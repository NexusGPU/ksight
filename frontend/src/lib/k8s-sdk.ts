import { EventsOn, EventsOff } from '@wailsjs/runtime/runtime'

// Types for cluster management
export interface ClusterInfo {
  id: string
  name: string
  context: string
  server: string
  status: 'connected' | 'disconnected' | 'error'
  lastError?: string
  isPinned: boolean
}

export interface ResourceWatchRequest {
  clusterId: string
  group: string
  version: string
  resource: string
  namespace?: string
}

export interface GroupVersionResource {
  group: string
  version: string
  resource: string
}

export interface ResourceEvent {
  type: 'ADDED' | 'MODIFIED' | 'DELETED'
  clusterId: string
  gvr: GroupVersionResource
  namespace: string
  name: string
  object: any
  oldObject?: any
  timestamp: string
}

// Wails backend method calls
declare global {
  interface Window {
    go: {
      main: {
        App: {
          AddCluster(name: string, kubeconfig: string, context: string): Promise<string>
          RemoveCluster(clusterId: string): Promise<void>
          GetClusters(): Promise<Record<string, ClusterInfo>>
          ToggleClusterPin(clusterId: string): Promise<void>
          AddResourceWatcher(clusterId: string, group: string, version: string, resource: string, namespace: string): Promise<void>
          RemoveResourceWatcher(clusterId: string, group: string, version: string, resource: string): Promise<void>
          GetResourceTypes(clusterId: string): Promise<GroupVersionResource[]>
          LoadKubeconfigFromFile(filePath: string): Promise<string>
          SaveKubeconfigToFile(content: string, fileName: string): Promise<string>
          GetKubeconfigFiles(): Promise<string[]>
          WatchDefaultKubeconfig(): Promise<void>
          Greet(name: string): Promise<string>
        }
      }
    }
  }
}

// K8s SDK Class
export class K8sSDK {
  private eventListeners: Map<string, Set<Function>> = new Map()

  async addCluster(name: string, kubeconfig: string, context: string = ''): Promise<string> {
    return window.go.main.App.AddCluster(name, kubeconfig, context)
  }

  async removeCluster(clusterId: string): Promise<void> {
    return window.go.main.App.RemoveCluster(clusterId)
  }

  async getClusters(): Promise<Record<string, ClusterInfo>> {
    return window.go.main.App.GetClusters()
  }

  async toggleClusterPin(clusterId: string): Promise<void> {
    return window.go.main.App.ToggleClusterPin(clusterId)
  }

  async addResourceWatcher(request: ResourceWatchRequest): Promise<void> {
    return window.go.main.App.AddResourceWatcher(
      request.clusterId,
      request.group,
      request.version,
      request.resource,
      request.namespace || ''
    )
  }

  async removeResourceWatcher(request: Omit<ResourceWatchRequest, 'namespace'>): Promise<void> {
    return window.go.main.App.RemoveResourceWatcher(
      request.clusterId,
      request.group,
      request.version,
      request.resource
    )
  }

  async getResourceTypes(clusterId: string): Promise<GroupVersionResource[]> {
    return window.go.main.App.GetResourceTypes(clusterId)
  }

  async loadKubeconfigFromFile(filePath: string): Promise<string> {
    return window.go.main.App.LoadKubeconfigFromFile(filePath)
  }

  async saveKubeconfigToFile(content: string, fileName: string): Promise<string> {
    return window.go.main.App.SaveKubeconfigToFile(content, fileName)
  }

  async getKubeconfigFiles(): Promise<string[]> {
    return window.go.main.App.GetKubeconfigFiles()
  }

  async watchDefaultKubeconfig(): Promise<void> {
    return window.go.main.App.WatchDefaultKubeconfig()
  }

  onClusterAdded(callback: (cluster: ClusterInfo) => void): () => void {
    return this.addEventListener('cluster:added', callback)
  }

  onClusterRemoved(callback: (clusterId: string) => void): () => void {
    return this.addEventListener('cluster:removed', callback)
  }

  onClusterUpdated(callback: (cluster: ClusterInfo) => void): () => void {
    return this.addEventListener('cluster:updated', callback)
  }

  onResourceEvent(callback: (event: ResourceEvent) => void): () => void {
    return this.addEventListener('resource:event', callback)
  }

  private addEventListener(eventName: string, callback: Function): () => void {
    if (!this.eventListeners.has(eventName)) {
      this.eventListeners.set(eventName, new Set())
      EventsOn(eventName, (data: any) => {
        const listeners = this.eventListeners.get(eventName)
        if (listeners) {
          listeners.forEach(listener => listener(data))
        }
      })
    }

    const listeners = this.eventListeners.get(eventName)!
    listeners.add(callback)

    return () => {
      listeners.delete(callback)
      if (listeners.size === 0) {
        this.eventListeners.delete(eventName)
        EventsOff(eventName)
      }
    }
  }

  get pods() {
    return { group: '', version: 'v1', resource: 'pods' }
  }

  get deployments() {
    return { group: 'apps', version: 'v1', resource: 'deployments' }
  }

  get services() {
    return { group: '', version: 'v1', resource: 'services' }
  }

  get nodes() {
    return { group: '', version: 'v1', resource: 'nodes' }
  }
}

// Global SDK instance
export const k8s = new K8sSDK()

// Make it available globally as 'k' for operations
declare global {
  interface Window {
    k: K8sSDK
  }
}

if (typeof window !== 'undefined') {
  window.k = k8s
}
