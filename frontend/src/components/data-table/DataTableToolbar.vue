<script setup lang="ts" generic="T">
import type { Row, Table } from '@tanstack/vue-table'
import type { ButtonVariants } from '@/components/ui/button'
import { ChevronDown, ChevronUp, Filter, XIcon } from 'lucide-vue-next'
import { computed, ref, watch } from 'vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { Input } from '@/components/ui/input'
import { cn } from '~/lib/utils'
import DataTableFacetedFilter from './DataTableFacetedFilter.vue'
import DataTableViewOptions from './DataTableViewOptions.vue'

export interface ToolbarActionButton {
  label?: string
  variant?: ButtonVariants['variant']
  icon?: Component
  click: () => void
}

export interface ToolbarBatchActionButton<T> {
  label?: string
  variant?: ButtonVariants['variant']
  icon?: Component
  click: (rows: Row<T>[]) => void
}

export interface ToolbarFilter {
  title: string
  columnName: string
  type?: 'select' | 'input'
  placeholder?: string
  options?: { label: string, value: string }[]
  onChange?: (value: string | string[] | undefined) => void
  defaultValue?: string | string[] | undefined
  priority?: 'primary' | 'secondary'
}

interface DataTableToolbarProps<T> {
  table: Table<T>
  staticActions?: ToolbarActionButton[]
  batchActions?: ToolbarBatchActionButton<T>[]
  filters?: ToolbarFilter[]
}

const props = defineProps<DataTableToolbarProps<T>>()

const isFiltered = computed(() => props.table.getState().columnFilters.length > 0)
const showAdvancedFilters = ref(false)
const userHasInteracted = ref(false)

const handleFilterChange = (columnName: string, value: string) => {
  props.table.getColumn(columnName)?.setFilterValue(value)
}

const selectedRows = computed(() => {
  return props.table.getSelectedRowModel().rows
})

const primaryFilters = computed(() =>
  props.filters?.filter(filter => !filter.priority || filter.priority === 'primary') || [],
)

const secondaryFilters = computed(() =>
  props.filters?.filter(filter => filter.priority === 'secondary') || [],
)

const hasSecondaryFilters = computed(() => secondaryFilters.value.length > 0)

// For badge display - count all filters with values (including defaults)
const activeSecondaryFiltersCount = computed(() => {
  return secondaryFilters.value.filter((filter) => {
    const column = props.table.getColumn(filter.columnName)
    const filterValue = column?.getFilterValue()
    return filterValue !== undefined && filterValue !== '' && filterValue !== null
  }).length
})

// For auto-expand - only count filters that differ from defaults
const hasActiveSecondaryFilters = computed(() => {
  return secondaryFilters.value.some((filter) => {
    const column = props.table.getColumn(filter.columnName)
    const filterValue = column?.getFilterValue()

    // Don't count default values as "active" for auto-expand
    if (filterValue === undefined || filterValue === '' || filterValue === null) {
      return false
    }

    // For array values (like multi-select), check if it's different from default
    if (Array.isArray(filterValue) && Array.isArray(filter.defaultValue)) {
      // Only consider active if values are different from default
      return JSON.stringify(filterValue.sort()) !== JSON.stringify(filter.defaultValue.sort())
    }

    // For single values, check if different from default
    if (filter.defaultValue !== undefined) {
      return filterValue !== filter.defaultValue
    }

    // If no default value set, any value means active
    return true
  })
})

watch(hasActiveSecondaryFilters, (newValue) => {
  if (newValue && userHasInteracted.value) {
    showAdvancedFilters.value = true
  }
})

watch(showAdvancedFilters, (newValue) => {
  if (newValue) {
    userHasInteracted.value = true
  }
})
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div class="flex flex-1 items-center space-x-2">
        <template v-for="filter in primaryFilters" :key="filter.columnName">
          <Input
            v-if="!filter.type || filter.type === 'input'"
            :placeholder="filter.placeholder || `Filter ${filter.title}...`"
            :model-value="(props.table.getColumn(filter.columnName)?.getFilterValue() as string) ?? ''"
            class="h-8 w-[150px] lg:w-[200px]"
            @input="handleFilterChange(filter.columnName, $event.target.value)"
          />
          <DataTableFacetedFilter
            v-else-if="filter.type === 'select' && props.table.getColumn(filter.columnName)"
            :column="props.table.getColumn(filter.columnName)"
            :title="filter.title"
            :options="filter.options || []"
            :default-value="filter.defaultValue"
            @change="filter.onChange"
          />
        </template>

        <Collapsible v-if="hasSecondaryFilters" v-model:open="showAdvancedFilters">
          <CollapsibleTrigger as-child>
            <Button
              variant="outline"
              size="sm"
              class="h-8 border-dashed"
              :class="hasActiveSecondaryFilters && 'border-primary'"
            >
              <Filter class="mr-2 h-4 w-4" />
              Advanced
              <component
                :is="showAdvancedFilters ? ChevronUp : ChevronDown"
                class="ml-2 h-4 w-4"
              />
              <Badge
                v-if="activeSecondaryFiltersCount > 0"
                variant="secondary"
                class="ml-1 h-4 w-4 rounded-full p-0 text-xs flex items-center justify-center"
              >
                {{ activeSecondaryFiltersCount }}
              </Badge>
            </Button>
          </CollapsibleTrigger>
        </Collapsible>

        <Button
          v-if="isFiltered"
          variant="ghost"
          class="h-8 px-2 lg:px-3"
          @click="props.table.resetColumnFilters()"
        >
          Reset
          <XIcon class="ml-2 h-4 w-4" />
        </Button>

        <Badge v-if="selectedRows.length > 0" variant="secondary" class="text-muted-foreground">
          {{ selectedRows.length }} {{ selectedRows.length > 1 ? 'items' : 'item' }} selected
        </Badge>

        <template v-for="action in props.batchActions" :key="action.label">
          <Button
            v-show="selectedRows.length > 0"
            :variant="action.variant || 'secondary'"
            class="h-8 px-3"
            @click="() => action.click(selectedRows)"
          >
            <component :is="action.icon" v-if="action.icon" class="mr-2 h-4 w-4" />
            {{ action.label }}
          </Button>
        </template>
      </div>

      <div class="flex items-center space-x-2">
        <template v-for="(action, index) in props.staticActions" :key="index">
          <Button
            :variant="action.variant || 'secondary'"
            class="h-8 px-3"
            @click="action.click"
          >
            <component :is="action.icon" v-if="action.icon" :class="cn('h-4 w-4', action.label ? 'mr-2' : '')" />
            {{ action.label ?? null }}
          </Button>
        </template>

        <DataTableViewOptions :table="props.table" />
      </div>
    </div>

    <Collapsible v-if="hasSecondaryFilters" v-model:open="showAdvancedFilters">
      <CollapsibleContent force-mount :class="showAdvancedFilters ? 'space-y-2' : 'hidden'">
        <div class="rounded-lg border bg-muted/20 p-4">
          <h4 class="text-sm font-medium mb-3 flex items-center">
            <Filter class="mr-2 h-4 w-4" />
            Advanced Filters
          </h4>
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            <template v-for="filter in secondaryFilters" :key="filter.columnName">
              <div class="space-y-2">
                <label class="text-xs font-medium text-muted-foreground block">
                  {{ filter.title }}
                </label>
                <Input
                  v-if="!filter.type || filter.type === 'input'"
                  :placeholder="filter.placeholder || `Filter ${filter.title}...`"
                  :model-value="(props.table.getColumn(filter.columnName)?.getFilterValue() as string) ?? ''"
                  class="h-8 w-full"
                  @input="handleFilterChange(filter.columnName, $event.target.value); userHasInteracted = true"
                />
                <DataTableFacetedFilter
                  v-else-if="filter.type === 'select' && props.table.getColumn(filter.columnName)"
                  :column="props.table.getColumn(filter.columnName)"
                  :title="filter.title"
                  :options="filter.options || []"
                  :default-value="filter.defaultValue"
                  :show-label="false"
                  @change="(value) => { filter.onChange?.(value); userHasInteracted = true }"
                />
              </div>
            </template>
          </div>
        </div>
      </CollapsibleContent>
    </Collapsible>
  </div>
</template>
