import { describe, it, expect, vi, beforeEach, afterEach, vi as vitest } from 'vitest'
import { ref } from 'vue'
import { useTaskProgressSync } from './useTaskProgressSync'

describe('useTaskProgressSync', () => {
  const createOptions = () => ({
    runs: ref([]),
    activeRuns: ref([]),
    activeRunLookup: { getActiveRunByTaskId: vi.fn() },
    loadData: vi.fn().mockResolvedValue(undefined),
    loadActiveRuns: vi.fn().mockResolvedValue(undefined),
  })

  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('should return expected methods', () => {
    const options = createOptions()
    const sync = useTaskProgressSync(options)

    expect(sync).toHaveProperty('getRunProgressFromSummary')
    expect(sync).toHaveProperty('getRealtimeProgressByRun')
    expect(sync).toHaveProperty('getRunningProgressByRun')
    expect(sync).toHaveProperty('getTaskCardProgressByTask')
    expect(sync).toHaveProperty('getRunningProgressByTask')
    expect(sync).toHaveProperty('formatBps')
    expect(sync).toHaveProperty('calcEtaFromAvg')
    expect(sync).toHaveProperty('triggerAutoRefresh')
  })

  it('getRunProgressFromSummary should parse valid summary from JSON string', () => {
    const options = createOptions()
    const { getRunProgressFromSummary } = useTaskProgressSync(options)

    const result = getRunProgressFromSummary({
      id: 1,
      summary: JSON.stringify({
        progress: { bytes: 1024, totalBytes: 2048, speed: 512, percentage: 50, completedFiles: 5, plannedFiles: 10 },
      }),
    })

    expect(result).not.toBeNull()
    expect(result?.bytes).toBe(1024)
    expect(result?.totalBytes).toBe(2048)
    expect(result?.percentage).toBe(50)
  })

  it('getRunProgressFromSummary should parse summary from object', () => {
    const options = createOptions()
    const { getRunProgressFromSummary } = useTaskProgressSync(options)

    const result = getRunProgressFromSummary({
      id: 2,
      summary: {
        progress: { bytes: 2048, totalBytes: 4096, speed: 1024, percentage: 50, completedFiles: 5, plannedFiles: 10 },
      },
    })

    expect(result).not.toBeNull()
    expect(result?.bytes).toBe(2048)
  })

  it('getRunProgressFromSummary should return cached frame for same run', () => {
    const options = createOptions()
    const { getRunProgressFromSummary } = useTaskProgressSync(options)

    const run1 = getRunProgressFromSummary({
      id: 3,
      summary: { progress: { bytes: 1000, totalBytes: 2000, percentage: 50 } },
    })
    const run2 = getRunProgressFromSummary({ id: 3, summary: {} })

    expect(run2).toEqual(run1)
  })

  it('getRunProgressFromSummary should handle invalid JSON', () => {
    const options = createOptions()
    const { getRunProgressFromSummary } = useTaskProgressSync(options)

    const result = getRunProgressFromSummary({ id: 4, summary: 'invalid-json' })
    expect(result).toBeNull()
  })

  it('getRunProgressFromSummary should calculate percentage when missing', () => {
    const options = createOptions()
    const { getRunProgressFromSummary } = useTaskProgressSync(options)

    const result = getRunProgressFromSummary({
      id: 5,
      summary: { progress: { bytes: 500, totalBytes: 1000 } },
    })

    expect(result?.percentage).toBe(50)
  })

  it('getRealtimeProgressByRun should return active progress when available', () => {
    const options = createOptions()
    const activeProgress = {
      bytes: 2048, totalBytes: 4096, speed: 1024, percentage: 50,
      completedFiles: 10, totalCount: 20, eta: 120,
    }
    options.activeRunLookup.getActiveRunByTaskId.mockReturnValue({ progress: activeProgress })

    const { getRealtimeProgressByRun } = useTaskProgressSync(options)
    const result = getRealtimeProgressByRun({ id: 1, taskId: 100 })

    expect(result).toEqual(activeProgress)
  })

  it('getRealtimeProgressByRun should fall back to summary when active not available', () => {
    const options = createOptions()
    options.activeRunLookup.getActiveRunByTaskId.mockReturnValue(null)

    const { getRealtimeProgressByRun } = useTaskProgressSync(options)
    const result = getRealtimeProgressByRun({
      id: 2, taskId: 101,
      summary: { progress: { bytes: 500, totalBytes: 1000, percentage: 50 } },
    })

    expect(result?.bytes).toBe(500)
  })

  it('getRealtimeProgressByRun should handle taskId variations', () => {
    const options = createOptions()
    options.activeRunLookup.getActiveRunByTaskId.mockReturnValue({
      progress: { bytes: 100, totalBytes: 200, percentage: 50 },
    })

    const { getRealtimeProgressByRun } = useTaskProgressSync(options)
    const result1 = getRealtimeProgressByRun({ id: 3, taskID: 102 })
    const result2 = getRealtimeProgressByRun({ id: 4, task_id: 103 })

    expect(result1?.bytes).toBe(100)
    expect(result2?.bytes).toBe(100)
  })

  it('getRunningProgressByRun should delegate to getRealtimeProgressByRun', () => {
    const options = createOptions()
    options.activeRunLookup.getActiveRunByTaskId.mockReturnValue({
      progress: { bytes: 256, totalBytes: 512, percentage: 50 },
    })

    const { getRunningProgressByRun } = useTaskProgressSync(options)
    const result = getRunningProgressByRun({ id: 5, taskId: 104 })

    expect(result?.bytes).toBe(256)
  })

  it('getTaskCardProgressByTask should return active progress', () => {
    const options = createOptions()
    const activeProgress = {
      bytes: 1024, totalBytes: 2048, speed: 512, percentage: 50,
      completedFiles: 10, totalCount: 20, eta: 120,
    }
    options.activeRunLookup.getActiveRunByTaskId.mockReturnValue({ progress: activeProgress })

    const { getTaskCardProgressByTask } = useTaskProgressSync(options)
    const result = getTaskCardProgressByTask(200)

    expect(result?.percentage).toBe(50)
  })

  it('getTaskCardProgressByTask should freeze progress at 100%', () => {
    const options = createOptions()
    const now = Date.now()
    vi.setSystemTime(now)

    options.activeRunLookup.getActiveRunByTaskId.mockReturnValue({
      progress: { bytes: 100, totalBytes: 100, percentage: 99.9999, speed: 100 },
    })

    const { getTaskCardProgressByTask } = useTaskProgressSync(options)
    const result = getTaskCardProgressByTask(201)

    expect(result?.percentage).toBe(100)
    expect(result?.speed).toBe(0)
    expect(result?.eta).toBe(0)
    expect(result?.phase).toBe('completed')
  })

  it('getTaskCardProgressByTask should use frozen frame for completed in window', () => {
    const options = createOptions()
    const now = Date.now()
    vi.setSystemTime(now)

    options.runs.value = [{ id: 1, taskId: 202, status: 'running' }]
    options.activeRunLookup.getActiveRunByTaskId.mockReturnValue({
      progress: { bytes: 100, totalBytes: 100, percentage: 99.9999 },
    })

    const { getTaskCardProgressByTask } = useTaskProgressSync(options)
    const result1 = getTaskCardProgressByTask(202)

    options.activeRunLookup.getActiveRunByTaskId.mockReturnValue(null)
    const result2 = getTaskCardProgressByTask(202)

    expect(result2).toEqual(result1)
  })

  it('getTaskCardProgressByTask should clear frozen frame after window', () => {
    const options = createOptions()
    const now = Date.now()
    vi.setSystemTime(now)

    options.runs.value = [{ id: 1, taskId: 203, status: 'running' }]
    options.activeRunLookup.getActiveRunByTaskId.mockReturnValue({
      progress: { bytes: 100, totalBytes: 100, percentage: 99.9999 },
    })

    const { getTaskCardProgressByTask } = useTaskProgressSync(options)
    getTaskCardProgressByTask(203)

    options.activeRunLookup.getActiveRunByTaskId.mockReturnValue(null)
    vi.setSystemTime(now + 16000)
    const result = getTaskCardProgressByTask(203)

    expect(result).toBeNull()
  })

  it('getTaskCardProgressByTask should clear frozen frame when task no longer running', () => {
    const options = createOptions()
    const now = Date.now()
    vi.setSystemTime(now)

    options.runs.value = [{ id: 1, taskId: 204, status: 'running' }]
    options.activeRunLookup.getActiveRunByTaskId.mockReturnValue({
      progress: { bytes: 100, totalBytes: 100, percentage: 99.9999 },
    })

    const { getTaskCardProgressByTask } = useTaskProgressSync(options)
    getTaskCardProgressByTask(204)

    options.activeRunLookup.getActiveRunByTaskId.mockReturnValue(null)
    options.runs.value = []
    const result = getTaskCardProgressByTask(204)

    expect(result).toBeNull()
  })

  it('getTaskCardProgressByTask should use summary progress when no frozen frame', () => {
    const options = createOptions()
    options.runs.value = [{
      id: 1, taskId: 205, status: 'running',
      summary: { progress: { bytes: 500, totalBytes: 1000, percentage: 50 } },
    }]

    const { getTaskCardProgressByTask } = useTaskProgressSync(options)
    const result = getTaskCardProgressByTask(205)

    expect(result?.percentage).toBe(50)
  })

  it('getTaskCardProgressByTask returns null for non-existent task', () => {
    const options = createOptions()
    const { getTaskCardProgressByTask } = useTaskProgressSync(options)

    const result = getTaskCardProgressByTask(999)
    expect(result).toBeNull()
  })

  it('getRunningProgressByTask should clamp percentage between 0 and 100', () => {
    const options = createOptions()
    options.runs.value = [{
      id: 1, taskId: 300, status: 'running',
      summary: { progress: { bytes: 100, totalBytes: 100, percentage: 120 } },
    }]

    const { getRunningProgressByTask } = useTaskProgressSync(options)
    const result1 = getRunningProgressByTask(300)
    expect(result1?.percentage).toBe(100)

    options.runs.value = [{
      id: 2, taskId: 301, status: 'running',
      summary: { progress: { bytes: 0, totalBytes: 100, percentage: -10 } },
    }]
    const result2 = getRunningProgressByTask(301)
    expect(result2?.percentage).toBe(0)
  })

  it('formatBps should format bytes per second', () => {
    const options = createOptions()
    const { formatBps } = useTaskProgressSync(options)

    expect(formatBps(0)).toBe('-')
    expect(formatBps(1024)).toBe('1.0 KB/s')
    expect(formatBps(1048576)).toBe('1.0 MB/s')
  })

  it('calcEtaFromAvg should calculate ETA', () => {
    const options = createOptions()
    const { calcEtaFromAvg } = useTaskProgressSync(options)

    const eta = calcEtaFromAvg(
      { startedAt: new Date().toISOString(), taskId: 400 },
      { bytes: 500, totalBytes: 1000, speed: 50, percentage: 50, completedFiles: 1, totalCount: 2, eta: 0 },
    )

    expect(eta).toBe(10)
  })

  it('calcEtaFromAvg returns null for invalid inputs', () => {
    const options = createOptions()
    const { calcEtaFromAvg } = useTaskProgressSync(options)

    expect(calcEtaFromAvg(null as any, null)).toBeNull()
    expect(calcEtaFromAvg({ startedAt: null }, null)).toBeNull()
    expect(calcEtaFromAvg({ startedAt: new Date().toISOString() }, null)).toBeNull()
    expect(calcEtaFromAvg(
      { startedAt: new Date().toISOString() },
      { bytes: 0, totalBytes: 1000, speed: 50, percentage: 0, completedFiles: 0, totalCount: 0, eta: 0 },
    )).toBeNull()
    expect(calcEtaFromAvg(
      { startedAt: new Date().toISOString() },
      { bytes: 500, totalBytes: 0, speed: 50, percentage: 0, completedFiles: 0, totalCount: 0, eta: 0 },
    )).toBeNull()
  })

  it('calcEtaFromAvg returns null for very large ETA', () => {
    const options = createOptions()
    const { calcEtaFromAvg } = useTaskProgressSync(options)

    const eta = calcEtaFromAvg(
      { startedAt: new Date().toISOString(), taskId: 402 },
      { bytes: 1, totalBytes: 1000000000000, speed: 1, percentage: 0, completedFiles: 1, totalCount: 2, eta: 0 },
    )

    expect(eta).toBeNull()
  })

  it('triggerAutoRefresh should be a function', () => {
    const options = createOptions()
    const { triggerAutoRefresh } = useTaskProgressSync(options)
    expect(typeof triggerAutoRefresh).toBe('function')
  })
})
