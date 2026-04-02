<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import BrowserView from './views/BrowserView.vue'
import TaskView from './views/TaskView.vue'
import LoginView from './views/LoginView.vue'
import * as api from './api'
import { isLoggedIn as checkAuth, getUser, logout, changePassword } from './api/auth'

const currentPage = ref('browser')
const taskViewKey = ref(0)
const version = ref('加载中...')
const isLight = ref(localStorage.getItem('theme') === 'light')
const isAuth = ref(false)
const authChecked = ref(false)
const showSettings = ref(false)
const showPasswordModal = ref(false)

const user = getUser()

const passwordForm = reactive({
  username: user?.username || '',
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

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

function handleLogout() {
  logout()
  isAuth.value = false
}

async function handleChangePassword() {
  if (passwordForm.newPassword && passwordForm.newPassword !== passwordForm.confirmPassword) {
    alert('新密码与确认密码不一致')
    return
  }
  if (passwordForm.newPassword && passwordForm.newPassword.length < 6) {
    alert('新密码长度至少6位')
    return
  }
  
  try {
    await changePassword(passwordForm.oldPassword, passwordForm.newPassword, passwordForm.username)
    alert('修改成功，请重新登录')
    showPasswordModal.value = false
    handleLogout()
  } catch (e: any) {
    alert(e.message || '修改失败')
  }
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
        <div class="header-actions">
          <button class="settings-btn" @click="showPasswordModal = true">
            {{ user?.username }}
          </button>
          <button class="theme-btn" @click="toggleTheme">
            {{ isLight ? '🌙' : '☀️' }}
          </button>
        </div>
      </header>

      <!-- Main Content -->
      <main class="main">
        <BrowserView v-if="currentPage === 'browser'" :version="version" />
        <TaskView v-if="currentPage === 'tasks'" :key="taskViewKey" />
      </main>

      <!-- 修改密码弹窗 -->
      <div v-if="showPasswordModal" class="modal-overlay" @click.self="showPasswordModal = false">
        <div class="modal-content">
          <div class="modal-header">
            <h3>修改密码</h3>
            <button class="close-btn" @click="showPasswordModal = false">×</button>
          </div>
          <div class="modal-body">
            <div class="field-item">
              <label>用户名</label>
              <input v-model="passwordForm.username" type="text" placeholder="输入新用户名" />
            </div>
            <div class="field-item">
              <label>旧密码</label>
              <input v-model="passwordForm.oldPassword" type="password" placeholder="输入旧密码（修改密码时必填）" />
            </div>
            <div class="field-item">
              <label>新密码</label>
              <input v-model="passwordForm.newPassword" type="password" placeholder="输入新密码（留空则不修改）" />
            </div>
            <div class="field-item">
              <label>确认新密码</label>
              <input v-model="passwordForm.confirmPassword" type="password" placeholder="再次输入新密码" />
            </div>
          </div>
          <div class="modal-footer">
            <button class="ghost" @click="showPasswordModal = false">取消</button>
            <button class="primary" @click="handleChangePassword">保存</button>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.settings-btn {
  padding: 6px 12px;
  background: #252525;
  border: 1px solid #333;
  border-radius: 8px;
  color: #ccc;
  font-size: 13px;
  cursor: pointer;
}

.settings-btn:hover {
  background: #333;
  color: #fff;
}

body.light .settings-btn {
  background: #f0f0f0;
  border-color: #ddd;
  color: #666;
}

body.light .settings-btn:hover {
  background: #e0e0e0;
  color: #333;
}

.theme-btn {
  padding: 6px 12px;
  background: transparent;
  border: 1px solid #333;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
}

body.light .theme-btn {
  border-color: #ddd;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: #1e1e2f;
  border-radius: 16px;
  padding: 24px;
  width: 90%;
  max-width: 400px;
  border: 1px solid #333;
}

body.light .modal-content {
  background: #fff;
  border-color: #ddd;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  color: #fff;
}

body.light .modal-header h3 {
  color: #1a1a1a;
}

.close-btn {
  background: transparent;
  border: none;
  color: #888;
  font-size: 24px;
  cursor: pointer;
}

.close-btn:hover {
  color: #fff;
}

.modal-body {
  margin-bottom: 20px;
}

.field-item {
  margin-bottom: 16px;
}

.field-item label {
  display: block;
  font-size: 13px;
  color: #888;
  margin-bottom: 6px;
}

.field-item input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #333;
  border-radius: 8px;
  background: #252525;
  color: #e0e0e0;
  font-size: 14px;
}

body.light .field-item input {
  background: #f5f5f5;
  border-color: #ddd;
  color: #1a1a1a;
}

.field-item input:focus {
  outline: none;
  border-color: #64b5f6;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.modal-footer button {
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
}

.modal-footer .ghost {
  background: transparent;
  border: 1px solid #333;
  color: #ccc;
}

.modal-footer .primary {
  background: #64b5f6;
  border: none;
  color: #fff;
}

body.light .modal-footer .ghost {
  border-color: #ddd;
  color: #666;
}
</style>
