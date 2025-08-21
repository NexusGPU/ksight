<template>
  <div class="p-6 space-y-6">
    <div class="border rounded-lg p-4">
      <h2 class="text-lg font-semibold mb-4">Cluster Management Test</h2>
      
      <!-- Add Cluster Form -->
      <div class="space-y-4 mb-6">
        <div class="grid grid-cols-2 gap-4">
          <input
            v-model="newCluster.name"
            placeholder="Cluster Name"
            class="px-3 py-2 border rounded"
          />
          <input
            v-model="newCluster.context"
            placeholder="Context (optional)"
            class="px-3 py-2 border rounded"
          />
        </div>
        <textarea
          v-model="newCluster.kubeconfig"
          placeholder="Kubeconfig content or file path"
          rows="4"
          class="w-full px-3 py-2 border rounded"
        />
        <button
          @click="addCluster"
          :disabled="!newCluster.name || !newCluster.kubeconfig"
          class="px-4 py-2 bg-blue-500 text-white rounded disabled:opacity-50"
        >
          Add Cluster
        </button>
      </div>

      <!-- Clusters List -->
      <div class="space-y-2">
        <h3 class="font-medium">Connected Clusters:</h3>
        <div v-if="Object.keys(clusters).length === 0" class="text-gray-500">
          No clusters connected
        </div>
        <div
          v-for="cluster in clusters"
          :key="cluster.id"
          class="flex items-center justify-between p-3 border rounded"
        >
          <div>
            <div class="font-medium">{{ cluster.name }}</div>
            <div class="text-sm text-gray-600">{{ cluster.server }}</div>
            <div class="text-xs" :class="getStatusColor(cluster.status)">
              {{ cluster.status }}
            </div>
          </div>
          <div class="flex gap-2">
            <button
              @click="togglePin(cluster.id)"
              class="px-2 py-1 text-xs rounded"
              :class="cluster.isPinned ? 'bg-yellow-100 text-yellow-800' : 'bg-gray-100'"
            >
              {{ cluster.isPinned ? 'Unpin' : 'Pin' }}
            </button>
            <button
              @click="removeCluster(cluster.id)"
              class="px-2 py-1 text-xs bg-red-100 text-red-800 rounded"
            >
              Remove
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Resource Watchers Test -->
    <div class="border rounded-lg p-4">
      <h2 class="text-lg font-semibold mb-4">Resource Watchers Test</h2>
      
      <div class="grid grid-cols-2 gap-4 mb-4">
        <select v-model="selectedClusterId" class="px-3 py-2 border rounded">
          <option value="">Select Cluster</option>
          <option v-for="cluster in clusters" :key="cluster.id" :value="cluster.id">
            {{ cluster.name }}
          </option>
        </select>
        <select v-model="selectedResource" class="px-3 py-2 border rounded">
          <option value="">Select Resource</option>
          <option value="pods">Pods</option>
          <option value="deployments">Deployments</option>
          <option value="services">Services</option>
          <option value="nodes">Nodes</option>
        </select>
      </div>
      
      <div class="flex gap-2 mb-4">
        <button
          @click="addWatcher"
          :disabled="!selectedClusterId || !selectedResource"
          class="px-4 py-2 bg-green-500 text-white rounded disabled:opacity-50"
        >
          Add Watcher
        </button>
        <button
          @click="removeWatcher"
          :disabled="!selectedClusterId || !selectedResource"
          class="px-4 py-2 bg-red-500 text-white rounded disabled:opacity-50"
        >
          Remove Watcher
        </button>
      </div>

      <!-- Resource Events -->
      <div class="space-y-2">
        <h3 class="font-medium">Resource Events (last 10):</h3>
        <div v-if="resourceEvents.length === 0" class="text-gray-500">
          No resource events yet
        </div>
        <div
          v-for="event in resourceEvents.slice(-10)"
          :key="`${event.clusterId}-${event.name}-${event.timestamp}`"
          class="p-2 border rounded text-sm"
        >
          <div class="flex justify-between">
            <span class="font-medium">{{ event.type }}</span>
            <span class="text-gray-500">{{ formatTime(event.timestamp) }}</span>
          </div>
          <div>{{ event.gvr.resource }}/{{ event.name }} in {{ event.namespace || 'default' }}</div>
        </div>
      </div>
    </div>

    <!-- Debug Info -->
    <div class="border rounded-lg p-4">
      <h2 class="text-lg font-semibold mb-4">Debug Info</h2>
      <div class="text-sm space-y-2">
        <div>Clusters: {{ Object.keys(clusters).length }}</div>
        <div>Resource Events: {{ resourceEvents.length }}</div>
        <div>SDK Available: {{ !!window.k }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { k8s, type ClusterInfo, type ResourceEvent } from '@/lib/k8s-sdk'

const clusters = ref<Record<string, ClusterInfo>>({})
const resourceEvents = ref<ResourceEvent[]>([])

const newCluster = ref({
  name: '',
  context: '',
  kubeconfig: ''
})

const selectedClusterId = ref('')
const selectedResource = ref('')

// Event listeners cleanup functions
const cleanupFunctions: (() => void)[] = []

onMounted(async () => {
  // Load existing clusters
  try {
    clusters.value = await k8s.getClusters()
  } catch (error) {
    console.error('Failed to load clusters:', error)
  }

  // Setup event listeners
  cleanupFunctions.push(
    k8s.onClusterAdded((cluster) => {
      clusters.value[cluster.id] = cluster
    }),
    k8s.onClusterRemoved((clusterId) => {
      delete clusters.value[clusterId]
    }),
    k8s.onClusterUpdated((cluster) => {
      clusters.value[cluster.id] = cluster
    }),
    k8s.onResourceEvent((event) => {
      resourceEvents.value.push(event)
      // Keep only last 100 events
      if (resourceEvents.value.length > 100) {
        resourceEvents.value = resourceEvents.value.slice(-100)
      }
    })
  )
})

onUnmounted(() => {
  // Cleanup event listeners
  cleanupFunctions.forEach(cleanup => cleanup())
})

async function addCluster() {
  try {
    const clusterId = await k8s.addCluster(
      newCluster.value.name,
      newCluster.value.kubeconfig,
      newCluster.value.context
    )
    console.log('Cluster added:', clusterId)
    
    // Reset form
    newCluster.value = { name: '', context: '', kubeconfig: '' }
  } catch (error) {
    console.error('Failed to add cluster:', error)
    alert('Failed to add cluster: ' + error)
  }
}

async function removeCluster(clusterId: string) {
  try {
    await k8s.removeCluster(clusterId)
    console.log('Cluster removed:', clusterId)
  } catch (error) {
    console.error('Failed to remove cluster:', error)
    alert('Failed to remove cluster: ' + error)
  }
}

async function togglePin(clusterId: string) {
  try {
    await k8s.toggleClusterPin(clusterId)
    console.log('Cluster pin toggled:', clusterId)
  } catch (error) {
    console.error('Failed to toggle pin:', error)
  }
}

async function addWatcher() {
  if (!selectedClusterId.value || !selectedResource.value) return

  const resourceMap: Record<string, any> = {
    pods: k8s.pods,
    deployments: k8s.deployments,
    services: k8s.services,
    nodes: k8s.nodes
  }

  const resource = resourceMap[selectedResource.value]
  
  try {
    await k8s.addResourceWatcher({
      clusterId: selectedClusterId.value,
      group: resource.group,
      version: resource.version,
      resource: resource.resource
    })
    console.log('Watcher added for:', selectedResource.value)
  } catch (error) {
    console.error('Failed to add watcher:', error)
    alert('Failed to add watcher: ' + error)
  }
}

async function removeWatcher() {
  if (!selectedClusterId.value || !selectedResource.value) return

  const resourceMap: Record<string, any> = {
    pods: k8s.pods,
    deployments: k8s.deployments,
    services: k8s.services,
    nodes: k8s.nodes
  }

  const resource = resourceMap[selectedResource.value]
  
  try {
    await k8s.removeResourceWatcher({
      clusterId: selectedClusterId.value,
      group: resource.group,
      version: resource.version,
      resource: resource.resource
    })
    console.log('Watcher removed for:', selectedResource.value)
  } catch (error) {
    console.error('Failed to remove watcher:', error)
    alert('Failed to remove watcher: ' + error)
  }
}

function getStatusColor(status: string) {
  switch (status) {
    case 'connected': return 'text-green-600'
    case 'disconnected': return 'text-gray-600'
    case 'error': return 'text-red-600'
    default: return 'text-gray-600'
  }
}

function formatTime(timestamp: string) {
  return new Date(timestamp).toLocaleTimeString()
}
</script>
