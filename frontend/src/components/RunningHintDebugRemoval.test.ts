import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))
const root = resolve(here, '..')

describe('running hint debug removal', () => {
  it('removes running debug setting from defaults modal', () => {
    const source = readFileSync(resolve(root, 'components/DefaultsModal.vue'), 'utf8')
    expect(source).not.toContain('RUNNING_HINT_DEBUG_ENABLED')
    expect(source).not.toContain("t('defaults.runningDebug')")
    expect(source).not.toContain("t('defaults.runningDebugLabel')")
  })

  it('removes debug controls from running hint modal', () => {
    const source = readFileSync(resolve(root, 'components/task/RunningHintModal.vue'), 'utf8')
    expect(source).not.toContain('toggle-debug')
    expect(source).not.toContain('debugOpen')
    expect(source).not.toContain('debugEnabled')
    expect(source).not.toContain("t('modal.expandDebug')")
    expect(source).not.toContain("t('modal.collapseDebug')")
  })
})
