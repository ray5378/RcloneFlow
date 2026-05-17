import { readFileSync } from 'fs'
import { resolve } from 'path'
import { describe, it, expect } from 'vitest'
import { createApp, h } from 'vue'
import TaskCard from './TaskCard.vue'

describe('TaskCard', () => {
  it('keeps task cards bordered, path details unboxed, and hover elevated without recoloring', () => {
    const here = resolve(__dirname)
    const componentSource = readFileSync(resolve(here, './TaskCard.vue'), 'utf8')
    const baseStyles = readFileSync(resolve(here, './listItemBase.css'), 'utf8')
    const globalStyles = readFileSync(resolve(here, '../../styles/global.css'), 'utf8')

    const taskPathsRules = [
      ...componentSource.matchAll(/\.task-paths\s*\{[^}]*\}/g),
      ...globalStyles.matchAll(/\.task-paths\s*\{[^}]*\}/g),
    ].map(match => match[0])
    const hoverRules = [
      ...componentSource.matchAll(/\.task-card:hover\s*\{[^}]*\}/g),
      ...baseStyles.matchAll(/:where\(\.task-card, \.run-item\):hover\s*\{[^}]*\}/g),
      ...baseStyles.matchAll(/body\.light\s+:where\(\.task-card, \.run-item\):hover\s*\{[^}]*\}/g),
    ].map(match => match[0])

    expect(taskPathsRules.length).toBeGreaterThan(0)
    for (const rule of taskPathsRules) {
      expect(rule).not.toMatch(/background\s*:/)
      expect(rule).not.toMatch(/border-radius\s*:/)
    }
    for (const rule of hoverRules) {
      expect(rule).not.toMatch(/background\s*:/)
    }
    expect(baseStyles).toContain('margin-bottom: 14px')
    expect(baseStyles).toContain('border: 1px solid')
    expect(baseStyles).toContain('box-shadow')
    expect(baseStyles).toContain('transform: translateY(-2px)')
    expect(componentSource).not.toContain('border-left-color')
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
