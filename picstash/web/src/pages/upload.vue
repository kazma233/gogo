<template>
  <div class="min-h-screen bg-gray-50 py-8">
    <div class="max-w-2xl mx-auto bg-white rounded-lg shadow-md p-8">
      <h1 class="text-2xl font-bold mb-6">上传图片</h1>

      <div v-if="authStore.isAuthenticated" class="mb-4 flex items-center gap-3">
        <button
          @click="handleSync"
          :disabled="syncing"
          class="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-lg transition flex items-center gap-2 disabled:opacity-50 cursor-pointer"
        >
          <span v-if="syncing">同步中...</span>
          <span v-else>从存储同步</span>
        </button>
        <router-link
          to="/sync"
          class="text-primary hover:underline text-sm"
        >
          查看同步日志
        </router-link>
      </div>

      <div v-if="syncResult" class="mb-4 p-3 bg-blue-50 rounded-lg text-sm">
        <p class="font-medium mb-1">✅ 同步完成</p>
        <p>新增: {{ syncResult.created_count }} | 更新: {{ syncResult.updated_count }} | 删除: {{ syncResult.deleted_count }} | 跳过: {{ syncResult.skipped_count }} | 错误: {{ syncResult.error_count }}</p>
      </div>

      <form @submit.prevent="handleUpload">
        <div class="mb-6">
          <label class="block text-gray-700 mb-2">选择文件</label>
          <input
            type="file"
            multiple
            accept="image/*"
            @change="handleFileSelect"
            class="w-full file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100 cursor-pointer"
          />
        </div>

        <div class="mb-6" v-if="files.length > 0">
          <h3 class="text-lg font-semibold mb-2">已选文件 ({{ files.length }})</h3>
          
          <div class="mb-4 -ml-2 overflow-x-auto flex gap-3 pb-2 pl-4 scrollbar-thin scrollbar-thumb-gray-300 pt-1">
            <div
              v-for="(file, index) in files"
              :key="index"
              class="relative flex-shrink-0"
              @click="selectedIndex = index"
            >
              <div class="w-16 h-16" :class="{ 'ring-2 ring-blue-500 rounded': selectedIndex === index }">
                <img
                  :src="file.previewUrl"
                  class="w-full h-full object-cover rounded cursor-pointer"
                />
              </div>
              <button
                @click.stop="removeFile(index)"
                type="button"
                class="absolute top-0 right-0 bg-red-500 text-white rounded-bl-lg w-4 h-4 flex items-center justify-center cursor-pointer hover:bg-red-600 transition text-xs"
              >
                ×
              </button>
            </div>
          </div>

          <div>
            <canvas
              ref="mainPreviewCanvas"
              class="w-full max-h-[70vh] object-contain rounded bg-gray-100"
            ></canvas>
          </div>
        </div>

        <div class="mb-6">
          <button
            type="button"
            @click="showWatermarkSettings = !showWatermarkSettings"
            class="flex items-center gap-2 text-gray-700 hover:text-gray-900 px-3 py-2 rounded-lg hover:bg-gray-100 transition cursor-pointer"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            水印设置
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 transition-transform" :class="{ 'rotate-180': showWatermarkSettings }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </button>

          <div v-if="showWatermarkSettings" class="mt-4 p-4 bg-gray-50 rounded-lg">
            <div class="mb-4">
              <label class="flex items-center gap-2">
                <input
                  type="checkbox"
                  v-model="watermarkConfig.enabled"
                  @change="updateMainPreview"
                  class="w-4 h-4"
                />
                <span class="text-gray-700">启用水印</span>
              </label>
            </div>

            <div class="mb-4">
              <label class="block text-gray-700 mb-1">水印文字</label>
              <input
                type="text"
                v-model="watermarkConfig.text"
                @input="updateMainPreview"
                placeholder="输入水印文字"
                class="w-full px-3 py-2 border rounded-lg"
              />
            </div>

            <div class="mb-4">
              <label class="block text-gray-700 mb-1">位置</label>
              <div class="flex items-center gap-2 mb-2">
                <input
                  type="checkbox"
                  v-model="watermarkConfig.fullscreen"
                  @change="updateMainPreview"
                  class="w-4 h-4"
                />
                <span class="text-gray-700">全屏铺满</span>
              </div>
              <select
                v-model="watermarkConfig.position"
                @change="updateMainPreview"
                class="w-full px-3 py-2 border rounded-lg disabled:bg-gray-100 disabled:text-gray-400 disabled:cursor-not-allowed"
                :disabled="watermarkConfig.fullscreen"
              >
                <option value="top-left">左上</option>
                <option value="top-center">中上</option>
                <option value="top-right">右上</option>
                <option value="center-left">左中</option>
                <option value="center">居中</option>
                <option value="center-right">右中</option>
                <option value="bottom-left">左下</option>
                <option value="bottom-center">中下</option>
                <option value="bottom-right">右下</option>
              </select>
            </div>

            <div class="mb-4">
              <label class="block text-gray-700 mb-1">字体大小</label>
              <input
                type="number"
                v-model.number="watermarkConfig.size"
                min="1"
                @input="updateMainPreview"
                class="w-full px-3 py-2 border rounded-lg"
              />
            </div>

            <div class="mb-4">
              <label class="block text-gray-700 mb-1">颜色</label>
              <div class="flex items-center gap-2">
                <input
                  type="color"
                  v-model="watermarkConfig.color"
                  @input="updateMainPreview"
                  class="w-10 h-10 border rounded-lg cursor-pointer"
                />
                <span class="text-gray-600">{{ watermarkConfig.color }}</span>
              </div>
            </div>

            <div class="mb-4">
              <label class="block text-gray-700 mb-1">
                透明度: {{ watermarkConfig.opacity }}
              </label>
              <input
                type="range"
                v-model.number="watermarkConfig.opacity"
                min="0.1"
                max="1"
                step="0.1"
                @input="updateMainPreview"
                class="w-full"
              />
            </div>

            <div class="mb-4" v-if="watermarkConfig.fullscreen">
              <label class="block text-gray-700 mb-1">
                间距: {{ watermarkConfig.spacing }}px
              </label>
              <input
                type="range"
                v-model.number="watermarkConfig.spacing"
                min="0"
                max="100"
                @input="updateMainPreview"
                class="w-full"
              />
            </div>

            <div class="mb-4" v-if="!watermarkConfig.fullscreen">
              <label class="block text-gray-700 mb-1">
                边距: {{ watermarkConfig.padding }}px
              </label>
              <input
                type="range"
                v-model.number="watermarkConfig.padding"
                min="0"
                max="100"
                @input="updateMainPreview"
                class="w-full"
              />
            </div>

            <div class="mb-4">
              <label class="block text-gray-700 mb-1">
                旋转角度: {{ watermarkConfig.rotation }}°
              </label>
              <input
                type="range"
                v-model.number="watermarkConfig.rotation"
                min="0"
                max="360"
                @input="updateMainPreview"
                class="w-full"
              />
            </div>
          </div>
        </div>

        <div class="mb-6">
          <div class="flex items-center justify-between mb-2">
            <label class="block text-gray-700">标签</label>
            <div class="flex gap-2">
              <button
                @click="loadTags"
                type="button"
                class="p-1.5 bg-gray-100 hover:bg-gray-200 rounded-lg flex items-center justify-center cursor-pointer"
                title="刷新标签"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
              </button>
<button
                  @click="openTagsPage"
                  type="button"
                class="p-1.5 bg-gray-100 hover:bg-gray-200 rounded-lg flex items-center justify-center cursor-pointer"
                title="新增标签"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
              </button>
            </div>
          </div>
          <div v-if="tags.length === 0" class="text-gray-400 text-sm py-2">
            暂无标签，请前往标签管理页面创建
          </div>
          <div v-else class="flex flex-wrap gap-2">
            <button
              v-for="tag in tags"
              :key="tag.id"
              type="button"
              @click="toggleTag(tag.id)"
              class="px-3 py-1.5 rounded-full text-sm transition border-2 cursor-pointer"
              :class="selectedTags.includes(String(tag.id)) 
                ? 'border-transparent text-white' 
                : 'border-gray-300 text-gray-700 hover:border-gray-400'"
              :style="selectedTags.includes(String(tag.id)) ? { backgroundColor: tag.color } : {}"
            >
              {{ tag.name }}
            </button>
          </div>
        </div>

        <button
          type="submit"
          :disabled="uploading || files.length === 0 || processing"
          class="w-full bg-blue-500 text-white py-3 rounded-lg hover:bg-blue-600 transition disabled:opacity-50 cursor-pointer"
        >
          {{ uploading ? '上传中...' : (processing ? '处理中...' : '上传') }}
        </button>
      </form>

      <div v-if="uploadMessage" class="mt-4" :class="uploadMessageType">
        {{ uploadMessage }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { imagesApi, tagsApi } from '@/api'
import type { Tag, WatermarkConfig, SyncResult } from '@/types'
import { defaultWatermarkConfig } from '@/types'
import { loadWatermarkConfig, saveWatermarkConfig } from '@/utils/storage'
import { createWatermarkCanvas, canvasToBlob, getImageDimensions } from '@/utils/watermark'

interface FileWithPreview extends File {
  previewUrl?: string
  img?: HTMLImageElement
  dimensions?: { width: number; height: number }
  needsThumbnail?: boolean
  thumbnailBlob?: Blob
}

const router = useRouter()
const authStore = useAuthStore()

const files = ref<FileWithPreview[]>([])
const tags = ref<Tag[]>([])
const selectedTags = ref<string[]>([])
const uploading = ref(false)
const processing = ref(false)
const uploadMessage = ref('')
const uploadMessageType = ref('text-green-600')
const objectUrls: string[] = []

const syncing = ref(false)
const syncResult = ref<SyncResult | null>(null)

const showWatermarkSettings = ref(false)
const watermarkConfig = ref<WatermarkConfig>({ ...defaultWatermarkConfig })
const mainPreviewCanvas = ref<HTMLCanvasElement | null>(null)
const selectedIndex = ref(0)
const thumbnailBlobs = ref<(Blob | null)[]>([])

const getPreviewUrl = (file: File): string => {
  const url = URL.createObjectURL(file)
  objectUrls.push(url)
  return url
}

const revokePreviewUrls = () => {
  objectUrls.forEach(url => URL.revokeObjectURL(url))
  objectUrls.length = 0
}

const handleSync = async () => {
  syncing.value = true
  syncResult.value = null
  try {
    const res = await imagesApi.sync()
    syncResult.value = res.data.data as SyncResult
    alert('同步成功')
  } catch (error) {
    alert('同步失败')
    console.error('同步失败:', error)
  } finally {
    syncing.value = false
  }
}

const loadTags = async () => {
  try {
    const res = await tagsApi.getAll()
    tags.value = (res.data as Tag[]) || []
  } catch (error) {
    console.error('加载标签失败:', error)
  }
}

onMounted(async () => {
  await loadTags()
  watermarkConfig.value = loadWatermarkConfig()
})

onUnmounted(() => {
  revokePreviewUrls()
})

const handleFileSelect = async (e: Event) => {
  const target = e.target as HTMLInputElement
  const newFiles = Array.from(target.files || [])

  if (newFiles.length === 0) return

  target.value = ''

  const processedNewFiles = await Promise.all(newFiles.map(async (file) => {
    const dimensions = await getImageDimensions(file)
    const needsThumbnail = watermarkConfig.value.enabled || dimensions.width >= 1920 || dimensions.height >= 1080

    const img = new Image()
    img.src = getPreviewUrl(file)
    await new Promise<void>((resolve) => {
      img.onload = () => resolve()
    })

    const fileWithPreview: FileWithPreview = Object.assign(file, {
      previewUrl: URL.createObjectURL(file),
      img,
      dimensions,
      needsThumbnail,
    })

    return fileWithPreview
  }))

  files.value = [...files.value, ...processedNewFiles]
  await nextTick()
  await updateMainPreview()
}

const removeFile = (index: number) => {
  const url = files.value[index].previewUrl
  if (url) {
    URL.revokeObjectURL(url)
    const urlIndex = objectUrls.indexOf(url)
    if (urlIndex > -1) objectUrls.splice(urlIndex, 1)
  }
  files.value.splice(index, 1)
  thumbnailBlobs.value.splice(index, 1)

  if (selectedIndex.value >= files.value.length) {
    selectedIndex.value = Math.max(0, files.value.length - 1)
  }

  if (files.value.length > 0) {
    updateMainPreview()
  } else {
    selectedIndex.value = 0
  }
}

const toggleTag = (tagId: number) => {
  const idStr = String(tagId)
  const index = selectedTags.value.indexOf(idStr)
  if (index > -1) {
    selectedTags.value.splice(index, 1)
  } else {
    selectedTags.value.push(idStr)
  }
}

const updateMainPreview = async () => {
  saveWatermarkConfig(watermarkConfig.value)

  if (!mainPreviewCanvas.value || files.value.length === 0) return

  const currentFile = files.value[selectedIndex.value]
  if (!currentFile || !currentFile.img) return

  const canvas = mainPreviewCanvas.value
  const ctx = canvas.getContext('2d')!
  
  canvas.width = currentFile.img.naturalWidth
  canvas.height = currentFile.img.naturalHeight

  const watermarkCanvas = createWatermarkCanvas(currentFile.img, watermarkConfig.value)
  ctx.drawImage(watermarkCanvas, 0, 0)
}

const processFiles = async (): Promise<{
  originalFiles: File[]
  watermarkFiles: File[]
  thumbnailFiles: File[]
  mapping: Array<{original: string, original_name: string, watermark: string | null, thumbnail: string | null}>
}> => {
  processing.value = true
  const originalFiles: File[] = []
  const watermarkFiles: File[] = []
  const thumbnailFiles: File[] = []
  const mapping: Array<{original: string, original_name: string, watermark: string | null, thumbnail: string | null}> = []

  for (let i = 0; i < files.value.length; i++) {
    const file = files.value[i]
    const uuid = crypto.randomUUID()
    const originalName = file.name

    const originalFilename = `original_${uuid}.jpg`
    originalFiles.push(new File([file], originalFilename, { type: file.type }))

    let watermarkFilename: string | null = null
    if (watermarkConfig.value.enabled && file.img) {
      const watermarkCanvas = createWatermarkCanvas(file.img, watermarkConfig.value)
      const watermarkBlob = await canvasToBlob(watermarkCanvas, 0.9)
      watermarkFilename = `watermark_${uuid}.jpg`
      watermarkFiles.push(new File([watermarkBlob], watermarkFilename, { type: 'image/jpeg' }))
    }

    let thumbnailFilename: string | null = null
    const needsThumbnail = file.dimensions && (
      file.dimensions.width >= 1920 ||
      file.dimensions.height >= 1080
    )

    if (needsThumbnail && file.img) {
      const thumbnailCanvas = document.createElement('canvas')
      thumbnailCanvas.width = file.img.naturalWidth
      thumbnailCanvas.height = file.img.naturalHeight
      const ctx = thumbnailCanvas.getContext('2d')!
      ctx.drawImage(file.img, 0, 0)
      const thumbnailBlob = await canvasToBlob(thumbnailCanvas, 0.9)
      thumbnailFilename = `thumb_${uuid}.jpg`
      thumbnailFiles.push(new File([thumbnailBlob], thumbnailFilename, { type: 'image/jpeg' }))
    }

    mapping.push({
      original: originalFilename,
      original_name: originalName,
      watermark: watermarkFilename,
      thumbnail: thumbnailFilename
    })
  }

  processing.value = false
  return { originalFiles, watermarkFiles, thumbnailFiles, mapping }
}

const openTagsPage = () => {
  window.open('/tags', '_blank')
}

const handleUpload = async () => {
  if (files.value.length === 0) {
    uploadMessage.value = '请选择文件'
    uploadMessageType.value = 'text-red-600'
    return
  }

  uploading.value = true
  uploadMessage.value = ''

  try {
    const { originalFiles, watermarkFiles, thumbnailFiles, mapping } = await processFiles()

    const formData = new FormData()
    
    originalFiles.forEach((file) => {
      formData.append('original_files', file)
    })
    
    watermarkFiles.forEach((file) => {
      formData.append('watermark_files', file)
    })
    
    thumbnailFiles.forEach((file) => {
      formData.append('thumbnail_files', file)
    })
    
    formData.append('file_mapping', JSON.stringify(mapping))
    
    selectedTags.value.forEach((tagId) => {
      formData.append('tag_ids', tagId)
    })

    const res = await imagesApi.upload(formData)
    const data = res.data as { count?: number } | undefined
    uploadMessage.value = `上传成功，共 ${data?.count || 0} 张图片`
    uploadMessageType.value = 'text-green-600'
    router.push('/')
  } catch (error: any) {
    uploadMessage.value = error.response?.data?.error || '上传失败'
    uploadMessageType.value = 'text-red-600'
  } finally {
    uploading.value = false
  }
}

watch(() => watermarkConfig.value, () => {
  if (files.value.length > 0) {
    updateMainPreview()
  }
}, { deep: true })

watch(selectedIndex, async () => {
  await nextTick()
  await updateMainPreview()
})
</script>
