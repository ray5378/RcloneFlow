import { ref, computed, onMounted } from 'vue'
import { t } from '../i18n'
import { showToast } from '../api/errors'
import * as api from '../api'
import type { FileItem } from '../types'
import { useBrowserClipboard } from './useBrowserClipboard'
import { useBrowserContextMenu } from './useBrowserContextMenu'
import { useBrowserFileOps } from './useBrowserFileOps'
import { useBrowserRemoteManagement } from './useBrowserRemoteManagement'

const REFRESH_DELAY_MS = 200
const DEFAULT_TIMEOUT_MS = 5000
const DEFAULT_INTERVAL_MS = 400

export interface UseBrowserOptions {
  emit?: (event: 'navigate', subview: string) => void
}

export function useBrowser(options: UseBrowserOptions = {}) {
  const browserFs = ref('')
  const browserPath = ref('')
  const browserItems = ref<FileItem[]>([])
  const browserError = ref('')
  const confirmModal = ref<{ show: boolean; title: string; message: string; onConfirm: () => void }>({
    show: false,
    title: '',
    message: '',
    onConfirm: () => {}
  })
  const showAddRemote = ref(false)
  const isEditMode = ref(false)
  const editRemoteName = ref('')
  const showEditDesc = ref(false)
  const editDescRemote = ref('')
  const subview = ref('explorer')

  function delay(ms: number) { return new Promise(res => setTimeout(res, ms)) }

  function hasItemByName(name: string): boolean {
    return browserItems.value.some(it => it.Name === name)
  }

  async function refreshUntilGone(names: string[], timeoutMs = DEFAULT_TIMEOUT_MS, intervalMs = DEFAULT_INTERVAL_MS) {
    const deadline = Date.now() + timeoutMs
    await delay(REFRESH_DELAY_MS)
    while (Date.now() < deadline) {
      await refreshBrowser()
      const stillPresent = names.some(n => hasItemByName(n))
      if (!stillPresent) return
      await delay(intervalMs)
    }
  }

  async function refreshBrowser() {
    if (!browserFs.value) return
    browserError.value = ''
    try {
      const data = await api.listPath(browserFs.value, browserPath.value)
      browserItems.value = data.items || []
    } catch (e) {
      browserError.value = (e as Error).message
    }
  }

  function hasPathTraversal(path: string): boolean {
    if (!path) return false
    
    const normalized = path.replace(/\\/g, '/')
    
    if (/\.\./.test(normalized)) return true
    
    if (normalized === '.') return true
    
    if (/^\.\//.test(normalized)) return true
    
    if (/\/\.\//.test(normalized)) return true
    
    if (/\/\.\.$/.test(normalized)) return true
    
    const lowerPath = normalized.toLowerCase()
    if (lowerPath.includes('%2e%2e') || lowerPath.includes('%2e.')) {
      return true
    }
    
    return false
  }

  function validatePath(path: string, name: string = t('browserView.path')): boolean {
    if (hasPathTraversal(path)) {
      showToast(t('browserView.pathTraversalRisk').replace('{name}', name), 'error')
      return false
    }
    return true
  }

  const clipboard = useBrowserClipboard({
    browserFs: () => browserFs.value,
    browserPath: () => browserPath.value,
    refreshBrowser,
    refreshUntilGone
  })

  const contextMenu = useBrowserContextMenu()

  const fileOps = useBrowserFileOps({
    browserFs: () => browserFs.value,
    browserPath: () => browserPath.value,
    refreshBrowser,
    refreshUntilGone,
    validatePath
  })

  const remoteMgmt = useBrowserRemoteManagement({
    async openRemote(name: string) {
      browserFs.value = name
      browserPath.value = ''
      await refreshBrowser()
      subview.value = 'explorer'
    },
    async loadRemotes() {
      await loadRemotesInternal()
    }
  })

  const breadcrumbs = computed(() => {
    const parts = browserPath.value.split('/').filter(Boolean)
    const crumbs = [{ name: browserFs.value + ':', path: '' }]
    let current = ''
    for (const p of parts) {
      current += '/' + p
      crumbs.push({ name: p, path: current })
    }
    return crumbs
  })

  async function loadRemotesInternal() {
    try {
      const data = await api.listRemotes()
      remoteMgmt.remotes.value = data.remotes || []

      const orderedRemotes = remoteMgmt.getOrderedRemotes()
      if (orderedRemotes.length > 0) {
        browserFs.value = orderedRemotes[0]
        browserPath.value = ''
        await refreshBrowser()
      }

      if (options.emit) {
        options.emit('navigate', 'explorer')
      }
    } catch (e) {
      browserError.value = (e as Error).message
    }
  }

  function enterItem(item: FileItem) {
    if (!item.IsDir) return
    const remotePrefix = browserFs.value + ':'
    let newPath = item.Path
    if (newPath.startsWith(remotePrefix)) {
      newPath = newPath.substring(remotePrefix.length)
    }
    newPath = newPath.replace(/^\/+/, '')
    browserPath.value = newPath
    refreshBrowser()
  }

  function formatTime(time: string) {
    if (!time) return '-'
    try {
      return new Date(time).toLocaleString('zh-CN', {
        year: 'numeric', month: '2-digit', day: '2-digit',
        hour: '2-digit', minute: '2-digit'
      })
    } catch {
      return time
    }
  }

  function formatSize(size: string) {
    if (!size || size === '-') return '-'
    const bytes = parseInt(size)
    if (isNaN(bytes)) return size
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' K'
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' M'
    return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' G'
  }

  function copyItem() {
    if (!contextMenu.contextMenu.value.item) return
    clipboard.copyItem(contextMenu.contextMenu.value.item)
    contextMenu.closeContextMenu()
  }

  function moveItem() {
    if (!contextMenu.contextMenu.value.item) return
    clipboard.moveItem(contextMenu.contextMenu.value.item)
    contextMenu.closeContextMenu()
  }

  function startRename() {
    fileOps.startRename(contextMenu.contextMenu.value.item, contextMenu.closeContextMenu)
  }

  function confirmDelete() {
    fileOps.confirmDelete(contextMenu.contextMenu.value.item, contextMenu.closeContextMenu)
  }

  async function handleDeleteRemote(name: string) {
    const modalConfig = await remoteMgmt.deleteRemote(name)
    confirmModal.value = {
      ...modalConfig,
      onConfirm: async () => {
        await modalConfig.onConfirm()
        confirmModal.value.show = false
      }
    }
  }

  function openManageStorage() {
    subview.value = 'manage-storage'
  }

  function openAddRemoteHandler() {
    isEditMode.value = false
    editRemoteName.value = ''
    showAddRemote.value = true
  }

  async function openEditRemote(name: string) {
    isEditMode.value = true
    editRemoteName.value = name
    showAddRemote.value = true
  }

  function openEditDesc(name: string) {
    editDescRemote.value = name
    showEditDesc.value = true
  }

  function saveDesc(desc: string) {
    remoteMgmt.saveDesc(editDescRemote.value, desc)
  }

  onMounted(async () => {
    await remoteMgmt.loadRemoteOrder()
    await loadRemotesInternal()
    const onContextClick = () => { contextMenu.contextMenu.value.show = false }
    document.addEventListener('click', onContextClick)
    return () => document.removeEventListener('click', onContextClick)
  })

  return {
    browserFs,
    browserPath,
    browserItems,
    browserError,
    confirmModal,
    showAddRemote,
    isEditMode,
    editRemoteName,
    showEditDesc,
    editDescRemote,
    subview,

    clipboard,
    contextMenu,
    fileOps,
    remoteMgmt,

    breadcrumbs,

    refreshBrowser,
    refreshUntilGone,
    loadRemotes: loadRemotesInternal,
    enterItem,
    formatTime,
    formatSize,
    copyItem,
    moveItem,
    startRename,
    confirmDelete,
    handleDeleteRemote,
    openManageStorage,
    openAddRemote: openAddRemoteHandler,
    openEditRemote,
    openEditDesc,
    saveDesc
  }
}
