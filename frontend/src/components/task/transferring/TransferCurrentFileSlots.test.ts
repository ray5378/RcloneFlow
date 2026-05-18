import { describe, it, expect } from 'vitest'
import { createApp, h } from 'vue'
import TransferCurrentFileCard from './TransferCurrentFileCard.vue'

function render(props: Record<string, any>) {
  const el = document.createElement('div')
  document.body.appendChild(el)
  const app = createApp({ render: () => h(TransferCurrentFileCard, props) })
  app.mount(el)
  return {
    el,
    unmount: () => {
      app.unmount()
      el.remove()
    },
  }
}

describe('TransferCurrentFileCard transfer slots', () => {
  it('renders current file cards plus placeholders up to transferSlots', () => {
    const { el, unmount } = render({
      currentFile: null,
      currentFiles: [
        { name: 'a.mkv', path: 'a.mkv', status: 'in_progress', bytes: 10, totalBytes: 100, speed: 1, percentage: 10 },
        { name: 'b.mkv', path: 'b.mkv', status: 'in_progress', bytes: 20, totalBytes: 100, speed: 2, percentage: 20 },
      ],
      trackingMode: 'normal',
      transferSlots: 4,
    })

    expect(el.querySelectorAll('.file-card')).toHaveLength(4)
    expect(el.querySelectorAll('.placeholder-card')).toHaveLength(2)
    expect(el.textContent || '').toContain('a.mkv')
    expect(el.textContent || '').toContain('等待传输')

    unmount()
  })

  it('renders one placeholder when no current file and transferSlots is missing', () => {
    const { el, unmount } = render({
      currentFile: null,
      currentFiles: [],
      trackingMode: 'normal',
    })

    expect(el.querySelectorAll('.file-card')).toHaveLength(1)
    expect(el.querySelectorAll('.placeholder-card')).toHaveLength(1)

    unmount()
  })
})
