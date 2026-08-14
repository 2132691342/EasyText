<script lang="ts" setup>
import { ref, computed, watch, onMounted } from 'vue'
import type { EditorTab } from '@/types'
import { ReadFileBytes } from '../../../wailsjs/go/main/App'
import { ChevronLeft, ChevronRight, ArrowRightToLine, Copy, RefreshCw } from 'lucide-vue-next'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  tab: EditorTab
}>()

// State
const hexData = ref<number[]>([])
const loading = ref(false)
const pageSize = 64 * 16 // 64 lines * 16 bytes per line = 1024 bytes per page
const currentPage = ref(1)
const totalPages = ref(1)
const showAscii = ref(true)

// Computed
const hexRows = computed(() => {
  const data = currentPageData.value
  const rows: { offset: number; hex: string[]; ascii: string }[] = []
  for (let i = 0; i < data.length; i += 16) {
    const chunk = data.slice(i, i + 16)
    const offset = (currentPage.value - 1) * pageSize + i
    const hex = chunk.map(b => b.toString(16).padStart(2, '0').toUpperCase())
    const ascii = chunk.map(b => (b >= 32 && b <= 126) ? String.fromCharCode(b) : '.').join('')
    rows.push({ offset, hex, ascii })
  }
  return rows
})

const currentPageData = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return hexData.value.slice(start, start + pageSize)
})

// Load data
async function loadHexData() {
  if (!props.tab.path) return
  loading.value = true
  try {
    const bytes = await ReadFileBytes(props.tab.path)
    hexData.value = bytes
    totalPages.value = Math.ceil(bytes.length / pageSize)
    currentPage.value = 1
  } catch (e) {
    ElMessage.error('读取文件失败')
  } finally {
    loading.value = false
  }
}

// Navigation
function prevPage() {
  if (currentPage.value > 1) currentPage.value--
}

function nextPage() {
  if (currentPage.value < totalPages.value) currentPage.value++
}

function gotoPage() {
  const input = prompt(`转到页 (1-${totalPages.value}):`, String(currentPage.value))
  if (input) {
    const page = parseInt(input)
    if (page >= 1 && page <= totalPages.value) {
      currentPage.value = page
    }
  }
}

// Copy hex or ascii
function copyHex() {
  const text = hexRows.value.map(r => r.hex.join(' ')).join('\n')
  navigator.clipboard.writeText(text).then(() => ElMessage.success('十六进制数据已复制'))
}

function copyAscii() {
  const text = hexRows.value.map(r => r.ascii).join('\n')
  navigator.clipboard.writeText(text).then(() => ElMessage.success('ASCII 文本已复制'))
}

// Format
function formatOffset(offset: number): string {
  return offset.toString(16).padStart(8, '0').toUpperCase()
}

onMounted(() => {
  loadHexData()
})
</script>

<template>
  <div class="h-full flex flex-col bg-white dark:bg-[#1e1e1e]">
    <!-- Toolbar -->
    <div class="flex items-center h-7 px-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#2d2d2d] gap-2 flex-shrink-0">
      <span class="text-xs text-gray-500 dark:text-gray-400">十六进制查看: {{ tab.name }}</span>
      <div class="w-px h-4 bg-gray-300 dark:bg-gray-600"></div>
      <button class="hex-btn" @click="prevPage" :disabled="currentPage <= 1" title="上一页">
        <ChevronLeft class="w-3.5 h-3.5" />
      </button>
      <span class="text-xs text-gray-500 dark:text-gray-400">页 {{ currentPage }} / {{ totalPages }}</span>
      <button class="hex-btn" @click="nextPage" :disabled="currentPage >= totalPages" title="下一页">
        <ChevronRight class="w-3.5 h-3.5" />
      </button>
      <button class="hex-btn" @click="gotoPage" title="转到页">
        <ArrowRightToLine class="w-3.5 h-3.5" />
      </button>
      <div class="w-px h-4 bg-gray-300 dark:bg-gray-600"></div>
      <button class="hex-btn" @click="loadHexData" title="刷新">
        <RefreshCw class="w-3.5 h-3.5" />
      </button>
      <div class="w-px h-4 bg-gray-300 dark:bg-gray-600"></div>
      <button class="hex-btn" @click="copyHex" title="复制十六进制">
        <Copy class="w-3.5 h-3.5" />
      </button>
      <span class="text-xs text-gray-500 dark:text-gray-400 ml-auto">
        {{ hexData.length.toLocaleString() }} 字节
      </span>
    </div>

    <!-- Hex content -->
    <div v-if="loading" class="flex-1 flex items-center justify-center text-xs text-gray-400">
      加载中...
    </div>
    <div v-else class="flex-1 overflow-auto font-mono text-xs">
      <div
        v-for="(row, idx) in hexRows"
        :key="idx"
        class="flex items-center hover:bg-blue-50 dark:hover:bg-[#094771] border-b border-gray-50 dark:border-gray-800"
      >
        <!-- Offset -->
        <span class="hex-offset text-gray-400 dark:text-gray-500 px-3 py-0.5 select-none">{{ formatOffset(row.offset) }}</span>
        <span class="text-gray-300 dark:text-gray-600 select-none">│</span>
        <!-- Hex bytes -->
        <span class="hex-bytes px-2 py-0.5">
          <span
            v-for="(b, bi) in row.hex"
            :key="bi"
            class="hex-byte"
            :class="bi === 7 ? 'mr-3' : ''"
          >{{ b }}</span>
        </span>
        <span class="text-gray-300 dark:text-gray-600 select-none">│</span>
        <!-- ASCII -->
        <span class="hex-ascii px-3 py-0.5 text-gray-500 dark:text-gray-400">{{ row.ascii }}</span>
      </div>
    </div>

    <!-- Footer -->
    <div class="flex items-center h-6 px-2 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#2d2d2d] text-xs text-gray-400 flex-shrink-0">
      <span>页大小: {{ pageSize }} 字节 | 偏移: {{ formatOffset((currentPage - 1) * pageSize) }} - {{ formatOffset(Math.min(currentPage * pageSize - 1, hexData.length)) }}</span>
    </div>
  </div>
</template>

<style scoped>
.hex-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 2px 4px;
  border: none;
  background: transparent;
  color: rgb(107, 114, 128);
  border-radius: 3px;
  cursor: pointer;
}
.hex-btn:hover:not(:disabled) {
  background: rgba(59, 130, 246, 0.1);
  color: rgb(59, 130, 246);
}
.hex-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.hex-offset, .hex-ascii {
  white-space: pre;
  line-height: 1.6;
}
.hex-byte {
  display: inline-block;
  width: 24px;
  text-align: center;
  color: #1a73e8;
}
html.dark .hex-byte {
  color: #7aa2f7;
}
</style>
