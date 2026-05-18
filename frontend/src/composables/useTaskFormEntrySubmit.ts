interface UseTaskFormEntrySubmitOptions {
  runTaskFormFlow: () => Promise<string>
  showToast: (message: string, type?: 'info' | 'success' | 'error') => void
}

export function useTaskFormEntrySubmit(options: UseTaskFormEntrySubmitOptions) {
  async function createTask() {
    const error = await options.runTaskFormFlow()
    if (error) {
      options.showToast(error, 'error')
    }
  }

  return {
    createTask,
  }
}
