const ERROR_1006_MESSAGE = `Unable to establish a WebSocket connection to the Pod. This is usually due to:
        
1. You do not have permission to access the Pod shell (you can ask the application owner to add permissions for you)
2. The Pod has not yet started and is still in Pending (this may be due to an incorrect Node selector, insufficient Node resources, etc.)
3. The Pod(Container) is constantly restarting

Please click the Pod logs or Pod events button to view detailed error information.`

export const ERROR_CODE_MESSAGE_MAP: Record<number, string> = {
  1006: ERROR_1006_MESSAGE,
}

export const DEFAULT_COLS = 216
export const DEFAULT_ROWS = 57
