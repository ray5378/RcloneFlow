<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-header">
        <h1>RcloneFlow</h1>
        <p>文件同步管理平台</p>
      </div>

      <form @submit.prevent="handleSubmit">
        <div class="field-item">
          <label>用户名</label>
          <input 
            v-model="form.username" 
            type="text" 
            placeholder="输入用户名"
            required
          />
        </div>

        <div class="field-item">
          <label>密码</label>
          <input 
            v-model="form.password" 
            type="password" 
            placeholder="输入密码"
            required
          />
        </div>

        <div v-if="error" class="error-message">{{ error }}</div>

        <button type="submit" class="primary-btn" :disabled="loading">
          {{ loading ? '登录中...' : '登录' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { login } from '../api/auth'

const emit = defineEmits<{
  (e: 'success'): void
}>()

const loading = ref(false)
const error = ref('')

const form = reactive({
  username: '',
  password: ''
})

async function handleSubmit() {
  error.value = ''
  loading.value = true

  try {
    const data = await login(form.username, form.password)
    
    // 保存token到localStorage
    localStorage.setItem('authToken', data.accessToken)
    localStorage.setItem('refreshToken', data.refreshToken)
    localStorage.setItem('user', JSON.stringify(data.user))
    
    emit('success')
  } catch (e: any) {
    error.value = e.message || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
}

.login-card {
  background: #1e1e2f;
  border-radius: 16px;
  padding: 40px;
  width: 100%;
  max-width: 400px;
  border: 1px solid #333;
}

body.light .login-card {
  background: #fff;
  border-color: #ddd;
}

.login-header {
  text-align: center;
  margin-bottom: 30px;
}

.login-header h1 {
  font-size: 28px;
  font-weight: 700;
  color: #64b5f6;
  margin: 0 0 8px 0;
}

.login-header p {
  color: #888;
  margin: 0;
  font-size: 14px;
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
  padding: 12px 16px;
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

.error-message {
  background: rgba(211, 47, 47, 0.1);
  border: 1px solid #d32f2f;
  border-radius: 8px;
  padding: 10px 14px;
  color: #ef5350;
  font-size: 13px;
  margin-bottom: 16px;
}

.primary-btn {
  width: 100%;
  padding: 14px;
  background: #64b5f6;
  border: none;
  border-radius: 8px;
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.primary-btn:hover:not(:disabled) {
  background: #42a5f5;
}

.primary-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
