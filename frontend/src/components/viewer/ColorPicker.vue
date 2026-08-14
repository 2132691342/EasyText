<template>
  <div class="color-picker p-4 bg-[var(--theme-bg)] text-[var(--theme-fg)] rounded-lg" style="width: 280px">
    <!-- Color swatch -->
    <div class="color-swatch w-full h-20 rounded-lg mb-3 border" :style="{ backgroundColor: currentColor }" />

    <!-- Pick button -->
    <ElButton type="primary" class="w-full mb-3" @click="pickColor">
      <svg class="w-4 h-4 mr-1 inline" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M3 17L17 3l4 4L7 21l-4-4z" /><path d="M13 7l4 4" />
      </svg>
      <template v-if="eyeDropperSupported">屏幕取色</template>
      <template v-else>输入颜色</template>
    </ElButton>

    <!-- Hex input -->
    <div class="flex items-center gap-2 mb-2">
      <span class="text-xs text-[var(--theme-comment)] w-8">HEX</span>
      <ElInput v-model="hexColor" size="small" @input="onHexInput" placeholder="#RRGGBB" />
      <ElButton size="small" @click="copyToClipboard(hexColor)">
        <Copy class="w-3 h-3" />
      </ElButton>
    </div>

    <!-- RGB display -->
    <div class="flex items-center gap-2 mb-2">
      <span class="text-xs text-[var(--theme-comment)] w-8">RGB</span>
      <span class="font-mono text-sm">{{ rgbStr }}</span>
      <ElButton size="small" @click="copyToClipboard(rgbStr)">
        <Copy class="w-3 h-3" />
      </ElButton>
    </div>

    <!-- HSL display -->
    <div class="flex items-center gap-2 mb-3">
      <span class="text-xs text-[var(--theme-comment)] w-8">HSL</span>
      <span class="font-mono text-sm">{{ hslStr }}</span>
    </div>

    <!-- Color format toggle -->
    <ElRadioGroup v-model="colorFormat" size="small" class="mb-3">
      <ElRadioButton value="hex">HEX</ElRadioButton>
      <ElRadioButton value="rgb">RGB</ElRadioButton>
      <ElRadioButton value="hsl">HSL</ElRadioButton>
    </ElRadioGroup>

    <!-- History -->
    <div v-if="history.length > 0">
      <p class="text-xs text-[var(--theme-comment)] mb-1">历史记录</p>
      <div class="flex flex-wrap gap-1">
        <div v-for="(c, i) in history" :key="i"
             class="w-6 h-6 rounded cursor-pointer border hover:scale-110 transition-transform"
             :style="{ backgroundColor: c }"
             @click="selectColor(c)" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElButton, ElInput, ElRadioGroup, ElRadioButton, ElMessage } from 'element-plus'
import { Copy } from 'lucide-vue-next'

const emit = defineEmits<{
  select: [color: string]
  close: []
}>()

const currentColor = ref('#ffffff')
const hexColor = ref('#ffffff')
const colorFormat = ref('hex')
const history = ref<string[]>([])
const eyeDropperSupported = ref(typeof window !== 'undefined' && 'EyeDropper' in window)

const rgbStr = computed(() => {
  const r = parseInt(currentColor.value.slice(1, 3), 16)
  const g = parseInt(currentColor.value.slice(3, 5), 16)
  const b = parseInt(currentColor.value.slice(5, 7), 16)
  return `rgb(${r}, ${g}, ${b})`
})

const hslStr = computed(() => {
  const r = parseInt(currentColor.value.slice(1, 3), 16) / 255
  const g = parseInt(currentColor.value.slice(3, 5), 16) / 255
  const b = parseInt(currentColor.value.slice(5, 7), 16) / 255
  const max = Math.max(r, g, b), min = Math.min(r, g, b)
  let h = 0, s = 0, l = (max + min) / 2
  if (max !== min) {
    const d = max - min
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min)
    switch (max) {
      case r: h = ((g - b) / d + (g < b ? 6 : 0)) / 6; break
      case g: h = ((b - r) / d + 2) / 6; break
      case b: h = ((r - g) / d + 4) / 6; break
    }
  }
  return `hsl(${Math.round(h * 360)}, ${Math.round(s * 100)}%, ${Math.round(l * 100)}%)`
})

async function pickColor() {
  if (!eyeDropperSupported.value) return

  try {
    const EyeDropperClass = window.EyeDropper
    if (!EyeDropperClass) throw new Error('EyeDropper unavailable')
    const eyeDropper = new EyeDropperClass()
    const result = await eyeDropper.open()
    selectColor(result.sRGBHex)
  } catch (err: any) {
    if (err?.name !== 'AbortError') {
      console.error('EyeDropper error:', err)
    }
  }
}

function onHexInput(val: string) {
  if (/^#[0-9a-fA-F]{6}$/.test(val)) {
    selectColor(val)
  }
}

function selectColor(color: string) {
  currentColor.value = color
  hexColor.value = color
  emit('select', color)

  // Add to history
  history.value = [color, ...history.value.filter(c => c !== color)].slice(0, 10)
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success('已复制到剪贴板')
  })
}
</script>
