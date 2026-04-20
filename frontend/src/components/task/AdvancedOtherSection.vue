<script setup lang="ts">
defineProps<{
  options: any
  updateOption: (key: string, value: any) => void
}>()
</script>

<template>
  <div class="advanced-group">
    <div class="advanced-group-title">其他参数</div>
    <div class="advanced-row inline"><input type="checkbox" :checked="options.dryRun" @change="updateOption('dryRun', ($event.target as HTMLInputElement).checked)" /><label>模拟运行 (dry-run)</label></div>
    <div class="advanced-row inline"><input type="checkbox" :checked="options.interactive" @change="updateOption('interactive', ($event.target as HTMLInputElement).checked)" /><label>交互模式</label></div>
    <div class="advanced-row inline"><input type="checkbox" :checked="options.checkFirst" @change="updateOption('checkFirst', ($event.target as HTMLInputElement).checked)" /><label>检查前先检查</label></div>
    <div class="advanced-row inline"><input type="checkbox" :checked="options.serverSideAcrossConfigs" @change="updateOption('serverSideAcrossConfigs', ($event.target as HTMLInputElement).checked)" /><label>服务器端跨配置</label></div>
    <div class="advanced-row">
      <label>检查器数</label>
      <input type="number" :value="options.checkers" @input="updateOption('checkers', Number(($event.target as HTMLInputElement).value))" min="1" max="100" />
    </div>
    <div class="advanced-row">
      <label>重试次数</label>
      <input type="number" :value="options.retries" @input="updateOption('retries', Number(($event.target as HTMLInputElement).value))" min="0" />
    </div>
    <div class="advanced-row">
      <label>备份目录</label>
      <input type="text" :value="options.backupDir" @input="updateOption('backupDir', ($event.target as HTMLInputElement).value)" placeholder="remote:path" />
    </div>
    <div class="advanced-row">
      <label>日志文件</label>
      <input type="text" :value="options.logFile" @input="updateOption('logFile', ($event.target as HTMLInputElement).value)" placeholder="/path/to/log" />
    </div>
  </div>
</template>

<style scoped>
.advanced-group { margin-bottom: 16px; padding-bottom: 16px; border-bottom: 1px solid #2a2a2a; }
body.light .advanced-group { border-bottom-color: #eee; }
.advanced-group:last-child { border-bottom: none; }
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
