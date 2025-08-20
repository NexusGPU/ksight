<template>
  <div class="flex items-center h-full overflow-hidden">
    <!-- Active Cluster Tabs -->
    <div
      v-for="tab in tabs"
      :key="tab.id"
      :class="[
        'cluster-tab flex items-center h-full px-4 cursor-pointer relative group transition-all duration-200 ease-out',
        'border-r border-gray-200/50 dark:border-gray-700/50',
        tab.active 
          ? 'bg-white dark:bg-gray-800 text-gray-900 dark:text-white shadow-sm border-b-2 border-blue-500' 
          : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100/80 dark:hover:bg-gray-800/80 hover:text-gray-800 dark:hover:text-gray-200'
      ]"
      @click="setActiveTab(tab.id)"
      style="--wails-draggable: no-drag"
    >
      <!-- Connection Status Indicator -->
      <div 
        :class="[
          'w-2 h-2 rounded-full mr-3 flex-shrink-0 transition-all duration-200',
          tab.connected 
            ? 'bg-green-500 shadow-sm shadow-green-500/50' 
            : 'bg-gray-400 dark:bg-gray-500'
        ]"
      />
      
      <!-- Cluster Icon -->
      <Server 
        :size="16" 
        class="mr-3 flex-shrink-0 transition-colors duration-200" 
        :class="tab.active ? 'text-blue-600 dark:text-blue-400' : ''"
      />
      
      <!-- Cluster Name -->
      <span class="text-sm font-medium truncate max-w-32 transition-colors duration-200">
        {{ tab.title }}
      </span>
      
      <!-- Pinned Indicator -->
      <Pin 
        v-if="tab.pinned" 
        :size="12" 
        class="ml-2 flex-shrink-0 text-blue-500 dark:text-blue-400 transition-colors duration-200" 
      />
      
      <!-- Close Button -->
      <button
        v-if="tabs.length > 1"
        @click.stop="closeTab(tab.id)"
        class="ml-2 w-5 h-5 flex items-center justify-center rounded-full hover:bg-red-100 dark:hover:bg-red-900/30 opacity-0 group-hover:opacity-100 transition-all duration-200 hover:text-red-600 dark:hover:text-red-400"
        title="Close tab"
      >
        <X :size="12" />
      </button>
    </div>
    
    <!-- Add Cluster Tab Button -->
    <button
      @click="addTab"
      class="add-cluster-btn h-full px-4 flex items-center justify-center hover:bg-gray-100/80 dark:hover:bg-gray-800/80 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 transition-all duration-200 group"
      style="--wails-draggable: no-drag"
      title="Connect to cluster"
    >
      <Plus :size="16" class="transition-transform duration-200 group-hover:scale-110" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { Server, Pin, Plus, X } from 'lucide-vue-next'

// Tab interface
export interface Tab {
  id: string
  title: string
  icon: any
  active: boolean
  connected: boolean
  pinned: boolean
  route?: string
}

// Props
interface Props {
  tabs: Tab[]
}

defineProps<Props>()

// Emits
const emit = defineEmits<{
  setActiveTab: [tabId: string]
  closeTab: [tabId: string]
  addTab: []
}>()

// Methods
const setActiveTab = (tabId: string) => {
  emit('setActiveTab', tabId)
}

const closeTab = (tabId: string) => {
  emit('closeTab', tabId)
}

const addTab = () => {
  emit('addTab')
}
</script>

<style scoped>
.cluster-tab {
  min-width: 140px;
  max-width: 220px;
  position: relative;
}

.cluster-tab::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(59, 130, 246, 0.3), transparent);
  opacity: 0;
  transition: opacity 0.2s ease;
}

.cluster-tab:hover::before {
  opacity: 1;
}

.cluster-tab.active::before {
  opacity: 0;
}

.add-cluster-btn {
  min-width: 44px;
}
</style>
