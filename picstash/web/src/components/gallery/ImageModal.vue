<template>
  <div
    v-if="show"
    class="fixed inset-0 bg-black bg-opacity-90 flex flex-col items-center justify-center z-50 p-4"
    @click="$emit('close')"
  >
    <div class="relative max-w-5xl w-full flex flex-col items-center">
      <img
        :src="image.watermark_url || image.url"
        :alt="image.original_filename"
        class="max-w-full max-h-[70vh] object-contain"
        @click.stop
      />
      <button
        @click="$emit('close')"
        class="absolute top-0 right-0 bg-white text-gray-800 rounded-full w-10 h-10 flex items-center justify-center hover:bg-gray-100"
      >
        ×
      </button>
      <div class="mt-4 bg-black bg-opacity-70 text-white p-4 rounded w-full max-w-5xl" @click.stop>
        <div class="flex flex-wrap items-start justify-between gap-4">
          <div class="flex-1">
            <h3 class="text-lg font-bold">{{ image.original_filename }}</h3>
            <p class="text-sm mt-1">
              {{ image.width }} × {{ image.height }} | {{ formatSize(image.size) }}
            </p>
            <div v-if="!authStore.isAuthenticated && image.tags && image.tags.length > 0" class="mt-3">
              <p class="text-xs text-gray-300 mb-2">标签</p>
              <div class="flex flex-wrap gap-2">
                <span
                  v-for="tag in image.tags"
                  :key="tag.id"
                  class="px-3 py-1 rounded-full text-sm"
                  :style="{ backgroundColor: tag.color, color: 'white' }"
                >
                  {{ tag.name }}
                </span>
              </div>
            </div>
            <div v-if="authStore.isAuthenticated" class="mt-3">
              <p class="text-xs text-gray-300 mb-2">标签</p>
              <div v-if="allTags.length === 0" class="text-gray-400 text-sm py-1">
                暂无标签
              </div>
              <div v-else class="flex flex-wrap gap-2">
                <span
                  v-for="tag in allTags"
                  :key="tag.id"
                  type="button"
                  @click="toggleTag(tag.id)"
                  class="px-3 py-1 rounded-full text-sm transition border-2 cursor-pointer"
                  :class="selectedTagIds.includes(String(tag.id))
                    ? 'border-transparent text-white'
                    : 'border-gray-500 text-gray-300 hover:border-gray-400'"
                  :style="selectedTagIds.includes(String(tag.id)) ? { backgroundColor: tag.color } : {}"
                >
                  {{ tag.name }}
                </span>
              </div>
            </div>
          </div>
          <div class="flex flex-col gap-2">
            <button
              v-if="showDelete"
              @click="handleDelete"
              :disabled="deleting"
              class="bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition disabled:opacity-50"
            >
              {{ deleting ? '删除中...' : '删除' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import type { Image, Tag } from '@/types'
import { imagesApi, tagsApi } from '@/api'
import { useAuthStore } from '@/stores/auth'

const props = defineProps<{
  show: boolean
  image: Image
  showDelete?: boolean
  onDeleted?: (id: number) => void
}>()

const emit = defineEmits(['close'])

const authStore = useAuthStore()
const deleting = ref(false)
const allTags = ref<Tag[]>([])
const selectedTagIds = ref<string[]>([])
const savingTags = ref(false)
let previousImageId: number | null = null

const formatSize = (bytes?: number) => {
  if (!bytes) return '-'
  const units = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${units[i]}`
}

const loadTags = async () => {
  try {
    const res = await tagsApi.getAll()
    allTags.value = (res.data as Tag[]) || []
  } catch (error) {
    console.error('加载标签失败:', error)
  }
}

const updateSelectedTags = () => {
  const currentIds = props.image.tags?.map(t => String(t.id)) || []
  const newIds = JSON.stringify(currentIds)
  const oldIds = JSON.stringify(selectedTagIds.value)
  if (newIds !== oldIds) {
    selectedTagIds.value = currentIds
  }
}

const toggleTag = async (tagId: number) => {
  const idStr = String(tagId)
  const index = selectedTagIds.value.indexOf(idStr)
  if (index > -1) {
    selectedTagIds.value.splice(index, 1)
  } else {
    selectedTagIds.value.push(idStr)
  }
  await handleSaveTags(selectedTagIds.value)
}

const handleSaveTags = async (tagIds: string[]) => {
  savingTags.value = true
  const numericIds = tagIds.map(id => parseInt(id))
  try {
    await imagesApi.updateTags(props.image.id, numericIds)
  } catch (error) {
    console.error('更新标签失败:', error)
    updateSelectedTags()
    alert('更新标签失败')
  } finally {
    savingTags.value = false
  }
}

const handleDelete = async () => {
  if (!confirm('确定要删除这张图片吗？')) return

  deleting.value = true
  try {
    await imagesApi.delete(props.image.id)
    emit('close')
    props.onDeleted?.(props.image.id)
  } catch (error) {
    console.error('删除失败:', error)
    alert('删除失败')
  } finally {
    deleting.value = false
  }
}

watch(() => props.image.id, () => {
  if (props.image.id !== previousImageId) {
    previousImageId = props.image.id
    updateSelectedTags()
  }
})

onMounted(async () => {
  await loadTags()
  previousImageId = props.image.id
  updateSelectedTags()
})
</script>
