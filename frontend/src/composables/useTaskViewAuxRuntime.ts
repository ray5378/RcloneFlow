import { useTaskWebhookConfig } from './useTaskWebhookConfig'
import { useTaskSingletonConfig } from './useTaskSingletonConfig'
import { useRunLogModal } from './useRunLogModal'
import { useTaskViewUi } from './useTaskViewUi'
import { useRunDisplayHelpers } from './useRunDisplayHelpers'

function formatScheduleSpec(spec: string): string {
  if (!spec) return ''
  const parts = spec.split('|')
  if (parts.length !== 5) return spec
  const [minute, hour, day, month, week] = parts
  return `${minute} ${hour} ${day} ${month} ${week}`
}

export function useTaskViewAuxRuntime(options: {
  loadData: () => Promise<void>
  showToast: (message: string, type?: 'info' | 'success' | 'error') => void
  taskApi: any
  getFinalSummary: (run: any) => any
}) {
  const {
    showWebhookModal,
    webhookForm,
    setWebhook,
    saveWebhook,
    testWebhook,
    regenerateWebhookSecret,
  } = useTaskWebhookConfig({
    loadData: options.loadData,
    showToast: options.showToast,
  })

  const {
    showSingletonModal,
    singletonForm,
    setSingletonMode,
    saveSingleton,
  } = useTaskSingletonConfig({
    taskApi: options.taskApi,
    loadData: options.loadData,
    showToast: options.showToast,
  })

  const {
    showLogModal,
    logModalTitle,
    logContent,
    openRunLog,
  } = useRunLogModal()

  const {
    openMenuId,
    confirmModal,
    toggleMenu,
    closeMenus,
    showConfirm,
    closeConfirm,
    confirmAndClose,
  } = useTaskViewUi()

  const {
    formatTime,
    formatDuration,
    getRunDurationText,
    getStatusClass,
    getStatusText,
  } = useRunDisplayHelpers({
    getFinalSummary: options.getFinalSummary,
  })

  return {
    showWebhookModal,
    webhookForm,
    setWebhook,
    saveWebhook,
    testWebhook,
    regenerateWebhookSecret,
    showSingletonModal,
    singletonForm,
    setSingletonMode,
    saveSingleton,
    showLogModal,
    logModalTitle,
    logContent,
    openRunLog,
    openMenuId,
    confirmModal,
    toggleMenu,
    closeMenus,
    showConfirm,
    closeConfirm,
    confirmAndClose,
    formatTime,
    formatDuration,
    getRunDurationText,
    getStatusClass,
    getStatusText,
    formatScheduleSpec,
  }
}
