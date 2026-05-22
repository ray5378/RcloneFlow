<script setup lang="ts">
import type { Provider } from '../../types'

defineProps<{
  selectedProvider: Provider | null
  remoteName: string
  creating: boolean
  success: boolean
  editMode?: boolean
}>()

const emit = defineEmits<{
  create: []
  close: []
}>()
</script>

<template>
  <div>
    <div class="card" style="background: #f0f9ff; margin-bottom: 16px">
      <div><strong>{{ $t('remote.typeLabel') }}:</strong> {{ selectedProvider?.Name }}</div>
      <div><strong>{{ $t('remote.nameLabel') }}:</strong> {{ remoteName }}</div>
    </div>
    <div class="actions" style="justify-content: space-between">
      <button v-if="!editMode" class="ghost" :disabled="success" @click="emit('close')">{{ $t('remote.previous') }}</button>
      <button v-if="!success" color="primary" :disabled="creating" @click="emit('create')">
        {{ creating ? $t('remote.saving') : (editMode ? $t('remote.saveEdit') : $t('remote.save')) }}
      </button>
      <button v-else color="primary" @click="emit('close')">
        {{ editMode ? $t('remote.editSuccessClose') : $t('remote.saveSuccessClose') }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.card { padding: 14px; border: 1px solid #d1d5db; border-radius: 12px; }
.actions { display: flex; gap: 8px; }
button { padding: 10px 14px; border-radius: 10px; border: none; background: #2563eb; color: white; cursor: pointer; }
button.ghost { background: #eef2ff; color: #1d4ed8; }
button:disabled { opacity: .6; cursor: not-allowed; }
</style>