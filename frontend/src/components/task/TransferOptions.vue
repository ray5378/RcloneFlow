<script setup lang="ts">
import { computed } from 'vue'
import type { TaskFormOptions } from './types'
import { t } from '../../i18n'

const props = defineProps<{ modelValue: TaskFormOptions }>()
const emit = defineEmits<{ 'update:modelValue': [value: TaskFormOptions] }>()

const options = computed<TaskFormOptions>({
  get: () => props.modelValue || {},
  set: (value) => emit('update:modelValue', value),
})

function updateField(key: string, value: string) {
  options.value = { ...options.value, [key]: value }
}
</script>

<template>
  <div class="section compact">
    <div class="section-title">{{ t('transferOptions.title') }}</div>
    <div class="field-grid compact-grid">
      <div class="field-item">
        <label>{{ t('transferOptions.transfers') }}</label>
        <input :value="options.transfers || ''" type="number" min="1" @input="updateField('transfers', ($event.target as HTMLInputElement).value)" />
      </div>
      <div class="field-item">
        <label>{{ t('transferOptions.checkers') }}</label>
        <input :value="options.checkers || ''" type="number" min="1" @input="updateField('checkers', ($event.target as HTMLInputElement).value)" />
      </div>
      <div class="field-item">
        <label>{{ t('transferOptions.maxDelete') }}</label>
        <input :value="options.maxDelete || ''" type="number" min="0" @input="updateField('maxDelete', ($event.target as HTMLInputElement).value)" />
      </div>
      <div class="field-item">
        <label>{{ t('transferOptions.maxTransfer') }}</label>
        <input :value="options.maxTransfer || ''" type="text" placeholder="10G / 500M" @input="updateField('maxTransfer', ($event.target as HTMLInputElement).value)" />
      </div>
      <div class="field-item">
        <label>{{ t('transferOptions.rateLimit') }}</label>
        <input :value="options.bwlimit || ''" type="text" placeholder="08:00,10M 23:00,off" @input="updateField('bwlimit', ($event.target as HTMLInputElement).value)" />
      </div>
      <div class="field-item field-item-wide">
        <label>{{ t('transferOptions.extraArgs') }}</label>
        <input :value="options.extraArgs || ''" type="text" placeholder="--fast-list --size-only" @input="updateField('extraArgs', ($event.target as HTMLInputElement).value)" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.section.compact { margin-top: 0; }
.section-title { font-size: 13px; color: var(--muted); margin-bottom: 12px; font-weight: 600; }
.compact-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
.field-item { display: flex; flex-direction: column; gap: 6px; }
.field-item-wide { grid-column: 1 / -1; }
.field-item label { font-size: 12px; color: var(--muted); }
.field-item input { width: 100%; padding: 10px 12px; border: 1px solid var(--border); border-radius: 8px; background: var(--surface); color: var(--text); box-sizing: border-box; }
</style>
