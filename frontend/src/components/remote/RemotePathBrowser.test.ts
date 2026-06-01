import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { createApp, h, nextTick } from 'vue'
import RemotePathBrowser from './RemotePathBrowser.vue'

vi.mock('../../api', () => ({
  getRemotes: vi.fn(),
  listPath: vi.fn(),
}))

vi.mock('../../i18n', () => ({
  t: (key: string) => key,
  locale: { value: 'zh' },
}))

import * as api from '../../api'

function mount(component: any, props: Record<string, unknown> = {}) {
  const el = document.createElement('div')
  document.body.appendChild(el)
  const app = createApp({
    render() {
      return h(component, props)
    },
  })
  app.config.globalProperties.$t = (key: string) => key
  app.mount(el)
  return { el, app, unmount: () => { app.unmount(); el.remove() } }
}

function wait() {
  return new Promise(resolve => setTimeout(resolve, 0))
}

async function flush() {
  await wait()
  await nextTick()
  await wait()
  await nextTick()
}

describe('RemotePathBrowser', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('remotes loading', () => {
    it('loads and displays remotes in dropdown', async () => {
      ;(api.getRemotes as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        remotes: ['gdrive', 'onedrive', 's3'],
      })

      const { el, unmount } = mount(RemotePathBrowser, { modelValue: '' })
      await wait()
      await nextTick()

      const select = el.querySelector('select')
      expect(select).not.toBeNull()
      const options = select!.querySelectorAll('option')
      const texts = Array.from(options).map(o => o.textContent?.trim())
      expect(texts).toContain('gdrive')
      expect(texts).toContain('onedrive')
      expect(texts).toContain('s3')

      unmount()
    })

    it('handles empty remotes list', async () => {
      ;(api.getRemotes as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        remotes: [],
      })

      const { el, unmount } = mount(RemotePathBrowser, { modelValue: '' })
      await wait()
      await nextTick()

      const select = el.querySelector('select')
      expect(select).not.toBeNull()
      const options = select!.querySelectorAll('option')
      expect(options.length).toBe(1)

      unmount()
    })

    it('handles errors when loading remotes', async () => {
      ;(api.getRemotes as ReturnType<typeof vi.fn>).mockRejectedValueOnce(
        new Error('Network error')
      )

      const { el, unmount } = mount(RemotePathBrowser, { modelValue: '' })
      await wait()
      await nextTick()

      expect(el.textContent).not.toContain('Network error')

      unmount()
    })
  })

  describe('remote selection and directory browsing', () => {
    it('lists root directories when a remote is selected', async () => {
      ;(api.getRemotes as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        remotes: ['gdrive'],
      })
      ;(api.listPath as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        fs: 'gdrive:',
        items: [
          { Path: 'photos', Name: 'photos', IsDir: true, Size: '0', ModTime: '' },
          { Path: 'backup', Name: 'backup', IsDir: true, Size: '0', ModTime: '' },
          { Path: 'readme.txt', Name: 'readme.txt', IsDir: false, Size: '100', ModTime: '' },
        ],
      })

      const { el, unmount } = mount(RemotePathBrowser, { modelValue: '' })
      await wait()
      await nextTick()

      const select = el.querySelector('select') as HTMLSelectElement
      select.value = 'gdrive'
      select.dispatchEvent(new Event('change'))
      await wait()
      await nextTick()

      const pathItems = el.querySelectorAll('.path-item')
      expect(pathItems.length).toBe(2)
      const names = Array.from(pathItems).map(item => item.querySelector('.item-name')?.textContent?.trim())
      expect(names).toContain('photos')
      expect(names).toContain('backup')
      expect(names).not.toContain('readme.txt')

      unmount()
    })

    it('navigates into subdirectories on click', async () => {
      ;(api.getRemotes as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        remotes: ['gdrive'],
      })
      ;(api.listPath as ReturnType<typeof vi.fn>)
        .mockResolvedValueOnce({
          fs: 'gdrive:',
          items: [
            { Path: 'photos', Name: 'photos', IsDir: true, Size: '0', ModTime: '' },
          ],
        })
        .mockResolvedValueOnce({
          fs: 'gdrive:photos',
          items: [
            { Path: 'photos/2024', Name: '2024', IsDir: true, Size: '0', ModTime: '' },
            { Path: 'photos/2025', Name: '2025', IsDir: true, Size: '0', ModTime: '' },
          ],
        })

      const { el, unmount } = mount(RemotePathBrowser, { modelValue: '' })
      await wait()
      await nextTick()

      const select = el.querySelector('select') as HTMLSelectElement
      select.value = 'gdrive'
      select.dispatchEvent(new Event('change'))
      await wait()
      await nextTick()

      const firstDir = el.querySelector('.path-item') as HTMLElement
      firstDir.click()
      await wait()
      await nextTick()

      expect(api.listPath).toHaveBeenCalledWith('gdrive', 'photos')

      const items = el.querySelectorAll('.path-item')
      const names = Array.from(items).map(item => item.querySelector('.item-name')?.textContent?.trim())
      expect(names).toContain('2024')
      expect(names).toContain('2025')

      unmount()
    })
  })

  describe('breadcrumb navigation', () => {
    it('displays breadcrumbs and allows navigating back', async () => {
      ;(api.getRemotes as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        remotes: ['gdrive'],
      })
      ;(api.listPath as ReturnType<typeof vi.fn>)
        .mockResolvedValueOnce({
          fs: 'gdrive:',
          items: [
            { Path: 'photos', Name: 'photos', IsDir: true, Size: '0', ModTime: '' },
          ],
        })
        .mockResolvedValueOnce({
          fs: 'gdrive:photos',
          items: [
            { Path: 'photos/2024', Name: '2024', IsDir: true, Size: '0', ModTime: '' },
          ],
        })
        .mockResolvedValueOnce({
          fs: 'gdrive:',
          items: [
            { Path: 'photos', Name: 'photos', IsDir: true, Size: '0', ModTime: '' },
          ],
        })

      const { el, unmount } = mount(RemotePathBrowser, { modelValue: '' })
      await wait()
      await nextTick()

      const select = el.querySelector('select') as HTMLSelectElement
      select.value = 'gdrive'
      select.dispatchEvent(new Event('change'))
      await wait()
      await nextTick()

      const firstDir = el.querySelector('.path-item') as HTMLElement
      firstDir.click()
      await wait()
      await nextTick()

      const crumbs = el.querySelectorAll('.crumb')
      const crumbTexts = Array.from(crumbs).map(c => c.textContent?.trim())
      expect(crumbTexts).toContain('gdrive')
      expect(crumbTexts).toContain('photos')

      const remoteCrumb = crumbs[0] as HTMLElement
      remoteCrumb.click()
      await wait()
      await nextTick()

      expect(api.listPath).toHaveBeenCalledWith('gdrive', '')

      unmount()
    })
  })

  describe('path emission', () => {
    it('emits correct path format remoteName:path when selecting directory', async () => {
      ;(api.getRemotes as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        remotes: ['gdrive'],
      })
      ;(api.listPath as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        fs: 'gdrive:',
        items: [
          { Path: 'backup', Name: 'backup', IsDir: true, Size: '0', ModTime: '' },
        ],
      })

      const emitted: string[] = []
      const { el, unmount, app } = mount(RemotePathBrowser, {
        modelValue: '',
        'onUpdate:modelValue': (val: string) => emitted.push(val),
      })
      await wait()
      await nextTick()

      const select = el.querySelector('select') as HTMLSelectElement
      select.value = 'gdrive'
      select.dispatchEvent(new Event('change'))
      await wait()
      await nextTick()

      const fullPaths = emitted.filter(v => v.includes(':'))
      expect(fullPaths.length).toBeGreaterThan(0)
      expect(fullPaths[fullPaths.length - 1]).toBe('gdrive:')

      unmount()
    })

    it('emits path with nested directory after navigation', async () => {
      ;(api.getRemotes as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        remotes: ['gdrive'],
      })
      ;(api.listPath as ReturnType<typeof vi.fn>)
        .mockResolvedValueOnce({
          fs: 'gdrive:',
          items: [
            { Path: 'data', Name: 'data', IsDir: true, Size: '0', ModTime: '' },
          ],
        })
        .mockResolvedValueOnce({
          fs: 'gdrive:data',
          items: [
            { Path: 'data/encrypted', Name: 'encrypted', IsDir: true, Size: '0', ModTime: '' },
          ],
        })

      const emitted: string[] = []
      const { el, unmount } = mount(RemotePathBrowser, {
        modelValue: '',
        'onUpdate:modelValue': (val: string) => emitted.push(val),
      })
      await wait()
      await nextTick()

      const select = el.querySelector('select') as HTMLSelectElement
      select.value = 'gdrive'
      select.dispatchEvent(new Event('change'))
      await wait()
      await nextTick()

      const dataDir = el.querySelector('.path-item') as HTMLElement
      dataDir.click()
      await wait()
      await nextTick()

      const pathValues = emitted.filter(v => v.includes(':'))
      expect(pathValues.length).toBeGreaterThanOrEqual(1)
      expect(pathValues[pathValues.length - 1]).toBe('gdrive:data')

      unmount()
    })
  })

  describe('select current directory button', () => {
    it('shows select button with current path and emits on click', async () => {
      ;(api.getRemotes as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        remotes: ['gdrive'],
      })
      ;(api.listPath as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        fs: 'gdrive:',
        items: [
          { Path: 'target', Name: 'target', IsDir: true, Size: '0', ModTime: '' },
        ],
      })

      const emitted: string[] = []
      const { el, unmount } = mount(RemotePathBrowser, {
        modelValue: '',
        'onUpdate:modelValue': (val: string) => emitted.push(val),
      })
      await wait()
      await nextTick()

      const select = el.querySelector('select') as HTMLSelectElement
      select.value = 'gdrive'
      select.dispatchEvent(new Event('change'))
      await wait()
      await nextTick()

      const selectBtn = el.querySelector('.select-btn') as HTMLElement
      expect(selectBtn).not.toBeNull()

      const lastBefore = emitted.filter(v => v.includes(':')).pop() || ''
      selectBtn.click()
      await wait()
      await nextTick()

      const lastAfter = emitted.filter(v => v.includes(':')).pop() || ''
      expect(lastAfter).toBe(lastBefore)

      unmount()
    })
  })

  describe('initial modelValue restoration', () => {
    it('restores state from modelValue', async () => {
      ;(api.getRemotes as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        remotes: ['gdrive', 'onedrive'],
      })
      ;(api.listPath as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        fs: 'gdrive:backup',
        items: [
          { Path: 'backup/2024', Name: '2024', IsDir: true, Size: '0', ModTime: '' },
        ],
      })

      const { el, unmount } = mount(RemotePathBrowser, {
        modelValue: 'gdrive:backup',
      })
      await flush()

      const select = el.querySelector('select') as HTMLSelectElement
      expect(select.value).toBe('gdrive')

      expect(api.listPath).toHaveBeenCalledWith('gdrive', 'backup')
      expect(api.listPath).toHaveBeenCalledTimes(1)

      const items = el.querySelectorAll('.path-item')
      const names = Array.from(items).map(item => item.querySelector('.item-name')?.textContent?.trim())
      expect(names).toContain('2024')

      unmount()
    })
  })

  describe('edge cases', () => {
    it('switching remote clears previous path', async () => {
      ;(api.getRemotes as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        remotes: ['gdrive', 'onedrive'],
      })
      ;(api.listPath as ReturnType<typeof vi.fn>)
        .mockResolvedValueOnce({
          fs: 'gdrive:',
          items: [
            { Path: 'photos', Name: 'photos', IsDir: true, Size: '0', ModTime: '' },
          ],
        })
        .mockResolvedValueOnce({
          fs: 'onedrive:',
          items: [
            { Path: 'docs', Name: 'docs', IsDir: true, Size: '0', ModTime: '' },
          ],
        })

      const { el, unmount } = mount(RemotePathBrowser, { modelValue: '' })
      await wait()
      await nextTick()

      const select = el.querySelector('select') as HTMLSelectElement
      select.value = 'gdrive'
      select.dispatchEvent(new Event('change'))
      await wait()
      await nextTick()

      expect(api.listPath).toHaveBeenCalledWith('gdrive', '')

      select.value = 'onedrive'
      select.dispatchEvent(new Event('change'))
      await wait()
      await nextTick()

      expect(api.listPath).toHaveBeenCalledWith('onedrive', '')

      unmount()
    })

    it('handles listPath error gracefully', async () => {
      ;(api.getRemotes as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        remotes: ['gdrive'],
      })
      ;(api.listPath as ReturnType<typeof vi.fn>).mockRejectedValueOnce(
        new Error('Permission denied')
      )

      const { el, unmount } = mount(RemotePathBrowser, { modelValue: '' })
      await wait()
      await nextTick()

      const select = el.querySelector('select') as HTMLSelectElement
      select.value = 'gdrive'
      select.dispatchEvent(new Event('change'))
      await wait()
      await nextTick()

      const errorEl = el.querySelector('.path-error')
      expect(errorEl).not.toBeNull()
      expect(errorEl!.textContent).toContain('Permission denied')

      unmount()
    })

    it('handles modelValue with no colon gracefully', async () => {
      ;(api.getRemotes as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        remotes: ['gdrive'],
      })

      const { el, unmount } = mount(RemotePathBrowser, {
        modelValue: 'just_a_string',
      })
      await wait()
      await nextTick()

      expect(api.listPath).not.toHaveBeenCalled()

      unmount()
    })
  })
})