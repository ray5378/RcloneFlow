<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-header">
        <h1>RcloneFlow</h1>
        <p>{{ t('login.subtitle') }}</p>
      </div>

      <form @submit.prevent="handleSubmit">
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
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { login } from '../api/auth'
import { t } from '../i18n'

const emit = defineEmits<{ (e: 'success'): void }>()
const loading = ref(false)
const error = ref('')
const form = reactive({ username: '', password: '' })

async function handleSubmit() {
  error.value = ''
  loading.value = true
  try {
    const data = await login(form.username, form.password)
    localStorage.setItem('authToken', data.accessToken)
    localStorage.setItem('refreshToken', data.refreshToken)
    localStorage.setItem('user', JSON.stringify(data.user))
    emit('success')
  } catch (e: any) {
    error.value = e.message || t('login.failed')
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
