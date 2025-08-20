// Simplified confirm composable without external dialog component
export interface ConfirmDialogProps {
  title: string
  description: string
  confirmText?: string
  cancelText?: string
  variant?: 'default' | 'destructive'
}

export function useConfirm() {
  const confirm = async (props: ConfirmDialogProps): Promise<boolean> => {
    // Simple browser confirm for now - can be replaced with a proper modal later
    const message = `${props.title}\n\n${props.description}`
    return window.confirm(message)
  }

  return {
    confirm
  }
}