import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskFormPrepare } from './useTaskFormPrepare'

vi.mock('../i18n', () => ({
  t: (key: string) => key,
}))

describe('useTaskFormPrepare', () => {
  function makeOptions() {
    return {
      createForm: ref({
        mode: 'copy',
        sourceRemote: '',
        sourcePath: '',
        targetRemote: '',
        targetPath: '',
        options: {},
      }),
      commandMode: ref(false),
      commandText: ref(''),
      normalizeTaskOptions: vi.fn((raw) => raw || {}),
      parseRcloneCommand: vi.fn().mockReturnValue({
        mode: 'sync',
        src: { remote: 'src', path: '/src' },
        dst: { remote: 'dst', path: '/dst' },
        options: { bwLimit: '10M' },
      }),
      validateTaskForm: vi.fn().mockReturnValue(''),
    }
  }

  it('should return empty when not in command mode', () => {
    const opts = makeOptions()
    const { prepareTaskFormSubmit } = useTaskFormPrepare(opts)
    expect(prepareTaskFormSubmit()).toBe('')
    expect(opts.parseRcloneCommand).not.toHaveBeenCalled()
  })

  it('should parse command and fill form', () => {
    const opts = makeOptions()
    opts.commandMode.value = true
    opts.commandText.value = 'rclone sync src:/src dst:/dst --bwlimit 10M'
    const { prepareTaskFormSubmit } = useTaskFormPrepare(opts)
    const result = prepareTaskFormSubmit()
    expect(result).toBe('')
    expect(opts.createForm.value.mode).toBe('sync')
    expect(opts.createForm.value.sourceRemote).toBe('src')
    expect(opts.createForm.value.targetRemote).toBe('dst')
  })

  it('should return error on parse failure', () => {
    const opts = makeOptions()
    opts.commandMode.value = true
    opts.parseRcloneCommand.mockImplementationOnce(() => { throw new Error('bad command') })
    const { prepareTaskFormSubmit } = useTaskFormPrepare(opts)
    const result = prepareTaskFormSubmit()
    expect(result).toContain('runtime.commandParseFailed')
  })

  it('should validate after prepare', () => {
    const opts = makeOptions()
    opts.commandMode.value = true
    const { validateTaskFormBeforeSubmit } = useTaskFormPrepare(opts)
    const result = validateTaskFormBeforeSubmit()
    expect(result).toBe('')
    expect(opts.validateTaskForm).toHaveBeenCalled()
  })

  it('should return prepare error before validation', () => {
    const opts = makeOptions()
    opts.commandMode.value = true
    opts.parseRcloneCommand.mockImplementationOnce(() => { throw new Error('bad') })
    const { validateTaskFormBeforeSubmit } = useTaskFormPrepare(opts)
    const result = validateTaskFormBeforeSubmit()
    expect(result).toContain('runtime.commandParseFailed')
  })
})
