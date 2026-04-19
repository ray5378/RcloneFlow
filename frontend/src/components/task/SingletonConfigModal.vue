<script setup lang="ts">
defineProps<{
  visible: boolean
  singletonEnabled: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'save'): void
  (e: 'update:singletonEnabled', value: boolean): void
}>()

function onToggle(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:singletonEnabled', target.checked)
}
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content" style="max-width:560px">
      <div class="modal-header">
        <h3>单例模式</h3>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <div class="detail-item full-width">
          <label class="trigger-opt">
            <input :checked="singletonEnabled" type="checkbox" @change="onToggle" />
            <span>开启单例模式</span>
          </label>
          <p class="hint">开启后，该任务触发时会检测全局是否有其他传输任务在运行。有则放弃本次执行，不排队，不等待，不重试。</p>
        </div>
      </div>
      <div class="modal-footer">
        <button class="primary" @click="emit('save')">保存</button>
        <button class="ghost" @click="emit('close')">取消</button>
      </div>
    </div>
  </div>
</template>
