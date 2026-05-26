import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useBrowserContextMenu } from './useBrowserContextMenu'
import type { FileItem } from '../types'

describe('useBrowserContextMenu.ts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('initial state', () => {
    it('should initialize with closed menu', () => {
      const menu = useBrowserContextMenu()
      expect(menu.contextMenu.value.show).toBe(false)
      expect(menu.contextMenu.value.item).toBeNull()
    })
  })

  describe('showContextMenuHandler', () => {
    it('should show menu for file item', () => {
      const menu = useBrowserContextMenu()
      const testItem: FileItem = {
        Name: 'test.txt',
        Path: '/path/to/test.txt',
        IsDir: false,
        Size: '1024',
        ModTime: '2024-01-01T00:00:00Z'
      }

      const preventDefault = vi.fn()
      menu.showContextMenuHandler({
        clientX: 100,
        clientY: 200,
        preventDefault
      } as unknown as MouseEvent, testItem)

      expect(preventDefault).toHaveBeenCalled()
      expect(menu.contextMenu.value.show).toBe(true)
      expect(menu.contextMenu.value.x).toBe(100)
      expect(menu.contextMenu.value.y).toBe(200)
      expect(menu.contextMenu.value.item).toEqual(testItem)
    })

    it('should show menu for directory item', () => {
      const menu = useBrowserContextMenu()
      const testItem: FileItem = {
        Name: 'test-dir',
        Path: '/path/to/test-dir',
        IsDir: true,
        Size: '0',
        ModTime: '2024-01-01T00:00:00Z'
      }

      menu.showContextMenuHandler({
        clientX: 150,
        clientY: 250,
        preventDefault: vi.fn()
      } as unknown as MouseEvent, testItem)

      expect(menu.contextMenu.value.show).toBe(true)
      expect(menu.contextMenu.value.x).toBe(150)
      expect(menu.contextMenu.value.y).toBe(250)
      expect(menu.contextMenu.value.item).toEqual(testItem)
    })

    it('should show menu at different coordinates', () => {
      const menu = useBrowserContextMenu()
      const testItem: FileItem = {
        Name: 'test.txt',
        Path: '/path/to/test.txt',
        IsDir: false,
        Size: '1024',
        ModTime: '2024-01-01T00:00:00Z'
      }

      menu.showContextMenuHandler({
        clientX: 500,
        clientY: 600,
        preventDefault: vi.fn()
      } as unknown as MouseEvent, testItem)

      expect(menu.contextMenu.value.x).toBe(500)
      expect(menu.contextMenu.value.y).toBe(600)
    })
  })

  describe('showBackgroundMenuHandler', () => {
    it('should show background menu without item', () => {
      const menu = useBrowserContextMenu()
      const preventDefault = vi.fn()

      menu.showBackgroundMenuHandler({
        clientX: 100,
        clientY: 200,
        preventDefault
      } as unknown as MouseEvent)

      expect(preventDefault).toHaveBeenCalled()
      expect(menu.contextMenu.value.show).toBe(true)
      expect(menu.contextMenu.value.x).toBe(100)
      expect(menu.contextMenu.value.y).toBe(200)
      expect(menu.contextMenu.value.item).toBeNull()
    })

    it('should show background menu at different coordinates', () => {
      const menu = useBrowserContextMenu()

      menu.showBackgroundMenuHandler({
        clientX: 300,
        clientY: 400,
        preventDefault: vi.fn()
      } as unknown as MouseEvent)

      expect(menu.contextMenu.value.x).toBe(300)
      expect(menu.contextMenu.value.y).toBe(400)
    })
  })

  describe('closeContextMenu', () => {
    it('should close open menu', () => {
      const menu = useBrowserContextMenu()
      const testItem: FileItem = {
        Name: 'test.txt',
        Path: '/path/to/test.txt',
        IsDir: false,
        Size: '1024',
        ModTime: '2024-01-01T00:00:00Z'
      }

      menu.showContextMenuHandler({
        clientX: 100,
        clientY: 200,
        preventDefault: vi.fn()
      } as unknown as MouseEvent, testItem)

      expect(menu.contextMenu.value.show).toBe(true)

      menu.closeContextMenu()

      expect(menu.contextMenu.value.show).toBe(false)
    })

    it('should do nothing if menu is already closed', () => {
      const menu = useBrowserContextMenu()

      expect(menu.contextMenu.value.show).toBe(false)

      menu.closeContextMenu()

      expect(menu.contextMenu.value.show).toBe(false)
    })

    it('should preserve item state when closing', () => {
      const menu = useBrowserContextMenu()
      const testItem: FileItem = {
        Name: 'test.txt',
        Path: '/path/to/test.txt',
        IsDir: false,
        Size: '1024',
        ModTime: '2024-01-01T00:00:00Z'
      }

      menu.showContextMenuHandler({
        clientX: 100,
        clientY: 200,
        preventDefault: vi.fn()
      } as unknown as MouseEvent, testItem)

      menu.closeContextMenu()

      expect(menu.contextMenu.value.item).toEqual(testItem)
    })
  })

  describe('menu navigation', () => {
    it('should support opening and closing multiple times', () => {
      const menu = useBrowserContextMenu()
      const testItem: FileItem = {
        Name: 'test.txt',
        Path: '/path/to/test.txt',
        IsDir: false,
        Size: '1024',
        ModTime: '2024-01-01T00:00:00Z'
      }

      menu.showContextMenuHandler({
        clientX: 100,
        clientY: 200,
        preventDefault: vi.fn()
      } as unknown as MouseEvent, testItem)
      expect(menu.contextMenu.value.show).toBe(true)

      menu.closeContextMenu()
      expect(menu.contextMenu.value.show).toBe(false)

      menu.showContextMenuHandler({
        clientX: 300,
        clientY: 400,
        preventDefault: vi.fn()
      } as unknown as MouseEvent, testItem)
      expect(menu.contextMenu.value.show).toBe(true)
    })
  })
})
