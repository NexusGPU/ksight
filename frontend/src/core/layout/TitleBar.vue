<template>
  <div 
    class="title-bar flex items-center bg-gray-50/80 dark:bg-gray-900/80 backdrop-blur-md border-b border-gray-200/50 dark:border-gray-700/50 h-9" 
    style="--wails-draggable: drag"
    @dblclick="handleDoubleClick"
  >
    <!-- Mac-style Window Controls (Left) -->
    <div class="flex items-center h-full pl-4 pr-3" style="--wails-draggable: no-drag">
      <div class="flex items-center space-x-2 p-1 rounded-lg bg-gray-100/50 dark:bg-gray-800/50">
        <!-- Close -->
        <button
          @click="closeWindow"
          class="mac-control w-3 h-3 rounded-full bg-red-500 hover:bg-red-600 flex items-center justify-center group relative overflow-hidden"
          title="Close"
        >
          <div class="absolute inset-0 rounded-full bg-gradient-to-br from-red-400 to-red-600 opacity-0 group-hover:opacity-100 transition-opacity duration-200"></div>
          <X :size="8" class="text-red-900 opacity-0 group-hover:opacity-100 transition-opacity duration-200 relative z-10" />
        </button>
        
        <!-- Minimize -->
        <button
          @click="minimizeWindow"
          class="mac-control w-3 h-3 rounded-full bg-yellow-500 hover:bg-yellow-600 flex items-center justify-center group relative overflow-hidden"
          title="Minimize"
        >
          <div class="absolute inset-0 rounded-full bg-gradient-to-br from-yellow-400 to-yellow-600 opacity-0 group-hover:opacity-100 transition-opacity duration-200"></div>
          <Minus :size="8" class="text-yellow-900 opacity-0 group-hover:opacity-100 transition-opacity duration-200 relative z-10" />
        </button>
        
        <!-- Maximize/Restore -->
        <button
          @click="toggleMaximize"
          class="mac-control w-3 h-3 rounded-full bg-green-500 hover:bg-green-600 flex items-center justify-center group relative overflow-hidden"
          :title="isMaximized ? 'Restore' : 'Maximize'"
        >
          <div class="absolute inset-0 rounded-full bg-gradient-to-br from-green-400 to-green-600 opacity-0 group-hover:opacity-100 transition-opacity duration-200"></div>
          <Minimize2 
            v-if="isMaximized"
            :size="8"
            class="text-green-900 opacity-0 group-hover:opacity-100 transition-opacity duration-200 relative z-10" 
          />
          <Maximize2 
            v-else
            :size="8"
            class="text-green-900 opacity-0 group-hover:opacity-100 transition-opacity duration-200 relative z-10" 
          />
        </button>
      </div>
    </div>

    <!-- Cluster Tabs Container -->
    <div class="flex-1 flex items-center h-full min-w-0">
      <ClusterTabs 
        :tabs="tabs" 
        @set-active-tab="setActiveTab"
        @close-tab="closeTab"
        @add-tab="addTab"
      />
    </div>

    <!-- Right Side Actions -->
    <div class="flex items-center h-full pr-4 space-x-2" style="--wails-draggable: no-drag">
      <!-- Theme Toggle -->
      <ThemeToggle />
      
      <!-- Settings -->
      <button
        @click="openSettings"
        class="settings-btn w-8 h-8 flex items-center justify-center rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 transition-all duration-200 group"
        title="Settings"
      >
        <Settings :size="16" class="transition-transform duration-200 group-hover:rotate-45" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { WindowMinimise, WindowMaximise, WindowUnmaximise, Quit } from '@/wailsjs/runtime/runtime'
import { X, Minus, Maximize2, Minimize2, Settings } from 'lucide-vue-next'
import ClusterTabs from '@/app/frame/ClusterTabs.vue'
import ThemeToggle from '@/app/frame/ThemeToggle.vue'
import type { Tab } from '@/app/frame/ClusterTabs.vue'

// Default icon component for tabs
const DefaultIcon = {
  template: `
    <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
    </svg>
  `
}

// State
const isMaximized = ref(false)
const tabs = ref<Tab[]>([
  {
    id: '1',
    title: 'local-cluster',
    icon: DefaultIcon,
    active: true,
    connected: true,
    pinned: false,
    route: '/dashboard'
  }
])

// Tab management methods
const setActiveTab = (tabId: string) => {
  tabs.value.forEach(tab => {
    tab.active = tab.id === tabId
  })
}

const closeTab = (tabId: string) => {
  const tabIndex = tabs.value.findIndex(tab => tab.id === tabId)
  if (tabIndex === -1) return
  
  const wasActive = tabs.value[tabIndex].active
  tabs.value.splice(tabIndex, 1)
  
  // If we closed the active tab, activate another one
  if (wasActive && tabs.value.length > 0) {
    const newActiveIndex = Math.min(tabIndex, tabs.value.length - 1)
    tabs.value[newActiveIndex].active = true
  }
}

const addTab = () => {
  const newId = Date.now().toString()
  const newTab: Tab = {
    id: newId,
    title: `new-cluster`,
    icon: DefaultIcon,
    active: false,
    connected: false,
    pinned: false,
    route: '/dashboard'
  }
  
  // Deactivate all tabs
  tabs.value.forEach(tab => tab.active = false)
  
  // Add and activate new tab
  newTab.active = true
  tabs.value.push(newTab)
}

// Settings handler
const openSettings = () => {
  // TODO: Implement settings dialog
  console.log('Open settings')
}

// Window controls
const minimizeWindow = async () => {
  try {
    await WindowMinimise()
  } catch (error) {
    console.error('WindowMinimise error:', error)
  }
}

const toggleMaximize = async () => {
  try {
    if (isMaximized.value) {
      await WindowUnmaximise()
    } else {
      await WindowMaximise()
    }
    isMaximized.value = !isMaximized.value
  } catch (error) {
    console.error('WindowMaximise/Unmaximise error:', error)
  }
}

const closeWindow = async () => {
  try {
    await Quit()
  } catch (error) {
    console.error('Quit error:', error)
  }
}

const handleDoubleClick = () => {
  toggleMaximize()
}

// Expose methods for parent components
defineExpose({
  addTab,
  setActiveTab,
  closeTab,
  tabs
})
</script>

<style scoped>
.title-bar {
  --wails-draggable: drag;
  -webkit-app-region: drag;
  user-select: none;
}

.title-bar button,
.mac-control,
.settings-btn {
  --wails-draggable: no-drag;
  -webkit-app-region: no-drag;
}

.mac-control {
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.mac-control:hover {
  transform: scale(1.15);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
}

.mac-control:active {
  transform: scale(0.95);
}

.settings-btn {
  position: relative;
}

.settings-btn::before {
  content: '';
  position: absolute;
  inset: -2px;
  border-radius: 10px;
  background: linear-gradient(45deg, transparent, rgba(59, 130, 246, 0.1), transparent);
  opacity: 0;
  transition: opacity 0.2s ease;
}

.settings-btn:hover::before {
  opacity: 1;
}
</style>
