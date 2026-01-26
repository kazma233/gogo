<template>
  <div class="min-h-screen bg-gray-50 flex items-center justify-center">
    <div class="max-w-md w-full bg-white rounded-lg shadow-md p-8">
      <h1 class="text-2xl font-bold text-center mb-6">登录</h1>

      <form @submit.prevent="handleSendCode" v-if="step === 1">
        <div class="mb-4">
          <label class="block text-gray-700 mb-2">邮箱</label>
          <input
            v-model="email"
            type="email"
            class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
            placeholder="your-email@example.com"
          />
        </div>
        <button
          type="submit"
          :disabled="sending || countdown > 0"
          class="w-full bg-primary text-white py-2 rounded-lg hover:bg-blue-600 transition"
        >
          {{ countdown > 0 ? `${countdown}秒后重试` : '发送验证码' }}
        </button>
      </form>

      <form @submit.prevent="handleVerifyCode" v-else>
        <div class="mb-4">
          <label class="block text-gray-700 mb-2">验证码</label>
          <input
            v-model="code"
            type="text"
            maxlength="6"
            class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-center text-2xl tracking-widest"
            placeholder="123456"
          />
        </div>
        <button
          type="button"
          @click="step = 1"
          class="w-full mb-3 bg-gray-200 text-gray-700 py-2 rounded-lg hover:bg-gray-300 transition"
        >
          返回
        </button>
        <button
          type="submit"
          :disabled="verifying"
          class="w-full bg-primary text-white py-2 rounded-lg hover:bg-blue-600 transition"
        >
          {{ verifying ? '验证中...' : '验证' }}
        </button>
      </form>

      <div v-if="message" class="mt-4 text-center" :class="messageType">
        {{ message }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const code = ref('')
const step = ref(1)
const sending = ref(false)
const verifying = ref(false)
const countdown = ref(0)
const message = ref('')
const messageType = ref('text-green-600')

let countdownInterval: any = null

const handleSendCode = async () => {
  if (!email.value) {
    message.value = '请输入邮箱'
    messageType.value = 'text-red-600'
    return
  }

  sending.value = true
  try {
    const res = await authStore.sendCode(email.value)
    message.value = res.message || '验证码已发送'
    messageType.value = 'text-green-600'
    step.value = 2
    startCountdown()
  } catch (error: any) {
    message.value = error.response?.data?.error || '发送失败'
    messageType.value = 'text-red-600'
  } finally {
    sending.value = false
  }
}

const handleVerifyCode = async () => {
  if (!code.value) {
    message.value = '请输入验证码'
    messageType.value = 'text-red-600'
    return
  }

  verifying.value = true
  try {
    await authStore.verifyCode(email.value, code.value)
    message.value = '登录成功'
    messageType.value = 'text-green-600'
    router.push('/')
  } catch (error: any) {
    message.value = error.response?.data?.error || '验证失败'
    messageType.value = 'text-red-600'
  } finally {
    verifying.value = false
  }
}

const startCountdown = () => {
  countdown.value = 60
  countdownInterval = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(countdownInterval)
    }
  }, 1000)
}
</script>
