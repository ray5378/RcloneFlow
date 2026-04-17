<script setup lang="ts">
const props = defineProps<{
  visible: boolean
  run: any | null
  phaseText: string
  progressText: string
  debugOpen: boolean
  debugCheckText: string
  debugProgressLine: string
  debugProgressJson: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'toggle-debug'): void
  (e: 'open-log', run: any): void
}>()
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content" style="max-width:520px">
      <div class="modal-header">
        <h3>任务运行中</h3>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <p>该任务仍在传输中，运行详情（历史）仅展示最终信息。</p>
        <p>实时日志与进度请点击"传输日志"或查看任务卡片上的实时进度。</p>
        <div class="hint-box">
          <div class="detail-item"><label>任务：</label><span>{{ run?.taskName || `#${run?.taskId}` }}</span></div>
          <div class="detail-item"><label>阶段：</label><span>{{ phaseText || '-' }}</span></div>
          <div class="detail-item"><label>实时：</label><span>{{ progressText || '-' }}</span></div>
          <div class="detail-item full-width">
            <button class="ghost debug-toggle" @click="emit('toggle-debug')">
              {{ debugOpen ? '收起调试详情' : '展开调试详情' }}
            </button>
          </div>
          <template v-if="debugOpen">
            <div class="detail-item"><label>自检：</label><span>{{ debugCheckText || '-' }}</span></div>
            <div class="detail-item full-width"><label>日志原文：</label><code class="inline-logline">{{ debugProgressLine || '-' }}</code></div>
            <div class="detail-item full-width"><label>接口进度：</label><code class="inline-logline">{{ debugProgressJson || '-' }}</code></div>
          </template>
        </div>
      </div>
      <div class="modal-footer">
        <button class="primary" @click="emit('open-log', run)">打开传输日志</button>
        <button class="ghost" @click="emit('close')">我知道了</button>
      </div>
    </div>
  </div>
</template>
