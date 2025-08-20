import type { V1ContainerStatusWithType } from './Pods.vue'
import type { TFPodWrapper } from '~/stores/tf-app'
import { usePodShellStore } from '~/stores/pod-shell'

/**
 * Hook for pod shell management
 * Provides methods to open and close shell sessions for pods
 */
export function usePodShell() {
  const podShellStore = usePodShellStore()

  /**
   * Open a shell for a selected pod container
   * @param pod The pod object
   * @param containerStatus The container status of the selected container
   * @returns The ID of the newly created shell session
   */
  const openPodShell = (pod: TFPodWrapper, containerStatus: V1ContainerStatusWithType) => {
    // Add pod to the store and get the shell ID
    const shellId = podShellStore.addShell(pod, containerStatus)
    return shellId
  }

  /**
   * Close a specific pod shell
   * @param shellId The ID of the shell to close
   */
  const closeShell = (shellId: string) => {
    podShellStore.removeShell(shellId)
  }

  /**
   * Set the active shell
   * @param shellId The ID of the shell to activate
   */
  const setActiveShell = (shellId: string) => {
    podShellStore.setActiveShell(shellId)
  }

  return {
    openPodShell,
    closeShell,
    setActiveShell,
    shells: computed(() => podShellStore.getShells),
    activeShellId: computed({ get: () => podShellStore.getActiveShellId, set: id => podShellStore.setActiveShell(id) }),
    activeShell: computed(() => podShellStore.getActiveShell),
    getShellById: (id: string) => podShellStore.getShellById(id),
    hasShells: computed(() => podShellStore.isShellOpen),
  }
}
