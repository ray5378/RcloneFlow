import { ref } from 'vue'

interface ConfirmModalState {
  show: boolean
  title: string
  message: string
  onConfirm: (() => void) | null
}

export function useTaskViewUi() {
  const openMenuId = ref<number | null>(null)
  const confirmModal = ref<ConfirmModalState>({
    show: false,
    title: '',
    message: '',
    onConfirm: null,
  })

  function toggleMenu(id: number) {
    openMenuId.value = openMenuId.value === id ? null : id
  }

  function closeMenus() {
    openMenuId.value = null
  }

  function showConfirm(title: string, message: string, onConfirm: () => void) {
    confirmModal.value = { show: true, title, message, onConfirm }
  }

  function closeConfirm() {
    confirmModal.value = { show: false, title: '', message: '', onConfirm: null }
  }

  function confirmAndClose() {
    const onConfirm = confirmModal.value.onConfirm
    closeConfirm()
    onConfirm?.()
  }

  return {
    openMenuId,
    confirmModal,
    toggleMenu,
    closeMenus,
    showConfirm,
    closeConfirm,
    confirmAndClose,
  }
}
