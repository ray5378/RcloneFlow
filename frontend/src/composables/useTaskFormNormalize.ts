const multilineOptionKeys = [
  'exclude',
  'excludeFrom',
  'excludeIfPresent',
  'include',
  'includeFrom',
  'filter',
  'filterFrom',
  'filesFrom',
  'filesFromRaw',
] as const

function toMultilineText(value: unknown): string {
  if (Array.isArray(value)) {
    return value.map(v => String(v ?? '').trim()).filter(Boolean).join('\n')
  }
  if (typeof value === 'string') {
    return value
  }
  return ''
}

function toStringArray(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value.map(v => String(v ?? '').trim()).filter(Boolean)
  }
  if (typeof value === 'string') {
    return value
      .split(/\r?\n/)
      .map(v => v.trim())
      .filter(Boolean)
  }
  return []
}

export function useTaskFormNormalize() {
  function normalizeTaskOptions(raw: Record<string, any> | undefined | null) {
    const options = { ...(raw || {}) }
    if (typeof options.enableStreaming === 'undefined') {
      options.enableStreaming = true
    }
    for (const key of multilineOptionKeys) {
      if (key in options) {
        options[key] = toStringArray(options[key])
      }
    }
    return options
  }

  function normalizeTaskOptionsForForm(raw: Record<string, any> | undefined | null) {
    const options = { ...(raw || {}) }
    if (typeof options.enableStreaming === 'undefined') {
      options.enableStreaming = true
    }
    for (const key of multilineOptionKeys) {
      options[key] = toMultilineText(options[key])
    }
    return options
  }

  return {
    normalizeTaskOptions,
    normalizeTaskOptionsForForm,
  }
}
