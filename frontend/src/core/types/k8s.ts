// Kubernetes SDK Types for window.k

// Core K8s resource interfaces
export interface K8sResource {
  apiVersion: string
  kind: string
  metadata: K8sMetadata
  spec?: Record<string, any>
  status?: Record<string, any>
}

export interface K8sMetadata {
  name: string
  namespace?: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
  resourceVersion?: string
  uid?: string
  creationTimestamp?: string
  [key: string]: any
}

// SDK Query interfaces
export interface ResourceQuery {
  namespace?: string
  labelSelector?: string
  fieldSelector?: string
  limit?: number
  continue?: string
}

export interface ListResult<T = K8sResource> {
  items: T[]
  metadata: {
    resourceVersion?: string
    continue?: string
    remainingItemCount?: number
  }
}

// SDK Operations
export interface ResourceAPI<T = K8sResource> {
  list(query?: ResourceQuery): Promise<ListResult<T>>
  get(name: string, namespace?: string): Promise<T>
  create(resource: Partial<T>): Promise<T>
  update(resource: T): Promise<T>
  patch(name: string, patch: any, namespace?: string): Promise<T>
  delete(name: string, namespace?: string): Promise<void>
  watch(callback: (event: WatchEvent<T>) => void, query?: ResourceQuery): WatchHandle
}

export interface WatchEvent<T = K8sResource> {
  type: 'ADDED' | 'MODIFIED' | 'DELETED'
  object: T
}

export interface WatchHandle {
  stop(): void
}

// High-level SDK interfaces
export interface PodAPI extends ResourceAPI<Pod> {
  exec(name: string, namespace: string, options: ExecOptions): Promise<ExecResult>
  logs(name: string, namespace: string, options?: LogOptions): Promise<string>
  portForward(name: string, namespace: string, localPort: number, podPort: number): Promise<PortForwardHandle>
}

export interface ExecOptions {
  container?: string
  command: string[]
  stdin?: boolean
  stdout?: boolean
  stderr?: boolean
  tty?: boolean
}

export interface ExecResult {
  stdout?: string
  stderr?: string
  exitCode?: number
}

export interface LogOptions {
  container?: string
  follow?: boolean
  previous?: boolean
  since?: string
  sinceTime?: Date
  timestamps?: boolean
  tailLines?: number
}

export interface PortForwardHandle {
  stop(): void
}

// Main SDK interface - this will be available as window.k
export interface KubernetesSDK {
  // Resource APIs
  pods: PodAPI
  deployments: ResourceAPI<Deployment>
  services: ResourceAPI<Service>
  nodes: ResourceAPI<Node>
  
  // Generic resource access
  resource(apiVersion: string, kind: string): ResourceAPI
  
  // Cluster operations
  cluster: {
    info(): Promise<ClusterInfo>
    contexts: ContextAPI
  }
  
  // Workflows
  workflow(name: string): WorkflowAPI
  
  // Utilities
  util: {
    conditionReady: (obj: K8sResource) => boolean
    waitForCondition: (resource: K8sResource, predicate: (obj: K8sResource) => boolean, timeout?: number) => Promise<K8sResource>
    retry: <T>(fn: () => Promise<T>, maxRetries?: number) => Promise<T>
  }
}

// Common K8s resource types
export interface Pod extends K8sResource {
  spec: {
    containers: Container[]
    nodeSelector?: Record<string, string>
    [key: string]: any
  }
  status: {
    phase: string
    conditions?: Condition[]
    [key: string]: any
  }
}

export interface Deployment extends K8sResource {
  spec: {
    replicas?: number
    selector: LabelSelector
    template: PodTemplate
    [key: string]: any
  }
  status: {
    replicas?: number
    readyReplicas?: number
    [key: string]: any
  }
}

export interface Service extends K8sResource {
  spec: {
    selector?: Record<string, string>
    ports: ServicePort[]
    type: string
    [key: string]: any
  }
}

export interface Node extends K8sResource {
  spec: {
    [key: string]: any
  }
  status: {
    conditions?: Condition[]
    addresses?: NodeAddress[]
    [key: string]: any
  }
}

// Supporting types
export interface Container {
  name: string
  image: string
  command?: string[]
  args?: string[]
  [key: string]: any
}

export interface Condition {
  type: string
  status: string
  lastTransitionTime?: string
  reason?: string
  message?: string
}

export interface LabelSelector {
  matchLabels?: Record<string, string>
  matchExpressions?: LabelSelectorRequirement[]
}

export interface LabelSelectorRequirement {
  key: string
  operator: string
  values?: string[]
}

export interface PodTemplate {
  metadata?: K8sMetadata
  spec: any
}

export interface ServicePort {
  name?: string
  port: number
  targetPort?: number | string
  protocol?: string
}

export interface NodeAddress {
  type: string
  address: string
}

export interface ClusterInfo {
  version: string
  serverVersion: any
  currentContext: string
}

export interface ContextAPI {
  list(): Promise<KubeContext[]>
  current(): Promise<KubeContext>
  switch(context: string): Promise<void>
}

export interface KubeContext {
  name: string
  cluster: string
  user: string
  namespace?: string
}

export interface WorkflowAPI {
  run(params?: Record<string, any>): Promise<WorkflowResult>
}

export interface WorkflowResult {
  success: boolean
  output: any
  error?: string
}

// Error types
export class K8sError extends Error {
  statusCode?: number
  reason?: string
  
  constructor(message: string, statusCode?: number, reason?: string) {
    super(message)
    this.name = 'K8sError'
    this.statusCode = statusCode
    this.reason = reason
  }
  
  isNotFound(): boolean {
    return this.statusCode === 404
  }
  
  isConflict(): boolean {
    return this.statusCode === 409
  }
  
  isRateLimited(): boolean {
    return this.statusCode === 429
  }
}