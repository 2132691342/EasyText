<script lang="ts" setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
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

/**
 * 对比规则（「对比 → 对比规则」对话框写入 localStorage）。
 * 此前只有写入方、没有任何读取方，用户设置的忽略空白/空行规则完全不生效。
 * 后端 CompareDiffLines 不支持选项，因此在前端预处理两侧文本，行为等价且即时生效。
 */
interface CmpRules {
  compareMode: 'before' | 'back' | 'all'
  blankMatch: boolean
  equalRatio: number
}

function loadCmpRules(): CmpRules {
  try {
    const raw = localStorage.getItem('file-cmp-rules')
    if (raw) {
      const r = JSON.parse(raw)
      return {
        compareMode: r.compareMode || 'before',
        blankMatch: r.blankMatch !== false,
        equalRatio: r.equalRatio || 50,
      }
    }
  } catch { /* 规则损坏时回退默认值 */ }
  return { compareMode: 'before', blankMatch: true, equalRatio: 50 }
}

function applyCmpRules(text: string): string {
  const r = loadCmpRules()
  let lines = text.split('\n')
  if (r.compareMode === 'before') lines = lines.map(l => l.replace(/^\s+/, ''))
  else if (r.compareMode === 'back') lines = lines.map(l => l.replace(/\s+$/, ''))
  else if (r.compareMode === 'all') lines = lines.map(l => l.replace(/\s+/g, ''))
  if (!r.blankMatch) lines = lines.filter(l => l.trim() !== '')
  return lines.join('\n')
}

/** 规则对话框点击「应用」后事件驱动重算，避免用户还要手动再点一次对比 */
function onCmpRulesUpdated() {
  if (leftContent.value && rightContent.value) compare()
}

async function compare() {
  if (!leftContent.value || !rightContent.value) return
  isComparing.value = true
  try {
    const blocks = await CompareDiffLines(applyCmpRules(leftContent.value), applyCmpRules(rightContent.value))
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

/** 当前差异块在 diffBlocks 中的下标（用于高亮与滚动定位） */
const currentBlockPos = computed(() => {
  const block = changedBlocks.value[currentDiffIndex.value]
  return block ? diffBlocks.value.indexOf(block) : -1
})

/**
 * 跳到上一处/下一处差异。
 * 此前只改 currentDiffIndex，界面既不高亮也不滚动，用户点了看不到任何变化。
 */
async function gotoDiff(dir: 1 | -1) {
  const n = changedBlocks.value.length
  if (n === 0) return
  currentDiffIndex.value = (currentDiffIndex.value + dir + n) % n
  await nextTick()
  document.querySelector('.diff-current')?.scrollIntoView({ block: 'center', behavior: 'smooth' })
}

function prevDiff() { gotoDiff(-1) }

onMounted(() => document.addEventListener('cmp-rules-updated', onCmpRulesUpdated))
onUnmounted(() => document.removeEventListener('cmp-rules-updated', onCmpRulesUpdated))

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

function nextDiff() { gotoDiff(1) }

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
  <ModalOverlay :visible="visible" title="文档对比 (Diff)" size="full" @close="emit('close')">
    <template #header-actions>
      <button
        class="dv-header-btn"
        :class="{ 'is-on': viewMode === 'side' }"
        title="左右分栏"
        @click="viewMode = 'side'"
      >
        <Columns :size="14" :stroke-width="1.6" />
      </button>
      <button
        class="dv-header-btn"
        :class="{ 'is-on': viewMode === 'unified' }"
        title="统一视图"
        @click="viewMode = 'unified'"
      >
        <Rows :size="14" :stroke-width="1.6" />
      </button>
      <button
        class="dv-header-btn"
        :class="{ 'is-warn': showCharDiff }"
        title="字符级差异高亮"
        @click="showCharDiff = !showCharDiff"
      >
        <Highlighter :size="14" :stroke-width="1.6" />
      </button>
      <button class="dv-header-action" title="导出 HTML" @click="exportDiff">
        <Download :size="12" :stroke-width="1.6" />
        导出
      </button>
      <span v-if="diffBlocks.length > 0" class="dv-stats">
        <span class="dv-stat-add">+{{ diffStats.added }}</span>
        <span class="dv-stat-rm">-{{ diffStats.removed }}</span>
        <span>{{ changedBlocks.length }} 处差异</span>
      </span>
      <button
        class="dv-header-btn"
        title="上一处差异"
        :disabled="changedBlocks.length === 0"
        @click="prevDiff"
      >
        <ChevronLeft :size="14" :stroke-width="1.6" />
      </button>
      <button
        class="dv-header-btn"
        title="下一处差异"
        :disabled="changedBlocks.length === 0"
        @click="nextDiff"
      >
        <ChevronRight :size="14" :stroke-width="1.6" />
      </button>
    </template>

    <div class="dv">
      <!-- File selectors -->
      <div class="dv-files">
        <div class="dv-file">
          <button class="et-btn-sm" @click="openLeftFile">
            <FolderOpen :size="12" :stroke-width="1.6" />
            打开
          </button>
          <span class="dv-file-label">{{ leftLabel }}</span>
        </div>
        <div class="dv-file">
          <button class="et-btn-sm" @click="openRightFile">
            <FolderOpen :size="12" :stroke-width="1.6" />
            打开
          </button>
          <span class="dv-file-label">{{ rightLabel }}</span>
        </div>
      </div>

      <!-- Empty state -->
      <div v-if="!leftContent && !rightContent" class="dv-empty">
        <div class="dv-empty-icon">⇄</div>
        <p>请选择要对比的两个文件</p>
      </div>

      <!-- Loading -->
      <div v-else-if="isComparing" class="dv-loading">
        <svg class="dv-spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" class="dv-spin-track"></circle>
          <path fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
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
                  :data-block="line.blockIdx"
                  :class="[getLineClass(line.type), { 'diff-current': line.blockIdx === currentBlockPos }]"
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
                  :data-block="line.blockIdx"
                  :class="[getLineClass(line.type), { 'diff-current': line.blockIdx === currentBlockPos }]"
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
              <tr
                class="diff-row"
                :data-block="line.blockIdx"
                :class="[getLineClass(line.type), { 'diff-current': line.blockIdx === currentBlockPos }]"
              >
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
  </ModalOverlay>
</template>

<style scoped>
/* ===== v2.1 — token 化的 DiffView 容器 ===== */
.dv {
  display: flex;
  flex-direction: column;
  height: calc(85vh - 100px);
  min-height: 400px;
}
.dv-files {
  display: flex;
  border-bottom: 1px solid var(--et-border);
  flex-shrink: 0;
}
.dv-file {
  flex: 1;
  display: flex;
  align-items: center;
  gap: var(--et-space-2);
  padding: var(--et-space-2) var(--et-space-3);
}
.dv-file + .dv-file { border-left: 1px solid var(--et-border); }
.dv-file-label {
  flex: 1;
  font-size: var(--et-text-sm);
  color: var(--et-fg-muted);
  font-family: var(--editor-font-family, 'Consolas', monospace);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.dv-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--et-fg-subtle);
}
.dv-empty-icon {
  font-size: 48px;
  margin-bottom: var(--et-space-3);
}
.dv-loading {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--et-accent);
}
.dv-spin { width: 24px; height: 24px; animation: dv-spin 1s linear infinite; }
@keyframes dv-spin { to { transform: rotate(360deg); } }
.dv-spin-track { opacity: .25; }

/* header 内联小按钮 */
.dv-header-btn,
.dv-header-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--et-space-1);
  background: transparent;
  border: 0;
  padding: 4px;
  height: 22px;
  color: var(--et-fg-subtle);
  border-radius: var(--et-radius-sm);
  cursor: pointer;
  transition: background-color 80ms ease, color 80ms ease;
  font-size: var(--et-text-xs);
}
.dv-header-btn:hover:not(:disabled),
.dv-header-action:hover:not(:disabled) {
  background: var(--et-bg-hover);
  color: var(--et-fg);
}
.dv-header-btn:disabled { opacity: .4; cursor: not-allowed; }
.dv-header-btn.is-on { color: var(--et-accent); }
.dv-header-btn.is-warn { color: var(--et-warn); }
.dv-header-action {
  padding: 0 6px;
  height: 22px;
}
.dv-stats {
  display: inline-flex;
  align-items: center;
  gap: var(--et-space-1);
  font-size: var(--et-text-xs);
  color: var(--et-fg-muted);
  padding: 0 var(--et-space-2);
}
.dv-stat-add { color: var(--et-success); font-weight: var(--et-fw-medium); }
.dv-stat-rm  { color: var(--et-danger);  font-weight: var(--et-fw-medium); }
</style>

<!-- 上文样式继续（原 scoped 保留）-->
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

/* 当前定位到的差异块：导航时可见，避免「点了上一处/下一处」毫无反馈 */
.diff-current {
  outline: 2px solid var(--et-accent);
  outline-offset: -2px;
}

.diff-unchanged {
  background-color: transparent;
}

.line-no {
  font-size: 11px;
  min-width: 40px;
}
</style>
