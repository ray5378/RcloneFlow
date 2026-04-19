import { ref } from 'vue'

export interface ToastMessage {
  id: number
  message: string
  type: 'info' | 'success' | 'error'
}

export function useToastCenter() {
  const toasts = ref<ToastMessage[]>([])
  let toastId = 0

  function showToast(message: string, type: 'info' | 'success' | 'error' = 'info') {
    const id = ++toastId
    toasts.value.push({ id, message, type })
    setTimeout(() => {
      toasts.value = toasts.value.filter(t => t.id !== id)
    }, 3000)
  }

  return {
    toasts,
    showToast,
  }
}
