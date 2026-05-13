import { describe, it, expect } from 'vitest'
import { useTaskFormState } from './useTaskFormState'

describe('useTaskFormState', () => {
  it('starts with expected create defaults and resets back to them', () => {
    const api = useTaskFormState()

    expect(api.createForm.value).toMatchObject({
      name: '',
      mode: 'copy',
      sourceRemote: '',
      sourcePath: '',
      targetRemote: '',
      targetPath: '',
      enableSchedule: false,
      scheduleMonth: '*',
      scheduleHour: '00',
      scheduleMinute: '00',
      options: { enableStreaming: true },
    })

    api.commandMode.value = true
    api.commandText.value = 'rclone copy a: b:'
    api.showAdvancedOptions.value = true
    api.editingTask.value = { id: 1 } as any
    api.createForm.value.name = 'changed'

    api.resetTaskFormForCreate()

    expect(api.commandMode.value).toBe(false)
    expect(api.commandText.value).toBe('')
    expect(api.showAdvancedOptions.value).toBe(false)
    expect(api.editingTask.value).toBeNull()
    expect(api.createForm.value.name).toBe('')
    expect(api.createForm.value.options).toEqual({ enableStreaming: true })
  })

  it('fills edit form with schedule parts and normalized options', () => {
    const api = useTaskFormState()
    const task = {
      id: 9,
      name: 'task-9',
      mode: 'move',
      sourceRemote: 'src',
      sourcePath: '/a',
      targetRemote: 'dst',
      targetPath: '/b',
      options: {
        exclude: ['foo', 'bar'],
        enableStreaming: false,
      },
    } as any

    const parts = api.fillTaskFormForEdit(task, '05|*|1|2|3')

    expect(parts).toEqual(['05', '*', '1', '2', '3'])
    expect(api.editingTask.value).toEqual(task)
    expect(api.commandMode.value).toBe(false)
    expect(api.commandText.value).toBe('')
    expect(api.showAdvancedOptions.value).toBe(false)
    expect(api.createForm.value).toMatchObject({
      name: 'task-9',
      mode: 'move',
      sourceRemote: 'src',
      sourcePath: '/a',
      targetRemote: 'dst',
      targetPath: '/b',
      enableSchedule: true,
      scheduleMinute: '05',
      scheduleHour: '*',
      scheduleDay: '1',
      scheduleMonth: '2',
      scheduleWeek: '3',
      options: {
        exclude: 'foo\nbar',
        enableStreaming: false,
      },
    })
  })

  it('fills edit form without schedule and falls back to disabled defaults', () => {
    const api = useTaskFormState()
    const task = {
      id: 10,
      name: 'task-10',
      mode: 'copy',
      sourceRemote: 'src',
      sourcePath: '',
      targetRemote: 'dst',
      targetPath: '',
      options: {},
    } as any

    const parts = api.fillTaskFormForEdit(task)

    expect(parts).toBeNull()
    expect(api.createForm.value.enableSchedule).toBe(false)
    expect(api.createForm.value.scheduleMonth).toBe('*')
    expect(api.createForm.value.scheduleWeek).toBe('')
    expect(api.createForm.value.scheduleDay).toBe('')
    expect(api.createForm.value.scheduleHour).toBe('00')
    expect(api.createForm.value.scheduleMinute).toBe('00')
    expect(api.createForm.value.options.enableStreaming).toBe(true)
  })
})
