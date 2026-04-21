<script setup lang="ts">
import { t } from '../../i18n'

defineProps<{
  page: number
  totalPages: number
  jumpPage: number | null
}>()

const emit = defineEmits<{
  (e: 'prev'): void
  (e: 'next'): void
  (e: 'jump'): void
  (e: 'update:jumpPage', value: number | null): void
}>()

function onJumpInput(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:jumpPage', target.value === '' ? null : Number(target.value))
}

function pageText(page: number, total: number) {
  return t('runtime.pageXofY').replace('{page}', String(page)).replace('{total}', String(total))
}
</script>

<template>
  <div class="pagination">
    <span class="page-current">{{ pageText(page, totalPages) }}</span>
    <button class="page-btn" :disabled="page <= 1" @click="emit('prev')">{{ t('runtime.prevPage') }}</button>
    <button class="page-btn" :disabled="page >= totalPages" @click="emit('next')">{{ t('runtime.nextPage') }}</button>
    <input type="number" class="page-input" :value="jumpPage ?? ''" min="1" :max="totalPages" @input="onJumpInput" @keyup.enter="emit('jump')" />
    <button class="page-btn" @click="emit('jump')">{{ t('runtime.jump') }}</button>
  </div>
</template>
