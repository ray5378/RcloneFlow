<script setup lang="ts">
defineProps<{
  options: any
  updateOption: (key: string, value: any) => void
}>()
</script>

<template>
  <div class="advanced-group">
    <div class="advanced-group-title">路径策略</div>
    <div class="advanced-row inline"><input type="checkbox" :checked="options.noTraverse" @change="updateOption('noTraverse', ($event.target as HTMLInputElement).checked)" /><label>不遍历</label></div>
    <div class="advanced-row inline"><input type="checkbox" :checked="options.noCheckDest" @change="updateOption('noCheckDest', ($event.target as HTMLInputElement).checked)" /><label>不检查目标</label></div>
    <div class="advanced-row">
      <label>比较目录</label>
      <input type="text" :value="options.compareDest" @input="updateOption('compareDest', ($event.target as HTMLInputElement).value)" placeholder="remote:path" />
    </div>
    <div class="advanced-row">
      <label>复制目录</label>
      <input type="text" :value="options.copyDest" @input="updateOption('copyDest', ($event.target as HTMLInputElement).value)" placeholder="remote:path" />
    </div>
  </div>
</template>

<style scoped>
.advanced-group { margin-bottom: 16px; padding-bottom: 16px; border-bottom: 1px solid #2a2a2a; }
body.light .advanced-group { border-bottom-color: #eee; }
.advanced-group-title { font-weight: 600; font-size: 13px; color: #64b5f6; margin-bottom: 12px; }
.advanced-row { margin-bottom: 12px; }
.advanced-row label { display: block; font-size: 12px; color: #888; margin-bottom: 4px; }
.advanced-row input[type="text"] { width: 100%; padding: 8px 12px; border: 1px solid #333; border-radius: 8px; background: #252525; color: #e0e0e0; font-size: 13px; box-sizing: border-box; font-family: inherit; }
body.light .advanced-row input[type="text"] { background: #fff; border-color: #ddd; color: #1a1a1a; }
.advanced-row input:focus { outline: none; border-color: #64b5f6; }
.advanced-row.inline { display: flex; align-items: center; gap: 8px; }
.advanced-row.inline label { margin-bottom: 0; display: inline; }
.advanced-row.inline input[type="checkbox"] { width: 16px; height: 16px; flex: 0 0 auto; }
</style>
