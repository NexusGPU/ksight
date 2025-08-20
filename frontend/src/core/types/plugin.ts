import type { Component } from 'vue'
import type { RouteRecordRaw } from 'vue-router'

export interface PluginManifest {
  id: string
  name: string
  version: string
  description: string
  icon: string
  order?: number
  dependencies?: string[]
}

export interface PluginRoute {
  path: string
  name: string
  component: Component
  meta?: {
    title?: string
    icon?: string
    requiresAuth?: boolean
  }
}

export interface PluginCommand {
  id: string
  title: string
  description?: string
  icon?: string
  shortcut?: string[]
  category?: string
  handler: (args?: any) => void | Promise<void>
}

export interface PluginView {
  id: string
  title: string
  icon: string
  component: Component
  defaultVisible?: boolean
}

export interface Plugin {
  manifest: PluginManifest
  routes?: PluginRoute[]
  commands?: PluginCommand[]
  views?: PluginView[]
  onActivate?: () => void | Promise<void>
  onDeactivate?: () => void | Promise<void>
}

export interface PluginContext {
  k: PluginKubernetesSDK
  router: any
  events: EventBus
  storage: Storage
  commands: CommandRegistry
}

// Plugin System interfaces
export interface PluginRegistry {
  register(plugin: Plugin): void
  unregister(pluginId: string): void
  getPlugin(pluginId: string): Plugin | undefined
  getAllPlugins(): Plugin[]
  getActivePlugins(): Plugin[]
}

export interface CommandRegistry {
  register(command: PluginCommand): void
  unregister(commandId: string): void
  getCommand(commandId: string): PluginCommand | undefined
  getAllCommands(): PluginCommand[]
  execute(commandId: string, args?: any): Promise<void>
}

// Core system interfaces that plugins will use
export interface EventBus {
  on(event: string, handler: (...args: any[]) => void): void
  off(event: string, handler?: (...args: any[]) => void): void
  emit(event: string, ...args: any[]): void
}

export interface Storage {
  get<T = any>(key: string): T | null
  set(key: string, value: any): void
  remove(key: string): void
  clear(): void
}

// K8s SDK interface placeholder - will be defined in k8s types
export interface PluginKubernetesSDK {
  // Will be filled in k8s.ts
}