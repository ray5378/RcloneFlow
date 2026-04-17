import { ref, onMounted, onUnmounted } from 'vue'

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

// Global WebSocket instance
let ws: WebSocket | null = null
let reconnectTimer: number | null = null
let reconnectAttempts = 0

const isConnected = ref(false)
const lastMessage = ref<WsMessage | null>(null)

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const {
    onMessage,
    onConnect,
    onDisconnect,
    reconnectInterval = 3000,
    maxReconnectAttempts = 10,
  } = options

  function connect() {
    if (ws && ws.readyState === WebSocket.OPEN) {
      return
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws`

    ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      isConnected.value = true
      reconnectAttempts = 0
      onConnect?.()
    }

    ws.onclose = () => {
      isConnected.value = false
      onDisconnect?.()
      // Auto reconnect
      if (reconnectAttempts < maxReconnectAttempts) {
        reconnectTimer = window.setTimeout(() => {
          reconnectAttempts++
          connect()
        }, reconnectInterval)
      }
    }

    ws.onerror = (err) => {
      console.error('WebSocket error:', err)
    }

    ws.onmessage = (event) => {
      try {
        const msg: WsMessage = JSON.parse(event.data)
        lastMessage.value = msg
        onMessage?.(msg)
        handleWsMessage(msg)
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e)
      }
    }
  }

  function disconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (ws) {
      ws.close()
      ws = null
    }
    isConnected.value = false
  }

  function send(msg: WsMessage) {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(msg))
    }
  }

  onMounted(() => {
    connect()
  })

  onUnmounted(() => {
    disconnect()
  })

  return {
    isConnected,
    lastMessage,
    connect,
    disconnect,
    send,
  }
}

// Subscribe to specific message types
const listeners: Map<string, Set<(data: any) => void>> = new Map()

export function onWsMessage(type_: string, callback: (data: any) => void) {
  if (!listeners.has(type_)) {
    listeners.set(type_, new Set())
  }
  listeners.get(type_)!.add(callback)

  // Return unsubscribe function
  return () => {
    listeners.get(type_)?.delete(callback)
  }
}

// Global message handler - call this to process incoming messages
export function handleWsMessage(msg: WsMessage) {
  if (msg.type && listeners.has(msg.type)) {
    listeners.get(msg.type)!.forEach((cb) => cb(msg.data))
  }
}
