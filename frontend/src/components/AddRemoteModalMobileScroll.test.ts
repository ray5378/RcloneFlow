import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))

describe('add remote modal mobile scrolling', () => {
  it('adds a dedicated scrollable body for long mobile content', () => {
    const source = readFileSync(resolve(here, 'AddRemoteModal.vue'), 'utf8')
    expect(source).toContain('class="add-remote-body"')
    expect(source).toContain('.add-remote-body {')
    expect(source).toContain('overflow-y: auto;')
    expect(source).toContain('@media (max-width: 768px)')
  })
})
