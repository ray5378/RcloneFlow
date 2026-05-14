<script setup lang="ts">
import type { TaskFormOptions, UpdateTaskOption } from './types'
import { t } from '../../i18n'

defineProps<{
  options: TaskFormOptions
  updateOption: UpdateTaskOption
}>()
</script>

<template>
  <div class="advanced-group">
    <div class="advanced-group-title">{{ t('advancedTask.transferStrategy') }}</div>
    <div class="advanced-row inline">
      <input type="checkbox" :checked="options.enableStreaming" @change="updateOption('enableStreaming', ($event.target as HTMLInputElement).checked)" />
      <label>{{ t('advancedTask.enableStreaming') }}</label>
    </div>
  </div>

  <div class="advanced-group">
    <div class="advanced-group-title">{{ t('advancedTask.transferControl') }}</div>
    <div class="advanced-row">
      <label>{{ t('advancedTask.concurrentTransfers') }}</label>
      <input type="number" :value="options.transfers" @input="updateOption('transfers', Number(($event.target as HTMLInputElement).value))" min="1" max="100" />
    </div>
    <div class="advanced-row">
      <label>{{ t('advancedTask.bandwidthLimit') }}</label>
      <input type="text" :value="options.bwLimit" @input="updateOption('bwLimit', ($event.target as HTMLInputElement).value)" :placeholder="t('advancedTask.bandwidthPlaceholder')" />
    </div>
    <div class="advanced-row inline">
      <input type="checkbox" :checked="options.multiThreadStreams" @change="updateOption('multiThreadStreams', ($event.target as HTMLInputElement).checked)" />
      <label>{{ t('advancedTask.multiThread') }}</label>
    </div>
    <div class="advanced-row">
      <label>{{ t('advancedTask.maxTransfer') }}</label>
      <input type="number" :value="options.maxTransfer" @input="updateOption('maxTransfer', Number(($event.target as HTMLInputElement).value))" min="0" :placeholder="t('advancedTask.maxTransferPlaceholder')" />
    </div>
    <div class="advanced-row">
      <label>{{ t('advancedTask.maxDuration') }}</label>
      <input type="number" :value="options.maxDuration" @input="updateOption('maxDuration', Number(($event.target as HTMLInputElement).value))" min="0" :placeholder="t('advancedTask.maxDurationPlaceholder')" />
    </div>
  </div>
</template>

<style scoped>
.advanced-group { margin-bottom: 16px; padding-bottom: 16px; border-bottom: 1px solid #2a2a2a; }
body.light .advanced-group { border-bottom-color: #eee; }
.advanced-group-title { font-weight: 600; font-size: 13px; color: #64b5f6; margin-bottom: 12px; }
.advanced-row { margin-bottom: 12px; }
.advanced-row label { display: block; font-size: 12px; color: #888; margin-bottom: 4px; }
.advanced-row input[type="text"], .advanced-row input[type="number"] { width: 100%; padding: 8px 12px; border: 1px solid #333; border-radius: 8px; background: #252525; color: #e0e0e0; font-size: 13px; box-sizing: border-box; font-family: inherit; }
body.light .advanced-row input[type="text"], body.light .advanced-row input[type="number"] { background: #fff; border-color: #ddd; color: #1a1a1a; }
.advanced-row input:focus { outline: none; border-color: #64b5f6; }
.advanced-row.inline { display: flex; align-items: center; gap: 8px; }
.advanced-row.inline label { margin-bottom: 0; display: inline; }
.advanced-row.inline input[type="checkbox"] { width: 16px; height: 16px; flex: 0 0 auto; }
</style>
