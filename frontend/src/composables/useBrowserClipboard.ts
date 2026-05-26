import { ref } from 'vue'
import { t } from '../i18n'
import { showToast } from '../api/errors'
import * as api from '../api'
import type { FileItem } from '../types'

export interface UseBrowserClipboardOptions {
  browserFs: () => string
  browserPath: () => string
  refreshBrowser: () => Promise<void>
  refreshUntilGone: (names: string[], timeoutMs?: number, intervalMs?: number) => Promise<void>
}

export interface UseBrowserClipboardReturn {
  clipboardItem: ReturnType<typeof ref<FileItem | null>>
  clipboardAction: ReturnType<typeof ref<'copy' | 'move' | null>>
  clipboardSrcFs: ReturnType<typeof ref<string>>
  validatePath: (path: string, name?: string) => boolean
  copyItem: (item: FileItem) => void
  moveItem: (item: FileItem) => void
  pasteItem: () => Promise<void>
  clearClipboard: () => void
}

export function useBrowserClipboard(options: UseBrowserClipboardOptions): UseBrowserClipboardReturn {
  const clipboardItem = ref<FileItem | null>(null)
  const clipboardAction = ref<'copy' | 'move' | null>(null)
  const clipboardSrcFs = ref<string>('')

  function hasPathTraversal(path: string): boolean {
    const normalized = path.replace(/\\/g, '/')
    if (normalized.includes('..') || normalized.includes('./') || normalized === '.') {
      return true
    }
    if (normalized.includes('%2e%2e') || normalized.includes('%2e.')) {
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

  function copyItem(item: FileItem) {
    clipboardItem.value = {
      ...item,
      Path: item.Path
    }
    clipboardAction.value = 'copy'
    clipboardSrcFs.value = options.browserFs()
  }

  function moveItem(item: FileItem) {
    clipboardItem.value = {
      ...item,
      Path: item.Path
    }
    clipboardAction.value = 'move'
    clipboardSrcFs.value = options.browserFs()
  }

  async function pasteItem() {
    if (!clipboardItem.value || !clipboardAction.value) {
      showToast(t('browserView.clipboardEmpty'), 'error')
      return
    }
    if (!options.browserFs()) {
      showToast(t('browserView.selectStorageFirst'), 'error')
      return
    }

    const srcFs = clipboardSrcFs.value || options.browserFs()
    const dstFs = options.browserFs()
    const srcPath = clipboardItem.value.Path.replace(/^\/+/, '')
    const dstPath = (options.browserPath() ? options.browserPath().replace(/^\/+/, '') + '/' : '') + clipboardItem.value.Name
    if (!validatePath(srcPath, t('taskCard.source'))) return
    if (!validatePath(dstPath, t('taskCard.target'))) return

    try {
      const isDir = clipboardItem.value.IsDir
      const actionWasMove = clipboardAction.value === 'move'
      const oldName = clipboardItem.value.Name

      if (clipboardAction.value === 'copy') {
        if (isDir) {
          await api.copyDir(srcFs, srcPath, dstFs, dstPath)
        } else {
          await api.copyFile(srcFs, srcPath, dstFs, dstPath)
        }
      } else if (clipboardAction.value === 'move') {
        if (isDir) {
          await api.moveDir(srcFs, srcPath, dstFs, dstPath)
        } else {
          await api.moveFile(srcFs, srcPath, dstFs, dstPath)
        }
      }

      clipboardItem.value = null
      clipboardAction.value = null
      if (actionWasMove) {
        await options.refreshUntilGone([oldName])
      } else {
        await options.refreshBrowser()
      }
    } catch (e) {
      showToast(`${t('browserView.pasteFailed')}: ${(e as Error).message}`, 'error')
    }
  }

  function clearClipboard() {
    clipboardItem.value = null
    clipboardAction.value = null
  }

  return {
    clipboardItem,
    clipboardAction,
    clipboardSrcFs,
    copyItem,
    moveItem,
    pasteItem,
    clearClipboard,
    validatePath
  }
}
