<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import * as api from '../api'
import type { Run, Task } from '../types'
import { t } from '../i18n'

const runs = ref<Run[]>([])
const tasks = ref<Task[]>([])
const refreshInterval = ref<number | null>(null)

onMounted(async () => {
  await loadData()
  refreshInterval.value = window.setInterval(loadData, 10000)
})

onUnmounted(() => {
  if (refreshInterval.value) clearInterval(refreshInterval.value)
})

async function loadData() {
  try {
    const [runData, taskData] = await Promise.all([api.listRuns(), api.listTasks()])
    runs.value = runData
    tasks.value = taskData
  } catch (e) {
    console.error(e)
  }
}

function getTaskName(taskId: number) {
  const task = tasks.value.find(t => t.id === taskId)
  return task?.name || `${t('run.taskFallback')} #${taskId}`
}

function formatTime(time: string) {
  return new Date(time).toLocaleString()
}

function getStatusClass(status: string) {
  switch (status) {
    case 'running': return 'running'
    case 'finished': return 'success'
    case 'failed': return 'error'
    default: return ''
  }
}

function getStatusText(status: string) {
  switch (status) {
    case 'running': return t('run.running')
    case 'finished': return t('run.finished')
    case 'failed': return t('run.failed')
    default: return status
  }
}

function getTriggerText(trigger: string) {
  if (trigger === 'schedule') return t('run.schedule')
  if (trigger === 'webhook') return 'Webhook'
  if (trigger === 'manual') return t('run.manual')
  return trigger
}

function formatSummary(summary?: Record<string, unknown>) {
  if (!summary) return ''
  const parts = []
  if (summary.transferredFiles) parts.push(`${t('run.files')}: ${summary.transferredFiles}`)
  if (summary.transferredSize) parts.push(`${t('run.size')}: ${summary.transferredSize}`)
  if (summary.elapsedTime) parts.push(`${t('run.elapsed')}: ${summary.elapsedTime}`)
  if (summary.speed) parts.push(`${t('run.speed')}: ${summary.speed}`)
  return parts.join(' | ')
}
</script>

<template>
  <div class="card">
    <div class="card-header">
      <div class="title">{{ t('run.title') }}</div>
      <div class="subtitle">{{ t('run.subtitle') }}</div>
    </div>
    <div class="list">
      <div v-for="run in runs" :key="run.id" class="item">
        <div class="name">
          <strong>{{ getTaskName(run.taskId) }}</strong>
          <div class="muted">{{ formatTime(run.createdAt) }} / {{ getTriggerText(run.trigger) }}</div>
          <div v-if="run.error" style="color: #ef5350; margin-top: 4px">{{ run.error }}</div>
          <div v-if="run.summary" class="muted" style="margin-top: 4px">{{ formatSummary(run.summary) }}</div>
        </div>
        <span :class="['badge', getStatusClass(run.status)]">{{ getStatusText(run.status) }}</span>
      </div>
      <div v-if="!runs.length" class="empty">{{ t('run.noData') }}</div>
    </div>
  </div>
</template>
