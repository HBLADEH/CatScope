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
}

export interface PackagePIDState {
  packageName: string
  pids?: number[]
  knownPids?: number[]
  lastPid?: number
}
