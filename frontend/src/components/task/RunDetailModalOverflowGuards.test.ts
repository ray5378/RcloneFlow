import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))

describe('run detail modal overflow guards', () => {
  it('includes dedicated wrapping and min-width guards in the modal styles', () => {
    const source = readFileSync(resolve(here, 'RunDetailModal.vue'), 'utf8')
    expect(source).toContain('.detail-item{min-width:0;}')
    expect(source).toContain('.summary-cell{min-width:0;')
    expect(source).toContain('.files-toolbar{display:flex;align-items:center;justify-content:space-between;gap:12px;margin-top:8px;max-width:1200px;min-width:0;')
    expect(source).toContain('.pager-inline{display:flex;align-items:center;gap:8px;flex-wrap:wrap;min-width:0;}')
  })
})
