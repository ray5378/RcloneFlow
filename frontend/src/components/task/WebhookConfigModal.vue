<script setup lang="ts">
defineProps<{
  visible: boolean
  triggerId: string
  postUrl: string
  wecomUrl: string
  notifyManual: boolean
  notifySchedule: boolean
  notifyWebhook: boolean
  statusSuccess: boolean
  statusFailed: boolean
  canTest: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'save'): void
  (e: 'test'): void
  (e: 'update:triggerId', value: string): void
  (e: 'update:postUrl', value: string): void
  (e: 'update:wecomUrl', value: string): void
  (e: 'update:notifyManual', value: boolean): void
  (e: 'update:notifySchedule', value: boolean): void
  (e: 'update:notifyWebhook', value: boolean): void
  (e: 'update:statusSuccess', value: boolean): void
  (e: 'update:statusFailed', value: boolean): void
}>()

function onTextInput(event: Event, key: 'triggerId' | 'postUrl' | 'wecomUrl') {
  const target = event.target as HTMLInputElement
  emit(`update:${key}` as any, target.value)
}

function onCheckbox(event: Event, key: 'notifyManual' | 'notifySchedule' | 'notifyWebhook' | 'statusSuccess' | 'statusFailed') {
  const target = event.target as HTMLInputElement
  emit(`update:${key}` as any, target.checked)
}
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content" style="max-width:560px">
      <div class="modal-header">
        <h3>Webhook 通知</h3>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <div class="detail-item full-width">
          <label>Webhook 接收（触发ID）：</label>
          <input :value="triggerId" type="text" placeholder="留空=使用任务ID（示例：/webhook/<任务ID> 或 /webhook/<你的ID>）" @input="onTextInput($event, 'triggerId')" />
        </div>
        <div class="detail-item full-width">
          <label>对外 POST 地址：</label>
          <input :value="postUrl" type="text" placeholder="https://example.com/hooks/endpoint" @input="onTextInput($event, 'postUrl')" />
          <p class="hint">任务完成或失败后，将以 POST 通知该地址。</p>
        </div>
        <div class="detail-item full-width">
          <label>企业微信地址：</label>
          <input :value="wecomUrl" type="text" placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=..." @input="onTextInput($event, 'wecomUrl')" />
          <p class="hint">若填写，将同时向企业微信机器人发送 Markdown 通知。</p>
        </div>
        <div class="detail-item">
          <label>触发来源：</label>
          <div class="trigger-row">
            <label class="trigger-opt"><input :checked="notifyManual" type="checkbox" @change="onCheckbox($event, 'notifyManual')" /><span>手动</span></label>
            <label class="trigger-opt"><input :checked="notifySchedule" type="checkbox" @change="onCheckbox($event, 'notifySchedule')" /><span>定时</span></label>
            <label class="trigger-opt"><input :checked="notifyWebhook" type="checkbox" @change="onCheckbox($event, 'notifyWebhook')" /><span>Webhook</span></label>
          </div>
        </div>
        <div class="detail-item">
          <label>状态过滤：</label>
          <div class="trigger-row">
            <label class="trigger-opt"><input :checked="statusSuccess" type="checkbox" @change="onCheckbox($event, 'statusSuccess')" /><span>成功</span></label>
            <label class="trigger-opt"><input :checked="statusFailed" type="checkbox" @change="onCheckbox($event, 'statusFailed')" /><span>失败</span></label>
          </div>
          <p class="hint">仅当匹配状态时发送通知；默认两个都勾选。</p>
        </div>
      </div>
      <div class="modal-footer">
        <button class="primary" @click="emit('save')">保存</button>
        <button class="ghost" :disabled="!canTest" @click="emit('test')">发送测试</button>
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

.trigger-row {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.trigger-opt {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.trigger-opt input {
  width: 16px;
  height: 16px;
}
</style>
