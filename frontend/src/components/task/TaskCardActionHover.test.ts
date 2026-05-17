import { describe, it, expect } from 'vitest'
import { readFileSync } from 'fs'
import { resolve } from 'path'

describe('Task card action button hover colors', () => {
  it('uses one unified hover background for all task action buttons', () => {
    const componentSource = readFileSync(resolve(__dirname, './TaskCard.vue'), 'utf8')
    const actionStyles = readFileSync(resolve(__dirname, './listItemActions.css'), 'utf8')

    expect(componentSource).not.toContain('action-info')
    expect(componentSource).not.toContain('action-danger')
    expect(actionStyles).toContain('.list-item-actions button.ghost.small:hover')
    expect(actionStyles).toContain('background: #374151 !important')
    expect(actionStyles).toContain(':global(body.light) .list-item-actions button.ghost.small:hover')
    expect(actionStyles).toContain('background: #e5e7eb !important')
    expect(actionStyles).not.toContain('rgba(59, 130, 246, 0.18)')
    expect(actionStyles).not.toContain('rgba(239, 68, 68, 0.18)')
    expect(actionStyles).not.toContain('#dbeafe')
    expect(actionStyles).not.toContain('#fee2e2')
  })
})
