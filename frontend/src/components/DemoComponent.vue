<template>
  <div class="p-8 max-w-md mx-auto bg-white rounded-xl shadow-lg space-y-4">
    <div class="text-center">
      <h1 class="text-2xl font-bold text-gray-900">
        {{ $t('welcome') }}
      </h1>
      <p class="text-gray-500">{{ $t('hello') }}</p>
    </div>
    
    <!-- Counter using Pinia store -->
    <div class="flex items-center justify-center space-x-4">
      <button 
        @click="store.increment()"
        class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
      >
        Count: {{ store.count }}
      </button>
    </div>
    
    <!-- Theme toggle -->
    <div class="flex items-center justify-center space-x-2">
      <Icon :icon="themeIcon" class="w-5 h-5" />
      <button 
        @click="store.toggleTheme()"
        class="px-3 py-1 border rounded"
        :class="themeButtonClass"
      >
        {{ $t('theme') }}: {{ store.theme }}
      </button>
    </div>
    
    <!-- Language switcher -->
    <div class="flex items-center justify-center space-x-2">
      <Icon icon="lucide:globe" class="w-5 h-5" />
      <select 
        v-model="currentLocale" 
        @change="changeLanguage"
        class="px-3 py-1 border border-gray-300 rounded"
      >
        <option value="en">English</option>
        <option value="zh">中文</option>
      </select>
    </div>
    
    <!-- Icons showcase -->
    <div class="flex justify-center space-x-4">
      <Icon icon="lucide:heart" class="w-6 h-6 text-red-500" />
      <Icon icon="lucide:star" class="w-6 h-6 text-yellow-500" />
      <Icon icon="lucide:thumbs-up" class="w-6 h-6 text-green-500" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import type { Locale } from '@/types'

// Auto-imported from Pinia store
const store = useMainStore()

// Auto-imported from vue-i18n
const { locale } = useI18n()

// Reactive reference for current locale with proper typing
const currentLocale = ref<Locale>(locale.value as Locale)

// Method to change language with proper typing
const changeLanguage = (): void => {
  if (!currentLocale.value) return
  
  locale.value = currentLocale.value
  store.setLocale(currentLocale.value)
}

// Watch for store locale changes with proper typing
watchEffect((): void => {
  currentLocale.value = store.locale
})

// Computed properties for better reactivity
const themeIcon = computed<string>(() => {
  return store.isLightTheme ? 'lucide:sun' : 'lucide:moon'
})

const themeButtonClass = computed<string>(() => {
  return store.isLightTheme ? 'border-gray-300' : 'border-gray-600'
})
</script>
