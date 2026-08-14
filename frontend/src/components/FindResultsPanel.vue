<script lang="ts" setup>
import { ref, computed, watch } from 'vue'
import { useEditorStore } from '@/stores'
import { X, ChevronUp, ChevronDown, Trash2, Copy } from 'lucide-vue-next'
import { ElMessage } from 'element-plus'

export interface FindResult {
  file: string
  line: number
  column?: number
  content: string
  match?: string
}

const props = defineProps<{
  visible: boolean
  results: FindResult[]
  title?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'clear'): void
  (e: 'goto', result: FindResult): void
}>()

const editorStore = useEditorStore()
const panelHeight = ref(150)
const isResizing = ref(false)

function handleResultClick(result: FindResult) {
  emit('goto', result)
}

function handleClear() {
  emit('clear')
}

function copyResults() {
  const text = props.results.map(r => `${r.file}:${r.line}: ${r.content}`).join('\n')
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success('结果已复制到剪贴板')
  })
}

// Resize handling
function startResize(e: MouseEvent) {
  isResizing.value = true
  document.addEventListener('mousemove', handleResize)
  document.addEventListener('mouseup', stopResize)
  e.preventDefault()
}

function handleResize(e: MouseEvent) {
  if (!isResizing.value) return
  const container = document.querySelector('.find-results-panel')
  if (!container) return
  const rect = container.getBoundingClientRect()
  const newHeight = rect.bottom - e.clientY
  panelHeight.value = Math.max(80, Math.min(500, newHeight))
}

function stopResize() {
  isResizing.value = false
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('mouseup', stopResize)
}

// Collapse/expand
const collapsed = ref(false)
function toggleCollapse() {
  collapsed.value = !collapsed.value
}
</script>

<template>
  <div
    v-if="visible"
    class="find-results-panel flex flex-col border-t border-gray-200 dark:border-gray-700 bg-white dark:bg-[#1e1e1e]"
    :style="{ height: collapsed ? '28px' : `${panelHeight}px` }"
  >
    <!-- Resize handle -->
    <div
      class="h-1 bg-transparent hover:bg-blue-400 cursor-ns-resize flex-shrink-0"
      @mousedown="startResize"
    ></div>

    <!-- Header -->
    <div class="flex items-center h-7 px-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#2d2d2d] flex-shrink-0">
      <button class="p-0.5 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-500" @click="toggleCollapse">
        <ChevronDown v-if="!collapsed" class="w-3.5 h-3.5" />
        <ChevronUp v-else class="w-3.5 h-3.5" />
      </button>
      <span class="text-xs font-medium ml-1 text-gray-600 dark:text-gray-300">{{ title || '查找结果' }} ({{ results.length }})</span>
      <div class="ml-auto flex items-center gap-1">
        <button 
          class="p-0.5 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-500" 
          title="复制结果" 
          @click="copyResults"
          :disabled="results.length === 0"
        >
          <Copy class="w-3.5 h-3.5" />
        </button>
        <button 
          class="p-0.5 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-500" 
          title="清除" 
          @click="handleClear"
        >
          <Trash2 class="w-3.5 h-3.5" />
        </button>
        <button 
          class="p-0.5 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-500" 
          title="关闭" 
          @click="emit('close')"
        >
          <X class="w-3.5 h-3.5" />
        </button>
      </div>
    </div>

    <!-- Results list -->
    <div v-if="!collapsed" class="flex-1 overflow-auto">
      <div v-if="results.length === 0" class="flex items-center justify-center h-full text-xs text-gray-400">
        没有匹配结果
      </div>
      <div
        v-for="(result, idx) in results"
        :key="idx"
        class="find-result-item flex items-center px-3 py-1 cursor-pointer hover:bg-blue-50 dark:hover:bg-[#094771] border-b border-gray-100 dark:border-gray-800"
        @click="handleResultClick(result)"
        @dblclick="handleResultClick(result)"
      >
        <span class="text-xs font-mono text-gray-500 dark:text-gray-400 mr-2 flex-shrink-0">
          {{ result.file.split(/[/\\]/).pop() }}:{{ result.line }}
        </span>
        <span class="text-xs truncate text-gray-700 dark:text-gray-300" v-html="result.content.substring(0, 200)"></span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.find-results-panel {
  min-height: 28px;
  user-select: none;
}
.find-result-item:last-child {
  border-bottom: none;
}
</style>
