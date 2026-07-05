import type {
  AnalysisResult,
  AIContextOptions,
  AppConfig,
  APKInfo,
  AndroidDevice,
  BuildInstallLaunchResult,
  BuildResult,
  FilterPreset,
  InstalledPackage,
  InstallOptions,
  InstallResult,
  LaunchResult,
  LogBatch,
  LogEntry,
  OfflineLogFileResult,
  PackagePIDState,
  ProjectConfig,
  SessionOpenResult,
  SessionSaveOptions,
  SessionSummary,
  WorkspaceConfig,
  LogStatus
} from '@/types/backend'

async function call<T>(method: string, ...args: unknown[]): Promise<T> {
  const app = window.go?.main?.App
  const fn = app?.[method]
  if (!fn) {
    throw new Error(`Wails backend method is unavailable: ${method}`)
  }
  return (await fn(...args)) as T
}

export const backend = {
  findADB: (configuredPath = '') => call<string>('FindADB', configuredPath),
  listDevices: () => call<AndroidDevice[]>('ListDevices'),
  getDeviceInfo: (serial: string) => call<AndroidDevice>('GetDeviceInfo', serial),
  setActiveDevice: (serial: string) => call<void>('SetActiveDevice', serial),
  listPackages: (serial: string, mode: string) => call<InstalledPackage[]>('ListPackages', serial, mode),
  setTrackedPackage: (serial: string, packageName: string) => call<void>('SetTrackedPackage', serial, packageName),
  getPackagePIDState: () => call<PackagePIDState>('GetPackagePIDState'),
  loadConfig: () => call<AppConfig>('LoadConfig'),
  saveConfig: (config: AppConfig) => call<void>('SaveConfig', config),
  resetConfig: () => call<AppConfig>('ResetConfig'),
  listWorkspaces: () => call<WorkspaceConfig[]>('ListWorkspaces'),
  saveWorkspace: (workspace: WorkspaceConfig) => call<AppConfig>('SaveWorkspace', workspace),
  deleteWorkspace: (id: string) => call<AppConfig>('DeleteWorkspace', id),
  setActiveWorkspace: (id: string) => call<AppConfig>('SetActiveWorkspace', id),
  listFilterPresets: () => call<FilterPreset[]>('ListFilterPresets'),
  saveFilterPreset: (preset: FilterPreset) => call<AppConfig>('SaveFilterPreset', preset),
  deleteFilterPreset: (id: string) => call<AppConfig>('DeleteFilterPreset', id),
  selectProjectDirectory: () => call<string>('SelectProjectDirectory'),
  getProjectConfig: () => call<ProjectConfig>('GetProjectConfig'),
  saveProjectConfig: (config: ProjectConfig) => call<void>('SaveProjectConfig', config),
  buildDebug: (projectPath: string) => call<BuildResult>('BuildDebug', projectPath),
  findLatestAPK: (projectPath: string) => call<APKInfo>('FindLatestAPK', projectPath),
  installAPK: (apkPath: string, options: InstallOptions) => call<InstallResult>('InstallAPK', apkPath, options),
  launchApp: (packageName: string) => call<LaunchResult>('LaunchApp', packageName),
  buildInstallLaunch: (config: ProjectConfig) => call<BuildInstallLaunchResult>('BuildInstallLaunch', config),
  analyzeLogs: (entries: LogEntry[]) => call<AnalysisResult[]>('AnalyzeLogs', entries),
  generateAIContext: (resultID: string, options: AIContextOptions) => call<string>('GenerateAIContext', resultID, options),
  copyAIContext: (resultID: string, options: AIContextOptions) => call<void>('CopyAIContext', resultID, options),
  exportAIContext: (resultID: string, options: AIContextOptions) => call<string>('ExportAIContext', resultID, options),
  startLogcat: (serial: string) => call<void>('StartLogcat', serial),
  stopLogcat: () => call<void>('StopLogcat'),
  exportLogs: (entries: unknown[]) => call<string>('ExportLogs', entries),
  exportLogsJSONL: (entries: unknown[]) => call<string>('ExportLogsJSONL', entries),
  saveSession: (path: string, options: SessionSaveOptions) => call<SessionSummary>('SaveSession', path, options),
  openSession: (path = '') => call<SessionOpenResult>('OpenSession', path),
  getCurrentSessionSummary: () => call<SessionSummary>('GetCurrentSessionSummary'),
  openLogFile: (path = '') => call<OfflineLogFileResult>('OpenLogFile', path),
  loadOfflineLogFile: (path: string) => call<OfflineLogFileResult>('LoadOfflineLogFile', path),
  returnToLiveMode: () => call<void>('ReturnToLiveMode'),
  clearLogs: () => call<void>('ClearLogs'),
  getLogBatch: (afterID: number, limit: number) => call<LogBatch>('GetLogBatch', afterID, limit),
  getLogStatus: () => call<LogStatus>('GetLogStatus')
}
