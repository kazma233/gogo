<template>
  <div class="min-h-screen bg-gray-50 py-8">
    <div class="max-w-2xl mx-auto bg-white rounded-lg shadow-md p-8">
      <h1 class="text-2xl font-bold mb-6">标签管理</h1>

      <div class="mb-6">
        <h3 class="text-lg font-semibold mb-2">创建标签</h3>
        <form @submit.prevent="createTag" class="flex gap-4">
          <input
            v-model="newTagName"
            type="text"
            class="flex-1 px-3 py-2 border rounded-lg"
            placeholder="标签名称"
          />
          <input
            v-model="newTagColor"
            type="color"
            class="w-20 h-10 border rounded-lg"
          />
          <button
            type="submit"
            :disabled="creating"
            class="bg-primary text-white px-6 py-2 rounded-lg hover:bg-blue-600 transition"
          >
            {{ creating ? '创建中...' : '创建' }}
          </button>
        </form>
      </div>

      <div>
        <h3 class="text-lg font-semibold mb-4">所有标签</h3>
        <div class="grid grid-cols-4 gap-4">
          <div
            v-for="tag in tags"
            :key="tag.id"
            class="p-3 border rounded-lg flex items-center justify-between"
            :style="{ borderColor: tag.color }"
          >
            <span>{{ tag.name }}</span>
            <button
              @click="deleteTag(tag.id)"
              type="button"
              class="text-red-500 hover:text-red-700 ml-2 p-1"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M3 6h18"/>
                <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"/>
                <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/>
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { tagsApi } from '@/api'
import type { Tag } from '@/types'

const tags = ref<Tag[]>([])
const newTagName = ref('')
const newTagColor = ref('#3B82F6')
const creating = ref(false)

onMounted(async () => {
  await loadTags()
})

const loadTags = async () => {
  try {
    const res = await tagsApi.getAll()
    tags.value = (res.data as Tag[]) || []
  } catch (error) {
    console.error('加载标签失败:', error)
  }
}

const createTag = async () => {
  if (!newTagName.value) {
    alert('请输入标签名称')
    return
  }

  creating.value = true
  try {
    await tagsApi.create({ name: newTagName.value, color: newTagColor.value })
    newTagName.value = ''
    await loadTags()
  } catch (error: any) {
    alert(error.response?.data?.error || '创建失败')
  } finally {
    creating.value = false
  }
}

const deleteTag = async (id: number) => {
  if (!confirm('确定要删除这个标签吗？')) return

  try {
    await tagsApi.delete(id)
    await loadTags()
  } catch (error: any) {
    alert(error.response?.data?.error || '删除失败')
  }
}
</script>
