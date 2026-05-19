import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('../api', () => ({
  updateTaskOptions: vi.fn(async () => ({})),
}))

vi.mock('../api/auth', () => ({
  getToken: () => 'test-token',
}))

vi.mock('../i18n', () => ({
  t: (key: string) => key,
}))

import * as api from '../api'
import { useTaskWebhookConfig } from './useTaskWebhookConfig'

describe('useTaskWebhookConfig', () => {
  const loadData = vi.fn(async () => {})
  const showToast = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('defaults hasTransfer to false and hydrates from task options', () => {
    const { webhookForm, setWebhook } = useTaskWebhookConfig({ loadData, showToast })

    expect(webhookForm.value.status).toEqual({ success: true, failed: true, hasTransfer: false })

    setWebhook({
      id: 1,
      options: {
        webhookNotifyStatus: {
          success: false,
          failed: true,
          hasTransfer: true,
        },
      },
    })

    expect(webhookForm.value.status).toEqual({ success: false, failed: true, hasTransfer: true })
  })

  it('falls back hasTransfer to false for legacy task options', () => {
    const { webhookForm, setWebhook } = useTaskWebhookConfig({ loadData, showToast })

    setWebhook({
      id: 2,
      options: {
        webhookNotifyStatus: {
          success: true,
          failed: false,
        },
      },
    })

    expect(webhookForm.value.status).toEqual({ success: true, failed: false, hasTransfer: false })
  })

  it('saves hasTransfer in webhookNotifyStatus payload', async () => {
    const { webhookForm, saveWebhook } = useTaskWebhookConfig({ loadData, showToast })

    webhookForm.value.taskId = 3
    webhookForm.value.triggerId = 'abc'
    webhookForm.value.status = { success: true, failed: false, hasTransfer: true }

    await saveWebhook()

    expect((api as any).updateTaskOptions).toHaveBeenCalledWith(3, expect.objectContaining({
      webhookNotifyStatus: {
        success: true,
        failed: false,
        hasTransfer: true,
      },
    }))
  })

  it('should open and close modal', () => {
    const { showWebhookModal, setWebhook } = useTaskWebhookConfig({ loadData, showToast })

    expect(showWebhookModal.value).toBe(false)

    setWebhook({ id: 1, options: {} })
    expect(showWebhookModal.value).toBe(true)
  })

  it('should handle string options', () => {
    const { webhookForm, setWebhook } = useTaskWebhookConfig({ loadData, showToast })

    setWebhook({
      id: 4,
      options: JSON.stringify({
        webhookPostUrl: 'http://example.com',
        wecomPostUrl: 'http://wecom.com',
        webhookId: 'trigger-1',
        webhookMatchText: 'match',
      }),
    })

    expect(webhookForm.value.postUrl).toBe('http://example.com')
    expect(webhookForm.value.wecomUrl).toBe('http://wecom.com')
    expect(webhookForm.value.triggerId).toBe('trigger-1')
    expect(webhookForm.value.matchText).toBe('match')
  })

  it('should handle invalid options', () => {
    const { webhookForm, setWebhook } = useTaskWebhookConfig({ loadData, showToast })

    setWebhook({ id: 5, options: 'invalid-json' })

    expect(webhookForm.value.postUrl).toBe('')
    expect(webhookForm.value.notify).toEqual({ manual: false, schedule: false, webhook: false })
  })

  it('should save webhook settings', async () => {
    const { webhookForm, saveWebhook } = useTaskWebhookConfig({ loadData, showToast })

    webhookForm.value.taskId = 6
    webhookForm.value.postUrl = 'http://example.com'
    webhookForm.value.notify = { manual: true, schedule: false, webhook: true }

    await saveWebhook()

    expect((api as any).updateTaskOptions).toHaveBeenCalled()
    expect(loadData).toHaveBeenCalled()
  })

  it('should do nothing when taskId is null', async () => {
    const { saveWebhook } = useTaskWebhookConfig({ loadData, showToast })

    await saveWebhook()

    expect((api as any).updateTaskOptions).not.toHaveBeenCalled()
  })

  it('should show error when no URL configured', async () => {
    const { webhookForm, testWebhook } = useTaskWebhookConfig({ loadData, showToast })

    webhookForm.value.taskId = 7
    webhookForm.value.postUrl = ''
    webhookForm.value.wecomUrl = ''

    await testWebhook()

    expect(showToast).toHaveBeenCalledWith(expect.any(String), 'error')
  })
})
