import { describe, it, expect } from 'vitest'
import { createApp, h } from 'vue'
import TaskListHeader from './TaskListHeader.vue'

function render(props: Record<string, any> = {}) {
  const el = document.createElement('div')
  document.body.appendChild(el)
  const app = createApp({
    render: () => h(TaskListHeader, {
      search: '',
      sorting: false,
      savingSort: false,
      ...props,
    }),
  })
  app.mount(el)
  return {
    el,
    unmount: () => {
      app.unmount()
      el.remove()
    },
  }
}

describe('TaskListHeader rendered action buttons', () => {
  it('marks add and sort buttons with the same compact size class', () => {
    const { el, unmount } = render()

    const sortButton = Array.from(el.querySelectorAll('button')).find(button => button.textContent?.includes('任务排序'))
    const addButton = Array.from(el.querySelectorAll('button')).find(button => button.textContent?.includes('添加任务'))

    expect(sortButton?.className).toContain('task-header-action-btn')
    expect(addButton?.className).toContain('task-header-action-btn')
    expect(addButton?.className).toContain('add-task-btn')
    expect(addButton?.textContent || '').not.toContain('+')

    unmount()
  })
})
