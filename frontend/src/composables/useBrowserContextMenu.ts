import { ref } from 'vue'
import type { FileItem } from '../types'

export interface UseBrowserContextMenuReturn {
  contextMenu: ReturnType<typeof ref<{ show: boolean; x: number; y: number; item: FileItem | null }>>
  showContextMenuHandler: (e: MouseEvent, item: FileItem) => void
  showBackgroundMenuHandler: (e: MouseEvent) => void
  closeContextMenu: () => void
}

export function useBrowserContextMenu(): UseBrowserContextMenuReturn {
  const contextMenu = ref({
    show: false,
    x: 0,
    y: 0,
    item: null as FileItem | null
  })

  function showContextMenuHandler(e: MouseEvent, item: FileItem) {
    e.preventDefault()
    contextMenu.value = {
      show: true,
      x: e.clientX,
      y: e.clientY,
      item
    }
  }

  function showBackgroundMenuHandler(e: MouseEvent) {
    e.preventDefault()
    contextMenu.value = {
      show: true,
      x: e.clientX,
      y: e.clientY,
      item: null
    }
  }

  function closeContextMenu() {
    contextMenu.value.show = false
  }

  return {
    contextMenu,
    showContextMenuHandler,
    showBackgroundMenuHandler,
    closeContextMenu
  }
}
