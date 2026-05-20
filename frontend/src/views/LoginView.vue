<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-header">
        <h1>RcloneFlow</h1>
        <p>{{ step === 'change' ? t('login.changeSubtitle') : t('login.subtitle') }}</p>
      </div>

      <form v-if="step === 'login'" @submit.prevent="handleLoginSubmit">
        <div class="field-item">
          <label>{{ t('login.username') }}</label>
          <input
            v-model="form.username"
            type="text"
            :placeholder="t('login.usernamePlaceholder')"
            required
          />
        </div>

        <div class="field-item">
          <label>{{ t('login.password') }}</label>
          <input
            v-model="form.password"
            type="password"
            :placeholder="t('login.passwordPlaceholder')"
            required
          />
        </div>

        <div v-if="error" class="error-message">{{ error }}</div>

        <button type="submit" class="primary-btn" :disabled="loading">
          {{ loading ? t('login.submitting') : t('login.submit') }}
        </button>
      </form>

      <form v-else-if="step === 'change'" @submit.prevent="handleChangeSubmit">
        <p class="change-hint">{{ t('login.changeHint') }}</p>

        <div class="field-item">
          <label>{{ t('login.newUsername') }}</label>
          <input
            v-model="changeForm.newUsername"
            type="text"
            :placeholder="t('login.newUsernamePlaceholder')"
            required
          />
        </div>

        <div class="field-item">
          <label>{{ t('login.newPassword') }}</label>
          <input
            v-model="changeForm.newPassword"
            type="password"
            :placeholder="t('login.newPasswordPlaceholder')"
            required
          />
        </div>

        <div class="field-item">
          <label>{{ t('login.confirmPassword') }}</label>
          <input
            v-model="changeForm.confirmPassword"
            type="password"
            :placeholder="t('login.confirmPasswordPlaceholder')"
            required
          />
        </div>

        <div v-if="error" class="error-message">{{ error }}</div>

        <button type="submit" class="primary-btn" :disabled="loading">
          {{ loading ? t('login.saving') : t('login.saveAndContinue') }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { login, setTokens, changePassword } from '../api/auth'
import { t } from '../i18n'

const emit = defineEmits<{ (e: 'success'): void }>()
const loading = ref(false)
const error = ref('')
const step = ref<'login' | 'change'>('login')
const form = reactive({ username: '', password: '' })
const changeForm = reactive({ newUsername: '', newPassword: '', confirmPassword: '' })
const loginPassword = ref('')

async function handleLoginSubmit() {
  error.value = ''
  loading.value = true
  try {
    const data = await login(form.username, form.password)
    if (data.mustChangePassword) {
      setTokens(data.accessToken, data.refreshToken)
      localStorage.setItem('user', JSON.stringify(data.user))
      loginPassword.value = form.password
      step.value = 'change'
    } else {
      setTokens(data.accessToken, data.refreshToken)
      localStorage.setItem('user', JSON.stringify(data.user))
      emit('success')
    }
  } catch (e: any) {
    error.value = e.message || t('login.failed')
  } finally {
    loading.value = false
  }
}

async function handleChangeSubmit() {
  error.value = ''
  if (changeForm.newPassword !== changeForm.confirmPassword) {
    error.value = t('account.mismatch')
    return
  }
  if (changeForm.newPassword.length < 6) {
    error.value = t('account.tooShort')
    return
  }
  loading.value = true
  try {
    await changePassword(loginPassword.value, changeForm.newPassword, changeForm.newUsername)
    emit('success')
  } catch (e: any) {
    error.value = e.message || t('login.changeFailed')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page { min-height: 100vh; display: flex; align-items: center; justify-content: center; background: var(--bg); }
.login-card { background: var(--card); border-radius: 16px; padding: 40px; width: 100%; max-width: 400px; border: 1px solid var(--border); }
body.light .login-card { background: var(--card); border-color: var(--border); }
.login-header { text-align: center; margin-bottom: 30px; }
.login-header h1 { font-size: 28px; font-weight: 700; color: var(--accent); margin: 0 0 8px 0; }
.login-header p { color: var(--muted); margin: 0; font-size: 14px; }
.change-hint { background: rgba(245, 158, 11, 0.1); border: 1px solid #f59e0b; border-radius: 8px; padding: 12px 14px; color: #f59e0b; font-size: 13px; margin-bottom: 20px; line-height: 1.6; }
.field-item { margin-bottom: 16px; }
.field-item label { display: block; font-size: 13px; color: var(--muted); margin-bottom: 6px; }
.field-item input { width: 100%; padding: 12px 16px; border: 1px solid var(--border); border-radius: 8px; background: var(--surface); color: var(--text); font-size: 14px; }
body.light .field-item input { background: var(--surface); border-color: var(--border); color: var(--text); }
.field-item input:focus { outline: none; border-color: var(--accent); }
.error-message { background: rgba(211, 47, 47, 0.1); border: 1px solid var(--danger); border-radius: 8px; padding: 10px 14px; color: var(--danger); font-size: 13px; margin-bottom: 16px; }
.primary-btn { width: 100%; padding: 14px; background: var(--accent); border: none; border-radius: 8px; color: #fff; font-size: 15px; font-weight: 600; cursor: pointer; transition: background 0.2s; }
.primary-btn:hover:not(:disabled) { background: var(--accent-strong); }
.primary-btn:disabled { opacity: 0.6; cursor: not-allowed; }
</style>