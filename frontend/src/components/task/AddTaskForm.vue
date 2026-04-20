<script setup lang="ts">
import { computed } from 'vue'
import ScheduleOptions from './ScheduleOptions.vue'
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

const optionsModel = computed({
  get: () => {
    const raw = props.createForm?.options
    return raw && typeof raw === 'object' ? raw : { enableStreaming: true }
  },
  set: (value: any) => {
    props.createForm.options = value && typeof value === 'object' ? value : { enableStreaming: true }
  },
})

function updateOption(key: string, value: any) {
  optionsModel.value = {
    ...optionsModel.value,
    [key]: value,
  }
}
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

      <div class="advanced-section">
        <div class="advanced-title">高级选项</div>

        <div class="advanced-options">
          <div class="advanced-group">
            <div class="advanced-group-title">传输策略</div>
            <div class="advanced-row inline">
              <input type="checkbox" :checked="optionsModel.enableStreaming" @change="updateOption('enableStreaming', ($event.target as HTMLInputElement).checked)" />
              <label>开启流式传输（推荐）</label>
            </div>
          </div>

          <div class="advanced-group">
            <div class="advanced-group-title">过滤参数</div>
            <div class="advanced-row">
              <label>排除 (exclude)</label>
              <textarea :value="optionsModel.exclude" @input="updateOption('exclude', ($event.target as HTMLTextAreaElement).value)" placeholder="每行一个规则, 如: *.txt&#10;备份/**" rows="3"></textarea>
            </div>
            <div class="advanced-row">
              <label>包含 (include)</label>
              <textarea :value="optionsModel.include" @input="updateOption('include', ($event.target as HTMLTextAreaElement).value)" placeholder="每行一个规则, 如: *.pdf&#10;文档/**" rows="3"></textarea>
            </div>
            <div class="advanced-row">
              <label>过滤规则 (filter)</label>
              <textarea :value="optionsModel.filter" @input="updateOption('filter', ($event.target as HTMLTextAreaElement).value)" placeholder="每行一个规则, 如: - *.tmp&#10;+ *.bak" rows="3"></textarea>
            </div>
            <div class="advanced-row inline">
              <input type="checkbox" :checked="optionsModel.ignoreCase" @change="updateOption('ignoreCase', ($event.target as HTMLInputElement).checked)" />
              <label>忽略大小写</label>
            </div>
            <div class="advanced-row inline">
              <input type="checkbox" :checked="optionsModel.ignoreExisting" @change="updateOption('ignoreExisting', ($event.target as HTMLInputElement).checked)" />
              <label>忽略已存在的文件</label>
            </div>
            <div class="advanced-row inline">
              <input type="checkbox" :checked="optionsModel.deleteExcluded" @change="updateOption('deleteExcluded', ($event.target as HTMLInputElement).checked)" />
              <label>删除被排除的文件</label>
            </div>
          </div>

          <div class="advanced-group">
            <div class="advanced-group-title">比较策略</div>
            <div class="advanced-row inline">
              <input type="checkbox" :checked="optionsModel.checksum" @change="updateOption('checksum', ($event.target as HTMLInputElement).checked)" />
              <label>校验和比较</label>
            </div>
            <div class="advanced-row inline">
              <input type="checkbox" :checked="optionsModel.sizeOnly" @change="updateOption('sizeOnly', ($event.target as HTMLInputElement).checked)" />
              <label>仅按大小</label>
            </div>
            <div class="advanced-row inline">
              <input type="checkbox" :checked="optionsModel.ignoreSize" @change="updateOption('ignoreSize', ($event.target as HTMLInputElement).checked)" />
              <label>忽略大小</label>
            </div>
            <div class="advanced-row inline">
              <input type="checkbox" :checked="optionsModel.ignoreTimes" @change="updateOption('ignoreTimes', ($event.target as HTMLInputElement).checked)" />
              <label>忽略时间</label>
            </div>
            <div class="advanced-row inline">
              <input type="checkbox" :checked="optionsModel.update" @change="updateOption('update', ($event.target as HTMLInputElement).checked)" />
              <label>更新较新的</label>
            </div>
            <div class="advanced-row">
              <label>时间窗口</label>
              <input type="text" :value="optionsModel.modifyWindow" @input="updateOption('modifyWindow', ($event.target as HTMLInputElement).value)" placeholder="如: 1h2s" />
            </div>
          </div>

          <div class="advanced-group">
            <div class="advanced-group-title">路径策略</div>
            <div class="advanced-row inline">
              <input type="checkbox" :checked="optionsModel.noTraverse" @change="updateOption('noTraverse', ($event.target as HTMLInputElement).checked)" />
              <label>不遍历</label>
            </div>
            <div class="advanced-row inline">
              <input type="checkbox" :checked="optionsModel.noCheckDest" @change="updateOption('noCheckDest', ($event.target as HTMLInputElement).checked)" />
              <label>不检查目标</label>
            </div>
            <div class="advanced-row">
              <label>比较目录</label>
              <input type="text" :value="optionsModel.compareDest" @input="updateOption('compareDest', ($event.target as HTMLInputElement).value)" placeholder="remote:path" />
            </div>
            <div class="advanced-row">
              <label>复制目录</label>
              <input type="text" :value="optionsModel.copyDest" @input="updateOption('copyDest', ($event.target as HTMLInputElement).value)" placeholder="remote:path" />
            </div>
          </div>

          <div class="advanced-group">
            <div class="advanced-group-title">传输控制</div>
            <div class="advanced-row">
              <label>并发传输数</label>
              <input type="number" :value="optionsModel.transfers" @input="updateOption('transfers', Number(($event.target as HTMLInputElement).value))" min="1" max="100" />
            </div>
            <div class="advanced-row">
              <label>带宽限制</label>
              <input type="text" :value="optionsModel.bwLimit" @input="updateOption('bwLimit', ($event.target as HTMLInputElement).value)" placeholder="如: 10M" />
            </div>
            <div class="advanced-row inline">
              <input type="checkbox" :checked="optionsModel.multiThreadStreams" @change="updateOption('multiThreadStreams', ($event.target as HTMLInputElement).checked)" />
              <label>多线程传输</label>
            </div>
            <div class="advanced-row">
              <label>最大传输</label>
              <input type="number" :value="optionsModel.maxTransfer" @input="updateOption('maxTransfer', Number(($event.target as HTMLInputElement).value))" min="0" placeholder="字节数, 0表示无限制" />
            </div>
            <div class="advanced-row">
              <label>最大时长</label>
              <input type="number" :value="optionsModel.maxDuration" @input="updateOption('maxDuration', Number(($event.target as HTMLInputElement).value))" min="0" placeholder="秒, 0表示无限制" />
            </div>
          </div>

          <div class="advanced-group">
            <div class="advanced-group-title">其他参数</div>
            <div class="advanced-row inline">
              <input type="checkbox" :checked="optionsModel.dryRun" @change="updateOption('dryRun', ($event.target as HTMLInputElement).checked)" />
              <label>模拟运行 (dry-run)</label>
            </div>
            <div class="advanced-row inline">
              <input type="checkbox" :checked="optionsModel.interactive" @change="updateOption('interactive', ($event.target as HTMLInputElement).checked)" />
              <label>交互模式</label>
            </div>
            <div class="advanced-row inline">
              <input type="checkbox" :checked="optionsModel.checkFirst" @change="updateOption('checkFirst', ($event.target as HTMLInputElement).checked)" />
              <label>检查前先检查</label>
            </div>
            <div class="advanced-row inline">
              <input type="checkbox" :checked="optionsModel.serverSideAcrossConfigs" @change="updateOption('serverSideAcrossConfigs', ($event.target as HTMLInputElement).checked)" />
              <label>服务器端跨配置</label>
            </div>
            <div class="advanced-row">
              <label>检查器数</label>
              <input type="number" :value="optionsModel.checkers" @input="updateOption('checkers', Number(($event.target as HTMLInputElement).value))" min="1" max="100" />
            </div>
            <div class="advanced-row">
              <label>重试次数</label>
              <input type="number" :value="optionsModel.retries" @input="updateOption('retries', Number(($event.target as HTMLInputElement).value))" min="0" />
            </div>
            <div class="advanced-row">
              <label>备份目录</label>
              <input type="text" :value="optionsModel.backupDir" @input="updateOption('backupDir', ($event.target as HTMLInputElement).value)" placeholder="remote:path" />
            </div>
            <div class="advanced-row">
              <label>日志文件</label>
              <input type="text" :value="optionsModel.logFile" @input="updateOption('logFile', ($event.target as HTMLInputElement).value)" placeholder="/path/to/log" />
            </div>
          </div>
        </div>
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
.advanced-title { font-size: 13px; color: var(--muted); margin-bottom: 12px; font-weight: 600; }
body.light .advanced-section { border-top-color: #ddd; }
.advanced-options { display: flex; flex-direction: column; gap: 0; }
.advanced-group { margin-bottom: 16px; padding-bottom: 16px; border-bottom: 1px solid #2a2a2a; }
body.light .advanced-group { border-bottom-color: #eee; }
.advanced-group:last-child { border-bottom: none; }
.advanced-group-title { font-weight: 600; font-size: 13px; color: #64b5f6; margin-bottom: 12px; }
.advanced-row { margin-bottom: 12px; }
.advanced-row label { display: block; font-size: 12px; color: #888; margin-bottom: 4px; }
.advanced-row input[type="text"],
.advanced-row input[type="number"],
.advanced-row textarea,
.advanced-row select {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #333;
  border-radius: 8px;
  background: #252525;
  color: #e0e0e0;
  font-size: 13px;
  box-sizing: border-box;
  font-family: inherit;
}
body.light .advanced-row input[type="text"],
body.light .advanced-row input[type="number"],
body.light .advanced-row textarea,
body.light .advanced-row select {
  background: #fff;
  border-color: #ddd;
  color: #1a1a1a;
}
.advanced-row input:focus,
.advanced-row textarea:focus,
.advanced-row select:focus {
  outline: none;
  border-color: #64b5f6;
}
.advanced-row textarea { resize: vertical; }
.advanced-row.inline { display: flex; align-items: center; gap: 8px; }
.advanced-row.inline label { margin-bottom: 0; display: inline; }
.advanced-row.inline input[type="checkbox"] { width: 16px; height: 16px; flex: 0 0 auto; }
.path-empty { padding: 20px; text-align: center; color: #666; font-size: 13px; }
.form-actions { margin-top: 20px; }
</style>
