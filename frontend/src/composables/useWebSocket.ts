import { ref, onUnmounted } from 'vue'

export interface WsMessage {
  type: string
  data?: any
}

export interface UseWebSocketOptions {
  onMessage?: (msg: WsMessage) => void
  onConnect?: () => void
  onDisconnect?: () => void
  reconnectInterval?: number
  maxReconnectAttempts?: number
}

export interface UseWebSocketReturn {
  isConnected: ReturnType<typeof ref<boolean>>
  lastMessage: ReturnType<typeof ref<WsMessage | null>>
  connect: () => void
  disconnect: () => void
  send: (msg: WsMessage) => void
  cleanup: () => void
}

const DEFAULT_CONFIG = {
  RECONNECT_INTERVAL_MS: 3000,
  MAX_RECONNECT_ATTEMPTS: 10,
  WS_PATH: '/ws',
}

let ws: WebSocket | null = null
let reconnectTimer: number | null = null
let reconnectAttempts = 0
let reconnectEnabled = true
let currentReconnectInterval = DEFAULT_CONFIG.RECONNECT_INTERVAL_MS
let currentMaxReconnectAttempts = DEFAULT_CONFIG.MAX_RECONNECT_ATTEMPTS

const isConnected = ref(false)
const lastMessage = ref<WsMessage | null>(null)

const listeners: Map<string, Set<(data: any) => void>> = new Map()
const messageSubscribers = new Set<(msg: WsMessage) => void>()
const connectSubscribers = new Set<() => void>()
const disconnectSubscribers = new Set<() => void>()

function notifyMessage(msg: WsMessage): void {
  messageSubscribers.forEach((cb) => cb(msg))
  if (msg.type && listeners.has(msg.type)) {
    listeners.get(msg.type)!.forEach((cb) => cb(msg.data))
  }
}

function buildWebSocketUrl(): string {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}${DEFAULT_CONFIG.WS_PATH}`
}

function scheduleReconnect(): void {
  if (!reconnectEnabled) return
  if (reconnectAttempts >= currentMaxReconnectAttempts) return
  if (reconnectTimer) return
  
  reconnectTimer = window.setTimeout(() => {
    reconnectTimer = null
    reconnectAttempts++
    connectGlobal()
  }, currentReconnectInterval)
}

function resetReconnectState(): void {
  reconnectAttempts = 0
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
}

function connectGlobal(): void {
  if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
    return
  }

  const wsUrl = buildWebSocketUrl()
  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    isConnected.value = true
    resetReconnectState()
    connectSubscribers.forEach((cb) => cb())
  }

  ws.onclose = () => {
    ws = null
    isConnected.value = false
    disconnectSubscribers.forEach((cb) => cb())
    scheduleReconnect()
  }

  ws.onerror = (err) => {
    console.error('[useWebSocket] WebSocket error:', err)
  }

  ws.onmessage = (event) => {
    try {
      const msg: WsMessage = JSON.parse(event.data)
      lastMessage.value = msg
      notifyMessage(msg)
    } catch (e) {
      console.warn('[useWebSocket] Failed to parse message:', event.data, e)
    }
  }
}

function disconnectGlobal(): void {
  reconnectEnabled = false
  resetReconnectState()
  
  if (ws) {
    const current = ws
    ws = null
    try {
      current.close()
    } catch (e) {
      console.warn('[useWebSocket] Error closing WebSocket:', e)
    }
  }
  
  isConnected.value = false
}

export function useWebSocket(options: UseWebSocketOptions = {}): UseWebSocketReturn {
  const {
    onMessage,
    onConnect,
    onDisconnect,
    reconnectInterval = DEFAULT_CONFIG.RECONNECT_INTERVAL_MS,
    maxReconnectAttempts = DEFAULT_CONFIG.MAX_RECONNECT_ATTEMPTS,
  } = options

  currentReconnectInterval = reconnectInterval
  currentMaxReconnectAttempts = maxReconnectAttempts
  reconnectEnabled = true

  if (onMessage) messageSubscribers.add(onMessage)
  if (onConnect) connectSubscribers.add(onConnect)
  if (onDisconnect) disconnectSubscribers.add(onDisconnect)

  function cleanup(): void {
    if (onMessage) messageSubscribers.delete(onMessage)
    if (onConnect) connectSubscribers.delete(onConnect)
    if (onDisconnect) disconnectSubscribers.delete(onDisconnect)
  }

  onUnmounted(() => {
    cleanup()
  })

  return {
    isConnected,
    lastMessage,
    connect: connectGlobal,
    disconnect: disconnectGlobal,
    send(msg: WsMessage): void {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(msg))
      }
    },
    cleanup,
  }
}

export function onWsMessage(type_: string, callback: (data: any) => void): () => void {
  if (!listeners.has(type_)) {
    listeners.set(type_, new Set())
  }
  listeners.get(type_)!.add(callback)

  return () => {
    const set = listeners.get(type_)
    set?.delete(callback)
    if (set && set.size === 0) {
      listeners.delete(type_)
    }
  }
}
