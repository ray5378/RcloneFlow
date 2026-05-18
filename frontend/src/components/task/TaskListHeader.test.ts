import { describe, it, expect } from 'vitest'
import { readFileSync } from 'fs'
import { resolve } from 'path'

describe('TaskListHeader', () => {
  it('renders add task without plus sign and matches task sort button sizing', () => {
    const source = readFileSync(resolve(__dirname, './TaskListHeader.vue'), 'utf8')

    expect(source).toContain('{{ t(\'taskUI.addTask\') }}')
    expect(source).not.toContain('+ {{ t(\'taskUI.addTask\') }}')
    expect(source).toContain('class="ghost small task-header-action-btn" @click="emit(\'toggle-sort\')"')
    expect(source).toContain('class="primary small task-header-action-btn add-task-btn"')
    expect(source).toContain('.task-header-action-btn')
    expect(source).toContain('.header-actions > button.primary.small.task-header-action-btn')
    expect(source).toContain('inline-size: 88px !important')
    expect(source).toContain('block-size: 32px !important')
    expect(source).toContain('flex: 0 0 88px !important')
  })
})
