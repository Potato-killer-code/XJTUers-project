<template>
  <div class="page">
    <div class="card">
      <div class="icon-wrapper store-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 5v14M5 12h14"/>
          <rect x="3" y="3" width="18" height="18" rx="2"/>
        </svg>
      </div>

      <h1>存入外卖</h1>
      <p class="subtitle">请输入 4 位数字密码，柜门将自动打开</p>

      <!-- 密码输入 -->
      <div class="input-group">
        <input
          v-model="code"
          type="text"
          inputmode="numeric"
          maxlength="4"
          placeholder=""
          :disabled="loading"
          @input="onCodeInput"
          @keyup.enter="handleStore"
        />
      </div>

      <!-- 操作按钮 -->
      <button
        class="btn-primary"
        :disabled="!canSubmit || loading"
        @click="handleStore"
      >
        <span v-if="loading" class="spinner"></span>
        {{ loading ? '处理中...' : '确认存入' }}
      </button>

      <!-- 结果提示 -->
      <div v-if="message" :class="['message', messageType]">
        {{ message }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { storeItem } from '../api/index.js'

const code = ref('')
const loading = ref(false)
const message = ref('')
const messageType = ref('')

const canSubmit = computed(() => code.value.length === 4)

function onCodeInput(e) {
  // 只允许数字
  code.value = e.target.value.replace(/\D/g, '')
}

async function handleStore() {
  if (!canSubmit.value || loading.value) return

  loading.value = true
  message.value = ''

  const res = await storeItem(code.value)

  if (res.code === 0) {
    messageType.value = 'success'
    message.value = res.message
    code.value = ''
  } else {
    messageType.value = 'error'
    message.value = res.message
  }

  loading.value = false
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.card {
  background: #fff;
  border-radius: 20px;
  padding: 48px 40px;
  width: 100%;
  max-width: 420px;
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.08);
  text-align: center;
}

.icon-wrapper {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 24px;
}

.icon-wrapper svg {
  width: 36px;
  height: 36px;
}

.store-icon {
  background: linear-gradient(135deg, #e0f2fe, #bae6fd);
  color: #0284c7;
}

h1 {
  font-size: 24px;
  font-weight: 700;
  color: #1e293b;
  margin: 0 0 8px;
}

.subtitle {
  font-size: 14px;
  color: #94a3b8;
  margin: 0 0 32px;
}

.input-group {
  margin-bottom: 24px;
}

input {
  width: 200px;
  padding: 14px 20px;
  border: 2px solid #e2e8f0;
  border-radius: 12px;
  font-size: 28px;
  text-align: center;
  letter-spacing: 12px;
  color: #1e293b;
  outline: none;
  transition: border-color 0.2s;
  font-family: 'Courier New', monospace;
}

input:focus {
  border-color: #0284c7;
}

input:disabled {
  background: #f8fafc;
}

.btn-primary {
  width: 100%;
  padding: 14px 0;
  background: linear-gradient(135deg, #0284c7, #0ea5e9);
  color: #fff;
  border: none;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s, transform 0.1s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-primary:hover:not(:disabled) {
  opacity: 0.9;
}

.btn-primary:active:not(:disabled) {
  transform: scale(0.98);
}

.btn-primary:disabled {
  background: #cbd5e1;
  cursor: not-allowed;
}

.spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.4);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.message {
  margin-top: 24px;
  padding: 12px 20px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
}

.message.success {
  background: #f0fdf4;
  color: #16a34a;
}

.message.error {
  background: #fef2f2;
  color: #dc2626;
}
</style>
