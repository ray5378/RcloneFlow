<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import Modal from './Modal.vue'

const props = defineProps<{ modelValue: boolean, taskId?: number }>()
const emit = defineEmits<{ 'update:modelValue': [boolean] }>()

const show = ref(false)
watch(() => props.modelValue, v => show.value = v, { immediate: true })

function close(){ emit('update:modelValue', false) }

// form state
const g = ref({
  postVerifyEnabled: true,
  postVerifyMode: 'mount',
  postVerifyInterval: '5s',
  postVerifyTimeout: '30m',
  postVerifyMtimeGrace: '60s',
  postVerifyMatch: 'size',
  minRerunInterval: '30m',
})

async function loadGlobal(){
  const res = await fetch('/api/settings/transfer', { headers: auth() })
  if(res.ok){ Object.assign(g.value, await res.json()) }
}

async function saveGlobal(){
  const res = await fetch('/api/settings/transfer', {
    method: 'PUT', headers: { 'Content-Type': 'application/json', ...auth() }, body: JSON.stringify(g.value)
  })
  if(!res.ok) throw new Error('保存失败')
}

async function saveTask(){
  if(!props.taskId) throw new Error('缺少任务ID')
  const opts:any = {
    'postVerify.enabled': g.value.postVerifyEnabled,
    'postVerify.mode': g.value.postVerifyMode,
    'postVerify.interval': g.value.postVerifyInterval,
    'postVerify.timeout': g.value.postVerifyTimeout,
    'postVerify.mtimeGrace': g.value.postVerifyMtimeGrace,
    'postVerify.match': g.value.postVerifyMatch,
    'minRerunInterval': g.value.minRerunInterval,
  }
  const res = await fetch('/api/tasks', {
    method: 'PATCH', headers: { 'Content-Type': 'application/json', ...auth() }, body: JSON.stringify({ id: props.taskId, options: opts })
  })
  if(!res.ok) throw new Error('保存失败')
}

function auth(){
  try{ const t = localStorage.getItem('authToken') || ''; return t? { Authorization: 'Bearer '+t }: {} }catch{ return {} }
}

onMounted(loadGlobal)
</script>

<template>
  <Modal :show="show" title="传输选项" @close="close">
    <div class="transfer-grid">
      <div class="row"><label><input v-model="g.postVerifyEnabled" type="checkbox"/> 启用收尾校验</label></div>
      <div class="row"><label>校验模式</label>
        <select v-model="g.postVerifyMode"><option value="mount">挂载路径</option></select>
      </div>
      <div class="row"><label>轮询间隔</label><input v-model="g.postVerifyInterval" type="text" placeholder="5s"/></div>
      <div class="row"><label>超时时间</label><input v-model="g.postVerifyTimeout" type="text" placeholder="30m"/></div>
      <div class="row"><label>mtime 稳定期</label><input v-model="g.postVerifyMtimeGrace" type="text" placeholder="60s"/></div>
      <div class="row"><label>匹配方式</label>
        <select v-model="g.postVerifyMatch"><option value="size">按大小</option></select>
      </div>
      <div class="row"><label>最短重跑间隔</label><input v-model="g.minRerunInterval" type="text" placeholder="30m"/></div>
    </div>
    <div class="modal-footer">
      <button class="ghost" @click="close">取消</button>
      <button class="primary" @click="props.taskId? saveTask(): saveGlobal()">保存</button>
    </div>
  </Modal>
</template>

<style scoped>
.transfer-grid{ display:grid; grid-template-columns:1fr 1fr; gap:12px; padding:12px }
.row{ display:flex; flex-direction:column; gap:6px }
.primary{ background:#2563eb; color:#fff; border:none; padding:8px 12px; border-radius:8px }
.ghost{ background:#263043; color:#d1d5db; border:none; padding:8px 12px; border-radius:8px; margin-right:8px }
.modal-footer{ display:flex; justify-content:flex-end; gap:8px; padding:12px }
</style>
