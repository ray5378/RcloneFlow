import { ref } from 'vue'
import type { Task } from '../types'
import { useTaskFormNormalize } from './useTaskFormNormalize'

export function useTaskFormState() {
  const { normalizeTaskOptionsForForm } = useTaskFormNormalize()
  const createForm = ref({
    name: '',
    mode: 'copy',
    sourceRemote: '',
    sourcePath: '',
    targetRemote: '',
    targetPath: '',
    enableSchedule: false,
    scheduleMonth: '*',
    scheduleWeek: '',
    scheduleDay: '',
    scheduleHour: '00',
    scheduleMinute: '00',
    options: { enableStreaming: true } as Record<string, any>,
  })

  const commandMode = ref(false)
  const commandText = ref('')
  const editingTask = ref<Task | null>(null)
  const showAdvancedOptions = ref(false)

  function resetTaskFormForCreate() {
    editingTask.value = null
    commandMode.value = false
    commandText.value = ''
    showAdvancedOptions.value = false
    createForm.value = {
      name: '',
      mode: 'copy',
      sourceRemote: '',
      sourcePath: '',
      targetRemote: '',
      targetPath: '',
      enableSchedule: false,
      scheduleMonth: '*',
      scheduleWeek: '',
      scheduleDay: '',
      scheduleHour: '00',
      scheduleMinute: '00',
      options: { enableStreaming: true },
    }
  }

  function fillTaskFormForEdit(task: Task, scheduleSpec?: string) {
    editingTask.value = task
    commandMode.value = false
    commandText.value = ''
    showAdvancedOptions.value = false

    if (scheduleSpec) {
      const parts = scheduleSpec.split('|')
      createForm.value = {
        name: task.name,
        mode: task.mode,
        sourceRemote: task.sourceRemote,
        sourcePath: task.sourcePath || '',
        targetRemote: task.targetRemote,
        targetPath: task.targetPath || '',
        enableSchedule: true,
        scheduleMinute: parts[0] || '00',
        scheduleHour: parts[1] || '*',
        scheduleDay: parts[2] || '*',
        scheduleMonth: parts[3] || '*',
        scheduleWeek: parts[4] || '*',
        options: normalizeTaskOptionsForForm(task.options as Record<string, any>),
      }
      return parts
    }

    createForm.value = {
      name: task.name,
      mode: task.mode,
      sourceRemote: task.sourceRemote,
      sourcePath: task.sourcePath || '',
      targetRemote: task.targetRemote,
      targetPath: task.targetPath || '',
      enableSchedule: false,
      scheduleMonth: '*',
      scheduleWeek: '',
      scheduleDay: '',
      scheduleHour: '00',
      scheduleMinute: '00',
      options: normalizeTaskOptionsForForm(task.options as Record<string, any>),
    }
    return null
  }

  return {
    createForm,
    commandMode,
    commandText,
    editingTask,
    showAdvancedOptions,
    resetTaskFormForCreate,
    fillTaskFormForEdit,
  }
}
