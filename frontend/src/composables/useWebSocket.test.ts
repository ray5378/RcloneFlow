import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useWebSocket, onWsMessage } from './useWebSocket'

describe('useWebSocket.ts', () => {
  let wsInstance: any

  beforeEach(() => {
    wsInstance = {
      readyState: 0,
      send: vi.fn(),
      close: vi.fn(),
      onopen: null,
      onclose: null,
      onerror: null,
      onmessage: null,
    }

    class MockWebSocketClass {
      constructor(_url: string) {
        return wsInstance
      }
    }
    vi.stubGlobal('WebSocket', MockWebSocketClass)
    vi.stubGlobal('window', {
      location: { protocol: 'http:', host: 'localhost:4200' },
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  describe('useWebSocket', () => {
    it('should return expected interface', () => {
      const ws = useWebSocket()
      expect(ws).toHaveProperty('isConnected')
      expect(ws).toHaveProperty('lastMessage')
      expect(ws).toHaveProperty('connect')
      expect(ws).toHaveProperty('disconnect')
      expect(ws).toHaveProperty('send')
      expect(ws).toHaveProperty('cleanup')
    })

    it('should parse and store messages', () => {
      const ws = useWebSocket()
      ws.connect()
      wsInstance.onopen()

      wsInstance.onmessage({ data: JSON.stringify({ type: 'test', data: { value: 42 } }) })
      expect(ws.lastMessage.value).toEqual({ type: 'test', data: { value: 42 } })
    })

    it('should ignore invalid JSON messages', () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      const ws = useWebSocket()
      ws.connect()
      wsInstance.onopen()

      wsInstance.onmessage({ data: 'not json' })
      expect(consoleSpy).toHaveBeenCalled()
      consoleSpy.mockRestore()
    })

    it('should not send when disconnected', () => {
      const ws = useWebSocket()
      ws.disconnect()
      ws.send({ type: 'ping' })
      expect(wsInstance.send).not.toHaveBeenCalled()
    })

    it('should disconnect and close websocket', () => {
      const ws = useWebSocket()
      ws.connect()
      wsInstance.onopen()

      ws.disconnect()
      expect(wsInstance.close).toHaveBeenCalled()
    })

    it('should notify message subscribers', () => {
      const onMessage = vi.fn()
      const ws = useWebSocket({ onMessage })
      ws.connect()
      wsInstance.onopen()

      wsInstance.onmessage({ data: JSON.stringify({ type: 'run_progress', data: { pct: 50 } }) })
      expect(onMessage).toHaveBeenCalledWith({ type: 'run_progress', data: { pct: 50 } })
    })

    it('should cleanup subscribers', () => {
      const onMessage = vi.fn()
      const ws = useWebSocket({ onMessage })
      ws.disconnect()
      ws.connect()
      const currentWsInstance = wsInstance
      currentWsInstance.onopen()

      ws.cleanup()

      currentWsInstance.onmessage({ data: JSON.stringify({ type: 'test' }) })
      expect(onMessage).not.toHaveBeenCalled()
    })
  })

  describe('onWsMessage', () => {
    it('should register and trigger callback for specific type', () => {
      const callback = vi.fn()
      const unsubscribe = onWsMessage('run_progress', callback)

      const ws = useWebSocket()
      ws.connect()
      wsInstance.onopen()
      wsInstance.onmessage({ data: JSON.stringify({ type: 'run_progress', data: { id: 1 } }) })

      expect(callback).toHaveBeenCalledWith({ id: 1 })
      unsubscribe()
    })

    it('should not trigger after unsubscribe', () => {
      const callback = vi.fn()
      const unsubscribe = onWsMessage('task_update', callback)
      unsubscribe()

      const ws = useWebSocket()
      ws.connect()
      wsInstance.onopen()
      wsInstance.onmessage({ data: JSON.stringify({ type: 'task_update', data: {} }) })

      expect(callback).not.toHaveBeenCalled()
    })

    it('should not trigger for different message type', () => {
      const callback = vi.fn()
      onWsMessage('type_a', callback)

      const ws = useWebSocket()
      ws.connect()
      wsInstance.onopen()
      wsInstance.onmessage({ data: JSON.stringify({ type: 'type_b', data: {} }) })

      expect(callback).not.toHaveBeenCalled()
    })
  })
})
