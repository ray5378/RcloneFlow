import { describe, it, expect } from 'vitest'
import { createApp, h } from 'vue'
import RunItem from './RunItem.vue'

describe('RunItem', () => {
  it('falls back to transfer total count when summary total is missing', () => {
    const el = document.createElement('div')
    document.body.appendChild(el)

    const app = createApp({
      render() {
        return h(RunItem, {
          run: {
            id: 21,
            taskId: 7,
            taskName: 'demo task',
            taskMode: 'copy',
            trigger: 'manual',
            status: 'finished',
            sourceRemote: 'src',
            sourcePath: '/a',
            targetRemote: 'dst',
            targetPath: '/b',
            startedAt: '2026-05-14T10:00:00Z',
          },
          progress: {
            logicalTotalCount: 2050,
            totalCount: 2050,
            plannedFiles: 2050,
            completedFiles: 2050,
          },
          summary: {
            counts: {
              copied: 2050,
              failed: 0,
            },
            totalBytes: 1024,
            transferredBytes: 1024,
          },
        })
      },
    })

    app.mount(el)

    expect(el.textContent || '').toContain('总计 2050')

    app.unmount()
    el.remove()
  })
})
