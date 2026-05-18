<script setup lang="ts">
import { computed } from 'vue'
import { t } from '../../i18n'

const props = defineProps<{
  page: number
  totalPages: number
  jumpPage: number | null
}>()

const emit = defineEmits<{
  (e: 'prev'): void
  (e: 'next'): void
  (e: 'jump'): void
  (e: 'first'): void
  (e: 'last'): void
  (e: 'update:jumpPage', value: number | null): void
}>()

function onJumpInput(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:jumpPage', target.value === '' ? null : Number(target.value))
}

function pageText(page: number, total: number) {
  return t('runtime.pageXofY').replace('{page}', String(page)).replace('{total}', String(total))
}

const pageItems = computed<(number | string)[]>(() => {
  const total = props.totalPages
  const current = props.page
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }

  const items: (number | string)[] = [1]
  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)

  if (start > 2) items.push('left-ellipsis')
  for (let i = start; i <= end; i++) {
    items.push(i)
  }
  if (end < total - 1) items.push('right-ellipsis')

  items.push(total)
  return items
})
</script>

<template>
  <div class="pagination">
    <div class="page-nav">
      <button class="page-btn edge-btn" :disabled="page <= 1" @click="emit('first')">«</button>

      <button class="page-btn nav-btn" :disabled="page <= 1" @click="emit('prev')">
        <span aria-hidden="true">←</span>
        <span>{{ t('runtime.prevPage') }}</span>
      </button>

      <div class="page-numbers">
        <template v-for="item in pageItems" :key="String(item)">
          <span v-if="typeof item !== 'number'" class="page-ellipsis">…</span>
          <button
            v-else
            class="page-btn page-number"
            :class="{ active: item === page }"
            :disabled="item === page"
            @click="emit('update:jumpPage', item); emit('jump')"
          >
            {{ item }}
          </button>
        </template>
      </div>

      <button class="page-btn nav-btn" :disabled="page >= totalPages" @click="emit('next')">
        <span>{{ t('runtime.nextPage') }}</span>
        <span aria-hidden="true">→</span>
      </button>

      <button class="page-btn edge-btn" :disabled="page >= totalPages" @click="emit('last')">»</button>

      <div class="page-jump">
        <span class="jump-label">#</span>
        <input
          type="number"
          class="page-input"
          :value="jumpPage ?? ''"
          min="1"
          :max="totalPages"
          @input="onJumpInput"
          @keyup.enter="emit('jump')"
        />
        <button class="page-btn jump-btn" @click="emit('jump')">{{ t('runtime.jump') }}</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  flex-wrap: wrap;
  padding-top: 10px;
}

.page-nav {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  flex-wrap: wrap;
}

.page-numbers {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.page-btn {
  min-height: 34px;
  padding: 0 12px;
  border-radius: 10px;
  font-size: 13px;
  background: transparent;
  border: 1px solid var(--border);
  color: var(--text);
  cursor: pointer;
  transition: all 0.18s ease;
}

.page-btn:hover:not(:disabled) {
  border-color: var(--accent);
  color: var(--accent);
  transform: translateY(-1px);
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

.nav-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.edge-btn {
  min-width: 38px;
  padding: 0 10px;
}

.page-number {
  min-width: 34px;
  padding: 0 10px;
}

.page-number.active {
  background: var(--accent, #4f46e5);
  border-color: var(--accent, #4f46e5);
  color: #fff;
  opacity: 1;
}

.page-ellipsis {
  width: 20px;
  text-align: center;
  color: var(--muted, #999);
  user-select: none;
}

.page-jump {
  display: flex;
  align-items: center;
  gap: 8px;
}

.jump-label {
  font-size: 12px;
  color: var(--muted, #999);
}

.page-input {
  width: 68px;
  min-width: 68px;
  height: 34px;
  padding: 0 10px;
  border: 1px solid #333;
  border-radius: 10px;
  background: #252525;
  color: #e0e0e0;
  box-sizing: border-box;
}

.page-input:focus {
  outline: none;
  border-color: var(--accent, #4f46e5);
}

.jump-btn {
  white-space: nowrap;
}

body.light .page-input {
  background: #fff;
  color: #111827;
  border-color: #ddd;
}

@media (max-width: 768px) {
  .page-nav,
  .page-jump {
    justify-content: center;
  }

  .page-jump {
    flex-wrap: wrap;
  }

  .page-btn {
    min-height: 36px;
  }
}
</style>
