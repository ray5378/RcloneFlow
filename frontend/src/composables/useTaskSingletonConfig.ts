import { ref } from 'vue'

export function useTaskSingletonConfig(options: {
  taskApi: { updateOptions: (id: number, options: Record<string, any>) => Promise<boolean> | Promise<any> }
  loadData: () => Promise<void>
  showToast: (message: string, type?: 'info' | 'success' | 'error') => void
}) {
  const showSingletonModal = ref(false)
  const singletonForm = ref<{ taskId: number | null; singletonEnabled: boolean }>({
    taskId: null,
    singletonEnabled: false,
  })

  function setSingletonMode(task: any) {
    singletonForm.value.taskId = task.id
    try {
      const opts = (task.options || task.Options || {}) as any
      singletonForm.value.singletonEnabled = !!opts.singletonMode
    } catch {
      singletonForm.value.singletonEnabled = false
    }
    showSingletonModal.value = true
  }

  async function saveSingleton() {
    if (!singletonForm.value.taskId) {
      showSingletonModal.value = false
      return
    }
    try {
      const id = singletonForm.value.taskId
      const opts = { singletonMode: singletonForm.value.singletonEnabled }
      await options.taskApi.updateOptions(id, opts)
      showSingletonModal.value = false
      await options.loadData()
    } catch (e: any) {
      options.showToast(e?.message || String(e), 'error')
    }
  }

  return {
    showSingletonModal,
    singletonForm,
    setSingletonMode,
    saveSingleton,
  }
}
