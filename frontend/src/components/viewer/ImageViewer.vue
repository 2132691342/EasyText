<script lang="ts" setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import type { EditorTab } from '@/types'
import { ReadFileBytes } from '../../../wailsjs/go/main/App'
import { normalizeBytes } from '@/utils'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  tab: EditorTab
}>()

const loading = ref(true)
const error = ref<string | null>(null)
const imageUrl = ref('')
const scale = ref(1.0)
const naturalSize = ref({ width: 0, height: 0 })

let objectUrl: string | null = null

function getMimeType(ext: string): string {
  const mimeMap: Record<string, string> = {
    'png': 'image/png',
    'jpg': 'image/jpeg',
    'jpeg': 'image/jpeg',
    'gif': 'image/gif',
    'bmp': 'image/bmp',
    'ico': 'image/x-icon',
    'svg': 'image/svg+xml',
    'webp': 'image/webp',
  }
  return mimeMap[ext] || 'image/png'
}

async function loadImage() {
  loading.value = true
  error.value = null

  // Revoke previous object URL to free memory
  if (objectUrl) {
    URL.revokeObjectURL(objectUrl)
    objectUrl = null
  }
  imageUrl.value = ''

  try {
    const data = await ReadFileBytes(props.tab.path)
    if (!data || data.length === 0) {
      error.value = '无法读取图片'
      loading.value = false
      return
    }

    const uint8Array = new Uint8Array(normalizeBytes(data))
    const ext = props.tab.name.split('.').pop()?.toLowerCase() || 'png'
    const mime = getMimeType(ext)

    // Use Blob + object URL instead of base64 for better memory efficiency
    const blob = new Blob([uint8Array], { type: mime })
    objectUrl = URL.createObjectURL(blob)
    imageUrl.value = objectUrl

    // Reset zoom on new image
    scale.value = 1.0
  } catch (err) {
    console.error('Failed to load image:', err)
    error.value = `加载图片失败: ${err}`
    ElMessage.error(`加载图片失败: ${err}`)
  } finally {
    loading.value = false
  }
}

function onImageLoad(e: Event) {
  const img = e.target as HTMLImageElement
  naturalSize.value = {
    width: img.naturalWidth,
    height: img.naturalHeight,
  }
}

function zoomIn() {
  scale.value = Math.min(5.0, +(scale.value + 0.25).toFixed(2))
}

function zoomOut() {
  scale.value = Math.max(0.1, +(scale.value - 0.25).toFixed(2))
}

function resetZoom() {
  scale.value = 1.0
}

onMounted(() => {
  loadImage()
})

onUnmounted(() => {
  if (objectUrl) {
    URL.revokeObjectURL(objectUrl)
    objectUrl = null
  }
})

watch(() => props.tab.path, (newPath, oldPath) => {
  if (newPath !== oldPath) {
    loadImage()
  }
})
</script>

<template>
  <div class="h-full flex flex-col bg-gray-900 dark:bg-[#111]">
    <!-- Image toolbar -->
    <div class="flex items-center justify-center gap-2 px-4 py-1.5 bg-gray-100 dark:bg-[#2d2d2d] border-b border-gray-200 dark:border-gray-700 flex-shrink-0">
      <span class="text-xs text-gray-500 dark:text-gray-400 mr-2">图片查看器</span>
      <button
        class="px-2 py-0.5 text-xs rounded bg-white dark:bg-[#3c3c3c] border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-[#4c4c4c]"
        @click="zoomOut"
      >−</button>
      <span class="text-xs text-gray-600 dark:text-gray-300 min-w-[3rem] text-center">{{ Math.round(scale * 100) }}%</span>
      <button
        class="px-2 py-0.5 text-xs rounded bg-white dark:bg-[#3c3c3c] border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-[#4c4c4c]"
        @click="zoomIn"
      >+</button>
      <button
        class="px-2 py-0.5 text-xs rounded bg-white dark:bg-[#3c3c3c] border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-[#4c4c4c]"
        @click="resetZoom"
      >1:1</button>

      <template v-if="naturalSize.width > 0">
        <div class="mx-1 h-4 w-px bg-gray-300 dark:bg-gray-600"></div>
        <span class="text-xs text-gray-400">{{ naturalSize.width }} × {{ naturalSize.height }}</span>
      </template>
    </div>

    <!-- Image content -->
    <div class="flex-1 overflow-auto flex items-center justify-center p-4">
      <!-- Loading -->
      <div v-if="loading" class="text-center">
        <div class="spinner w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full mx-auto mb-2 animate-spin"></div>
        <p class="text-sm text-gray-400">加载图片中...</p>
      </div>

      <!-- Error -->
      <div v-else-if="error" class="text-center">
        <p class="text-red-500 text-sm">{{ error }}</p>
      </div>

      <!-- Image -->
      <img
        v-else-if="imageUrl"
        :src="imageUrl"
        :alt="tab.name"
        class="max-w-full max-h-full object-contain"
        :style="{ transform: `scale(${scale})`, transformOrigin: 'center center', transition: 'transform 0.15s ease' }"
        @load="onImageLoad"
        draggable="false"
      />
    </div>
  </div>
</template>

<style scoped>
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
.animate-spin {
  animation: spin 1s linear infinite;
}
</style>
