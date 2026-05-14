import { describe, it, expect } from 'vitest'
import { createApp, h } from 'vue'
import TaskCard from './TaskCard.vue'

describe('TaskCard', () => {
  it('shows total count when the running total is exactly 1', () => {
    const el = document.createElement('div')
    document.body.appendChild(el)

    const app = createApp({
      render() {
        return h(TaskCard, {
          task: {
            id: 7,
            name: 'demo task',
            mode: 'copy',
            sourceRemote: 'src',
            sourcePath: '/a',
            targetRemote: 'dst',
            targetPath: '/b',
          },
          progress: {
            percentage: 100,
            bytes: 1024,
            totalBytes: 1024,
            speed: 0,
            eta: 0,
            completedFiles: 1,
            plannedFiles: 1,
            logicalTotalCount: 1,
            totalCount: 1,
          },
          runningTaskId: 7,
        })
      },
    })

    app.mount(el)

    expect(el.textContent || '').toContain('总数量 1')

    app.unmount()
    el.remove()
  })
})
