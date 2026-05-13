import { describe, it, expect, vi } from 'vitest'
import { ref, nextTick } from 'vue'
import { useTaskPathBrowse } from './useTaskPathBrowse'

describe('useTaskPathBrowse', () => {
  it('loads source/target paths, updates current path, and builds breadcrumbs', async () => {
    const createForm = ref({
      sourceRemote: 'src',
      sourcePath: '',
      targetRemote: 'dst',
      targetPath: '',
    })
    const listPath = vi.fn(async (remote: string, path: string) => ({
      items: [{ Path: `${remote}:${path || '/'}` }],
    }))

    const api = useTaskPathBrowse({ createForm, listPath })

    await api.loadSourcePath('src', '/a/b')
    await api.loadTargetPath('dst', '/c/d')
    await nextTick()

    expect(api.sourcePathOptions.value).toEqual([{ Path: 'src:/a/b' }])
    expect(api.targetPathOptions.value).toEqual([{ Path: 'dst:/c/d' }])
    expect(api.sourceCurrentPath.value).toBe('/a/b')
    expect(api.targetCurrentPath.value).toBe('/c/d')
    expect(api.sourceBreadcrumbs.value).toEqual([
      { name: 'src:', path: '' },
      { name: 'a', path: '/a' },
      { name: 'b', path: '/a/b' },
    ])
    expect(api.targetBreadcrumbs.value).toEqual([
      { name: 'dst:', path: '' },
      { name: 'c', path: '/c' },
      { name: 'd', path: '/c/d' },
    ])
  })

  it('toggles input visibility and resets browse state', () => {
    const createForm = ref({ sourceRemote: '', sourcePath: '', targetRemote: '', targetPath: '' })
    const api = useTaskPathBrowse({ createForm, listPath: vi.fn(async () => ({ items: [] })) })

    api.setShowSourcePathInput(true)
    api.setShowTargetPathInput(true)
    api.sourcePathOptions.value = [{ Path: '/a' }]
    api.targetPathOptions.value = [{ Path: '/b' }]
    api.sourceCurrentPath.value = '/a'
    api.targetCurrentPath.value = '/b'

    expect(api.showSourcePathInput.value).toBe(true)
    expect(api.showTargetPathInput.value).toBe(true)

    api.resetTaskPathBrowse()

    expect(api.sourcePathOptions.value).toEqual([])
    expect(api.targetPathOptions.value).toEqual([])
    expect(api.showSourcePathInput.value).toBe(false)
    expect(api.showTargetPathInput.value).toBe(false)
    expect(api.sourceCurrentPath.value).toBe('')
    expect(api.targetCurrentPath.value).toBe('')
  })

  it('restores path browsing from existing task remotes and paths', async () => {
    const createForm = ref({ sourceRemote: '', sourcePath: '', targetRemote: '', targetPath: '' })
    const listPath = vi.fn(async (_remote: string, path: string) => ({ items: [{ Path: path }] }))
    const api = useTaskPathBrowse({ createForm, listPath })

    await api.restoreTaskPathBrowse({
      id: 1,
      name: 'demo',
      mode: 'copy',
      sourceRemote: 'src',
      sourcePath: '/from',
      targetRemote: 'dst',
      targetPath: '/to',
      createdAt: '',
    })

    expect(listPath).toHaveBeenNthCalledWith(1, 'src', '/from')
    expect(listPath).toHaveBeenNthCalledWith(2, 'dst', '/to')
    expect(api.sourceCurrentPath.value).toBe('/from')
    expect(api.targetCurrentPath.value).toBe('/to')
  })

  it('reacts to remote changes and clears form paths when remote is empty', async () => {
    const createForm = ref({
      sourceRemote: 'src',
      sourcePath: '/old-source',
      targetRemote: 'dst',
      targetPath: '/old-target',
    })
    const listPath = vi.fn(async () => ({ items: [{ Path: '/' }] }))
    const api = useTaskPathBrowse({ createForm, listPath })

    api.onSourceRemoteChange()
    api.onTargetRemoteChange()
    await Promise.resolve()

    expect(listPath).toHaveBeenNthCalledWith(1, 'src', '')
    expect(listPath).toHaveBeenNthCalledWith(2, 'dst', '')
    expect(createForm.value.sourcePath).toBe('')
    expect(createForm.value.targetPath).toBe('')

    createForm.value.sourceRemote = ''
    createForm.value.targetRemote = ''
    api.sourcePathOptions.value = [{ Path: '/keep?' }]
    api.targetPathOptions.value = [{ Path: '/keep?' }]
    api.onSourceRemoteChange()
    api.onTargetRemoteChange()

    expect(api.sourcePathOptions.value).toEqual([])
    expect(api.targetPathOptions.value).toEqual([])
  })

  it('handles breadcrumb and item click/arrow navigation', async () => {
    const createForm = ref({
      sourceRemote: 'src',
      sourcePath: '',
      targetRemote: 'dst',
      targetPath: '',
    })
    const listPath = vi.fn(async (_remote: string, path: string) => ({ items: [{ Path: path }] }))
    const api = useTaskPathBrowse({ createForm, listPath })

    api.showSourcePathInput.value = true
    api.showTargetPathInput.value = true
    api.onSourceClick({ Path: '/picked-source' })
    api.onTargetClick({ Path: '/picked-target' })
    expect(createForm.value.sourcePath).toBe('/picked-source')
    expect(createForm.value.targetPath).toBe('/picked-target')
    expect(api.showSourcePathInput.value).toBe(false)
    expect(api.showTargetPathInput.value).toBe(false)

    api.onSourceArrow({ Path: '/deep-source' })
    api.onTargetArrow({ Path: '/deep-target' })
    await Promise.resolve()
    expect(createForm.value.sourcePath).toBe('/deep-source')
    expect(createForm.value.targetPath).toBe('/deep-target')
    expect(listPath).toHaveBeenCalledWith('src', '/deep-source')
    expect(listPath).toHaveBeenCalledWith('dst', '/deep-target')

    api.sourceCurrentPath.value = '/deep-source'
    api.targetCurrentPath.value = '/deep-target'
    api.onSourceBreadcrumbClick('/deep-source')
    api.onTargetBreadcrumbClick('/deep-target')
    expect(listPath).toHaveBeenCalledTimes(2)

    api.onSourceBreadcrumbClick('/root-source')
    api.onTargetBreadcrumbClick('/root-target')
    await Promise.resolve()
    expect(listPath).toHaveBeenCalledWith('src', '/root-source')
    expect(listPath).toHaveBeenCalledWith('dst', '/root-target')
  })

  it('swallows load errors and keeps prior state', async () => {
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    const createForm = ref({ sourceRemote: 'src', sourcePath: '', targetRemote: 'dst', targetPath: '' })
    const listPath = vi.fn(async () => { throw new Error('boom') })
    const api = useTaskPathBrowse({ createForm, listPath })

    api.sourcePathOptions.value = [{ Path: '/old' }]
    api.targetPathOptions.value = [{ Path: '/old2' }]
    api.sourceCurrentPath.value = '/old'
    api.targetCurrentPath.value = '/old2'

    await api.loadSourcePath('src', '/new')
    await api.loadTargetPath('dst', '/new2')

    expect(errorSpy).toHaveBeenCalled()
    expect(api.sourcePathOptions.value).toEqual([{ Path: '/old' }])
    expect(api.targetPathOptions.value).toEqual([{ Path: '/old2' }])
    expect(api.sourceCurrentPath.value).toBe('/old')
    expect(api.targetCurrentPath.value).toBe('/old2')
  })
})
