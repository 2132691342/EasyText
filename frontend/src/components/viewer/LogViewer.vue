<script lang="ts" setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { ReadFile, ReadPartial, StartFileWatch, StopFileWatch } from '../../../wailsjs/go/main/App'
import { ElMessage } from 'element-plus'
import { Play, Pause, Filter, Download } from 'lucide-vue-next'
import { EventsOn } from '../../../wailsjs/runtime/runtime'
import type { EditorTab } from '@/types'

const props = defineProps<{ tab: EditorTab }>()

// ============ 状态 ============
const isTailing = ref(false)
const lines = ref<string[]>([])
const filteredLines = ref<string[]>([])
const scrollContainer = ref<HTMLElement | null>(null)
const lineCount = ref(0)
const loadedCount = ref(0)

// ============ 日志级别过滤 ============
const levelFilter = ref({
  error: true,
  warn: true,
  info: true,
  debug: true,
  trace: true,
})

// ============ 关键字过滤 ============
const keywords = ref<string[]>([])

// 日志级别正则
const LOG_PATTERNS: Record<string, RegExp> = {
  error: /\b(ERROR|FATAL|CRITICAL|ERR|SEVERE|panic)\b/i,
  warn: /\b(WARN|WARNING|WRN)\b/i,
  info: /\b(INFO|INFORMATION|INF)\b/i,
  debug: /\b(DEBUG|DBG|FINE)\b/i,
  trace: /\b(TRACE|FINE|FINER|FINEST)\b/i,
}

function getLogLevel(line: string): string {
  for (const [level, pattern] of Object.entries(LOG_PATTERNS)) {
    if (pattern.test(line)) return level
  }
  return 'unknown'
}

const LEVEL_COLORS: Record<string, string> = {
  error: 'text-red-500 bg-red-50 dark:bg-red-900/20',
  warn: 'text-yellow-600 bg-yellow-50 dark:bg-yellow-900/20',
  info: 'text-green-600 bg-green-50 dark:bg-green-900/20',
  debug: 'text-gray-400 bg-gray-50 dark:bg-gray-800/50',
  trace: 'text-blue-400 bg-blue-50 dark:bg-blue-900/20',
  unknown: 'text-gray-600 dark:text-gray-300',
}

// ============ 分块加载 ============
const CHUNK_SIZE = 500

async function loadContent() {
  if (!props.tab?.path) return
  try {
    const result = await ReadFile(props.tab.path)
    if (result) {
      lines.value = result.content.split('\n')
      lineCount.value = lines.value.length
      loadedCount.value = lines.value.length
      applyFilter()
    }
  } catch (e: any) {
    // 大文件使用分片加载
    if (e?.code === 1003) {
      await loadChunked()
    }
  }
}

async function loadChunked() {
  if (!props.tab?.path) return
  try {
    const chunk = await ReadPartial(props.tab.path, 0, CHUNK_SIZE)
    if (chunk) {
      lines.value = chunk.split('\n')
      loadedCount.value = lines.value.length
      applyFilter()
      loadMoreInBackground()
    }
  } catch (e) {
    console.error('Failed to load log chunk:', e)
  }
}

async function loadMoreInBackground() {
  if (!props.tab?.path) return
  let offset = loadedCount.value
  while (offset < CHUNK_SIZE * 5) {
    try {
      const chunk = await ReadPartial(props.tab.path, offset, CHUNK_SIZE)
      if (!chunk) break
      const newLines = chunk.split('\n')
      lines.value.push(...newLines)
      loadedCount.value = lines.value.length
      applyFilter()
      offset += CHUNK_SIZE
    } catch {
      break
    }
  }
}

function applyFilter() {
  let result = lines.value
  // 级别过滤
  const activeLevels = Object.entries(levelFilter.value).filter(([, v]) => v).map(([k]) => k)
  result = result.filter(line => {
    const level = getLogLevel(line)
    return activeLevels.includes(level)
  })
  // 关键字过滤
  if (keywords.value.length > 0) {
    result = result.filter(line => keywords.value.some(kw => line.toLowerCase().includes(kw.toLowerCase())))
  }
  filteredLines.value = result
}

watch(levelFilter, applyFilter, { deep: true })
watch(keywords, applyFilter, { deep: true })

// ============ 尾部实时刷新 ============
// EventsOn 返回 cancel 函数，startTail 时保存，stopTail 时调用——否则反复
// start/stop 会累积前端事件监听器（isTailing 检查虽能拦截大部分，但监听器本身仍在）。
let fileWatcherCleanup: (() => void) | null = null

async function startTail() {
  if (!props.tab?.path) return
  isTailing.value = true
  try {
    await StartFileWatch(props.tab.path)
    // 重复启动前先清理旧监听器，避免叠加
    fileWatcherCleanup?.()
    fileWatcherCleanup = EventsOn('file:change', async (evt: any) => {
      if (!isTailing.value || evt?.path !== props.tab.path) return
      try {
        const result = await ReadFile(props.tab.path)
        if (result) {
          const newLines = result.content.split('\n')
          // 只追加新增的行
          if (newLines.length > lineCount.value) {
            const added = newLines.slice(lineCount.value)
            lines.value.push(...added)
            lineCount.value = newLines.length
            applyFilter()
            scrollToBottom()
          }
        }
      } catch { /* ignore file watch errors */ }
    })
    ElMessage.success('已开始实时跟踪')
  } catch (e: any) {
    ElMessage.error(`跟踪失败: ${e?.message || ''}`)
  }
}

async function stopTail() {
  isTailing.value = false
  // 先移除前端事件监听器，再停后端文件监视
  fileWatcherCleanup?.()
  fileWatcherCleanup = null
  if (props.tab?.path) {
    try { await StopFileWatch(props.tab.path) } catch { /* ignore */ }
  }
  ElMessage.info('已停止跟踪')
}

function scrollToBottom() {
  nextTick(() => {
    if (scrollContainer.value) {
      scrollContainer.value.scrollTop = scrollContainer.value.scrollHeight
    }
  })
}

import { nextTick } from 'vue'

function toggleLevel(level: string) {
  levelFilter.value[level as keyof typeof levelFilter.value] = !levelFilter.value[level as keyof typeof levelFilter.value]
}

function exportLog() {
  const content = filteredLines.value.join('\n')
  const blob = new Blob([content], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${props.tab.name}_filtered.log`
  a.click()
  URL.revokeObjectURL(url)
}

onMounted(() => {
  loadContent()
})

onUnmounted(() => {
  if (isTailing.value) stopTail()
})
</script>

<template>
  <div class="h-full flex flex-col bg-white dark:bg-[#1e1e1e]">
    <!-- 工具栏 -->
    <div class="flex items-center gap-2 px-3 py-1.5 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#252525]">
      <!-- 跟踪按钮 -->
      <button
        class="flex items-center gap-1 px-2 py-0.5 text-xs rounded border"
        :class="isTailing ? 'border-green-400 text-green-600 bg-green-50 dark:bg-green-900/20' : 'border-gray-300 dark:border-gray-500 text-gray-500'"
        @click="isTailing ? stopTail() : startTail()"
      >
        <Play v-if="!isTailing" class="w-3 h-3" />
        <Pause v-else class="w-3 h-3" />
        {{ isTailing ? '停止跟踪' : '实时跟踪' }}
      </button>

      <div class="w-px h-4 bg-gray-300 dark:bg-gray-600"></div>

      <!-- 级别过滤 -->
      <Filter class="w-3.5 h-3.5 text-gray-400" />
      <button
        v-for="level in ['error', 'warn', 'info', 'debug', 'trace']" :key="level"
        class="px-1.5 py-0.5 text-[10px] rounded border cursor-pointer"
        :class="levelFilter[level as keyof typeof levelFilter] ? LEVEL_COLORS[level] + ' border-current' : 'border-gray-300 dark:border-gray-600 text-gray-400'"
        @click="toggleLevel(level)"
      >
        {{ level.toUpperCase() }}
      </button>

      <div class="flex-1"></div>

      <button class="p-1 text-gray-400 hover:text-gray-600" @click="exportLog" title="导出过滤结果">
        <Download class="w-3.5 h-3.5" />
      </button>
    </div>

    <!-- 日志内容 -->
    <div ref="scrollContainer" class="flex-1 overflow-auto font-mono text-xs leading-5">
      <div
        v-for="(line, idx) in filteredLines" :key="idx"
        class="flex hover:bg-gray-50 dark:hover:bg-[#2a2a2a]"
        :class="idx % 2 === 0 ? '' : 'bg-gray-50/30 dark:bg-[#222]'"
      >
        <span class="w-12 text-right px-2 text-gray-300 dark:text-gray-600 select-none flex-shrink-0 border-r border-gray-100 dark:border-gray-800">{{ idx + 1 }}</span>
        <span
          class="flex-1 px-2 whitespace-pre"
          :class="LEVEL_COLORS[getLogLevel(line)] || 'text-gray-700 dark:text-gray-300'"
        >{{ line }}</span>
      </div>
      <div v-if="filteredLines.length === 0 && lines.length > 0" class="px-4 py-8 text-center text-xs text-gray-400">
        当前过滤条件下无匹配行
      </div>
      <div v-if="lines.length === 0" class="px-4 py-8 text-center text-xs text-gray-400">
        加载中…
      </div>
    </div>

    <!-- 状态栏 -->
    <div class="flex items-center px-3 py-0.5 border-t border-gray-200 dark:border-gray-700 text-[10px] text-gray-400 bg-gray-50 dark:bg-[#252525]">
      <span>总计 {{ lines.length }} 行</span>
      <span v-if="filteredLines.length !== lines.length" class="ml-2">显示 {{ filteredLines.length }} 行</span>
      <span v-if="isTailing" class="ml-2 text-green-500">● 跟踪中</span>
    </div>
  </div>
</template>