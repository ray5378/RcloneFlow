import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskFormPrepare } from './useTaskFormPrepare'

describe('useTaskFormPrepare', () => {
  it('skips command parsing when command mode is off and defers to validation', () => {
    const createForm = ref({ options: { transfers: 2 } })
    const parseRcloneCommand = vi.fn()
    const validateTaskForm = vi.fn(() => '')
    const normalizeTaskOptions = vi.fn((raw) => raw || {})

    const api = useTaskFormPrepare({
      createForm,
      commandMode: ref(false),
      commandText: ref('rclone copy a: b:'),
      normalizeTaskOptions,
      parseRcloneCommand,
      validateTaskForm,
    })

    expect(api.prepareTaskFormSubmit()).toBe('')
    expect(parseRcloneCommand).not.toHaveBeenCalled()
    expect(normalizeTaskOptions).not.toHaveBeenCalled()
    expect(api.validateTaskFormBeforeSubmit()).toBe('')
    expect(validateTaskForm).toHaveBeenCalledTimes(1)
  })

  it('parses command mode input and merges normalized existing options with parsed options', () => {
    const createForm = ref({
      mode: 'sync',
      sourceRemote: '',
      sourcePath: '',
      targetRemote: '',
      targetPath: '',
      options: { raw: true },
    })
    const parseRcloneCommand = vi.fn(() => ({
      mode: 'copy',
      src: { remote: 'src', path: '/from' },
      dst: { remote: 'dst', path: '/to' },
      options: { transfers: 4, dryRun: true },
    }))
    const validateTaskForm = vi.fn(() => '')
    const normalizeTaskOptions = vi.fn(() => ({ checkers: 8, dryRun: false }))

    const api = useTaskFormPrepare({
      createForm,
      commandMode: ref(true),
      commandText: ref('rclone copy src:/from dst:/to --transfers=4'),
      normalizeTaskOptions,
      parseRcloneCommand,
      validateTaskForm,
    })

    expect(api.prepareTaskFormSubmit()).toBe('')
    expect(parseRcloneCommand).toHaveBeenCalledWith('rclone copy src:/from dst:/to --transfers=4')
    expect(createForm.value.mode).toBe('copy')
    expect(createForm.value.sourceRemote).toBe('src')
    expect(createForm.value.sourcePath).toBe('/from')
    expect(createForm.value.targetRemote).toBe('dst')
    expect(createForm.value.targetPath).toBe('/to')
    expect(createForm.value.options).toEqual({ checkers: 8, dryRun: true, transfers: 4 })

    expect(api.validateTaskFormBeforeSubmit()).toBe('')
    expect(validateTaskForm).toHaveBeenCalledTimes(1)
  })

  it('returns translated parse error and stops validation when parsing fails', () => {
    const createForm = ref({ options: {} })
    const parseRcloneCommand = vi.fn(() => {
      throw new Error('boom')
    })
    const validateTaskForm = vi.fn(() => 'should-not-run')

    const api = useTaskFormPrepare({
      createForm,
      commandMode: ref(true),
      commandText: ref('bad command'),
      normalizeTaskOptions: vi.fn((raw) => raw || {}),
      parseRcloneCommand,
      validateTaskForm,
    })

    const result = api.prepareTaskFormSubmit()
    expect(result).toContain('boom')

    const beforeSubmit = api.validateTaskFormBeforeSubmit()
    expect(beforeSubmit).toContain('boom')
    expect(validateTaskForm).not.toHaveBeenCalled()
  })
})
