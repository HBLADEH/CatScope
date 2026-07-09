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
  SessionFilters,
  SessionSummary,
  LaunchResult,
  WorkspaceConfig
} from '@/types/backend'
import { getCurrentLocale, localizePresetName, t } from '@/i18n'
import { backend } from '@/utils/wails'

const ALL_LEVELS = ['V', 'D', 'I', 'W', 'E', 'F']
const UI_LOG_LIMIT = 100000

type QueryField = 'any' | 'tag' | 'message' | 'package' | 'process' | 'pid' | 'tid' | 'level' | 'raw'
type MatchMode = 'contains' | 'equals' | 'startsWith' | 'regex'

interface QueryFilter {
  field: QueryField
  mode: MatchMode
  value: string
  negative: boolean
  regex: RegExp | null
}

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
    language: getCurrentLocale()
  }
}

function defaultAppConfig(): AppConfig {
  const workspace = defaultWorkspaceConfig()
  return {
    activeWorkspaceId: workspace.id,
    adbPath: '',
    workspaces: [workspace],
    filterPresets: []
  }
}

function defaultWorkspaceConfig(): WorkspaceConfig {
  return {
    id: 'default',
    workspaceName: t('defaults.defaultWorkspace'),
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
    lastID: 0,
    source: 'live'
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
  const adbPathInput = ref('')
  const workspaces = ref<WorkspaceConfig[]>(initialConfig.workspaces)
  const activeWorkspaceID = ref(initialConfig.activeWorkspaceId)
  const filterPresets = ref<FilterPreset[]>(initialConfig.filterPresets)
  const selectedPresetID = ref('')
  const presetDraftName = ref('')
  const presetManagerOpen = ref(false)
  const aiContextOptions = ref<AIContextOptions>(defaultAIContextOptions())
  const workspaceName = ref(t('defaults.defaultWorkspace'))
  const autoStartLogcat = ref(false)
  const autoClearOnLaunch = ref(false)
  const offlinePathInput = ref('')
  const offlineLoading = ref(false)
  const sessionPathInput = ref('')
  const sessionNameInput = ref('')
  const sessionNotes = ref('')
  const sessionLoading = ref(false)
  const currentSession = ref<SessionSummary | null>(null)

  const running = computed(() => status.value.running)
  const logSource = computed(() => status.value.source ?? 'live')
  const isOffline = computed(() => logSource.value === 'offline')
  const isSession = computed(() => logSource.value === 'session')
  const isStaticSource = computed(() => isOffline.value || isSession.value)
  const selectedDevice = computed(() => devices.value.find((device) => device.serial === selectedSerial.value))
  const resolvedADBPath = computed(() => status.value.adbPath || '')
  const selectedDeviceState = computed(() => selectedDevice.value?.state ?? 'unknown')
  const canStart = computed(() => !isStaticSource.value && Boolean(selectedSerial.value) && selectedDeviceState.value === 'device' && !running.value)
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
      label: localizePresetName(preset.id, preset.name),
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
      return t('hint.startPidTracking')
    }
    if (currentPIDs.value.length === 0) {
      return t('hint.appNotRunning')
    }
    return t('hint.trackingPid', { pids: currentPIDs.value.join(', ') })
  })
  const deviceHint = computed(() => {
    if (devices.value.length === 0) {
      return t('hint.noDevices')
    }
    if (selectedDeviceState.value === 'unauthorized') {
      return t('hint.unauthorized')
    }
    if (selectedDeviceState.value === 'offline') {
      return t('hint.offline')
    }
    if (selectedSerial.value && selectedDeviceState.value !== 'device') {
      return t('hint.badState', { state: selectedDeviceState.value })
    }
    return ''
  })
  const tableEmptyMessage = computed(() => {
    if (isSession.value && logs.value.length === 0) {
      return t('empty.session')
    }
    if (isOffline.value && logs.value.length === 0) {
      return t('empty.offline')
    }
    if (!isStaticSource.value && devices.value.length === 0) {
      return t('empty.noDevice')
    }
    if (!running.value && logs.value.length === 0) {
      return t('empty.stopped')
    }
    if (running.value && logs.value.length === 0) {
      return t('empty.running')
    }
    if (selectedPackage.value && running.value && currentPIDs.value.length === 0 && filteredLogs.value.length === 0) {
      return t('hint.appNotRunning')
    }
    if (logs.value.length > 0 && filteredLogs.value.length === 0) {
      return t('empty.noMatch')
    }
    return t('empty.default')
  })
  const filteredLogs = computed(() => {
    const activeLevels = new Set(levels.value)
    const query = parseSearchQuery(search.value)
    const keyword = query.text.trim().toLowerCase()
    const exclude = excludeKeyword.value.trim().toLowerCase()
    const tags = parseTags(tagFilter.value)
    const regex = regexEnabled.value && query.text.trim() ? safeRegex(query.text.trim()) : null

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
      if (query.filters.length > 0 && !matchesQueryFilters(entry, query.filters)) {
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
        notice.value = t('notice.autoStart')
        await start()
      }
    } catch (err) {
      setError(err)
    } finally {
      loading.value = false
    }
  }

  async function chooseADBExecutable() {
    try {
      const path = await backend.selectADBExecutable()
      if (!path) {
        return
      }
      adbPathInput.value = path
      await useADBPath(path)
    } catch (err) {
      setError(err)
    }
  }

  async function useADBPath(path = adbPathInput.value) {
    loading.value = true
    error.value = ''
    notice.value = ''
    try {
      const resolved = await backend.useADBPath(path.trim())
      adbPathInput.value = resolved
      devices.value = []
      selectedSerial.value = ''
      packages.value = []
      selectedPackage.value = ''
      packagePIDState.value = { packageName: '' }
      await fetchStatus()
      await loadConfig()
      notice.value = resolved
        ? t('notice.usingAdb', { path: resolved })
        : t('notice.adbCleared')
      await refreshDevices()
    } catch (err) {
      setError(err)
      await fetchStatus()
    } finally {
      loading.value = false
    }
  }

  async function start() {
    if (!selectedSerial.value) {
      error.value = t('error.selectDevice')
      return
    }
    if (selectedDeviceState.value === 'unauthorized') {
      error.value = t('hint.unauthorized')
      return
    }
    if (selectedDeviceState.value === 'offline') {
      error.value = t('hint.offline')
      return
    }
    if (selectedDeviceState.value !== 'device') {
      error.value = t('hint.badState', { state: selectedDeviceState.value })
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
      currentSession.value = null
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

  async function openLogFile(path = offlinePathInput.value) {
    offlineLoading.value = true
    error.value = ''
    notice.value = ''
    try {
      stopPolling()
      const result = await backend.openLogFile(path.trim())
      logs.value = result.entries ?? []
      selectedLog.value = null
      analysisResults.value = []
      selectedAnalysis.value = null
      analysisIDs.clear()
      mergeAnalysisResults(result.analysisResults ?? [])
      currentSession.value = null
      paused.value = false
      offlinePathInput.value = result.filePath
      await fetchStatus()
      notice.value = t('notice.openedLog', {
        fileName: result.fileName,
        count: result.count,
        rawCount: result.parseFailedCount
      })
    } catch (err) {
      setError(err)
      await fetchStatus()
    } finally {
      offlineLoading.value = false
    }
  }

  async function returnToLiveMode() {
    loading.value = true
    error.value = ''
    notice.value = ''
    try {
      stopPolling()
      await backend.returnToLiveMode()
      logs.value = []
      selectedLog.value = null
      analysisResults.value = []
      selectedAnalysis.value = null
      analysisIDs.clear()
      currentSession.value = null
      paused.value = false
      await fetchStatus()
      notice.value = t('notice.returnedLive')
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
    if (currentSession.value) {
      currentSession.value = {
        ...currentSession.value,
        logCount: 0,
        analysisCount: 0
      }
    }
    await fetchStatus()
  }

  function clearSearch() {
    search.value = ''
    levels.value = [...ALL_LEVELS]
    tagFilter.value = ''
    excludeKeyword.value = ''
    regexEnabled.value = false
  }

  function setTagSearchFilter(tag: string, negative: boolean) {
    const normalizedTag = tag.trim()
    if (!normalizedTag) {
      return
    }
    const tokens = tokenizeQuery(search.value)
    const nextToken = `${negative ? '-' : ''}tag:${formatQueryValue(normalizedTag)}`
    if (!tokens.some((token) => isSameTagQueryToken(token, normalizedTag, negative))) {
      tokens.push(nextToken)
    }
    search.value = tokens.join(' ')
  }

  async function exportFiltered() {
    error.value = ''
    notice.value = ''
    try {
      const path = await backend.exportLogs(filteredLogs.value)
      notice.value = t('notice.exported', { count: filteredLogs.value.length, path })
    } catch (err) {
      setError(err)
    }
  }

  async function exportFilteredJSONL() {
    error.value = ''
    notice.value = ''
    try {
      const path = await backend.exportLogsJSONL(filteredLogs.value)
      notice.value = t('notice.exportedJsonl', { count: filteredLogs.value.length, path })
    } catch (err) {
      setError(err)
    }
  }

  function currentSessionFilters(): SessionFilters {
    return {
      level: [...levels.value],
      packageName: selectedPackage.value,
      keyword: search.value,
      regexEnabled: regexEnabled.value,
      tags: parseTags(tagFilter.value),
      excludeKeyword: excludeKeyword.value
    }
  }

  async function saveSession(path = sessionPathInput.value) {
    sessionLoading.value = true
    error.value = ''
    notice.value = ''
    try {
      if (logs.value.length === 0) {
        throw new Error(t('error.noLogsToSave'))
      }
      const summary = await backend.saveSession(path.trim(), {
        name: sessionNameInput.value.trim() || currentSession.value?.name || workspaceName.value,
        filters: currentSessionFilters(),
        aiContextOptions: contextOptionsForRequest(),
        notes: sessionNotes.value
      })
      currentSession.value = summary
      sessionPathInput.value = summary.filePath
      sessionNameInput.value = summary.name
      notice.value = t('notice.savedSession', { logs: summary.logCount, issues: summary.analysisCount })
    } catch (err) {
      setError(err)
    } finally {
      sessionLoading.value = false
    }
  }

  async function openSession(path = sessionPathInput.value) {
    sessionLoading.value = true
    error.value = ''
    notice.value = ''
    try {
      stopPolling()
      const result = await backend.openSession(path.trim())
      logs.value = result.entries ?? []
      selectedLog.value = null
      analysisResults.value = []
      selectedAnalysis.value = null
      analysisIDs.clear()
      mergeAnalysisResults(result.analysisResults ?? result.session.analysisResults ?? [])
      applySessionFilters(result.session.filters)
      aiContextOptions.value = {
        ...defaultAIContextOptions(),
        ...(result.session.aiContextOptions ?? {}),
        language: getCurrentLocale()
      }
      projectConfig.value = {
        ...projectConfig.value,
        projectPath: result.session.projectPath || projectConfig.value.projectPath,
        packageName: result.session.packageName || projectConfig.value.packageName
      }
      selectedSerial.value = result.session.selectedDevice || ''
      selectedPackage.value = result.session.filters?.packageName || result.session.packageName || ''
      packagePIDState.value = {
        packageName: selectedPackage.value,
        knownPids: result.session.knownPids ?? []
      }
      currentSession.value = result.summary
      sessionPathInput.value = result.summary.filePath
      sessionNameInput.value = result.summary.name
      sessionNotes.value = result.session.notes || ''
      paused.value = false
      await fetchStatus()
      notice.value = t('notice.openedSession', {
        name: result.summary.name,
        logs: result.summary.logCount,
        issues: result.summary.analysisCount
      })
    } catch (err) {
      setError(err)
      await fetchStatus()
    } finally {
      sessionLoading.value = false
    }
  }

  function contextOptionsForRequest(): AIContextOptions {
    return {
      ...defaultAIContextOptions(),
      ...aiContextOptions.value,
      language: getCurrentLocale(),
      packageFilter: selectedPackage.value,
      levelFilter: [...levels.value],
      searchKeyword: search.value
    }
  }

  async function copyAIContext(resultID?: string) {
    const targetID = resultID || selectedAnalysis.value?.id
    if (!targetID) {
      throw new Error(t('error.selectAnalysis'))
    }
    await backend.copyAIContext(targetID, contextOptionsForRequest())
  }

  async function exportAIContext(resultID?: string) {
    const targetID = resultID || selectedAnalysis.value?.id
    if (!targetID) {
      throw new Error(t('error.selectAnalysis'))
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
      notice.value = t('notice.configReset')
    } catch (err) {
      setError(err)
    }
  }

  async function saveCurrentWorkspace() {
    try {
      const workspace = currentWorkspaceSnapshot()
      const config = await backend.saveWorkspace(workspace)
      applyAppConfig(config, false)
      notice.value = t('notice.workspaceSaved', { name: workspace.workspaceName })
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
        workspaceName: t('defaults.workspace', { index })
      }
      const config = await backend.saveWorkspace(workspace)
      applyAppConfig(config)
      notice.value = t('notice.workspaceCreated', { name: workspace.workspaceName })
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
      notice.value = t('notice.workspaceSelected', { name: activeWorkspace.value?.workspaceName ?? id })
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
      notice.value = t('notice.workspaceDeleted')
    } catch (err) {
      setError(err)
    }
  }

  async function saveCurrentFilter() {
    try {
      const name = presetDraftName.value.trim() || t('defaults.preset', { index: filterPresets.value.length + 1 })
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
      notice.value = t('notice.filterSaved', { name })
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
    notice.value = t('notice.filterApplied', { name: localizePresetName(preset.id, preset.name) })
  }

  function applySessionFilters(filters?: SessionFilters) {
    if (!filters) {
      return
    }
    levels.value = filters.level?.length ? [...filters.level] : [...ALL_LEVELS]
    selectedPackage.value = filters.packageName || ''
    search.value = filters.keyword || ''
    regexEnabled.value = filters.regexEnabled
    tagFilter.value = (filters.tags ?? []).join(', ')
    excludeKeyword.value = filters.excludeKeyword || ''
  }

  async function chooseProjectDirectory() {
    try {
      const path = await backend.selectProjectDirectory()
      if (!path) {
        return
      }
      projectConfig.value.projectPath = path
      await saveProjectConfig()
      notice.value = t('notice.projectSelected', { path })
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
      notice.value = t('notice.foundApk', { fileName: apk.fileName })
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
        notice.value = t('notice.buildSucceeded', { fileName: result.apk?.fileName ?? t('notice.apkGenerated') })
      } else {
        error.value = result.error || t('error.buildFailed')
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
        notice.value = t('notice.installSucceeded')
      } else {
        error.value = result.error || t('error.installFailed')
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
        if (autoClearOnLaunch.value && !isStaticSource.value) {
          await clear()
        }
        await selectPackage(result.packageName)
        await fetchPackagePIDState()
        notice.value = running.value
          ? t('notice.launched', { packageName: result.packageName })
          : t('notice.launchedNeedStart', { packageName: result.packageName })
      } else {
        error.value = result.error || t('error.launchFailed')
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
        if (autoClearOnLaunch.value && !isStaticSource.value) {
          await clear()
        }
        await selectPackage(result.launch.packageName)
        await fetchPackagePIDState()
        notice.value = running.value
          ? t('notice.buildInstallLaunch', { packageName: result.launch.packageName })
          : t('notice.buildInstallLaunchNeedStart', { packageName: result.launch.packageName })
      } else if (!result.build?.success) {
        error.value = result.build?.error || t('error.buildFailed')
      } else if (!result.install?.success) {
        error.value = result.install?.error || t('error.installFailed')
      } else {
        error.value = result.launch?.error || t('error.launchIncomplete')
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
      notice.value = t('notice.analyzed', { logs: filteredLogs.value.length, issues: results.length })
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
    adbPathInput.value = appConfig.value.adbPath || ''
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
    workspaceName.value = workspace.workspaceName || t('defaults.defaultWorkspace')
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
      ...(workspace.aiContextOptions ?? {}),
      language: getCurrentLocale()
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
      workspaceName: workspaceName.value || base.workspaceName || t('defaults.defaultWorkspace'),
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
      entry.packageName,
      entry.pid ? String(entry.pid) : '',
      entry.tid ? String(entry.tid) : '',
      entry.level,
      entry.tag,
      entry.message,
      entry.raw,
      ...(entry.multiline ?? [])
    ].join('\n')
  }

  function parseSearchQuery(value: string) {
    const filters: QueryFilter[] = []
    const textTerms: string[] = []

    for (const token of tokenizeQuery(value)) {
      const fieldMatch = token.match(/^(-?)(tag|message|msg|package|pkg|app|process|pid|tid|level|raw)(~:|\^:|\^=|~=|=|:)(.+)$/i)
      if (!fieldMatch) {
        textTerms.push(unquote(token))
        continue
      }
      const field = normalizeQueryField(fieldMatch[2])
      const mode = normalizeMatchMode(fieldMatch[3])
      const rawValue = unquote(fieldMatch[4].trim())
      if (!rawValue) {
        continue
      }
      filters.push({
        field,
        mode,
        value: normalizeQueryValue(field, rawValue),
        negative: fieldMatch[1] === '-',
        regex: mode === 'regex' ? safeRegex(rawValue) : null
      })
    }

    return {
      text: textTerms.join(' '),
      filters
    }
  }

  function tokenizeQuery(value: string) {
    const tokens: string[] = []
    let current = ''
    let quote = ''

    for (const char of value.trim()) {
      if ((char === '"' || char === "'") && (!quote || quote === char)) {
        quote = quote ? '' : char
        current += char
        continue
      }
      if (/\s/.test(char) && !quote) {
        if (current.trim()) {
          tokens.push(current.trim())
        }
        current = ''
        continue
      }
      current += char
    }

    if (current.trim()) {
      tokens.push(current.trim())
    }
    return tokens
  }

  function unquote(value: string) {
    const trimmed = value.trim()
    if (trimmed.length >= 2) {
      const first = trimmed[0]
      const last = trimmed[trimmed.length - 1]
      if ((first === '"' && last === '"') || (first === "'" && last === "'")) {
        return trimmed.slice(1, -1)
      }
    }
    return trimmed
  }

  function isTagQueryToken(value: string) {
    return /^-?tag(~:|\^:|\^=|~=|=|:).+$/i.test(value)
  }

  function isSameTagQueryToken(value: string, tag: string, negative: boolean) {
    const match = value.match(/^(-?)tag:(.+)$/i)
    if (!match) {
      return false
    }
    return (match[1] === '-') === negative && unquote(match[2]).toLowerCase() === tag.toLowerCase()
  }

  function formatQueryValue(value: string) {
    if (!/[\s"']/.test(value)) {
      return value
    }
    return `"${value.replace(/"/g, '')}"`
  }

  function normalizeQueryField(field: string): QueryField {
    switch (field.toLowerCase()) {
      case 'msg':
      case 'message':
        return 'message'
      case 'pkg':
      case 'app':
      case 'package':
        return 'package'
      case 'process':
        return 'process'
      case 'pid':
        return 'pid'
      case 'tid':
        return 'tid'
      case 'level':
        return 'level'
      case 'raw':
        return 'raw'
      default:
        return 'tag'
    }
  }

  function normalizeMatchMode(operator: string): MatchMode {
    if (operator === '=') {
      return 'equals'
    }
    if (operator === '^=' || operator === '^:') {
      return 'startsWith'
    }
    if (operator === '~=' || operator === '~:') {
      return 'regex'
    }
    return 'contains'
  }

  function normalizeQueryValue(field: QueryField, value: string) {
    if (field !== 'level') {
      return value
    }
    const level = value.trim().toLowerCase()
    const aliases: Record<string, string> = {
      verbose: 'V',
      debug: 'D',
      info: 'I',
      warn: 'W',
      warning: 'W',
      error: 'E',
      fatal: 'F',
      assert: 'F'
    }
    return aliases[level] ?? value.toUpperCase()
  }

  function matchesQueryFilters(entry: LogEntry, filters: QueryFilter[]) {
    const includeTagFilters = filters.filter((filter) => filter.field === 'tag' && !filter.negative)
    const otherFilters = filters.filter((filter) => filter.field !== 'tag' || filter.negative)

    if (includeTagFilters.length > 0 && !includeTagFilters.some((filter) => matchesQueryFilter(entry, filter))) {
      return false
    }

    return otherFilters.every((filter) => {
      const matched = matchesQueryFilter(entry, filter)
      return filter.negative ? !matched : matched
    })
  }

  function matchesQueryFilter(entry: LogEntry, filter: QueryFilter) {
    const value = queryFieldValue(entry, filter.field)
    if (filter.mode === 'regex') {
      return Boolean(filter.regex?.test(value))
    }
    const haystack = value.toLowerCase()
    const needle = filter.value.toLowerCase()
    if (filter.mode === 'equals') {
      return haystack === needle
    }
    if (filter.mode === 'startsWith') {
      return haystack.startsWith(needle)
    }
    return haystack.includes(needle)
  }

  function queryFieldValue(entry: LogEntry, field: QueryField) {
    switch (field) {
      case 'tag':
        return entry.tag || ''
      case 'message':
        return [entry.message, ...(entry.multiline ?? [])].join('\n')
      case 'package':
        return entry.packageName || ''
      case 'process':
        return [entry.packageName, entry.pid ? String(entry.pid) : ''].join('\n')
      case 'pid':
        return entry.pid ? String(entry.pid) : ''
      case 'tid':
        return entry.tid ? String(entry.tid) : ''
      case 'level':
        return entry.level || ''
      case 'raw':
        return [entry.raw, ...(entry.multiline ?? [])].join('\n')
      default:
        return entryHaystack(entry)
    }
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
    adbPathInput,
    resolvedADBPath,
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
    offlinePathInput,
    offlineLoading,
    sessionPathInput,
    sessionNameInput,
    sessionNotes,
    sessionLoading,
    currentSession,
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
    logSource,
    isOffline,
    isSession,
    isStaticSource,
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
    chooseADBExecutable,
    useADBPath,
    refreshPackages,
    setPackageMode,
    selectDevice,
    selectPackage,
    start,
    stop,
    clear,
    clearSearch,
    setTagSearchFilter,
    exportFiltered,
    exportFilteredJSONL,
    saveSession,
    openSession,
    openLogFile,
    returnToLiveMode,
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
