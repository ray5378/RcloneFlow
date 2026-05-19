import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskViewRefreshLifecycle } from './useTaskViewRefreshLifecycle'

describe('useTaskViewRefreshLifecycle', () => {
  const createOptions = () => ({
    tasks: ref([]),
    activeRuns: ref([]),
    currentModule: ref<'history' | 'add' | 'tasks'>('tasks'),
    getRunningProgressByTask: vi.fn().mockReturnValue(null),
    loadData: vi.fn().mockResolvedValue(undefined),
    loadActiveRuns: vi.fn().mockResolvedValue(undefined),
    setupRealtimeSync: vi.fn(),
    stuckMs: 30000,
  })

  it('should return empty object', () => {
    const options = createOptions()
    const result = useTaskViewRefreshLifecycle(options)
    expect(result).toEqual({})
  })
})
