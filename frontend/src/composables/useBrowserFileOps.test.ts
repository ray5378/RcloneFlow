import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useBrowserFileOps } from './useBrowserFileOps'

vi.mock('../api', () => ({
  moveFile: vi.fn(),
  moveDir: vi.fn(),
  deleteFile: vi.fn(),
  purgeDir: vi.fn(),
  testRemote: vi.fn(),
}))

vi.mock('../api/errors', () => ({
  showToast: vi.fn(),
}))

vi.mock('../i18n', () => ({
  t: (key: string) => key,
}))

describe('useBrowserFileOps.ts', () => {
  const mockOptions = {
    browserFs: vi.fn(() => 'test-fs'),
    browserPath: vi.fn(() => '/test/path'),
    refreshBrowser: vi.fn(),
    refreshUntilGone: vi.fn(),
    validatePath: vi.fn(() => true),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('initial state', () => {
    it('should initialize with hidden dialogs', () => {
      const ops = useBrowserFileOps(mockOptions)
      expect(ops.showDeleteConfirm.value).toBe(false)
      expect(ops.showRenameInput.value).toBe(false)
      expect(ops.deletingItem.value).toBe(null)
      expect(ops.renamingItem.value).toBe(null)
      expect(ops.renameInput.value).toBe('')
    })

    it('should initialize with empty testState', () => {
      const ops = useBrowserFileOps(mockOptions)
      expect(ops.testState.value).toEqual({})
    })
  })

  describe('confirmDelete and executeDelete', () => {
    it('should open delete dialog and then execute', async () => {
      const ops = useBrowserFileOps(mockOptions)
      const testItem = {
        Name: 'test.txt',
        Path: '/test/path/test.txt',
        IsDir: false,
        Size: '1024',
        ModTime: '2024-01-01T00:00:00Z',
      }
      const closeMenu = vi.fn()

      ops.confirmDelete(testItem, closeMenu)
      expect(ops.showDeleteConfirm.value).toBe(true)
      expect(ops.deletingItem.value).toEqual(testItem)
      expect(closeMenu).toHaveBeenCalled()
    })

    it('should not open delete dialog when item is null', () => {
      const ops = useBrowserFileOps(mockOptions)
      const closeMenu = vi.fn()

      ops.confirmDelete(null, closeMenu)
      expect(ops.showDeleteConfirm.value).toBe(false)
      expect(closeMenu).not.toHaveBeenCalled()
    })

    it('should execute delete for file', async () => {
      const { showToast } = await import('../api/errors')
      const api = await import('../api')
      const ops = useBrowserFileOps(mockOptions)

      const testItem = {
        Name: 'test.txt',
        Path: '/test/path/test.txt',
        IsDir: false,
        Size: '1024',
        ModTime: '2024-01-01T00:00:00Z',
      }

      ops.confirmDelete(testItem, vi.fn())
      await ops.executeDelete()

      expect(api.deleteFile).toHaveBeenCalledWith('test-fs', testItem.Path)
      expect(ops.showDeleteConfirm.value).toBe(false)
      expect(ops.deletingItem.value).toBe(null)
      expect(mockOptions.refreshBrowser).toHaveBeenCalled()
    })

    it('should execute delete for directory', async () => {
      const api = await import('../api')
      const ops = useBrowserFileOps(mockOptions)

      const testItem = {
        Name: 'test-dir',
        Path: '/test/path/test-dir',
        IsDir: true,
        Size: '0',
        ModTime: '2024-01-01T00:00:00Z',
      }

      ops.confirmDelete(testItem, vi.fn())
      await ops.executeDelete()

      expect(api.purgeDir).toHaveBeenCalledWith('test-fs', testItem.Path)
    })

    it('should not execute delete when no browserFs', async () => {
      const { showToast } = await import('../api/errors')
      const optionsWithoutFs = {
        ...mockOptions,
        browserFs: vi.fn(() => ''),
      }
      const ops = useBrowserFileOps(optionsWithoutFs)

      const testItem = {
        Name: 'test.txt',
        Path: '/test/path/test.txt',
        IsDir: false,
        Size: '1024',
        ModTime: '2024-01-01T00:00:00Z',
      }

      ops.confirmDelete(testItem, vi.fn())
      await ops.executeDelete()

      expect(showToast).toHaveBeenCalledWith(expect.stringContaining('selectStorageFirst'), 'error')
    })

    it('should not execute delete when validation fails', async () => {
      const { showToast } = await import('../api/errors')
      const optionsWithValidationFails = {
        ...mockOptions,
        validatePath: vi.fn(() => false),
      }
      const ops = useBrowserFileOps(optionsWithValidationFails)

      const testItem = {
        Name: 'test.txt',
        Path: '/test/path/test.txt',
        IsDir: false,
        Size: '1024',
        ModTime: '2024-01-01T00:00:00Z',
      }

      ops.confirmDelete(testItem, vi.fn())
      await ops.executeDelete()

      expect(showToast).not.toHaveBeenCalled()
    })

    it('should handle delete error', async () => {
      const { showToast } = await import('../api/errors')
      const api = await import('../api')
      ;(api.deleteFile as any).mockRejectedValue(new Error('Delete failed'))

      const ops = useBrowserFileOps(mockOptions)

      const testItem = {
        Name: 'test.txt',
        Path: '/test/path/test.txt',
        IsDir: false,
        Size: '1024',
        ModTime: '2024-01-01T00:00:00Z',
      }

      ops.confirmDelete(testItem, vi.fn())
      await ops.executeDelete()

      expect(showToast).toHaveBeenCalledWith(expect.stringContaining('Delete failed'), 'error')
    })
  })

  describe('startRename and confirmRename', () => {
    it('should open rename dialog and prepare input', async () => {
      const ops = useBrowserFileOps(mockOptions)
      const testItem = {
        Name: 'test.txt',
        Path: '/test/path/test.txt',
        IsDir: false,
        Size: '1024',
        ModTime: '2024-01-01T00:00:00Z',
      }
      const closeMenu = vi.fn()

      ops.startRename(testItem, closeMenu)
      expect(ops.showRenameInput.value).toBe(true)
      expect(ops.renamingItem.value).toEqual(testItem)
      expect(ops.renameInput.value).toEqual('test.txt')
      expect(closeMenu).toHaveBeenCalled()
    })

    it('should not open rename dialog when item is null', () => {
      const ops = useBrowserFileOps(mockOptions)
      const closeMenu = vi.fn()

      ops.startRename(null, closeMenu)
      expect(ops.showRenameInput.value).toBe(false)
      expect(closeMenu).not.toHaveBeenCalled()
    })

    it('should confirm rename for file', async () => {
      const api = await import('../api')
      const ops = useBrowserFileOps(mockOptions)

      const testItem = {
        Name: 'old.txt',
        Path: '/test/path/old.txt',
        IsDir: false,
        Size: '1024',
        ModTime: '2024-01-01T00:00:00Z',
      }

      ops.startRename(testItem, vi.fn())
      ops.renameInput.value = 'new.txt'
      await ops.confirmRename()

      expect(api.moveFile).toHaveBeenCalled()
      expect(ops.showRenameInput.value).toBe(false)
      expect(ops.renamingItem.value).toBe(null)
      expect(mockOptions.refreshUntilGone).toHaveBeenCalledWith(['old.txt'])
    })

    it('should confirm rename for directory', async () => {
      const api = await import('../api')
      const ops = useBrowserFileOps(mockOptions)

      const testItem = {
        Name: 'old-dir',
        Path: '/test/path/old-dir',
        IsDir: true,
        Size: '0',
        ModTime: '2024-01-01T00:00:00Z',
      }

      ops.startRename(testItem, vi.fn())
      ops.renameInput.value = 'new-dir'
      await ops.confirmRename()

      expect(api.moveDir).toHaveBeenCalled()
    })

    it('should not rename when no browserFs', async () => {
      const { showToast } = await import('../api/errors')
      const optionsWithoutFs = {
        ...mockOptions,
        browserFs: vi.fn(() => ''),
      }
      const ops = useBrowserFileOps(optionsWithoutFs)

      const testItem = {
        Name: 'test.txt',
        Path: '/test/path/test.txt',
        IsDir: false,
        Size: '1024',
        ModTime: '2024-01-01T00:00:00Z',
      }

      ops.startRename(testItem, vi.fn())
      ops.renameInput.value = 'new.txt'
      await ops.confirmRename()

      expect(showToast).toHaveBeenCalledWith(expect.stringContaining('selectStorageFirst'), 'error')
    })

    it('should not rename when renameInput is empty', async () => {
      const ops = useBrowserFileOps(mockOptions)

      const testItem = {
        Name: 'test.txt',
        Path: '/test/path/test.txt',
        IsDir: false,
        Size: '1024',
        ModTime: '2024-01-01T00:00:00Z',
      }

      ops.startRename(testItem, vi.fn())
      ops.renameInput.value = ''
      await ops.confirmRename()

      expect(mockOptions.refreshUntilGone).not.toHaveBeenCalled()
    })

    it('should handle rename error', async () => {
      const { showToast } = await import('../api/errors')
      const api = await import('../api')
      ;(api.moveFile as any).mockRejectedValue(new Error('Rename failed'))

      const ops = useBrowserFileOps(mockOptions)

      const testItem = {
        Name: 'old.txt',
        Path: '/test/path/old.txt',
        IsDir: false,
        Size: '1024',
        ModTime: '2024-01-01T00:00:00Z',
      }

      ops.startRename(testItem, vi.fn())
      ops.renameInput.value = 'new.txt'
      await ops.confirmRename()

      expect(showToast).toHaveBeenCalledWith(expect.stringContaining('Rename failed'), 'error')
    })

    it('should handle rename with subdirectory path', async () => {
      const api = await import('../api')
      const optionsWithSubPath = {
        ...mockOptions,
        browserPath: vi.fn(() => '/sub/path'),
      }
      const ops = useBrowserFileOps(optionsWithSubPath)

      const testItem = {
        Name: 'test.txt',
        Path: '/sub/path/test.txt',
        IsDir: false,
        Size: '1024',
        ModTime: '2024-01-01T00:00:00Z',
      }

      ops.startRename(testItem, vi.fn())
      ops.renameInput.value = 'renamed.txt'
      await ops.confirmRename()

      expect(api.moveFile).toHaveBeenCalledWith(
        'test-fs',
        'sub/path/test.txt',
        'test-fs',
        'sub/path/renamed.txt'
      )
    })
  })

  describe('testRemote', () => {
    it('should test remote successfully', async () => {
      const api = await import('../api')
      ;(api.testRemote as any).mockResolvedValue(undefined)

      const ops = useBrowserFileOps(mockOptions)
      await ops.testRemote('test-remote')

      expect(api.testRemote).toHaveBeenCalledWith('test-remote')
      expect(ops.testState.value['test-remote']).toBe('success')
    })

    it('should handle test remote failure', async () => {
      const api = await import('../api')
      ;(api.testRemote as any).mockRejectedValue(new Error('Test failed'))

      const ops = useBrowserFileOps(mockOptions)
      await ops.testRemote('test-remote')

      expect(ops.testState.value['test-remote']).toBe('failed')
    })

    it('should not test if already testing', async () => {
      const api = await import('../api')

      const ops = useBrowserFileOps(mockOptions)
      ops.testState.value['test-remote'] = 'testing'

      await ops.testRemote('test-remote')

      expect(api.testRemote).not.toHaveBeenCalled()
    })
  })

  describe('getTestText', () => {
    it('should return correct text for testing state', () => {
      const ops = useBrowserFileOps(mockOptions)
      ops.testState.value['test-remote'] = 'testing'

      expect(ops.getTestText('test-remote')).toBe('browserView.testing')
    })

    it('should return correct text for success state', () => {
      const ops = useBrowserFileOps(mockOptions)
      ops.testState.value['test-remote'] = 'success'

      expect(ops.getTestText('test-remote')).toBe('browserView.testSuccess')
    })

    it('should return correct text for failed state', () => {
      const ops = useBrowserFileOps(mockOptions)
      ops.testState.value['test-remote'] = 'failed'

      expect(ops.getTestText('test-remote')).toBe('browserView.testFailed')
    })

    it('should return default text for idle state', () => {
      const ops = useBrowserFileOps(mockOptions)
      expect(ops.getTestText('test-remote')).toBe('browserView.test')
    })
  })
})
