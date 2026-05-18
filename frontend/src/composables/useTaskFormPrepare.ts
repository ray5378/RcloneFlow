import type { Ref } from 'vue'
import type { CreateForm, ParsedRcloneCommand, TaskFormOptions } from '../components/task/types'
import { t } from '../i18n'

interface UseTaskFormPrepareOptions {
  createForm: Ref<CreateForm>
  commandMode: Ref<boolean>
  commandText: Ref<string>
  normalizeTaskOptions: (raw: TaskFormOptions | undefined | null) => TaskFormOptions
  parseRcloneCommand: (cmd: string) => ParsedRcloneCommand
  validateTaskForm: () => string
}

export function useTaskFormPrepare(options: UseTaskFormPrepareOptions) {
  function prepareTaskFormSubmit() {
    if (!options.commandMode.value) return ''
    try {
      const parsed = options.parseRcloneCommand(options.commandText.value)
      options.createForm.value.mode = parsed.mode
      options.createForm.value.sourceRemote = parsed.src.remote
      options.createForm.value.sourcePath = parsed.src.path
      options.createForm.value.targetRemote = parsed.dst.remote
      options.createForm.value.targetPath = parsed.dst.path
      options.createForm.value.options = {
        ...options.normalizeTaskOptions(options.createForm.value.options),
        ...parsed.options,
      }
      return ''
    } catch (e) {
      return t('runtime.commandParseFailed').replace('{message}', (e as Error).message)
    }
  }

  function validateTaskFormBeforeSubmit() {
    const prepareError = prepareTaskFormSubmit()
    if (prepareError) return prepareError
    return options.validateTaskForm()
  }

  return {
    prepareTaskFormSubmit,
    validateTaskFormBeforeSubmit,
  }
}
