import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useBrowserRemoteManagement } from './useBrowserRemoteManagement'

vi.mock('../api', () => ({
  getRemoteOrder: vi.fn(),
  saveRemoteOrder: vi.fn(),
  deleteRemote: vi.fn(),
}))

vi.mock('../api/errors', () => ({
  showToast: vi.fn(),
}))

vi.mock('../i18n', () => ({
  t: (key: string) => key,
}))

describe('useBrowserRemoteManagement.ts', () => {
  const mockOptions = {
    openRemote: vi.fn(),
    loadRemotes: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  describe('initial state', () => {
    it('should initialize with empty remotes', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      expect(mgmt.remotes.value).toEqual([])
    })

    it('should initialize with empty remoteOrder from localStorage', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      expect(mgmt.remoteOrder.value).toEqual([])
    })

    it('should initialize with empty descriptions from localStorage', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      expect(mgmt.descriptions.value).toEqual({})
    })

    it('should load remoteOrder from localStorage if available', () => {
      localStorage.setItem('remoteOrder', JSON.stringify(['remote1', 'remote2']))
      const mgmt = useBrowserRemoteManagement(mockOptions)
      expect(mgmt.remoteOrder.value).toEqual(['remote1', 'remote2'])
    })

    it('should load descriptions from localStorage if available', () => {
      localStorage.setItem('remoteDescriptions', JSON.stringify({ remote1: 'desc1' }))
      const mgmt = useBrowserRemoteManagement(mockOptions)
      expect(mgmt.descriptions.value).toEqual({ remote1: 'desc1' })
    })

    it('should initialize with empty draggedRemote', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      expect(mgmt.draggedRemote.value).toBe('')
    })
  })

  describe('getOrderedRemotes', () => {
    it('should return remotes in order', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      mgmt.remotes.value = ['a', 'b', 'c']
      expect(mgmt.getOrderedRemotes()).toEqual(['a', 'b', 'c'])
    })

    it('should return remotes in custom order', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      mgmt.remotes.value = ['a', 'b', 'c', 'd']
      mgmt.remoteOrder.value = ['c', 'a']
      expect(mgmt.getOrderedRemotes()).toEqual(['c', 'a', 'b', 'd'])
    })

    it('should handle empty order', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      mgmt.remotes.value = ['a', 'b', 'c']
      mgmt.remoteOrder.value = []
      expect(mgmt.getOrderedRemotes()).toEqual(['a', 'b', 'c'])
    })

    it('should handle partial order with new remotes', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      mgmt.remotes.value = ['a', 'b', 'c', 'new-remote']
      mgmt.remoteOrder.value = ['b', 'a']
      const ordered = mgmt.getOrderedRemotes()
      expect(ordered).toContain('b')
      expect(ordered).toContain('a')
      expect(ordered).toContain('c')
      expect(ordered).toContain('new-remote')
      expect(ordered.indexOf('b')).toBeLessThan(ordered.indexOf('a'))
      expect(ordered.indexOf('a')).toBeLessThan(ordered.indexOf('c'))
      expect(ordered.indexOf('c')).toBeLessThan(ordered.indexOf('new-remote'))
    })

    it('should filter out remotes not in the list', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      mgmt.remotes.value = ['a', 'b']
      mgmt.remoteOrder.value = ['a', 'removed', 'b']
      const ordered = mgmt.getOrderedRemotes()
      expect(ordered).not.toContain('removed')
      expect(ordered).toEqual(['a', 'b'])
    })
  })

  describe('saveRemoteOrder', () => {
    it('should save remote order to API', async () => {
      const api = await import('../api')
      const mgmt = useBrowserRemoteManagement(mockOptions)
      mgmt.remotes.value = ['a', 'b', 'c']

      await mgmt.saveRemoteOrder()

      expect(api.saveRemoteOrder).toHaveBeenCalledWith(['a', 'b', 'c'])
    })

    it('should fall back to localStorage on API error', async () => {
      const api = await import('../api')
      ;(api.saveRemoteOrder as any).mockRejectedValue(new Error('API error'))

      const mgmt = useBrowserRemoteManagement(mockOptions)
      mgmt.remotes.value = ['a', 'b']

      await mgmt.saveRemoteOrder()

      expect(localStorage.getItem('remoteOrder')).toBe(JSON.stringify(['a', 'b']))
    })
  })

  describe('loadRemoteOrder', () => {
    it('should load remote order from API', async () => {
      const api = await import('../api')
      ;(api.getRemoteOrder as any).mockResolvedValue(['c', 'a', 'b'])

      const mgmt = useBrowserRemoteManagement(mockOptions)
      await mgmt.loadRemoteOrder()

      expect(mgmt.remoteOrder.value).toEqual(['c', 'a', 'b'])
      expect(localStorage.getItem('remoteOrder')).toBe(JSON.stringify(['c', 'a', 'b']))
    })

    it('should not update if API returns empty', async () => {
      const api = await import('../api')
      ;(api.getRemoteOrder as any).mockResolvedValue([])

      const mgmt = useBrowserRemoteManagement(mockOptions)
      const initialOrder = mgmt.remoteOrder.value

      await mgmt.loadRemoteOrder()

      expect(mgmt.remoteOrder.value).toEqual(initialOrder)
    })

    it('should handle API error gracefully', async () => {
      const api = await import('../api')
      ;(api.getRemoteOrder as any).mockRejectedValue(new Error('API error'))

      const mgmt = useBrowserRemoteManagement(mockOptions)
      const initialOrder = mgmt.remoteOrder.value

      await mgmt.loadRemoteOrder()

      expect(mgmt.remoteOrder.value).toEqual(initialOrder)
    })
  })

  describe('deleteRemote', () => {
    it('should return delete confirmation config', async () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      const config = await mgmt.deleteRemote('test-remote')
      expect(config.show).toBe(true)
      expect(typeof config.onConfirm).toBe('function')
    })

    it('should call API on confirm', async () => {
      const api = await import('../api')
      ;(api.deleteRemote as any).mockResolvedValue(undefined)

      const mgmt = useBrowserRemoteManagement(mockOptions)
      const config = await mgmt.deleteRemote('test-remote')

      await config.onConfirm()

      expect(api.deleteRemote).toHaveBeenCalledWith('test-remote')
    })

    it('should call loadRemotes after delete', async () => {
      const api = await import('../api')
      ;(api.deleteRemote as any).mockResolvedValue(undefined)

      const mgmt = useBrowserRemoteManagement(mockOptions)
      const config = await mgmt.deleteRemote('test-remote')

      await config.onConfirm()

      expect(mockOptions.loadRemotes).toHaveBeenCalled()
    })
  })

  describe('onDragStart', () => {
    it('should set dragged remote', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      mgmt.onDragStart('test-remote')
      expect(mgmt.draggedRemote.value).toBe('test-remote')
    })
  })

  describe('onDragOver', () => {
    it('should prevent default on drag over', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      const event = {
        preventDefault: vi.fn(),
      } as unknown as DragEvent

      mgmt.onDragOver(event, 'target-remote')
      expect(event.preventDefault).toHaveBeenCalled()
    })
  })

  describe('onDrop', () => {
    it('should move remote in list', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      mgmt.remotes.value = ['a', 'b', 'c']
      mgmt.draggedRemote.value = 'a'

      const event = {
        preventDefault: vi.fn(),
      } as unknown as DragEvent

      mgmt.onDrop(event, 'c')

      expect(mgmt.remotes.value).toContain('a')
      expect(mgmt.draggedRemote.value).toBe('')
    })

    it('should not move if dropped on itself', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      mgmt.remotes.value = ['a', 'b', 'c']
      mgmt.draggedRemote.value = 'b'

      const event = {
        preventDefault: vi.fn(),
      } as unknown as DragEvent

      mgmt.onDrop(event, 'b')

      expect(mgmt.remotes.value).toEqual(['a', 'b', 'c'])
    })

    it('should save order after drop', async () => {
      const api = await import('../api')
      ;(api.saveRemoteOrder as any).mockResolvedValue(undefined)

      const mgmt = useBrowserRemoteManagement(mockOptions)
      mgmt.remotes.value = ['a', 'b']
      mgmt.draggedRemote.value = 'a'

      const event = {
        preventDefault: vi.fn(),
      } as unknown as DragEvent

      await mgmt.onDrop(event, 'b')

      expect(api.saveRemoteOrder).toHaveBeenCalled()
    })
  })

  describe('onDragEnd', () => {
    it('should clear dragged remote', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      mgmt.draggedRemote.value = 'test-remote'

      mgmt.onDragEnd()

      expect(mgmt.draggedRemote.value).toBe('')
    })
  })

  describe('saveDesc', () => {
    it('should save description to state', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      mgmt.saveDesc('test-remote', 'test description')
      expect(mgmt.descriptions.value['test-remote']).toBe('test description')
    })

    it('should persist to localStorage', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      mgmt.saveDesc('test-remote', 'test description')
      expect(localStorage.getItem('remoteDescriptions')).toBe(
        JSON.stringify({ 'test-remote': 'test description' })
      )
    })

    it('should update existing description', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      mgmt.descriptions.value = { 'test-remote': 'old description' }

      mgmt.saveDesc('test-remote', 'new description')

      expect(mgmt.descriptions.value['test-remote']).toBe('new description')
    })

    it('should preserve other descriptions', () => {
      const mgmt = useBrowserRemoteManagement(mockOptions)
      mgmt.descriptions.value = { 'other-remote': 'other description' }

      mgmt.saveDesc('test-remote', 'test description')

      expect(mgmt.descriptions.value['other-remote']).toBe('other description')
      expect(mgmt.descriptions.value['test-remote']).toBe('test description')
    })
  })
})
