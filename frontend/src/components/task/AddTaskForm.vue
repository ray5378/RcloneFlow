<script setup lang="ts">
import { computed } from 'vue'
import type { CreateForm, PathBreadcrumb, PathBrowseItem, TaskFormOptions, TaskFormOptionValue, UpdateTaskOption } from './types'
import type { Task } from '../../types'
import AdvancedTransferSection from './AdvancedTransferSection.vue'
import AdvancedFilterSection from './AdvancedFilterSection.vue'
import AdvancedCompareSection from './AdvancedCompareSection.vue'
import AdvancedPathSection from './AdvancedPathSection.vue'
import AdvancedOtherSection from './AdvancedOtherSection.vue'
import { PathItem } from '../path'
import { t } from '../../i18n'

const props = defineProps<{
  commandMode: boolean
  commandText: string
  createForm: CreateForm
  remotes: string[]
  showSourcePathInput: boolean
  showTargetPathInput: boolean
  sourceBreadcrumbs: PathBreadcrumb[]
  sourceCurrentPath: string
  sourcePathOptions: PathBrowseItem[]
  targetBreadcrumbs: PathBreadcrumb[]
  targetCurrentPath: string
  targetPathOptions: PathBrowseItem[]
  showAdvancedOptions: boolean
  creatingState: string
  editingTask: Task | null
}>()

const emit = defineEmits<{
  'update:commandMode': [value: boolean]
  'update:commandText': [value: string]
  'update:showSourcePathInput': [value: boolean]
  'update:showTargetPathInput': [value: boolean]
  'update:showAdvancedOptions': [value: boolean]
  'source-remote-change': []
  'target-remote-change': []
  'load-source-path': [remote: string, path: string]
  'load-target-path': [remote: string, path: string]
  'source-arrow': [item: PathBrowseItem]
  'source-click': [item: PathBrowseItem]
  'target-arrow': [item: PathBrowseItem]
  'target-click': [item: PathBrowseItem]
  'submit': []
}>()

const commandModeModel = computed({
  get: () => props.commandMode,
  set: (value: boolean) => emit('update:commandMode', value),
})

const commandTextModel = computed({
  get: () => props.commandText,
  set: (value: string) => emit('update:commandText', value),
})

const showSourcePathInputModel = computed({
  get: () => props.showSourcePathInput,
  set: (value: boolean) => emit('update:showSourcePathInput', value),
})

const showTargetPathInputModel = computed({
  get: () => props.showTargetPathInput,
  set: (value: boolean) => emit('update:showTargetPathInput', value),
})

const commandPlaceholder = computed(() => 'rclone copy source:dir target:dir --progress')

const optionsModel = computed<TaskFormOptions>({
  get: () => {
    const raw = props.createForm?.options
    return raw && typeof raw === 'object' ? raw : { enableStreaming: true }
  },
  set: (value: TaskFormOptions) => {
    props.createForm.options = value && typeof value === 'object' ? value : { enableStreaming: true }
  },
})

const updateOption: UpdateTaskOption = (key, value) => {
  optionsModel.value = {
    ...optionsModel.value,
    [key]: value as TaskFormOptionValue,
  }
}
</script>

<template>
  <div class="card">
    <div class="card-header"><div class="title">{{ t('addTask.title') }}</div></div>
    <div class="form-content">
      <div class="field-item">
        <label class="inline-label">
          <input type="checkbox" v-model="commandModeModel" />
          <span style="margin-left:8px">{{ t('addTask.commandMode') }}</span>
        </label>
        <textarea v-if="commandModeModel" v-model="commandTextModel" class="cmd-textarea" rows="4" :placeholder="commandPlaceholder"></textarea>
        <p v-if="commandModeModel" class="hint">{{ t('addTask.commandHint') }}</p>
      </div>
      <div class="field-item">
        <label>{{ t('addTask.taskName') }} <span style="color: #dc2626">*</span></label>
        <input v-model="createForm.name" type="text" :placeholder="t('addTask.taskNamePlaceholder')" />
      </div>
      <div class="field-item">
        <label>{{ t('addTask.mode') }}</label>
        <select v-model="createForm.mode">
          <option value="copy">{{ t('addTask.copy') }} (copy)</option>
          <option value="sync">{{ t('addTask.sync') }} (sync)</option>
          <option value="move">{{ t('addTask.move') }} (move)</option>
        </select>
      </div>
      <div class="field-item">
        <label>{{ t('addTask.sourceStorage') }} <span style="color: #dc2626">*</span></label>
        <select v-model="createForm.sourceRemote" @change="$emit('source-remote-change')">
          <option value="">{{ t('addTask.selectSourceStorage') }}</option>
          <option v-for="r in remotes" :key="r" :value="r">{{ r }}</option>
        </select>
      </div>
      <div class="field-item">
        <label>{{ t('addTask.sourcePath') }}</label>
        <div class="path-selector">
          <div class="path-browse">
            <div class="pathbar">
              <template v-for="(crumb, i) in sourceBreadcrumbs" :key="crumb.path">
                <span v-if="i > 0" class="sep">/</span>
                <button
                  class="crumb"
                  :class="{ current: i === sourceBreadcrumbs.length - 1 }"
                  @click="crumb.path !== sourceCurrentPath && $emit('source-breadcrumb-click', crumb.path)"
                >
                  {{ crumb.name }}
                </button>
              </template>
            </div>
            <div class="path-list">
              <PathItem
                v-for="item in sourcePathOptions"
                :key="item.Path"
                :item="item"
                @enter="$emit('source-arrow', item)"
                @click="$emit('source-click', item)"
              />
              <div v-if="!sourcePathOptions.length" class="path-empty">{{ t('addTask.emptyDir') }}</div>
            </div>
          </div>
          <button type="button" class="ghost small" @click="showSourcePathInputModel = !showSourcePathInputModel">{{ t('addTask.manualInput') }}</button>
        </div>
        <input v-if="showSourcePathInputModel" v-model="createForm.sourcePath" type="text" :placeholder="t('addTask.manualPathPlaceholder')" style="margin-top: 8px" />
      </div>
      <div class="field-item">
        <label>{{ t('addTask.targetStorage') }} <span style="color: #dc2626">*</span></label>
        <select v-model="createForm.targetRemote" @change="$emit('target-remote-change')">
          <option value="">{{ t('addTask.selectTargetStorage') }}</option>
          <option v-for="r in remotes" :key="r" :value="r">{{ r }}</option>
        </select>
      </div>
      <div class="field-item">
        <label>{{ t('addTask.targetPath') }}</label>
        <div class="path-selector">
          <div class="path-browse">
            <div class="pathbar">
              <template v-for="(crumb, i) in targetBreadcrumbs" :key="crumb.path">
                <span v-if="i > 0" class="sep">/</span>
                <button
                  class="crumb"
                  :class="{ current: i === targetBreadcrumbs.length - 1 }"
                  @click="crumb.path !== targetCurrentPath && $emit('target-breadcrumb-click', crumb.path)"
                >
                  {{ crumb.name }}
                </button>
              </template>
            </div>
            <div class="path-list">
              <PathItem
                v-for="item in targetPathOptions"
                :key="item.Path"
                :item="item"
                @enter="$emit('target-arrow', item)"
                @click="$emit('target-click', item)"
              />
              <div v-if="!targetPathOptions.length" class="path-empty">{{ t('addTask.emptyDir') }}</div>
            </div>
          </div>
          <button type="button" class="ghost small" @click="showTargetPathInputModel = !showTargetPathInputModel">{{ t('addTask.manualInput') }}</button>
        </div>
        <input v-if="showTargetPathInputModel" v-model="createForm.targetPath" type="text" :placeholder="t('addTask.manualPathPlaceholder')" style="margin-top: 8px" />
      </div>

      <div class="advanced-section">
        <div class="advanced-title">{{ t('addTask.advancedOptions') }}</div>
        <div class="advanced-options">
          <AdvancedTransferSection :options="optionsModel" :update-option="updateOption" />
          <AdvancedFilterSection :options="optionsModel" :update-option="updateOption" />
          <AdvancedCompareSection :options="optionsModel" :update-option="updateOption" />
          <AdvancedPathSection :options="optionsModel" :update-option="updateOption" />
          <AdvancedOtherSection :options="optionsModel" :update-option="updateOption" />
        </div>
      </div>

      <div class="form-actions">
        <button class="primary" :class="{ 'btn-success': creatingState === 'done' }" :disabled="creatingState === 'loading'" @click="$emit('submit')">
          <template v-if="creatingState === 'loading'">{{ t('addTask.creating') }}</template>
          <template v-else-if="creatingState === 'done'">{{ t('addTask.done') }}</template>
          <template v-else-if="editingTask">{{ t('addTask.saveEdit') }}</template>
          <template v-else>{{ t('addTask.createTask') }}</template>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.form-content { padding: 20px; }
.form-content .field-item { margin-bottom: 16px; }
.form-content label { display: block; margin-bottom: 6px; font-size: 13px; color: var(--muted); }
.form-content input,
.form-content select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--surface);
  color: var(--text);
  box-sizing: border-box;
}
.cmd-textarea { width:100%; min-height:120px; padding:12px 14px; border-radius:10px; border:1px solid var(--border); background:var(--surface); color:var(--text); font-size:14px; box-sizing:border-box; resize:vertical; }
body.light .cmd-textarea { background:var(--surface); border-color:var(--border); color:var(--text); }
.form-content label.inline-label { display:flex !important; align-items:center; gap:8px; margin:0 0 6px 0; }
.form-content label.inline-label input[type="checkbox"] { width:16px; height:16px; }
body.light .form-content input,
body.light .form-content select { background:#fff; border-color:#ddd; color:#333; }
.path-selector { display:flex; gap:8px; align-items:flex-start; }
.path-browse { flex:1; border:1px solid var(--border); border-radius:8px; background:var(--surface); overflow:hidden; }
body.light .path-browse { border-color:var(--border); background:var(--surface); }
.pathbar { display:flex; flex-wrap:wrap; gap:0; padding:8px 10px; border-bottom:1px solid var(--border); background:color-mix(in srgb, var(--surface) 92%, var(--text) 8%); }
.sep { color: var(--muted); margin: 0 4px; }
.crumb { background: transparent; border: none; color: var(--accent); cursor: pointer; padding: 0; }
.crumb.current { color: var(--text); font-weight: 600; cursor: default; }
.path-list { max-height: 200px; overflow-y: auto; padding: 8px; }
.advanced-section { margin-top: 16px; padding-top: 16px; border-top: 1px solid #333; }
.advanced-title { font-size: 13px; color: var(--muted); margin-bottom: 12px; font-weight: 600; }
body.light .advanced-section { border-top-color: #ddd; }
.advanced-options { display: flex; flex-direction: column; gap: 0; }
.path-empty { padding: 20px; text-align: center; color: #666; font-size: 13px; }
.form-actions { margin-top: 20px; }
</style>
