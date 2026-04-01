<script setup lang="ts">
import { ref } from 'vue'
import BrowserView from './views/BrowserView.vue'
import TaskView from './views/TaskView.vue'
import ScheduleView from './views/ScheduleView.vue'
import RunView from './views/RunView.vue'

const currentPage = ref('browser')
const version = ref('加载中...')

const pages = {
  browser: ['文件管理', '浏览存储和目录。'],
  tasks: ['任务管理', '创建复制 / 同步 / 移动任务。'],
  schedules: ['定时任务', '配置多个周期任务。'],
  runs: ['运行记录', '查看任务执行状态。'],
}

function switchPage(page: string) {
  currentPage.value = page
}
</script>

<template>
  <div class="app">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="brand">
        RcloneFlow
        <small>{{ version }}</small>
      </div>
      <nav class="nav">
        <button
          v-for="(info, key) in pages"
          :key="key"
          :class="{ active: currentPage === key }"
          @click="switchPage(key)"
        >
          {{ info[0] }}
        </button>
      </nav>
    </aside>

    <!-- Main Content -->
    <main class="main">
      <BrowserView v-if="currentPage === 'browser'" :version="version" />
      <TaskView v-if="currentPage === 'tasks'" />
      <ScheduleView v-if="currentPage === 'schedules'" />
      <RunView v-if="currentPage === 'runs'" />
    </main>
  </div>
</template>
