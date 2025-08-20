// Core application types
export type Locale = 'en' | 'zh'
export type Theme = 'light' | 'dark'

// Export plugin system types
export * from './plugin'
export * from './layout'
export * from './views'
export * from './k8s'

// Legacy store interface (to be refactored)
export interface MainStore {
  count: number
  theme: Theme
  locale: Locale
  increment: () => void
  toggleTheme: () => void
  setLocale: (newLocale: Locale) => void
}

// I18n message structure
export interface LocaleMessages {
  hello: string
  welcome: string
  language: string
  theme: string
  [key: string]: string
}

// Wails API types
export interface WailsAPI {
  Greet: (name: string) => Promise<string>
  // K8s operations will be added here
  ListPods: (namespace?: string) => Promise<any[]>
  GetPod: (name: string, namespace: string) => Promise<any>
  // More K8s operations...
}

// Global type augmentations
declare global {
  interface Window {
    go?: {
      main: {
        App: WailsAPI
      }
    }
    // K8s SDK will be available globally
    k?: import('./k8s').KubernetesSDK
  }
}
