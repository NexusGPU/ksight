<script setup lang="ts" generic="T">
import type { RowAction } from './DataTable.vue'

const props = defineProps<{ action: RowAction<T>, row: T }>()
</script>

<template>
  <DropdownMenuItem
    v-if="!action.children?.(props.row).length"
    :disabled="action.disabled?.(props.row)"
    @click="action.click?.(props.row)"
  >
    <component :is="action.icon" v-if="action.icon" class="mr-2 h-4 w-4" />
    <component :is="action.render?.(props.row)" v-if="action.render" />
    <template v-else>
      {{ action.label }}
    </template>
  </DropdownMenuItem>
  <DropdownMenuSub v-else>
    <DropdownMenuSubTrigger>
      <div class="flex items-center gap-2">
        <component :is="action.icon" v-if="action.icon" class="mr-2 h-4 w-4" />
        {{ action.label }}
      </div>
    </DropdownMenuSubTrigger>
    <DropdownMenuPortal>
      <DropdownMenuSubContent>
        <DropdownMenuItem
          v-for="(child, index) in action.children?.(props.row)"
          :key="child.label ?? index"
          :disabled="child.disabled?.(props.row)"
          @click="child.click?.(props.row)"
        >
          <component :is="child.icon" v-if="child.icon" class="mr-2 h-4 w-4" />
          <component :is="child.render?.(props.row)" v-if="child.render" />
          <template v-else>
            {{ child.label }}
          </template>
        </DropdownMenuItem>
      </DropdownMenuSubContent>
    </DropdownMenuPortal>
  </DropdownMenuSub>
</template>

<style scoped></style>
