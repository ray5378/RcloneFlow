<script setup lang="ts">
interface Options {
  enableStreaming?: boolean
  exclude?: string
  include?: string
  filter?: string
  ignoreCase?: boolean
  ignoreExisting?: boolean
  deleteExcluded?: boolean
  checksum?: boolean
  sizeOnly?: boolean
  ignoreSize?: boolean
  ignoreTimes?: boolean
  update?: boolean
  modifyWindow?: string
  noTraverse?: boolean
  noCheckDest?: boolean
  compareDest?: string
  copyDest?: string
  transfers?: number
  bwLimit?: string
  multiThreadStreams?: boolean
  maxTransfer?: number
  maxDuration?: number
  dryRun?: boolean
  interactive?: boolean
  checkFirst?: boolean
  serverSideAcrossConfigs?: boolean
  checkers?: number
  retries?: number
  backupDir?: string
  logFile?: string
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
    <!-- 传输策略 -->
    <div class="advanced-group">
      <div class="advanced-group-title">传输策略</div>
      <div class="advanced-row inline">
        <label>开启流式传输（推荐）</label>
        <input type="checkbox" :checked="modelValue?.enableStreaming" @change="update('enableStreaming', ($event.target as HTMLInputElement).checked)" />
      </div>
    </div>

    <!-- 过滤参数 -->
    <div class="advanced-group">
      <div class="advanced-group-title">过滤参数</div>
      <div class="advanced-row">
        <label>排除 (exclude)</label>
        <textarea :value="modelValue?.exclude" @input="update('exclude', ($event.target as HTMLTextAreaElement).value)" placeholder="每行一个规则, 如: *.txt&#10;备份/**" rows="3"></textarea>
      </div>
      <div class="advanced-row">
        <label>包含 (include)</label>
        <textarea :value="modelValue?.include" @input="update('include', ($event.target as HTMLTextAreaElement).value)" placeholder="每行一个规则, 如: *.pdf&#10;文档/**" rows="3"></textarea>
      </div>
      <div class="advanced-row">
        <label>过滤规则 (filter)</label>
        <textarea :value="modelValue?.filter" @input="update('filter', ($event.target as HTMLTextAreaElement).value)" placeholder="每行一个规则, 如: - *.tmp&#10;+ *.bak" rows="3"></textarea>
      </div>
      <div class="advanced-row inline">
        <label>忽略大小写</label>
        <input type="checkbox" :checked="modelValue?.ignoreCase" @change="update('ignoreCase', ($event.target as HTMLInputElement).checked)" />
      </div>
      <div class="advanced-row inline">
        <label>忽略已存在的文件</label>
        <input type="checkbox" :checked="modelValue?.ignoreExisting" @change="update('ignoreExisting', ($event.target as HTMLInputElement).checked)" />
      </div>
      <div class="advanced-row inline">
        <label>删除被排除的文件</label>
        <input type="checkbox" :checked="modelValue?.deleteExcluded" @change="update('deleteExcluded', ($event.target as HTMLInputElement).checked)" />
      </div>
    </div>

    <!-- 比较策略 -->
    <div class="advanced-group">
      <div class="advanced-group-title">比较策略</div>
      <div class="advanced-row inline">
        <label>校验和比较</label>
        <input type="checkbox" :checked="modelValue?.checksum" @change="update('checksum', ($event.target as HTMLInputElement).checked)" />
      </div>
      <div class="advanced-row inline">
        <label>仅按大小</label>
        <input type="checkbox" :checked="modelValue?.sizeOnly" @change="update('sizeOnly', ($event.target as HTMLInputElement).checked)" />
      </div>
      <div class="advanced-row inline">
        <label>忽略大小</label>
        <input type="checkbox" :checked="modelValue?.ignoreSize" @change="update('ignoreSize', ($event.target as HTMLInputElement).checked)" />
      </div>
      <div class="advanced-row inline">
        <label>忽略时间</label>
        <input type="checkbox" :checked="modelValue?.ignoreTimes" @change="update('ignoreTimes', ($event.target as HTMLInputElement).checked)" />
      </div>
      <div class="advanced-row inline">
        <label>更新较新的</label>
        <input type="checkbox" :checked="modelValue?.update" @change="update('update', ($event.target as HTMLInputElement).checked)" />
      </div>
      <div class="advanced-row">
        <label>时间窗口</label>
        <input type="text" :value="modelValue?.modifyWindow" @input="update('modifyWindow', ($event.target as HTMLInputElement).value)" placeholder="如: 1h2s" />
      </div>
    </div>

    <!-- 路径策略 -->
    <div class="advanced-group">
      <div class="advanced-group-title">路径策略</div>
      <div class="advanced-row inline">
        <label>不遍历</label>
        <input type="checkbox" :checked="modelValue?.noTraverse" @change="update('noTraverse', ($event.target as HTMLInputElement).checked)" />
      </div>
      <div class="advanced-row inline">
        <label>不检查目标</label>
        <input type="checkbox" :checked="modelValue?.noCheckDest" @change="update('noCheckDest', ($event.target as HTMLInputElement).checked)" />
      </div>
      <div class="advanced-row">
        <label>比较目录</label>
        <input type="text" :value="modelValue?.compareDest" @input="update('compareDest', ($event.target as HTMLInputElement).value)" placeholder="remote:path" />
      </div>
      <div class="advanced-row">
        <label>复制目录</label>
        <input type="text" :value="modelValue?.copyDest" @input="update('copyDest', ($event.target as HTMLInputElement).value)" placeholder="remote:path" />
      </div>
    </div>

    <!-- 传输控制 -->
    <div class="advanced-group">
      <div class="advanced-group-title">传输控制</div>
      <div class="advanced-row">
        <label>并发传输数</label>
        <input type="number" :value="modelValue?.transfers" @input="update('transfers', Number(($event.target as HTMLInputElement).value))" min="1" max="100" />
      </div>
      <div class="advanced-row">
        <label>带宽限制</label>
        <input type="text" :value="modelValue?.bwLimit" @input="update('bwLimit', ($event.target as HTMLInputElement).value)" placeholder="如: 10M" />
      </div>
      <div class="advanced-row inline">
        <label>多线程传输</label>
        <input type="checkbox" :checked="modelValue?.multiThreadStreams" @change="update('multiThreadStreams', ($event.target as HTMLInputElement).checked)" />
      </div>
      <div class="advanced-row">
        <label>最大传输</label>
        <input type="number" :value="modelValue?.maxTransfer" @input="update('maxTransfer', Number(($event.target as HTMLInputElement).value))" min="0" placeholder="字节数, 0表示无限制" />
      </div>
      <div class="advanced-row">
        <label>最大时长</label>
        <input type="number" :value="modelValue?.maxDuration" @input="update('maxDuration', Number(($event.target as HTMLInputElement).value))" min="0" placeholder="秒, 0表示无限制" />
      </div>
    </div>

    <!-- 其他参数 -->
    <div class="advanced-group">
      <div class="advanced-group-title">其他参数</div>
      <div class="advanced-row inline">
        <label>模拟运行 (dry-run)</label>
        <input type="checkbox" :checked="modelValue?.dryRun" @change="update('dryRun', ($event.target as HTMLInputElement).checked)" />
      </div>
      <div class="advanced-row inline">
        <label>交互模式</label>
        <input type="checkbox" :checked="modelValue?.interactive" @change="update('interactive', ($event.target as HTMLInputElement).checked)" />
      </div>
      <div class="advanced-row inline">
        <label>检查前先检查</label>
        <input type="checkbox" :checked="modelValue?.checkFirst" @change="update('checkFirst', ($event.target as HTMLInputElement).checked)" />
      </div>
      <div class="advanced-row inline">
        <label>服务器端跨配置</label>
        <input type="checkbox" :checked="modelValue?.serverSideAcrossConfigs" @change="update('serverSideAcrossConfigs', ($event.target as HTMLInputElement).checked)" />
      </div>
      <div class="advanced-row">
        <label>检查器数</label>
        <input type="number" :value="modelValue?.checkers" @input="update('checkers', Number(($event.target as HTMLInputElement).value))" min="1" max="100" />
      </div>
      <div class="advanced-row">
        <label>重试次数</label>
        <input type="number" :value="modelValue?.retries" @input="update('retries', Number(($event.target as HTMLInputElement).value))" min="0" />
      </div>
      <div class="advanced-row">
        <label>备份目录</label>
        <input type="text" :value="modelValue?.backupDir" @input="update('backupDir', ($event.target as HTMLInputElement).value)" placeholder="remote:path" />
      </div>
      <div class="advanced-row">
        <label>日志文件</label>
        <input type="text" :value="modelValue?.logFile" @input="update('logFile', ($event.target as HTMLInputElement).value)" placeholder="/path/to/log" />
      </div>
    </div>
  </div>
</template>
