import { describe, it, expect } from 'vitest'
import { useTaskFormState } from './useTaskFormState'

describe('useTaskFormState', () => {
  it('should initialize with default create form', () => {
    const { createForm, commandMode, commandText, editingTask, showAdvancedOptions } = useTaskFormState()
    expect(createForm.value.mode).toBe('copy')
    expect(createForm.value.name).toBe('')
    expect(commandMode.value).toBe(false)
    expect(commandText.value).toBe('')
    expect(editingTask.value).toBeNull()
    expect(showAdvancedOptions.value).toBe(false)
  })

  it('should reset form for create', () => {
    const { createForm, editingTask, commandMode, commandText, showAdvancedOptions, resetTaskFormForCreate } = useTaskFormState()
    createForm.value.name = 'test'
    editingTask.value = { id: 1 } as any
    commandMode.value = true
    commandText.value = 'rclone copy'
    showAdvancedOptions.value = true
    resetTaskFormForCreate()
    expect(createForm.value.name).toBe('')
    expect(editingTask.value).toBeNull()
    expect(commandMode.value).toBe(false)
    expect(commandText.value).toBe('')
    expect(showAdvancedOptions.value).toBe(false)
  })

  it('should fill form for edit', () => {
    const { createForm, editingTask, fillTaskFormForEdit } = useTaskFormState()
    const task = {
      id: 1,
      name: 'my-task',
      mode: 'sync',
      sourceRemote: 'src',
      sourcePath: '/src',
      targetRemote: 'dst',
      targetPath: '/dst',
      options: { bwLimit: '10M' },
    } as any
    fillTaskFormForEdit(task)
    expect(createForm.value.name).toBe('my-task')
    expect(createForm.value.mode).toBe('sync')
    expect(createForm.value.sourceRemote).toBe('src')
    expect(createForm.value.targetRemote).toBe('dst')
    expect(editingTask.value?.id).toBe(1)
  })

  it('should normalize invalid mode to copy', () => {
    const { createForm, fillTaskFormForEdit } = useTaskFormState()
    fillTaskFormForEdit({ mode: 'invalid' } as any)
    expect(createForm.value.mode).toBe('copy')
  })
})
