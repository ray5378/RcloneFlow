import { describe, it, expect } from 'vitest'
import { createApp, h } from 'vue'
import FileItem from './FileItem.vue'

describe('FileItem', () => {
  it('shows only the basename for file paths', () => {
    const el = document.createElement('div')
    document.body.appendChild(el)

    const app = createApp({
      render() {
        return h(FileItem, {
          item: {
            path: 'folder/sub/demo-video.mkv',
            status: 'Copied',
            at: '2026-05-15T01:00:00Z',
            sizeBytes: 2048,
          },
        })
      },
    })

    app.mount(el)

    const text = el.textContent || ''
    expect(text).toContain('demo-video.mkv')
    expect(text).not.toContain('folder/sub/demo-video.mkv')

    app.unmount()
    el.remove()
  })
})
