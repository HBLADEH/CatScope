export interface AndroidDevice {
  serial: string
  state: string
  model?: string
  brand?: string
  androidVersion?: string
  sdkVersion?: string
  abi?: string
  isEmulator?: boolean
}

export interface InstalledPackage {
  packageName: string
  label?: string
}

export interface LogEntry {
  id: number
  timestamp: string
  pid: number
  tid: number
  level: string
  tag: string
  message: string
  packageName?: string
  raw: string
  multiline?: string[]
}

export interface LogBatch {
  entries: LogEntry[]
  count: number
  discardedCount: number
  lastID: number
}

export interface LogStatus {
  running: boolean
  serial: string
  lastError?: string
  count: number
  discardedCount: number
  lastID: number
  adbPath?: string
  source?: 'live' | 'offline' | 'session'
  offlineFilePath?: string
  offlineFileName?: string
  offlineParseFailedCount?: number
  sessionFilePath?: string
  sessionName?: string
}

export interface OfflineLogFileResult {
  filePath: string
  fileName: string
  entries: LogEntry[]
  count: number
  parseFailedCount: number
  analysisResults?: AnalysisResult[]
}

export interface PackagePIDState {
  packageName: string
  pids?: number[]
  knownPids?: number[]
  lastPid?: number
}

export type AnalysisType =
  | 'java_crash'
  | 'native_crash'
  | 'anr'
  | 'jni_error'
  | 'install_error'
  | 'unknown'

export type AnalysisSeverity = 'info' | 'warning' | 'error' | 'fatal'

export interface AnalysisResult {
  id: string
  type: AnalysisType
  severity: AnalysisSeverity
  title: string
  summary: string
  packageName?: string
  pid?: number
  tid?: number
  timestamp?: string
  primaryTag?: string
  primaryMessage?: string
  exceptionType?: string
  threadName?: string
  signal?: string
  libraryName?: string
  reason?: string
  keyFrames?: string[]
  relatedEntryIds?: number[]
  rawText?: string
  suggestions?: string[]
}

export interface AIContextOptions {
  includeDeviceInfo: boolean
  includePackageInfo: boolean
  includeAnalysisSummary: boolean
  includeRelatedLogs: boolean
  includeBeforeContextLines: number
  includeAfterContextLines: number
  includeRawText: boolean
  includeSuggestions: boolean
  language: 'zh-CN' | 'en-US'
  packageFilter?: string
  levelFilter?: string[]
  searchKeyword?: string
}

export interface WorkspaceConfig {
  id: string
  workspaceName: string
  projectPath: string
  packageName: string
  lastApkPath: string
  defaultBuildTask: string
  installOptions: InstallOptions
  selectedDeviceSerial: string
  selectedLogLevel: string[]
  searchKeyword: string
  selectedPackageMode: 'thirdParty' | 'all'
  maxLogLines: number
  autoStartLogcat: boolean
  autoClearOnLaunch: boolean
  aiContextOptions: AIContextOptions
}

export interface FilterPreset {
  id: string
  name: string
  level: string[]
  packageName: string
  keyword: string
  regexEnabled: boolean
  tags: string[]
  excludeKeyword: string
  builtIn?: boolean
}

export interface AppConfig {
  activeWorkspaceId: string
  workspaces: WorkspaceConfig[]
  filterPresets: FilterPreset[]
}

export interface APKInfo {
  apkPath: string
  fileName: string
  modifiedTime: string
  size: number
}

export interface BuildResult {
  success: boolean
  projectPath: string
  task: string
  durationMillis: number
  output: string
  error?: string
  apk?: APKInfo
}

export interface InstallOptions {
  allowDowngrade: boolean
  grantPermissions: boolean
  allowTestOnly: boolean
}

export interface InstallResult {
  success: boolean
  apkPath: string
  durationMillis: number
  output: string
  error?: string
  analysisResults?: AnalysisResult[]
}

export interface LaunchResult {
  success: boolean
  packageName: string
  durationMillis: number
  output: string
  error?: string
}

export interface ProjectConfig {
  projectPath: string
  packageName: string
  lastApkPath: string
  defaultBuildTask: string
  installOptions: InstallOptions
}

export interface BuildInstallLaunchResult {
  build: BuildResult
  install: InstallResult
  launch: LaunchResult
  packageName: string
  apk?: APKInfo
  analysisResults?: AnalysisResult[]
}

export interface SessionFilters {
  level: string[]
  packageName: string
  keyword: string
  regexEnabled: boolean
  tags: string[]
  excludeKeyword: string
}

export interface SessionSaveOptions {
  name: string
  filters: SessionFilters
  aiContextOptions: AIContextOptions
  notes: string
}

export interface CatScopeSession {
  version: number
  sessionId: string
  name: string
  createdAt: string
  updatedAt: string
  sourceMode: 'live' | 'offline' | 'session'
  sourceName: string
  sourcePath: string
  workspaceId: string
  workspaceName: string
  projectPath: string
  packageName: string
  selectedDevice: string
  knownPids: number[]
  filters: SessionFilters
  aiContextOptions: AIContextOptions
  logEntries: LogEntry[]
  analysisResults: AnalysisResult[]
  notes: string
}

export interface SessionSummary {
  sessionId: string
  name: string
  filePath: string
  createdAt: string
  updatedAt: string
  sourceMode: 'live' | 'offline' | 'session'
  sourceName: string
  workspaceName: string
  packageName: string
  logCount: number
  analysisCount: number
}

export interface SessionOpenResult {
  session: CatScopeSession
  summary: SessionSummary
  entries: LogEntry[]
  analysisResults: AnalysisResult[]
}
