<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { t } from '../../i18n'
import {
  buildScheduleFormFieldsFromTemp,
  createEmptyScheduleTempState,
  parseScheduleFormToTemp,
  scheduleFieldOptions,
  toggleAllScheduleTempField,
  toggleScheduleTempField,
  weekLabels,
  type ScheduleField,
  type ScheduleFormLike,
} from './scheduleOptions'

const props = defineProps<{
  visible: boolean
  title?: string
  modelValue: ScheduleFormLike
  saving?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: ScheduleFormLike]
  save: [value: ScheduleFormLike]
  close: []
}>()

const draft = ref<ScheduleFormLike>({
  enableSchedule: false,
  scheduleMinute: '00',
  scheduleHour: '00',
  scheduleDay: '*',
  scheduleMonth: '*',
  scheduleWeek: '*',
})
const temp = ref(createEmptyScheduleTempState())

function cloneForm(value?: Partial<ScheduleFormLike> | null): ScheduleFormLike {
  return {
    enableSchedule: !!value?.enableSchedule,
    scheduleMinute: value?.scheduleMinute || '00',
    scheduleHour: value?.scheduleHour || '00',
    scheduleDay: value?.scheduleDay || '*',
    scheduleMonth: value?.scheduleMonth || '*',
    scheduleWeek: value?.scheduleWeek || '*',
  }
}

watch(() => [props.visible, props.modelValue] as const, () => {
  draft.value = cloneForm(props.modelValue)
  temp.value = parseScheduleFormToTemp(props.modelValue)
}, { immediate: true, deep: true })

function updateDraftFields() {
  draft.value = {
    ...draft.value,
    ...buildScheduleFormFieldsFromTemp(temp.value),
  }
}

function setEnabled(enabled: boolean) {
  draft.value = {
    ...draft.value,
    enableSchedule: enabled,
  }
}

function toggleField(field: ScheduleField, value: string) {
  temp.value = toggleScheduleTempField(temp.value, field, value)
  updateDraftFields()
}

function toggleAll(field: ScheduleField) {
  temp.value = toggleAllScheduleTempField(temp.value, field)
  updateDraftFields()
}

const previewText = computed(() => {
  return `${draft.value.scheduleMinute || '*'}|${draft.value.scheduleHour || '*'}|${draft.value.scheduleDay || '*'}|${draft.value.scheduleMonth || '*'}|${draft.value.scheduleWeek || '*'}`
})

function save() {
  const next = cloneForm(draft.value)
  emit('update:modelValue', next)
  emit('save', next)
}
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content schedule-modal">
      <div class="modal-header">
        <h3>{{ title || t('schedule.configTitle') }}</h3>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <div class="toggle-row">
          <button class="toggle-btn enable-btn" :class="{ active: draft.enableSchedule }" @click="setEnabled(true)">{{ t('schedule.enableNow') }}</button>
          <button class="toggle-btn disable-btn" :class="{ active: !draft.enableSchedule }" @click="setEnabled(false)">{{ t('schedule.disableNow') }}</button>
        </div>

        <div class="status-line">
          <span class="status-label">{{ t('schedule.currentStatus') }}</span>
          <span :class="['status-pill', draft.enableSchedule ? 'enabled' : 'disabled']">{{ draft.enableSchedule ? t('taskCard.enabled') : t('taskCard.disabled') }}</span>
        </div>

        <div v-if="draft.enableSchedule" class="schedule-fields">
          <div class="schedule-row">
            <label>{{ t('schedule.month') }}</label>
            <div class="chip-select">
              <button
                v-for="m in scheduleFieldOptions.month"
                :key="m"
                :class="['chip-btn', temp.month.includes(m) && 'active']"
                @click="toggleField('month', m)"
              >{{ m }}{{ t('schedule.monthSuffix') }}</button>
              <button class="chip-btn all-btn" @click="toggleAll('month')">{{ t('schedule.selectAll') }}</button>
            </div>
          </div>

          <div class="schedule-row">
            <label>{{ t('schedule.day') }}</label>
            <div class="chip-select">
              <button
                v-for="d in scheduleFieldOptions.day"
                :key="d"
                :class="['chip-btn', temp.day.includes(d) && 'active']"
                @click="toggleField('day', d)"
              >{{ d }}</button>
              <button class="chip-btn all-btn" @click="toggleAll('day')">{{ t('schedule.selectAll') }}</button>
            </div>
          </div>

          <div class="schedule-row">
            <label>{{ t('schedule.week') }}</label>
            <div class="chip-select">
              <button
                v-for="(w, i) in weekLabels"
                :key="i"
                :class="['chip-btn', temp.week.includes(String(i)) && 'active']"
                @click="toggleField('week', String(i))"
              >{{ w }}</button>
              <button class="chip-btn all-btn" @click="toggleAll('week')">{{ t('schedule.selectAll') }}</button>
            </div>
          </div>

          <div class="schedule-row">
            <label>{{ t('schedule.hour') }}</label>
            <div class="chip-select">
              <button
                v-for="h in scheduleFieldOptions.hour"
                :key="h"
                :class="['chip-btn', temp.hour.includes(h) && 'active']"
                @click="toggleField('hour', h)"
              >{{ h }}{{ t('schedule.hourSuffix') }}</button>
              <button class="chip-btn all-btn" @click="toggleAll('hour')">{{ t('schedule.selectAll') }}</button>
            </div>
          </div>

          <div class="schedule-row">
            <label>{{ t('schedule.minute') }}</label>
            <div class="chip-select minute-chips">
              <button
                v-for="m in scheduleFieldOptions.minute"
                :key="m"
                :class="['chip-btn', temp.minute.includes(m) && 'active']"
                @click="toggleField('minute', m)"
              >{{ m }}</button>
              <button class="chip-btn all-btn" @click="toggleAll('minute')">{{ t('schedule.selectAll') }}</button>
            </div>
          </div>
        </div>

        <div class="schedule-preview">
          <span class="hint">{{ t('schedule.format') }}: {{ previewText }}</span>
        </div>

        <div class="modal-actions">
          <button class="ghost" @click="emit('close')">{{ t('common.cancel') }}</button>
          <button class="primary" :disabled="saving" @click="save">{{ saving ? t('common.saving') : t('common.save') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.schedule-modal { width: min(960px, 92vw); max-width: 960px; }
.toggle-row { display:flex; gap:12px; margin-bottom:16px; }
.toggle-btn { flex:1; padding:12px 16px; border-radius:10px; border:1px solid #475569; background:#111827; color:#e5e7eb; font-weight:700; }
.toggle-btn.active.enable-btn { background:#166534; border-color:#166534; color:#fff; }
.toggle-btn.active.disable-btn { background:#b91c1c; border-color:#b91c1c; color:#fff; }
.status-line { display:flex; align-items:center; gap:10px; margin-bottom:16px; }
.status-label { color:#94a3b8; font-size:13px; }
.status-pill { padding:4px 10px; border-radius:999px; font-size:12px; font-weight:700; }
.status-pill.enabled { background:#22c55e33; color:#22c55e; }
.status-pill.disabled { background:#7c2d1233; color:#fca5a5; }
.schedule-fields { display:flex; flex-direction:column; gap:12px; padding:12px; background:var(--surface); border-radius:8px; }
.schedule-row { display:flex; flex-direction:column; gap:6px; }
.schedule-row label { font-size:12px; color:#888; }
.chip-select { display:flex; flex-wrap:wrap; gap:4px; }
.chip-btn { padding:4px 8px; font-size:11px; border:1px solid var(--border); border-radius:4px; background:var(--surface); color:var(--text); cursor:pointer; transition:all .2s; }
.chip-btn:hover { border-color:var(--accent); }
.chip-btn.active { background:var(--accent); border-color:var(--accent); color:#fff; }
.minute-chips .chip-btn { padding:2px 4px; font-size:10px; min-width:24px; text-align:center; }
.schedule-preview { margin-top:14px; }
.hint { font-size:12px; color:#888; }
.modal-actions { display:flex; justify-content:flex-end; gap:10px; margin-top:20px; }
body.light .toggle-btn { background:#fff; color:#222; border-color:#cbd5e1; }
body.light .schedule-fields { background:#f8fafc; }
</style>
