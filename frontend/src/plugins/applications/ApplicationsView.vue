<template>
  <div class="flex flex-col h-full">
    <!-- View Header -->
    <div class="p-4 border-b">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <div class="w-5 h-5 bg-current opacity-60"></div>
          <h1 class="text-xl font-semibold">Applications</h1>
        </div>
        <div class="flex items-center gap-2">
          <!-- View switcher will go here -->
          <Button variant="outline" size="sm">
            <div class="w-4 h-4 bg-current opacity-60 mr-2"></div>
            View
          </Button>
        </div>
      </div>
    </div>
    
    <!-- Content Area -->
    <div class="flex-1 overflow-hidden">
      <!-- Filters and toolbar will go here -->
      <div class="p-4">
        <p class="text-muted-foreground mb-4">
          Applications view shows Pods grouped by app labels. This will contain:
        </p>
        <ul class="list-disc list-inside space-y-2 text-sm text-muted-foreground">
          <li>Data table with pods grouped by app labels</li>
          <li>Filters: App names, namespaces, resource types, labels, images, nodes</li>
          <li>Group by: App labels, environment, workload type, node</li>
          <li>Actions: Detail, edit YAML, exec, logs, port-forward, etc.</li>
        </ul>
        
        <!-- Placeholder for the actual pod table -->
        <div class="mt-8 p-6 border rounded-lg bg-muted/50">
          <p class="text-center text-muted-foreground">
            Pod data table will be implemented here using the existing pod.vue and table-definition-pod.ts
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Button } from '@/components/ui/button'
</script>