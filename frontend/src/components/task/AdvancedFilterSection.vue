<script setup lang="ts">
defineProps<{
  options: any
  updateOption: (key: string, value: any) => void
}>()
</script>

<template>
  <div class="advanced-group">
    <div class="advanced-group-title">过滤参数</div>
    <div class="advanced-row">
      <label>排除 (exclude)</label>
      <textarea :value="options.exclude" @input="updateOption('exclude', ($event.target as HTMLTextAreaElement).value)" placeholder="每行一个规则, 如: *.txt&#10;备份/**" rows="3"></textarea>
    </div>
    <div class="advanced-row">
      <label>包含 (include)</label>
      <textarea :value="options.include" @input="updateOption('include', ($event.target as HTMLTextAreaElement).value)" placeholder="每行一个规则, 如: *.pdf&#10;文档/**" rows="3"></textarea>
    </div>
    <div class="advanced-row">
      <label>过滤规则 (filter)</label>
      <textarea :value="options.filter" @input="updateOption('filter', ($event.target as HTMLTextAreaElement).value)" placeholder="每行一个规则, 如: - *.tmp&#10;+ *.bak" rows="3"></textarea>
    </div>
    <div class="advanced-row inline">
      <input type="checkbox" :checked="options.ignoreCase" @change="updateOption('ignoreCase', ($event.target as HTMLInputElement).checked)" />
      <label>忽略大小写</label>
    </div>
    <div class="advanced-row inline">
      <input type="checkbox" :checked="options.ignoreExisting" @change="updateOption('ignoreExisting', ($event.target as HTMLInputElement).checked)" />
      <label>忽略已存在的文件</label>
    </div>
    <div class="advanced-row inline">
      <input type="checkbox" :checked="options.deleteExcluded" @change="updateOption('deleteExcluded', ($event.target as HTMLInputElement).checked)" />
      <label>删除被排除的文件</label>
    </div>
  </div>
</template>

<style scoped>
.advanced-group { margin-bottom: 16px; padding-bottom: 16px; border-bottom: 1px solid #2a2a2a; }
body.light .advanced-group { border-bottom-color: #eee; }
.advanced-group-title { font-weight: 600; font-size: 13px; color: #64b5f6; margin-bottom: 12px; }
.advanced-row { margin-bottom: 12px; }
.advanced-row label { display: block; font-size: 12px; color: #888; margin-bottom: 4px; }
.advanced-row textarea { width: 100%; padding: 8px 12px; border: 1px solid #333; border-radius: 8px; background: #252525; color: #e0e0e0; font-size: 13px; box-sizing: border-box; font-family: inherit; resize: vertical; }
body.light .advanced-row textarea { background: #fff; border-color: #ddd; color: #1a1a1a; }
.advanced-row textarea:focus { outline: none; border-color: #64b5f6; }
.advanced-row.inline { display: flex; align-items: center; gap: 8px; }
.advanced-row.inline label { margin-bottom: 0; display: inline; }
.advanced-row.inline input[type="checkbox"] { width: 16px; height: 16px; flex: 0 0 auto; }
</style>
