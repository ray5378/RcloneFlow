interface UseTaskFormFlowOptions {
  handleTaskFormDoneClick: () => boolean
  validateTaskFormBeforeSubmit: () => string
  executeTaskFormSubmit: () => Promise<string>
}

export function useTaskFormFlow(options: UseTaskFormFlowOptions) {
  async function runTaskFormFlow() {
    if (options.handleTaskFormDoneClick()) {
      return ''
    }

    const validationError = options.validateTaskFormBeforeSubmit()
    if (validationError) {
      return validationError
    }

    return await options.executeTaskFormSubmit()
  }

  return {
    runTaskFormFlow,
  }
}
