import type { Ref } from 'vue'

interface ParsedRcloneCommand {
  mode: string
  src: { remote: string; path: string }
  dst: { remote: string; path: string }
  options: Record<string, any>
}

interface UseTaskFormPrepareOptions {
  createForm: Ref<any>
  commandMode: Ref<boolean>
  commandText: Ref<string>
  normalizeTaskOptions: (raw: Record<string, any> | undefined | null) => Record<string, any>
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
      return '命令解析失败：' + (e as Error).message
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
