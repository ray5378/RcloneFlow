import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))

describe('run log modal overflow guards', () => {
  it('adds wrap and min-width protections for long log lines', () => {
    const source = readFileSync(resolve(here, 'RunLogModal.vue'), 'utf8')
    expect(source).toContain('.log-modal .modal-body {')
    expect(source).toContain('min-width: 0;')
    expect(source).toContain('.log-box {')
    expect(source).toContain('overflow: hidden;')
    expect(source).toContain('.log-pre {')
    expect(source).toContain('overflow-wrap: anywhere;')
    expect(source).toContain('word-break: break-word;')
  })
})
