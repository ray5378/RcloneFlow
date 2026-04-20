<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getSettings, saveSettings, resetSettings } from '../api/settings'

const emit = defineEmits<{ (e: 'close'): void }>()

const loading = ref(true)
const saving = ref(false)
const saved = ref(false)
const showResetConfirm = ref(false)
const data = ref<any>(null)
const form = ref<Record<string,string>>({})
const errors = ref<Record<string,string>>({})
const saveFailed = ref(false)
const durationRe = /^\s*\d+\s*(ms|s|m|h|d)\s*$/i

function validate(){
  errors.value = {}
  // 数字字段 >=0（整数）
  const intFields = ['FINAL_SUMMARY_RETENTION_DAYS','CLEANUP_INTERVAL_HOURS','WEBHOOK_MAX_FILES','LOG_RETENTION_DAYS']
  for (const k of intFields){
    const v = (form.value as any)[k]
    if (v!=='' && (isNaN(Number(v)) || !Number.isFinite(Number(v)) || Number(v) < 0)){
      errors.value[k] = '请输入大于或等于 0 的数字'
    }
  }
  // MB 字段单独校验，可小数
  const mb = (form.value as any)['PROGRESS_FLUSH_MIN_DELTA_BYTES']
  if (mb!=='' && (isNaN(Number(mb)) || Number(mb) < 0)){
    errors.value['PROGRESS_FLUSH_MIN_DELTA_BYTES'] = '请输入大于或等于 0 的数字（单位：MB，可小数）'
  }
  // 百分比 0-100
  const pct = (form.value as any)['PROGRESS_FLUSH_MIN_DELTA_PCT']
  if (pct!=='' && (isNaN(Number(pct)) || Number(pct) < 0 || Number(pct) > 100)){
    errors.value['PROGRESS_FLUSH_MIN_DELTA_PCT'] = '请输入 0-100 之间的数字（可带小数）'
  }
  // 时长字段
  const durFields = ['ACCESS_TOKEN_TTL','REFRESH_TOKEN_TTL','PROGRESS_FLUSH_INTERVAL','FINISH_WAIT_INTERVAL','FINISH_WAIT_TIMEOUT']
  for (const k of durFields){
    const v = (form.value as any)[k]
    if (v && !durationRe.test(String(v))){
      errors.value[k] = '格式应为 数字+单位（ms/s/m/h/d），例如：5s、24h、90d'
    }
  }
  return Object.keys(errors.value).length === 0
}

function flat(resp:any){
  const out:Record<string,string> = {}
  const patch = (grp:any)=>{ Object.keys(grp||{}).forEach(k=> out[k] = grp[k]?.effective??'') }
  patch(resp.auth); patch(resp.log); patch(resp.history); patch(resp.precheck); patch(resp.progress); patch(resp.webdav); patch(resp.webhook)
  return out
}

async function load(){
  loading.value = true
  try{
    const resp = await getSettings();
    data.value = resp;
    form.value = flat(resp);
    // 将字节转为 MB（保留两位小数显示）
    const b = Number((form.value as any).PROGRESS_FLUSH_MIN_DELTA_BYTES || 0)
    if (!isNaN(b) && isFinite(b) && b>0){
      (form.value as any).PROGRESS_FLUSH_MIN_DELTA_BYTES = String(Math.round((b/1048576)*100)/100)
    }
  } finally { loading.value=false }
}

async function onSave(){
  if (!validate()) return
  // 将 MB 转回字节提交
  // 统一把所有字段转成字符串，避免后端 map[string]string 丢键
  const payload: Record<string,string> = {}
  Object.keys(form.value).forEach(k => {
    const v = (form.value as any)[k]
    payload[k] = v === undefined || v === null ? '' : String(v)
  })
  // 将 MB 转回字节提交
  const mb = Number(payload.PROGRESS_FLUSH_MIN_DELTA_BYTES)
  if (!isNaN(mb) && isFinite(mb) && mb>=0){
    payload.PROGRESS_FLUSH_MIN_DELTA_BYTES = String(Math.round(mb * 1048576))
  }
  saving.value = true
  saveFailed.value = false
  try{
    await saveSettings(payload)
    await load()
    saved.value = true
    setTimeout(()=>{ saved.value=false }, 10000)
  } catch(e:any){
    console.error(e)
    saveFailed.value = true
    setTimeout(()=>{ saveFailed.value=false }, 3000)
  } finally {
    saving.value=false
  }
}

function onSavedClick(){
  if (saved.value) emit('close')
}
async function onReset(){
  showResetConfirm.value = true
}
async function doConfirmReset(){
  saving.value = true
  try{
    await resetSettings()
    await load()
    saved.value = true
    setTimeout(()=>{ saved.value=false }, 10000)
  }catch(e:any){
    alert(e?.message||e)
  }finally{
    saving.value = false
    showResetConfirm.value = false
  }
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
            <label title="JWT 访问令牌有效期，示例：24h">访问令牌有效期 <small class="subkey">ACCESS_TOKEN_TTL</small></label>
            <input v-model="form.ACCESS_TOKEN_TTL" placeholder="如 24h" />
            <div class="error" v-if="errors.ACCESS_TOKEN_TTL">{{ errors.ACCESS_TOKEN_TTL }}</div>
            <label title="Refresh 令牌有效期，示例：90d">刷新令牌有效期 <small class="subkey">REFRESH_TOKEN_TTL</small></label>
            <input v-model="form.REFRESH_TOKEN_TTL" placeholder="如 90d" />
            <div class="error" v-if="errors.REFRESH_TOKEN_TTL">{{ errors.REFRESH_TOKEN_TTL }}</div>
          </div>
        </div>

        <div class="section">
          <div class="section-title">日志与等级（实时生效）</div>
          <div class="grid">
            <label title="日志级别（保存后即时生效）">日志级别 <small class="subkey">LOG_LEVEL</small></label>
            <select v-model="form.LOG_LEVEL">
              <option value="debug">调试</option>
              <option value="info">信息</option>
              <option value="warn">警告</option>
              <option value="error">错误</option>
            </select>
            <label title="运行日志文件保留天数">日志文件保留天数 <small class="subkey">LOG_RETENTION_DAYS</small></label>
            <input v-model="form.LOG_RETENTION_DAYS" type="number" min="1" placeholder="默认7天" />
            <div class="error" v-if="errors.LOG_RETENTION_DAYS">{{ errors.LOG_RETENTION_DAYS }}</div>
          </div>
        </div>

        <div class="section">
          <div class="section-title">历史与保留（定期清理）</div>
          <div class="grid">
            <label title="运行总结（finalSummary）保留天数">运行总结保留天数 <small class="subkey">FINAL_SUMMARY_RETENTION_DAYS</small></label>
            <input v-model="form.FINAL_SUMMARY_RETENTION_DAYS" type="number" min="0" />
            <div class="error" v-if="errors.FINAL_SUMMARY_RETENTION_DAYS">{{ errors.FINAL_SUMMARY_RETENTION_DAYS }}</div>
            <label title="清理任务执行间隔（小时）">清理扫描间隔（小时） <small class="subkey">CLEANUP_INTERVAL_HOURS</small></label>
            <input v-model="form.CLEANUP_INTERVAL_HOURS" type="number" min="0" />
            <div class="error" v-if="errors.CLEANUP_INTERVAL_HOURS">{{ errors.CLEANUP_INTERVAL_HOURS }}</div>
          </div>
        </div>

        <div class="section">
          <div class="section-title">运行中进度限流写库（热生效）</div>
          <div class="grid">
            <label title="进度写库最小时间间隔">写库最小间隔 <small class="subkey">PROGRESS_FLUSH_INTERVAL</small></label>
            <input v-model="form.PROGRESS_FLUSH_INTERVAL" placeholder="如 5s" />
            <div class="error" v-if="errors.PROGRESS_FLUSH_INTERVAL">{{ errors.PROGRESS_FLUSH_INTERVAL }}</div>
            <label title="两次写库的最小增量（百分比）">写库最小增量（百分比） <small class="subkey">PROGRESS_FLUSH_MIN_DELTA_PCT</small></label>
            <input v-model="form.PROGRESS_FLUSH_MIN_DELTA_PCT" type="number" min="0" step="0.1" />
            <div class="error" v-if="errors.PROGRESS_FLUSH_MIN_DELTA_PCT">{{ errors.PROGRESS_FLUSH_MIN_DELTA_PCT }}</div>
            <label title="两次写库的最小增量（MB）">写库最小增量（MB） <small class="subkey">PROGRESS_FLUSH_MIN_DELTA_BYTES</small></label>
            <input v-model="form.PROGRESS_FLUSH_MIN_DELTA_BYTES" type="number" min="0" step="0.01" />
            <div class="error" v-if="errors.PROGRESS_FLUSH_MIN_DELTA_BYTES">{{ errors.PROGRESS_FLUSH_MIN_DELTA_BYTES }}</div>
          </div>
        </div>

        <div class="section">
          <div class="section-title">WebDAV 收尾（完成确认）</div>
          <div class="grid">
            <label title="WebDAV 收尾轮询间隔">收尾轮询间隔 <small class="subkey">FINISH_WAIT_INTERVAL</small></label>
            <input v-model="form.FINISH_WAIT_INTERVAL" placeholder="如 5s" />
            <div class="error" v-if="errors.FINISH_WAIT_INTERVAL">{{ errors.FINISH_WAIT_INTERVAL }}</div>
            <label title="WebDAV 收尾最长等待">收尾超时时间 <small class="subkey">FINISH_WAIT_TIMEOUT</small></label>
            <input v-model="form.FINISH_WAIT_TIMEOUT" placeholder="如 5h" />
            <div class="error" v-if="errors.FINISH_WAIT_TIMEOUT">{{ errors.FINISH_WAIT_TIMEOUT }}</div>
          </div>
        </div>

        <div class="section">
          <div class="section-title">Webhook 通知</div>
          <div class="grid">
            <label title="Webhook POST 负载中的文件名数量上限；设为 0 表示不限制（传递所有文件名）">文件名上限（0=不限制） <small class="subkey">WEBHOOK_MAX_FILES</small></label>
            <input v-model="form.WEBHOOK_MAX_FILES" type="number" min="0" placeholder="0=不限制" />
          </div>
        </div>

        <!-- 文件管理后端固定：浏览=RC，操作=CLI；此处不再暴露切换选项 -->

      </div>
      <div class="modal-footer">
        <button class="ghost" @click="onReset" :disabled="saving">重置为默认</button>
        <button :class="['primary', { saved: saved, failed: saveFailed }]" @click="saved ? onSavedClick() : onSave()" :disabled="saving">{{ saved ? '已保存生效' : (saveFailed ? '保存失败' : '保存') }}</button>
      </div>
    </div>
  </div>

  <!-- 确认重置弹窗（与现有确认风格对齐） -->
  <div v-if="showResetConfirm" class="modal-overlay" @click.self="showResetConfirm=false">
    <div class="modal-content confirm-modal">
      <div class="modal-header">
        <h3>重置为默认</h3>
        <button class="close-btn" @click="showResetConfirm=false">×</button>
      </div>
      <div class="modal-body">
        <p>确定将所有设置重置为默认值？此操作不可撤销。</p>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="showResetConfirm=false">取消</button>
        <button class="primary danger" @click="doConfirmReset">确认</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-content.large{ max-width: 720px; width: 92vw; }
.section{ margin-bottom: 16px; }
.section-title{ font-weight: 700; margin-bottom: 8px; color: var(--text) }
.grid{ display:grid; grid-template-columns: 220px 1fr; gap: 8px 12px }
input, select{ padding: 8px 10px; border-radius: 8px; border: 1px solid var(--border); background: var(--surface); color: var(--text) }
.subkey{opacity:.6;margin-left:6px;font-weight:400;color:var(--muted)}
.modal-footer .primary.saved{ background: var(--success); color:#fff; border:none }
.modal-footer .primary.failed{ background: var(--danger); color:#fff; border:none }
.error{ color: var(--danger); font-size:12px; margin-top:4px }
</style>
