import { describe, it, expect, vi } from 'vitest'
import { useRunLogModal } from './useRunLogModal'

vi.mock('../api/auth', () => ({
  getToken: () => 'test-token',
}))

vi.mock('../i18n', () => ({
  t: (key: string) => key,
}))

describe('useRunLogModal', () => {
  it('should initialize with default values', () => {
    const { showLogModal, logModalTitle, logContent } = useRunLogModal()
    expect(showLogModal.value).toBe(false)
    expect(logModalTitle.value).toBe('modal.transferLog')
    expect(logContent.value).toBe('')
  })

  it('should open log modal and fetch log', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      arrayBuffer: () => new TextEncoder().encode('log data').buffer,
    })

    const { showLogModal, logModalTitle, logContent, openRunLog } = useRunLogModal()
    await openRunLog({ id: 42 })

    expect(showLogModal.value).toBe(true)
    expect(logModalTitle.value).toBe('modal.transferLog #42')
    expect(logContent.value).toBe('log data')
  })

  it('should handle fetch failure with error message', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      text: () => 'error',
    })

    const { logContent, openRunLog } = useRunLogModal()
    await openRunLog({ id: 1 })
    expect(logContent.value).toContain('modal.loadFailed')
    expect(logContent.value).toContain('500')
  })

  it('should handle network error', async () => {
    global.fetch = vi.fn().mockRejectedValue(new Error('network error'))
    const { logContent, openRunLog } = useRunLogModal()
    await openRunLog({ id: 3 })
    expect(logContent.value).toContain('modal.loadError')
  })
})
