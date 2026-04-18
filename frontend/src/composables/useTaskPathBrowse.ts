import { computed, ref } from 'vue'
import type { Ref } from 'vue'
import type { Task } from '../types'

interface PathBrowseItem {
  IsDir?: boolean
  Path: string
}

interface UseTaskPathBrowseOptions {
  createForm: Ref<any>
  listPath: (remote: string, path: string) => Promise<{ items?: any[] }>
}

export function useTaskPathBrowse(options: UseTaskPathBrowseOptions) {
  const sourcePathOptions = ref<any[]>([])
  const targetPathOptions = ref<any[]>([])
  const showSourcePathInput = ref(false)
  const showTargetPathInput = ref(false)
  const sourceCurrentPath = ref('')
  const targetCurrentPath = ref('')

  function setShowSourcePathInput(value: boolean) {
    showSourcePathInput.value = value
  }

  function setShowTargetPathInput(value: boolean) {
    showTargetPathInput.value = value
  }

  async function loadSourcePath(remote: string, path: string) {
    try {
      const data = await options.listPath(remote, path)
      sourcePathOptions.value = data.items || []
      sourceCurrentPath.value = path
    } catch (e) {
      console.error(e)
    }
  }

  async function loadTargetPath(remote: string, path: string) {
    try {
      const data = await options.listPath(remote, path)
      targetPathOptions.value = data.items || []
      targetCurrentPath.value = path
    } catch (e) {
      console.error(e)
    }
  }

  function resetTaskPathBrowse() {
    sourcePathOptions.value = []
    targetPathOptions.value = []
    showSourcePathInput.value = false
    showTargetPathInput.value = false
    sourceCurrentPath.value = ''
    targetCurrentPath.value = ''
  }

  async function restoreTaskPathBrowse(task: Task) {
    if (task.sourceRemote) {
      const sourcePath = task.sourcePath || ''
      sourceCurrentPath.value = sourcePath
      await loadSourcePath(task.sourceRemote, sourcePath)
    }

    if (task.targetRemote) {
      const targetPath = task.targetPath || ''
      targetCurrentPath.value = targetPath
      await loadTargetPath(task.targetRemote, targetPath)
    }
  }

  function onSourceRemoteChange() {
    sourceCurrentPath.value = ''
    if (options.createForm.value.sourceRemote) {
      loadSourcePath(options.createForm.value.sourceRemote, '')
    } else {
      sourcePathOptions.value = []
    }
    options.createForm.value.sourcePath = ''
  }

  function onTargetRemoteChange() {
    targetCurrentPath.value = ''
    if (options.createForm.value.targetRemote) {
      loadTargetPath(options.createForm.value.targetRemote, '')
    } else {
      targetPathOptions.value = []
    }
    options.createForm.value.targetPath = ''
  }

  function onSourceBreadcrumbClick(path: string) {
    if (!options.createForm.value.sourceRemote || path === sourceCurrentPath.value) return
    loadSourcePath(options.createForm.value.sourceRemote, path)
  }

  function onTargetBreadcrumbClick(path: string) {
    if (!options.createForm.value.targetRemote || path === targetCurrentPath.value) return
    loadTargetPath(options.createForm.value.targetRemote, path)
  }

  function onSourceClick(item: PathBrowseItem) {
    options.createForm.value.sourcePath = item.Path
    showSourcePathInput.value = false
  }

  function onSourceArrow(item: PathBrowseItem) {
    options.createForm.value.sourcePath = item.Path
    loadSourcePath(options.createForm.value.sourceRemote, item.Path)
  }

  function onTargetClick(item: PathBrowseItem) {
    options.createForm.value.targetPath = item.Path
    showTargetPathInput.value = false
  }

  function onTargetArrow(item: PathBrowseItem) {
    options.createForm.value.targetPath = item.Path
    loadTargetPath(options.createForm.value.targetRemote, item.Path)
  }

  const sourceBreadcrumbs = computed(() => {
    if (!options.createForm.value.sourceRemote) return []
    const parts = (sourceCurrentPath.value || '').split('/').filter(Boolean)
    const crumbs = [{ name: options.createForm.value.sourceRemote + ':', path: '' }]
    let current = ''
    for (const part of parts) {
      current += '/' + part
      crumbs.push({ name: part, path: current })
    }
    return crumbs
  })

  const targetBreadcrumbs = computed(() => {
    if (!options.createForm.value.targetRemote) return []
    const parts = (targetCurrentPath.value || '').split('/').filter(Boolean)
    const crumbs = [{ name: options.createForm.value.targetRemote + ':', path: '' }]
    let current = ''
    for (const part of parts) {
      current += '/' + part
      crumbs.push({ name: part, path: current })
    }
    return crumbs
  })

  return {
    sourcePathOptions,
    targetPathOptions,
    showSourcePathInput,
    showTargetPathInput,
    sourceCurrentPath,
    targetCurrentPath,
    sourceBreadcrumbs,
    targetBreadcrumbs,
    setShowSourcePathInput,
    setShowTargetPathInput,
    resetTaskPathBrowse,
    restoreTaskPathBrowse,
    onSourceRemoteChange,
    onTargetRemoteChange,
    onSourceBreadcrumbClick,
    onTargetBreadcrumbClick,
    loadSourcePath,
    loadTargetPath,
    onSourceClick,
    onSourceArrow,
    onTargetClick,
    onTargetArrow,
  }
}
