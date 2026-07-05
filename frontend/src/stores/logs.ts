import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import type {
  AnalysisResult,
  AIContextOptions,
  AppConfig,
  APKInfo,
  AndroidDevice,
  BuildResult,
  FilterPreset,
  InstalledPackage,
  InstallResult,
  LogBatch,
  LogEntry,
  LogStatus,
  PackagePIDState,
  ProjectConfig,
  LaunchResult,
  WorkspaceConfig
} from '@/types/backend'
import { backend } from '@/utils/wails'

const ALL_LEVELS = ['V', 'D', 'I', 'W', 'E', 'F']
const UI_LOG_LIMIT = 100000

function defaultProjectConfig(): ProjectConfig {
  return {
    projectPath: '',
    packageName: '',
    lastApkPath: '',
    defaultBuildTask: 'assembleDebug',
    installOptions: {
      allowDowngrade: false,
      grantPermissions: true,
      allowTestOnly: false
    }
  }
}

function defaultAIContextOptions(): AIContextOptions {
  return {
    includeDeviceInfo: true,
    includePackageInfo: true,
    includeAnalysisSummary: true,
    includeRelatedLogs: true,
    includeBeforeContextLines: 50,
    includeAfterContextLines: 50,
    includeRawText: true,
    includeSuggestions: true,
    language: 'zh-CN'
  }
}

function defaultAppConfig(): AppConfig {
  const workspace = defaultWorkspaceConfig()
  return {
    activeWorkspaceId: workspace.id,
    workspaces: [workspace],
    filterPresets: []
  }
}

function defaultWorkspaceConfig(): WorkspaceConfig {
  return {
    id: 'default',
    workspaceName: 'Default Workspace',
    projectPath: '',
    packageName: '',
    lastApkPath: '',
    defaultBuildTask: 'assembleDebug',
    installOptions: defaultProjectConfig().installOptions,
    selectedDeviceSerial: '',
    selectedLogLevel: [...ALL_LEVELS],
    searchKeyword: '',
    selectedPackageMode: 'thirdParty',
    maxLogLines: UI_LOG_LIMIT,
    autoStartLogcat: false,
    autoClearOnLaunch: false,
    aiContextOptions: defaultAIContextOptions()
  }
}

export const useLogStore = defineStore('logs', () => {
  const initialConfig = defaultAppConfig()
  const devices = ref<AndroidDevice[]>([])
  const selectedSerial = ref('')
  const logs = ref<LogEntry[]>([])
  const selectedLog = ref<LogEntry | null>(null)
  const analysisResults = ref<AnalysisResult[]>([])
  const selectedAnalysis = ref<AnalysisResult | null>(null)
  const packages = ref<InstalledPackage[]>([])
  const packageMode = ref<'thirdParty' | 'all'>('thirdParty')
  const selectedPackage = ref('')
  const packageLoading = ref(false)
  const packagePIDState = ref<PackagePIDState>({ packageName: '' })
  const levels = ref<string[]>([...ALL_LEVELS])
  const search = ref('')
  const regexEnabled = ref(false)
  const tagFilter = ref('')
  const excludeKeyword = ref('')
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
  const projectConfig = ref<ProjectConfig>(defaultProjectConfig())
  const latestAPK = ref<APKInfo | null>(null)
  const buildResult = ref<BuildResult | null>(null)
  const installResult = ref<InstallResult | null>(null)
  const launchResult = ref<LaunchResult | null>(null)
  const buildOutput = ref('')
  const installOutput = ref('')
  const buildLoading = ref(false)
  const installLoading = ref(false)
  const launchLoading = ref(false)
  const appConfig = ref<AppConfig>(initialConfig)
  const workspaces = ref<WorkspaceConfig[]>(initialConfig.workspaces)
  const activeWorkspaceID = ref(initialConfig.activeWorkspaceId)
  const filterPresets = ref<FilterPreset[]>(initialConfig.filterPresets)
  const selectedPresetID = ref('')
  const presetDraftName = ref('')
  const presetManagerOpen = ref(false)
  const aiContextOptions = ref<AIContextOptions>(defaultAIContextOptions())
  const workspaceName = ref('Default Workspace')
  const autoStartLogcat = ref(false)
  const autoClearOnLaunch = ref(false)

  const running = computed(() => status.value.running)
  const selectedDevice = computed(() => devices.value.find((device) => device.serial === selectedSerial.value))
  const selectedDeviceState = computed(() => selectedDevice.value?.state ?? 'unknown')
  const canStart = computed(() => Boolean(selectedSerial.value) && selectedDeviceState.value === 'device' && !running.value)
  const canSelectPackage = computed(() => Boolean(selectedSerial.value) && selectedDeviceState.value === 'device')
  const canUseDeviceActions = computed(() => Boolean(selectedSerial.value) && selectedDeviceState.value === 'device')
  const canBuildProject = computed(() => Boolean(projectConfig.value.projectPath.trim()))
  const canInstallAPK = computed(() => canUseDeviceActions.value && Boolean(projectConfig.value.lastApkPath.trim()))
  const launchPackageName = computed(() => projectConfig.value.packageName.trim() || selectedPackage.value)
  const canLaunchProject = computed(() => canUseDeviceActions.value && Boolean(launchPackageName.value))
  const activeWorkspace = computed(() =>
    workspaces.value.find((workspace) => workspace.id === activeWorkspaceID.value) ?? null
  )
  const workspaceOptions = computed(() =>
    workspaces.value.map((workspace) => ({
      label: workspace.workspaceName,
      value: workspace.id
    }))
  )
  const presetOptions = computed(() =>
    filterPresets.value.map((preset) => ({
      label: preset.name,
      value: preset.id
    }))
  )
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
    const exclude = excludeKeyword.value.trim().toLowerCase()
    const tags = parseTags(tagFilter.value)
    const regex = regexEnabled.value && search.value.trim() ? safeRegex(search.value.trim()) : null

    return logs.value.filter((entry) => {
      if (!activeLevels.has(entry.level)) {
        return false
      }
      if (selectedPackage.value && entry.packageName !== selectedPackage.value) {
        return false
      }
      if (tags.length > 0 && !tags.some((tag) => entry.tag.toLowerCase().includes(tag))) {
        return false
      }
      if (!keyword) {
        return !exclude || !entryHaystack(entry).toLowerCase().includes(exclude)
      }
      const rawHaystack = entryHaystack(entry)
      const haystack = rawHaystack.toLowerCase()
      if (exclude && haystack.includes(exclude)) {
        return false
      }
      if (regexEnabled.value) {
        return Boolean(regex?.test(rawHaystack))
      }
      return haystack.includes(keyword)
    })
  })

  let pollTimer: number | undefined
  const analysisIDs = new Set<string>()

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
        await backend.setActiveDevice(selectedSerial.value)
        await refreshPackages()
      }
      await fetchStatus()
      if (autoStartLogcat.value && canStart.value) {
        notice.value = 'Auto-starting Logcat for restored workspace.'
        await start()
      }
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
      analysisResults.value = []
      selectedAnalysis.value = null
      analysisIDs.clear()
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
    await backend.setActiveDevice(nextSerial)
    selectedPackage.value = ''
    packages.value = []
    packagePIDState.value = { packageName: '' }
    await backend.setTrackedPackage(nextSerial, '')
    if (nextSerial) {
      await refreshPackages()
    }
    selectedLog.value = null
    selectedAnalysis.value = null
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
    if (selectedPackage.value && projectConfig.value.packageName !== selectedPackage.value) {
      projectConfig.value.packageName = selectedPackage.value
      void saveProjectConfig()
    }
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
    analysisResults.value = []
    selectedAnalysis.value = null
    analysisIDs.clear()
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

  function contextOptionsForRequest(): AIContextOptions {
    return {
      ...defaultAIContextOptions(),
      ...aiContextOptions.value,
      packageFilter: selectedPackage.value,
      levelFilter: [...levels.value],
      searchKeyword: search.value
    }
  }

  async function copyAIContext(resultID?: string) {
    const targetID = resultID || selectedAnalysis.value?.id
    if (!targetID) {
      throw new Error('请先选择一个分析结果。')
    }
    await backend.copyAIContext(targetID, contextOptionsForRequest())
  }

  async function exportAIContext(resultID?: string) {
    const targetID = resultID || selectedAnalysis.value?.id
    if (!targetID) {
      throw new Error('请先选择一个分析结果。')
    }
    return await backend.exportAIContext(targetID, contextOptionsForRequest())
  }

  async function loadConfig() {
    try {
      const config = await backend.loadConfig()
      applyAppConfig(config)
    } catch (err) {
      setError(err)
    }
  }

  async function saveProjectConfig() {
    await saveCurrentWorkspace()
  }

  async function resetConfig() {
    try {
      const config = await backend.resetConfig()
      applyAppConfig(config)
      notice.value = 'Configuration reset to defaults.'
    } catch (err) {
      setError(err)
    }
  }

  async function saveCurrentWorkspace() {
    try {
      const workspace = currentWorkspaceSnapshot()
      const config = await backend.saveWorkspace(workspace)
      applyAppConfig(config, false)
      notice.value = `Workspace saved: ${workspace.workspaceName}`
    } catch (err) {
      setError(err)
    }
  }

  async function createWorkspace() {
    try {
      const index = workspaces.value.length + 1
      const workspace = {
        ...currentWorkspaceSnapshot(),
        id: `workspace-${Date.now()}`,
        workspaceName: `Workspace ${index}`
      }
      const config = await backend.saveWorkspace(workspace)
      applyAppConfig(config)
      notice.value = `Workspace created: ${workspace.workspaceName}`
    } catch (err) {
      setError(err)
    }
  }

  async function selectWorkspace(id: string | number | null) {
    if (typeof id !== 'string' || id === activeWorkspaceID.value) {
      return
    }
    try {
      if (running.value) {
        await stop()
      }
      const config = await backend.setActiveWorkspace(id)
      applyAppConfig(config)
      notice.value = `Workspace selected: ${activeWorkspace.value?.workspaceName ?? id}`
    } catch (err) {
      setError(err)
    }
  }

  async function deleteCurrentWorkspace() {
    if (!activeWorkspaceID.value) {
      return
    }
    try {
      const config = await backend.deleteWorkspace(activeWorkspaceID.value)
      applyAppConfig(config)
      notice.value = 'Workspace deleted.'
    } catch (err) {
      setError(err)
    }
  }

  async function saveCurrentFilter() {
    try {
      const name = presetDraftName.value.trim() || `Preset ${filterPresets.value.length + 1}`
      const preset: FilterPreset = {
        id: `preset-${Date.now()}`,
        name,
        level: [...levels.value],
        packageName: selectedPackage.value,
        keyword: search.value,
        regexEnabled: regexEnabled.value,
        tags: parseTags(tagFilter.value),
        excludeKeyword: excludeKeyword.value
      }
      const config = await backend.saveFilterPreset(preset)
      applyAppConfig(config, false)
      selectedPresetID.value = preset.id
      presetDraftName.value = ''
      notice.value = `Filter preset saved: ${name}`
    } catch (err) {
      setError(err)
    }
  }

  async function renamePreset(preset: FilterPreset, name: string) {
    if (preset.builtIn) {
      return
    }
    try {
      const config = await backend.saveFilterPreset({ ...preset, name })
      applyAppConfig(config, false)
    } catch (err) {
      setError(err)
    }
  }

  async function deletePreset(id: string) {
    try {
      const config = await backend.deleteFilterPreset(id)
      applyAppConfig(config, false)
      if (selectedPresetID.value === id) {
        selectedPresetID.value = ''
      }
    } catch (err) {
      setError(err)
    }
  }

  async function applyPreset(id: string | number | null) {
    selectedPresetID.value = typeof id === 'string' ? id : ''
    const preset = filterPresets.value.find((item) => item.id === selectedPresetID.value)
    if (!preset) {
      return
    }
    levels.value = preset.level.length ? [...preset.level] : [...ALL_LEVELS]
    regexEnabled.value = preset.regexEnabled
    search.value = preset.keyword
    tagFilter.value = preset.tags.join(', ')
    excludeKeyword.value = preset.excludeKeyword
    const presetPackage = preset.packageName === '$current'
      ? projectConfig.value.packageName || selectedPackage.value
      : preset.packageName
    if (presetPackage !== selectedPackage.value) {
      await selectPackage(presetPackage || null)
    }
    notice.value = `Filter preset applied: ${preset.name}`
  }

  async function chooseProjectDirectory() {
    try {
      const path = await backend.selectProjectDirectory()
      if (!path) {
        return
      }
      projectConfig.value.projectPath = path
      await saveProjectConfig()
      notice.value = `Project selected: ${path}`
    } catch (err) {
      setError(err)
    }
  }

  async function findLatestAPK() {
    error.value = ''
    notice.value = ''
    try {
      const apk = await backend.findLatestAPK(projectConfig.value.projectPath)
      latestAPK.value = apk
      projectConfig.value.lastApkPath = apk.apkPath
      await saveProjectConfig()
      notice.value = `Found APK: ${apk.fileName}`
      return apk
    } catch (err) {
      setError(err)
      return null
    }
  }

  async function buildDebug() {
    buildLoading.value = true
    error.value = ''
    notice.value = ''
    buildOutput.value = ''
    try {
      await saveProjectConfig()
      const result = await backend.buildDebug(projectConfig.value.projectPath)
      buildResult.value = result
      buildOutput.value = result.output || result.error || ''
      if (result.apk) {
        latestAPK.value = result.apk
        projectConfig.value.lastApkPath = result.apk.apkPath
        await saveProjectConfig()
      }
      if (result.success) {
        notice.value = `Build succeeded: ${result.apk?.fileName ?? 'APK generated'}`
      } else {
        error.value = result.error || 'Build failed.'
      }
      return result
    } catch (err) {
      setError(err)
      return null
    } finally {
      buildLoading.value = false
    }
  }

  async function installAPK(apkPath = projectConfig.value.lastApkPath) {
    installLoading.value = true
    error.value = ''
    notice.value = ''
    installOutput.value = ''
    try {
      const result = await backend.installAPK(apkPath, projectConfig.value.installOptions)
      installResult.value = result
      installOutput.value = result.output || result.error || ''
      mergeAnalysisResults(result.analysisResults ?? [])
      if (result.success) {
        projectConfig.value.lastApkPath = result.apkPath
        await saveProjectConfig()
        notice.value = 'Install succeeded.'
      } else {
        error.value = result.error || 'Install failed.'
      }
      return result
    } catch (err) {
      setError(err)
      return null
    } finally {
      installLoading.value = false
    }
  }

  async function buildAndInstall() {
    const built = await buildDebug()
    if (!built?.success || !built.apk) {
      return
    }
    await installAPK(built.apk.apkPath)
  }

  async function launchApp() {
    launchLoading.value = true
    error.value = ''
    notice.value = ''
    try {
      const packageName = launchPackageName.value
      const result = await backend.launchApp(packageName)
      launchResult.value = result
      if (result.success) {
        projectConfig.value.packageName = result.packageName
        await saveProjectConfig()
        if (autoClearOnLaunch.value) {
          await clear()
        }
        await selectPackage(result.packageName)
        await fetchPackagePIDState()
        notice.value = running.value
          ? `Launched ${result.packageName}.`
          : `Launched ${result.packageName}. Start Logcat to stream logs.`
      } else {
        error.value = result.error || 'Launch failed.'
      }
      return result
    } catch (err) {
      setError(err)
      return null
    } finally {
      launchLoading.value = false
    }
  }

  async function buildInstallLaunch() {
    buildLoading.value = true
    installLoading.value = true
    launchLoading.value = true
    error.value = ''
    notice.value = ''
    buildOutput.value = ''
    installOutput.value = ''
    try {
      await saveProjectConfig()
      const result = await backend.buildInstallLaunch(projectConfig.value)
      buildResult.value = result.build
      installResult.value = result.install
      launchResult.value = result.launch
      buildOutput.value = result.build?.output || result.build?.error || ''
      installOutput.value = result.install?.output || result.install?.error || ''
      mergeAnalysisResults(result.analysisResults ?? result.install?.analysisResults ?? [])
      if (result.apk) {
        latestAPK.value = result.apk
        projectConfig.value.lastApkPath = result.apk.apkPath
      }
      if (result.launch?.success) {
        projectConfig.value.packageName = result.launch.packageName
        await saveProjectConfig()
        if (autoClearOnLaunch.value) {
          await clear()
        }
        await selectPackage(result.launch.packageName)
        await fetchPackagePIDState()
        notice.value = running.value
          ? `Built, installed, and launched ${result.launch.packageName}.`
          : `Built, installed, and launched ${result.launch.packageName}. Start Logcat to stream logs.`
      } else if (!result.build?.success) {
        error.value = result.build?.error || 'Build failed.'
      } else if (!result.install?.success) {
        error.value = result.install?.error || 'Install failed.'
      } else {
        error.value = result.launch?.error || 'Launch did not complete.'
      }
    } catch (err) {
      setError(err)
    } finally {
      buildLoading.value = false
      installLoading.value = false
      launchLoading.value = false
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
        await analyzeIncremental(batch.entries)
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

  async function analyzeIncremental(entries: LogEntry[]) {
    if (entries.length === 0) {
      return
    }
    try {
      const results = await backend.analyzeLogs(entries)
      mergeAnalysisResults(results)
    } catch (err) {
      setError(err)
    }
  }

  async function analyzeCurrentLogs() {
    error.value = ''
    notice.value = ''
    try {
      const results = await backend.analyzeLogs(filteredLogs.value)
      analysisResults.value = []
      selectedAnalysis.value = null
      analysisIDs.clear()
      mergeAnalysisResults(results)
      notice.value = `Analyzed ${filteredLogs.value.length} logs, found ${results.length} issue(s).`
    } catch (err) {
      setError(err)
    }
  }

  function mergeAnalysisResults(results: AnalysisResult[]) {
    for (const result of results) {
      if (analysisIDs.has(result.id)) {
        continue
      }
      analysisIDs.add(result.id)
      analysisResults.value.unshift(result)
    }
  }

  function selectAnalysis(result: AnalysisResult) {
    selectedAnalysis.value = result
    const firstID = result.relatedEntryIds?.[0]
    if (!firstID) {
      return
    }
    const related = logs.value.find((entry) => entry.id === firstID)
    if (related) {
      selectedLog.value = related
    }
  }

  function applyAppConfig(config: AppConfig, applyWorkspace = true) {
    appConfig.value = {
      ...defaultAppConfig(),
      ...config
    }
    workspaces.value = appConfig.value.workspaces ?? []
    filterPresets.value = appConfig.value.filterPresets ?? []
    activeWorkspaceID.value = appConfig.value.activeWorkspaceId
    if (applyWorkspace) {
      const workspace = activeWorkspace.value ?? workspaces.value[0] ?? defaultWorkspaceConfig()
      applyWorkspaceConfig(workspace)
    }
  }

  function applyWorkspaceConfig(workspace: WorkspaceConfig) {
    projectConfig.value = {
      projectPath: workspace.projectPath,
      packageName: workspace.packageName,
      lastApkPath: workspace.lastApkPath,
      defaultBuildTask: workspace.defaultBuildTask || 'assembleDebug',
      installOptions: {
        ...defaultProjectConfig().installOptions,
        ...(workspace.installOptions ?? {})
      }
    }
    workspaceName.value = workspace.workspaceName || 'Default Workspace'
    autoStartLogcat.value = workspace.autoStartLogcat
    autoClearOnLaunch.value = workspace.autoClearOnLaunch
    selectedSerial.value = workspace.selectedDeviceSerial || ''
    void backend.setActiveDevice(selectedSerial.value)
    levels.value = workspace.selectedLogLevel?.length ? [...workspace.selectedLogLevel] : [...ALL_LEVELS]
    search.value = workspace.searchKeyword || ''
    packageMode.value = workspace.selectedPackageMode === 'all' ? 'all' : 'thirdParty'
    selectedPackage.value = workspace.packageName || ''
    packagePIDState.value = { packageName: selectedPackage.value }
    aiContextOptions.value = {
      ...defaultAIContextOptions(),
      ...(workspace.aiContextOptions ?? {})
    }
    latestAPK.value = workspace.lastApkPath
      ? {
          apkPath: workspace.lastApkPath,
          fileName: workspace.lastApkPath.split(/[\\/]/).pop() || workspace.lastApkPath,
          modifiedTime: '',
          size: 0
        }
      : null
  }

  function currentWorkspaceSnapshot(): WorkspaceConfig {
    const base = activeWorkspace.value ?? defaultWorkspaceConfig()
    return {
      ...base,
      workspaceName: workspaceName.value || base.workspaceName || 'Default Workspace',
      projectPath: projectConfig.value.projectPath,
      packageName: projectConfig.value.packageName || selectedPackage.value,
      lastApkPath: projectConfig.value.lastApkPath,
      defaultBuildTask: projectConfig.value.defaultBuildTask || 'assembleDebug',
      installOptions: projectConfig.value.installOptions,
      selectedDeviceSerial: selectedSerial.value,
      selectedLogLevel: [...levels.value],
      searchKeyword: search.value,
      selectedPackageMode: packageMode.value,
      maxLogLines: UI_LOG_LIMIT,
      autoStartLogcat: autoStartLogcat.value,
      autoClearOnLaunch: autoClearOnLaunch.value,
      aiContextOptions: aiContextOptions.value
    }
  }

  function entryHaystack(entry: LogEntry) {
    return [
      entry.tag,
      entry.message,
      entry.raw,
      ...(entry.multiline ?? [])
    ].join('\n')
  }

  function parseTags(value: string) {
    return value
      .split(/[,\s]+/)
      .map((item) => item.trim().toLowerCase())
      .filter(Boolean)
  }

  function safeRegex(pattern: string) {
    try {
      return new RegExp(pattern, 'i')
    } catch {
      return null
    }
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
    analysisResults,
    selectedAnalysis,
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
    appConfig,
    workspaces,
    activeWorkspaceID,
    activeWorkspace,
    workspaceOptions,
    filterPresets,
    presetOptions,
    selectedPresetID,
    presetDraftName,
    presetManagerOpen,
    workspaceName,
    autoStartLogcat,
    autoClearOnLaunch,
    regexEnabled,
    tagFilter,
    excludeKeyword,
    aiContextOptions,
    projectConfig,
    latestAPK,
    buildResult,
    installResult,
    launchResult,
    buildOutput,
    installOutput,
    buildLoading,
    installLoading,
    launchLoading,
    running,
    selectedDeviceState,
    canStart,
    canSelectPackage,
    canUseDeviceActions,
    canBuildProject,
    canInstallAPK,
    canLaunchProject,
    launchPackageName,
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
    copyAIContext,
    exportAIContext,
    loadConfig,
    resetConfig,
    saveProjectConfig,
    saveCurrentWorkspace,
    createWorkspace,
    selectWorkspace,
    deleteCurrentWorkspace,
    saveCurrentFilter,
    renamePreset,
    deletePreset,
    applyPreset,
    chooseProjectDirectory,
    findLatestAPK,
    buildDebug,
    installAPK,
    buildAndInstall,
    launchApp,
    buildInstallLaunch,
    analyzeCurrentLogs,
    togglePause,
    fetchStatus,
    startPolling,
    stopPolling,
    selectLog,
    selectAnalysis
  }
})
