<script setup lang="ts">
import { ref, watch } from 'vue'
import Modal from './Modal.vue'
import { t } from '../i18n'

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
  <Modal :show="show" :title="t('modal.descTitle')" @close="emit('close')">
    <div class="card" style="background: #f0f9ff; margin-bottom: 16px">
      <div><strong>{{ t('modal.storageName') }}:</strong> {{ remoteName }}</div>
    </div>
    <div class="field-item">
      <label>{{ t('modal.descLabel') }}</label>
      <input v-model="text" type="text" :placeholder="t('modal.descPlaceholder')" style="width: 100%" />
      <div class="help">{{ t('modal.descHelp') }}</div>
    </div>
    <div class="actions" style="margin-top: 16px; justify-content: flex-end">
      <button class="ghost" @click="emit('close')">{{ t('modal.cancel') }}</button>
      <button @click="save">{{ t('modal.save') }}</button>
    </div>
  </Modal>
</template>
