import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

// Plugin views - these will be registered dynamically by the plugin system
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/plugins/applications'
  },
  {
    path: '/plugins/applications',
    name: 'applications',
    component: () => import('@/plugins/applications/ApplicationsView.vue')
  },
  {
    path: '/plugins/operations',
    name: 'operations', 
    component: () => import('@/plugins/operations/OperationsView.vue')
  },
  {
    path: '/plugins/nodes',
    name: 'nodes',
    component: () => import('@/plugins/nodes/NodesView.vue')
  },
  {
    path: '/plugins/resources',
    name: 'resources',
    component: () => import('@/plugins/resources/ResourcesView.vue')
  },
  {
    path: '/plugins/templates',
    name: 'templates',
    component: () => import('@/plugins/templates/TemplatesView.vue')
  },
  {
    path: '/plugins/workflows',
    name: 'workflows',
    component: () => import('@/plugins/workflows/WorkflowsView.vue')
  }
]

export const router = createRouter({
  history: createWebHistory(),
  routes
})

// Plugin registration function
export function registerPluginRoutes(pluginRoutes: RouteRecordRaw[]) {
  pluginRoutes.forEach(route => {
    router.addRoute(route)
  })
}

export default router