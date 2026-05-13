import { describe, it, expect, vi, beforeEach } from 'vitest'
import { getSchedules, createSchedule, updateSchedule, deleteSchedule } from './schedule'

vi.mock('./client', () => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  del: vi.fn(),
}))

import { get, post, put, del } from './client'

describe('schedule API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('gets schedules', async () => {
    const payload = [{ id: 1, taskId: 2 }]
    ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(payload)
    await expect(getSchedules()).resolves.toEqual(payload)
    expect(get).toHaveBeenCalledWith('/api/schedules')
  })

  it('creates schedules', async () => {
    const body = { taskId: 2, enabled: true, spec: '0 * * * *' } as any
    const created = { id: 1, ...body }
    ;(post as ReturnType<typeof vi.fn>).mockResolvedValueOnce(created)
    await expect(createSchedule(body)).resolves.toEqual(created)
    expect(post).toHaveBeenCalledWith('/api/schedules', body)
  })

  it('updates schedule without blank spec', async () => {
    await updateSchedule(3, false, '   ')
    expect(put).toHaveBeenCalledWith('/api/schedules/3', { enabled: false })
  })

  it('updates schedule with spec and deletes schedules', async () => {
    await updateSchedule(4, true, '*/5 * * * *')
    expect(put).toHaveBeenCalledWith('/api/schedules/4', { enabled: true, spec: '*/5 * * * *' })

    await deleteSchedule(9)
    expect(del).toHaveBeenCalledWith('/api/schedules/9')
  })
})
