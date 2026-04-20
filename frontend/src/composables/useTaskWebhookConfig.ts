import { ref } from 'vue'
import * as api from '../api'
import { getToken } from '../api/auth'
import { formatBytes, formatBytesPerSec } from '../utils/format'

export function useTaskWebhookConfig(options: {
  loadData: () => Promise<void>
  showToast: (message: string, type?: 'info' | 'success' | 'error') => void
}) {
  const showWebhookModal = ref(false)
  const webhookForm = ref<any>({
    taskId: null,
    postUrl: '',
    triggerId: '',
    notify: { manual: false, schedule: false, webhook: false },
    status: { success: true, failed: true },
    wecomUrl: '',
  })

  function setWebhook(task: any) {
    webhookForm.value.taskId = task.id
    try {
      const raw = task.options || task.Options || {}
      const opts = typeof raw === 'string' ? JSON.parse(raw || '{}') : raw
      webhookForm.value.postUrl = opts?.webhookPostUrl || ''
      webhookForm.value.wecomUrl = opts?.wecomPostUrl || ''
      webhookForm.value.triggerId = opts?.webhookId || ''
      const n = opts?.webhookNotifyOn || {}
      webhookForm.value.notify = {
        manual: !!n.manual,
        schedule: !!n.schedule,
        webhook: !!n.webhook,
      }
      const s = opts?.webhookNotifyStatus || {}
      webhookForm.value.status = { success: s.success !== false, failed: s.failed !== false }
    } catch {
      webhookForm.value.postUrl = ''
      webhookForm.value.wecomUrl = ''
      webhookForm.value.triggerId = ''
      webhookForm.value.notify = { manual: false, schedule: false, webhook: false }
      webhookForm.value.status = { success: true, failed: true }
    }
    showWebhookModal.value = true
  }

  async function saveWebhook() {
    if (!webhookForm.value.taskId) {
      showWebhookModal.value = false
      return
    }
    try {
      const id = webhookForm.value.taskId
      const payload = {
        webhookId: webhookForm.value.triggerId,
        webhookPostUrl: webhookForm.value.postUrl,
        wecomPostUrl: webhookForm.value.wecomUrl || '',
        webhookNotifyOn: webhookForm.value.notify,
        webhookNotifyStatus: webhookForm.value.status,
      }
      const fn: any = (api as any).updateTaskOptions
      if (typeof fn === 'function') {
        await fn(id, payload)
      } else {
        await fetch('/api/tasks', {
          method: 'PATCH',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${getToken() || ''}`,
          },
          body: JSON.stringify({ id, options: payload }),
        })
      }
      showWebhookModal.value = false
      await options.loadData()
    } catch (e: any) {
      options.showToast(e?.message || String(e), 'error')
    }
  }

  function buildWecomMarkdown(p: any) {
    const taskName = p?.task?.name || '测试任务'
    const statusZh = p?.statusZh || '演示'
    const triggerZh = p?.triggerZh || '测试'
    const mode = (p?.task?.mode || '').toLowerCase()
    const okLabel = mode === 'move' ? '移动' : '成功'
    const s = p?.summary || {}
    const total = s.totalCount ?? 0
    const ok = s.completedCount ?? 0
    const fail = s.failedCount ?? 0
    const skipped = s.skippedCount ?? 0
    const bytesFmt = formatBytes(Number(s.totalBytes || 0))
    const txFmt = formatBytes(Number(s.transferredBytes || 0))
    const spFmt = formatBytesPerSec(Number(s.avgSpeedBps || 0))
    const dur = p?.run?.durationText || '-'
    let md = `**任务** <font color="info">${taskName}</font> 已${statusZh}（${triggerZh}）\n`
    md += `> 总计: ${total}  ${okLabel}: ${ok}  失败: ${fail}  其他: ${skipped}\n`
    md += `> 体积: ${bytesFmt} / 已传: ${txFmt}\n`
    md += `> 均速: ${spFmt}  耗时: ${dur}\n`
    if (Array.isArray(p?.files)) {
      for (const f of p.files) md += `> ${f}\n`
    }
    return md
  }

  async function testWebhook() {
    try {
      const url1 = (webhookForm.value.postUrl || '').trim()
      const url2 = String(webhookForm.value.wecomUrl || '').trim()
      if (!url1 && !url2) {
        options.showToast('请先填写对外 POST 地址或企业微信地址', 'error')
        return
      }
      const payload: any = {
        title: 'RcloneFlow 测试通知',
        triggerZh: '测试',
        statusZh: '演示',
        summaryZh: '这是一条测试消息，用于验证通知接收是否正常（自动识别企业微信/通用 Webhook）。',
        task: { id: 0, name: '测试任务', mode: 'test' },
        run: { id: 0, trigger: 'manual', status: 'success', startedAt: new Date().toISOString(), finishedAt: new Date().toISOString(), durationSeconds: 1, durationText: '1秒' },
        summary: { totalCount: 5, completedCount: 5, failedCount: 0, skippedCount: 0, totalBytes: 10485760, transferredBytes: 10485760, avgSpeedBps: 10485760 },
        files: ['test-a.mp4', 'test-b.mp4', 'test-c.mp4', 'test-d.mp4', 'test-e.mp4'],
        omittedCount: 0,
      }
      const send = async (url: string) => {
        const isWecom = url.includes('qyapi.weixin.qq.com')
        const body = isWecom
          ? JSON.stringify({ msgtype: 'markdown', markdown: { content: buildWecomMarkdown(payload) } })
          : JSON.stringify(payload)
        const resp = await fetch(url, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body })
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
      }
      const fails: string[] = []
      if (url1) { try { await send(url1) } catch (e: any) { fails.push(`通用: ${e?.message || e}`) } }
      if (url2) { try { await send(url2) } catch (e: any) { fails.push(`企业微信: ${e?.message || e}`) } }
      if (fails.length) {
        options.showToast(`测试部分失败：${fails.join('；')}`, 'error')
      } else {
        options.showToast('测试通知已发送（请在接收端查看）', 'success')
      }
    } catch (e: any) {
      options.showToast(`测试发送失败：${e?.message || e}`, 'error')
    }
  }

  return {
    showWebhookModal,
    webhookForm,
    setWebhook,
    saveWebhook,
    testWebhook,
  }
}
