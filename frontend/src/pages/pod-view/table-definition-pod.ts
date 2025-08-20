import type { V1Container } from '@kubernetes/client-node'
import type { ColumnDef } from '@tanstack/vue-table'
import type { ToolbarFilter } from '@/components/data-table/DataTableToolbar.vue'
import Badge from '@/components/ui/badge/Badge.vue'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { getItemAge } from '@/lib/utils'

// TODO final way should be like this, currently it is "tensor-fusion.ai/tflops-requests-containerName: 100m"
// labels:
//   tensor-fusion.ai/enabled: 'true'
// annotations:
//   tensor-fusion.ai/pool: dev-pool
//   tensor-fusion.ai/requests: tflops=100m,vram=1Gi | auto
//   tensor-fusion.ai/limits: tflops=2,vram=1Gi  | auto
//   tensor-fusion.ai/containers: python,app # default to all containers
//   tensor-fusion.ai/gpu-replicas: 3  # only set on Deployment type, not for Pod type
//   -- alternative --
//   tensor-fusion.ai/profile: abc # if profile set, all others could be removed
function resolveGPUResources(annotations: Record<string, string>) {
  const containerMap: Record<
    string,
    {
      vramRequests: string
      tflopsRequests: string
      vramLimits: string
      tflopsLimits: string
    }
  > = {}

  for (const [key, value] of Object.entries(annotations)) {
    if (
      key.startsWith('tensor-fusion.ai/tflops')
      || key.startsWith('tensor-fusion.ai/vram')
    ) {
      const resKey = key.replace('tensor-fusion.ai/', '')
      const parts = resKey.split('-')
      const containerName = parts[parts.length - 1]
      const resourceType = parts[parts.length - 3]

      if (!containerName || !resourceType)
        continue

      if (!containerMap[containerName]) {
        containerMap[containerName] = {
          vramRequests: '0Gi',
          tflopsRequests: '0',
          vramLimits: '0Gi',
          tflopsLimits: '0',
        }
      }

      if (resourceType === 'vram' && key.includes('limit')) {
        containerMap[containerName].vramLimits = value
      }
      else if (resourceType === 'vram' && key.includes('request')) {
        containerMap[containerName].vramRequests = value
      }
      else if (resourceType === 'tflops' && key.includes('limit')) {
        containerMap[containerName].tflopsLimits = value
      }
      else if (resourceType === 'tflops' && key.includes('request')) {
        containerMap[containerName].tflopsRequests = value
      }
    }
  }

  return Object.entries(containerMap).map(([name, resources]) => ({
    name,
    ...resources,
  }))
}

function renderGPUResources(
  annotations: Partial<Record<string, string>> | undefined,
) {
  if (!annotations || !annotations['tensor-fusion.ai/pool']) {
    return `-`
  }
  const resources = resolveGPUResources(annotations as Record<string, string>)
  return h(
    'div',
    { class: 'flex flex-col flex-auto gap-1 min-w-[120px]' },
    resources.map((resource) => {
      return h('div', { class: 'flex flex-col gap-2 text-sm' }, [
        h(
          'span',
          `TFlops: ${resource.tflopsRequests} - ${resource.tflopsLimits}`,
        ),
        h('span', `VRAM: ${resource.vramRequests} -${resource.vramLimits}`),
      ])
    }),
  )
}

function getPodBadgeVariant(status: string) {
  switch (status) {
    case 'Running':
      return 'success'
    case 'Pending':
      return 'warning'
    case 'Succeeded':
      return 'success'
    case 'Failed':
      return 'destructive'
    case 'Unknown':
      return 'default'
    default:
      return 'default'
  }
}

export function renderSquare(
  container: V1Container,
  row: TFPodWrapper,
) {
  const status = row.status?.containerStatuses?.find(
    c => c.name === container.name,
  )

  return h('div', {
    class: [
      'w-[0.68rem] h-[0.68rem] rounded-[4px]',
      status?.ready ? 'bg-green-500' : 'bg-amber-500',
    ],
  })
}

export function renderContainer(
  container: V1Container,
  row: TFPodWrapper,
) {
  const status = row.status?.containerStatuses?.find(
    c => c.name === container.name,
  )
  return h(
    Tooltip,
    {},
    {
      default: () => [
        h(TooltipTrigger, {}, () => renderSquare(container, row)),
        h(TooltipContent, {}, () => [
          h('p', { class: 'font-medium' }, container.name),
          h(
            'pre',
            { class: 'text-sm text-muted-foreground' },
            `${status ? JSON.stringify(status, null, 2) : 'N/A'}`,
          ),
        ]),
      ],
    },
  )
}

function renderContainers(
  containers: V1Container[],
  row: TFPodWrapper,
) {
  return h(
    'div',
    { class: 'flex flex-row flex-auto gap-1' },
    containers.map(container => renderContainer(container, row)),
  )
}

export const columns: ColumnDef<TFPodWrapper>[] = [
  {
    id: 'podName',
    accessorKey: 'metadata.name',
    header: 'Pod Name',
    cell: ({ row }) =>
      h('div', { class: 'min-w-[150px] flex flex-col' }, [
        h('span', {}, row.getValue('podName')),
        h('span', { class: 'text-xs text-muted-foreground' }, row.getValue('namespace')),
      ]),
    enableSorting: true,
    enableHiding: false,
  },
  {
    id: 'namespace',
    accessorKey: 'metadata.namespace',
    header: 'Namespace',
    cell: ({ row }) =>
      h('div', { class: 'min-w-[60px]' }, row.getValue('namespace')),
    enableSorting: true,
    enableHiding: true,
    filterFn: (row, columnId, filterValue) => {
      if (!filterValue || filterValue.length === 0 || filterValue.includes(''))
        return true
      return filterValue.includes(row.getValue(columnId))
    },
  },
  {
    id: 'metadata.labels',
    accessorKey: 'metadata.labels',
    header: 'Owner',
    cell: ({ row }) => {
      // just display owner info
      const labels = row.getValue('metadata.labels') as Record<string, string> | undefined
      if (!labels)
        return h('div', {}, '-')
      return h('div', { class: 'min-w-[100px]' }, labels['tensor-fusion.ai/owner'] ?? '-')
    },
    enableSorting: false,
    enableHiding: true,
    filterFn: (row, columnId, filterValue) => {
      if (!filterValue || filterValue === '')
        return true

      const labels = row.getValue(columnId) as Record<string, string> | undefined
      if (!labels)
        return false

      const owner = labels['tensor-fusion.ai/owner']
      if (!owner)
        return false

      return owner.toLowerCase().includes(String(filterValue).toLowerCase())
    },
  },
  {
    id: 'containers',
    accessorKey: 'spec.containers',
    header: 'Containers',
    cell: ({ row }) =>
      renderContainers(row.getValue('containers'), row.original),
    enableSorting: false,
    enableHiding: true,
  },
  {
    id: 'gpuRes',
    accessorKey: 'metadata.annotations',
    header: 'GPU Resources',
    cell: ({ row }) => renderGPUResources(row.getValue('gpuRes')),
    enableSorting: false,
    enableHiding: true,
  },
  {
    id: 'podIP',
    accessorKey: 'status.podIP',
    header: 'Pod IP',
    cell: ({ row }) => {
      const podIP = row.getValue<string>('podIP')
      const containers = row.getValue('containers') as V1Container[]

      // also show hostport if available
      const portElements = containers
        .filter(container => container.ports && container.ports.length > 0)
        .flatMap(container =>
          (container.ports || []).map(port =>
            h('span', { class: 'text-xs text-muted-foreground' }, `${port.name || 'port'}: ${port.containerPort}${port.protocol ? `/${port.protocol}` : ''}${port.hostPort ? `, ${port.hostPort}` : ''}`),
          ),
        )

      return h('div', { class: 'min-w-[150px] flex flex-col' }, [
        h('span', {}, podIP),
        ...portElements,
      ])
    },
    enableSorting: false,
    enableHiding: true,
  },
  {
    id: 'hostIP',
    accessorKey: 'status.hostIP',
    header: 'Host IP',
    cell: ({ row }) =>
      h('div', { class: 'min-w-[80px]' }, row.getValue('hostIP')),
    enableSorting: false,
    enableHiding: true,
  },
  {
    id: 'podStatus',
    accessorKey: 'status.phase',
    header: 'Pod Status',
    cell: ({ row }) =>
      h(
        Badge,
        {
          class: 'min-w-[90px] justify-center',
          variant: getPodBadgeVariant(row.getValue('podStatus')),
        },
        () => row.getValue('podStatus'),
      ),
    enableSorting: true,
    enableHiding: true,
    filterFn: (row, columnId, filterValue) => {
      if (!filterValue || filterValue.length === 0)
        return true
      return filterValue.includes(row.getValue(columnId))
    },
  },
  {
    id: 'creationTimestamp',
    accessorKey: 'metadata.creationTimestamp',
    header: 'Age',
    cell: ({ row }) =>
      h(
        'div',
        { class: 'min-w-[80px]' },
        getItemAge(row.getValue('creationTimestamp')),
      ),
    enableSorting: true,
    enableHiding: true,
  },
]

export const filters: ToolbarFilter[] = [
  {
    title: 'Pod Name',
    columnName: 'podName',
    type: 'input',
    placeholder: 'Filter by name...',
    priority: 'primary',
  },
  {
    title: 'Pod Status',
    columnName: 'podStatus',
    type: 'select',
    priority: 'primary',
    options: [
      { label: 'Running', value: 'Running' },
      { label: 'Pending', value: 'Pending' },
      { label: 'Succeeded', value: 'Succeeded' },
      { label: 'Failed', value: 'Failed' },
      { label: 'Unknown', value: 'Unknown' },
    ],
  },
  {
    title: 'Owner',
    columnName: 'metadata.labels',
    type: 'input',
    placeholder: 'Filter by owner...',
    priority: 'secondary',
  },
]
