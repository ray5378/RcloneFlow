<script setup lang="ts">
import { t } from '../../i18n'

defineProps<{
  options: any
  updateOption: (key: string, value: any) => void
}>()
</script>

<template>
  <div class="advanced-group">
    <div class="advanced-group-title">{{ t('advancedTask.otherOptions') }}</div>
    <div class="advanced-row inline"><input type="checkbox" :checked="options.dryRun" @change="updateOption('dryRun', ($event.target as HTMLInputElement).checked)" /><label>{{ t('advancedTask.dryRun') }}</label></div>
    <div class="advanced-row inline"><input type="checkbox" :checked="options.interactive" @change="updateOption('interactive', ($event.target as HTMLInputElement).checked)" /><label>{{ t('advancedTask.interactive') }}</label></div>
    <div class="advanced-row inline"><input type="checkbox" :checked="options.checkFirst" @change="updateOption('checkFirst', ($event.target as HTMLInputElement).checked)" /><label>{{ t('advancedTask.checkFirst') }}</label></div>
    <div class="advanced-row inline"><input type="checkbox" :checked="options.serverSideAcrossConfigs" @change="updateOption('serverSideAcrossConfigs', ($event.target as HTMLInputElement).checked)" /><label>{{ t('advancedTask.serverSideAcrossConfigs') }}</label></div>
    <div class="advanced-row inline"><input type="checkbox" :checked="options.useJsonLog" @change="updateOption('useJsonLog', ($event.target as HTMLInputElement).checked)" /><label>{{ t('advancedTask.useJsonLog') }}</label></div>
    <div class="advanced-row hint-row">{{ t('advancedTask.useJsonLogHint') }}</div>
    <div class="advanced-row inline"><input type="checkbox" :checked="options.openlistCasCompatible" @change="updateOption('openlistCasCompatible', ($event.target as HTMLInputElement).checked)" /><label>{{ t('advancedTask.openlistCasCompatible') }}</label></div>
    <div class="advanced-row hint-row">{{ t('advancedTask.openlistCasCompatibleHint') }}</div>
    <div class="advanced-row">
      <label>{{ t('advancedTask.checkersCount') }}</label>
      <input type="number" :value="options.checkers" @input="updateOption('checkers', Number(($event.target as HTMLInputElement).value))" min="1" max="100" />
    </div>
    <div class="advanced-row">
      <label>{{ t('advancedTask.retries') }}</label>
      <input type="number" :value="options.retries" @input="updateOption('retries', Number(($event.target as HTMLInputElement).value))" min="0" />
    </div>
    <div class="advanced-row">
      <label>{{ t('advancedTask.backupDir') }}</label>
      <input type="text" :value="options.backupDir" @input="updateOption('backupDir', ($event.target as HTMLInputElement).value)" :placeholder="t('advancedTask.remotePathPlaceholder')" />
    </div>
    <div class="advanced-row">
      <label>{{ t('advancedTask.logFile') }}</label>
      <input type="text" :value="options.logFile" @input="updateOption('logFile', ($event.target as HTMLInputElement).value)" :placeholder="t('advancedTask.logFilePlaceholder')" />
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
.hint-row { font-size: 12px; color: #777; line-height: 1.5; margin-top: -6px; }
body.light .hint-row { color: #666; }
.advanced-row input[type="text"], .advanced-row input[type="number"] { width: 100%; padding: 8px 12px; border: 1px solid #333; border-radius: 8px; background: #252525; color: #e0e0e0; font-size: 13px; box-sizing: border-box; font-family: inherit; }
body.light .advanced-row input[type="text"], body.light .advanced-row input[type="number"] { background: #fff; border-color: #ddd; color: #1a1a1a; }
.advanced-row input:focus { outline: none; border-color: #64b5f6; }
.advanced-row.inline { display: flex; align-items: center; gap: 8px; }
.advanced-row.inline label { margin-bottom: 0; display: inline; }
.advanced-row.inline input[type="checkbox"] { width: 16px; height: 16px; flex: 0 0 auto; }
</style>
