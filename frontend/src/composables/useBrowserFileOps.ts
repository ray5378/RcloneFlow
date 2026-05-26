import { ref } from 'vue'
import { t } from '../i18n'
import { showToast } from '../api/errors'
import * as api from '../api'
import type { FileItem, RemoteTestState } from '../types'

export interface UseBrowserFileOpsOptions {
  browserFs: () => string
  browserPath: () => string
  refreshBrowser: () => Promise<void>
  refreshUntilGone: (names: string[], timeoutMs?: number, intervalMs?: number) => Promise<void>
  validatePath: (path: string, name?: string) => boolean
}

export interface UseBrowserFileOpsReturn {
  showDeleteConfirm: ReturnType<typeof ref<boolean>>
  deletingItem: ReturnType<typeof ref<FileItem | null>>
  showRenameInput: ReturnType<typeof ref<boolean>>
  renamingItem: ReturnType<typeof ref<FileItem | null>>
  renameInput: ReturnType<typeof ref<string>>
  testState: ReturnType<typeof ref<Record<string, RemoteTestState>>>
  startRename: (item: FileItem | null, closeMenu: () => void) => void
  confirmRename: () => Promise<void>
  confirmDelete: (item: FileItem | null, closeMenu: () => void) => void
  executeDelete: () => Promise<void>
  testRemote: (name: string) => Promise<void>
  getTestText: (name: string) => string
}

export function useBrowserFileOps(options: UseBrowserFileOpsOptions): UseBrowserFileOpsReturn {
  const showDeleteConfirm = ref(false)
  const deletingItem = ref<FileItem | null>(null)
  const showRenameInput = ref(false)
  const renamingItem = ref<FileItem | null>(null)
  const renameInput = ref('')
  const testState = ref<Record<string, RemoteTestState>>({})

  function startRename(item: FileItem | null, closeMenu: () => void) {
    if (!item) return
    renamingItem.value = item
    renameInput.value = item.Name
    showRenameInput.value = true
    closeMenu()
  }

  async function confirmRename() {
    if (!renamingItem.value || !renameInput.value) return
    if (!options.browserFs()) {
      showToast(t('browserView.selectStorageFirst'), 'error')
      return
    }
    try {
      const isDir = renamingItem.value.IsDir
      const srcPath = renamingItem.value.Path.replace(/^\/+/, '')
      const dstPath = (options.browserPath() ? options.browserPath().replace(/^\/+/, '') + '/' : '') + renameInput.value

      if (isDir) {
        await api.moveDir(options.browserFs(), srcPath, options.browserFs(), dstPath)
      } else {
        await api.moveFile(options.browserFs(), srcPath, options.browserFs(), dstPath)
      }

      showRenameInput.value = false
      const oldName = renamingItem.value.Name
      renamingItem.value = null
      await options.refreshUntilGone([oldName])
    } catch (e) {
      showToast(`${t('browserView.renameFailed')}: ${(e as Error).message}`, 'error')
    }
  }

  function confirmDelete(item: FileItem | null, closeMenu: () => void) {
    if (!item) return
    deletingItem.value = item
    showDeleteConfirm.value = true
    closeMenu()
  }

  async function executeDelete() {
    if (!deletingItem.value) return
    if (!options.browserFs()) {
      showToast(t('browserView.selectStorageFirst'), 'error')
      return
    }

    if (!options.validatePath(deletingItem.value.Path, t('browserView.deletePath'))) return

    try {
      if (deletingItem.value.IsDir) {
        await api.purgeDir(options.browserFs(), deletingItem.value.Path)
      } else {
        await api.deleteFile(options.browserFs(), deletingItem.value.Path)
      }
      showDeleteConfirm.value = false
      deletingItem.value = null
      await options.refreshBrowser()
    } catch (e) {
      showToast(`${t('browserView.deleteFailed')}: ${(e as Error).message}`, 'error')
    }
  }

  async function testRemote(name: string) {
    if (testState.value[name] === 'testing') return
    testState.value[name] = 'testing'
    try {
      await api.testRemote(name)
      testState.value[name] = 'success'
    } catch {
      testState.value[name] = 'failed'
    }
    setTimeout(() => {
      testState.value[name] = 'idle'
    }, 5000)
  }

  function getTestText(name: string) {
    const s = testState.value[name]
    if (s === 'testing') return t('browserView.testing')
    if (s === 'success') return t('browserView.testSuccess')
    if (s === 'failed') return t('browserView.testFailed')
    return t('browserView.test')
  }

  return {
    showDeleteConfirm,
    deletingItem,
    showRenameInput,
    renamingItem,
    renameInput,
    testState,
    startRename,
    confirmRename,
    confirmDelete,
    executeDelete,
    testRemote,
    getTestText
  }
}
