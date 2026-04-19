import { ref } from 'vue'
import { getToken } from '../api/auth'

export function useRunLogModal() {
  const showLogModal = ref(false)
  const logModalTitle = ref('传输日志')
  const logContent = ref('')

  async function openRunLog(run: any) {
    logModalTitle.value = `传输日志 #${run.id}`
    logContent.value = '加载中…'
    showLogModal.value = true
    try {
      const token = getToken() || ''
      const resp = await fetch(`/api/runs/${run.id}/log`, {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      })
      if (!resp.ok) {
        const fallback = await fetch(`/api/runs/${run.id}/log?auth=${token}`)
        if (!fallback.ok) {
          const txt = await fallback.text()
          logContent.value = `加载失败：${fallback.status} ${txt}`
          return
        }
        const buf2 = await fallback.arrayBuffer()
        const dec2 = new TextDecoder('utf-8')
        logContent.value = dec2.decode(buf2)
        return
      }
      const buf = await resp.arrayBuffer()
      const dec = new TextDecoder('utf-8')
      logContent.value = dec.decode(buf)
    } catch (e: any) {
      logContent.value = `加载异常：${e?.message || e}`
    }
  }

  return {
    showLogModal,
    logModalTitle,
    logContent,
    openRunLog,
  }
}
