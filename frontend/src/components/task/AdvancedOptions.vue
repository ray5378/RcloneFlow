<script setup lang="ts">
interface Options {
  dryRun?: boolean
  interactive?: boolean
  checkFirst?: boolean
  serverSideAcrossConfigs?: boolean
  checkers?: number
  retries?: number
  backupDir?: string
  logFile?: string
  transfers?: number
  multiThreadStreams?: number
  multiThreadCutoff?: string
  bufferSize?: string
  timeout?: string
}

const props = defineProps<{
  modelValue: Options
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Options]
}>()

function update(key: keyof Options, value: any) {
  emit('update:modelValue', {
    ...props.modelValue,
    [key]: value,
  })
}
</script>

<template>
  <div class="advanced-options">
    <div class="field-item">
      <label class="inline-label">
        <input type="checkbox" :checked="modelValue?.dryRun" @change="update('dryRun', ($event.target as HTMLInputElement).checked)" />
        <span style="margin-left:8px">模拟运行 (dry-run)</span>
      </label>
    </div>

    <div class="field-item">
      <label class="inline-label">
        <input type="checkbox" :checked="modelValue?.interactive" @change="update('interactive', ($event.target as HTMLInputElement).checked)" />
        <span style="margin-left:8px">交互模式</span>
      </label>
    </div>

    <div class="field-item">
      <label class="inline-label">
        <input type="checkbox" :checked="modelValue?.checkFirst" @change="update('checkFirst', ($event.target as HTMLInputElement).checked)" />
        <span style="margin-left:8px">检查前先检查</span>
      </label>
    </div>

    <div class="field-item">
      <label class="inline-label">
        <input type="checkbox" :checked="modelValue?.serverSideAcrossConfigs" @change="update('serverSideAcrossConfigs', ($event.target as HTMLInputElement).checked)" />
        <span style="margin-left:8px">服务器端跨配置</span>
      </label>
    </div>

    <div class="advanced-grid">
      <div class="field-item">
        <label>传输线程数</label>
        <input type="number" :value="modelValue?.transfers" @input="update('transfers', Number(($event.target as HTMLInputElement).value))" min="1" max="100" />
      </div>

      <div class="field-item">
        <label>检查器数</label>
        <input type="number" :value="modelValue?.checkers" @input="update('checkers', Number(($event.target as HTMLInputElement).value))" min="1" max="100" />
      </div>

      <div class="field-item">
        <label>重试次数</label>
        <input type="number" :value="modelValue?.retries" @input="update('retries', Number(($event.target as HTMLInputElement).value))" min="0" />
      </div>

      <div class="field-item">
        <label>多线程数</label>
        <input type="number" :value="modelValue?.multiThreadStreams" @input="update('multiThreadStreams', Number(($event.target as HTMLInputElement).value))" min="1" max="32" />
      </div>

      <div class="field-item">
        <label>多线程阈值</label>
        <input type="text" :value="modelValue?.multiThreadCutoff" @input="update('multiThreadCutoff', ($event.target as HTMLInputElement).value)" placeholder="如: 1G" />
      </div>

      <div class="field-item">
        <label>缓冲区大小</label>
        <input type="text" :value="modelValue?.bufferSize" @input="update('bufferSize', ($event.target as HTMLInputElement).value)" placeholder="如: 16M" />
      </div>

      <div class="field-item">
        <label>超时时间</label>
        <input type="text" :value="modelValue?.timeout" @input="update('timeout', ($event.target as HTMLInputElement).value)" placeholder="如: 5m" />
      </div>
    </div>

    <div class="field-item">
      <label>备份目录</label>
      <input type="text" :value="modelValue?.backupDir" @input="update('backupDir', ($event.target as HTMLInputElement).value)" placeholder="remote:path" />
    </div>

    <div class="field-item">
      <label>日志文件</label>
      <input type="text" :value="modelValue?.logFile" @input="update('logFile', ($event.target as HTMLInputElement).value)" placeholder="/path/to/log" />
    </div>
  </div>
</template>

<style scoped>
.advanced-options {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.advanced-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
}
.field-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.field-item label {
  font-size: 12px;
  color: #888;
}
.field-item input[type="text"],
.field-item input[type="number"] {
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--surface);
  color: var(--text);
}
.field-item input:focus {
  outline: none;
  border-color: var(--accent);
}
</style>
