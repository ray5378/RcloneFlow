<script setup lang="ts">
defineProps<{
  remoteName: string
}>()

const emit = defineEmits<{
  'update:remoteName': [value: string]
}>()
</script>

<template>
  <div>
    <div class="field-item">
      <label>{{ $t('remote.nameLabel') }} <span style="color: #dc2626">*</span></label>
      <input
        :value="remoteName"
        type="text"
        :placeholder="$t('remote.namePlaceholder')"
        @input="emit('update:remoteName', ($event.target as HTMLInputElement).value)"
      />
      <div class="field-help">{{ $t('remote.nameHelp') }}</div>
    </div>
  </div>
</template>

<style scoped>
.field-item { display: grid; gap: 8px; }
.field-item label { font-weight: 600; }
.field-item input { width: 100%; padding: 10px 12px; border-radius: 10px; border: 1px solid #d1d5db; font: inherit; }
.field-help { font-size: 12px; color: #6b7280; line-height: 1.5; }
</style>