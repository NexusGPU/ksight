<script setup lang="ts" generic="T">
import type {
  ColumnDef,
  ColumnFiltersState,
  SortingState,
  TableOptions,
  VisibilityState,
} from '@tanstack/vue-table'
import type { ToolbarActionButton, ToolbarBatchActionButton, ToolbarFilter } from './DataTableToolbar.vue'
import {
  FlexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useVueTable,
} from '@tanstack/vue-table'
import { Checkbox } from '~/components/ui/checkbox'
import { valueUpdater } from '~/lib/utils'
import DataTableColumnHeader from './DataTableColumnHeader.vue'
import DataTablePagination from './DataTablePagination.vue'
import DataTableRowActions from './DataTableRowActions.vue'
import DataTableToolbar from './DataTableToolbar.vue'

export interface RowAction<D> {
  label?: string
  icon?: Component
  group?: string
  showOutside?: boolean
  click?: (row: D) => void
  render?: (row: D) => Component
  children?: (row: D) => Omit<RowAction<D>, 'children'>[]
  disabled?: (row: D) => boolean
}

interface DataTableProps {
  batchSelection?: boolean
  loading?: boolean
  onRowClick?: (row: T) => void
  onCellClick?: Record<string, (row: T) => void>
  columns: ColumnDef<T, any>[]
  data: T[]
  filters?: ToolbarFilter[]
  defaultFilters?: Record<string, string | number | undefined>
  batchActions?: ToolbarBatchActionButton<T>[]
  staticActions?: ToolbarActionButton[]
  rowActions?: RowAction<T>[]
  defaultHiddenColumns?: Record<string, boolean>
  vueTableOptions?: Partial<Omit<TableOptions<T>, 'data' | 'columns'>>
}

const props = defineProps<DataTableProps>()

const sorting = ref<SortingState>([])
const columnFilters = ref<ColumnFiltersState>([])
const columnVisibility = ref<VisibilityState>(props.defaultHiddenColumns || {})
const rowSelection = ref({})

const checkCol: ColumnDef<T> = {
  id: 'select',
  header: ({ table }) =>
    h(Checkbox, {
      'checked':
        table.getIsAllPageRowsSelected()
        || (table.getIsSomePageRowsSelected() && 'indeterminate'),
      'onUpdate:checked': value => table.toggleAllPageRowsSelected(!!value),
      'ariaLabel': 'Select all',
      'class': 'translate-y-0.5',
    }),
  cell: ({ row }) =>
    h(Checkbox, {
      'checked': row.getIsSelected(),
      'onUpdate:checked': value => row.toggleSelected(!!value),
      'ariaLabel': 'Select row',
      'class': 'translate-y-0.5',
    }),
  enableSorting: false,
  enableHiding: false,
}

const actionCol: ColumnDef<T> = {
  id: 'actions',
  header: 'Actions',
  enableSorting: false,
  enableHiding: false,
  cell: ({ row }) =>
    h(DataTableRowActions, { row, rowActions: props.rowActions as any }),
}

const columnsMerged = computed(() => {
  const columns: ColumnDef<T, any>[] = props.columns.map(
    col =>
      ({
        ...col,
        header:
          typeof col.header === 'function'
            ? col.header
            : ({ column }) =>
                h(DataTableColumnHeader<T>, {
                  column,
                  title: col.header as string,
                }),
      }) as ColumnDef<T, any>,
  )

  if (props.batchSelection) {
    columns.unshift(checkCol)
  }
  if (props.rowActions && props.rowActions.length > 0) {
    columns.push(actionCol)
  }
  return columns
})

const table = useVueTable({
  get data() {
    return props.data
  },
  get columns() {
    return columnsMerged.value
  },
  enableRowSelection: true,
  onSortingChange: updaterOrValue => valueUpdater(updaterOrValue, sorting),
  onColumnFiltersChange: updaterOrValue =>
    valueUpdater(updaterOrValue, columnFilters),
  onColumnVisibilityChange: updaterOrValue =>
    valueUpdater(updaterOrValue, columnVisibility),
  onRowSelectionChange: updaterOrValue =>
    valueUpdater(updaterOrValue, rowSelection),
  getCoreRowModel: getCoreRowModel(),
  getFilteredRowModel: getFilteredRowModel(),
  getPaginationRowModel: getPaginationRowModel(),
  getSortedRowModel: getSortedRowModel(),
  getFacetedRowModel: getFacetedRowModel(),
  getFacetedUniqueValues: getFacetedUniqueValues(),
  ...props.vueTableOptions,
  state: {
    get sorting() {
      return sorting.value
    },
    get columnFilters() {
      return columnFilters.value
    },
    get columnVisibility() {
      return columnVisibility.value
    },
    get rowSelection() {
      return rowSelection.value
    },
    ...props.vueTableOptions?.state,
  },
})

onMounted(() => {
  if (props.defaultFilters) {
    Object.entries(props.defaultFilters).forEach(([columnName, value]) => {
      const column = table.getColumn(columnName)
      if (column) {
        column.setFilterValue(value)
      }
    })
  }
})
</script>

<template>
  <div class="space-y-4 w-full">
    <DataTableToolbar
      :table="table"
      :filters="filters"
      :batch-actions="batchActions"
      :static-actions="staticActions"
    />
    <div class="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow
            v-for="headerGroup in table.getHeaderGroups()"
            :key="headerGroup.id"
          >
            <TableHead v-for="header in headerGroup.headers" :key="header.id">
              <FlexRender
                v-if="!header.isPlaceholder"
                :render="header.column.columnDef.header"
                :props="header.getContext()"
              />
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <template v-if="table.getRowModel().rows?.length">
            <TableRow
              v-for="row in table.getRowModel().rows"
              :key="row.id"
              :data-state="row.getIsSelected() && 'selected'"
              :class="props.onRowClick && 'cursor-pointer'"
              @click="props.onRowClick?.(row.original)"
            >
              <TableCell
                v-for="cell in row.getVisibleCells()"
                :key="cell.id"
                :class="
                  props.onCellClick?.[cell.column.id] ? 'cursor-pointer' : ''
                "
                @click="props.onCellClick?.[cell.column.id]?.(row.original)"
              >
                <FlexRender
                  :render="cell.column.columnDef.cell"
                  :props="cell.getContext()"
                />
              </TableCell>
            </TableRow>
          </template>

          <TableRow v-else>
            <TableCell
              :colspan="
                batchSelection ? columns.length + 2 : columns.length + 1
              "
              class="h-24 text-center"
            >
              {{ loading ? "Loading..." : "No results." }}
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <DataTablePagination :table="table" />
  </div>
</template>
