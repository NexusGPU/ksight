import { defineStore } from 'pinia'
import type { Theme, Locale } from '@/types'

export const useMainStore = defineStore('main', () => {
  // State
  const count = ref<number>(0)
  const theme = ref<Theme>('light')
  const locale = ref<Locale>('en')

  // Actions
  const increment = (): void => {
    count.value++
  }

  const toggleTheme = (): void => {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
  }

  const setLocale = (newLocale: Locale): void => {
    locale.value = newLocale
  }

  // Getters
  const isLightTheme = computed<boolean>(() => theme.value === 'light')
  const isDarkTheme = computed<boolean>(() => theme.value === 'dark')
  const currentLocaleLabel = computed<string>(() => {
    return locale.value === 'en' ? 'English' : '中文'
  })

  return {
    // State
    count: readonly(count),
    theme: readonly(theme),
    locale: readonly(locale),
    
    // Getters
    isLightTheme,
    isDarkTheme,
    currentLocaleLabel,
    
    // Actions
    increment,
    toggleTheme,
    setLocale
  }
})
