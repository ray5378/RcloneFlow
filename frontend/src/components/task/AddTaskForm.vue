<script setup lang="ts">
import { computed } from 'vue'
import { ScheduleOptions, AdvancedOptions } from './'
import { PathItem } from '../path'

const props = defineProps<{
  commandMode: boolean
  commandText: string
  createForm: any
  remotes: string[]
  showSourcePathInput: boolean
  showTargetPathInput: boolean
  sourceBreadcrumbs: Array<{ name: string; path: string }>
  sourceCurrentPath: string
  sourcePathOptions: any[]
  targetBreadcrumbs: Array<{ name: string; path: string }>
  targetCurrentPath: string
  targetPathOptions: any[]
  showAdvancedOptions: boolean
  creatingState: string
  editingTask: any
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
  'source-arrow': [item: any]
  'source-click': [item: any]
  'target-arrow': [item: any]
  'target-click': [item: any]
  submit: []
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

const showAdvancedOptionsModel = computed({
  get: () => props.showAdvancedOptions,
  set: (value: boolean) => emit('update:showAdvancedOptions', value),
})
</script>

<template>
  <div class="card">
    <div class="card-header"><div class="title">添加任务</div></div>
    <div class="form-content">
      <div class="field-item">
        <label class="inline-label">
          <input type="checkbox" v-model="commandModeModel" />
          <span style="margin-left:8px">命令行模式（可粘贴 rclone 命令）</span>
        </label>
        <textarea v-if="commandModeModel" v-model="commandTextModel" class="cmd-textarea" rows="4" placeholder='例如: rclone copy FNOS:/HDD/media openlist:/影音媒体/天翼5050 --bwlimit "07:30,2M;17:40,2M;23:00,2M" --use-server-modtime --size-only --verbose --transfers 2'></textarea>
        <p v-if="commandModeModel" class="hint">保存时将自动解析命令，填充"模式/源/目标/选项"。任务名称仍需手动填写。</p>
      </div>
      <div class="field-item">
        <label>任务名称 <span style="color: #dc2626">*</span></label>
        <input v-model="createForm.name" type="text" placeholder="输入任务名称" />
      </div>
      <div class="field-item">
        <label>模式</label>
        <select v-model="createForm.mode">
          <option value="copy">复制 (copy)</option>
          <option value="sync">同步 (sync)</option>
          <option value="move">移动 (move)</option>
        </select>
      </div>
      <div class="field-item">
        <label>源存储 <span style="color: #dc2626">*</span></label>
        <select v-model="createForm.sourceRemote" @change="$emit('source-remote-change')">
          <option value="">选择源存储</option>
          <option v-for="r in remotes" :key="r" :value="r">{{ r }}</option>
        </select>
      </div>
      <div class="field-item">
        <label>源路径</label>
        <div class="path-selector">
          <div class="path-browse">
            <div class="pathbar">
              <template v-for="(crumb, i) in sourceBreadcrumbs" :key="crumb.path">
                <span v-if="i > 0" class="sep">/</span>
                <button
                  class="crumb"
                  :class="{ current: i === sourceBreadcrumbs.length - 1 }"
                  @click="crumb.path !== sourceCurrentPath && $emit('load-source-path', createForm.sourceRemote, crumb.path)"
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
              <div v-if="!sourcePathOptions.length" class="path-empty">空目录</div>
            </div>
          </div>
          <button type="button" class="ghost small" @click="showSourcePathInputModel = !showSourcePathInputModel">手动输入</button>
        </div>
        <input v-if="showSourcePathInputModel" v-model="createForm.sourcePath" type="text" placeholder="手动输入路径" style="margin-top: 8px" />
      </div>
      <div class="field-item">
        <label>目标存储 <span style="color: #dc2626">*</span></label>
        <select v-model="createForm.targetRemote" @change="$emit('target-remote-change')">
          <option value="">选择目标存储</option>
          <option v-for="r in remotes" :key="r" :value="r">{{ r }}</option>
        </select>
      </div>
      <div class="field-item">
        <label>目标路径</label>
        <div class="path-selector">
          <div class="path-browse">
            <div class="pathbar">
              <template v-for="(crumb, i) in targetBreadcrumbs" :key="crumb.path">
                <span v-if="i > 0" class="sep">/</span>
                <button
                  class="crumb"
                  :class="{ current: i === targetBreadcrumbs.length - 1 }"
                  @click="crumb.path !== targetCurrentPath && $emit('load-target-path', createForm.targetRemote, crumb.path)"
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
              <div v-if="!targetPathOptions.length" class="path-empty">空目录</div>
            </div>
          </div>
          <button type="button" class="ghost small" @click="showTargetPathInputModel = !showTargetPathInputModel">手动输入</button>
        </div>
        <input v-if="showTargetPathInputModel" v-model="createForm.targetPath" type="text" placeholder="手动输入路径" style="margin-top: 8px" />
      </div>

      <ScheduleOptions
        :model-value="createForm"
        @update:model-value="Object.assign(createForm, $event)"
      />

      <button type="button" class="ghost small" @click="showAdvancedOptionsModel = !showAdvancedOptionsModel">
        {{ showAdvancedOptionsModel ? '收起高级选项' : '+ 高级选项' }}
      </button>
      <div v-if="showAdvancedOptionsModel" class="advanced-section">
        <AdvancedOptions v-model="createForm.options" />
      </div>

      <div class="form-actions">
        <button
          class="primary"
          :class="{ 'btn-success': creatingState === 'done' }"
          :disabled="creatingState === 'loading'"
          @click="$emit('submit')"
        >
          <template v-if="creatingState === 'loading'">创建中...</template>
          <template v-else-if="creatingState === 'done'">完成（点击返回任务列表）</template>
          <template v-else-if="editingTask">保存修改</template>
          <template v-else>创建任务</template>
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
body.light .advanced-section { border-top-color: #ddd; }
.path-empty { padding: 20px; text-align: center; color: #666; font-size: 13px; }
.form-actions { margin-top: 20px; }
</style>
