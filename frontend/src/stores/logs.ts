import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import type { AndroidDevice, InstalledPackage, LogBatch, LogEntry, LogStatus, PackagePIDState } from '@/types/backend'
import { backend } from '@/utils/wails'

const ALL_LEVELS = ['V', 'D', 'I', 'W', 'E', 'F']
const UI_LOG_LIMIT = 100000

export const useLogStore = defineStore('logs', () => {
  const devices = ref<AndroidDevice[]>([])
  const selectedSerial = ref('')
  const logs = ref<LogEntry[]>([])
  const selectedLog = ref<LogEntry | null>(null)
  const packages = ref<InstalledPackage[]>([])
  const packageMode = ref<'thirdParty' | 'all'>('thirdParty')
  const selectedPackage = ref('')
  const packageLoading = ref(false)
  const packagePIDState = ref<PackagePIDState>({ packageName: '' })
  const levels = ref<string[]>([...ALL_LEVELS])
  const search = ref('')
  const paused = ref(false)
  const loading = ref(false)
  const error = ref('')
  const notice = ref('')
  const status = ref<LogStatus>({
    running: false,
    serial: '',
    count: 0,
    discardedCount: 0,
    lastID: 0
  })

  const running = computed(() => status.value.running)
  const selectedDevice = computed(() => devices.value.find((device) => device.serial === selectedSerial.value))
  const selectedDeviceState = computed(() => selectedDevice.value?.state ?? 'unknown')
  const canStart = computed(() => Boolean(selectedSerial.value) && selectedDeviceState.value === 'device' && !running.value)
  const canSelectPackage = computed(() => Boolean(selectedSerial.value) && selectedDeviceState.value === 'device')
  const currentPIDs = computed(() => packagePIDState.value.pids ?? [])
  const knownPIDs = computed(() => packagePIDState.value.knownPids ?? [])
  const packageHint = computed(() => {
    if (!selectedPackage.value) {
      return ''
    }
    if (!running.value) {
      return 'Start Logcat to track package PIDs.'
    }
    if (currentPIDs.value.length === 0) {
      return '目标 App 当前未运行，等待 PID 出现。'
    }
    return `Tracking PID ${currentPIDs.value.join(', ')}`
  })
  const deviceHint = computed(() => {
    if (devices.value.length === 0) {
      return 'No Android devices found. Connect a device and refresh.'
    }
    if (selectedDeviceState.value === 'unauthorized') {
      return '请在手机上允许 USB 调试授权，然后刷新设备。'
    }
    if (selectedDeviceState.value === 'offline') {
      return '设备处于 offline 状态，请重新连接数据线或刷新设备。'
    }
    if (selectedSerial.value && selectedDeviceState.value !== 'device') {
      return `设备状态为 ${selectedDeviceState.value}，暂不能启动 Logcat。`
    }
    return ''
  })
  const tableEmptyMessage = computed(() => {
    if (devices.value.length === 0) {
      return 'No device connected.'
    }
    if (!running.value && logs.value.length === 0) {
      return 'Logcat is stopped. Choose a device and press Start.'
    }
    if (running.value && logs.value.length === 0) {
      return 'Logcat is running. Waiting for log entries...'
    }
    if (selectedPackage.value && running.value && currentPIDs.value.length === 0 && filteredLogs.value.length === 0) {
      return '目标 App 当前未运行，等待 PID 出现。'
    }
    if (logs.value.length > 0 && filteredLogs.value.length === 0) {
      return 'No logs match the current search or level filter.'
    }
    return 'No log entries yet.'
  })
  const filteredLogs = computed(() => {
    const activeLevels = new Set(levels.value)
    const keyword = search.value.trim().toLowerCase()

    return logs.value.filter((entry) => {
      if (!activeLevels.has(entry.level)) {
        return false
      }
      if (selectedPackage.value && entry.packageName !== selectedPackage.value) {
        return false
      }
      if (!keyword) {
        return true
      }
      const haystack = [
        entry.tag,
        entry.message,
        entry.raw,
        ...(entry.multiline ?? [])
      ]
        .join('\n')
        .toLowerCase()
      return haystack.includes(keyword)
    })
  })

  let pollTimer: number | undefined

  async function refreshDevices() {
    loading.value = true
    error.value = ''
    try {
      const found = await backend.listDevices()
      devices.value = found

      const enriched = await Promise.all(
        found.map(async (device) => {
          if (device.state !== 'device') {
            return device
          }
          try {
            const info = await backend.getDeviceInfo(device.serial)
            return { ...device, ...info, state: device.state }
          } catch {
            return device
          }
        })
      )
      devices.value = enriched

      if (!enriched.some((device) => device.serial === selectedSerial.value)) {
        await selectDevice(enriched[0]?.serial ?? '')
      } else if (selectedSerial.value) {
        await refreshPackages()
      }
      await fetchStatus()
    } catch (err) {
      setError(err)
    } finally {
      loading.value = false
    }
  }

  async function start() {
    if (!selectedSerial.value) {
      error.value = '请选择一个已连接的 Android 设备。'
      return
    }
    if (selectedDeviceState.value === 'unauthorized') {
      error.value = '请在手机上允许 USB 调试授权，然后刷新设备。'
      return
    }
    if (selectedDeviceState.value === 'offline') {
      error.value = '设备处于 offline 状态，请重新连接数据线或刷新设备。'
      return
    }
    if (selectedDeviceState.value !== 'device') {
      error.value = `设备状态为 ${selectedDeviceState.value}，暂不能启动 Logcat。`
      return
    }

    loading.value = true
    error.value = ''
    notice.value = ''
    try {
      logs.value = []
      selectedLog.value = null
      paused.value = false
      await backend.startLogcat(selectedSerial.value)
      if (selectedPackage.value) {
        await backend.setTrackedPackage(selectedSerial.value, selectedPackage.value)
      }
      await fetchStatus()
      await fetchPackagePIDState()
      startPolling()
    } catch (err) {
      setError(err)
      await fetchStatus()
    } finally {
      loading.value = false
    }
  }

  async function stop() {
    loading.value = true
    try {
      await backend.stopLogcat()
      stopPolling()
      await fetchStatus()
    } catch (err) {
      setError(err)
    } finally {
      loading.value = false
    }
  }

  async function selectDevice(serial: string | null) {
    const nextSerial = serial ?? ''
    if (nextSerial === selectedSerial.value) {
      return
    }
    if (running.value) {
      await stop()
    }
    selectedSerial.value = nextSerial
    selectedPackage.value = ''
    packages.value = []
    packagePIDState.value = { packageName: '' }
    await backend.setTrackedPackage(nextSerial, '')
    if (nextSerial) {
      await refreshPackages()
    }
    selectedLog.value = null
    error.value = ''
  }

  async function refreshPackages() {
    packages.value = []
    if (!canSelectPackage.value) {
      return
    }
    packageLoading.value = true
    try {
      packages.value = await backend.listPackages(selectedSerial.value, packageMode.value)
    } catch (err) {
      setError(err)
    } finally {
      packageLoading.value = false
    }
  }

  async function setPackageMode(mode: 'thirdParty' | 'all') {
    packageMode.value = mode
    await refreshPackages()
  }

  async function selectPackage(packageName: string | null) {
    selectedPackage.value = packageName ?? ''
    selectedLog.value = null
    packagePIDState.value = { packageName: selectedPackage.value }
    if (!selectedSerial.value) {
      return
    }
    try {
      await backend.setTrackedPackage(selectedSerial.value, selectedPackage.value)
      await fetchPackagePIDState()
    } catch (err) {
      setError(err)
    }
  }

  async function clear() {
    await backend.clearLogs()
    logs.value = []
    selectedLog.value = null
    await fetchStatus()
  }

  function clearSearch() {
    search.value = ''
    levels.value = [...ALL_LEVELS]
  }

  async function exportFiltered() {
    error.value = ''
    notice.value = ''
    try {
      const path = await backend.exportLogs(filteredLogs.value)
      notice.value = `Exported ${filteredLogs.value.length} log entries to ${path}`
    } catch (err) {
      setError(err)
    }
  }

  function togglePause() {
    paused.value = !paused.value
    if (paused.value) {
      stopPolling()
    } else if (running.value) {
      startPolling()
      void fetchBatch()
    }
  }

  async function fetchBatch() {
    if (paused.value) {
      return
    }

    try {
      const lastLog = logs.value[logs.value.length - 1]
      const afterID = lastLog?.id ?? 0
      const batch: LogBatch = await backend.getLogBatch(afterID, 2000)
      status.value = {
        ...status.value,
        count: batch.count,
        discardedCount: batch.discardedCount,
        lastID: batch.lastID
      }
      await fetchPackagePIDState()

      if (batch.entries.length > 0) {
        logs.value.push(...batch.entries)
        if (logs.value.length > UI_LOG_LIMIT) {
          logs.value.splice(0, logs.value.length - UI_LOG_LIMIT)
        }
      }
    } catch (err) {
      setError(err)
      stopPolling()
      await fetchStatus()
    }
  }

  async function fetchStatus() {
    try {
      status.value = await backend.getLogStatus()
      if (status.value.lastError) {
        error.value = status.value.lastError
      }
    } catch (err) {
      setError(err)
    }
  }

  async function fetchPackagePIDState() {
    try {
      packagePIDState.value = await backend.getPackagePIDState()
    } catch (err) {
      setError(err)
    }
  }

  function startPolling() {
    if (pollTimer !== undefined) {
      return
    }
    pollTimer = window.setInterval(() => {
      void fetchBatch()
      void fetchStatus()
      void fetchPackagePIDState()
    }, 250)
  }

  function stopPolling() {
    if (pollTimer === undefined) {
      return
    }
    window.clearInterval(pollTimer)
    pollTimer = undefined
  }

  function selectLog(entry: LogEntry) {
    selectedLog.value = entry
  }

  function setError(err: unknown) {
    error.value = err instanceof Error ? err.message : String(err)
  }

  return {
    devices,
    selectedSerial,
    selectedDevice,
    logs,
    filteredLogs,
    selectedLog,
    packages,
    packageMode,
    selectedPackage,
    packageLoading,
    packagePIDState,
    currentPIDs,
    knownPIDs,
    levels,
    search,
    paused,
    loading,
    error,
    notice,
    status,
    running,
    selectedDeviceState,
    canStart,
    canSelectPackage,
    deviceHint,
    packageHint,
    tableEmptyMessage,
    refreshDevices,
    refreshPackages,
    setPackageMode,
    selectDevice,
    selectPackage,
    start,
    stop,
    clear,
    clearSearch,
    exportFiltered,
    togglePause,
    fetchStatus,
    startPolling,
    stopPolling,
    selectLog
  }
})
