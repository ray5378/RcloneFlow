import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('WebhookConfigModal source', () => {
  it('includes hasTransfer prop, event, and checkbox label wiring', () => {
    const source = readFileSync(resolve(__dirname, './WebhookConfigModal.vue'), 'utf8')

    expect(source).toContain('statusHasTransfer')
    expect(source).toContain("update:statusHasTransfer")
    expect(source).toContain("taskUI.hasTransfer")
  })
})
