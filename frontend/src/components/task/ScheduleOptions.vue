<script setup lang="ts">
import { ref, watch } from 'vue'

interface ScheduleForm {
  enableSchedule: boolean
  scheduleMinute: string
  scheduleHour: string
  scheduleDay: string
  scheduleMonth: string
  scheduleWeek: string
}

const props = defineProps<{
  modelValue: ScheduleForm
}>()

const emit = defineEmits<{
  'update:modelValue': [value: ScheduleForm]
}>()

const temp = ref({
  minute: [] as string[],
  hour: [] as string[],
  day: [] as string[],
  month: [] as string[],
  week: [] as string[],
})

watch(() => props.modelValue, (val) => {
  if (val) {
    temp.value = {
      minute: val.scheduleMinute && val.scheduleMinute !== '*' ? val.scheduleMinute.split(',') : [],
      hour: val.scheduleHour && val.scheduleHour !== '*' ? val.scheduleHour.split(',') : [],
      day: val.scheduleDay && val.scheduleDay !== '*' ? val.scheduleDay.split(',') : [],
      month: val.scheduleMonth && val.scheduleMonth !== '*' ? val.scheduleMonth.split(',') : [],
      week: val.scheduleWeek && val.scheduleWeek !== '*' ? val.scheduleWeek.split(',') : [],
    }
  }
}, { immediate: true })

function updateSchedule(field: string, val: string[]) {
  const parts = {
    minute: temp.value.minute.join(',') || '*',
    hour: temp.value.hour.join(',') || '*',
    day: temp.value.day.join(',') || '*',
    month: temp.value.month.join(',') || '*',
    week: temp.value.week.join(',') || '*',
  }
  parts[field as keyof typeof parts] = val.join(',') || '*'
  
  emit('update:modelValue', {
    ...props.modelValue,
    scheduleMinute: parts.minute,
    scheduleHour: parts.hour,
    scheduleDay: parts.day,
    scheduleMonth: parts.month,
    scheduleWeek: parts.week,
  })
}

function toggleEnable() {
  emit('update:modelValue', {
    ...props.modelValue,
    enableSchedule: !props.modelValue.enableSchedule,
  })
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
            v-for="m in ['1','2','3','4','5','6','7','8','9','10','11','12']" 
            :key="m"
            :class="['chip-btn', temp.month.includes(m) && 'active']"
            @click="temp.month.includes(m) ? temp.month = temp.month.filter(x=>x!==m) : temp.month.push(m); updateSchedule('month', temp.month)"
          >{{ m }}月</button>
        </div>
      </div>

      <div class="schedule-row">
        <label>日期</label>
        <div class="chip-select">
          <button 
            v-for="d in ['1','2','3','4','5','6','7','8','9','10','11','12','13','14','15','16','17','18','19','20','21','22','23','24','25','26','27','28','29','30','31']" 
            :key="d"
            :class="['chip-btn', temp.day.includes(d) && 'active']"
            @click="temp.day.includes(d) ? temp.day = temp.day.filter(x=>x!==d) : temp.day.push(d); updateSchedule('day', temp.day)"
          >{{ d }}</button>
        </div>
      </div>

      <div class="schedule-row">
        <label>星期</label>
        <div class="chip-select">
          <button 
            v-for="(w, i) in ['周日','周一','周二','周三','周四','周五','周六']" 
            :key="i"
            :class="['chip-btn', temp.week.includes(String(i)) && 'active']"
            @click="temp.week.includes(String(i)) ? temp.week = temp.week.filter(x=>x!==String(i)) : temp.week.push(String(i)); updateSchedule('week', temp.week)"
          >{{ w }}</button>
        </div>
      </div>

      <div class="schedule-row">
        <label>时</label>
        <div class="chip-select">
          <button 
            v-for="h in 24" 
            :key="h-1"
            :class="['chip-btn', temp.hour.includes(String(h-1).padStart(2,'0')) && 'active']"
            @click="temp.hour.includes(String(h-1).padStart(2,'0')) ? temp.hour = temp.hour.filter(x=>x!==String(h-1).padStart(2,'0')) : temp.hour.push(String(h-1).padStart(2,'0')); updateSchedule('hour', temp.hour)"
          >{{ String(h-1).padStart(2,'0') }}时</button>
        </div>
      </div>

      <div class="schedule-row">
        <label>分</label>
        <div class="chip-select minute-chips">
          <button 
            v-for="m in 60" 
            :key="m-1"
            :class="['chip-btn', temp.minute.includes(String(m-1).padStart(2,'0')) && 'active']"
            @click="temp.minute.includes(String(m-1).padStart(2,'0')) ? temp.minute = temp.minute.filter(x=>x!==String(m-1).padStart(2,'0')) : temp.minute.push(String(m-1).padStart(2,'0')); updateSchedule('minute', temp.minute)"
          >{{ String(m-1).padStart(2,'0') }}</button>
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
</style>
