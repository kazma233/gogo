<template>
  <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="mb-6 flex flex-wrap gap-2 items-center">
      <span class="text-gray-700 font-medium">标签筛选:</span>
      <button
        v-for="tag in tags"
        :key="tag.id"
        @click="toggleTagFilter(tag.id)"
        class="px-3 py-1 rounded-full text-sm transition"
        :class="selectedTagId === tag.id ? 'ring-2 ring-offset-1' : ''"
        :style="{
          backgroundColor: tag.color,
          color: 'white',
          opacity: selectedTagId === tag.id ? 1 : 0.5
        }"
      >
        {{ tag.name }}
      </button>
      <button
        v-if="selectedTagId !== null"
        @click="clearTagFilter"
        class="text-gray-500 hover:text-gray-700 text-sm ml-2"
      >
        清除筛选
      </button>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        <div
        v-for="item in images"
        :key="item.id"
        class="group relative overflow-hidden rounded-lg shadow hover:shadow-lg transition-shadow cursor-pointer"
        @click="selectImage(item)"
      >
        <img
          :src="item.thumbnail_url || item.url"
          :alt="item.filename"
          class="w-full h-64 object-cover"
        />
        <div class="absolute inset-0 bg-gradient-to-t from-black/70 to-transparent opacity-0 group-hover:opacity-100 transition-opacity">
          <div class="absolute top-2 right-2">
            <button
              v-if="authStore.isAuthenticated"
              @click.stop="copyImageUrl(item.url)"
              class="p-2 bg-white/20 hover:bg-white/30 rounded-lg backdrop-blur-sm transition"
              title="复制图片链接"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
              </svg>
            </button>
          </div>
          <div class="absolute bottom-0 left-0 right-0 p-3">
            <p class="text-white text-sm truncate">{{ item.original_filename }}</p>
            <div v-if="item.tags && item.tags.length > 0" class="flex gap-1 mt-1 flex-wrap">
              <span
                v-for="tag in item.tags.slice(0, 3)"
                :key="tag.id"
                class="px-1.5 py-0.5 rounded text-xs"
                :style="{ backgroundColor: tag.color, color: 'white' }"
              >
                {{ tag.name }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div v-if="!noMore && images.length > 0" class="text-center py-6">
      <button
        @click="loadImages(true)"
        :disabled="loading"
        class="bg-primary text-white px-6 py-2 rounded-lg hover:bg-blue-600 transition disabled:opacity-50"
      >
        {{ loading ? '加载中...' : '加载更多' }}
      </button>
    </div>
    <div v-else-if="noMore" class="text-center py-8 text-gray-500">
      没有更多图片了
    </div>

    <div v-if="!loading && images.length === 0" class="text-center py-16">
      <p class="text-gray-500 text-lg">暂无图片</p>
      <router-link v-if="authStore.isAuthenticated" to="/upload" class="text-primary hover:underline mt-4 inline-block">
        去上传第一张图片
      </router-link>
    </div>
  </main>

    <ImageModal
      v-if="selectedImage"
      :show="true"
      :image="selectedImage"
      :showDelete="authStore.isAuthenticated"
      :onDeleted="removeImage"
      @close="selectedImage = null"
    />
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { imagesApi, tagsApi } from '@/api'
import type { Image, Tag } from '@/types'
import ImageModal from '@/components/gallery/ImageModal.vue'

const showToast = (message: string) => {
  const toast = document.createElement('div')
  toast.style.cssText = 'position:fixed;bottom:20px;left:50%;transform:translateX(-50%);background:#10B981;color:white;padding:8px 16px;border-radius:8px;font-size:14px;z-index:9999;'
  toast.textContent = message
  document.body.appendChild(toast)
  setTimeout(() => toast.remove(), 2000)
}

const authStore = useAuthStore()
const images = ref<Image[]>([])
const tags = ref<Tag[]>([])
const loading = ref(false)
const noMore = ref(false)
  const page = ref(1)
  const selectedTagId = ref<number | null>(null)
  const selectedImage = ref<Image | null>(null)

const loadTags = async () => {
  try {
    const res = await tagsApi.getAll()
    tags.value = (res.data as Tag[]) || []
  } catch (error) {
    console.error('加载标签失败:', error)
  }
}

const loadImages = async (loadMore = false) => {
  if (loading.value || noMore.value) return

  if (loadMore) {
    page.value++
  } else {
    page.value = 1
  }

  const limit = 20
  loading.value = true
  try {
    const params: { page: number; limit: number; tag_id?: number } = { page: page.value, limit }
    if (selectedTagId.value !== null) {
      params.tag_id = selectedTagId.value
    }

    const res = await imagesApi.getList(params)
    const result = res.data as { data: Image[] }
    if (result?.data && result.data.length > 0) {
      if (loadMore) {
        images.value.push(...result.data)
      } else {
        images.value = result.data
      }
      noMore.value = result.data.length < limit
    } else {
      if (loadMore) {
        noMore.value = true
      } else {
        images.value = []
        noMore.value = true
      }
    }
  } catch (error) {
    console.error('加载图片失败:', error)
  } finally {
    loading.value = false
  }
}

const toggleTagFilter = (tagId: number) => {
  if (selectedTagId.value === tagId) {
    selectedTagId.value = null
  } else {
    selectedTagId.value = tagId
  }
  noMore.value = false
  loadImages()
}

const clearTagFilter = () => {
  selectedTagId.value = null
  noMore.value = false
  loadImages()
}

const selectImage = async (image: Image) => {
  try {
    const res = await imagesApi.getById(image.id)
    if (res.data) {
      selectedImage.value = res.data as Image
    }
  } catch (error) {
    console.error('获取图片详情失败:', error)
  }
}

const removeImage = (id: number) => {
  images.value = images.value.filter(img => img.id !== id)
}

const copyImageUrl = async (url: string) => {
  try {
    await navigator.clipboard.writeText(url)
    showToast('已复制链接')
  } catch (error) {
    console.error('复制失败:', error)
  }
}

onMounted(async () => {
  await loadTags()
  await loadImages()
})
</script>
