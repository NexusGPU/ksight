import { useConfirmDialog } from '@vueuse/core'
import { h, render } from 'vue'
import ConfirmDialog from '@/components/confirm/ConfirmDialog.vue'

export interface ConfirmDialogProps {
  title: string
  description: string
}

export function useConfirm() {
  const { isRevealed, reveal, confirm, cancel } = useConfirmDialog()

  let container: HTMLElement | null = null
  let cleanup: (() => void) | null = null

  const showConfirm = ({ title, description }: ConfirmDialogProps) => {
    const cleanupDialog = () => {
      if (cleanup) {
        cleanup()
        cleanup = null
      }
    }
    // Create vnode for ConfirmDialog
    const vnode = h(ConfirmDialog, {
      open: true,
      onConfirm: () => {
        confirm()
        cleanupDialog()
      },
      onCancel: () => {
        cancel()
        cleanupDialog()
      },
    }, {
      title: () => h('span', title),
      description: () => h('span', description),
    })

    // Create a container for the dialog if it doesn't exist
    if (!container) {
      container = document.createElement('div')
      document.body.appendChild(container)
    }

    // Render the dialog
    render(vnode, container)

    // Store cleanup function
    cleanup = () => {
      if (container) {
        render(null, container)
        container.remove()
        container = null
      }
    }

    // Return the promise from reveal
    return reveal()
  }

  return {
    confirm: async (props: ConfirmDialogProps): Promise<boolean> => {
      const { isCanceled } = await showConfirm(props)
      return !isCanceled
    },
    isRevealed,
  }
}
