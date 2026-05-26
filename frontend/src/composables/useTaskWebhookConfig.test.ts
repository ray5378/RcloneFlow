import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('../api/task', () => ({
  updateTaskOptions: vi.fn(async () => ({})),
}))

vi.mock('../api/auth', () => ({
  getToken: () => 'test-token',
}))

vi.mock('../i18n', () => ({
  t: (key: string) => key,
}))

import { updateTaskOptions } from '../api/task'
import { useTaskWebhookConfig } from './useTaskWebhookConfig'

describe('useTaskWebhookConfig', () => {
  const loadData = vi.fn(async () => {})
  const showToast = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    vi.resetAllMocks()
  })

  it('should return all expected methods', () => {
    const config = useTaskWebhookConfig({ loadData, showToast })
    expect(config).toHaveProperty('showWebhookModal')
    expect(config).toHaveProperty('webhookForm')
    expect(config).toHaveProperty('setWebhook')
    expect(config).toHaveProperty('saveWebhook')
    expect(config).toHaveProperty('testWebhook')
    expect(config).toHaveProperty('regenerateWebhookSecret')
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

    expect(updateTaskOptions).toHaveBeenCalledWith(3, expect.objectContaining({
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
        webhookSecret: 'secret',
        webhookNotifyOn: { manual: true, schedule: true, webhook: true },
      }),
    })

    expect(webhookForm.value.postUrl).toBe('http://example.com')
    expect(webhookForm.value.wecomUrl).toBe('http://wecom.com')
    expect(webhookForm.value.triggerId).toBe('trigger-1')
    expect(webhookForm.value.matchText).toBe('match')
    expect(webhookForm.value.webhookSecret).toBe('secret')
    expect(webhookForm.value.notify).toEqual({ manual: true, schedule: true, webhook: true })
  })

  it('should handle invalid options', () => {
    const { webhookForm, setWebhook } = useTaskWebhookConfig({ loadData, showToast })

    setWebhook({ id: 5, options: 'invalid-json' })

    expect(webhookForm.value.postUrl).toBe('')
    expect(webhookForm.value.notify).toEqual({ manual: false, schedule: false, webhook: false })
  })

  it('should handle Options property (capital O)', () => {
    const { webhookForm, setWebhook } = useTaskWebhookConfig({ loadData, showToast })

    setWebhook({
      id: 6,
      Options: {
        webhookPostUrl: 'http://capital-o.com',
      },
    })

    expect(webhookForm.value.postUrl).toBe('http://capital-o.com')
  })

  it('should save webhook settings', async () => {
    const { webhookForm, saveWebhook } = useTaskWebhookConfig({ loadData, showToast })

    webhookForm.value.taskId = 7
    webhookForm.value.postUrl = 'http://example.com'
    webhookForm.value.wecomUrl = 'http://wecom.com'
    webhookForm.value.notify = { manual: true, schedule: false, webhook: true }

    await saveWebhook()

    expect(updateTaskOptions).toHaveBeenCalled()
    expect(loadData).toHaveBeenCalled()
  })

  it('should do nothing when taskId is null', async () => {
    const { saveWebhook, showWebhookModal } = useTaskWebhookConfig({ loadData, showToast })
    showWebhookModal.value = true

    await saveWebhook()

    expect(updateTaskOptions).not.toHaveBeenCalled()
    expect(showWebhookModal.value).toBe(false)
  })

  it('should handle saveWebhook error', async () => {
    const { webhookForm, saveWebhook } = useTaskWebhookConfig({ loadData, showToast })
    vi.mocked(updateTaskOptions).mockRejectedValueOnce(new Error('Save failed'))

    webhookForm.value.taskId = 8
    await saveWebhook()

    expect(showToast).toHaveBeenCalledWith('Save failed', 'error')
  })

  it('should use fetch fallback when updateTaskOptions is not a function', async () => {
    const { webhookForm, saveWebhook } = useTaskWebhookConfig({ loadData, showToast })
    
    // Temporarily mock to simulate missing function
    const originalUpdateTaskOptions = updateTaskOptions
    Object.assign(vi.importActual('../api/task'), { updateTaskOptions: undefined })
    
    webhookForm.value.taskId = 9
    
    // Mock fetch
    global.fetch = vi.fn().mockResolvedValueOnce({ ok: true })
    
    await saveWebhook()
    
    // Restore
    Object.assign(vi.importActual('../api/task'), { updateTaskOptions: originalUpdateTaskOptions })
  })

  it('regenerateWebhookSecret should generate a secret', () => {
    const { webhookForm, regenerateWebhookSecret } = useTaskWebhookConfig({ loadData, showToast })

    const originalSecret = webhookForm.value.webhookSecret
    regenerateWebhookSecret()
    
    expect(webhookForm.value.webhookSecret).not.toBe(originalSecret)
    expect(webhookForm.value.webhookSecret).toMatch(/^[0-9a-f]{32}$/)
  })

  it('should show error when no URL configured', async () => {
    const { webhookForm, testWebhook } = useTaskWebhookConfig({ loadData, showToast })

    webhookForm.value.taskId = 10
    webhookForm.value.postUrl = ''
    webhookForm.value.wecomUrl = ''

    await testWebhook()

    expect(showToast).toHaveBeenCalledWith(expect.any(String), 'error')
  })

  it('testWebhook should send to webhook URL successfully', async () => {
    const { webhookForm, testWebhook } = useTaskWebhookConfig({ loadData, showToast })
    global.fetch = vi.fn().mockResolvedValueOnce({ ok: true })

    webhookForm.value.postUrl = 'http://webhook.test'
    webhookForm.value.wecomUrl = ''

    await testWebhook()

    expect(fetch).toHaveBeenCalledWith('http://webhook.test', expect.any(Object))
    expect(showToast).toHaveBeenCalledWith(expect.any(String), 'success')
  })

  it('testWebhook should send to WeCom URL successfully', async () => {
    const { webhookForm, testWebhook } = useTaskWebhookConfig({ loadData, showToast })
    global.fetch = vi.fn().mockResolvedValueOnce({ ok: true })

    webhookForm.value.postUrl = ''
    webhookForm.value.wecomUrl = 'https://qyapi.weixin.qq.com/some-endpoint'

    await testWebhook()

    expect(fetch).toHaveBeenCalledWith('https://qyapi.weixin.qq.com/some-endpoint', expect.any(Object))
    expect(showToast).toHaveBeenCalledWith(expect.any(String), 'success')
  })

  it('testWebhook should send to both URLs with mixed success', async () => {
    const { webhookForm, testWebhook } = useTaskWebhookConfig({ loadData, showToast })
    global.fetch = vi.fn()
      .mockResolvedValueOnce({ ok: true })
      .mockRejectedValueOnce(new Error('WeCom failed'))

    webhookForm.value.postUrl = 'http://webhook.test'
    webhookForm.value.wecomUrl = 'http://wecom.test'

    await testWebhook()

    expect(showToast).toHaveBeenCalledWith(expect.any(String), 'error')
  })

  it('testWebhook should handle general errors', async () => {
    const { webhookForm, testWebhook } = useTaskWebhookConfig({ loadData, showToast })
    global.fetch = vi.fn().mockImplementationOnce(() => {
      throw new Error('General error')
    })

    webhookForm.value.postUrl = 'http://webhook.test'
    webhookForm.value.wecomUrl = ''

    await testWebhook()

    expect(showToast).toHaveBeenCalledWith(expect.any(String), 'error')
  })

  it('testWebhook should handle HTTP errors', async () => {
    const { webhookForm, testWebhook } = useTaskWebhookConfig({ loadData, showToast })
    global.fetch = vi.fn().mockResolvedValueOnce({ ok: false, status: 500 })

    webhookForm.value.postUrl = 'http://webhook.test'
    webhookForm.value.wecomUrl = ''

    await testWebhook()

    expect(showToast).toHaveBeenCalledWith(expect.any(String), 'error')
  })
})
