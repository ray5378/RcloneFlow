<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import BrowserView from './views/BrowserView.vue'
import TaskView from './views/TaskView.vue'
import LoginView from './views/LoginView.vue'
import * as api from './api'
import { isLoggedIn as checkAuth } from './api/auth'

const currentPage = ref('browser')
const taskViewKey = ref(0)
const version = ref('加载中...')
const isLight = ref(localStorage.getItem('theme') === 'light')
const isAuth = ref(false)
const authChecked = ref(false)

const pages = {
  browser: '文件管理',
  tasks: '任务管理',
}

function switchPage(page: string) {
  currentPage.value = page
  if (page === 'tasks') {
    // 切换到任务管理时重置视图
    taskViewKey.value++
  }
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

function handleLoginSuccess() {
  isAuth.value = true
}

onMounted(async () => {
  if (isLight.value) {
    document.body.classList.add('light')
  }
  
  // 检查登录状态
  isAuth.value = checkAuth()
  authChecked.value = true
  
  if (!isAuth.value) {
    return
  }
  
  try {
    const data = await api.listRemotes()
    version.value = data.version || '未知版本'
  } catch {
    // 如果请求失败（可能是401），可能需要重新登录
    version.value = '未连接'
  }
})
</script>

<template>
  <div class="app">
    <!-- 登录页面 -->
    <LoginView v-if="authChecked && !isAuth" @success="handleLoginSuccess" />
    
    <!-- 已登录的主应用 -->
    <template v-else-if="authChecked && isAuth">
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
        <TaskView v-if="currentPage === 'tasks'" :key="taskViewKey" />
      </main>
    </template>
  </div>
</template>
