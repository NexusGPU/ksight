<script setup lang="ts">
import type { RowAction } from './DataTable.vue'
import { groupBy } from 'lodash'
import { Ellipsis } from 'lucide-vue-next'
import DataTableRowAction from '~/components/app/data-table/DataTableRowAction.vue'

const props = defineProps<{
  row: any
  rowActions?: RowAction<any>[]
}>()

const noGroupActions = computed(() => {
  return (
    props.rowActions?.filter(
      action => !action.group && !action.showOutside,
    ) || []
  )
})
const groupedActions = computed(() => {
  return (
    groupBy(
      props.rowActions?.filter(action => action.group),
      'group',
    ) || {}
  )
})

const outsideActions = computed(() => {
  return props.rowActions?.filter(action => action.showOutside) || []
})
</script>

<template>
  <div class="flex items-center">
    <Tooltip v-for="action in outsideActions" :key="action.label">
      <TooltipTrigger as-child>
        <Button
          variant="ghost"
          class="flex h-8 w-8 p-0 data-[state=open]:bg-muted"
          :disabled="action.disabled?.(row.original)"
          @click="action.click?.(row.original)"
        >
          <component :is="action.icon" v-if="action.icon" class="mr-2 h-4 w-4" />
          <span class="sr-only">{{ action.label }}</span>
        </Button>
      </TooltipTrigger>
      <TooltipContent>
        {{ action.label }}
      </TooltipContent>
    </Tooltip>

    <DropdownMenu
      v-if="noGroupActions.length + Object.keys(groupedActions).length > 0"
    >
      <DropdownMenuTrigger as-child>
        <Button
          variant="ghost"
          class="flex h-8 w-8 p-0 data-[state=open]:bg-muted"
        >
          <Ellipsis class="h-4 w-4" />
          <span class="sr-only">Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" class="min-w-[200px]">
        <DataTableRowAction
          v-for="action in noGroupActions"
          :key="action.label"
          :action="action"
          :row="row.original"
        />

        <DropdownMenuSeparator v-if="Object.keys(groupedActions).length > 0" />

        <DropdownMenuSub
          v-for="group in Object.keys(groupedActions)"
          :key="group"
        >
          <DropdownMenuSubTrigger>{{ group }}</DropdownMenuSubTrigger>
          <DropdownMenuSubContent>
            <DropdownMenuItem
              v-for="action in groupedActions[group]"
              :key="action.label"
              :disabled="action.disabled?.(row.original)"
              @click="action.click?.(row.original)"
            >
              <DataTableRowAction :action="action" :row="row.original" />
            </DropdownMenuItem>
          </DropdownMenuSubContent>
        </DropdownMenuSub>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
</template>
