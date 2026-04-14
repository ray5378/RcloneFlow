<script setup lang="ts">
import { ref, watch } from 'vue'
import Modal from './Modal.vue'

const props = defineProps<{
  show: boolean
  remoteName: string
  description: string
}>()

const emit = defineEmits<{
  close: []
  save: [desc: string]
}>()

const text = ref('')

watch(() => props.show, (val) => {
  if (val) {
    text.value = props.description
  }
})

function save() {
  emit('save', text.value)
  emit('close')
}
</script>

<template>
  <Modal :show="show" title="自定义介绍" @close="emit('close')">
    <div class="card" style="background: #f0f9ff; margin-bottom: 16px">
      <div><strong>存储名称:</strong> {{ remoteName }}</div>
    </div>
    <div class="field-item">
      <label>介绍文字</label>
      <input v-model="text" type="text" placeholder="输入该存储的自定义介绍文字..." style="width: 100%" />
      <div class="help">设置后将在存储节点页面显示</div>
    </div>
    <div class="actions" style="margin-top: 16px; justify-content: flex-end">
      <button class="ghost" @click="emit('close')">取消</button>
      <button @click="save">保存</button>
    </div>
  </Modal>
</template>
