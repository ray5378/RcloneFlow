<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import * as api from '../api'
import type { Run, Task } from '../types'

const runs = ref<Run[]>([])
const tasks = ref<Task[]>([])
const refreshInterval = ref<number | null>(null)

onMounted(async () => {
  await loadData()
  // Auto refresh every 10s for running status
  refreshInterval.value = window.setInterval(loadData, 10000)
})

onUnmounted(() => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value)
  }
})

async function loadData() {
  try {
    const [runData, taskData] = await Promise.all([
      api.listRuns(),
      api.listTasks(),
    ])
    runs.value = runData
    tasks.value = taskData
  } catch (e) {
    console.error(e)
  }
}

function getTaskName(taskId: number) {
  const task = tasks.value.find(t => t.id === taskId)
  return task?.name || `任务 #${taskId}`
}

function formatTime(time: string) {
  return new Date(time).toLocaleString()
}

function getStatusClass(status: string) {
  switch (status) {
    case 'running':
      return { color: '#2563eb', bg: '#dbeafe' }
    case 'finished':
      return { color: '#16a34a', bg: '#dcfce7' }
    case 'failed':
      return { color: '#dc2626', bg: '#fee2e2' }
    default:
      return { color: '#6b7280', bg: '#f3f4f6' }
  }
}

function formatSummary(summary?: Record<string, unknown>) {
  if (!summary) return ''
  const parts = []
  if (summary.transferredFiles) parts.push(`文件: ${summary.transferredFiles}`)
  if (summary.transferredSize) parts.push(`大小: ${summary.transferredSize}`)
  if (summary.elapsedTime) parts.push(`耗时: ${summary.elapsedTime}`)
  if (summary.speed) parts.push(`速度: ${summary.speed}`)
  return parts.join(' | ')
}
</script>

<template>
  <div class="card">
    <div class="topbar">
      <div>
        <div style="font-size: 18px; font-weight: 600">运行记录</div>
        <div class="muted">查看任务执行状态</div>
      </div>
      <div class="actions">
        <button class="ghost small" @click="loadData">刷新</button>
      </div>
    </div>
  </div>

  <div class="card">
    <div class="list">
      <div v-for="run in runs" :key="run.id" class="item">
        <div class="manage-row">
          <div>
            <strong>{{ getTaskName(run.taskId) }}</strong>
            <div class="muted">
              {{ formatTime(run.createdAt) }} / {{ run.trigger }}
            </div>
            <div v-if="run.error" style="color: #dc2626; margin-top: 4px">
              {{ run.error }}
            </div>
            <div v-if="run.summary" class="muted" style="margin-top: 4px">
              {{ formatSummary(run.summary) }}
            </div>
          </div>
          <span
            :style="{
              color: getStatusClass(run.status).color,
              background: getStatusClass(run.status).bg,
              padding: '4px 8px',
              borderRadius: '8px',
              fontSize: '12px',
              fontWeight: 500,
            }"
          >
            {{ run.status }}
          </span>
        </div>
      </div>
      <div v-if="!runs.length" class="empty">暂无运行记录</div>
    </div>
  </div>
</template>
