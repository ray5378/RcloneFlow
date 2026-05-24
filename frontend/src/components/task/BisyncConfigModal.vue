<script setup lang="ts">
import type { BisyncOptions } from './types'

const props = defineProps<{
  visible: boolean
  bisyncOptions: BisyncOptions
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'save'): void
  (e: 'update:bisyncOptions', value: BisyncOptions): void
}>()

function updateOption<K extends keyof BisyncOptions>(key: K, value: BisyncOptions[K]) {
  emit('update:bisyncOptions', {
    ...props.bisyncOptions,
    [key]: value,
  })
}
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content" style="max-width: 600px">
      <div class="modal-header">
        <h3>Bisync 双向同步设置</h3>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <div class="detail-item full-width">
          <label class="inline-label">
            <input
              :checked="bisyncOptions.resync"
              type="checkbox"
              @change="updateOption('resync', ($event.target as HTMLInputElement).checked)"
            />
            <span>重新同步 (Resync)</span>
          </label>
          <p class="hint">首次运行或从错误状态恢复时使用，会清理旧状态重新开始同步</p>
        </div>

        <div class="detail-item full-width">
          <label>比较方式 (Compare)</label>
          <select
            :value="bisyncOptions.compare"
            @change="updateOption('compare', ($event.target as HTMLSelectElement).value || undefined)"
          >
            <option value="">默认</option>
            <option value="size">仅检查大小</option>
            <option value="modtime">检查修改时间</option>
            <option value="checksum">检查校验和</option>
          </select>
        </div>

        <div class="detail-item full-width">
          <label>最大删除比例 (Max Delete)</label>
          <input
            type="text"
            :value="bisyncOptions.maxDelete"
            @input="updateOption('maxDelete', ($event.target as HTMLInputElement).value || undefined)"
            placeholder="例如: 25 (限制删除不超过 25%)"
          />
          <p class="hint">设置删除操作的安全限制，防止意外大量删除</p>
        </div>

        <div class="detail-item full-width">
          <label class="inline-label">
            <input
              :checked="bisyncOptions.checkAccess"
              type="checkbox"
              @change="updateOption('checkAccess', ($event.target as HTMLInputElement).checked)"
            />
            <span>检查访问权限 (Check Access)</span>
          </label>
        </div>

        <div class="detail-item full-width">
          <label>冲突解决策略 (Conflict Resolve)</label>
          <select
            :value="bisyncOptions.conflictResolve"
            @change="updateOption('conflictResolve', ($event.target as HTMLSelectElement).value || undefined)"
          >
            <option value="">默认</option>
            <option value="path1">优先源端</option>
            <option value="path2">优先目标端</option>
            <option value="newer">优先较新文件</option>
            <option value="older">优先较旧文件</option>
          </select>
        </div>

        <div class="detail-item full-width">
          <label>冲突处理 (Conflict Loser)</label>
          <select
            :value="bisyncOptions.conflictLoser"
            @change="updateOption('conflictLoser', ($event.target as HTMLSelectElement).value || undefined)"
          >
            <option value="">默认</option>
            <option value="backup">备份冲突文件</option>
            <option value="delete">删除冲突文件</option>
          </select>
        </div>

        <div class="detail-item full-width">
          <label>历史备份数量</label>
          <input
            type="number"
            :value="bisyncOptions.lstBackupCount || 5"
            min="1"
            max="50"
            @input="updateOption('lstBackupCount', parseInt(($event.target as HTMLInputElement).value) || 5)"
            placeholder="默认: 5"
          />
          <p class="hint">设置保留的 lst 文件历史备份数量，建议 3-10 个版本</p>
        </div>
      </div>
      <div class="modal-footer">
        <button class="primary" @click="emit('save')">保存</button>
        <button class="ghost" @click="emit('close')">取消</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.hint {
  margin-top: 8px;
  color: var(--muted, #94a3b8);
  font-size: 13px;
  line-height: 1.5;
}

.modal-content input,
.modal-content select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--surface);
  color: var(--text);
  box-sizing: border-box;
}

body.light .modal-content input,
body.light .modal-content select {
  background: #fff;
  border-color: #ddd;
  color: #333;
}

.modal-content label {
  display: block;
  margin-bottom: 6px;
  font-size: 13px;
  color: var(--muted);
}

.modal-content label.inline-label {
  display: flex !important;
  align-items: center;
  gap: 8px;
  margin: 0 0 6px 0;
}

.modal-content label.inline-label input[type="checkbox"] {
  width: 16px;
  height: 16px;
}
</style>
