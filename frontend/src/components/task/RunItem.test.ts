import { readFileSync } from 'fs'
import { resolve } from 'path'
import { describe, it, expect } from 'vitest'
import { createApp, h } from 'vue'
import RunItem from './RunItem.vue'

describe('RunItem', () => {
  it('uses the same inset and spacing layout as task cards in history lists', () => {
    const here = resolve(__dirname)
    const historySource = readFileSync(resolve(here, './TaskHistoryPanel.vue'), 'utf8')
    const baseStyles = readFileSync(resolve(here, './listItemBase.css'), 'utf8')
    const runSource = readFileSync(resolve(here, './RunItem.vue'), 'utf8')

    const hoverRules = [
      ...baseStyles.matchAll(/:where\(\.task-card, \.run-item\):hover\s*\{[^}]*\}/g),
      ...baseStyles.matchAll(/body\.light\s+:where\(\.task-card, \.run-item\):hover\s*\{[^}]*\}/g),
    ].map(match => match[0])

    expect(historySource).toContain('class="list history-list-cards"')
    expect(historySource).toContain('gap: 8px')
    expect(historySource).toContain('margin-top: 16px')
    expect(historySource).toContain('padding: 0 16px')
    expect(historySource).toContain('box-sizing: border-box')
    expect(historySource).toContain('margin-bottom: 0')
    expect(baseStyles).toContain('border: 1px solid')
    expect(baseStyles).toContain('border-radius: 12px')
    expect(baseStyles).toContain('transform: translateY(-2px)')
    for (const rule of hoverRules) {
      expect(rule).not.toMatch(/background\s*:/)
    }
    expect(runSource).not.toContain('border-left-color')
  })

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
