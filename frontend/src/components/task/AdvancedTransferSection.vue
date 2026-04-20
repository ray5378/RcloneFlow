<script setup lang="ts">
defineProps<{
  options: any
  updateOption: (key: string, value: any) => void
}>()
</script>

<template>
  <div class="advanced-group">
    <div class="advanced-group-title">传输策略</div>
    <div class="advanced-row inline">
      <input type="checkbox" :checked="options.enableStreaming" @change="updateOption('enableStreaming', ($event.target as HTMLInputElement).checked)" />
      <label>开启流式传输（推荐）</label>
    </div>
  </div>

  <div class="advanced-group">
    <div class="advanced-group-title">传输控制</div>
    <div class="advanced-row">
      <label>并发传输数</label>
      <input type="number" :value="options.transfers" @input="updateOption('transfers', Number(($event.target as HTMLInputElement).value))" min="1" max="100" />
    </div>
    <div class="advanced-row">
      <label>带宽限制</label>
      <input type="text" :value="options.bwLimit" @input="updateOption('bwLimit', ($event.target as HTMLInputElement).value)" placeholder="如: 10M" />
    </div>
    <div class="advanced-row inline">
      <input type="checkbox" :checked="options.multiThreadStreams" @change="updateOption('multiThreadStreams', ($event.target as HTMLInputElement).checked)" />
      <label>多线程传输</label>
    </div>
    <div class="advanced-row">
      <label>最大传输</label>
      <input type="number" :value="options.maxTransfer" @input="updateOption('maxTransfer', Number(($event.target as HTMLInputElement).value))" min="0" placeholder="字节数, 0表示无限制" />
    </div>
    <div class="advanced-row">
      <label>最大时长</label>
      <input type="number" :value="options.maxDuration" @input="updateOption('maxDuration', Number(($event.target as HTMLInputElement).value))" min="0" placeholder="秒, 0表示无限制" />
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
