import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))

describe('mobile transfer list layout', () => {
  it('removes status tag column from completed list template', () => {
    const source = readFileSync(resolve(here, 'TransferCompletedList.vue'), 'utf8')
    expect(source).not.toContain('<span class="tag">')
  })

  it('removes status tag column from pending list template', () => {
    const source = readFileSync(resolve(here, 'TransferPendingList.vue'), 'utf8')
    expect(source).not.toContain('<span class="tag">')
  })
})
