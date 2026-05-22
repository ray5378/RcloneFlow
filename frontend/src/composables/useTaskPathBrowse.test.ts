import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskPathBrowse } from './useTaskPathBrowse'

describe('useTaskPathBrowse', () => {
  function makeOptions() {
    return {
      createForm: ref({
        sourceRemote: '',
        sourcePath: '',
        targetRemote: '',
        targetPath: '',
      }),
      listPath: vi.fn().mockResolvedValue({ items: [{ Path: '/file.txt', IsDir: false }] }),
    }
  }

  it('should load source path', async () => {
    const opts = makeOptions()
    const { loadSourcePath, sourcePathOptions, sourceCurrentPath } = useTaskPathBrowse(opts)
    await loadSourcePath('remote', '/test')
    expect(sourcePathOptions.value).toHaveLength(1)
    expect(sourceCurrentPath.value).toBe('/test')
  })

  it('should load target path', async () => {
    const opts = makeOptions()
    const { loadTargetPath, targetPathOptions, targetCurrentPath } = useTaskPathBrowse(opts)
    await loadTargetPath('remote', '/dst')
    expect(targetPathOptions.value).toHaveLength(1)
    expect(targetCurrentPath.value).toBe('/dst')
  })

  it('should handle load path failure', async () => {
    const opts = makeOptions()
    opts.listPath.mockRejectedValueOnce(new Error('failed'))
    const { loadSourcePath } = useTaskPathBrowse(opts)
    await expect(loadSourcePath('remote', '/test')).rejects.toThrow('failed')
  })

  it('should reset browse state', () => {
    const opts = makeOptions()
    const { resetTaskPathBrowse, sourcePathOptions, targetPathOptions, showSourcePathInput, showTargetPathInput, sourceCurrentPath, targetCurrentPath } = useTaskPathBrowse(opts)
    sourcePathOptions.value = [{ Path: '/a' }] as any
    targetPathOptions.value = [{ Path: '/b' }] as any
    showSourcePathInput.value = true
    showTargetPathInput.value = true
    sourceCurrentPath.value = '/a'
    targetCurrentPath.value = '/b'
    resetTaskPathBrowse()
    expect(sourcePathOptions.value).toEqual([])
    expect(targetPathOptions.value).toEqual([])
    expect(showSourcePathInput.value).toBe(false)
    expect(showTargetPathInput.value).toBe(false)
    expect(sourceCurrentPath.value).toBe('')
    expect(targetCurrentPath.value).toBe('')
  })

  it('should handle source remote change with remote', async () => {
    const opts = makeOptions()
    const { onSourceRemoteChange, sourceCurrentPath } = useTaskPathBrowse(opts)
    opts.createForm.value.sourceRemote = 'myremote'
    await onSourceRemoteChange()
    expect(sourceCurrentPath.value).toBe('')
    expect(opts.listPath).toHaveBeenCalledWith('myremote', '')
  })

  it('should handle source remote change without remote', () => {
    const opts = makeOptions()
    const { onSourceRemoteChange, sourcePathOptions } = useTaskPathBrowse(opts)
    opts.createForm.value.sourceRemote = ''
    onSourceRemoteChange()
    expect(sourcePathOptions.value).toEqual([])
  })

  it('should handle target remote change with remote', async () => {
    const opts = makeOptions()
    const { onTargetRemoteChange, targetCurrentPath } = useTaskPathBrowse(opts)
    opts.createForm.value.targetRemote = 'myremote'
    await onTargetRemoteChange()
    expect(targetCurrentPath.value).toBe('')
    expect(opts.listPath).toHaveBeenCalledWith('myremote', '')
  })

  it('should handle onSourceClick', () => {
    const opts = makeOptions()
    const { onSourceClick, showSourcePathInput } = useTaskPathBrowse(opts)
    showSourcePathInput.value = true
    onSourceClick({ Path: '/selected.txt' } as any)
    expect(opts.createForm.value.sourcePath).toBe('/selected.txt')
    expect(showSourcePathInput.value).toBe(false)
  })

  it('should handle onTargetClick', () => {
    const opts = makeOptions()
    const { onTargetClick, showTargetPathInput } = useTaskPathBrowse(opts)
    showTargetPathInput.value = true
    onTargetClick({ Path: '/target.txt' } as any)
    expect(opts.createForm.value.targetPath).toBe('/target.txt')
    expect(showTargetPathInput.value).toBe(false)
  })

  it('should generate source breadcrumbs', () => {
    const opts = makeOptions()
    opts.createForm.value.sourceRemote = 'remote'
    const { sourceBreadcrumbs, sourceCurrentPath } = useTaskPathBrowse(opts)
    sourceCurrentPath.value = '/a/b'
    expect(sourceBreadcrumbs.value).toHaveLength(3)
    expect(sourceBreadcrumbs.value[0].name).toBe('remote:')
    expect(sourceBreadcrumbs.value[1].name).toBe('a')
    expect(sourceBreadcrumbs.value[2].name).toBe('b')
  })

  it('should return empty breadcrumbs without remote', () => {
    const opts = makeOptions()
    opts.createForm.value.sourceRemote = ''
    const { sourceBreadcrumbs } = useTaskPathBrowse(opts)
    expect(sourceBreadcrumbs.value).toEqual([])
  })

  it('should handle breadcrumb click', async () => {
    const opts = makeOptions()
    opts.createForm.value.sourceRemote = 'remote'
    const { onSourceBreadcrumbClick } = useTaskPathBrowse(opts)
    await onSourceBreadcrumbClick('/new')
    expect(opts.listPath).toHaveBeenCalledWith('remote', '/new')
  })

  it('should skip breadcrumb click if same path', async () => {
    const opts = makeOptions()
    opts.createForm.value.sourceRemote = 'remote'
    const { onSourceBreadcrumbClick, sourceCurrentPath } = useTaskPathBrowse(opts)
    sourceCurrentPath.value = '/same'
    await onSourceBreadcrumbClick('/same')
    expect(opts.listPath).not.toHaveBeenCalled()
  })
})
