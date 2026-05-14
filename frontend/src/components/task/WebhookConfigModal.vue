<script setup lang="ts">
import { t } from '../../i18n'

defineProps<{
  visible: boolean
  triggerId: string
  matchText: string
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
  (e: 'update:matchText', value: string): void
  (e: 'update:postUrl', value: string): void
  (e: 'update:wecomUrl', value: string): void
  (e: 'update:notifyManual', value: boolean): void
  (e: 'update:notifySchedule', value: boolean): void
  (e: 'update:notifyWebhook', value: boolean): void
  (e: 'update:statusSuccess', value: boolean): void
  (e: 'update:statusFailed', value: boolean): void
}>()

function onTextInput(event: Event, key: 'triggerId' | 'matchText' | 'postUrl' | 'wecomUrl') {
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
        <h3>{{ t('webhookModal.title') }}</h3>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <div class="detail-item full-width">
          <label>{{ t('webhookModal.triggerId') }}</label>
          <input :value="triggerId" type="text" :placeholder="t('webhookModal.triggerPlaceholder')" @input="onTextInput($event, 'triggerId')" />
        </div>
        <div class="detail-item full-width">
          <label>{{ t('webhookModal.matchText') }}</label>
          <input :value="matchText" type="text" :placeholder="t('webhookModal.matchPlaceholder')" @input="onTextInput($event, 'matchText')" />
          <p class="hint">{{ t('webhookModal.matchHint') }}</p>
        </div>
        <div class="detail-item full-width">
          <label>{{ t('webhookModal.postUrl') }}</label>
          <input :value="postUrl" type="text" placeholder="https://example.com/hooks/endpoint" @input="onTextInput($event, 'postUrl')" />
          <p class="hint">{{ t('webhookModal.postHint') }}</p>
        </div>
        <div class="detail-item full-width">
          <label>{{ t('webhookModal.wecomUrl') }}</label>
          <input :value="wecomUrl" type="text" placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=..." @input="onTextInput($event, 'wecomUrl')" />
          <p class="hint">{{ t('webhookModal.wecomHint') }}</p>
        </div>
        <div class="detail-item">
          <label>{{ t('webhookModal.triggerSource') }}</label>
          <div class="trigger-row">
            <label class="trigger-opt"><input :checked="notifyManual" type="checkbox" @change="onCheckbox($event, 'notifyManual')" /><span>{{ t('webhookModal.manual') }}</span></label>
            <label class="trigger-opt"><input :checked="notifySchedule" type="checkbox" @change="onCheckbox($event, 'notifySchedule')" /><span>{{ t('webhookModal.schedule') }}</span></label>
            <label class="trigger-opt"><input :checked="notifyWebhook" type="checkbox" @change="onCheckbox($event, 'notifyWebhook')" /><span>{{ t('taskCard.webhook') }}</span></label>
          </div>
        </div>
        <div class="detail-item">
          <label>{{ t('webhookModal.statusFilter') }}</label>
          <div class="trigger-row">
            <label class="trigger-opt"><input :checked="statusSuccess" type="checkbox" @change="onCheckbox($event, 'statusSuccess')" /><span>{{ t('webhookModal.success') }}</span></label>
            <label class="trigger-opt"><input :checked="statusFailed" type="checkbox" @change="onCheckbox($event, 'statusFailed')" /><span>{{ t('webhookModal.failed') }}</span></label>
          </div>
          <p class="hint">{{ t('webhookModal.statusHint') }}</p>
        </div>
      </div>
      <div class="modal-footer">
        <button class="primary" @click="emit('save')">{{ t('common.save') }}</button>
        <button class="ghost" :disabled="!canTest" @click="emit('test')">{{ t('webhookModal.test') }}</button>
        <button class="ghost" @click="emit('close')">{{ t('common.cancel') }}</button>
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
