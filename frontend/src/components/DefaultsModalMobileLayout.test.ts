import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))

describe('defaults modal mobile layout', () => {
  it('stacks labels and inputs vertically on small screens', () => {
    const source = readFileSync(resolve(here, 'DefaultsModal.vue'), 'utf8')
    expect(source).toContain('@media (max-width: 768px)')
    expect(source).toContain('.grid {')
    expect(source).toContain('grid-template-columns: 1fr;')
  })
})
