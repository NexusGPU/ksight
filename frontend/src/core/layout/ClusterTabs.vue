<template>
  <div class="h-10 bg-muted/50 border-b flex items-center px-2 gap-1">
    <!-- Cluster Connection Tabs -->
    <div class="flex items-center gap-1 flex-1">
      <div
        v-for="tab in clusterTabs"
        :key="tab.id"
        class="flex items-center gap-2 px-3 py-1 rounded-md text-sm bg-background border"
        :class="{ 'bg-primary text-primary-foreground': tab.active }"
      >
        <div class="w-2 h-2 rounded-full bg-green-500"></div>
        {{ tab.name }}
        <Button v-if="tab.pinned" variant="ghost" size="icon" class="h-4 w-4">
          <Pin class="w-3 h-3" />
        </Button>
        <Button variant="ghost" size="icon" class="h-4 w-4">
          <X class="w-3 h-3" />
        </Button>
      </div>
      
      <!-- Add Cluster Button -->
      <Button variant="ghost" size="icon" class="h-6 w-6">
        <Plus class="w-4 h-4" />
      </Button>
    </div>
    
    <!-- Settings -->
    <Button variant="ghost" size="icon" class="h-6 w-6">
      <Settings class="w-4 h-4" />
    </Button>
  </div>
</template>

<script setup lang="ts">
import type { ClusterTab } from '@/core/types'
import { Button } from '@/components/ui/button'
import { Pin, X, Plus, Settings } from 'lucide-vue-next'

// Mock data - will be replaced with actual store
const clusterTabs = ref<ClusterTab[]>([
  {
    id: '1',
    name: 'local-cluster',
    context: 'docker-desktop',
    active: true,
    pinned: false
  }
])
</script>