<script lang="ts" setup>
/**
 * 日志查看器
 *
 * 本轮修复：
 *  1. 大文件：此前无论多大都先 ReadFile 全量读入（>100MB 才回落分块），
 *     且分块加载有硬上限 2500 行、没有「加载更多」入口 —— 后半段日志根本看不到。
 *     现改为：先读文件信息判断体积，小文件全量、大文件按 500 行分块，
 *     底部提供「加载更多」并显示 已加载/总行数（总行数取自 FileInfo.lineCount）。
 *  2. 切换标签不重载：相同组件实例复用，加 path 监听。
 *  3. 关键字过滤：逻辑存在但界面没有输入框，属于死代码，补上 UI。
 *  4. 导出：a[download] 在 WebView2 中不可靠，改用后端另存为对话框。
 *  5. 样式令牌化。
 */
import { ref, computed, nextTick, onMounted, onUnmounted, watch } from 'vue'
import { ReadFile, ReadPartial, GetFileInfo, SaveFile, SaveFileDialog, StartFileWatch, StopFileWatch } from '../../../wailsjs/go/main/App'
import { ElMessage } from 'element-plus'
import { Play, Pause, Filter, Download, ChevronDown, Search } from 'lucide-vue-next'
import { EventsOn } from '../../../wailsjs/runtime/runtime'
import type { EditorTab } from '@/types'

const props = defineProps<{ tab: EditorTab }>()

// ============ 状态 ============
const isTailing = ref(false)
const lines = ref<string[]>([])
const filteredLines = ref<string[]>([])
const scrollContainer = ref<HTMLElement | null>(null)
const loadedCount = ref(0)
const totalLines = ref(0)
const loadingMore = ref(false)

/** 超过该体积走分块加载，避免一次性把几百 MB 日志读进内存 */
const MAX_INLINE_BYTES = 8 * 1024 * 1024
const CHUNK_SIZE = 500

const hasMore = computed(() => loadedCount.value < totalLines.value)

// ============ 级别过滤 ============
const levelFilter = ref({
  error: true,
  warn: true,
  info: true,
  debug: true,
  trace: true,
})

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

// 级别只用于行前缀着色，配色交给 CSS 变量
const LEVEL_CLASS: Record<string, string> = {
  error: 'lv-error',
  warn: 'lv-warn',
  info: 'lv-info',
  debug: 'lv-debug',
  trace: 'lv-trace',
  unknown: 'lv-unknown',
}

// ============ 关键字过滤 ============
const keywordsInput = ref('')
const keywords = computed(() =>
  keywordsInput.value.split(/\s+/).map(s => s.trim()).filter(Boolean),
)

// ============ 加载 ============
async function loadContent() {
  const path = props.tab?.path
  if (!path) return
  lines.value = []
  loadedCount.value = 0
  totalLines.value = 0
  try {
    const info = await GetFileInfo(path)
    const size = Number(info?.size ?? 0)
    totalLines.value = Number(info?.lineCount ?? 0)
    if (size > MAX_INLINE_BYTES) {
      await loadChunk(0)
      ElMessage.info('日志较大，已按分块加载，可点击「加载更多」继续')
    } else {
      const result = await ReadFile(path)
      if (result) {
        lines.value = result.content.split('\n')
        loadedCount.value = lines.value.length
        totalLines.value = Math.max(totalLines.value, lines.value.length)
      }
    }
  } catch {
    // 信息读取失败时退回全量读取（老路径，保证可用）
    try {
      const result = await ReadFile(path)
      if (result) {
        lines.value = result.content.split('\n')
        loadedCount.value = lines.value.length
        totalLines.value = lines.value.length
      }
    } catch (e: unknown) {
      const code = (e as { code?: number })?.code
      if (code === 1003) {
        await loadChunk(0)
      } else {
        ElMessage.error('日志读取失败')
      }
    }
  }
}

/** 从 from 行开始读取一块（ReadPartial 按完整行返回，追加时不会截断半行） */
async function loadChunk(from: number): Promise<number> {
  const path = props.tab?.path
  if (!path) return 0
  const chunk = await ReadPartial(path, from, CHUNK_SIZE)
  const newLines = chunk ? chunk.split('\n') : []
  if (from === 0) lines.value = newLines
  else lines.value.push(...newLines)
  loadedCount.value = lines.value.length
  if (totalLines.value < loadedCount.value) totalLines.value = loadedCount.value
  return newLines.length
}

async function loadMore() {
  if (loadingMore.value || !hasMore.value) return
  loadingMore.value = true
  try {
    const n = await loadChunk(loadedCount.value)
    if (n === 0) totalLines.value = loadedCount.value // 已到文件末尾
  } catch {
    ElMessage.error('加载更多失败')
  } finally {
    loadingMore.value = false
  }
}

function applyFilter() {
  let result = lines.value
  const activeLevels = Object.entries(levelFilter.value).filter(([, v]) => v).map(([k]) => k)
  result = result.filter(line => activeLevels.includes(getLogLevel(line)))
  const kws = keywords.value
  if (kws.length > 0) {
    result = result.filter(line => kws.some(kw => line.toLowerCase().includes(kw.toLowerCase())))
  }
  filteredLines.value = result
}

watch([levelFilter, keywords], applyFilter, { deep: true })

// ============ 尾部实时刷新 ============
// EventsOn 返回 cancel 函数，startTail 时保存、stopTail 时调用，
// 否则反复 start/stop 会累积前端事件监听器。
let fileWatcherCleanup: (() => void) | null = null

async function startTail() {
  const path = props.tab?.path
  if (!path) return
  isTailing.value = true
  try {
    await StartFileWatch(path)
    fileWatcherCleanup?.()
    fileWatcherCleanup = EventsOn('file:change', async (evt: any) => {
      if (!isTailing.value || evt?.path !== path) return
      try {
        const result = await ReadFile(path)
        if (result) {
          const newLines = result.content.split('\n')
          if (newLines.length > loadedCount.value) {
            lines.value.push(...newLines.slice(loadedCount.value))
            loadedCount.value = newLines.length
            totalLines.value = Math.max(totalLines.value, newLines.length)
            applyFilter()
            scrollToBottom()
          }
        }
      } catch { /* 文件可能被删除或独占，忽略本次刷新 */ }
    })
    ElMessage.success('已开始实时跟踪')
  } catch (e: any) {
    ElMessage.error(`跟踪失败: ${e?.message || ''}`)
  }
}

async function stopTail() {
  isTailing.value = false
  fileWatcherCleanup?.()
  fileWatcherCleanup = null
  if (props.tab?.path) {
    try { await StopFileWatch(props.tab.path) } catch { /* 已停止 */ }
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

function toggleLevel(level: string) {
  levelFilter.value[level as keyof typeof levelFilter.value] = !levelFilter.value[level as keyof typeof levelFilter.value]
}

/** 导出过滤结果：走后端保存对话框，避免 WebView2 忽略 a[download] */
async function exportLog() {
  if (filteredLines.value.length === 0) {
    ElMessage.warning('当前没有可导出的内容')
    return
  }
  try {
    const path = await SaveFileDialog(`${props.tab.name}_filtered.log`)
    if (!path) return
    await SaveFile(path, filteredLines.value.join('\n'), 'UTF-8')
    ElMessage.success('已导出过滤结果')
  } catch {
    ElMessage.error('导出失败')
  }
}

watch(() => props.tab?.path, (p, old) => {
  if (p && p !== old) {
    if (isTailing.value) stopTail()
    loadContent()
  }
})

onMounted(loadContent)

onUnmounted(() => {
  if (isTailing.value) stopTail()
})
</script>

<template>
  <div class="log-view">
    <!-- 工具栏 -->
    <div class="log-bar">
      <button
        class="log-tail-btn"
        :class="{ 'is-on': isTailing }"
        @click="isTailing ? stopTail() : startTail()"
      >
        <Play v-if="!isTailing" :size="12" />
        <Pause v-else :size="12" />
        {{ isTailing ? '停止跟踪' : '实时跟踪' }}
      </button>

      <span class="log-sep" />

      <Filter :size="13" class="log-icon" />
      <button
        v-for="level in ['error', 'warn', 'info', 'debug', 'trace']"
        :key="level"
        class="log-lv-btn"
        :class="[`lv-${level}`, levelFilter[level as keyof typeof levelFilter] ? 'is-on' : 'is-off']"
        @click="toggleLevel(level)"
      >
        {{ level.toUpperCase() }}
      </button>

      <span class="log-sep" />

      <Search :size="13" class="log-icon" />
      <input
        v-model="keywordsInput"
        class="log-search"
        placeholder="关键字过滤（空格分隔多个）"
      />

      <span class="log-spacer" />

      <button class="log-icon-btn" title="导出过滤结果" @click="exportLog">
        <Download :size="14" />
      </button>
    </div>

    <!-- 日志内容 -->
    <div ref="scrollContainer" class="log-body">
      <div
        v-for="(line, idx) in filteredLines" :key="idx"
        class="log-row"
      >
        <span class="log-no">{{ idx + 1 }}</span>
        <span class="log-text" :class="LEVEL_CLASS[getLogLevel(line)]">{{ line }}</span>
      </div>

      <div v-if="filteredLines.length === 0 && lines.length > 0" class="log-empty">
        当前过滤条件下无匹配行
      </div>
      <div v-if="lines.length === 0" class="log-empty">加载中…</div>

      <!-- 分块加载：此前 2500 行之后的内容没有任何入口可以看到 -->
      <div v-if="hasMore" class="log-more">
        <button class="log-more-btn" :disabled="loadingMore" @click="loadMore">
          <ChevronDown :size="13" />
          {{ loadingMore ? '加载中…' : `加载更多（剩余 ${Math.max(0, totalLines - loadedCount)} 行）` }}
        </button>
      </div>
    </div>

    <!-- 状态栏 -->
    <div class="log-foot">
      <span>总计 {{ totalLines || lines.length }} 行</span>
      <span v-if="filteredLines.length !== lines.length">显示 {{ filteredLines.length }} 行</span>
      <span v-if="isTailing" class="log-tailing">● 跟踪中</span>
    </div>
  </div>
</template>

<style scoped>
/* 令牌化：此前整块使用 Tailwind gray/red/green + dark: 变体字面色值 */
.log-view {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--et-bg);
  color: var(--et-fg);
}

.log-bar {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  flex-shrink: 0;
  background: var(--et-bg-sunken);
  border-bottom: 1px solid var(--et-border);
  font-size: 12px;
}
.log-sep { width: 1px; height: 14px; background: var(--et-border); margin: 0 4px; flex-shrink: 0; }
.log-spacer { flex: 1 1 auto; }
.log-icon { color: var(--et-fg-subtle); flex-shrink: 0; }

.log-tail-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  font-size: 11px;
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius-sm);
  background: var(--et-bg);
  color: var(--et-fg-muted);
  cursor: pointer;
}
.log-tail-btn:hover { background: var(--et-bg-hover); color: var(--et-fg); }
.log-tail-btn.is-on { border-color: #22c55e; color: #22c55e; background: rgba(34, 197, 94, .1); }

.log-lv-btn {
  padding: 2px 6px;
  font-size: 10px;
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius-sm);
  background: transparent;
  cursor: pointer;
  transition: opacity .12s ease;
}
.log-lv-btn.is-off { opacity: .45; color: var(--et-fg-subtle); }

.log-search {
  width: 200px;
  padding: 2px 6px;
  font-size: 11px;
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius-sm);
  background: var(--et-bg);
  color: var(--et-fg);
  outline: none;
}
.log-search:focus { border-color: var(--et-accent); }
.log-search::placeholder { color: var(--et-fg-subtle); }

.log-icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 22px;
  border: none;
  border-radius: var(--et-radius-sm);
  background: transparent;
  color: var(--et-fg-muted);
  cursor: pointer;
}
.log-icon-btn:hover { background: var(--et-bg-hover); color: var(--et-fg); }

.log-body { flex: 1; overflow: auto; font-family: 'Cascadia Mono', Consolas, monospace; font-size: 12px; line-height: 1.5; }
.log-row { display: flex; }
.log-row:hover { background: var(--et-bg-hover); }
.log-no {
  width: 56px;
  padding: 0 8px;
  text-align: right;
  flex-shrink: 0;
  user-select: none;
  color: var(--et-fg-subtle);
  border-right: 1px solid var(--et-border);
}
.log-text { flex: 1; padding: 0 8px; white-space: pre; }

.lv-error { color: #ef4444; }
.lv-warn { color: #d97706; }
.lv-info { color: #16a34a; }
.lv-debug { color: var(--et-fg-subtle); }
.lv-trace { color: #3b82f6; }
.lv-unknown { color: var(--et-fg); }

.log-empty { padding: 32px 16px; text-align: center; font-size: 12px; color: var(--et-fg-subtle); }

.log-more { padding: 8px; text-align: center; }
.log-more-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  font-size: 12px;
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius-sm);
  background: var(--et-bg);
  color: var(--et-fg-muted);
  cursor: pointer;
}
.log-more-btn:hover:not(:disabled) { background: var(--et-bg-hover); color: var(--et-fg); }
.log-more-btn:disabled { opacity: .5; cursor: default; }

.log-foot {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 3px 8px;
  flex-shrink: 0;
  font-size: 11px;
  color: var(--et-fg-subtle);
  background: var(--et-bg-sunken);
  border-top: 1px solid var(--et-border);
}
.log-tailing { color: #22c55e; }
</style>
