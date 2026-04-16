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

const tempSchedule = ref({
  month: [] as string[],
  week: [] as string[],
  day: [] as string[],
  hour: [] as string[],
  minute: [] as string[],
})

watch(() => props.modelValue, (val) => {
  if (val) {
    tempSchedule.value = {
      month: val.scheduleMonth && val.scheduleMonth !== '*' ? val.scheduleMonth.split(',') : [],
      week: val.scheduleWeek && val.scheduleWeek !== '*' ? val.scheduleWeek.split(',') : [],
      day: val.scheduleDay && val.scheduleDay !== '*' ? val.scheduleDay.split(',') : [],
      hour: val.scheduleHour && val.scheduleHour !== '*' ? val.scheduleHour.split(',') : [],
      minute: val.scheduleMinute && val.scheduleMinute !== '*' ? val.scheduleMinute.split(',') : [],
    }
  }
}, { immediate: true })

function confirmField(field: 'month' | 'week' | 'day' | 'hour' | 'minute') {
  const val = tempSchedule.value[field].join(',') || '*'
  emit('update:modelValue', {
    ...props.modelValue,
    scheduleMonth: field === 'month' ? val : props.modelValue.scheduleMonth,
    scheduleWeek: field === 'week' ? val : props.modelValue.scheduleWeek,
    scheduleDay: field === 'day' ? val : props.modelValue.scheduleDay,
    scheduleHour: field === 'hour' ? val : props.modelValue.scheduleHour,
    scheduleMinute: field === 'minute' ? val : props.modelValue.scheduleMinute,
  })
}
</script>

<template>
  <div class="schedule-section">
    <div class="section-header">
      <label class="schedule-toggle">
        <input type="checkbox" :checked="modelValue?.enableSchedule" @change="emit('update:modelValue', { ...modelValue, enableSchedule: !modelValue.enableSchedule })" />
        <span>启用定时任务</span>
      </label>
    </div>

    <div v-if="modelValue?.enableSchedule" class="schedule-grid">
      <!-- 月 -->
      <div class="schedule-item">
        <label>月</label>
        <select v-model="tempSchedule.month" multiple size="6" @dblclick="confirmField('month')">
          <option value="*">每月</option>
          <option v-for="m in [1,2,3,4,5,6,7,8,9,10,11,12]" :key="m" :value="String(m)">{{ m }}月</option>
        </select>
        <button type="button" class="ghost small" @click="confirmField('month')">确定</button>
        <span class="selected-val">{{ modelValue.scheduleMonth === '*' ? '每月' : (modelValue.scheduleMonth || '每月') }}</span>
      </div>
      <!-- 周 -->
      <div class="schedule-item">
        <label>周</label>
        <select v-model="tempSchedule.week" multiple size="6" @dblclick="confirmField('week')">
          <option value="*">每日</option>
          <option value="">不设置</option>
          <option v-for="(w, idx) in ['周一','周二','周三','周四','周五','周六','周日']" :key="w" :value="String(idx+1)">{{ w }}</option>
        </select>
        <button type="button" class="ghost small" @click="confirmField('week')">确定</button>
        <span class="selected-val">{{ modelValue.scheduleWeek === '*' ? '每日' : (modelValue.scheduleWeek || '不设置') }}</span>
      </div>
      <!-- 日 -->
      <div class="schedule-item">
        <label>日</label>
        <select v-model="tempSchedule.day" multiple size="6" @dblclick="confirmField('day')">
          <option value="*">每日</option>
          <option value="">不设置</option>
          <option v-for="d in 31" :key="d" :value="String(d)">{{ d }}日</option>
        </select>
        <button type="button" class="ghost small" @click="confirmField('day')">确定</button>
        <span class="selected-val">{{ modelValue.scheduleDay === '*' ? '每日' : (modelValue.scheduleDay || '不设置') }}</span>
      </div>
      <!-- 时 -->
      <div class="schedule-item">
        <label>时</label>
        <select v-model="tempSchedule.hour" multiple size="6" @dblclick="confirmField('hour')">
          <option value="*">每时</option>
          <option v-for="h in 24" :key="h-1" :value="String(h-1).padStart(2,'0')">{{ String(h-1).padStart(2,'0') }}时</option>
        </select>
        <button type="button" class="ghost small" @click="confirmField('hour')">确定</button>
        <span class="selected-val">{{ modelValue.scheduleHour === '*' ? '每时' : (modelValue.scheduleHour || '00') + '时' }}</span>
      </div>
      <!-- 分 -->
      <div class="schedule-item">
        <label>分</label>
        <select v-model="tempSchedule.minute" multiple size="6" @dblclick="confirmField('minute')">
          <option value="*">每分</option>
          <option v-for="m in 60" :key="m-1" :value="String(m-1).padStart(2,'0')">{{ String(m-1).padStart(2,'0') }}分</option>
        </select>
        <button type="button" class="ghost small" @click="confirmField('minute')">确定</button>
        <span class="selected-val">{{ modelValue.scheduleMinute === '*' ? '每分' : (modelValue.scheduleMinute || '00') + '分' }}</span>
      </div>
    </div>
  </div>
</template>
