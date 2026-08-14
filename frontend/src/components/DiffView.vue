<script lang="ts" setup>
import { ref, computed, watch } from 'vue'
import { CompareDiffLines, CompareCharacters, OpenFileDialog, ReadFile, SaveFile, SaveFileDialog } from '../../wailsjs/go/main/App'
import { X, ChevronLeft, ChevronRight, FolderOpen, Columns, Rows, Download, Highlighter } from 'lucide-vue-next'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  visible: boolean
  leftPath?: string
  rightPath?: string
}>()

const emit = defineEmits(['close'])

const leftContent = ref('')
const rightContent = ref('')
const leftLabel = ref('原始文件')
const rightLabel = ref('修改文件')
const diffBlocks = ref<any[]>([])
const isComparing = ref(false)
const currentDiffIndex = ref(-1)

// 🆕 V2.0.0 视图模式
const viewMode = ref<'side' | 'unified'>('side')

const diffStats = computed(() => {
  let added = 0, removed = 0, unchanged = 0
  for (const block of diffBlocks.value) {
    for (const line of block.lines) {
      if (line.type === 'added') added++
      else if (line.type === 'removed') removed++
      else unchanged++
    }
  }
  return { added, removed, unchanged }
})

const changedBlocks = computed(() =>
  diffBlocks.value.filter((b: any) => b.lines.some((l: any) => l.type !== 'unchanged'))
)

async function openLeftFile() {
  try {
    const path = await OpenFileDialog()
    if (!path) return
    const result = await ReadFile(path)
    if (result) {
      leftContent.value = result.content
      leftLabel.value = path.split(/[/\\]/).pop() || path
      if (rightContent.value) await compare()
    }
  } catch (e) {
    console.error('Failed to open left file:', e)
    ElMessage.error('打开文件失败')
  }
}

async function openRightFile() {
  try {
    const path = await OpenFileDialog()
    if (!path) return
    const result = await ReadFile(path)
    if (result) {
      rightContent.value = result.content
      rightLabel.value = path.split(/[/\\]/).pop() || path
      if (leftContent.value) await compare()
    }
  } catch (e) {
    console.error('Failed to open right file:', e)
    ElMessage.error('打开文件失败')
  }
}

async function compare() {
  if (!leftContent.value || !rightContent.value) return
  isComparing.value = true
  try {
    const blocks = await CompareDiffLines(leftContent.value, rightContent.value)
    diffBlocks.value = blocks || []
    currentDiffIndex.value = changedBlocks.value.length > 0 ? 0 : -1
    ElMessage.success('对比完成')
  } catch (e) {
    console.error('Diff failed:', e)
    ElMessage.error('对比失败')
  } finally {
    isComparing.value = false
  }
}

function prevDiff() {
  if (changedBlocks.value.length === 0) return
  currentDiffIndex.value = (currentDiffIndex.value - 1 + changedBlocks.value.length) % changedBlocks.value.length
}

// 自动加载外部传入的左右文件路径
async function loadByPath(side: 'left' | 'right', path: string) {
  try {
    const result = await ReadFile(path)
    if (!result) return
    if (side === 'left') { leftContent.value = result.content; leftLabel.value = path.split(/[/\\]/).pop() || path }
    else { rightContent.value = result.content; rightLabel.value = path.split(/[/\\]/).pop() || path }
  } catch (e) { console.warn(e) }
}
watch(() => props.leftPath, async (p) => { if (p) { await loadByPath('left', p); if (rightContent.value) compare() } })
watch(() => props.rightPath, async (p) => { if (p) { await loadByPath('right', p); if (leftContent.value) compare() } })

function nextDiff() {
  if (changedBlocks.value.length === 0) return
  currentDiffIndex.value = (currentDiffIndex.value + 1) % changedBlocks.value.length
}

function getLineClass(type: string) {
  switch (type) {
    case 'added': return 'diff-added'
    case 'removed': return 'diff-removed'
    default: return 'diff-unchanged'
  }
}

const leftLines = computed(() => {
  const lines: Array<{ lineNo: number | null; content: string; type: string; blockIdx: number; lineIdx: number }> = []
  let bi = 0
  for (const block of diffBlocks.value) {
    let li = 0
    for (const line of block.lines) {
      if (line.type === 'removed' || line.type === 'unchanged') {
        lines.push({ lineNo: line.oldLine || null, content: line.content, type: line.type, blockIdx: bi, lineIdx: li })
      } else if (line.type === 'added') {
        lines.push({ lineNo: null, content: '', type: 'placeholder', blockIdx: bi, lineIdx: li })
      }
      li++
    }
    bi++
  }
  return lines
})

const rightLines = computed(() => {
  const lines: Array<{ lineNo: number | null; content: string; type: string; blockIdx: number; lineIdx: number }> = []
  let bi = 0
  for (const block of diffBlocks.value) {
    let li = 0
    for (const line of block.lines) {
      if (line.type === 'added' || line.type === 'unchanged') {
        lines.push({ lineNo: line.newLine || null, content: line.content, type: line.type, blockIdx: bi, lineIdx: li })
      } else if (line.type === 'removed') {
        lines.push({ lineNo: null, content: '', type: 'placeholder', blockIdx: bi, lineIdx: li })
      }
      li++
    }
    bi++
  }
  return lines
})

// 🆕 V2.0.0 字符级差异高亮
const showCharDiff = ref(false)
// 存储每个 diff block 中修改行的字符级 HTML 渲染结果
// key: `${blockIndex}:${lineIndexInBlock}`, value: HTML string
const charDiffHtml = ref<Record<string, string>>({})
const charDiffLoading = ref(false)

async function computeCharDiffs() {
  if (!showCharDiff.value || diffBlocks.value.length === 0) {
    charDiffHtml.value = {}
    return
  }
  charDiffLoading.value = true
  const results: Record<string, string> = {}
  try {
    for (let bi = 0; bi < diffBlocks.value.length; bi++) {
      const block = diffBlocks.value[bi]
      const lines = block.lines
      for (let li = 0; li < lines.length; li++) {
        const line = lines[li]
        if (line.type === 'changed' || line.type === 'added' || line.type === 'removed') {
          // 对于修改行：找到配对的 removed 和 added 行
          if (line.type === 'removed' && li + 1 < lines.length && lines[li + 1].type === 'added') {
            try {
              const html = await CompareCharacters(line.content, lines[li + 1].content)
              results[`${bi}:${li}`] = html
              results[`${bi}:${li + 1}`] = html
            } catch { /* ignore */ }
          } else if (line.type === 'added' && li > 0 && lines[li - 1].type === 'removed') {
            // 已经在前一个 removed 行处理过了
            continue
          }
        }
      }
    }
    charDiffHtml.value = results
  } catch { /* ignore */ }
  charDiffLoading.value = false
}

watch(showCharDiff, async (v) => {
  if (v) await computeCharDiffs()
})

// 获取字符级差异 HTML（用于 v-html）
function getCharDiffHtml(blockIdx: number, lineIdx: number, lineContent: string): string {
  if (!showCharDiff.value) return escapeHtml(lineContent)
  const key = `${blockIdx}:${lineIdx}`
  const html = charDiffHtml.value[key]
  if (!html) return escapeHtml(lineContent)
  // 转义 HTML 中非 <ins>/<del> 标签的内容
  return html
}

// 在 compare 后重新计算字符级差异
watch(diffBlocks, async () => {
  if (showCharDiff.value && diffBlocks.value.length > 0) {
    await computeCharDiffs()
  }
})

// 🆕 V2.0.0 统一视图
const unifiedLines = computed(() => {
  const lines: Array<{ oldLine: number | null; newLine: number | null; content: string; type: string; blockIdx: number; lineIdx: number }> = []
  let bi = 0
  for (const block of diffBlocks.value) {
    let li = 0
    for (const line of block.lines) {
      lines.push({
        oldLine: line.oldLine || null,
        newLine: line.newLine || null,
        content: line.content,
        type: line.type,
        blockIdx: bi,
        lineIdx: li,
      })
      li++
    }
    bi++
  }
  return lines
})

// 🆕 V2.0.0 导出 HTML
async function exportDiff() {
  try {
    const path = await SaveFileDialog('diff.html')
    if (!path) return
    let html = `<!DOCTYPE html><html><head><meta charset="utf-8"><style>
      body{font-family:Consolas,monospace;font-size:12px;background:#1e1e1e;color:#d4d4d4;margin:0;padding:8px}
      .header{display:flex;border-bottom:1px solid #444;padding-bottom:8px;margin-bottom:8px}
      .header div{flex:1;font-weight:bold;color:#888}
      .added{background:rgba(34,197,94,0.15)}
      .removed{background:rgba(239,68,68,0.15)}
      .line-no{color:#666;padding:0 8px;text-align:right;min-width:40px;user-select:none}
      td{padding:1px 4px;white-space:pre-wrap}
      .sign{color:#888;padding:0 4px;text-align:center}
      .added .sign{color:#4ade80}
      .removed .sign{color:#f87171}
      </style></head><body><div class="header"><div>${leftLabel.value}</div><div>${rightLabel.value}</div></div><table>`
    for (const line of unifiedLines.value) {
      const sign = line.type === 'added' ? '+' : line.type === 'removed' ? '-' : ' '
      html += `<tr class="${line.type}"><td class="line-no">${line.oldLine ?? ''}</td><td class="line-no">${line.newLine ?? ''}</td><td class="sign">${sign}</td><td>${escapeHtml(line.content)}</td></tr>`
    }
    html += '</table></body></html>'
    await SaveFile(path, html, 'UTF-8')
    ElMessage.success('Diff 已导出为 HTML')
  } catch (e: any) {
    ElMessage.error(`导出失败: ${e?.message || ''}`)
  }
}

function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}
</script>

<template>
  <div v-if="visible" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
    <div class="bg-white dark:bg-[#1e1e1e] rounded-lg shadow-2xl w-[90vw] h-[85vh] flex flex-col overflow-hidden">
      <!-- Header -->
      <div class="flex items-center justify-between px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#252526]">
        <h2 class="text-sm font-semibold text-gray-700 dark:text-gray-200">文档对比 (Diff)</h2>
        <div class="flex items-center gap-2">
          <!-- 🆕 V2.0.0 视图模式切换 -->
          <button
            class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 text-xs"
            :class="viewMode === 'side' ? 'text-blue-500' : 'text-gray-400'"
            title="左右分栏"
            @click="viewMode = 'side'"
          >
            <Columns class="w-3.5 h-3.5" />
          </button>
          <button
            class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 text-xs"
            :class="viewMode === 'unified' ? 'text-blue-500' : 'text-gray-400'"
            title="统一视图"
            @click="viewMode = 'unified'"
          >
            <Rows class="w-3.5 h-3.5" />
          </button>
          <!-- 🆕 V2.0.0 字符级高亮开关 -->
          <button
            class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 text-xs"
            :class="showCharDiff ? 'text-orange-500' : 'text-gray-400'"
            title="字符级差异高亮"
            @click="showCharDiff = !showCharDiff"
          >
            <Highlighter class="w-3.5 h-3.5" />
          </button>
          <!-- 🆕 导出 -->
          <button
            class="flex items-center gap-1 px-2 py-0.5 text-xs rounded bg-gray-100 dark:bg-[#3c3c3c] hover:bg-gray-200 dark:hover:bg-[#4a4a4a]"
            title="导出 HTML"
            @click="exportDiff"
          >
            <Download class="w-3 h-3" />
          </button>
          <span v-if="diffBlocks.length > 0" class="text-xs text-gray-500 dark:text-gray-400">
            <span class="text-green-600 dark:text-green-400">+{{ diffStats.added }}</span>
            <span class="mx-1 text-red-600 dark:text-red-400">-{{ diffStats.removed }}</span>
            <span class="text-gray-400">{{ changedBlocks.length }} 处差异</span>
          </span>
          <!-- Navigate diffs -->
          <button
            class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-40"
            title="上一处差异"
            :disabled="changedBlocks.length === 0"
            @click="prevDiff"
          >
            <ChevronLeft class="w-4 h-4" />
          </button>
          <button
            class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-40"
            title="下一处差异"
            :disabled="changedBlocks.length === 0"
            @click="nextDiff"
          >
            <ChevronRight class="w-4 h-4" />
          </button>
          <button class="p-1 rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c]" @click="emit('close')">
            <X class="w-4 h-4" />
          </button>
        </div>
      </div>

      <!-- File selectors -->
      <div class="flex border-b border-gray-200 dark:border-gray-700">
        <div class="flex-1 flex items-center gap-2 px-3 py-2 border-r border-gray-200 dark:border-gray-700">
          <button
            class="flex items-center gap-1 px-2 py-1 text-xs rounded bg-gray-100 dark:bg-[#3c3c3c] hover:bg-gray-200 dark:hover:bg-[#4a4a4a]"
            @click="openLeftFile"
          >
            <FolderOpen class="w-3 h-3" />
            打开
          </button>
          <span class="text-xs text-gray-600 dark:text-gray-300 truncate flex-1 font-mono">{{ leftLabel }}</span>
        </div>
        <div class="flex-1 flex items-center gap-2 px-3 py-2">
          <button
            class="flex items-center gap-1 px-2 py-1 text-xs rounded bg-gray-100 dark:bg-[#3c3c3c] hover:bg-gray-200 dark:hover:bg-[#4a4a4a]"
            @click="openRightFile"
          >
            <FolderOpen class="w-3 h-3" />
            打开
          </button>
          <span class="text-xs text-gray-600 dark:text-gray-300 truncate flex-1 font-mono">{{ rightLabel }}</span>
        </div>
      </div>

      <!-- Empty state -->
      <div v-if="!leftContent && !rightContent" class="flex-1 flex items-center justify-center text-gray-400 dark:text-gray-500">
        <div class="text-center">
          <div class="text-4xl mb-3">⇄</div>
          <p class="text-sm">请选择要对比的两个文件</p>
        </div>
      </div>

      <!-- Loading -->
      <div v-else-if="isComparing" class="flex-1 flex items-center justify-center">
        <svg class="animate-spin h-6 w-6 text-blue-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
      </div>

      <!-- Diff content - side by side -->
      <div v-else-if="viewMode === 'side'" class="flex-1 overflow-hidden flex">
        <!-- Left panel -->
        <div class="flex-1 overflow-auto border-r border-gray-200 dark:border-gray-700">
          <div v-if="diffBlocks.length === 0" class="p-4 text-xs text-gray-400 font-mono whitespace-pre-wrap">{{ leftContent }}</div>
          <table v-else class="w-full text-xs font-mono border-collapse">
            <tbody>
              <template v-for="(line, i) in leftLines" :key="i">
                <tr
                  class="diff-row"
                  :class="getLineClass(line.type)"
                >
                  <td class="line-no select-none w-10 text-right pr-2 text-gray-400 dark:text-gray-600 border-r border-gray-200 dark:border-gray-700 sticky left-0 bg-inherit">
                    {{ line.lineNo ?? '' }}
                  </td>
                  <td class="px-2 py-0.5 whitespace-pre-wrap break-all">
                    <span v-if="showCharDiff && (line.type === 'removed' || line.type === 'added') && getCharDiffHtml(line.blockIdx, line.lineIdx, line.content) !== escapeHtml(line.content)" v-html="getCharDiffHtml(line.blockIdx, line.lineIdx, line.content)"></span>
                    <span v-else>{{ line.content }}</span>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

        <!-- Right panel -->
        <div class="flex-1 overflow-auto">
          <div v-if="diffBlocks.length === 0" class="p-4 text-xs text-gray-400 font-mono whitespace-pre-wrap">{{ rightContent }}</div>
          <table v-else class="w-full text-xs font-mono border-collapse">
            <tbody>
              <template v-for="(line, i) in rightLines" :key="i">
                <tr
                  class="diff-row"
                  :class="getLineClass(line.type)"
                >
                  <td class="line-no select-none w-10 text-right pr-2 text-gray-400 dark:text-gray-600 border-r border-gray-200 dark:border-gray-700 sticky left-0 bg-inherit">
                    {{ line.lineNo ?? '' }}
                  </td>
                  <td class="px-2 py-0.5 whitespace-pre-wrap break-all">
                    <span v-if="showCharDiff && (line.type === 'added' || line.type === 'removed') && getCharDiffHtml(line.blockIdx, line.lineIdx, line.content) !== escapeHtml(line.content)" v-html="getCharDiffHtml(line.blockIdx, line.lineIdx, line.content)"></span>
                    <span v-else>{{ line.content }}</span>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 🆕 V2.0.0 统一视图 -->
      <div v-else class="flex-1 overflow-auto">
        <div v-if="diffBlocks.length === 0" class="p-4 text-xs text-gray-400 font-mono whitespace-pre-wrap">{{ leftContent }}</div>
        <table v-else class="w-full text-xs font-mono border-collapse">
          <tbody>
            <template v-for="(line, i) in unifiedLines" :key="i">
              <tr class="diff-row" :class="getLineClass(line.type)">
                <td class="line-no select-none w-10 text-right pr-2 text-gray-400 dark:text-gray-600 border-r border-gray-200 dark:border-gray-700 bg-inherit">
                  {{ line.oldLine ?? '' }}
                </td>
                <td class="line-no select-none w-10 text-right pr-2 text-gray-400 dark:text-gray-600 border-r border-gray-200 dark:border-gray-700 bg-inherit">
                  {{ line.newLine ?? '' }}
                </td>
                <td class="sign-col select-none w-6 text-center text-gray-500 bg-inherit">
                  {{ line.type === 'added' ? '+' : line.type === 'removed' ? '-' : '' }}
                </td>
                <td class="px-2 py-0.5 whitespace-pre-wrap break-all">
                    <span v-if="showCharDiff && (line.type === 'added' || line.type === 'removed') && getCharDiffHtml(line.blockIdx, line.lineIdx, line.content) !== escapeHtml(line.content)" v-html="getCharDiffHtml(line.blockIdx, line.lineIdx, line.content)"></span>
                    <span v-else>{{ line.content }}</span>
                  </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
.diff-row {
  line-height: 1.5;
}

.diff-added {
  background-color: rgba(34, 197, 94, 0.15);
}

html.dark .diff-added {
  background-color: rgba(34, 197, 94, 0.1);
}

.diff-removed {
  background-color: rgba(239, 68, 68, 0.15);
}

html.dark .diff-removed {
  background-color: rgba(239, 68, 68, 0.1);
}

.diff-placeholder {
  background-color: rgba(156, 163, 175, 0.07);
}

.diff-unchanged {
  background-color: transparent;
}

.line-no {
  font-size: 11px;
  min-width: 40px;
}
</style>
