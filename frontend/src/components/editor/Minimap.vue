<script lang="ts" setup>
import { ref, watch, onMounted, onUnmounted, nextTick } from 'vue'

/**
 * 文档地图（minimap）：canvas 绘制的代码密度缩略图 + 视口指示器。
 * 点击/拖拽定位到对应行。
 */
const props = defineProps<{
  content: string
  viewport: { scrollTop: number; scrollHeight: number; clientHeight: number }
}>()

const emit = defineEmits<{ (e: 'seek', line: number): void }>()

const canvasRef = ref<HTMLCanvasElement | null>(null)
const WIDTH = 90

function lines(): string[] {
  return props.content.split('\n')
}

function lineColor(text: string, isDark: boolean): string {
  const t = text.trim()
  if (!t) return 'transparent'
  // 注释行 → 绿色调
  if (/^(\/\/|#|--|;|%|\/\*|\*)/.test(t)) return isDark ? '#2f6b3f' : '#b5dcb5'
  // 代码行：按行长加深
  const len = Math.min(text.length, 120)
  const a = 0.3 + (len / 120) * 0.65
  return isDark ? `rgba(150,170,190,${a})` : `rgba(90,110,140,${a})`
}

function draw() {
  const canvas = canvasRef.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  const W = canvas.clientWidth || WIDTH
  const H = canvas.clientHeight
  if (H <= 0) return
  const dpr = window.devicePixelRatio || 1
  if (canvas.width !== Math.round(W * dpr) || canvas.height !== Math.round(H * dpr)) {
    canvas.width = Math.round(W * dpr)
    canvas.height = Math.round(H * dpr)
  }
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
  ctx.clearRect(0, 0, W, H)

  const ls = lines()
  const isDark = document.documentElement.classList.contains('dark')
  const lineH = H / Math.max(ls.length, 1)
  const blockH = Math.max(1, Math.min(2, lineH))
  for (let i = 0; i < ls.length; i++) {
    ctx.fillStyle = lineColor(ls[i], isDark)
    ctx.fillRect(0, i * lineH, W, blockH)
  }

  // 视口指示器
  if (props.viewport.scrollHeight > 0) {
    const vh = Math.max((props.viewport.clientHeight / props.viewport.scrollHeight) * H, 2)
    const vt = (props.viewport.scrollTop / props.viewport.scrollHeight) * H
    ctx.fillStyle = 'rgba(59,130,246,0.28)'
    ctx.fillRect(0, vt, W, vh)
  }
}

function seek(e: MouseEvent) {
  const canvas = canvasRef.value
  if (!canvas) return
  const rect = canvas.getBoundingClientRect()
  if (rect.height <= 0) return
  const ls = lines()
  const line = Math.floor(((e.clientY - rect.top) / rect.height) * ls.length) + 1
  emit('seek', Math.min(Math.max(line, 1), ls.length))
}

let dragging = false
function onMouseDown(e: MouseEvent) { dragging = true; seek(e) }
function onMouseMove(e: MouseEvent) { if (dragging) seek(e) }
function onMouseUp() { dragging = false }

watch(() => [props.content, props.viewport.scrollTop, props.viewport.scrollHeight, props.viewport.clientHeight], draw)

let raf = 0
function onResize() {
  cancelAnimationFrame(raf)
  raf = requestAnimationFrame(draw)
}

onMounted(() => {
  nextTick(draw)
  window.addEventListener('resize', onResize)
})
onUnmounted(() => {
  window.removeEventListener('resize', onResize)
  cancelAnimationFrame(raf)
})
</script>

<template>
  <canvas
    ref="canvasRef"
    class="minimap"
    @mousedown="onMouseDown"
    @mousemove="onMouseMove"
    @mouseup="onMouseUp"
    @mouseleave="onMouseUp"
  />
</template>

<style scoped>
.minimap {
  width: 90px;
  height: 100%;
  flex-shrink: 0;
  border-left: 1px solid #e5e7eb;
  cursor: pointer;
  display: block;
}
html.dark .minimap { border-left-color: #333; }
</style>
