<template>
  <div class="theme-toggle-container">
    <button
      @click="toggleTheme"
      class="theme-toggle w-8 h-8 flex items-center justify-center rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 transition-all duration-200 group"
      :title="`Switch to ${nextTheme} theme`"
    >
      <Sun 
        v-if="currentTheme === 'light'" 
        :size="16" 
        class="transition-transform duration-200 group-hover:rotate-45" 
      />
      <Moon 
        v-else-if="currentTheme === 'dark'" 
        :size="16" 
        class="transition-transform duration-200 group-hover:scale-110" 
      />
      <Monitor 
        v-else 
        :size="16" 
        class="transition-transform duration-200 group-hover:scale-110" 
      />
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Sun, Moon, Monitor } from 'lucide-vue-next'

type Theme = 'light' | 'dark' | 'system'

const currentTheme = ref<Theme>('system')

const nextTheme = computed(() => {
  switch (currentTheme.value) {
    case 'light': return 'dark'
    case 'dark': return 'system'
    case 'system': return 'light'
    default: return 'light'
  }
})

const applyTheme = (theme: Theme) => {
  const html = document.documentElement
  
  if (theme === 'system') {
    // Remove explicit theme classes and let system preference take over
    html.classList.remove('light', 'dark')
    // Check system preference
    if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
      html.classList.add('dark')
    } else {
      html.classList.add('light')
    }
  } else {
    html.classList.remove('light', 'dark', 'system')
    html.classList.add(theme)
  }
}

const toggleTheme = () => {
  const themes: Theme[] = ['light', 'dark', 'system']
  const currentIndex = themes.indexOf(currentTheme.value)
  const nextIndex = (currentIndex + 1) % themes.length
  
  currentTheme.value = themes[nextIndex]
  applyTheme(currentTheme.value)
  
  // Save to localStorage
  localStorage.setItem('theme', currentTheme.value)
}

// Initialize theme on mount
onMounted(() => {
  const savedTheme = localStorage.getItem('theme') as Theme
  if (savedTheme && ['light', 'dark', 'system'].includes(savedTheme)) {
    currentTheme.value = savedTheme
  } else {
    currentTheme.value = 'system'
  }
  
  applyTheme(currentTheme.value)
  
  // Listen for system theme changes when in system mode
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  const handleSystemThemeChange = () => {
    if (currentTheme.value === 'system') {
      applyTheme('system')
    }
  }
  
  mediaQuery.addEventListener('change', handleSystemThemeChange)
  
  // Cleanup
  return () => {
    mediaQuery.removeEventListener('change', handleSystemThemeChange)
  }
})
</script>

<style scoped>
.theme-toggle {
  position: relative;
}

.theme-toggle::before {
  content: '';
  position: absolute;
  inset: -2px;
  border-radius: 10px;
  background: linear-gradient(45deg, transparent, rgba(59, 130, 246, 0.1), transparent);
  opacity: 0;
  transition: opacity 0.2s ease;
}

.theme-toggle:hover::before {
  opacity: 1;
}
</style>
