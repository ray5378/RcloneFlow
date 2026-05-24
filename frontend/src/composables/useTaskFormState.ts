import { ref } from 'vue'
import type { CreateForm, TaskFormOptions, TaskMode, BisyncOptions } from '../components/task/types'
import type { Task } from '../types'
import { useTaskFormNormalize } from './useTaskFormNormalize'

function normalizeTaskMode(mode: string): TaskMode {
  return mode === 'sync' || mode === 'move' || mode === 'copy' || mode === 'bisync' ? (mode as TaskMode) : 'copy'
}

export function useTaskFormState() {
  const { normalizeTaskOptionsForForm } = useTaskFormNormalize()
  const createForm = ref<CreateForm>({
    name: '',
    mode: 'copy',
    sourceRemote: '',
    sourcePath: '',
    targetRemote: '',
    targetPath: '',
    options: { enableStreaming: true } as TaskFormOptions,
    bisyncOptions: {} as BisyncOptions,
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
      options: { enableStreaming: true },
      bisyncOptions: {},
    }
  }

  function fillTaskFormForEdit(task: Task): void {
    editingTask.value = task
    commandMode.value = false
    commandText.value = ''
    showAdvancedOptions.value = false

    createForm.value = {
      name: task.name,
      mode: normalizeTaskMode(task.mode),
      sourceRemote: task.sourceRemote,
      sourcePath: task.sourcePath || '',
      targetRemote: task.targetRemote,
      targetPath: task.targetPath || '',
      options: normalizeTaskOptionsForForm(task.options as TaskFormOptions | undefined),
      bisyncOptions: (task.bisyncOptions || {}) as BisyncOptions,
    }
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
