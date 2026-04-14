<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import BrowserView from './views/BrowserView.vue'
import TaskView from './views/TaskView.vue'
import LoginView from './views/LoginView.vue'
import DefaultsModal from './components/DefaultsModal.vue'

import * as api from './api'
import { isLoggedIn as checkAuth, getUser, logout, changePassword } from './api/auth'

const currentPage = ref(localStorage.getItem('currentPage') || (location.hash.replace('#','')||'browser'))
const taskViewKey = ref(0)
const version = ref('加载中...')
const isLight = ref(localStorage.getItem('theme') === 'light')
const isAuth = ref(false)
const authChecked = ref(false)
const showSettingsModal = ref(false)
const showPasswordModal = ref(false)
const showDefaultsModal = ref(false)
const showMobileMenu = ref(false)

const user = getUser()

const passwordForm = reactive({
  username: user?.username || '',
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const pages: Record<string, { name: string; icon: string }> = {
  browser: { name: '文件', icon: '📁' },
  tasks: { name: '任务', icon: '📋' },
}

// 检测是否移动端
const isMobile = ref(false)
function checkMobile() {
  isMobile.value = window.innerWidth <= 768
}

function switchPage(page: string) {
  currentPage.value = page
  localStorage.setItem('currentPage', page)
  location.hash = page
  if (page === 'tasks') {
    taskViewKey.value++
  }
  showMobileMenu.value = false
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

async function handleLoginSuccess() {
  isAuth.value = true
  try {
    const data = await api.listRemotes()
    version.value = data.version || '未知版本'
  } catch {
    version.value = '未连接'
  }
}

function handleLogout() {
  logout()
  isAuth.value = false
  showSettingsModal.value = false
  showMobileMenu.value = false
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
    showSettingsModal.value = false
    handleLogout()
  } catch (e: any) {
    alert(e.message || '修改失败')
  }
}

function openGitHub() {
  window.open('https://github.com/ray5378/RcloneFlow/tree/master', '_blank')
}

onMounted(async () => {
  if (isLight.value) {
    document.body.classList.add('light')
  }

  // 恢复页面状态
  const hash = (location.hash || '').replace('#','')
  if (hash && ['browser','tasks'].includes(hash)) {
    currentPage.value = hash
  }

  isAuth.value = checkAuth()
  authChecked.value = true

  // 检测移动端
  checkMobile()
  window.addEventListener('resize', checkMobile)

  if (!isAuth.value) {
    return
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
    <!-- 登录页面 -->
    <LoginView v-if="authChecked && !isAuth" @success="handleLoginSuccess" />

    <!-- 已登录的主应用 -->
    <template v-else-if="authChecked && isAuth">
      <!-- Header -->
      <header class="header">
        <!-- 移动端汉堡菜单 -->
        <button v-if="isMobile" class="mobile-menu-btn" @click="showMobileMenu = !showMobileMenu">
          <span v-if="!showMobileMenu">☰</span>
          <span v-else>✕</span>
        </button>

        <div class="header-brand">
          RcloneFlow
          <small v-if="!isMobile">{{ version }}</small>
        </div>

        <!-- 桌面端导航 -->
        <nav v-if="!isMobile" class="header-nav">
          <button
            v-for="(page, key) in pages"
            :key="key"
            :class="{ active: currentPage === key }"
            @click="switchPage(key)"
          >
            {{ page.name }}
          </button>
        </nav>

        <div class="header-actions">
          <button class="settings-btn" @click="showSettingsModal = true">
            <span v-if="isMobile">⚙️</span>
            <span v-else>⚙️ 设置</span>
          </button>
        </div>
      </header>

      <!-- 移动端侧边菜单 -->
      <transition name="slide">
        <div v-if="isMobile && showMobileMenu" class="mobile-menu-overlay" @click="showMobileMenu = false">
          <div class="mobile-menu" @click.stop>
            <div class="mobile-menu-header">
              <span class="mobile-menu-title">RcloneFlow</span>
              <button class="close-btn" @click="showMobileMenu = false">×</button>
            </div>
            <nav class="mobile-menu-nav">
              <button
                v-for="(page, key) in pages"
                :key="key"
                :class="{ active: currentPage === key }"
                @click="switchPage(key)"
              >
                <span class="nav-icon">{{ page.icon }}</span>
                <span>{{ page.name }}管理</span>
              </button>
            </nav>
            <div class="mobile-menu-footer">
              <div class="version-info">{{ version }}</div>
            </div>
          </div>
        </div>
      </transition>

      <!-- Main Content -->
      <main class="main">
        <BrowserView v-if="currentPage === 'browser'" :version="version" />
        <TaskView v-if="currentPage === 'tasks'" :key="taskViewKey" />
      </main>

      <!-- 移动端底部导航 -->
      <nav v-if="isMobile" class="mobile-bottom-nav">
        <div
          v-for="(page, key) in pages"
          :key="key"
          class="nav-item"
          :class="{ active: currentPage === key }"
          @click="switchPage(key)"
        >
          <span class="icon">{{ page.icon }}</span>
          <span>{{ page.name }}</span>
        </div>
      </nav>

      <!-- 设置弹窗 -->
      <div v-if="showSettingsModal" class="modal-overlay" @click.self="showSettingsModal = false">
        <div class="modal-content settings-modal">
          <div class="modal-header">
            <h3>设置</h3>
            <button class="close-btn" @click="showSettingsModal = false">×</button>
          </div>
          <div class="settings-list">
            <div class="settings-item" @click="showPasswordModal = true; passwordForm.username = user?.username || ''">
              <span class="settings-icon">👤</span>
              <span class="settings-text">账号管理</span>
              <span class="settings-arrow">›</span>
            </div>
            <div class="settings-item" @click="toggleTheme">
              <span class="settings-icon">{{ isLight ? '🌙' : '☀️' }}</span>
              <span class="settings-text">{{ isLight ? '深色模式' : '浅色模式' }}</span>
              <span class="settings-arrow">›</span>
            </div>
            <div class="settings-item" @click="showDefaultsModal = true">
              <span class="settings-icon">🛠️</span>
              <span class="settings-text">修改默认设置</span>
              <span class="settings-arrow">›</span>
            </div>
            <div class="settings-item" @click="openGitHub">
              <span class="settings-icon">⭐</span>
              <span class="settings-text">给项目点个Star</span>
              <span class="settings-arrow">↗</span>
            </div>
            <div class="settings-item danger" @click="handleLogout">
              <span class="settings-icon">🚪</span>
              <span class="settings-text">退出登录</span>
              <span class="settings-arrow">›</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 修改密码弹窗 -->
      <div v-if="showPasswordModal" class="modal-overlay" @click.self="showPasswordModal = false">
        <div class="modal-content">
          <div class="modal-header">
            <h3>账号管理</h3>
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

    <!-- 修改默认设置弹窗（根级） -->
    <DefaultsModal v-if="showDefaultsModal" @close="showDefaultsModal=false" />
  </div>
</template>

<style scoped>
.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.settings-btn {
  padding: 6px 14px;
  background: transparent;
  border: 1px solid #333;
  border-radius: 8px;
  color: #ccc;
  font-size: 14px;
  cursor: pointer;
}

.settings-btn:hover {
  background: #252525;
  color: #fff;
}

body.light .settings-btn {
  border-color: #ddd;
  color: #666;
}

body.light .settings-btn:hover {
  background: #f0f0f0;
  color: #333;
}

/* 移动端菜单按钮 */
.mobile-menu-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  background: transparent;
  border: none;
  color: var(--text);
  font-size: 20px;
  cursor: pointer;
  border-radius: 8px;
  padding: 0;
}

.mobile-menu-btn:active {
  background: var(--border);
}

/* 移动端侧边菜单 */
.mobile-menu-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 200;
}

.mobile-menu {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 280px;
  background: var(--surface);
  display: flex;
  flex-direction: column;
  box-shadow: 4px 0 24px rgba(0, 0, 0, 0.3);
}

.mobile-menu-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border);
}

.mobile-menu-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--text);
}

.mobile-menu-nav {
  flex: 1;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.mobile-menu-nav button {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: transparent;
  border: none;
  border-radius: 10px;
  color: var(--text);
  font-size: 15px;
  cursor: pointer;
  text-align: left;
  width: 100%;
}

.mobile-menu-nav button .nav-icon {
  font-size: 18px;
}

.mobile-menu-nav button:hover {
  background: var(--border);
}

.mobile-menu-nav button.active {
  background: var(--accent);
  color: #fff;
}

.mobile-menu-footer {
  padding: 16px 20px;
  border-top: 1px solid var(--border);
}

.version-info {
  font-size: 12px;
  color: var(--muted);
}

/* 侧边菜单动画 */
.slide-enter-active,
.slide-leave-active {
  transition: opacity 0.3s ease;
}

.slide-enter-active .mobile-menu,
.slide-leave-active .mobile-menu {
  transition: transform 0.3s ease;
}

.slide-enter-from,
.slide-leave-to {
  opacity: 0;
}

.slide-enter-from .mobile-menu,
.slide-leave-to .mobile-menu {
  transform: translateX(-100%);
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

.settings-modal {
  max-width: 320px;
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

.settings-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.settings-item {
  display: flex;
  align-items: center;
  padding: 14px 16px;
  background: #252525;
  border-radius: 10px;
  cursor: pointer;
  transition: background 0.2s;
}

.settings-item:hover {
  background: #333;
}

body.light .settings-item {
  background: #f5f5f5;
}

body.light .settings-item:hover {
  background: #e8e8e8;
}

.settings-item.danger .settings-icon,
.settings-item.danger .settings-text {
  color: #ef5350;
}

.settings-item.danger:hover {
  background: rgba(239, 83, 80, 0.1);
}

.settings-icon {
  font-size: 18px;
  margin-right: 12px;
}

.settings-text {
  flex: 1;
  color: #e0e0e0;
  font-size: 14px;
}

body.light .settings-text {
  color: #333;
}

.settings-arrow {
  color: #666;
  font-size: 18px;
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

/* 移动端适配 */
@media (max-width: 768px) {
  .settings-modal {
    max-width: 100%;
  }

  .mobile-menu-overlay {
    align-items: stretch;
  }

  .mobile-menu {
    width: 280px;
  }
}
</style>
