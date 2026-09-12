<template>
  <div class="color-picker">
    <!-- Color swatch -->
    <div class="color-swatch" :style="{ backgroundColor: currentColor }" />

    <!-- Pick button -->
    <ElButton type="primary" class="color-pick-btn" @click="pickColor">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6">
        <path d="M3 17L17 3l4 4L7 21l-4-4z" /><path d="M13 7l4 4" />
      </svg>
      <template v-if="eyeDropperSupported">屏幕取色</template>
      <template v-else>输入颜色</template>
    </ElButton>

    <!-- Hex input -->
    <div class="color-row">
      <span class="color-row-label">HEX</span>
      <ElInput v-model="hexColor" size="small" @input="onHexInput" placeholder="#RRGGBB" />
      <ElButton size="small" @click="copyToClipboard(hexColor)">
        <Copy :size="12" :stroke-width="1.6" />
      </ElButton>
    </div>

    <!-- RGB display -->
    <div class="color-row">
      <span class="color-row-label">RGB</span>
      <span class="color-mono">{{ rgbStr }}</span>
      <ElButton size="small" @click="copyToClipboard(rgbStr)">
        <Copy :size="12" :stroke-width="1.6" />
      </ElButton>
    </div>

    <!-- HSL display -->
    <div class="color-row">
      <span class="color-row-label">HSL</span>
      <span class="color-mono">{{ hslStr }}</span>
    </div>

    <!-- Color format toggle -->
    <ElRadioGroup v-model="colorFormat" size="small" class="color-fmt">
      <ElRadioButton value="hex">HEX</ElRadioButton>
      <ElRadioButton value="rgb">RGB</ElRadioButton>
      <ElRadioButton value="hsl">HSL</ElRadioButton>
    </ElRadioGroup>

    <!-- History -->
    <div v-if="history.length > 0" class="color-history">
      <p class="color-history-title">历史记录</p>
      <div class="color-history-grid">
        <div v-for="(c, i) in history" :key="i"
             class="color-history-swatch"
             :style="{ backgroundColor: c }"
             @click="selectColor(c)"
             :title="c" />
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
  history.value = [color, ...history.value.filter(c => c !== color)].slice(0, 10)
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success('已复制到剪贴板')
  })
}
</script>

<style scoped>
.color-picker {
  display: flex;
  flex-direction: column;
  gap: var(--et-space-2);
  color: var(--et-fg);
}
.color-swatch {
  width: 100%;
  height: 80px;
  border-radius: var(--et-radius);
  border: 1px solid var(--et-border);
}
.color-pick-btn {
  width: 100%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--et-space-1);
}
.color-row {
  display: flex;
  align-items: center;
  gap: var(--et-space-2);
}
.color-row-label {
  width: 32px;
  font-size: var(--et-text-xs);
  color: var(--et-fg-subtle);
  flex-shrink: 0;
}
.color-mono {
  font-family: var(--editor-font-family, 'Consolas', monospace);
  font-size: var(--et-text-md);
  color: var(--et-fg);
  flex: 1;
}
.color-fmt {
  align-self: flex-start;
}
.color-history {
  border-top: 1px solid var(--et-border);
  padding-top: var(--et-space-2);
}
.color-history-title {
  font-size: var(--et-text-xs);
  color: var(--et-fg-subtle);
  margin: 0 0 var(--et-space-1);
}
.color-history-grid {
  display: flex;
  flex-wrap: wrap;
  gap: var(--et-space-1);
}
.color-history-swatch {
  width: 24px;
  height: 24px;
  border-radius: var(--et-radius-sm);
  border: 1px solid var(--et-border);
  cursor: pointer;
  transition: transform 80ms ease;
}
.color-history-swatch:hover {
  transform: scale(1.1);
}
</style>
