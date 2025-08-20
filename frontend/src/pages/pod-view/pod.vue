<script setup lang="ts">
import type { V1ContainerStatus } from '@kubernetes/client-node'
import type { RowAction } from '@/components/data-table/DataTable.vue'
import type {
  ToolbarActionButton,
  ToolbarBatchActionButton,
  ToolbarFilter,
} from '@/components/data-table/DataTableToolbar.vue'
import yaml from 'js-yaml'
import { castArray } from 'lodash'
import { Edit, TerminalSquare, Trash, X } from 'lucide-vue-next'
import DataTable from '@/components/data-table/DataTable.vue'
import { dumpKubeYaml } from '@/lib/utils'
import ShellExecutor from '@/components/exec/index.vue'
import { usePodShell } from '@/components/exec/shell-manager.ts'
import { filters as _filters, columns, renderSquare } from './table-definition-pod.ts'

export type V1ContainerType = 'normal' | 'init' | 'ephemeral'

export type V1ContainerStatusWithType = V1ContainerStatus & {
  type: V1ContainerType
}

const { labelSelector = '', staticActions, contextNamespaces = ['default'], rowActions: customRowActions, inOperator } = defineProps<{
  labelSelector?: string
  view?: 'main' | 'lab'
  staticActions?: ToolbarActionButton[]
  contextNamespaces?: string[]
  rowActions?: RowAction<TFPodWrapper>[]
  inOperator?: boolean
}>()

const editVisible = ref(false)
const mode = ref<'Add' | 'Update'>('Add')
const editPodData = ref<string>('')
const { confirm } = useConfirm()

// Set up shell management
const {
  openPodShell,
  closeShell,
  shells,
  activeShellId,
  hasShells,
} = usePodShell()

const {
  items,
  store: podStore,
  loading,
} = useK8s('TF_POD', {
  labelSelector,
  limit: 500,
})

onMounted(() => {
  podStore.updateContext({ contextNamespaces })
})

const { items: namespaces } = useK8s('TF_NAMESPACE', { limit: 500 })

watch(namespaces, (val) => {
  podStore.updateContext({ allNamespaces: val?.map(namespace => namespace.metadata.name) })
})

const batchActions: ToolbarBatchActionButton<TFPodWrapper>[] = [
  {
    label: 'Batch Delete',
    icon: markRaw(Trash),
    variant: 'destructive',
    click: async (rows) => {
      const confirmed = await confirm({
        title: 'Delete',
        description: `Are you sure you want to delete ${rows.length} ${rows.length === 1 ? 'pod' : 'pods'}?`,
      })
      if (confirmed) {
        for (const row of rows) {
          await podStore.api.delete({
            name: row.original.getName(),
            namespace: row.original.getNs(),
          })
        }
        toast.success(`Delete ${rows.length} Pods succeeded`)
      }
    },
  },
]

const filters = computed<ToolbarFilter[]>(() => {
  return [
    ..._filters.filter(i => inOperator ? i.title !== 'Owner' : true),
    ...([
      {
        title: 'Namespace',
        columnName: 'namespace',
        type: 'select' as const,
        options: [
          { label: 'All', value: '' },
          ...namespaces.value.map(namespace => ({
            label: namespace.metadata.name,
            value: namespace.metadata.name,
          })),
        ],
        defaultValue: ['default'],
        onChange: (_value) => {
          const value = _value ? castArray(_value) : []
          podStore.updateContext({ contextNamespaces: value })
        },
      },
    ] as ToolbarFilter[]),
  ]
})

const onCellClick = (row: TFPodWrapper) => {
  mode.value = 'Update'
  editVisible.value = true
  editPodData.value = dumpKubeYaml(row, true)
}

const markContainerType = (type: V1ContainerType) => {
  return (container: V1ContainerStatus) => ({ ...container, type })
}

const getContainers = (pod: TFPodWrapper) => {
  const initContainerStatuses = pod.status?.initContainerStatuses
    ?.filter(cv => cv.state?.running)
    .map(markContainerType('init'))
  const containerStatuses = pod.status?.containerStatuses
    ?.map(markContainerType('normal'))
    .concat(
      (pod.status?.ephemeralContainerStatuses ?? []).map(
        markContainerType('ephemeral'),
      ),
    )
  return containerStatuses ?? initContainerStatuses
}

const rowActions: RowAction<TFPodWrapper>[] = [
  {
    label: 'Pod shell',
    icon: markRaw(TerminalSquare),
    disabled: (pod) => {
      return (getContainers(pod)?.length ?? 0) === 0
    },
    children: (pod) => {
      const containers = getContainers(pod)
      return (
        containers?.map((containerStatus) => {
          return {
            label: containerStatus.name,
            render: (pod) => {
              return h('div', { class: 'flex items-center gap-2' }, [
                renderSquare(containerStatus, pod),
                containerStatus.name,
              ])
            },
            click: () => {
              // Use the shell manager to open a new shell session
              openPodShell(pod, containerStatus)
            },
          }
        }) ?? []
      )
    },
  },
  ...(customRowActions ?? []),
  {
    label: inOperator ? 'View' : 'Update',
    icon: markRaw(Edit),
    click: onCellClick,
  },
  {
    label: 'Delete',
    icon: markRaw(Trash),
    click: async (row) => {
      const confirmed = await confirm({
        title: 'Delete',
        description: `Are you sure you want to delete Pod ${row.getName()}?`,
      })
      if (!confirmed) {
        return
      }
      await podStore.api.delete({
        name: row.getName(),
        namespace: row.getNs(),
      })
      toast.success(`Delete ${row.getName()} succeeded`)
    },
  },
]

async function savePod() {
  const obj = yaml.load(editPodData.value) as TFPodWrapper
  const itemId = {
    name: obj.metadata.name,
    namespace: obj.metadata.namespace ?? 'default',
  }
  if (mode.value === 'Add') {
    await podStore.api.create(itemId, obj)
    editVisible.value = false
    toast.success(`create Pod ${itemId.name}in ${itemId.namespace} succeeded`)
  }
  else if (mode.value === 'Update') {
    const obj = yaml.load(editPodData.value) as TFPodWrapper
    await podStore.api.update(itemId, obj)
    toast.success(
      `update Pod ${itemId.name} in namespace ${itemId.namespace} succeeded`,
    )
  }
}
</script>

<template>
  <div class="w-full h-[calc(100vh-140px)]">
    <ResizablePanelGroup direction="vertical" class="w-full rounded-lg">
      <ResizablePanel :default-size="40">
        <div class="flex h-full w-full rounded-t-lg overflow-auto">
          <DataTable
            :data="items"
            :columns="columns.filter(i => inOperator ? i.header !== 'GPU Resources' : true)"
            :batch-selection="!inOperator"
            :row-actions="rowActions"
            :filters="filters"
            :batch-actions="batchActions"
            :static-actions="staticActions"
            :on-cell-click="{ podName: onCellClick }"
            :loading="loading"
            :default-hidden-columns="{
              'hostIP': false,
              'status.hostIP': false,
            }"
          />
        </div>
      </ResizablePanel>
      <ResizableHandle v-if="hasShells" />
      <ResizablePanel v-if="hasShells" id="pod-shell" :default-size="60" class="mr-4">
        <div class="flex flex-col h-full border rounded-b-lg">
          <!-- Tab Bar for Shell Sessions -->
          <div v-if="hasShells" class="flex items-center border-b bg-muted/20">
            <Tabs
              v-model="activeShellId"
              class="w-full overflow-x-auto flex-nowrap"
            >
              <TabsList class="w-full justify-start">
                <TabsTrigger
                  v-for="shell in shells"
                  :key="shell.id"
                  :value="shell.id"
                  class="flex items-center gap-2 whitespace-nowrap"
                >
                  <div class="flex items-center gap-1.5 max-w-[200px]">
                    <span class="truncate">{{ shell.title }}</span>
                    <Button
                      variant="ghost"
                      size="icon"
                      class="h-5 w-5 hover:bg-destructive/10 rounded-full"
                      @click.stop="closeShell(shell.id)"
                    >
                      <X class="h-3 w-3" />
                    </Button>
                  </div>
                </TabsTrigger>
              </TabsList>
            </Tabs>
          </div>

          <!-- Content Area -->
          <div class="flex-1 overflow-hidden relative">
            <!-- Empty State -->
            <div v-if="!hasShells" class="flex h-full items-center justify-center">
              <span class="font-semibold text-muted-foreground">Select a pod and container to open shell</span>
            </div>

            <!-- Shell Sessions -->
            <div v-for="shell in shells" :key="shell.id" class="absolute inset-0" :class="{ hidden: activeShellId !== shell.id }">
              <ShellExecutor
                v-if="activeShellId === shell.id"
                :params="shell.shellParams"
              />
            </div>
          </div>
        </div>
      </ResizablePanel>
    </ResizablePanelGroup>

    <AddOrEdit
      v-model:content="editPodData"
      v-model:is-sheet-open="editVisible"
      :is-add="mode === 'Add'"
      entity-name="Pod"
      :read-only="inOperator"
      @save-content="savePod"
    />
  </div>
</template>
