import { describe, it, expect } from 'vitest'
import { ref } from 'vue'
import { useTaskViewModalBindings } from './useTaskViewModalBindings'

describe('useTaskViewModalBindings', () => {
  const createOptions = () => ({
    showWebhookModal: ref(false),
    webhookForm: ref({ notify: {}, status: {} }),
    showSingletonModal: ref(false),
    singletonForm: ref({ singletonEnabled: false }),
    showLogModal: ref(false),
    commandMode: ref(false),
    commandText: ref(''),
    showAdvancedOptions: ref(false),
    showGlobalStatsModal: ref(false),
  })

  it('should close modals', () => {
    const opts = createOptions()
    opts.showWebhookModal.value = true
    opts.showSingletonModal.value = true
    opts.showLogModal.value = true
    opts.showGlobalStatsModal.value = true

    const { closeWebhookModal, closeSingletonModal, closeLogModal, closeGlobalStatsModal } = useTaskViewModalBindings(opts)

    closeWebhookModal()
    expect(opts.showWebhookModal.value).toBe(false)

    closeSingletonModal()
    expect(opts.showSingletonModal.value).toBe(false)

    closeLogModal()
    expect(opts.showLogModal.value).toBe(false)

    closeGlobalStatsModal()
    expect(opts.showGlobalStatsModal.value).toBe(false)
  })

  it('should set webhook form fields', () => {
    const opts = createOptions()
    const { setWebhookTriggerId, setWebhookMatchText, setWebhookPostUrl, setWebhookWecomUrl } = useTaskViewModalBindings(opts)

    setWebhookTriggerId('trigger-1')
    setWebhookMatchText('match')
    setWebhookPostUrl('http://example.com')
    setWebhookWecomUrl('http://wecom.com')

    expect(opts.webhookForm.value.triggerId).toBe('trigger-1')
    expect(opts.webhookForm.value.matchText).toBe('match')
    expect(opts.webhookForm.value.postUrl).toBe('http://example.com')
    expect(opts.webhookForm.value.wecomUrl).toBe('http://wecom.com')
  })

  it('should set webhook notify flags', () => {
    const opts = createOptions()
    opts.webhookForm.value = {} as any
    const { setWebhookNotifyManual, setWebhookNotifySchedule, setWebhookNotifyWebhook } = useTaskViewModalBindings(opts)

    setWebhookNotifyManual(true)
    setWebhookNotifySchedule(true)
    setWebhookNotifyWebhook(false)

    expect(opts.webhookForm.value.notify.manual).toBe(true)
    expect(opts.webhookForm.value.notify.schedule).toBe(true)
    expect(opts.webhookForm.value.notify.webhook).toBe(false)
  })

  it('should set webhook status flags', () => {
    const opts = createOptions()
    opts.webhookForm.value = {} as any
    const { setWebhookStatusSuccess, setWebhookStatusFailed, setWebhookStatusHasTransfer } = useTaskViewModalBindings(opts)

    setWebhookStatusSuccess(true)
    setWebhookStatusFailed(false)
    setWebhookStatusHasTransfer(true)

    expect(opts.webhookForm.value.status.success).toBe(true)
    expect(opts.webhookForm.value.status.failed).toBe(false)
    expect(opts.webhookForm.value.status.hasTransfer).toBe(true)
  })

  it('should set singleton and command fields', () => {
    const opts = createOptions()
    const { setSingletonEnabled, setCommandMode, setCommandText, setShowAdvancedOptions } = useTaskViewModalBindings(opts)

    setSingletonEnabled(true)
    setCommandMode(true)
    setCommandText('rclone copy src: dst:')
    setShowAdvancedOptions(true)

    expect(opts.singletonForm.value.singletonEnabled).toBe(true)
    expect(opts.commandMode.value).toBe(true)
    expect(opts.commandText.value).toBe('rclone copy src: dst:')
    expect(opts.showAdvancedOptions.value).toBe(true)
  })
})
