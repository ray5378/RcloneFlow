export function useTaskFormNormalize() {
  function normalizeTaskOptions(raw: Record<string, any> | undefined | null) {
    const options = { ...(raw || {}) }
    if (typeof options.enableStreaming === 'undefined') {
      options.enableStreaming = true
    }
    return options
  }

  return {
    normalizeTaskOptions,
  }
}
