import { describe, it, expect } from 'vitest'
import { createApp, h } from 'vue'
import TransferCompletedList from './TransferCompletedList.vue'
import TransferPendingList from './TransferPendingList.vue'

function render(component: any, props: Record<string, any>) {
  const el = document.createElement('div')
  document.body.appendChild(el)
  const app = createApp({ render: () => h(component, props) })
  app.mount(el)
  return {
    el,
    unmount: () => {
      app.unmount()
      el.remove()
    },
  }
}

describe('transfer list placeholder rows', () => {
  it('keeps completed list at ten rows by rendering empty slots', () => {
    const { el, unmount } = render(TransferCompletedList, {
      items: [
        { name: 'done-a.mkv', path: 'done-a.mkv', status: 'copied', sizeBytes: 100 },
        { name: 'done-b.mkv', path: 'done-b.mkv', status: 'copied', sizeBytes: 200 },
      ],
      trackingMode: 'normal',
      total: 2,
      page: 1,
      totalPages: 1,
      jumpPage: 1,
    })

    expect(el.querySelectorAll('.row')).toHaveLength(10)
    expect(el.querySelectorAll('.placeholder-row')).toHaveLength(8)
    expect(el.textContent || '').toContain('done-a.mkv')
    expect(el.textContent || '').toContain('等待完成')

    unmount()
  })

  it('keeps pending list at ten rows by rendering empty slots', () => {
    const { el, unmount } = render(TransferPendingList, {
      items: [
        { name: 'pending-a.mkv', path: 'pending-a.mkv', status: 'pending', sizeBytes: 100 },
      ],
      trackingMode: 'normal',
      total: 1,
      page: 1,
      totalPages: 1,
      jumpPage: 1,
    })

    expect(el.querySelectorAll('.row')).toHaveLength(10)
    expect(el.querySelectorAll('.placeholder-row')).toHaveLength(9)
    expect(el.textContent || '').toContain('pending-a.mkv')
    expect(el.textContent || '').toContain('等待文件')

    unmount()
  })
})
