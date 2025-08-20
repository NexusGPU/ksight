<template>
  <div class="h-screen flex flex-col bg-background">
    <!-- Top: Cluster Tabs -->
    <ClusterTabs />
    
    <div class="flex-1 flex overflow-hidden">
      <!-- Left: Sidebar -->
      <Sidebar class="w-12" />
      
      <!-- Main Content Area -->
      <div class="flex-1 flex flex-col overflow-hidden">
        <router-view />
      </div>
      
      <!-- Right: AI Chat (collapsible) -->
      <ChatPanel v-if="showChat" class="w-80" />
    </div>
    
    <!-- Bottom: Command Tabs -->
    <CommandTabs v-if="showCommandTabs" class="h-64" />
  </div>
</template>

<script setup lang="ts">
import ClusterTabs from './ClusterTabs.vue'
import Sidebar from './Sidebar.vue'
import ChatPanel from './ChatPanel.vue'
import CommandTabs from './CommandTabs.vue'

// Layout state
const showChat = ref(false)
const showCommandTabs = ref(false)
</script>