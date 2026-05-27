import { ref } from 'vue'
import { t } from '../i18n'
import { showToast } from '../api/errors'
import * as api from '../api'

export interface UseBrowserRemoteManagementOptions {
  openRemote: (name: string) => Promise<void>
  loadRemotes: () => Promise<void>
}

export interface UseBrowserRemoteManagementReturn {
  remotes: ReturnType<typeof ref<string[]>>
  remoteOrder: ReturnType<typeof ref<string[]>>
  draggedRemote: ReturnType<typeof ref<string>>
  descriptions: ReturnType<typeof ref<Record<string, string>>>
  getOrderedRemotes: () => string[]
  saveRemoteOrder: () => Promise<void>
  loadRemoteOrder: () => Promise<void>
  onDragStart: (name: string) => void
  onDragOver: (e: DragEvent, name: string) => void
  onDrop: (e: DragEvent, targetName: string) => void
  onDragEnd: () => void
  openRemote: (name: string) => Promise<void>
  deleteRemote: (name: string) => Promise<{ show: boolean; title: string; message: string; onConfirm: () => Promise<void> }>
  saveDesc: (name: string, desc: string) => void
}

export function useBrowserRemoteManagement(options: UseBrowserRemoteManagementOptions): UseBrowserRemoteManagementReturn {
  const remotes = ref<string[]>([])
  const remoteOrder = ref<string[]>(JSON.parse(localStorage.getItem('remoteOrder') || '[]'))
  const draggedRemote = ref('')
  const descriptions = ref<Record<string, string>>(
    JSON.parse(localStorage.getItem('remoteDescriptions') || '{}')
  )

  function getOrderedRemotes() {
    const remotesList = remotes.value
    const order = remoteOrder.value
    if (!order.length) return remotesList
    const ordered = order.filter(r => remotesList.includes(r))
    const newOnes = remotesList.filter(r => !order.includes(r))
    return [...ordered, ...newOnes]
  }

  async function saveRemoteOrder() {
    remoteOrder.value = remotes.value
    try {
      await api.saveRemoteOrder(remoteOrder.value)
    } catch {
      localStorage.setItem('remoteOrder', JSON.stringify(remoteOrder.value))
    }
  }

  async function loadRemoteOrder() {
    try {
      const order = await api.getRemoteOrder()
      if (order.length > 0) {
        remoteOrder.value = order
        localStorage.setItem('remoteOrder', JSON.stringify(order))
        return
      }
    } catch {
    }
  }

  function onDragStart(name: string) {
    draggedRemote.value = name
  }

  function onDragOver(e: DragEvent, _name: string) {
    e.preventDefault()
  }

  function onDrop(e: DragEvent, targetName: string) {
    e.preventDefault()
    if (draggedRemote.value === targetName) return
    
    const list = [...remotes.value]
    const fromIndex = list.indexOf(draggedRemote.value)
    const toIndex = list.indexOf(targetName)
    
    if (fromIndex !== -1 && toIndex !== -1) {
      list.splice(fromIndex, 1)
      list.splice(toIndex, 0, draggedRemote.value)
      remotes.value = list
      saveRemoteOrder()
    }
    
    draggedRemote.value = ''
  }

  function onDragEnd() {
    draggedRemote.value = ''
  }

  async function deleteRemote(name: string) {
    return {
      show: true,
      title: t('remote.deleteStorage'),
      message: t('remote.deleteStorageConfirm').replace('{name}', name),
      onConfirm: async () => {
        try {
          await api.deleteRemote(name)
          delete descriptions.value[name]
          localStorage.setItem('remoteDescriptions', JSON.stringify(descriptions.value))
          await options.loadRemotes()
        } catch (e) {
          showToast((e as Error).message, 'error')
        }
      }
    }
  }

  function saveDesc(name: string, desc: string) {
    descriptions.value[name] = desc
    localStorage.setItem('remoteDescriptions', JSON.stringify(descriptions.value))
  }

  return {
    remotes,
    remoteOrder,
    draggedRemote,
    descriptions,
    getOrderedRemotes,
    saveRemoteOrder,
    loadRemoteOrder,
    onDragStart,
    onDragOver,
    onDrop,
    onDragEnd,
    openRemote: options.openRemote,
    deleteRemote,
    saveDesc
  }
}
