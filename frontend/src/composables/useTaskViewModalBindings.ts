import type { Ref } from 'vue'

export function useTaskViewModalBindings(options: {
  showWebhookModal: Ref<boolean>
  webhookForm: Ref<any>
  showSingletonModal: Ref<boolean>
  singletonForm: Ref<any>
  showLogModal: Ref<boolean>
  commandMode: Ref<boolean>
  commandText: Ref<string>
  showAdvancedOptions: Ref<boolean>
  showGlobalStatsModal: Ref<boolean>
}) {
  const closeWebhookModal = () => { options.showWebhookModal.value = false }
  const closeSingletonModal = () => { options.showSingletonModal.value = false }
  const closeLogModal = () => { options.showLogModal.value = false }
  const closeGlobalStatsModal = () => { options.showGlobalStatsModal.value = false }

  function ensureWebhookFormShape() {
    if (!options.webhookForm.value.notify) {
      options.webhookForm.value.notify = { manual: false, schedule: false, webhook: false }
    }
    if (!options.webhookForm.value.status) {
      options.webhookForm.value.status = { success: true, failed: true, hasTransfer: false }
      return
    }
    if (typeof options.webhookForm.value.status.hasTransfer !== 'boolean') {
      options.webhookForm.value.status.hasTransfer = false
    }
  }

  const setWebhookTriggerId = (value: string) => { options.webhookForm.value.triggerId = value }
  const setWebhookMatchText = (value: string) => { options.webhookForm.value.matchText = value }
  const setWebhookPostUrl = (value: string) => { options.webhookForm.value.postUrl = value }
  const setWebhookWecomUrl = (value: string) => { options.webhookForm.value.wecomUrl = value }
  const setWebhookNotifyManual = (value: boolean) => { ensureWebhookFormShape(); options.webhookForm.value.notify.manual = value }
  const setWebhookNotifySchedule = (value: boolean) => { ensureWebhookFormShape(); options.webhookForm.value.notify.schedule = value }
  const setWebhookNotifyWebhook = (value: boolean) => { ensureWebhookFormShape(); options.webhookForm.value.notify.webhook = value }
  const setWebhookStatusSuccess = (value: boolean) => { ensureWebhookFormShape(); options.webhookForm.value.status.success = value }
  const setWebhookStatusFailed = (value: boolean) => { ensureWebhookFormShape(); options.webhookForm.value.status.failed = value }
  const setWebhookStatusHasTransfer = (value: boolean) => { ensureWebhookFormShape(); options.webhookForm.value.status.hasTransfer = value }

  const setSingletonEnabled = (value: boolean) => { options.singletonForm.value.singletonEnabled = value }
  const setCommandMode = (value: boolean) => { options.commandMode.value = value }
  const setCommandText = (value: string) => { options.commandText.value = value }
  const setShowAdvancedOptions = (value: boolean) => { options.showAdvancedOptions.value = value }

  return {
    closeWebhookModal,
    closeSingletonModal,
    closeLogModal,
    closeGlobalStatsModal,
    setWebhookTriggerId,
    setWebhookMatchText,
    setWebhookPostUrl,
    setWebhookWecomUrl,
    setWebhookNotifyManual,
    setWebhookNotifySchedule,
    setWebhookNotifyWebhook,
    setWebhookStatusSuccess,
    setWebhookStatusFailed,
    setWebhookStatusHasTransfer,
    setSingletonEnabled,
    setCommandMode,
    setCommandText,
    setShowAdvancedOptions,
  }
}
