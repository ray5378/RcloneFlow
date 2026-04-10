<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import Modal from './Modal.vue'

const props = defineProps<{ modelValue: boolean, taskId?: number }>()
const emit = defineEmits<{ 'update:modelValue': [boolean] }>()

const show = ref(false)
watch(() => props.modelValue, v => show.value = v, { immediate: true })

function close(){ emit('update:modelValue', false) }

// form state（移除收尾校验相关字段，仅保留通用传输参数容器）
const g = ref<Record<string, any>>({})

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
  const opts:any = {}
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
