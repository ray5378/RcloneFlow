import { describe, it, expect } from 'vitest'
import { useTaskFormNormalize } from './useTaskFormNormalize'

describe('useTaskFormNormalize', () => {
  it('normalizes task options for execution payload', () => {
    const { normalizeTaskOptions } = useTaskFormNormalize()

    const result = normalizeTaskOptions({
      exclude: ' a\n\n b ',
      includeFrom: [' x ', '', ' y '],
      filesFromRaw: undefined,
      enableStreaming: undefined,
      retries: 3,
    })

    expect(result).toEqual({
      exclude: ['a', 'b'],
      includeFrom: ['x', 'y'],
      filesFromRaw: [],
      enableStreaming: true,
      retries: 3,
    })
  })

  it('normalizes task options for form display', () => {
    const { normalizeTaskOptionsForForm } = useTaskFormNormalize()

    const result = normalizeTaskOptionsForForm({
      exclude: ['a', ' b '],
      include: 'foo\nbar',
      enableStreaming: undefined,
      retries: 2,
    })

    expect(result).toMatchObject({
      exclude: 'a\nb',
      include: 'foo\nbar',
      enableStreaming: true,
      retries: 2,
    })
    expect(result.excludeFrom).toBe('')
    expect(result.filter).toBe('')
  })

  it('handles nullish or primitive multiline inputs safely', () => {
    const { normalizeTaskOptions, normalizeTaskOptionsForForm } = useTaskFormNormalize()

    expect(normalizeTaskOptions(null)).toEqual({ enableStreaming: true })
    expect(normalizeTaskOptionsForForm(undefined)).toMatchObject({ enableStreaming: true })

    const result = normalizeTaskOptions({ excludeFrom: 123 as any })
    expect(result.excludeFrom).toEqual([])

    const form = normalizeTaskOptionsForForm({ includeFrom: 123 as any })
    expect(form.includeFrom).toBe('')
  })
})
