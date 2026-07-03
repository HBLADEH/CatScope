import type { AndroidDevice, InstalledPackage, LogBatch, PackagePIDState, LogStatus } from '@/types/backend'

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
  listPackages: (serial: string, mode: string) => call<InstalledPackage[]>('ListPackages', serial, mode),
  setTrackedPackage: (serial: string, packageName: string) => call<void>('SetTrackedPackage', serial, packageName),
  getPackagePIDState: () => call<PackagePIDState>('GetPackagePIDState'),
  startLogcat: (serial: string) => call<void>('StartLogcat', serial),
  stopLogcat: () => call<void>('StopLogcat'),
  exportLogs: (entries: unknown[]) => call<string>('ExportLogs', entries),
  clearLogs: () => call<void>('ClearLogs'),
  getLogBatch: (afterID: number, limit: number) => call<LogBatch>('GetLogBatch', afterID, limit),
  getLogStatus: () => call<LogStatus>('GetLogStatus')
}
