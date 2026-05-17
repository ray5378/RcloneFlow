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
    const listSectionSource = readFileSync(resolve(here, './TaskListSection.vue'), 'utf8')
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
    expect(hoverRules.join('\n')).toContain('background: #252525')
    expect(baseStyles).toContain(':global(body.light) :where(.task-card, .run-item):hover')
    expect(hoverRules.join('\n')).toContain('background: #f8f8f8')
    expect(listSectionSource).toContain('class="list task-list-cards"')
    expect(listSectionSource).toContain('v-if="!sorting"')
    expect(listSectionSource).not.toContain('tasksTotal > tasksPageSize')
    expect(listSectionSource).toContain('gap: 8px')
    expect(listSectionSource).toContain('margin-top: 16px')
    expect(listSectionSource).toContain('padding: 0 16px')
    expect(listSectionSource).toContain('box-sizing: border-box')
    expect(listSectionSource).toContain('task-list-pagination-wrap')
    expect(listSectionSource).toContain('margin-bottom: 16px')
    expect(listSectionSource).toContain('margin-bottom: 0')
    expect(baseStyles).toContain('border: 1px solid')
    expect(baseStyles).toContain('box-shadow')
    expect(baseStyles).toContain('transform: translateY(-2px)')
    expect(componentSource).toContain('background: #f5f7fa')
    expect(componentSource).toContain('border-color: rgba(25, 118, 210, 0.22)')
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
