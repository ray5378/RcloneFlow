<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as api from '../api'
import type { Schedule, Task } from '../types'
import { t } from '../i18n'

const schedules = ref<Schedule[]>([])
const tasks = ref<Task[]>([])
const scheduleMenu = ref<string | number>('')
const scheduleForm = ref({ taskId: 0, spec: '5m' })

onMounted(async () => { await loadData() })

async function loadData() {
  try {
    const [scheduleData, taskData] = await Promise.all([api.listSchedules(), api.listTasks()])
    schedules.value = scheduleData
    tasks.value = taskData
  } catch (e) {
    console.error(e)
  }
}

function getTaskName(taskId: number) {
  const task = tasks.value.find(t => t.id === taskId)
  return task?.name || `${t('schedule.taskFallback')} #${taskId}`
}

async function createSchedule() {
  if (!scheduleForm.value.taskId) {
    alert(t('schedule.pleaseSelectTask'))
    return
  }
  try {
    await api.createSchedule({ taskId: scheduleForm.value.taskId, spec: '@every ' + scheduleForm.value.spec, enabled: true })
    scheduleForm.value = { taskId: 0, spec: '5m' }
    await loadData()
  } catch (e) {
    alert((e as Error).message)
  }
}

async function deleteSchedule(id: number) {
  if (!confirm(t('schedule.deleteConfirm'))) return
  try {
    await api.deleteSchedule(id)
    await loadData()
  } catch (e) {
    alert((e as Error).message)
  }
}

function getSpecLabel(spec: string) {
  const map: Record<string, string> = {
    '5m': t('schedule.every5m'),
    '10m': t('schedule.every10m'),
    '30m': t('schedule.every30m'),
    '1h': t('schedule.every1h'),
    '6h': t('schedule.every6h'),
    '12h': t('schedule.every12h'),
    '24h': t('schedule.every24h'),
  }
  return map[spec] || spec
}
</script>

<template>
  <div class="card">
    <div class="card-header">
      <div class="title">{{ t('schedule.title') }}</div>
      <div class="subtitle">{{ t('schedule.subtitle') }}</div>
    </div>
    <div class="card-content">
      <div class="field-grid">
        <div class="field-item">
          <label>{{ t('schedule.selectTask') }}</label>
          <select v-model="scheduleForm.taskId">
            <option value="0">{{ t('schedule.selectTask') }}</option>
            <option v-for="tItem in tasks" :key="tItem.id" :value="tItem.id">{{ tItem.name }} ({{ tItem.mode }})</option>
          </select>
        </div>
        <div class="field-item">
          <label>{{ t('schedule.interval') }}</label>
          <select v-model="scheduleForm.spec">
            <option value="5m">{{ getSpecLabel('5m') }}</option>
            <option value="10m">{{ getSpecLabel('10m') }}</option>
            <option value="30m">{{ getSpecLabel('30m') }}</option>
            <option value="1h">{{ getSpecLabel('1h') }}</option>
            <option value="6h">{{ getSpecLabel('6h') }}</option>
            <option value="12h">{{ getSpecLabel('12h') }}</option>
            <option value="24h">{{ getSpecLabel('24h') }}</option>
          </select>
        </div>
      </div>
      <div style="margin-top: 16px"><button class="primary" @click="createSchedule">{{ t('schedule.create') }}</button></div>
    </div>
  </div>

  <div class="card">
    <div class="card-header"><div class="title">{{ t('schedule.listTitle') }}</div></div>
    <div class="list">
      <div v-for="s in schedules" :key="s.id" class="item">
        <div class="name">
          <strong>{{ getTaskName(s.taskId) }}</strong>
          <div class="muted">{{ s.spec.replace('@every ', '') }} / {{ s.enabled ? t('common.enabled') : t('common.disabled') }}</div>
        </div>
        <div class="actions" @click.stop>
          <div class="menu-area">
            <button class="menu-btn" @click="scheduleMenu = scheduleMenu === s.id ? '' : s.id">⋮</button>
            <div v-if="scheduleMenu === s.id" class="menu-pop">
              <button class="danger" @click="deleteSchedule(s.id); scheduleMenu = ''">{{ t('common.delete') }}</button>
            </div>
          </div>
        </div>
      </div>
      <div v-if="!schedules.length" class="empty">{{ t('schedule.noData') }}</div>
    </div>
  </div>
</template>
