<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import BrowserView from './views/BrowserView.vue'
import TaskView from './views/TaskView.vue'
import LoginView from './views/LoginView.vue'
import DefaultsModal from './components/DefaultsModal.vue'

import * as api from './api'
import { getSettings } from './api/settings'
import { isLoggedIn as checkAuth, getUser, logout, changePassword } from './api/auth'
import { locale, toggleLocale, t } from './i18n'

const currentPage = ref(localStorage.getItem('currentPage') || (location.hash.replace('#', '') || 'browser'))
const taskViewKey = ref(0)
const version = ref(t('common.loading'))
const isLight = ref(localStorage.getItem('theme') === 'light')
const isAuth = ref(false)
const authChecked = ref(false)
const showSettingsModal = ref(false)
const showPasswordModal = ref(false)
const showDefaultsModal = ref(false)
const showMobileMenu = ref(false)
const runningHintDebugEnabled = ref(false)

const user = getUser()
const passwordForm = reactive({ username: user?.username || '', oldPassword: '', newPassword: '', confirmPassword: '' })

const pages = computed<Record<string, { name: string; icon: string }>>(() => ({
  browser: { name: t('browser.title'), icon: '📁' },
  tasks: { name: t('task.title'), icon: '📋' },
}))

const localeLabel = computed(() => locale.value === 'zh' ? t('locale.en') : t('locale.zh'))

const isMobile = ref(false)
function checkMobile() { isMobile.value = window.innerWidth <= 768 }

function switchPage(page: string) {
  currentPage.value = page
  localStorage.setItem('currentPage', page)
  location.hash = page
  if (page === 'tasks') taskViewKey.value++
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

async function loadRuntimeSettings() {
  try {
    const resp = await getSettings()
    runningHintDebugEnabled.value = String(resp.webhook?.RUNNING_HINT_DEBUG_ENABLED?.effective || 'false') === 'true'
  } catch {
    runningHintDebugEnabled.value = false
  }
}

function handleDefaultsSaved(values: Record<string, string>) {
  runningHintDebugEnabled.value = String(values.RUNNING_HINT_DEBUG_ENABLED || 'false') === 'true'
}

async function handleLoginSuccess() {
  isAuth.value = true
  try {
    const data = await api.listRemotes()
    version.value = data.version || t('common.unknown')
  } catch {
    version.value = t('common.offline')
  }
  await loadRuntimeSettings()
}

function handleLogout() {
  logout()
  isAuth.value = false
  showSettingsModal.value = false
  showMobileMenu.value = false
}

async function handleChangePassword() {
  if (passwordForm.newPassword && passwordForm.newPassword !== passwordForm.confirmPassword) {
    alert(t('account.mismatch'))
    return
  }
  if (passwordForm.newPassword && passwordForm.newPassword.length < 6) {
    alert(t('account.tooShort'))
    return
  }
  try {
    await changePassword(passwordForm.oldPassword, passwordForm.newPassword, passwordForm.username)
    alert(t('account.updateSuccess'))
    showPasswordModal.value = false
    showSettingsModal.value = false
    handleLogout()
  } catch (e: any) {
    alert(e.message || t('account.updateFailed'))
  }
}

function openGitHub() {
  window.open('https://github.com/ray5378/RcloneFlow/tree/master', '_blank')
}

onMounted(async () => {
  if (isLight.value) document.body.classList.add('light')
  const hash = (location.hash || '').replace('#', '')
  if (hash && ['browser', 'tasks'].includes(hash)) currentPage.value = hash
  isAuth.value = checkAuth()
  authChecked.value = true
  checkMobile()
  window.addEventListener('resize', checkMobile)
  if (!isAuth.value) return
  try {
    const data = await api.listRemotes()
    version.value = data.version || t('common.unknown')
  } catch {
    version.value = t('common.offline')
  }
  await loadRuntimeSettings()
})
</script>

<template>
  <div class="app">
    <LoginView v-if="authChecked && !isAuth" @success="handleLoginSuccess" />
    <template v-else-if="authChecked && isAuth">
      <header class="header">
        <button v-if="isMobile" class="mobile-menu-btn" @click="showMobileMenu = !showMobileMenu">
          <span v-if="!showMobileMenu">☰</span>
          <span v-else>✕</span>
        </button>
        <div class="header-brand">RcloneFlow <small v-if="!isMobile">{{ version }}</small></div>
        <nav v-if="!isMobile" class="header-nav">
          <button v-for="(page, key) in pages" :key="key" :class="{ active: currentPage === key }" @click="switchPage(key)">
            {{ page.name }}
          </button>
        </nav>
        <div class="header-actions">
          <button class="settings-btn" @click="showSettingsModal = true">
            <span v-if="isMobile">⚙️</span>
            <span v-else>⚙️ {{ t('settings.title') }}</span>
          </button>
        </div>
      </header>

      <transition name="slide">
        <div v-if="isMobile && showMobileMenu" class="mobile-menu-overlay" @click="showMobileMenu = false">
          <div class="mobile-menu" @click.stop>
            <div class="mobile-menu-header">
              <span class="mobile-menu-title">RcloneFlow</span>
              <button class="close-btn" @click="showMobileMenu = false">×</button>
            </div>
            <nav class="mobile-menu-nav">
              <button v-for="(page, key) in pages" :key="key" :class="{ active: currentPage === key }" @click="switchPage(key)">
                <span class="nav-icon">{{ page.icon }}</span>
                <span>{{ page.name }}</span>
              </button>
            </nav>
            <div class="mobile-menu-footer"><div class="version-info">{{ version }}</div></div>
          </div>
        </div>
      </transition>

      <main class="main">
        <BrowserView v-if="currentPage === 'browser'" :version="version" />
        <TaskView v-if="currentPage === 'tasks'" :key="taskViewKey" :running-hint-debug-enabled="runningHintDebugEnabled" />
      </main>

      <nav v-if="isMobile" class="mobile-bottom-nav">
        <div v-for="(page, key) in pages" :key="key" class="nav-item" :class="{ active: currentPage === key }" @click="switchPage(key)">
          <span class="icon">{{ page.icon }}</span>
          <span>{{ page.name }}</span>
        </div>
      </nav>

      <div v-if="showSettingsModal" class="modal-overlay" @click.self="showSettingsModal = false">
        <div class="modal-content settings-modal">
          <div class="modal-header">
            <h3>{{ t('settings.title') }}</h3>
            <button class="close-btn" @click="showSettingsModal = false">×</button>
          </div>
          <div class="settings-list">
            <div class="settings-item" @click="toggleLocale">
              <span class="settings-icon">🌐</span>
              <span class="settings-text">{{ t('settings.language') }}：{{ locale.value === 'zh' ? t('locale.zh') : t('locale.en') }}</span>
              <span class="settings-arrow">›</span>
            </div>
            <div class="settings-item" @click="showPasswordModal = true; passwordForm.username = user?.username || ''">
              <span class="settings-icon">👤</span>
              <span class="settings-text">{{ t('settings.account') }}</span>
              <span class="settings-arrow">›</span>
            </div>
            <div class="settings-item" @click="toggleTheme">
              <span class="settings-icon">{{ isLight ? '🌙' : '☀️' }}</span>
              <span class="settings-text">{{ isLight ? t('settings.darkMode') : t('settings.lightMode') }}</span>
              <span class="settings-arrow">›</span>
            </div>
            <div class="settings-item" @click="showDefaultsModal = true">
              <span class="settings-icon">🛠️</span>
              <span class="settings-text">{{ t('settings.defaults') }}</span>
              <span class="settings-arrow">›</span>
            </div>
            <div class="settings-item" @click="openGitHub">
              <span class="settings-icon">⭐</span>
              <span class="settings-text">{{ t('settings.star') }}</span>
              <span class="settings-arrow">↗</span>
            </div>
            <div class="settings-item danger" @click="handleLogout">
              <span class="settings-icon">🚪</span>
              <span class="settings-text">{{ t('settings.logout') }}</span>
              <span class="settings-arrow">›</span>
            </div>
          </div>
        </div>
      </div>

      <div v-if="showPasswordModal" class="modal-overlay" @click.self="showPasswordModal = false">
        <div class="modal-content settings-modal">
          <div class="modal-header">
            <h3>{{ t('account.title') }}</h3>
            <button class="close-btn" @click="showPasswordModal = false">×</button>
          </div>
          <div class="modal-body">
            <div class="field-item">
              <label>{{ t('account.username') }}</label>
              <input v-model="passwordForm.username" :placeholder="t('account.usernamePlaceholder')" />
            </div>
            <div class="field-item">
              <label>{{ t('account.oldPassword') }}</label>
              <input v-model="passwordForm.oldPassword" type="password" :placeholder="t('account.oldPasswordPlaceholder')" />
            </div>
            <div class="field-item">
              <label>{{ t('account.newPassword') }}</label>
              <input v-model="passwordForm.newPassword" type="password" :placeholder="t('account.newPasswordPlaceholder')" />
            </div>
            <div class="field-item">
              <label>{{ t('account.confirmPassword') }}</label>
              <input v-model="passwordForm.confirmPassword" type="password" :placeholder="t('account.confirmPasswordPlaceholder')" />
            </div>
          </div>
          <div class="modal-footer">
            <button class="ghost" @click="showPasswordModal = false">{{ t('common.cancel') }}</button>
            <button class="primary" @click="handleChangePassword">{{ t('common.save') }}</button>
          </div>
        </div>
      </div>

      <DefaultsModal v-if="showDefaultsModal" @close="showDefaultsModal = false" @settings-saved="handleDefaultsSaved" />
    </template>
  </div>
</template>

<style scoped>
.app { min-height: 100vh; background: #111; color: #fff; }
.header { height: 56px; display: flex; align-items: center; justify-content: space-between; padding: 0 16px; border-bottom: 1px solid #222; position: sticky; top: 0; z-index: 20; background: rgba(17, 17, 17, .9); backdrop-filter: blur(8px); }
.header-brand { font-size: 20px; font-weight: 800; }
.header-brand small { margin-left: 8px; font-size: 12px; color: #999; }
.header-nav { display: flex; gap: 12px; }
.header-nav button, .mobile-bottom-nav .nav-item, .mobile-menu-nav button, .settings-btn, .close-btn, .modal-footer button { cursor: pointer; }
.header-nav button { background: transparent; border: none; color: #ccc; padding: 8px 12px; border-radius: 8px; }
.header-nav button.active { background: #64b5f6; color: #fff; }
.header-actions { display: flex; gap: 8px; }
.settings-btn { background: transparent; border: 1px solid #333; color: #ddd; border-radius: 10px; padding: 8px 12px; }
.settings-btn:hover { border-color: #64b5f6; color: #fff; }
.locale-btn { min-width: 82px; }
.main { min-height: calc(100vh - 56px); }
.mobile-menu-btn, .close-btn { background: transparent; border: none; color: inherit; font-size: 22px; }
.mobile-menu-overlay, .modal-overlay { position: fixed; inset: 0; background: rgba(0, 0, 0, .45); display: flex; align-items: center; justify-content: center; z-index: 50; }
.mobile-menu { width: 320px; max-width: 92vw; height: 100%; background: #161616; padding: 16px; box-sizing: border-box; }
.mobile-menu-header, .modal-header { display: flex; align-items: center; justify-content: space-between; }
.mobile-menu-title { font-size: 18px; font-weight: 700; }
.mobile-menu-nav { display: flex; flex-direction: column; gap: 10px; margin-top: 24px; }
.mobile-menu-nav button { display: flex; align-items: center; gap: 12px; background: #222; color: #eee; border: 1px solid #2f2f2f; border-radius: 10px; padding: 12px; }
.mobile-menu-nav button.active { border-color: #64b5f6; background: rgba(100, 181, 246, .15); }
.mobile-menu-footer { position: absolute; bottom: 16px; left: 16px; right: 16px; }
.version-info { color: #999; font-size: 12px; }
.mobile-bottom-nav { position: sticky; bottom: 0; display: flex; justify-content: space-around; border-top: 1px solid #222; background: rgba(17, 17, 17, .95); backdrop-filter: blur(8px); padding: 8px 0; }
.mobile-bottom-nav .nav-item { display: flex; flex-direction: column; align-items: center; color: #bbb; font-size: 12px; }
.mobile-bottom-nav .nav-item.active { color: #64b5f6; }
.modal-content { background: #1a1a1a; color: #eee; border: 1px solid #2f2f2f; border-radius: 14px; width: min(520px, 92vw); padding: 18px; box-sizing: border-box; }
.settings-modal { width: min(460px, 92vw); }
.settings-list { display: flex; flex-direction: column; gap: 10px; }
.settings-item { display: flex; align-items: center; gap: 10px; padding: 14px; border-radius: 12px; background: #222; border: 1px solid #2f2f2f; }
.settings-item:hover { border-color: #64b5f6; }
.settings-item.danger .settings-icon, .settings-item.danger .settings-text { color: #ff8a80; }
.settings-icon { font-size: 18px; }
.settings-text { flex: 1; }
.settings-arrow { color: #666; font-size: 18px; }
.modal-body { margin-bottom: 20px; }
.field-item { margin-bottom: 16px; }
.field-item label { display: block; font-size: 13px; color: #888; margin-bottom: 6px; }
.field-item input { width: 100%; padding: 10px 12px; border: 1px solid #333; border-radius: 8px; background: #252525; color: #e0e0e0; font-size: 14px; box-sizing: border-box; }
.field-item input:focus { outline: none; border-color: #64b5f6; }
.modal-footer { display: flex; justify-content: flex-end; gap: 12px; }
.modal-footer button { padding: 10px 20px; border-radius: 8px; font-size: 14px; }
.modal-footer .ghost { background: transparent; border: 1px solid #333; color: #ccc; }
.modal-footer .primary { background: #64b5f6; border: none; color: #fff; }
body.light .app { background: #f5f7fb; color: #1a1a1a; }
body.light .header { background: rgba(255,255,255,.86); border-bottom-color: #e6e8ec; }
body.light .header-brand small, body.light .version-info, body.light .field-item label, body.light .settings-arrow { color: #666; }
body.light .header-nav button { color: #555; }
body.light .settings-btn { border-color: #ddd; color: #444; }
body.light .settings-btn:hover, body.light .settings-item:hover { border-color: #64b5f6; }
body.light .mobile-menu, body.light .modal-content { background: #fff; color: #1a1a1a; border-color: #e6e8ec; }
body.light .mobile-menu-nav button, body.light .settings-item { background: #f8f9fb; border-color: #e8ebef; color: #1f2937; }
body.light .field-item input { background: #f5f5f5; border-color: #ddd; color: #1a1a1a; }
body.light .modal-footer .ghost { border-color: #ddd; color: #666; }
@media (max-width: 768px) {
  .settings-modal { max-width: 100%; }
  .mobile-menu-overlay { align-items: stretch; }
  .mobile-menu { width: 280px; }
}
</style>
tyle>
