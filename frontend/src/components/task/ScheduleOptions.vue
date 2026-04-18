<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  buildScheduleFormFieldsFromTemp,
  createEmptyScheduleTempState,
  parseScheduleFormToTemp,
  scheduleFieldOptions,
  toggleAllScheduleTempField,
  toggleScheduleTempField,
  weekLabels,
  type ScheduleField,
  type ScheduleFormLike as ScheduleForm,
} from './scheduleOptions'

const props = defineProps<{
  modelValue: ScheduleForm
}>()

const emit = defineEmits<{
  'update:modelValue': [value: ScheduleForm]
}>()

const temp = ref(createEmptyScheduleTempState())

watch(() => props.modelValue, (val) => {
  if (val) {
    temp.value = parseScheduleFormToTemp(val)
  }
}, { immediate: true })

function updateSchedule() {
  emit('update:modelValue', {
    ...props.modelValue,
    ...buildScheduleFormFieldsFromTemp(temp.value),
  })
}

function toggleEnable() {
  emit('update:modelValue', {
    ...props.modelValue,
    enableSchedule: !props.modelValue.enableSchedule,
  })
}

function toggleField(field: ScheduleField, value: string) {
  temp.value = toggleScheduleTempField(temp.value, field, value)
  updateSchedule()
}

function toggleAll(field: ScheduleField) {
  temp.value = toggleAllScheduleTempField(temp.value, field)
  updateSchedule()
}
</script>

<template>
  <div class="schedule-options">
    <div class="field-item">
      <label class="inline-label">
        <span>启用定时调度</span>
        <input type="checkbox" :checked="modelValue?.enableSchedule" @change="toggleEnable" />
      </label>
    </div>

    <div v-if="modelValue?.enableSchedule" class="schedule-fields">
      <div class="schedule-row">
        <label>月份</label>
        <div class="chip-select">
          <button
            v-for="m in scheduleFieldOptions.month"
            :key="m"
            :class="['chip-btn', temp.month.includes(m) && 'active']"
            @click="toggleField('month', m)"
          >{{ m }}月</button>
          <button class="chip-btn all-btn" @click="toggleAll('month')">全选</button>
        </div>
      </div>

      <div class="schedule-row">
        <label>日期</label>
        <div class="chip-select">
          <button
            v-for="d in scheduleFieldOptions.day"
            :key="d"
            :class="['chip-btn', temp.day.includes(d) && 'active']"
            @click="toggleField('day', d)"
          >{{ d }}</button>
          <button class="chip-btn all-btn" @click="toggleAll('day')">全选</button>
        </div>
      </div>

      <div class="schedule-row">
        <label>星期</label>
        <div class="chip-select">
          <button
            v-for="(w, i) in weekLabels"
            :key="i"
            :class="['chip-btn', temp.week.includes(String(i)) && 'active']"
            @click="toggleField('week', String(i))"
          >{{ w }}</button>
          <button class="chip-btn all-btn" @click="toggleAll('week')">全选</button>
        </div>
      </div>

      <div class="schedule-row">
        <label>时</label>
        <div class="chip-select">
          <button
            v-for="h in scheduleFieldOptions.hour"
            :key="h"
            :class="['chip-btn', temp.hour.includes(h) && 'active']"
            @click="toggleField('hour', h)"
          >{{ h }}时</button>
          <button class="chip-btn all-btn" @click="toggleAll('hour')">全选</button>
        </div>
      </div>

      <div class="schedule-row">
        <label>分</label>
        <div class="chip-select minute-chips">
          <button
            v-for="m in scheduleFieldOptions.minute"
            :key="m"
            :class="['chip-btn', temp.minute.includes(m) && 'active']"
            @click="toggleField('minute', m)"
          >{{ m }}</button>
          <button class="chip-btn all-btn" @click="toggleAll('minute')">全选</button>
        </div>
      </div>

      <div class="schedule-preview">
        <span class="hint">格式: {{ modelValue?.scheduleMinute || '*' }}|{{ modelValue?.scheduleHour || '*' }}|{{ modelValue?.scheduleDay || '*' }}|{{ modelValue?.scheduleMonth || '*' }}|{{ modelValue?.scheduleWeek || '*' }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.schedule-options {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.schedule-fields {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 12px;
  background: var(--surface);
  border-radius: 8px;
}
.field-item {
  display: flex;
  align-items: center;
}
.field-item .inline-label {
  display: inline-flex;
  flex-direction: row;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}
.field-item .inline-label input[type="checkbox"] {
  margin: 0;
}
.schedule-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.schedule-row label {
  font-size: 12px;
  color: #888;
}
.chip-select {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.chip-btn {
  padding: 4px 8px;
  font-size: 11px;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: var(--surface);
  color: var(--text);
  cursor: pointer;
  transition: all 0.2s;
}
.chip-btn:hover {
  border-color: var(--accent);
}
.chip-btn.active {
  background: var(--accent);
  border-color: var(--accent);
  color: white;
}
.mini-btn {
  padding: 2px 4px;
  font-size: 10px;
  min-width: 20px;
}
.minute-chips .chip-btn {
  padding: 2px 4px;
  font-size: 10px;
  min-width: 24px;
  text-align: center;
}
.schedule-preview {
  margin-top: 8px;
}
.hint {
  font-size: 12px;
  color: #888;
}
</style>
