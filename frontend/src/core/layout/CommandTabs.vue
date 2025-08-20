<template>
  <div class="bg-background border-t flex flex-col">
    <!-- Command Tab Headers -->
    <div class="h-10 bg-muted/30 border-b flex items-center px-2 gap-1">
      <div
        v-for="tab in commandTabs"
        :key="tab.id"
        class="flex items-center gap-2 px-3 py-1 rounded-md text-sm bg-background border"
        :class="{ 'bg-primary text-primary-foreground': tab.active }"
      >
        {{ tab.title }}
        <Button variant="ghost" size="icon" class="h-4 w-4">
          <x class="w-3 h-3" />
        </Button>
      </div>
      
      <!-- Add Command Tab Button -->
      <Popover>
        <PopoverTrigger asChild>
          <Button variant="ghost" size="icon" class="h-6 w-6">
            <plus class="w-4 h-4" />
          </Button>
        </PopoverTrigger>
        <PopoverContent class="w-48" align="start">
          <div class="space-y-1">
            <Button variant="ghost" class="w-full justify-start" size="sm">
              <file-plus class="w-4 h-4 mr-2" />
              Add Resource
            </Button>
            <Button variant="ghost" class="w-full justify-start" size="sm">
              <code class="w-4 h-4 mr-2" />
              K Script
            </Button>
            <Button variant="ghost" class="w-full justify-start" size="sm">
              <terminal class="w-4 h-4 mr-2" />
              Pod Shell
            </Button>
            <Button variant="ghost" class="w-full justify-start" size="sm">
              <scroll class="w-4 h-4 mr-2" />
              Pod Logs
            </Button>
            <Button variant="ghost" class="w-full justify-start" size="sm">
              <folder class="w-4 h-4 mr-2" />
              Pod Files
            </Button>
            <Button variant="ghost" class="w-full justify-start" size="sm">
              <git-compare class="w-4 h-4 mr-2" />
              Resource Diff
            </Button>
          </div>
        </PopoverContent>
      </Popover>
    </div>
    
    <!-- Command Tab Content -->
    <div class="flex-1 p-3">
      <p class="text-sm text-muted-foreground">
        Ad-hoc command functionality will be implemented here.
        YAML editors, shell access, logs, file browser, etc.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { CommandTab } from '@/core/types'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

// Mock command tabs
const commandTabs = ref<CommandTab[]>([])
</script>