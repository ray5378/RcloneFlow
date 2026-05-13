import { describe, it, expect, vi, beforeEach } from 'vitest'
import { listPath, copyFile, moveFile, copyDir, moveDir, deleteFile, purgeDir, mkdir } from './browser'

vi.mock('./client', () => ({
  get: vi.fn(),
  post: vi.fn(),
}))

import { get, post } from './client'

describe('browser API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('lists path with encoded params', async () => {
    const payload = { fs: 'r1', items: [] }
    ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(payload)

    await expect(listPath('remote a', '/目录/a b')).resolves.toEqual(payload)
    expect(get).toHaveBeenCalledWith('/api/browser/list?remote=remote%20a&path=%2F%E7%9B%AE%E5%BD%95%2Fa%20b')
  })

  it('copies and moves files', async () => {
    await copyFile('src', '/a', 'dst', '/b')
    expect(post).toHaveBeenCalledWith('/api/fs/copy', {
      srcFs: 'src:', srcRemote: '/a', dstFs: 'dst:', dstRemote: '/b',
    })

    await moveFile('src', '/a', 'dst', '/b')
    expect(post).toHaveBeenCalledWith('/api/fs/move', {
      srcFs: 'src:', srcRemote: '/a', dstFs: 'dst:', dstRemote: '/b',
    })
  })

  it('copies and moves directories', async () => {
    await copyDir('src', 'dirA', 'dst', 'dirB')
    expect(post).toHaveBeenCalledWith('/api/fs/copyDir', {
      srcFs: 'src:dirA', dstFs: 'dst:dirB', createEmptySrcDirs: true,
    })

    await moveDir('src', 'dirA', 'dst', 'dirB')
    expect(post).toHaveBeenCalledWith('/api/fs/moveDir', {
      srcFs: 'src:dirA', dstFs: 'dst:dirB', createEmptySrcDirs: true, deleteEmptySrcDirs: true,
    })
  })

  it('deletes, purges and creates directories', async () => {
    await deleteFile('r1', '/a')
    expect(post).toHaveBeenCalledWith('/api/fs/delete', { fs: 'r1:', remote: '/a' })

    await purgeDir('r1', '/a')
    expect(post).toHaveBeenCalledWith('/api/fs/purge', { fs: 'r1:', remote: '/a' })

    await mkdir('r1', '/a')
    expect(post).toHaveBeenCalledWith('/api/fs/mkdir', { fs: 'r1:', remote: '/a' })
  })
})
