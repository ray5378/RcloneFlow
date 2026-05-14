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
    <div class="advanced-group-title">{{ t('advancedTask.filterParams') }}</div>

    <div class="subgroup">
      <div class="subgroup-title">{{ t('advancedTask.ruleMatching') }}</div>
      <div class="advanced-row">
        <label>{{ t('advancedTask.exclude') }}</label>
        <textarea :value="options.exclude" @input="updateOption('exclude', ($event.target as HTMLTextAreaElement).value)" :placeholder="t('advancedTask.excludePlaceholder')" rows="3"></textarea>
      </div>
      <div class="advanced-row">
        <label>{{ t('advancedTask.include') }}</label>
        <textarea :value="options.include" @input="updateOption('include', ($event.target as HTMLTextAreaElement).value)" :placeholder="t('advancedTask.includePlaceholder')" rows="3"></textarea>
      </div>
      <div class="advanced-row">
        <label>{{ t('advancedTask.filter') }}</label>
        <textarea :value="options.filter" @input="updateOption('filter', ($event.target as HTMLTextAreaElement).value)" :placeholder="t('advancedTask.filterPlaceholder')" rows="3"></textarea>
        <div class="hint-row">{{ t('advancedTask.filterHint') }}</div>
      </div>
      <div class="advanced-row">
        <label>{{ t('advancedTask.includeFrom') }}</label>
        <textarea :value="options.includeFrom" @input="updateOption('includeFrom', ($event.target as HTMLTextAreaElement).value)" :placeholder="t('advancedTask.includeFromPlaceholder')" rows="2"></textarea>
        <div class="hint-row">{{ t('advancedTask.includeFromHint') }}</div>
      </div>
      <div class="advanced-row">
        <label>{{ t('advancedTask.excludeFrom') }}</label>
        <textarea :value="options.excludeFrom" @input="updateOption('excludeFrom', ($event.target as HTMLTextAreaElement).value)" :placeholder="t('advancedTask.excludeFromPlaceholder')" rows="2"></textarea>
        <div class="hint-row">{{ t('advancedTask.excludeFromHint') }}</div>
      </div>
      <div class="advanced-row">
        <label>{{ t('advancedTask.filterFrom') }}</label>
        <textarea :value="options.filterFrom" @input="updateOption('filterFrom', ($event.target as HTMLTextAreaElement).value)" :placeholder="t('advancedTask.filterFromPlaceholder')" rows="2"></textarea>
      </div>
    </div>

    <div class="subgroup">
      <div class="subgroup-title">{{ t('advancedTask.fileLists') }}</div>
      <div class="advanced-row">
        <label>{{ t('advancedTask.filesFrom') }}</label>
        <textarea :value="options.filesFrom" @input="updateOption('filesFrom', ($event.target as HTMLTextAreaElement).value)" :placeholder="t('advancedTask.filesFromPlaceholder')" rows="2"></textarea>
        <div class="hint-row">{{ t('advancedTask.filesFromHint') }}</div>
      </div>
      <div class="advanced-row">
        <label>{{ t('advancedTask.filesFromRaw') }}</label>
        <textarea :value="options.filesFromRaw" @input="updateOption('filesFromRaw', ($event.target as HTMLTextAreaElement).value)" :placeholder="t('advancedTask.filesFromRawPlaceholder')" rows="2"></textarea>
        <div class="hint-row">{{ t('advancedTask.filesFromRawHint') }}</div>
      </div>
      <div class="advanced-row">
        <label>{{ t('advancedTask.excludeIfPresent') }}</label>
        <textarea :value="options.excludeIfPresent" @input="updateOption('excludeIfPresent', ($event.target as HTMLTextAreaElement).value)" :placeholder="t('advancedTask.excludeIfPresentPlaceholder')" rows="2"></textarea>
        <div class="hint-row">{{ t('advancedTask.excludeIfPresentHint') }}</div>
      </div>
    </div>

    <div class="subgroup">
      <div class="subgroup-title">{{ t('advancedTask.rangeLimits') }}</div>
      <div class="advanced-row size-age-grid">
        <div>
          <label>{{ t('advancedTask.minSize') }}</label>
          <input type="text" :value="options.minSize" @input="updateOption('minSize', ($event.target as HTMLInputElement).value)" :placeholder="t('advancedTask.minSizePlaceholder')" />
        </div>
        <div>
          <label>{{ t('advancedTask.maxSize') }}</label>
          <input type="text" :value="options.maxSize" @input="updateOption('maxSize', ($event.target as HTMLInputElement).value)" :placeholder="t('advancedTask.maxSizePlaceholder')" />
        </div>
      </div>
      <div class="advanced-row size-age-grid">
        <div>
          <label>{{ t('advancedTask.minAge') }}</label>
          <input type="text" :value="options.minAge" @input="updateOption('minAge', ($event.target as HTMLInputElement).value)" :placeholder="t('advancedTask.minAgePlaceholder')" />
        </div>
        <div>
          <label>{{ t('advancedTask.maxAge') }}</label>
          <input type="text" :value="options.maxAge" @input="updateOption('maxAge', ($event.target as HTMLInputElement).value)" :placeholder="t('advancedTask.maxAgePlaceholder')" />
        </div>
      </div>
    </div>

    <div class="subgroup">
      <div class="subgroup-title">{{ t('advancedTask.behaviorSwitches') }}</div>
      <div class="advanced-row inline">
        <input type="checkbox" :checked="options.ignoreCase" @change="updateOption('ignoreCase', ($event.target as HTMLInputElement).checked)" />
        <label>{{ t('advancedTask.ignoreCase') }}</label>
      </div>
      <div class="advanced-row inline">
        <input type="checkbox" :checked="options.ignoreExisting" @change="updateOption('ignoreExisting', ($event.target as HTMLInputElement).checked)" />
        <label>{{ t('advancedTask.ignoreExisting') }}</label>
      </div>
      <div class="advanced-row inline">
        <input type="checkbox" :checked="options.deleteExcluded" @change="updateOption('deleteExcluded', ($event.target as HTMLInputElement).checked)" />
        <label>{{ t('advancedTask.deleteExcluded') }}</label>
      </div>
    </div>
  </div>
</template>

<style scoped>
.advanced-group { margin-bottom: 16px; padding-bottom: 16px; border-bottom: 1px solid #2a2a2a; }
body.light .advanced-group { border-bottom-color: #eee; }
.advanced-group-title { font-weight: 600; font-size: 13px; color: #64b5f6; margin-bottom: 12px; }
.subgroup { margin-top: 14px; padding: 12px; border: 1px solid #2f2f2f; border-radius: 10px; background: rgba(255,255,255,0.01); }
body.light .subgroup { border-color: #e5e7eb; background: #fafafa; }
.subgroup-title { font-size: 12px; font-weight: 600; color: #8ab4f8; margin-bottom: 10px; }
body.light .subgroup-title { color: #2563eb; }
.advanced-row { margin-bottom: 12px; }
.advanced-row label { display: block; font-size: 12px; color: #888; margin-bottom: 4px; }
.advanced-row textarea,
.advanced-row input[type="text"] { width: 100%; padding: 8px 12px; border: 1px solid #333; border-radius: 8px; background: #252525; color: #e0e0e0; font-size: 13px; box-sizing: border-box; font-family: inherit; }
.advanced-row textarea { resize: vertical; }
body.light .advanced-row textarea,
body.light .advanced-row input[type="text"] { background: #fff; border-color: #ddd; color: #1a1a1a; }
.advanced-row textarea:focus,
.advanced-row input[type="text"]:focus { outline: none; border-color: #64b5f6; }
.size-age-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
.advanced-row.inline { display: flex; align-items: center; gap: 8px; }
.advanced-row.inline label { margin-bottom: 0; display: inline; }
.advanced-row.inline input[type="checkbox"] { width: 16px; height: 16px; flex: 0 0 auto; }
.hint-row { font-size: 12px; color: #777; line-height: 1.5; margin-top: 6px; }
body.light .hint-row { color: #666; }
@media (max-width: 720px) {
  .size-age-grid { grid-template-columns: 1fr; }
}
</style>
