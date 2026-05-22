import { ref } from 'vue'
import { getToken } from '../api/auth'
import { t } from '../i18n'

export function useRunLogModal() {
  const showLogModal = ref(false)
  const logModalTitle = ref(t('modal.transferLog'))
  const logContent = ref('')

  async function openRunLog(run: any) {
    logModalTitle.value = `${t('modal.transferLog')} #${run.id}`
    logContent.value = t('modal.loading')
    showLogModal.value = true
    try {
      const token = getToken() || ''
      const resp = await fetch(`/api/runs/${run.id}/log`, {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      })
      if (!resp.ok) {
        const txt = await resp.text()
        logContent.value = `${t('modal.loadFailed')} ${resp.status} ${txt}`
        return
      }
      const buf = await resp.arrayBuffer()
      const dec = new TextDecoder('utf-8')
      logContent.value = dec.decode(buf)
    } catch (e: any) {
      logContent.value = `${t('modal.loadError')} ${e?.message || e}`
    }
  }

  return {
    showLogModal,
    logModalTitle,
    logContent,
    openRunLog,
  }
}
