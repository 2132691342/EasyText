<template>
  <div class="image-editor flex flex-col h-full bg-[var(--theme-bg)] text-[var(--theme-fg)]">
    <!-- Toolbar -->
    <div class="toolbar flex items-center gap-2 px-3 py-2 border-b flex-shrink-0 flex-wrap"
         style="border-color: var(--theme-gutter-bg)">
      <!-- Crop -->
      <ElButton size="small" @click="toggleCrop" :type="isCropping ? 'primary' : 'default'">
        <FileImage class="w-4 h-4 mr-1" /> 裁剪
      </ElButton>

      <!-- Rotate -->
      <ElButtonGroup size="small">
        <ElButton @click="rotate(-90)"><RotateCcw class="w-4 h-4" /> 左旋</ElButton>
        <ElButton @click="rotate(90)"><RotateCw class="w-4 h-4" /> 右旋</ElButton>
        <ElButton @click="rotate(180)"><RotateCw class="w-4 h-4" /> 180°</ElButton>
      </ElButtonGroup>

      <el-divider direction="vertical" />

      <!-- Scale -->
      <span class="text-xs text-[var(--theme-comment)] mr-1">缩放</span>
      <ElButtonGroup size="small">
        <ElButton @click="setZoom(0.5)">50%</ElButton>
        <ElButton @click="setZoom(0.75)">75%</ElButton>
        <ElButton @click="setZoom(1)">100%</ElButton>
        <ElButton @click="setZoom(2)">200%</ElButton>
      </ElButtonGroup>
      <ElSlider v-model="zoomLevel" :min="10" :max="400" :step="5" class="w-24 mx-2" />
      <span class="text-xs w-12">{{ Math.round(zoomLevel) }}%</span>

      <el-divider direction="vertical" />

      <!-- Format -->
      <span class="text-xs text-[var(--theme-comment)] mr-1">格式</span>
      <ElRadioGroup v-model="outputFormat" size="small">
        <ElRadioButton value="png" label="png">PNG</ElRadioButton>
        <ElRadioButton value="jpeg" label="jpeg">JPG</ElRadioButton>
        <ElRadioButton value="webp" label="webp">WebP</ElRadioButton>
      </ElRadioGroup>

      <template v-if="outputFormat !== 'png'">
        <span class="text-xs text-[var(--theme-comment)] ml-2 mr-1">质量</span>
        <ElSlider v-model="quality" :min="10" :max="100" :step="5" class="w-20" />
      </template>

      <div class="flex-1" />

      <ElButton type="primary" size="small" @click="saveImage">
        <Save class="w-4 h-4 mr-1" /> 保存
      </ElButton>
    </div>

    <!-- Canvas Area -->
    <div class="canvas-area flex-1 overflow-auto relative flex items-center justify-center bg-gray-100 dark:bg-gray-900"
         ref="canvasContainer" @mousedown="onCropMouseDown" @mousemove="onCropMouseMove" @mouseup="onCropMouseUp">
      <canvas ref="canvasRef" class="max-w-none" :style="{ transform: `scale(${zoomLevel / 100})` }" />

      <!-- Crop overlay -->
      <div v-if="isCropping && cropArea" class="crop-overlay absolute border-2 border-blue-500 bg-blue-200 bg-opacity-20"
           :style="{
             left: cropArea.x * zoomLevel / 100 + 'px',
             top: cropArea.y * zoomLevel / 100 + 'px',
             width: cropArea.w * zoomLevel / 100 + 'px',
             height: cropArea.h * zoomLevel / 100 + 'px'
           }" />

      <div v-if="!imageLoaded" class="absolute text-gray-400 dark:text-gray-600">
        <FileImage class="w-16 h-16 mx-auto mb-2" />
        <p class="text-sm">暂无图片</p>
      </div>
    </div>

    <!-- Info bar -->
    <div class="info-bar flex items-center gap-4 px-3 py-1 text-xs border-t flex-shrink-0"
         style="border-color: var(--theme-gutter-bg); color: var(--theme-comment)">
      <span>尺寸: {{ imageWidth }} × {{ imageHeight }}</span>
      <span>缩放: {{ Math.round(zoomLevel) }}%</span>
      <span>格式: {{ outputFormat.toUpperCase() }}</span>
      <span v-if="imageSize">文件大小: {{ formatSize(imageSize) }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue'
import { ElButton, ElButtonGroup, ElSlider, ElRadioGroup, ElRadioButton, ElDivider, ElMessage } from 'element-plus'
import { RotateCw, RotateCcw, FileImage, Save } from 'lucide-vue-next'
import { SaveFileBytes, ReadFileBytes } from '../../../wailsjs/go/main/App'
import { normalizeBytes } from '@/utils'

const props = defineProps<{
  filePath: string
  src?: string
}>()

const canvasRef = ref<HTMLCanvasElement | null>(null)
const canvasContainer = ref<HTMLDivElement | null>(null)
const imageLoaded = ref(false)
const imageWidth = ref(0)
const imageHeight = ref(0)
const imageSize = ref(0)
const zoomLevel = ref(100)
const outputFormat = ref('png')
const quality = ref(85)
const isCropping = ref(false)
const cropArea = ref<{ x: number; y: number; w: number; h: number } | null>(null)

// Internal state
let originalImage: HTMLImageElement | null = null
let currentRotation = 0
let cropStart: { x: number; y: number } | null = null

function getCtx(): CanvasRenderingContext2D | null {
  return canvasRef.value?.getContext('2d') || null
}

function loadImage(src: string) {
  const img = new Image()
  img.onload = () => {
    originalImage = img
    imageWidth.value = img.width
    imageHeight.value = img.height
    imageSize.value = img.src.length
    imageLoaded.value = true
    currentRotation = 0
    renderImage()
  }
  img.src = src
}

function renderImage() {
  const canvas = canvasRef.value
  const ctx = getCtx()
  if (!canvas || !ctx || !originalImage) return

  let w = originalImage.width
  let h = originalImage.height

  // Apply rotation
  if (currentRotation === 90 || currentRotation === 270) {
    [w, h] = [h, w]
  }

  canvas.width = w * (zoomLevel.value / 100)
  canvas.height = h * (zoomLevel.value / 100)

  ctx.save()
  ctx.translate(canvas.width / 2, canvas.height / 2)
  ctx.rotate((currentRotation * Math.PI) / 180)
  ctx.drawImage(originalImage, -originalImage.width / 2, -originalImage.height / 2, originalImage.width, originalImage.height)
  ctx.restore()
}

function setZoom(scale: number) {
  zoomLevel.value = scale * 100
}

watch(zoomLevel, () => {
  if (imageLoaded.value) renderImage()
})

function rotate(deg: number) {
  currentRotation = (currentRotation + deg) % 360
  if (currentRotation < 0) currentRotation += 360
  renderImage()
}

function toggleCrop() {
  isCropping.value = !isCropping.value
  cropArea.value = null
  cropStart = null
}

function onCropMouseDown(e: MouseEvent) {
  if (!isCropping.value) return
  const rect = canvasRef.value?.getBoundingClientRect()
  if (!rect) return
  cropStart = {
    x: (e.clientX - rect.left) / (zoomLevel.value / 100),
    y: (e.clientY - rect.top) / (zoomLevel.value / 100),
  }
}

function onCropMouseMove(e: MouseEvent) {
  if (!isCropping.value || !cropStart) return
  const rect = canvasRef.value?.getBoundingClientRect()
  if (!rect) return

  const cx = (e.clientX - rect.left) / (zoomLevel.value / 100)
  const cy = (e.clientY - rect.top) / (zoomLevel.value / 100)
  const x = Math.min(cropStart.x, cx)
  const y = Math.min(cropStart.y, cy)
  const w = Math.abs(cx - cropStart.x)
  const h = Math.abs(cy - cropStart.y)

  cropArea.value = { x, y, w, h }
}

function onCropMouseUp() {
  if (!isCropping.value) return
  cropStart = null
}

async function saveImage() {
  const canvas = canvasRef.value
  if (!canvas || !originalImage) return

  // If cropping, use crop area; otherwise use full canvas
  let sourceCanvas = canvas
  let sx = 0, sy = 0, sw = canvas.width, sh = canvas.height

  if (cropArea.value && cropArea.value.w > 0 && cropArea.value.h > 0) {
    const ca = cropArea.value
    sx = ca.x * (zoomLevel.value / 100)
    sy = ca.y * (zoomLevel.value / 100)
    sw = ca.w * (zoomLevel.value / 100)
    sh = ca.h * (zoomLevel.value / 100)
  }

  const outCanvas = document.createElement('canvas')
  outCanvas.width = sw
  outCanvas.height = sh
  const outCtx = outCanvas.getContext('2d')
  if (!outCtx) return

  outCtx.drawImage(sourceCanvas, sx, sy, sw, sh, 0, 0, sw, sh)

  const mimeType = outputFormat.value === 'jpeg' ? 'image/jpeg' : `image/${outputFormat.value}`
  const qualityVal = quality.value / 100

  try {
    const blob = await new Promise<Blob | null>((resolve) => {
      outCanvas.toBlob((b) => resolve(b), mimeType, qualityVal)
    })
    if (!blob) {
      ElMessage.error('图片编码失败')
      return
    }

    const buffer = await blob.arrayBuffer()
    const bytes = new Uint8Array(buffer)
    const intArray: number[] = Array.from(bytes)

    // Save via Wails backend
    await SaveFileBytes(props.filePath, intArray)
    ElMessage.success('图片已保存')
  } catch (err: any) {
    ElMessage.error('保存失败: ' + (err?.message || String(err)))
  }
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

onMounted(() => {
  if (props.src) {
    loadImage(props.src)
  } else if (props.filePath) {
    loadImageFromFile(props.filePath)
  }
})

watch(() => props.src, (newSrc) => {
  if (newSrc) loadImage(newSrc)
})

watch(() => props.filePath, (newPath) => {
  if (newPath && !props.src) loadImageFromFile(newPath)
})

async function loadImageFromFile(filePath: string) {
  try {
    const bytes = await ReadFileBytes(filePath)
    const arr = normalizeBytes(bytes)
    if (!arr || arr.length === 0) {
      ElMessage.error('无法读取图片文件')
      return
    }
    const ext = filePath.split('.').pop()?.toLowerCase() || 'png'
    const mimeMap: Record<string, string> = {
      png: 'image/png', jpg: 'image/jpeg', jpeg: 'image/jpeg',
      gif: 'image/gif', webp: 'image/webp', bmp: 'image/bmp',
      svg: 'image/svg+xml', ico: 'image/x-icon',
    }
    const mime = mimeMap[ext] || 'image/png'
    const blob = new Blob([new Uint8Array(arr)], { type: mime })
    const url = URL.createObjectURL(blob)
    loadImage(url)
    imageSize.value = arr.length
  } catch (e: any) {
    ElMessage.error('加载图片失败: ' + (e?.message || String(e)))
  }
}
</script>

<style scoped>
.image-editor {
  --editor-bg: var(--theme-bg, #ffffff);
}
.canvas-area {
  cursor: default;
}
.canvas-area:has(.crop-overlay) {
  cursor: crosshair;
}
</style>
