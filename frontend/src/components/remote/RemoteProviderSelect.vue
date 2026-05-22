<script setup lang="ts">
import type { Provider } from '../../types'
import { getProviderDescription } from './remoteUtils'

defineProps<{
  providers: Provider[]
  search: string
  selectedProviderName: string
}>()

const emit = defineEmits<{
  'update:search': [value: string]
  select: [provider: Provider]
}>()
</script>

<template>
  <div>
    <input
      :value="search"
      type="text"
      :placeholder="$t('remote.searchPlaceholder')"
      style="width: 100%; margin-bottom: 16px"
      @input="emit('update:search', ($event.target as HTMLInputElement).value)"
    />
    <div class="provider-grid">
      <div
        v-for="p in providers"
        :key="p.Name"
        class="provider-card"
        :class="{ selected: selectedProviderName === p.Name }"
        @click="emit('select', p)"
      >
        <strong>{{ p.Name }}</strong>
        <div style="margin-top: 6px; color: #6b7280; font-size: 12px">{{ getProviderDescription(p) }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.provider-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 12px; max-height: 360px; overflow: auto; }
.provider-card { padding: 14px; border: 1px solid #374151; border-radius: 12px; cursor: pointer; background: #1f2937; color: #e5e7eb; }
.provider-card:hover { border-color: #60a5fa; background: #111827; }
.provider-card.selected { border-color: #2563eb; background: #111827; }
body.light .provider-card { border-color: #e5e7eb; background: #fff; color: #111827; }
body.light .provider-card:hover { border-color: #93c5fd; background: #eff6ff; }
body.light .provider-card.selected { border-color: #2563eb; background: #eff6ff; }
</style>