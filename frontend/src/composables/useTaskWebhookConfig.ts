import { ref } from 'vue'
import * as api from '../api'
import { getToken } from '../api/auth'
import { formatBytes, formatBytesPerSec } from '../utils/format'
import { t } from '../i18n'

export function useTaskWebhookConfig(options: {
  loadData: () => Promise<void>
  showToast: (message: string, type?: 'info' | 'success' | 'error') => void
}) {
  const showWebhookModal = ref(false)
  const webhookForm = ref<any>({
    taskId: null,
    postUrl: '',
    triggerId: '',
    matchText: '',
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
      webhookForm.value.matchText = opts?.webhookMatchText || ''
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
      webhookForm.value.matchText = ''
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
        webhookMatchText: webhookForm.value.matchText,
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
    const taskName = p?.task?.name || t('runtime.webhookTestTask')
    const statusZh = p?.statusZh || t('runtime.webhookTestStatus')
    const triggerZh = p?.triggerZh || t('runtime.webhookTestTrigger')
    const mode = (p?.task?.mode || '').toLowerCase()
    const okLabel = mode === 'move' ? t('runtime.webhookMove') : t('runtime.webhookSuccess')
    const s = p?.summary || {}
    const total = s.totalCount ?? 0
    const ok = s.completedCount ?? 0
    const fail = s.failedCount ?? 0
    const skipped = s.skippedCount ?? 0
    const bytesFmt = formatBytes(Number(s.totalBytes || 0))
    const txFmt = formatBytes(Number(s.transferredBytes || 0))
    const spFmt = formatBytesPerSec(Number(s.avgSpeedBps || 0))
    const dur = p?.run?.durationText || '-'
    let md = `**${t('task.title')}** <font color="info">${taskName}</font> ${statusZh}（${triggerZh}）\n`
    md += `> ${t('runtime.webhookTotal')}: ${total}  ${okLabel}: ${ok}  ${t('runtime.webhookFailed')}: ${fail}  ${t('runtime.webhookOther')}: ${skipped}\n`
    md += `> ${t('runtime.webhookBytes')}: ${bytesFmt} / ${t('runtime.webhookTransferred')}: ${txFmt}\n`
    md += `> ${t('runtime.webhookAvgSpeed')}: ${spFmt}  ${t('runtime.webhookDuration')}: ${dur}\n`
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
        options.showToast(t('runtime.webhookNeedUrl'), 'error')
        return
      }
      const payload: any = {
        title: t('runtime.webhookTestTitle'),
        triggerZh: t('runtime.webhookTestTrigger'),
        statusZh: t('runtime.webhookTestStatus'),
        summaryZh: t('runtime.webhookTestSummary'),
        task: { id: 0, name: t('runtime.webhookTestTask'), mode: 'test' },
        run: { id: 0, trigger: 'manual', status: 'success', startedAt: new Date().toISOString(), finishedAt: new Date().toISOString(), durationSeconds: 1, durationText: `1${t('runtime.secondSuffix')}` },
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
      if (url1) { try { await send(url1) } catch (e: any) { fails.push(`Webhook: ${e?.message || e}`) } }
      if (url2) { try { await send(url2) } catch (e: any) { fails.push(`WeCom: ${e?.message || e}`) } }
      if (fails.length) {
        options.showToast(t('runtime.webhookPartialFail').replace('{message}', fails.join('; ')), 'error')
      } else {
        options.showToast(t('runtime.webhookSent'), 'success')
      }
    } catch (e: any) {
      options.showToast(t('runtime.webhookSendFail').replace('{message}', e?.message || e), 'error')
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
