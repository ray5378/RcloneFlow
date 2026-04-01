<script setup lang="ts">
import { ref, onMounted } from 'vue'
import BrowserView from './views/BrowserView.vue'
import TaskView from './views/TaskView.vue'
import RunView from './views/RunView.vue'
import * as api from './api'

const currentPage = ref('browser')
const version = ref('加载中...')
const isLight = ref(localStorage.getItem('theme') === 'light')

const pages = {
  browser: '文件管理',
  tasks: '任务管理',
  runs: '运行记录',
}

function switchPage(page: string) {
  currentPage.value = page
}

function toggleTheme() {
  isLight.value = !isLight.value
  if (isLight.value) {
    document.body.classList.add('light')
    localStorage.setItem('theme', 'light')
  } else {
    document.body.classList.remove('light')
    localStorage.setItem('theme', 'dark')
  }
}

onMounted(async () => {
  if (isLight.value) {
    document.body.classList.add('light')
  }
  try {
    const data = await api.listRemotes()
    version.value = data.version || '未知版本'
  } catch {
    version.value = '未连接'
  }
})
</script>

<template>
  <div class="app">
    <!-- Header -->
    <header class="header">
      <div class="header-brand">RcloneFlow <small>{{ version }}</small></div>
      <nav class="header-nav">
        <button
          v-for="(name, key) in pages"
          :key="key"
          :class="{ active: currentPage === key }"
          @click="switchPage(key)"
        >
          {{ name }}
        </button>
      </nav>
      <button class="theme-btn" @click="toggleTheme">
        {{ isLight ? '🌙 深色模式' : '☀️ 浅色模式' }}
      </button>
    </header>

    <!-- Main Content -->
    <main class="main">
      <BrowserView v-if="currentPage === 'browser'" :version="version" />
      <TaskView v-if="currentPage === 'tasks'" />
      <RunView v-if="currentPage === 'runs'" />
    </main>
  </div>
</template>
