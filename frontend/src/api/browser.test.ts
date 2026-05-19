import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { listPath, copyFile, moveFile, copyDir, moveDir, deleteFile, purgeDir, mkdir } from './browser'

const mockFetch = vi.fn()
globalThis.fetch = mockFetch as any

describe('browser.ts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.setItem('authToken', 'test-token')
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('listPath', () => {
    it('should list directory contents', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          fs: 'myRemote:',
          items: [
            { Name: 'file.txt', Path: 'file.txt', IsDir: false, Size: '100', ModTime: '2024-01-01' },
            { Name: 'folder', Path: 'folder', IsDir: true, Size: '0', ModTime: '2024-01-01' }
          ]
        })
      })

      const result = await listPath('myRemote', '/path/to/dir')

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/browser/list?remote=myRemote&path=%2Fpath%2Fto%2Fdir'),
        expect.any(Object)
      )
      expect(result.fs).toBe('myRemote:')
      expect(result.items).toHaveLength(2)
    })
  })

  describe('copyFile', () => {
    it('should copy a file', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({}) })

      await copyFile('src', '/src.txt', 'dst', '/dst.txt')

      expect(mockFetch).toHaveBeenCalledWith('/api/fs/copy', expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({
          srcFs: 'src:',
          srcRemote: '/src.txt',
          dstFs: 'dst:',
          dstRemote: '/dst.txt'
        })
      }))
    })
  })

  describe('moveFile', () => {
    it('should move a file', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({}) })

      await moveFile('src', '/src.txt', 'dst', '/dst.txt')

      expect(mockFetch).toHaveBeenCalledWith('/api/fs/move', expect.objectContaining({
        body: JSON.stringify({
          srcFs: 'src:',
          srcRemote: '/src.txt',
          dstFs: 'dst:',
          dstRemote: '/dst.txt'
        })
      }))
    })
  })

  describe('copyDir', () => {
    it('should copy a directory', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({}) })

      await copyDir('src', '/srcDir', 'dst', '/dstDir')

      expect(mockFetch).toHaveBeenCalledWith('/api/fs/copyDir', expect.objectContaining({
        body: JSON.stringify({
          srcFs: 'src:/srcDir',
          dstFs: 'dst:/dstDir',
          createEmptySrcDirs: true
        })
      }))
    })
  })

  describe('moveDir', () => {
    it('should move a directory', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({}) })

      await moveDir('src', '/srcDir', 'dst', '/dstDir')

      expect(mockFetch).toHaveBeenCalledWith('/api/fs/moveDir', expect.objectContaining({
        body: JSON.stringify({
          srcFs: 'src:/srcDir',
          dstFs: 'dst:/dstDir',
          createEmptySrcDirs: true,
          deleteEmptySrcDirs: true
        })
      }))
    })
  })

  describe('deleteFile', () => {
    it('should delete a file', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({}) })

      await deleteFile('myRemote', '/file.txt')

      expect(mockFetch).toHaveBeenCalledWith('/api/fs/delete', expect.objectContaining({
        body: JSON.stringify({ fs: 'myRemote:', remote: '/file.txt' })
      }))
    })
  })

  describe('purgeDir', () => {
    it('should purge a directory', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({}) })

      await purgeDir('myRemote', '/olddir')

      expect(mockFetch).toHaveBeenCalledWith('/api/fs/purge', expect.objectContaining({
        body: JSON.stringify({ fs: 'myRemote:', remote: '/olddir' })
      }))
    })
  })

  describe('mkdir', () => {
    it('should create a directory', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({}) })

      await mkdir('myRemote', '/newdir')

      expect(mockFetch).toHaveBeenCalledWith('/api/fs/mkdir', expect.objectContaining({
        body: JSON.stringify({ fs: 'myRemote:', remote: '/newdir' })
      }))
    })
  })
})
