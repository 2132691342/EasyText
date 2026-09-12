<script lang="ts" setup>
/**
 * 十六进制查看器
 *
 *  - 按页调用 ReadFileChunk（ReadAt 随机访问），内存占用恒定为一页
 *  - 切换标签时重新加载当前文件
 *  - 翻页命令（pre/next/goto-hex-page）经 editor-command 接线
 *  - 转到页使用 Element Plus 弹窗（WebView2 下 window.prompt 不可靠）
 */
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import type { EditorTab } from '@/types'
import { ReadFileChunk, ReadFileBytes } from '../../../wailsjs/go/main/App'
import { ChevronLeft, ChevronRight, ArrowRightToLine, Copy, RefreshCw, FileText } from 'lucide-vue-next'
import { ElMessage, ElMessageBox } from 'element-plus'

const props = defineProps<{
  tab: EditorTab
}>()

const pageSize = 64 * 16 // 64 行 × 16 字节 = 1024 字节/页
const buffer = ref<number[]>([])
const totalBytes = ref(0)
const currentPage = ref(1)
const loading = ref(false)

const totalPages = computed(() => Math.max(1, Math.ceil(totalBytes.value / pageSize)))

const hexRows = computed(() => {
  const rows: { offset: number; hex: string[]; ascii: string }[] = []
  for (let i = 0; i < buffer.value.length; i += 16) {
    const chunk = buffer.value.slice(i, i + 16)
    const offset = (currentPage.value - 1) * pageSize + i
    const hex = chunk.map(b => b.toString(16).padStart(2, '0').toUpperCase())
    const ascii = chunk.map(b => (b >= 32 && b <= 126) ? String.fromCharCode(b) : '.').join('')
    rows.push({ offset, hex, ascii })
  }
  return rows
})

/** 降级模式：绑定不可用时退回整文件读取（功能可用但占内存） */
const fallback = ref(false)
let fallbackBytes: number[] | null = null

/** 加载指定页（数据不进内存常驻，只保留当前页） */
async function loadPage(page: number) {
  const path = props.tab.path
  if (!path) return
  const target = Math.min(Math.max(1, page), totalPages.value)
  loading.value = true
  try {
    if (!fallback.value) {
      try {
        const res = await ReadFileChunk(path, (target - 1) * pageSize, pageSize)
        buffer.value = (res?.data ?? []) as unknown as number[]
        totalBytes.value = Number(res?.total ?? 0)
        currentPage.value = target
        return
      } catch (e) {
        // 分片接口不可用（例如绑定未注册）时自动降级，保证十六进制视图仍可查看
        console.warn('ReadFileChunk unavailable, falling back to full read:', e)
        fallback.value = true
      }
    }
    if (!fallbackBytes) {
      fallbackBytes = (await ReadFileBytes(path)) as unknown as number[]
    }
    totalBytes.value = fallbackBytes.length
    buffer.value = fallbackBytes.slice((target - 1) * pageSize, target * pageSize)
    currentPage.value = target
  } catch {
    ElMessage.error('读取文件失败')
    buffer.value = []
  } finally {
    loading.value = false
  }
}

function prevPage() {
  if (currentPage.value > 1) loadPage(currentPage.value - 1)
}

function nextPage() {
  if (currentPage.value < totalPages.value) loadPage(currentPage.value + 1)
}

async function gotoPage() {
  try {
    const { value } = await ElMessageBox.prompt(
      `转到页 (1 - ${totalPages.value})`,
      '转到页',
      {
        inputValue: String(currentPage.value),
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputPattern: /^\d+$/,
        inputErrorMessage: '请输入页码数字',
      },
    )
    const page = parseInt(value, 10)
    if (page >= 1 && page <= totalPages.value) await loadPage(page)
    else ElMessage.warning(`页码范围 1 - ${totalPages.value}`)
  } catch { /* 用户取消 */ }
}

function copyHex() {
  const text = hexRows.value.map(r => r.hex.join(' ')).join('\n')
  navigator.clipboard.writeText(text)
    .then(() => ElMessage.success('十六进制数据已复制'))
    .catch(() => ElMessage.error('复制失败'))
}

function copyAscii() {
  const text = hexRows.value.map(r => r.ascii).join('\n')
  navigator.clipboard.writeText(text)
    .then(() => ElMessage.success('ASCII 文本已复制'))
    .catch(() => ElMessage.error('复制失败'))
}

function formatOffset(offset: number): string {
  return offset.toString(16).padStart(8, '0').toUpperCase()
}

/** 菜单 / 工具栏的十六进制翻页命令此前无人处理，接上后按钮与菜单都能用 */
function onEditorCommand(e: Event) {
  const detail = (e as CustomEvent).detail
  const cmd = typeof detail === 'string' ? detail : detail?.cmd
  if (cmd === 'pre-hex-page') prevPage()
  else if (cmd === 'next-hex-page') nextPage()
  else if (cmd === 'goto-hex-page') gotoPage()
}

/** 切换标签时重新加载：相同组件实例复用，props.tab 变化不会触发 onMounted */
watch(() => props.tab.path, (p, old) => {
  if (p && p !== old) {
    buffer.value = []
    totalBytes.value = 0
    fallback.value = false
    fallbackBytes = null
    loadPage(1)
  }
})

onMounted(() => {
  loadPage(1)
  document.addEventListener('editor-command', onEditorCommand)
})
onUnmounted(() => document.removeEventListener('editor-command', onEditorCommand))
</script>

<template>
  <div class="hex-view">
    <!-- Toolbar -->
    <div class="hex-bar">
      <FileText :size="14" class="hex-bar-icon" />
      <span class="hex-bar-name" :title="tab.path">{{ tab.name }}</span>
      <span class="hex-sep" />
      <button class="hex-btn" :disabled="currentPage <= 1" title="上一页" @click="prevPage">
        <ChevronLeft :size="14" />
      </button>
      <span class="hex-page">页 {{ currentPage }} / {{ totalPages }}</span>
      <button class="hex-btn" :disabled="currentPage >= totalPages" title="下一页" @click="nextPage">
        <ChevronRight :size="14" />
      </button>
      <button class="hex-btn" title="转到页" @click="gotoPage">
        <ArrowRightToLine :size="14" />
      </button>
      <span class="hex-sep" />
      <button class="hex-btn" title="重新加载" @click="loadPage(currentPage)">
        <RefreshCw :size="14" />
      </button>
      <button class="hex-btn" title="复制十六进制" @click="copyHex">
        <Copy :size="14" />
      </button>
      <button class="hex-btn" title="复制 ASCII" @click="copyAscii">
        <Copy :size="14" />
      </button>
      <span class="hex-spacer" />
      <span class="hex-info">{{ totalBytes.toLocaleString() }} 字节</span>
    </div>

    <!-- Content -->
    <div v-if="loading" class="hex-empty">加载中…</div>
    <div v-else-if="hexRows.length === 0" class="hex-empty">文件为空</div>
    <div v-else class="hex-body">
      <div v-for="(row, idx) in hexRows" :key="idx" class="hex-row">
        <span class="hex-offset">{{ formatOffset(row.offset) }}</span>
        <span class="hex-pipe">│</span>
        <span class="hex-bytes">
          <span
            v-for="(b, bi) in row.hex"
            :key="bi"
            class="hex-byte"
            :class="{ 'hex-byte-gap': bi === 7 }"
          >{{ b }}</span>
        </span>
        <span class="hex-pipe">│</span>
        <span class="hex-ascii">{{ row.ascii }}</span>
      </div>
    </div>

    <!-- Footer -->
    <div class="hex-foot">
      <span>页大小 {{ pageSize }} 字节</span>
      <span class="hex-sep" />
      <span>偏移 {{ formatOffset((currentPage - 1) * pageSize) }} - {{ formatOffset(Math.max(0, Math.min(currentPage * pageSize - 1, totalBytes - 1))) }}</span>
    </div>
  </div>
</template>

<style scoped>
/* 令牌化：此前工具栏/行分隔线/字节色均为字面色值，深浅色下与整体不一致 */
.hex-view {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--et-bg);
  color: var(--et-fg);
  font-family: 'Cascadia Mono', Consolas, Monaco, monospace;
}

.hex-bar {
  display: flex;
  align-items: center;
  gap: 2px;
  height: 30px;
  padding: 0 8px;
  flex-shrink: 0;
  background: var(--et-bg-sunken);
  border-bottom: 1px solid var(--et-border);
  font-size: 12px;
}
.hex-bar-icon { color: var(--et-fg-subtle); flex-shrink: 0; }
.hex-bar-name {
  max-width: 40%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--et-fg);
  margin-left: 4px;
}
.hex-page { color: var(--et-fg-muted); padding: 0 4px; }
.hex-info { color: var(--et-fg-subtle); }
.hex-spacer { flex: 1 1 auto; }
.hex-sep { width: 1px; height: 14px; background: var(--et-border); margin: 0 6px; flex-shrink: 0; }

.hex-btn {
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
  transition: background .12s ease, color .12s ease;
}
.hex-btn:hover:not(:disabled) { background: var(--et-bg-hover); color: var(--et-accent); }
.hex-btn:disabled { opacity: .35; cursor: default; }

.hex-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--et-fg-subtle);
  font-size: 12px;
}

.hex-body { flex: 1; overflow: auto; font-size: 12px; }
.hex-row { display: flex; align-items: center; padding: 1px 0; }
.hex-row:hover { background: var(--et-bg-hover); }

.hex-offset {
  color: var(--et-fg-subtle);
  padding: 0 10px;
  user-select: none;
  white-space: pre;
}
.hex-pipe { color: var(--et-border); user-select: none; }
.hex-bytes { padding: 0 8px; white-space: pre; }
.hex-byte {
  display: inline-block;
  width: 24px;
  text-align: center;
  color: var(--et-accent);
}
.hex-byte-gap { margin-right: 10px; }
.hex-ascii { padding: 0 10px; color: var(--et-fg-muted); white-space: pre; }

.hex-foot {
  display: flex;
  align-items: center;
  height: 24px;
  padding: 0 8px;
  flex-shrink: 0;
  font-size: 11px;
  color: var(--et-fg-subtle);
  background: var(--et-bg-sunken);
  border-top: 1px solid var(--et-border);
}
</style>
