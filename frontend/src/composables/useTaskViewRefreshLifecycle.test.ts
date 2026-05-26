import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useTaskViewRefreshLifecycle } from './useTaskViewRefreshLifecycle'
import { ref } from 'vue'

describe('useTaskViewRefreshLifecycle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('should be a function', () => {
    expect(typeof useTaskViewRefreshLifecycle).toBe('function')
  })

  it('should accept correct options', () => {
    const loadData = vi.fn().mockResolvedValue(undefined)
    const loadActiveRuns = vi.fn().mockResolvedValue(undefined)
    const getRunningProgressByTask = vi.fn().mockReturnValue({ percentage: 0, completedFiles: 0 })

    expect(() => {
      useTaskViewRefreshLifecycle({
        tasks: ref([]),
        activeRuns: ref([]),
        getRunningProgressByTask,
        loadData,
        loadActiveRuns,
        stuckMs: 25000,
      })
    }).not.toThrow()
  })

  it('should handle currentModule option', () => {
    const loadData = vi.fn().mockResolvedValue(undefined)
    const loadActiveRuns = vi.fn().mockResolvedValue(undefined)
    const getRunningProgressByTask = vi.fn().mockReturnValue({ percentage: 0, completedFiles: 0 })

    expect(() => {
      useTaskViewRefreshLifecycle({
        tasks: ref([]),
        activeRuns: ref([]),
        currentModule: ref<'tasks' | 'history' | 'add'>('tasks'),
        getRunningProgressByTask,
        loadData,
        loadActiveRuns,
        stuckMs: 25000,
      })
    }).not.toThrow()
  })

  it('should handle setupRealtimeSync option', () => {
    const loadData = vi.fn().mockResolvedValue(undefined)
    const loadActiveRuns = vi.fn().mockResolvedValue(undefined)
    const getRunningProgressByTask = vi.fn().mockReturnValue({ percentage: 0, completedFiles: 0 })
    const setupRealtimeSync = vi.fn()

    expect(() => {
      useTaskViewRefreshLifecycle({
        tasks: ref([]),
        activeRuns: ref([]),
        getRunningProgressByTask,
        loadData,
        loadActiveRuns,
        setupRealtimeSync,
        stuckMs: 25000,
      })
    }).not.toThrow()
  })

  it('should return empty object', () => {
    const loadData = vi.fn().mockResolvedValue(undefined)
    const loadActiveRuns = vi.fn().mockResolvedValue(undefined)
    const getRunningProgressByTask = vi.fn().mockReturnValue({ percentage: 0, completedFiles: 0 })

    const result = useTaskViewRefreshLifecycle({
      tasks: ref([]),
      activeRuns: ref([]),
      getRunningProgressByTask,
      loadData,
      loadActiveRuns,
      stuckMs: 25000,
    })

    expect(result).toEqual({})
  })
})
