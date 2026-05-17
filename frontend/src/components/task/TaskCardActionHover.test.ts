import { describe, it, expect } from 'vitest'
import { readFileSync } from 'fs'
import { resolve } from 'path'

describe('Task card action button hover colors', () => {
  it('uses semantic hover backgrounds for normal, info, and danger actions', () => {
    const componentSource = readFileSync(resolve(__dirname, './TaskCard.vue'), 'utf8')
    const actionStyles = readFileSync(resolve(__dirname, './listItemActions.css'), 'utf8')

    expect(componentSource).toContain('action-info')
    expect(componentSource).toContain('action-danger')
    expect(actionStyles).toContain('.list-item-actions :where(button.ghost.small):hover')
    expect(actionStyles).toContain('background: #374151')
    expect(actionStyles).toContain('.list-item-actions :where(button.ghost.small.action-info):hover')
    expect(actionStyles).toContain('background: rgba(59, 130, 246, 0.18)')
    expect(actionStyles).toContain('.list-item-actions :where(button.ghost.small.action-danger):hover')
    expect(actionStyles).toContain('background: rgba(239, 68, 68, 0.18)')
    expect(actionStyles).toContain(':global(body.light) .list-item-actions :where(button.ghost.small):hover')
    expect(actionStyles).toContain('background: #e5e7eb')
    expect(actionStyles).toContain('background: #dbeafe')
    expect(actionStyles).toContain('background: #fee2e2')
  })
})
