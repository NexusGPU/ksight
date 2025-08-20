<template>
  <div class="bg-background border-l flex flex-col">
    <div class="p-3 border-b">
      <div class="flex items-center justify-between">
        <h3 class="font-medium">AI Assistant</h3>
        <Button variant="ghost" size="icon" class="h-6 w-6">
          <x class="w-4 h-4" />
        </Button>
      </div>
    </div>
    
    <div class="flex-1 p-3">
      <p class="text-sm text-muted-foreground">
        AI chat functionality will be implemented here.
        Use k.workflow('name').run() and other K8s operations.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Button } from '@/components/ui/button'
</script>