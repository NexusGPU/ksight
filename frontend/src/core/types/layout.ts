// Layout and UI types for KSight

export interface ClusterTab {
  id: string
  name: string
  context: string
  active: boolean
  pinned: boolean
}

export interface SidebarItem {
  id: string
  title: string
  icon: string
  order: number
  pluginId: string
  active?: boolean
  badge?: string | number
}

export interface ViewState {
  id: string
  name: string
  pluginId: string
  filters: FilterCriteria[]
  groupBy?: GroupByOption
  orderBy?: OrderByOption
  columns: string[]
  isPinned: boolean
  isShared: boolean
  shareLink?: string
}

export interface FilterCriteria {
  field: string
  operator: 'equals' | 'contains' | 'startsWith' | 'endsWith' | 'in' | 'notIn'
  value: any
  label?: string
}

export interface GroupByOption {
  field: string
  label: string
  showCount: boolean
}

export interface OrderByOption {
  field: string
  direction: 'asc' | 'desc'
}

export interface CommandTab {
  id: string
  type: 'yaml-editor' | 'script-runner' | 'shell' | 'logs' | 'files' | 'diff'
  title: string
  active: boolean
  data: any // Type varies by tab type
}

export interface YamlEditorTab extends CommandTab {
  type: 'yaml-editor'
  data: {
    content: string
    resource?: any
    mode: 'create' | 'edit'
  }
}

export interface ShellTab extends CommandTab {
  type: 'shell'
  data: {
    podName: string
    namespace: string
    container?: string
    sessionId: string
  }
}

export interface LogsTab extends CommandTab {
  type: 'logs'
  data: {
    podName: string
    namespace: string
    container?: string
    follow: boolean
  }
}

export interface FilesTab extends CommandTab {
  type: 'files'
  data: {
    podName: string
    namespace: string
    container?: string
    basePath: string
  }
}

export interface DiffTab extends CommandTab {
  type: 'diff'
  data: {
    left: {
      title: string
      content: string
      resource?: any
    }
    right: {
      title: string
      content: string
      resource?: any
    }
  }
}