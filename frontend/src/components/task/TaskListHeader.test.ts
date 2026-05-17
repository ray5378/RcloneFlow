import { describe, it, expect } from 'vitest'
import { readFileSync } from 'fs'
import { resolve } from 'path'

describe('TaskListHeader', () => {
  it('renders add task without plus sign and matches task sort button sizing', () => {
    const source = readFileSync(resolve(__dirname, './TaskListHeader.vue'), 'utf8')

    expect(source).toContain('{{ t(\'taskUI.addTask\') }}')
    expect(source).not.toContain('+ {{ t(\'taskUI.addTask\') }}')
    expect(source).toContain('class="ghost small action-btn" @click="emit(\'toggle-sort\')"')
    expect(source).toContain('class="primary small action-btn add-task-btn"')
    expect(source).toContain('.action-btn')
    expect(source).toContain('.header-actions > button.action-btn')
    expect(source).toContain('width: 88px !important')
    expect(source).toContain('height: 32px !important')
    expect(source).toContain('padding: 0 12px !important')
  })
})
