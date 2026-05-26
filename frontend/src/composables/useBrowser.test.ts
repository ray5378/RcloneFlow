import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useBrowser } from './useBrowser'
import * as api from '../api'

vi.mock('../api', () => ({
  listPath: vi.fn().mockResolvedValue({ items: [] }),
  listRemotes: vi.fn().mockResolvedValue({ remotes: ['remote1'] }),
}))

vi.mock('../api/errors', () => ({
  showToast: vi.fn(),
}))

vi.mock('../i18n', () => ({
  t: vi.fn((key: string) => key),
}))

vi.mock('./useBrowserClipboard', () => ({
  useBrowserClipboard: vi.fn(() => ({
    copyItem: vi.fn(),
    moveItem: vi.fn(),
  })),
}))

vi.mock('./useBrowserContextMenu', () => ({
  useBrowserContextMenu: vi.fn(() => ({
    contextMenu: { value: { item: null, show: false } },
    closeContextMenu: vi.fn(),
  })),
}))

vi.mock('./useBrowserFileOps', () => ({
  useBrowserFileOps: vi.fn(() => ({
    startRename: vi.fn(),
    confirmDelete: vi.fn(),
  })),
}))

vi.mock('./useBrowserRemoteManagement', () => ({
  useBrowserRemoteManagement: vi.fn(() => ({
    remotes: { value: [] },
    getOrderedRemotes: vi.fn(() => ['remote1']),
    loadRemoteOrder: vi.fn().mockResolvedValue(undefined),
    deleteRemote: vi.fn().mockResolvedValue({ show: true, title: '', message: '', onConfirm: vi.fn() }),
    saveDesc: vi.fn(),
  })),
}))

describe('useBrowser', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('should be a function', () => {
    expect(typeof useBrowser).toBe('function')
  })

  it('should have correct initial state', () => {
    const browser = useBrowser()
    
    expect(browser.browserFs.value).toBe('')
    expect(browser.browserPath.value).toBe('')
    expect(browser.browserItems.value).toEqual([])
    expect(browser.browserError.value).toBe('')
    expect(browser.subview.value).toBe('explorer')
  })

  it('should have correct refreshBrowser method', () => {
    const browser = useBrowser()
    
    expect(typeof browser.refreshBrowser).toBe('function')
  })

  it('should have correct refreshUntilGone method', () => {
    const browser = useBrowser()
    
    expect(typeof browser.refreshUntilGone).toBe('function')
  })

  it('should have formatTime method', () => {
    const browser = useBrowser()
    
    expect(typeof browser.formatTime).toBe('function')
  })

  it('should format size correctly', () => {
    const browser = useBrowser()
    
    expect(browser.formatSize('')).toBe('-')
    expect(browser.formatSize('-')).toBe('-')
    expect(browser.formatSize('100')).toBe('100 B')
    expect(browser.formatSize('2048')).toBe('2.0 K')
    expect(browser.formatSize('2097152')).toBe('2.0 M')
    expect(browser.formatSize('2147483648')).toBe('2.0 G')
    expect(browser.formatSize('invalid')).toBe('invalid')
  })

  it('should generate breadcrumbs correctly', () => {
    const browser = useBrowser()
    
    browser.browserFs.value = 'remote1'
    browser.browserPath.value = ''
    expect(browser.breadcrumbs.value).toEqual([{ name: 'remote1:', path: '' }])
    
    browser.browserPath.value = 'folder1/subfolder'
    expect(browser.breadcrumbs.value).toEqual([
      { name: 'remote1:', path: '' },
      { name: 'folder1', path: '/folder1' },
      { name: 'subfolder', path: '/folder1/subfolder' },
    ])
  })

  it('should enter item correctly', () => {
    const browser = useBrowser()
    
    browser.browserFs.value = 'remote1'
    browser.browserPath.value = ''
    
    browser.enterItem({ Name: 'folder', IsDir: true, Path: 'remote1:/folder' } as any)
    
    expect(browser.browserPath.value).toBe('folder')
  })

  it('should not enter non-directory item', () => {
    const browser = useBrowser()
    const refreshSpy = vi.spyOn(browser, 'refreshBrowser')
    
    browser.enterItem({ Name: 'file.txt', IsDir: false, Path: 'remote1:/file.txt' } as any)
    
    expect(refreshSpy).not.toHaveBeenCalled()
  })

  it('should copy item', () => {
    const browser = useBrowser()
    const copySpy = vi.spyOn(browser.clipboard, 'copyItem')
    const closeSpy = vi.spyOn(browser.contextMenu, 'closeContextMenu')
    
    browser.contextMenu.contextMenu.value.item = { Name: 'test.txt' } as any
    browser.copyItem()
    
    expect(copySpy).toHaveBeenCalledWith({ Name: 'test.txt' })
    expect(closeSpy).toHaveBeenCalled()
  })

  it('should move item', () => {
    const browser = useBrowser()
    const moveSpy = vi.spyOn(browser.clipboard, 'moveItem')
    const closeSpy = vi.spyOn(browser.contextMenu, 'closeContextMenu')
    
    browser.contextMenu.contextMenu.value.item = { Name: 'test.txt' } as any
    browser.moveItem()
    
    expect(moveSpy).toHaveBeenCalledWith({ Name: 'test.txt' })
    expect(closeSpy).toHaveBeenCalled()
  })

  it('should start rename', () => {
    const browser = useBrowser()
    
    browser.contextMenu.contextMenu.value.item = { Name: 'test.txt' } as any
    browser.startRename()
  })

  it('should confirm delete', () => {
    const browser = useBrowser()
    
    browser.contextMenu.contextMenu.value.item = { Name: 'test.txt' } as any
    browser.confirmDelete()
  })

  it('should open manage storage', () => {
    const browser = useBrowser()
    
    browser.openManageStorage()
    
    expect(browser.subview.value).toBe('manage-storage')
  })

  it('should open add remote', () => {
    const browser = useBrowser()
    
    browser.openAddRemote()
    
    expect(browser.showAddRemote.value).toBe(true)
    expect(browser.isEditMode.value).toBe(false)
    expect(browser.editRemoteName.value).toBe('')
  })

  it('should open edit remote', () => {
    const browser = useBrowser()
    
    browser.openEditRemote('remote1')
    
    expect(browser.showAddRemote.value).toBe(true)
    expect(browser.isEditMode.value).toBe(true)
    expect(browser.editRemoteName.value).toBe('remote1')
  })

  it('should open edit desc', () => {
    const browser = useBrowser()
    
    browser.openEditDesc('remote1')
    
    expect(browser.showEditDesc.value).toBe(true)
    expect(browser.editDescRemote.value).toBe('remote1')
  })

  it('should save desc', () => {
    const browser = useBrowser()
    const saveDescSpy = vi.spyOn(browser.remoteMgmt, 'saveDesc')
    
    browser.editDescRemote.value = 'remote1'
    browser.saveDesc('new description')
    
    expect(saveDescSpy).toHaveBeenCalledWith('remote1', 'new description')
  })

  it('should handle copy/move without item', () => {
    const browser = useBrowser()
    const copySpy = vi.spyOn(browser.clipboard, 'copyItem')
    const moveSpy = vi.spyOn(browser.clipboard, 'moveItem')
    
    browser.contextMenu.contextMenu.value.item = null
    browser.copyItem()
    browser.moveItem()
    
    expect(copySpy).not.toHaveBeenCalled()
    expect(moveSpy).not.toHaveBeenCalled()
  })
})
