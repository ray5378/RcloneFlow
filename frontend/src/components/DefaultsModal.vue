<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getSettings, saveSettings, resetSettings } from '../api/settings'

const loading = ref(true)
const saving = ref(false)
const data = ref<any>(null)
const form = ref<Record<string,string>>({})

function flat(resp:any){
  const out:Record<string,string> = {}
  const patch = (grp:any)=>{ Object.keys(grp||{}).forEach(k=> out[k] = grp[k]?.effective??'') }
  patch(resp.auth); patch(resp.log); patch(resp.history); patch(resp.precheck); patch(resp.progress); patch(resp.webdav)
  return out
}

async function load(){
  loading.value = true
  try{ const resp = await getSettings(); data.value = resp; form.value = flat(resp) } finally { loading.value=false }
}

async function onSave(){
  saving.value = true
  try{ await saveSettings(form.value); alert('已保存并生效'); await load() } catch(e:any){ alert(e?.message||e) } finally { saving.value=false }
}
async function onReset(){
  if (!confirm('确认重置为默认？')) return
  saving.value = true
  try{ await resetSettings(); alert('已重置为默认'); await load() } catch(e:any){ alert(e?.message||e) } finally { saving.value=false }
}

onMounted(load)
</script>

<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content large">
      <div class="modal-header">
        <h3>修改默认设置</h3>
        <button class="close-btn" @click="$emit('close')">×</button>
      </div>
      <div class="modal-body" v-if="!loading">
        <div class="section">
          <div class="section-title">认证</div>
          <div class="grid">
            <label>ACCESS_TOKEN_TTL</label>
            <input v-model="form.ACCESS_TOKEN_TTL" placeholder="如 24h" />
            <label>REFRESH_TOKEN_TTL</label>
            <input v-model="form.REFRESH_TOKEN_TTL" placeholder="如 90d" />
          </div>
        </div>

        <div class="section">
          <div class="section-title">日志与等级</div>
          <div class="grid">
            <label>LOG_LEVEL</label>
            <select v-model="form.LOG_LEVEL">
              <option value="debug">debug</option>
              <option value="info">info</option>
              <option value="warn">warn</option>
              <option value="error">error</option>
            </select>
            <label>LOG_OUTPUT</label>
            <input v-model="form.LOG_OUTPUT" placeholder="stdout" />
          </div>
        </div>

        <div class="section">
          <div class="section-title">历史与保留</div>
          <div class="grid">
            <label>FINAL_SUMMARY_RETENTION_DAYS</label>
            <input v-model="form.FINAL_SUMMARY_RETENTION_DAYS" type="number" min="0" />
            <label>CLEANUP_INTERVAL_HOURS</label>
            <input v-model="form.CLEANUP_INTERVAL_HOURS" type="number" min="0" />
          </div>
        </div>

        <div class="section">
          <div class="section-title">预检/统计概览</div>
          <div class="grid">
            <label>PRECHECK_MODE</label>
            <select v-model="form.PRECHECK_MODE">
              <option value="none">none</option>
              <option value="size">size</option>
            </select>
          </div>
        </div>

        <div class="section">
          <div class="section-title">运行中进度限流写库</div>
          <div class="grid">
            <label>PROGRESS_FLUSH_INTERVAL</label>
            <input v-model="form.PROGRESS_FLUSH_INTERVAL" placeholder="如 5s" />
            <label>PROGRESS_FLUSH_MIN_DELTA_PCT</label>
            <input v-model="form.PROGRESS_FLUSH_MIN_DELTA_PCT" type="number" min="0" step="0.1" />
            <label>PROGRESS_FLUSH_MIN_DELTA_BYTES</label>
            <input v-model="form.PROGRESS_FLUSH_MIN_DELTA_BYTES" type="number" min="0" />
          </div>
        </div>

        <div class="section">
          <div class="section-title">WebDAV 收尾</div>
          <div class="grid">
            <label>FINISH_WAIT_INTERVAL</label>
            <input v-model="form.FINISH_WAIT_INTERVAL" placeholder="如 5s" />
            <label>FINISH_WAIT_TIMEOUT</label>
            <input v-model="form.FINISH_WAIT_TIMEOUT" placeholder="如 5h" />
          </div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="onReset" :disabled="saving">重置为默认</button>
        <button class="primary" @click="onSave" :disabled="saving">保存</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-content.large{ max-width: 720px; width: 92vw; }
.section{ margin-bottom: 16px; }
.section-title{ font-weight: 700; margin-bottom: 8px; color: #e0e0e0 }
body.light .section-title{ color: #111827 }
.grid{ display:grid; grid-template-columns: 220px 1fr; gap: 8px 12px }
input, select{ padding: 8px 10px; border-radius: 8px; border: 1px solid #333; background:#252525; color:#e0e0e0 }
body.light input, body.light select{ background:#fff; color:#111; border-color:#ddd }
</style>
