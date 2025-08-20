// View and data table types

export interface DataTableColumn {
  id: string
  label: string
  sortable: boolean
  width?: number
  minWidth?: number
  visible: boolean
  accessor: string | ((item: any) => any)
  formatter?: (value: any, item: any) => string
  component?: any // Vue component for custom rendering
}

export interface DataTableRow {
  id: string
  data: any
  selected: boolean
  expanded?: boolean
  children?: DataTableRow[]
}

export interface DataTableGroup {
  id: string
  title: string
  count: number
  items: DataTableRow[]
  expanded: boolean
  metadata?: Record<string, any>
}

export interface DataTableAction {
  id: string
  label: string
  icon: string
  handler: (items: DataTableRow[]) => void
  disabled?: (items: DataTableRow[]) => boolean
  visible?: (items: DataTableRow[]) => boolean
  category?: 'primary' | 'secondary' | 'danger'
}

export interface PaginationState {
  page: number
  pageSize: number
  total: number
  hasNext: boolean
  hasPrev: boolean
}

export interface DetailDrawerState {
  open: boolean
  item: any
  width: number
}

// Plugin view interfaces
export interface PluginViewConfig {
  columns: DataTableColumn[]
  actions: DataTableAction[]
  filters: FilterOption[]
  groupByOptions: import('./layout').GroupByOption[]
  defaultView: import('./layout').ViewState
}

export interface FilterOption {
  id: string
  label: string
  type: 'text' | 'select' | 'multiselect' | 'date' | 'boolean'
  field: string
  options?: { label: string; value: any }[]
  defaultValue?: any
}

// Metrics and monitoring types
export interface MetricValue {
  timestamp: number
  value: number
}

export interface MetricSeries {
  name: string
  unit: string
  values: MetricValue[]
}

export interface ResourceMetrics {
  cpu: MetricSeries
  memory: MetricSeries
  network?: {
    in: MetricSeries
    out: MetricSeries
  }
  storage?: MetricSeries
}

// Topology view types
export interface TopologyNode {
  id: string
  type: string
  name: string
  namespace?: string
  status: 'healthy' | 'warning' | 'error' | 'unknown'
  metadata: Record<string, any>
  position?: { x: number; y: number }
}

export interface TopologyEdge {
  id: string
  source: string
  target: string
  type: 'owns' | 'controls' | 'depends'
  metadata?: Record<string, any>
}

export interface TopologyGraph {
  nodes: TopologyNode[]
  edges: TopologyEdge[]
}