import { describe, it, expect, vi, beforeEach } from 'vitest'
import { getActiveTransfer, getActiveTransferCompleted, getActiveTransferPending } from './activeTransfer'

vi.mock('./client', () => ({
  get: vi.fn(),
}))

import { get } from './client'

describe('activeTransfer API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('gets active transfer overview', async () => {
    const payload = { taskId: 1, runId: 2, trackingMode: 'cas', summary: { completedCount: 1 } }
    ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(payload)

    const result = await getActiveTransfer(1)

    expect(get).toHaveBeenCalledWith('/api/tasks/1/active-transfer')
    expect(result).toEqual(payload)
  })

  it('gets completed list with paging', async () => {
    const payload = { total: 1, items: [{ name: 'a', status: 'copied' }] }
    ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(payload)

    const result = await getActiveTransferCompleted(7, 5, 9)

    expect(get).toHaveBeenCalledWith('/api/tasks/7/active-transfer/completed?offset=5&limit=9')
    expect(result).toEqual(payload)
  })

  it('gets pending list with defaults', async () => {
    const payload = { total: 0, items: [] }
    ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(payload)

    const result = await getActiveTransferPending(8)

    expect(get).toHaveBeenCalledWith('/api/tasks/8/active-transfer/pending?offset=0&limit=100')
    expect(result).toEqual(payload)
  })

  it('maps known backend errors to localized messages', async () => {
    ;(get as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('active run not found'))
    await expect(getActiveTransfer(1)).rejects.toThrow('当前没有运行中的任务')

    ;(get as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('active transfer not found'))
    await expect(getActiveTransferCompleted(1)).rejects.toThrow('当前没有可恢复的传输状态')

    ;(get as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('invalid task id'))
    await expect(getActiveTransferPending(0)).rejects.toThrow('任务 ID 无效')
  })

  it('rethrows unknown errors', async () => {
    ;(get as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('boom'))
    await expect(getActiveTransfer(3)).rejects.toThrow('boom')
  })
})
