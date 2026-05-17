import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('BrowserView storage node hover styling', () => {
  it('uses task-card-like hover highlight and lift for storage tiles', () => {
    const source = readFileSync(resolve(__dirname, './BrowserView.vue'), 'utf8')

    expect(source).toContain('.tile:hover')
    expect(source).toContain('transform: translateY(-2px)')
    expect(source).toContain('box-shadow: 0 10px 24px')
    expect(source).toContain('border-color: rgba(99, 102, 241, 0.42)')
    expect(source).toContain('body.light .tile:hover')
    expect(source).toContain('border-color: rgba(25, 118, 210, 0.30)')
  })
})
