import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useBrowserClipboard } from './useBrowserClipboard'
import type { FileItem } from '../types'

vi.mock('../api', () => ({
  copyFile: vi.fn(),
  copyDir: vi.fn(),
  moveFile: vi.fn(),
  moveDir: vi.fn()
}))

vi.mock('../api/errors', () => ({
  showToast: vi.fn()
}))

vi.mock('../i18n', () => ({
  t: (key: string) => key
}))

describe('useBrowserClipboard.ts', () => {
  const mockOptions = {
    browserFs: vi.fn(() => 'test-fs'),
    browserPath: vi.fn(() => '/test/path'),
    refreshBrowser: vi.fn(),
    refreshUntilGone: vi.fn()
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('initial state', () => {
    it('should initialize with empty clipboard', () => {
      const clipboard = useBrowserClipboard(mockOptions)
      expect(clipboard.clipboardItem.value).toBeNull()
      expect(clipboard.clipboardAction.value).toBeNull()
    })
  })

  describe('copyItem', () => {
    it('should set clipboard state for copy action', () => {
      const clipboard = useBrowserClipboard(mockOptions)
      const testItem: FileItem = {
        Name: 'test.txt',
        Path: '/path/to/test.txt',
        IsDir: false,
        Size: 1024,
        ModTime: '2024-01-01T00:00:00Z'
      }

      clipboard.copyItem(testItem)

      expect(clipboard.clipboardItem.value).toEqual(testItem)
      expect(clipboard.clipboardAction.value).toBe('copy')
      expect(clipboard.clipboardSrcFs.value).toBe('test-fs')
    })
  })

  describe('moveItem', () => {
    it('should set clipboard state for move action', () => {
      const clipboard = useBrowserClipboard(mockOptions)
      const testItem: FileItem = {
        Name: 'test.txt',
        Path: '/path/to/test.txt',
        IsDir: false,
        Size: 1024,
        ModTime: '2024-01-01T00:00:00Z'
      }

      clipboard.moveItem(testItem)

      expect(clipboard.clipboardItem.value).toEqual(testItem)
      expect(clipboard.clipboardAction.value).toBe('move')
      expect(clipboard.clipboardSrcFs.value).toBe('test-fs')
    })
  })

  describe('validatePath', () => {
    it('should validate normal path successfully', () => {
      const clipboard = useBrowserClipboard(mockOptions)
      expect(clipboard.validatePath('normal/path/file.txt')).toBe(true)
    })

    it('should detect path traversal with ..', () => {
      const clipboard = useBrowserClipboard(mockOptions)
      expect(clipboard.validatePath('../malicious/path')).toBe(false)
    })

    it('should detect path traversal with ./', () => {
      const clipboard = useBrowserClipboard(mockOptions)
      expect(clipboard.validatePath('./malicious/path')).toBe(false)
    })

    it('should detect just . as malicious', () => {
      const clipboard = useBrowserClipboard(mockOptions)
      expect(clipboard.validatePath('.')).toBe(false)
    })

    it('should detect URL encoded path traversal', () => {
      const clipboard = useBrowserClipboard(mockOptions)
      expect(clipboard.validatePath('%2e%2e/path')).toBe(false)
    })

    it('should detect partial URL encoded path traversal', () => {
      const clipboard = useBrowserClipboard(mockOptions)
      expect(clipboard.validatePath('%2e./path')).toBe(false)
    })

    it('should handle backslashes correctly', () => {
      const clipboard = useBrowserClipboard(mockOptions)
      expect(clipboard.validatePath('\\..\\malicious')).toBe(false)
    })
  })

  describe('pasteItem', () => {
    it('should not paste if clipboard is empty', async () => {
      const clipboard = useBrowserClipboard(mockOptions)
      const { showToast } = await import('../api/errors')

      await clipboard.pasteItem()

      expect(showToast).toHaveBeenCalledWith('browserView.clipboardEmpty', 'error')
    })

    it('should not paste if no browserFs is selected', async () => {
      const clipboard = useBrowserClipboard({
        ...mockOptions,
        browserFs: vi.fn(() => '')
      })
      const testItem: FileItem = {
        Name: 'test.txt',
        Path: '/path/to/test.txt',
        IsDir: false,
        Size: 1024,
        ModTime: '2024-01-01T00:00:00Z'
      }
      const { showToast } = await import('../api/errors')

      clipboard.copyItem(testItem)
      await clipboard.pasteItem()

      expect(showToast).toHaveBeenCalledWith('browserView.selectStorageFirst', 'error')
    })

    it('should not paste if source path is malicious', async () => {
      const clipboard = useBrowserClipboard(mockOptions)
      const testItem: FileItem = {
        Name: 'test.txt',
        Path: '../malicious/file.txt',
        IsDir: false,
        Size: 1024,
        ModTime: '2024-01-01T00:00:00Z'
      }

      clipboard.copyItem(testItem)
      await clipboard.pasteItem()

      const api = await import('../api')
      expect(api.copyFile).not.toHaveBeenCalled()
    })

    it('should not paste if destination path is malicious', async () => {
      const clipboard = useBrowserClipboard({
        ...mockOptions,
        browserPath: vi.fn(() => '../malicious/path')
      })
      const testItem: FileItem = {
        Name: 'test.txt',
        Path: 'safe/path/test.txt',
        IsDir: false,
        Size: 1024,
        ModTime: '2024-01-01T00:00:00Z'
      }

      clipboard.copyItem(testItem)
      await clipboard.pasteItem()

      const api = await import('../api')
      expect(api.copyFile).not.toHaveBeenCalled()
    })

    it('should paste file with copy action successfully', async () => {
      const clipboard = useBrowserClipboard(mockOptions)
      const testItem: FileItem = {
        Name: 'test.txt',
        Path: '/path/to/test.txt',
        IsDir: false,
        Size: 1024,
        ModTime: '2024-01-01T00:00:00Z'
      }
      const api = await import('../api')

      clipboard.copyItem(testItem)
      await clipboard.pasteItem()

      expect(api.copyFile).toHaveBeenCalled()
      expect(clipboard.clipboardItem.value).toBeNull()
      expect(clipboard.clipboardAction.value).toBeNull()
      expect(mockOptions.refreshBrowser).toHaveBeenCalled()
    })

    it('should paste directory with copy action successfully', async () => {
      const clipboard = useBrowserClipboard(mockOptions)
      const testItem: FileItem = {
        Name: 'test-dir',
        Path: '/path/to/test-dir',
        IsDir: true,
        Size: 0,
        ModTime: '2024-01-01T00:00:00Z'
      }
      const api = await import('../api')

      clipboard.copyItem(testItem)
      await clipboard.pasteItem()

      expect(api.copyDir).toHaveBeenCalled()
    })

    it('should paste file with move action successfully', async () => {
      const clipboard = useBrowserClipboard(mockOptions)
      const testItem: FileItem = {
        Name: 'test.txt',
        Path: '/path/to/test.txt',
        IsDir: false,
        Size: 1024,
        ModTime: '2024-01-01T00:00:00Z'
      }
      const api = await import('../api')

      clipboard.moveItem(testItem)
      await clipboard.pasteItem()

      expect(api.moveFile).toHaveBeenCalled()
      expect(mockOptions.refreshUntilGone).toHaveBeenCalledWith(['test.txt'])
    })

    it('should paste directory with move action successfully', async () => {
      const clipboard = useBrowserClipboard(mockOptions)
      const testItem: FileItem = {
        Name: 'test-dir',
        Path: '/path/to/test-dir',
        IsDir: true,
        Size: 0,
        ModTime: '2024-01-01T00:00:00Z'
      }
      const api = await import('../api')

      clipboard.moveItem(testItem)
      await clipboard.pasteItem()

      expect(api.moveDir).toHaveBeenCalled()
    })

    it('should handle paste errors gracefully', async () => {
      const clipboard = useBrowserClipboard(mockOptions)
      const testItem: FileItem = {
        Name: 'test.txt',
        Path: '/path/to/test.txt',
        IsDir: false,
        Size: 1024,
        ModTime: '2024-01-01T00:00:00Z'
      }
      const api = await import('../api')
      const { showToast } = await import('../api/errors')
      vi.mocked(api.copyFile).mockRejectedValue(new Error('Paste failed'))

      clipboard.copyItem(testItem)
      await clipboard.pasteItem()

      expect(showToast).toHaveBeenCalled()
    })

    it('should handle empty srcFs correctly', async () => {
      const clipboard = useBrowserClipboard(mockOptions)
      const testItem: FileItem = {
        Name: 'test.txt',
        Path: '/path/to/test.txt',
        IsDir: false,
        Size: 1024,
        ModTime: '2024-01-01T00:00:00Z'
      }
      const api = await import('../api')

      clipboard.clipboardSrcFs.value = ''
      clipboard.clipboardItem.value = testItem
      clipboard.clipboardAction.value = 'copy'

      await clipboard.pasteItem()

      expect(api.copyFile).toHaveBeenCalled()
    })

    it('should handle empty browserPath correctly', async () => {
      const clipboard = useBrowserClipboard({
        ...mockOptions,
        browserPath: vi.fn(() => '')
      })
      const testItem: FileItem = {
        Name: 'test.txt',
        Path: '/path/to/test.txt',
        IsDir: false,
        Size: 1024,
        ModTime: '2024-01-01T00:00:00Z'
      }
      const api = await import('../api')

      clipboard.copyItem(testItem)
      await clipboard.pasteItem()

      expect(api.copyFile).toHaveBeenCalled()
    })
  })

  describe('clearClipboard', () => {
    it('should clear clipboard state', () => {
      const clipboard = useBrowserClipboard(mockOptions)
      const testItem: FileItem = {
        Name: 'test.txt',
        Path: '/path/to/test.txt',
        IsDir: false,
        Size: 1024,
        ModTime: '2024-01-01T00:00:00Z'
      }

      clipboard.copyItem(testItem)
      clipboard.clearClipboard()

      expect(clipboard.clipboardItem.value).toBeNull()
      expect(clipboard.clipboardAction.value).toBeNull()
    })
  })
})
