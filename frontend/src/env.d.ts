/// <reference types="vite/client" />

export {}

declare global {
  interface Window {
    go?: {
      main?: {
        App?: Record<string, (...args: unknown[]) => Promise<unknown>>
      }
    }
  }
}
