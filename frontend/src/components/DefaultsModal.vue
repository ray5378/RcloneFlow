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
            <label>访问令牌有效期 <small class="subkey">ACCESS_TOKEN_TTL</small></label>
            <input v-model="form.ACCESS_TOKEN_TTL" placeholder="如 24h" />
            <label>刷新令牌有效期 <small class="subkey">REFRESH_TOKEN_TTL</small></label>
            <input v-model="form.REFRESH_TOKEN_TTL" placeholder="如 90d" />
          </div>
        </div>

        <div class="section">
          <div class="section-title">日志与等级（实时生效）</div>
          <div class="grid">
            <label>日志级别 <small class="subkey">LOG_LEVEL</small></label>
            <select v-model="form.LOG_LEVEL">
              <option value="debug">调试</option>
              <option value="info">信息</option>
              <option value="warn">警告</option>
              <option value="error">错误</option>
            </select>
            <label>日志输出 <small class="subkey">LOG_OUTPUT</small></label>
            <input v-model="form.LOG_OUTPUT" placeholder="stdout" />
          </div>
        </div>

        <div class="section">
          <div class="section-title">历史与保留（定期清理）</div>
          <div class="grid">
            <label>运行总结保留天数 <small class="subkey">FINAL_SUMMARY_RETENTION_DAYS</small></label>
            <input v-model="form.FINAL_SUMMARY_RETENTION_DAYS" type="number" min="0" />
            <label>清理扫描间隔（小时） <small class="subkey">CLEANUP_INTERVAL_HOURS</small></label>
            <input v-model="form.CLEANUP_INTERVAL_HOURS" type="number" min="0" />
          </div>
        </div>

        <div class="section">
          <div class="section-title">预检/统计概览（任务启动前）</div>
          <div class="grid">
            <label>预检模式 <small class="subkey">PRECHECK_MODE</small></label>
            <select v-model="form.PRECHECK_MODE">
              <option value="none">关闭</option>
              <option value="size">按大小预检</option>
            </select>
          </div>
        </div>

        <div class="section">
          <div class="section-title">运行中进度限流写库（热生效）</div>
          <div class="grid">
            <label>写库最小间隔 <small class="subkey">PROGRESS_FLUSH_INTERVAL</small></label>
            <input v-model="form.PROGRESS_FLUSH_INTERVAL" placeholder="如 5s" />
            <label>写库最小增量（百分比） <small class="subkey">PROGRESS_FLUSH_MIN_DELTA_PCT</small></label>
            <input v-model="form.PROGRESS_FLUSH_MIN_DELTA_PCT" type="number" min="0" step="0.1" />
            <label>写库最小增量（字节） <small class="subkey">PROGRESS_FLUSH_MIN_DELTA_BYTES</small></label>
            <input v-model="form.PROGRESS_FLUSH_MIN_DELTA_BYTES" type="number" min="0" />
          </div>
        </div>

        <div class="section">
          <div class="section-title">WebDAV 收尾（完成确认）</div>
          <div class="grid">
            <label>收尾轮询间隔 <small class="subkey">FINISH_WAIT_INTERVAL</small></label>
            <input v-model="form.FINISH_WAIT_INTERVAL" placeholder="如 5s" />
            <label>收尾超时时间 <small class="subkey">FINISH_WAIT_TIMEOUT</small></label>
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
.subkey{opacity:.6;margin-left:6px;font-weight:400}
</style>
