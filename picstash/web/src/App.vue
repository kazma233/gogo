<template>
  <div id="app" class="min-h-screen bg-gray-50">
    <nav class="bg-white shadow">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex items-center">
            <router-link to="/" class="text-xl font-bold text-primary">Picstash</router-link>
          </div>
          <div class="flex items-center space-x-4">
            <router-link v-if="!authStore.isAuthenticated" to="/login" class="text-gray-700 hover:text-primary">
              登录
            </router-link>
            <template v-else>
              <router-link to="/upload" class="text-gray-700 hover:text-primary">上传</router-link>
              <router-link to="/tags" class="text-gray-700 hover:text-primary">标签</router-link>
              <button @click="authStore.logout" class="text-gray-700 hover:text-danger">
                退出
              </button>
            </template>
          </div>
        </div>
      </div>
    </nav>
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
</script>
