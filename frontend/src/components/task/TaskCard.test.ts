import { readFileSync } from 'fs'
import { resolve } from 'path'
import { describe, it, expect } from 'vitest'
import { createApp, h } from 'vue'
import TaskCard from './TaskCard.vue'

describe('TaskCard', () => {
  it('keeps task list cards borderless and highlights the whole card', () => {
    const here = resolve(__dirname)
    const componentSource = readFileSync(resolve(here, './TaskCard.vue'), 'utf8')
    const baseStyles = readFileSync(resolve(here, './listItemBase.css'), 'utf8')

    const taskPathsRule = componentSource.match(/\.task-paths\s*\{[^}]*\}/)?.[0] || ''

    expect(taskPathsRule).not.toMatch(/background\s*:/)
    expect(taskPathsRule).not.toMatch(/border-radius\s*:/)
    expect(componentSource).toContain('rgba(99, 102, 241, 0.10)')
    expect(componentSource).toContain('rgba(25, 118, 210, 0.08)')
    expect(componentSource).not.toContain('border-left-color')
    expect(baseStyles).not.toContain('border-bottom:')
    expect(baseStyles).not.toContain('border-left:')
  })

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
