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
})
