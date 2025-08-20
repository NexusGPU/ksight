<template>
  <div class="flex flex-col h-full">
    <!-- View Header -->
    <div class="p-4 border-b">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <server class="w-5 h-5" />
          <h1 class="text-xl font-semibold">Nodes</h1>
        </div>
        <div class="flex items-center gap-2">
          <Button variant="outline" size="sm">
            <cpu class="w-4 h-4 mr-2" />
            CPU
          </Button>
          <Button variant="outline" size="sm">
            <memory-stick class="w-4 h-4 mr-2" />
            Memory
          </Button>
          <Button variant="outline" size="sm">
            <play-circle class="w-4 h-4 mr-2" />
            Simulate Schedule
          </Button>
        </div>
      </div>
    </div>
    
    <!-- Content Area -->
    <div class="flex-1 overflow-hidden">
      <div class="p-4">
        <p class="text-muted-foreground mb-4">
          Nodes view for cluster node management. Features:
        </p>
        <ul class="list-disc list-inside space-y-2 text-sm text-muted-foreground">
          <li>Node listing with resource allocation thumbnails</li>
          <li>Filters: Node name/IP, labels, resource/device types</li>
          <li>Group by: Region/zone, machine type, architecture, tenant</li>
          <li>Actions: Cordon/uncordon, drain, node shell, resource usage</li>
          <li>Pod scheduling simulation</li>
          <li>Cost tracking per node</li>
        </ul>
        
        <!-- Placeholder for NodeList component -->
        <div class="mt-8 p-6 border rounded-lg bg-muted/50">
          <p class="text-center text-muted-foreground">
            Node data table will be implemented here using the existing NodeList.vue
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Button } from '@/components/ui/button'
</script>