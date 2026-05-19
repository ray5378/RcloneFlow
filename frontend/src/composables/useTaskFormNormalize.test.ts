import { describe, it, expect } from 'vitest'
import { useTaskFormNormalize } from './useTaskFormNormalize'

describe('useTaskFormNormalize', () => {
  const { normalizeTaskOptions, normalizeTaskOptionsForForm } = useTaskFormNormalize()

  it('should handle null/undefined input', () => {
    const normalized = normalizeTaskOptions(null)
    expect(normalized.enableStreaming).toBe(true)
    const normalized2 = normalizeTaskOptionsForForm(null)
    expect(normalized2.enableStreaming).toBe(true)
  })

  it('should default enableStreaming to true', () => {
    const result = normalizeTaskOptions({})
    expect(result.enableStreaming).toBe(true)
  })

  it('should preserve existing enableStreaming', () => {
    const result = normalizeTaskOptions({ enableStreaming: false })
    expect(result.enableStreaming).toBe(false)
  })

  it('should convert multiline options from string to array', () => {
    const result = normalizeTaskOptions({
      exclude: '*.tmp\n*.log',
      include: '*.jpg',
    })
    expect(result.exclude).toEqual(['*.tmp', '*.log'])
    expect(result.include).toEqual(['*.jpg'])
  })

  it('should preserve array options in normalizeTaskOptions', () => {
    const result = normalizeTaskOptions({
      exclude: ['*.tmp', '*.log'],
    })
    expect(result.exclude).toEqual(['*.tmp', '*.log'])
  })

  it('should convert multiline options to string in normalizeTaskOptionsForForm', () => {
    const result = normalizeTaskOptionsForForm({
      exclude: ['*.tmp', '*.log'],
      include: '*.jpg',
    })
    expect(result.exclude).toBe('*.tmp\n*.log')
    expect(result.include).toBe('*.jpg')
  })

  it('should handle empty multiline values', () => {
    const result = normalizeTaskOptions({ exclude: '' })
    expect(result.exclude).toEqual([])

    const result2 = normalizeTaskOptions({ exclude: ['  ', '  '] })
    expect(result2.exclude).toEqual([])
  })

  it('should handle all multiline option keys', () => {
    const keys = ['exclude', 'excludeFrom', 'excludeIfPresent', 'include', 'includeFrom', 'filter', 'filterFrom', 'filesFrom', 'filesFromRaw']
    for (const key of keys) {
      const result = normalizeTaskOptions({ [key]: 'a\nb' })
      expect(result[key]).toEqual(['a', 'b'])
    }
  })
})
