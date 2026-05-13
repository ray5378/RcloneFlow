import { describe, it, expect, vi, beforeEach } from 'vitest'
import { getRemotes, createRemote, updateRemote, getRemoteConfig, deleteRemote, testRemote, getProviders } from './remote'

vi.mock('./client', () => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  del: vi.fn(),
}))

import { get, post, put, del } from './client'

describe('remote API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('gets remotes', async () => {
    const payload = { remotes: ['a'], version: '1.0.0' }
    ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(payload)
    await expect(getRemotes()).resolves.toEqual(payload)
    expect(get).toHaveBeenCalledWith('/api/remotes')
  })

  it('creates and updates remotes', async () => {
    const params = { token: 'x' }
    await createRemote('r1', 'webdav', params)
    expect(post).toHaveBeenCalledWith('/api/remotes', { name: 'r1', type: 'webdav', parameters: params })

    await updateRemote('r1', 'webdav', params)
    expect(put).toHaveBeenCalledWith('/api/remotes', { name: 'r1', type: 'webdav', parameters: params })
  })

  it('encodes remote config and delete paths', async () => {
    await getRemoteConfig('a/b c')
    expect(get).toHaveBeenCalledWith('/api/remotes/config/a%2Fb%20c')

    await deleteRemote('a/b c')
    expect(del).toHaveBeenCalledWith('/api/config/a%2Fb%20c')
  })

  it('tests remote and gets providers', async () => {
    const tested = { ok: true, count: 3 }
    ;(post as ReturnType<typeof vi.fn>).mockResolvedValueOnce(tested)
    await expect(testRemote('demo')).resolves.toEqual(tested)
    expect(post).toHaveBeenCalledWith('/api/remotes/test', { name: 'demo' })

    const providers = { providers: [{ Name: 'x' }] }
    ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(providers)
    await expect(getProviders()).resolves.toEqual(providers)
    expect(get).toHaveBeenCalledWith('/api/providers')
  })
})
