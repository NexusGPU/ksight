// Locale types
export type Locale = 'en' | 'zh'

// Theme types
export type Theme = 'light' | 'dark'

// Store interfaces
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
}

// Component props interfaces
export interface HelloWorldProps {
  // Add props if needed in the future
}

export interface DemoComponentProps {
  // Add props if needed in the future
}

// Wails Go function types
export interface WailsAPI {
  Greet: (name: string) => Promise<string>
}

// Global type augmentations
declare global {
  interface Window {
    go?: {
      main: {
        App: WailsAPI
      }
    }
  }
}
