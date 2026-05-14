import { ref, onUnmounted } from 'vue'

// WebSocket message types
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

let ws: WebSocket | null = null
let reconnectTimer: number | null = null
let reconnectAttempts = 0
let reconnectEnabled = true
let currentReconnectInterval = 3000
let currentMaxReconnectAttempts = 10

const isConnected = ref(false)
const lastMessage = ref<WsMessage | null>(null)

const listeners: Map<string, Set<(data: any) => void>> = new Map()
const messageSubscribers = new Set<(msg: WsMessage) => void>()
const connectSubscribers = new Set<() => void>()
const disconnectSubscribers = new Set<() => void>()

function notifyMessage(msg: WsMessage) {
  messageSubscribers.forEach((cb) => cb(msg))
  if (msg.type && listeners.has(msg.type)) {
    listeners.get(msg.type)!.forEach((cb) => cb(msg.data))
  }
}

function scheduleReconnect() {
  if (!reconnectEnabled) return
  if (reconnectAttempts >= currentMaxReconnectAttempts) return
  if (reconnectTimer) return
  reconnectTimer = window.setTimeout(() => {
    reconnectTimer = null
    reconnectAttempts++
    connectGlobal()
  }, currentReconnectInterval)
}

function connectGlobal() {
  if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
    return
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws`
  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    isConnected.value = true
    reconnectAttempts = 0
    connectSubscribers.forEach((cb) => cb())
  }

  ws.onclose = () => {
    ws = null
    isConnected.value = false
    disconnectSubscribers.forEach((cb) => cb())
    scheduleReconnect()
  }

  ws.onerror = (err) => {
    console.error('WebSocket error:', err)
  }

  ws.onmessage = (event) => {
    try {
      const msg: WsMessage = JSON.parse(event.data)
      lastMessage.value = msg
      notifyMessage(msg)
    } catch (e) {
      console.error('Failed to parse WebSocket message:', e)
    }
  }
}

function disconnectGlobal() {
  reconnectEnabled = false
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  if (ws) {
    const current = ws
    ws = null
    current.close()
  }
  isConnected.value = false
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const {
    onMessage,
    onConnect,
    onDisconnect,
    reconnectInterval = 3000,
    maxReconnectAttempts = 10,
  } = options

  currentReconnectInterval = reconnectInterval
  currentMaxReconnectAttempts = maxReconnectAttempts
  reconnectEnabled = true

  if (onMessage) messageSubscribers.add(onMessage)
  if (onConnect) connectSubscribers.add(onConnect)
  if (onDisconnect) disconnectSubscribers.add(onDisconnect)

  function cleanup() {
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
    send(msg: WsMessage) {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(msg))
      }
    },
    cleanup,
  }
}

export function onWsMessage(type_: string, callback: (data: any) => void) {
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
