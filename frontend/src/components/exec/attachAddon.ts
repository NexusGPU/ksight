import type { AttachAddon as IAttachApi } from '@xterm/addon-attach'
import type { IDisposable, ITerminalAddon, Terminal } from '@xterm/xterm'

interface IAttachOptions {
  bidirectional?: boolean
  keepAlive?: boolean
}

function addSocketListener<K extends keyof WebSocketEventMap>(
  socket: WebSocket,
  type: K,
  handler: (this: WebSocket, ev: WebSocketEventMap[K]) => any,
): IDisposable {
  socket.addEventListener(type, handler)
  return {
    dispose: () => {
      if (!handler) {
        // Already disposed
        return
      }
      socket.removeEventListener(type, handler)
    },
  }
}

export class AttachAddon implements ITerminalAddon, IAttachApi {
  private readonly _socket: WebSocket
  private readonly _bidirectional: boolean
  private readonly _keepAlive: boolean
  private _disposables: IDisposable[] = []

  constructor(socket: WebSocket, options?: IAttachOptions) {
    this._socket = socket
    // always set binary type to arraybuffer, we do not handle blobs
    this._socket.binaryType = 'arraybuffer'
    this._bidirectional = !(options && options.bidirectional === false)
    this._keepAlive = options?.keepAlive ?? false
  }

  public activate(terminal: Terminal): void {
    this._disposables.push(
      addSocketListener(this._socket, 'message', (ev) => {
        const data: ArrayBuffer | string = ev.data
        terminal.write(typeof data === 'string' ? data : new Uint8Array(data))
      }),
    )

    if (this._bidirectional) {
      // send Binary anyway
      this._disposables.push(terminal.onData(data => this._sendBinary(data)))
      this._disposables.push(
        terminal.onBinary(data => this._sendBinary(data)),
      )
    }
    if (this._keepAlive) {
      this._disposables.push(this._heartbeat())
    }

    this._disposables.push(
      addSocketListener(this._socket, 'close', () => this.dispose()),
    )
    this._disposables.push(
      addSocketListener(this._socket, 'error', () => this.dispose()),
    )
  }

  public dispose(): void {
    for (const d of this._disposables) {
      d.dispose()
    }
  }

  private _heartbeat(): IDisposable {
    const timer = window.setInterval(() => {
      this._sendBinary('h', 8)
    }, 20 * 1000)
    return {
      dispose() {
        window.clearInterval(timer)
      },
    }
  }

  private _sendData(data: string): void {
    if (!this._checkOpenSocket()) {
      return
    }
    this._socket.send(data)
  }

  private _sendBinary(_data: string, firstFrame = 0): void {
    if (this._socket.readyState !== 1) {
      return
    }
    const data = firstFrame + _data
    const buffer = new Uint8Array(data.length)

    for (let i = 1; i < data.length; ++i) {
      buffer[i] = data.charCodeAt(i) & 255
    }

    if (firstFrame !== 0) {
      buffer[0] = firstFrame
    }
    this._socket.send(buffer)
  }

  public sendSizeData(data: any): void {
    if (this._socket.readyState !== 1) {
      return
    }
    const { cols: width, rows: height } = data
    const sizeData = {
      width,
      height,
    }
    this._sendBinary(JSON.stringify(sizeData), 4)
  }

  private _checkOpenSocket(): boolean {
    switch (this._socket.readyState) {
      case WebSocket.OPEN:
        return true
      case WebSocket.CONNECTING:
        throw new Error('Attach addon was loaded before socket was open')
      case WebSocket.CLOSING:
        console.warn('Attach addon socket is closing')
        return false
      case WebSocket.CLOSED:
        throw new Error('Attach addon socket is closed')
      default:
        throw new Error('Unexpected socket state')
    }
  }
}
