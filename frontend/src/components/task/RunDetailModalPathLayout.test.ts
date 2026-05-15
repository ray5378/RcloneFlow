import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))

describe('run detail path layout', () => {
  it('marks source and target paths with a dedicated wrapping class', () => {
    const source = readFileSync(resolve(here, 'RunDetailModal.vue'), 'utf8')
    expect(source).toContain('<span class="detail-path">{{ props.runDetail.sourceRemote }}:{{ props.runDetail.sourcePath || \'/\' }}</span>')
    expect(source).toContain('<span class="detail-path">{{ props.runDetail.targetRemote }}:{{ props.runDetail.targetPath || \'/\' }}</span>')
  })
})
