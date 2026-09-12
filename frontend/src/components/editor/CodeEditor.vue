<script lang="ts" setup>
/**
 * CodeEditor — 编辑器容器
 *
 * 架构：容器 + composables/（theme/language/completion/bookmark/columnMode/macro/markdown）
 *       + ext/（keymap/context-menu）
 *
 * 关键机制：
 *  - 各关注点一个 compartment，运行时切换走 dispatch reconfigure，不销毁重建
 *  - tab 切换就地换绑文档（保留 undo 历史），换绑前后保存/恢复滚动位置
 *  - 内容变更防抖同步 editorStore，同时采集宏步骤与光标位置
 *  - undo/redo 走 CodeMirror 命令；位置前进/后退走 goBackPosition/goForwardPosition
 *  - minimap 视口矩形 {scrollTop, scrollHeight, clientHeight}，rAF 滚动同步
 */
import { computed, nextTick, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { EditorTab, Snippet } from '@/types'
import { useEditorStore, useSettingStore } from '@/stores'
import { AddBookmark, RemoveBookmark, FormatJSON, MinifyJSON, ValidateJSON } from '../../../wailsjs/go/main/App'
import { EditorView, Decoration, lineNumbers, highlightActiveLine, highlightActiveLineGutter, highlightSpecialChars, rectangularSelection, crosshairCursor, dropCursor } from '@codemirror/view'
import { Compartment, EditorState, Prec, StateEffect, StateField, RangeSetBuilder, RangeSet, type Extension } from '@codemirror/state'
import { syntaxHighlighting, foldGutter } from '@codemirror/language'
import { closeBrackets, closeBracketsKeymap, autocompletion, snippetCompletion, completionKeymap } from '@codemirror/autocomplete'
import { defaultKeymap, history, historyKeymap, indentWithTab, toggleComment, toggleBlockComment, undo, redo } from '@codemirror/commands'
import { foldKeymap, indentOnInput, bracketMatching } from '@codemirror/language'
import { lintKeymap } from '@codemirror/lint'
import { keymap } from '@codemirror/view'
import Minimap from './Minimap.vue'

import { useEditorTheme } from './composables/useEditorTheme'
import { useEditorLanguage } from './composables/useEditorLanguage'
import { useEditorCompletion } from './composables/useEditorCompletion'
import { useEditorBookmark } from './composables/useEditorBookmark'
import { useEditorColumnMode } from './composables/useEditorColumnMode'
import { useEditorMacro } from './composables/useEditorMacro'
import { useEditorMarkdown } from './composables/useEditorMarkdown'
import { useEditorKeymap, CMD_ALIASES } from './ext/ext-keymap'
import { useEditorContextMenu, CONTEXT_MENU_SECTIONS } from './ext/ext-context-menu'

// ==================== Props & Store ====================
const props = defineProps<{ tab: EditorTab }>()
const editorStore = useEditorStore()
const settingStore = useSettingStore()
const config = computed(() => settingStore.config)
const colors = computed(() => settingStore.currentThemeColors as any)

// ==================== 实例状态 ====================
const editorContainer = ref<HTMLElement | null>(null)
const editorView = shallowRef<EditorView | null>(null)
let isInitializing = false

// ==================== Composable 装配 ====================
const theme = useEditorTheme(colors as any, config as any)
const language = useEditorLanguage()
const completion = useEditorCompletion(
  computed(() => props.tab.language) as any,
  computed(() => editorStore.snippets ?? []) as any,
)
const bookmark = useEditorBookmark()
const columnMode = useEditorColumnMode()
const macro = useEditorMacro()
const markdown = useEditorMarkdown(
  computed(() => props.tab.language) as any,
  computed(() => props.tab.content) as any,
  colors as any,
)
const keymapExt = useEditorKeymap()
const contextMenu = useEditorContextMenu()

// ==================== Compartments ====================
const appearanceCompartment = new Compartment()
const syntaxHighlightCompartment = new Compartment()
const langCompartment = new Compartment()
const foldCompartment = new Compartment()
const wordWrapCompartment = new Compartment()
const showWhitespaceCompartment = new Compartment()
const webAddrCompartment = new Compartment()
const tabSizeCompartment = new Compartment()

// ==================== 多色 mark（原版逻辑） ====================
const MARK_COLORS = [
  'rgba(255,212,0,0.45)',   // 黄
  'rgba(255,120,120,0.45)', // 红
  'rgba(120,180,255,0.45)', // 蓝
  'rgba(120,220,150,0.45)', // 绿
  'rgba(200,140,255,0.45)', // 紫
]
let currentMarkColor = 0
interface MarkRange { from: number; to: number; color: number }
const setMarks = StateEffect.define<MarkRange[]>()
const addMarkRanges = StateEffect.define<MarkRange[]>()
const clearMarksEffect = StateEffect.define()
const markField = StateField.define<RangeSet<Decoration>>({
  create() { return RangeSet.empty },
  update(marks, tr) {
    let next = marks
    for (const e of tr.effects) {
      if (e.is(setMarks)) {
        const builder = new RangeSetBuilder<Decoration>()
        const sorted = [...e.value].sort((a, b) => a.from - b.from)
        for (const r of sorted) {
          if (r.from < r.to) builder.add(r.from, r.to, Decoration.mark({ attributes: { style: `background-color:${MARK_COLORS[r.color % MARK_COLORS.length]};border-radius:2px;` } }))
        }
        next = builder.finish()
      } else if (e.is(addMarkRanges)) {
        const existing: MarkRange[] = []
        const iter = marks.iter()
        while (iter.value) { existing.push({ from: iter.from, to: iter.to, color: 0 }); iter.next() }
        const merged = [...existing, ...e.value].sort((a, b) => a.from - b.from)
        const builder = new RangeSetBuilder<Decoration>()
        for (const r of merged) {
          if (r.from < r.to) builder.add(r.from, r.to, Decoration.mark({ attributes: { style: `background-color:${MARK_COLORS[r.color % MARK_COLORS.length]};border-radius:2px;` } }))
        }
        next = builder.finish()
      } else if (e.is(clearMarksEffect)) {
        next = RangeSet.empty
      }
    }
    return next
  },
  provide: f => EditorView.decorations.from(f, v => v),
})

// ---- URL highlight（视图→显示网页地址） ----
const webAddrField = StateField.define<RangeSet<Decoration>>({
  create(state) { return buildWebAddrDecorations(state) },
  update(decos, tr) {
    if (tr.docChanged) return buildWebAddrDecorations(tr.state)
    return decos
  },
  provide: f => EditorView.decorations.from(f, v => v),
})
const URL_RE = /\b(https?|ftp|file):\/\/[^\s<>"']+/gi
function buildWebAddrDecorations(state: any): RangeSet<Decoration> {
  const builder = new RangeSetBuilder<Decoration>()
  const doc = state.doc.toString()
  for (const m of doc.matchAll(URL_RE)) {
    if (m.index === undefined) continue
    builder.add(m.index, m.index + m[0].length, Decoration.mark({ class: 'cm-webaddr' }))
  }
  return builder.finish()
}

// ---- 高亮当前词（双击） ----
const addHighlightWord = StateEffect.define<{ from: number; to: number }>()
const clearHighlightWord = StateEffect.define()
const highlightWordField = StateField.define<RangeSet<Decoration>>({
  create() { return RangeSet.empty },
  update(highlights, tr) {
    for (const e of tr.effects) {
      if (e.is(addHighlightWord)) {
        const { from, to } = e.value
        const builder = new RangeSetBuilder<Decoration>()
        builder.add(from, to, Decoration.mark({ class: 'cm-word-highlight', attributes: { style: 'background-color:rgba(255,200,0,0.3);border-radius:2px;' } }))
        return builder.finish()
      } else if (e.is(clearHighlightWord)) {
        return RangeSet.empty
      }
    }
    if (tr.docChanged) return RangeSet.empty
    return highlights
  },
  provide: f => EditorView.decorations.from(f, v => v),
})

// ==================== 工具函数 ====================
let lastSearchTerm = ''
function setSearchTerm(term: string) { lastSearchTerm = term }
function escapeRegExp(s: string) { return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') }

function getLineCount(): number { return editorView.value?.state.doc.lines || 0 }
function getSelectedText(): string {
  return editorView.value ? editorView.value.state.sliceDoc(editorView.value.state.selection.main.from, editorView.value.state.selection.main.to) : ''
}

function gotoLine(lineNum: number) {
  const v = editorView.value
  if (!v || lineNum < 1) return
  const line = Math.min(lineNum, v.state.doc.lines)
  const pos = v.state.doc.line(line).from
  v.dispatch({ selection: { anchor: pos, head: pos }, scrollIntoView: true })
  v.focus()
}

// —— Goto line 弹窗 ——
const showGotoLine = ref(false)
const gotoLineInput = ref<number | null>(null)
function showGotoLineDialog() {
  showGotoLine.value = true
  setTimeout(() => { (document.querySelector('.goto-line-input') as HTMLInputElement | null)?.focus() }, 50)
}
function handleGotoLine() {
  if (gotoLineInput.value) gotoLine(gotoLineInput.value)
  showGotoLine.value = false
}

// ==================== 编辑操作 ====================
function cutSelection() {
  const v = editorView.value
  if (!v) return
  const { from, to } = v.state.selection.main
  if (from === to) return
  const text = v.state.sliceDoc(from, to)
  // 先写应用内剪贴板（可靠），再尽力写系统剪贴板（WebView2 可能拒权限）
  editorStore.pushClipboard(text)
  navigator.clipboard?.writeText(text).catch(() => {})
  v.dispatch({ changes: { from, to } })
  v.focus()
}
function copySelection() {
  const v = editorView.value
  if (!v) return
  const { from, to } = v.state.selection.main
  if (from === to) return
  const text = v.state.sliceDoc(from, to)
  editorStore.pushClipboard(text)
  navigator.clipboard?.writeText(text).catch(() => {})
}
async function pasteAtCursor() {
  const v = editorView.value
  if (!v) return
  let clip = editorStore.clipboardHistory[0] || ''
  if (!clip) {
    try { clip = await navigator.clipboard.readText() } catch { clip = '' }
  }
  if (clip) { v.dispatch({ changes: { from: v.state.selection.main.head, insert: clip } }); v.focus() }
}
function selectAll() {
  const v = editorView.value
  if (!v) return
  v.dispatch({ selection: { anchor: 0, head: v.state.doc.length } })
  v.focus()
}
function undoAction() {
  if (!editorView.value) return
  undo(editorView.value)
}
function redoAction() {
  if (!editorView.value) return
  redo(editorView.value)
}
function toggleCommentAction() {
  if (!editorView.value) return
  toggleComment(editorView.value)
}
function deleteSelectionOrChar() {
  const v = editorView.value
  if (!v) return
  const { from, to } = v.state.selection.main
  const end = from !== to ? to : Math.min(v.state.doc.length, from + 1)
  if (from === end) return
  v.dispatch({ changes: { from, to: end } })
  v.focus()
}
/** 剪切当前行：整行进剪贴板并连行尾换行删除（原版实现） */
function cutCurrentLine() {
  const v = editorView.value
  if (!v) return
  const state = v.state
  const line = state.doc.lineAt(state.selection.main.head)
  const text = line.text
  navigator.clipboard?.writeText(text).catch(() => {})
  editorStore.pushClipboard(text)
  const from = line.number > 1 ? line.from - 1 : line.from
  const to = line.to < state.doc.length ? line.to + 1 : line.to
  v.dispatch({ changes: { from, to: Math.min(to, state.doc.length) } })
  v.focus()
}

function doEditorAction(action: 'cut' | 'copy' | 'paste' | 'selectAll') {
  if (action === 'selectAll') selectAll()
  else if (action === 'cut') cutSelection()
  else if (action === 'copy') copySelection()
  else if (action === 'paste') pasteAtCursor()
}

// ==================== 行操作 / 排版（原版逐字搬运） ====================
function lineOperation(op: string) {
  const v = editorView.value
  if (!v) return
  const state = v.state
  const from = state.selection.main.from
  const to = state.selection.main.to
  const fromLine = state.doc.lineAt(from)
  const toLine = state.doc.lineAt(to)
  switch (op) {
    case 'duplicate': {
      const text = state.sliceDoc(fromLine.from, toLine.to)
      v.dispatch({ changes: { from: toLine.to, insert: '\n' + text } })
      break
    }
    case 'remove': {
      const end = toLine.to + 1 > state.doc.length ? state.doc.length : toLine.to + 1
      v.dispatch({ changes: { from: fromLine.from, to: end } })
      break
    }
    case 'moveUp': {
      if (fromLine.number <= 1) return
      const prevLine = state.doc.line(fromLine.number - 1)
      const text = state.sliceDoc(fromLine.from, toLine.to)
      const hasNL = toLine.to < state.doc.length && state.sliceDoc(toLine.to, toLine.to + 1) === '\n'
      v.dispatch({ changes: [{ from: fromLine.from - (prevLine.length + 1), to: toLine.to + (hasNL ? 1 : 0), insert: text + (hasNL ? '\n' : '') + prevLine.text }] })
      break
    }
    case 'moveDown': {
      if (toLine.number >= state.doc.lines) return
      const nextLine = state.doc.line(toLine.number + 1)
      const text = state.sliceDoc(fromLine.from, toLine.to)
      v.dispatch({ changes: [{ from: fromLine.from, to: nextLine.to, insert: nextLine.text + '\n' + text }] })
      break
    }
    case 'removeEmpty': {
      const lines = state.doc.toString().split('\n').filter(l => l.trim() !== '')
      v.dispatch({ changes: { from: 0, to: state.doc.length, insert: lines.join('\n') } })
      break
    }
    case 'removeBlank': {
      const lines = state.doc.toString().split('\n').filter(l => l.length > 0)
      v.dispatch({ changes: { from: 0, to: state.doc.length, insert: lines.join('\n') } })
      break
    }
    case 'split': {
      if (to > from) {
        const text = state.sliceDoc(from, to)
        v.dispatch({ changes: { from, to, insert: [...text].join('\n') } })
      }
      break
    }
    case 'join': {
      if (to > from) {
        v.dispatch({ changes: { from, to, insert: state.sliceDoc(from, to).replace(/\n/g, ' ') } })
      }
      break
    }
    case 'removeDuplicate': {
      const lines = state.doc.toString().split('\n')
      const seen = new Set<string>()
      v.dispatch({ changes: { from: 0, to: state.doc.length, insert: lines.filter(l => { if (seen.has(l)) return false; seen.add(l); return true }).join('\n') } })
      break
    }
    case 'removeConsecutiveDuplicate': {
      const lines = state.doc.toString().split('\n')
      const result: string[] = []; let last = ''
      for (const l of lines) { if (l !== last) { result.push(l); last = l } }
      v.dispatch({ changes: { from: 0, to: state.doc.length, insert: result.join('\n') } })
      break
    }
    case 'reverse': {
      const lines = state.doc.toString().split('\n')
      v.dispatch({ changes: { from: 0, to: state.doc.length, insert: lines.reverse().join('\n') } })
      break
    }
    case 'randomize': {
      const lines = state.doc.toString().split('\n')
      for (let i = lines.length - 1; i > 0; i--) { const j = Math.floor(Math.random() * (i + 1)); [lines[i], lines[j]] = [lines[j], lines[i]] }
      v.dispatch({ changes: { from: 0, to: state.doc.length, insert: lines.join('\n') } })
      break
    }
    case 'insertAbove': insertBlankLine(true); break
    case 'insertBelow': insertBlankLine(false); break
  }
}

function insertBlankLine(above: boolean) {
  const v = editorView.value
  if (!v) return
  const state = v.state
  const pos = state.selection.main.head
  const line = state.doc.lineAt(pos)
  if (above) v.dispatch({ changes: { from: line.from, insert: '\n' }, selection: { anchor: line.from } })
  else v.dispatch({ changes: { from: line.to, insert: '\n' }, selection: { anchor: line.to + 1 } })
}

function convertTabsSpaces(type: string) {
  const v = editorView.value
  if (!v) return
  const state = v.state
  const tabSize = config.value?.editor?.tabSize || 4
  const text = state.doc.toString()
  let result = text
  switch (type) {
    case 'tabToSpaces': result = text.replace(/\t/g, ' '.repeat(tabSize)); break
    case 'spacesAllToTabs': result = text.replace(new RegExp(` {${tabSize}}`, 'g'), '\t'); break
    case 'spacesLeadingToTabs': {
      result = text.split('\n').map(line => {
        const m = line.match(/^ +/)
        if (!m) return line
        const spaces = m[0].length
        return '\t'.repeat(Math.floor(spaces / tabSize)) + ' '.repeat(spaces % tabSize) + line.slice(spaces)
      }).join('\n')
      break
    }
  }
  if (result !== text) v.dispatch({ changes: { from: 0, to: state.doc.length, insert: result } })
}

function trimWhitespace(mode: string) {
  const v = editorView.value
  if (!v) return
  const state = v.state
  const text = state.doc.toString()
  let result = text
  switch (mode) {
    case 'head': result = text.split('\n').map(l => l.replace(/^[ \t]+/, '')).join('\n'); break
    case 'tail': result = text.split('\n').map(l => l.replace(/[ \t]+$/, '')).join('\n'); break
    case 'both': result = text.split('\n').map(l => l.replace(/^[ \t]+/, '').replace(/[ \t]+$/, '')).join('\n'); break
  }
  if (result !== text) v.dispatch({ changes: { from: 0, to: state.doc.length, insert: result } })
}

function sortLines(direction: string) {
  const v = editorView.value
  if (!v) return
  const lines = v.state.doc.toString().split('\n')
  const isCI = direction.endsWith('-ci') || direction.endsWith('-case')
  const isInt = direction.startsWith('int-')
  const isFloat = direction.startsWith('float-')
  const isComma = direction.startsWith('comma-')
  const isDesc = direction.includes('desc')
  const toNum = (s: string) => {
    let n: number
    if (isComma) n = parseFloat(s.replace(',', '.'))
    else if (isInt) n = parseInt(s)
    else n = parseFloat(s)
    return isNaN(n) ? 0 : n
  }
  let sorted: string[]
  if (isInt || isFloat || isComma) sorted = [...lines].sort((a, b) => (isDesc ? toNum(b) - toNum(a) : toNum(a) - toNum(b)))
  else if (isCI) sorted = [...lines].sort((a, b) => { const cmp = a.toLowerCase().localeCompare(b.toLowerCase()); return isDesc ? -cmp : cmp })
  else sorted = [...lines].sort((a, b) => { const cmp = a.localeCompare(b); return isDesc ? -cmp : cmp })
  v.dispatch({ changes: { from: 0, to: v.state.doc.length, insert: sorted.join('\n') } })
}

function indentLines() {
  const v = editorView.value
  if (!v) return
  const state = v.state
  const ts = config.value?.editor?.tabSize || 4
  const useSpaces = config.value?.editor?.insertSpaces ?? true
  const indent = useSpaces ? ' '.repeat(ts) : '\t'
  const fl = state.doc.lineAt(state.selection.main.from)
  const tl = state.doc.lineAt(state.selection.main.to)
  const changes = []
  for (let i = fl.number; i <= tl.number; i++) changes.push({ from: state.doc.line(i).from, insert: indent })
  v.dispatch({ changes })
}

function dedentLines() {
  const v = editorView.value
  if (!v) return
  const state = v.state
  const ts = config.value?.editor?.tabSize || 4
  const fl = state.doc.lineAt(state.selection.main.from)
  const tl = state.doc.lineAt(state.selection.main.to)
  const changes = []
  for (let i = fl.number; i <= tl.number; i++) {
    const line = state.doc.line(i)
    const t = line.text
    if (t.startsWith('\t')) changes.push({ from: line.from, to: line.from + 1 })
    else if (t.startsWith(' '.repeat(ts))) changes.push({ from: line.from, to: line.from + ts })
    else if (t.startsWith(' ')) { const n = t.match(/^ +/)![0].length; changes.push({ from: line.from, to: line.from + Math.min(n, ts) }) }
  }
  if (changes.length > 0) v.dispatch({ changes })
}

// ---- Case transform（原版逐字搬运） ----
function splitWords(text: string): string[] {
  return text.trim().split(/[^A-Za-z0-9]+/).filter(Boolean).map(w => w.toLowerCase())
}
function toPascalWords(text: string): string {
  return splitWords(text).map(w => w.charAt(0).toUpperCase() + w.slice(1)).join('')
}
function transformCase(type: string) {
  const v = editorView.value
  if (!v) return
  const state = v.state
  const selection = state.selection.main
  const isSelection = selection.from !== selection.to
  const text = isSelection ? state.sliceDoc(selection.from, selection.to) : state.sliceDoc(state.doc.lineAt(selection.from).from, state.doc.lineAt(selection.from).to)
  let result = text
  switch (type) {
    case 'upper': result = text.toUpperCase(); break
    case 'lower': result = text.toLowerCase(); break
    case 'invert': result = [...text].map(c => c === c.toUpperCase() ? c.toLowerCase() : c.toUpperCase()).join(''); break
    case 'title': result = text.replace(/\b\w/g, c => c.toUpperCase()); break
    case 'title-blend': result = text.replace(/\b\w/g, c => c.toUpperCase()).replace(/\B\w/g, c => c.toLowerCase()); break
    case 'sentence': result = text.replace(/(^\w|[.!?]\s+\w)/g, c => c.toUpperCase()); break
    case 'sentence-blend': result = text.replace(/(^\w|[.!?]\s+\w)/g, c => c.toUpperCase()); break
    case 'random': result = [...text].map(c => Math.random() > 0.5 ? c.toUpperCase() : c.toLowerCase()).join(''); break
    case 'pascal': result = toPascalWords(text); break
    case 'camel': {
      const p = toPascalWords(text)
      result = p ? p.charAt(0).toLowerCase() + p.slice(1) : p
      break
    }
    case 'snake': result = splitWords(text).join('_'); break
    case 'kebab': result = splitWords(text).join('-'); break
  }
  if (result !== text) {
    const from = isSelection ? selection.from : state.doc.lineAt(selection.from).from
    const to = isSelection ? selection.to : state.doc.lineAt(selection.from).to
    v.dispatch({ changes: { from, to, insert: result } })
  }
}

// ==================== Toggle 类（原版语义） ====================
function toggleWordWrap(enable?: boolean) {
  const v = editorView.value
  if (!v) return
  const cur = v.lineWrapping
  const w = enable !== undefined ? enable : !cur
  v.dispatch({ effects: wordWrapCompartment.reconfigure(w ? EditorView.lineWrapping : []) })
  if (settingStore.config) { settingStore.config.editor.wordWrap = w; settingStore.saveConfig() }
}

function toggleShowWhitespace(show: boolean) {
  const v = editorView.value
  if (!v) return
  const isDark = colors.value.isDark
  const gutterFg = isDark ? '%23666' : '%23bbb'
  const specialChars = show ? EditorView.theme({
    '& .cm-space': { backgroundImage: `url("data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' width='6' height='20'><circle cx='3' cy='12' r='1' fill='${gutterFg}'/></svg>")`, backgroundRepeat: 'no-repeat' },
    '& .cm-tab': { backgroundImage: `url("data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' width='20' height='20'><path d='M2 12 L16 12' stroke='${gutterFg}' stroke-width='0.5' fill='none'/><path d='M14 9 L18 12 L14 15' stroke='${gutterFg}' stroke-width='0.5' fill='none'/></svg>")`, backgroundRepeat: 'no-repeat' },
  }) : EditorView.theme({})
  v.dispatch({ effects: showWhitespaceCompartment.reconfigure(specialChars) })
}

function toggleEol(show: boolean) {
  const v = editorView.value
  if (!v) return
  if (show) v.dom.style.setProperty('--show-eol', "'¶'")
  else v.dom.style.removeProperty('--show-eol')
}

function toggleShowAll() {
  const show = !(settingStore.config?.editor?.showWhitespace)
  toggleShowWhitespace(show)
  toggleEol(show)
  if (settingStore.config) { settingStore.config.editor.showWhitespace = show; settingStore.saveConfig() }
}

function toggleWebAddr() {
  const v = editorView.value
  if (!v) return
  const on = !(settingStore.config?.ui?.showWebAddr)
  if (settingStore.config) { settingStore.config.ui.showWebAddr = on; settingStore.saveConfig() }
  v.dispatch({ effects: webAddrCompartment.reconfigure(on ? [webAddrField] : []) })
}

// ==================== Marks / Highlight ====================
function markAll(term: string) {
  const v = editorView.value
  if (!v) return
  const re = new RegExp(escapeRegExp(term), 'gi')
  const ranges: MarkRange[] = []
  let m: RegExpExecArray | null
  while ((m = re.exec(v.state.doc.toString())) !== null) {
    if (m[0]) ranges.push({ from: m.index, to: m.index + m[0].length, color: currentMarkColor })
    if (m.index === re.lastIndex) re.lastIndex++
  }
  v.dispatch({ effects: setMarks.of(ranges) })
}
function markSelectionOrWord() {
  const v = editorView.value
  if (!v) return
  const sel = v.state.selection.main
  if (sel.from !== sel.to) {
    v.dispatch({ effects: addMarkRanges.of([{ from: sel.from, to: sel.to, color: currentMarkColor }]) })
  } else {
    const word = v.state.wordAt(sel.head)
    if (word) v.dispatch({ effects: addMarkRanges.of([{ from: word.from, to: word.to, color: currentMarkColor }]) })
  }
}
function markKeywords(keywords: string[]) {
  const v = editorView.value
  if (!v) return
  const text = v.state.doc.toString()
  const ranges: MarkRange[] = []
  for (const kw of keywords) {
    if (!kw) continue
    const re = new RegExp(escapeRegExp(kw), 'gi')
    let m: RegExpExecArray | null
    while ((m = re.exec(text)) !== null) {
      if (m[0]) ranges.push({ from: m.index, to: m.index + m[0].length, color: currentMarkColor })
      if (m.index === re.lastIndex) re.lastIndex++
    }
  }
  v.dispatch({ effects: setMarks.of(ranges) })
}
function highlightRanges(idxs: number[], len: number) {
  const v = editorView.value
  if (!v) return
  const ranges: MarkRange[] = idxs.map(i => ({ from: i, to: i + len, color: currentMarkColor }))
  v.dispatch({ effects: setMarks.of(ranges) })
}
function highlightWordAtCursor() {
  const v = editorView.value
  if (!v) return
  const pos = v.state.selection.main.head
  const word = v.state.wordAt(pos)
  if (word) v.dispatch({ effects: addHighlightWord.of({ from: word.from, to: word.to }) })
}
function clearWordHighlight() { editorView.value?.dispatch({ effects: clearHighlightWord.of(null) }) }
function clearAllMarks() { editorView.value?.dispatch({ effects: clearMarksEffect.of(null) }) }

// ==================== 查找（原版 doFindAction） ====================
function doFindAction(dir: 'next' | 'prev') {
  const v = editorView.value
  if (!v || !lastSearchTerm) return
  const state = v.state
  const doc = state.doc.toString()
  const from = state.selection.main.head
  const regex = new RegExp(escapeRegExp(lastSearchTerm), 'gi')
  let target: number | null = null
  if (dir === 'next') {
    const after = doc.slice(from).match(regex)
    target = after ? from + (after.index ?? 0) : (doc.match(regex)?.index ?? null)
    if (target !== null && target < from) target = doc.match(regex)?.index ?? null
  } else {
    const matches = [...doc.slice(0, from).matchAll(new RegExp(escapeRegExp(lastSearchTerm), 'gi'))]
    target = matches.length > 0 ? (matches[matches.length - 1].index ?? null) : ([...doc.matchAll(new RegExp(escapeRegExp(lastSearchTerm), 'gi'))].pop()?.index ?? null)
  }
  if (target !== null) {
    v.dispatch({ selection: { anchor: target, head: target + lastSearchTerm.length }, scrollIntoView: true })
    v.focus()
  }
}

// ==================== JSON / XML 格式化（原版实现） ====================
async function formatJsonSelection() {
  const v = editorView.value
  if (!v) return
  const { from, to } = v.state.selection.main
  const text = from === to ? v.state.doc.toString() : v.state.sliceDoc(from, to)
  try {
    const r = await FormatJSON(text, 2)
    if (r && r.success && r.content) {
      if (from === to) v.dispatch({ changes: { from: 0, to: v.state.doc.length, insert: r.content } })
      else v.dispatch({ changes: { from, to, insert: r.content } })
      ElMessage.success('JSON 格式化成功')
    } else if (r && !r.success) {
      ElMessage.error(`JSON 格式化失败: ${r.error?.message || '未知错误'}`)
    }
  } catch (e: any) { ElMessage.error(e?.message || 'JSON 格式化失败') }
}
function prettyPrintXml(xml: string): string {
  try {
    const parser = new DOMParser()
    const doc = parser.parseFromString(xml, 'application/xml')
    if (doc.querySelector('parsererror')) return ''
    const serialize = (node: Element, level: number): string => {
      const pad = '  '.repeat(level)
      let out = `${pad}<${node.nodeName}`
      for (const attr of Array.from(node.attributes)) out += ` ${attr.name}="${attr.value}"`
      const text = (node.textContent || '').trim()
      if (node.children.length === 0) {
        out += `>${text}</${node.nodeName}>`
      } else {
        out += '>\n'
        for (const child of Array.from(node.children)) out += serialize(child, level + 1) + '\n'
        out += `${pad}</${node.nodeName}>`
      }
      return out
    }
    const root = doc.documentElement
    return '<?xml version="1.0" encoding="UTF-8"?>\n' + serialize(root, 0)
  } catch { return '' }
}
function formatXmlSelection() {
  const v = editorView.value
  if (!v) return
  const { from, to } = v.state.selection.main
  const text = from === to ? v.state.doc.toString() : v.state.sliceDoc(from, to)
  const formatted = prettyPrintXml(text)
  if (!formatted) { ElMessage?.warning?.('XML 格式化失败'); return }
  if (from === to) v.dispatch({ changes: { from: 0, to: v.state.doc.length, insert: formatted } })
  else v.dispatch({ changes: { from, to, insert: formatted } })
}
async function minifyJsonSelection() {
  const v = editorView.value
  if (!v) return
  const { from, to } = v.state.selection.main
  const text = from === to ? v.state.doc.toString() : v.state.sliceDoc(from, to)
  try {
    const r = await MinifyJSON(text)
    if (r && r.success && r.content) {
      if (from === to) v.dispatch({ changes: { from: 0, to: v.state.doc.length, insert: r.content } })
      else v.dispatch({ changes: { from, to, insert: r.content } })
      ElMessage.success('JSON 压缩成功')
    } else if (r && !r.success) {
      ElMessage.error(`JSON 压缩失败: ${r.error?.message || '未知错误'}`)
    }
  } catch (e: any) { ElMessage.error(e?.message || 'JSON 压缩失败') }
}
async function validateJsonSelection() {
  const v = editorView.value
  if (!v) return
  const { from, to } = v.state.selection.main
  const text = from === to ? v.state.doc.toString() : v.state.sliceDoc(from, to)
  try {
    const r = await ValidateJSON(text)
    if (r) {
      if (r.success) ElMessage.success('JSON 格式正确')
      else ElMessage.error(`JSON 格式错误: ${r.error?.message || r.error || '未知错误'}`)
    }
  } catch (e: any) { ElMessage.error(e?.message || 'JSON 校验失败') }
}

// ==================== 括号 / 位置历史（原版实现） ====================
function gotoBracket() {
  const v = editorView.value
  if (!v) return
  const head = v.state.selection.main.head
  const doc = v.state.doc.toString()
  const pairs: Record<string, string> = { '(': ')', '[': ']', '{': '}', ')': '(', ']': '[', '}': '{' }
  const open = '([{'
  const close = ')]}'
  for (let off = 0; off <= 1; off++) {
    const ch = doc[head + off]
    if (ch && (open.includes(ch) || close.includes(ch))) {
      const forward = open.includes(ch)
      const target = pairs[ch]
      let depth = 0, i = head + off
      if (forward) {
        for (i = head + off + 1; i < doc.length; i++) {
          if (doc[i] === ch) depth++
          else if (doc[i] === target) { if (depth === 0) break; depth-- }
        }
      } else {
        for (i = head + off - 1; i >= 0; i--) {
          if (doc[i] === ch) depth++
          else if (doc[i] === target) { if (depth === 0) break; depth-- }
        }
      }
      if (i >= 0 && i < doc.length) {
        v.dispatch({ selection: { anchor: i }, scrollIntoView: true })
        v.focus()
      }
      return
    }
  }
}

// ---- Position history（原版：pushPosition + goBack/ForwardPosition） ----
let lastPushedPos = -1
let lastPushedLine = -1
let posTimer: any = null
function recordPosition(pos: number) {
  const v = editorView.value
  if (!v) return
  const line = v.state.doc.lineAt(pos).number
  clearTimeout(posTimer)
  posTimer = setTimeout(() => {
    if (Math.abs(line - lastPushedLine) >= 5 || lastPushedLine < 0) {
      editorStore.pushPosition(props.tab.id, pos, 0)
      lastPushedPos = pos; lastPushedLine = line
    }
  }, 350)
}
function gotoPrevPosition() {
  const r = editorStore.goBackPosition(props.tab.id)
  if (r && editorView.value) { editorView.value.dispatch({ selection: { anchor: r.pos }, scrollIntoView: true }) }
}
function gotoNextPosition() {
  const r = editorStore.goForwardPosition(props.tab.id)
  if (r && editorView.value) { editorView.value.dispatch({ selection: { anchor: r.pos }, scrollIntoView: true }) }
}

// ==================== 编辑器创建（原版装配顺序） ====================
function createEditor() {
  if (!editorContainer.value) return
  isInitializing = true
  if (editorView.value) { editorView.value.destroy(); editorView.value = null }

  const langExtensions = language.getLanguageExtension(props.tab.language)
  const isDark = colors.value.isDark
  const ed = config.value?.editor

  if (props.tab.language === 'markdown') markdown.initMermaid(isDark)

  const state = EditorState.create({
    doc: props.tab.content,
    extensions: [
      lineNumbers(),
      highlightActiveLineGutter(),
      highlightSpecialChars(),
      history(),
      foldGutter({}),
      dropCursor(),
      EditorState.allowMultipleSelections.of(true),
      indentOnInput(),
      syntaxHighlightCompartment.of(theme.buildSyntaxHighlight()),
      bracketMatching(),
      closeBrackets(),
      autocompletion({
        override: [completion.snippetCompletionSource as any],
      }),
      rectangularSelection(),
      crosshairCursor(),
      highlightActiveLine(),
      keymap.of([
        ...closeBracketsKeymap,
        ...defaultKeymap,
        ...historyKeymap,
        ...foldKeymap,
        ...completionKeymap,
        ...lintKeymap,
        indentWithTab,
      ]),
      langCompartment.of(langExtensions),
      foldCompartment.of(theme.buildFoldGutter(isDark)),
      appearanceCompartment.of(theme.buildAppearanceTheme()),
      wordWrapCompartment.of(ed?.wordWrap ? EditorView.lineWrapping : []),
      showWhitespaceCompartment.of(EditorView.theme({})),
      tabSizeCompartment.of(EditorState.tabSize.of(ed?.tabSize || 4)),
      webAddrCompartment.of(settingStore.config?.ui?.showWebAddr ? [webAddrField] : []),
      markField,
      highlightWordField,
      EditorView.updateListener.of((update) => {
        if (update.docChanged && !isInitializing) {
          editorStore.updateTabContent(props.tab.id, update.state.doc.toString())
          macro.recordMacroStep(update, editorView.value as any, isInitializing)
        }
        if (update.selectionSet) {
          const pos = update.state.selection.main.head
          const line = update.state.doc.lineAt(pos)
          editorStore.updateCursorPosition(props.tab.id, line.number, pos - line.from + 1)
          recordPosition(pos)
        }
      }),
    ],
  })

  const view = new EditorView({ state, parent: editorContainer.value })
  editorView.value = view
  Promise.resolve().then(() => { isInitializing = false })

  // 文档地图：滚动时同步视口（rAF 节流，原版实现）
  const scroller = view.scrollDOM
  let minimapRaf = 0
  const syncMinimap = () => {
    minimapRaf = 0
    minimapViewport.value = {
      scrollTop: scroller.scrollTop,
      scrollHeight: scroller.scrollHeight,
      clientHeight: scroller.clientHeight,
    }
  }
  scroller.addEventListener('scroll', () => {
    if (!minimapRaf) minimapRaf = requestAnimationFrame(syncMinimap)
  }, { passive: true })
  syncMinimap()

  // 恢复上次光标位置（原版实现）
  if (props.tab.cursorPosition.line > 1) {
    try {
      const line = state.doc.line(props.tab.cursorPosition.line)
      view.dispatch({ selection: { anchor: line.from + props.tab.cursorPosition.column - 1 }, scrollIntoView: true })
    } catch (e) { console.warn(e) }
  }

  // bug #1 修复：书签从 DB 同步（tab 打开时）
  bookmark.scheduleSyncFromDB(props.tab)
}

// ==================== Tab 切换 / 状态保存（原版 in-place swap） ====================
function saveEditorState() {
  const v = editorView.value
  if (!v) return
  const pos = v.state.selection.main.head
  const line = v.state.doc.lineAt(pos)
  editorStore.updateCursorPosition(props.tab.id, line.number, pos - line.from + 1)
  editorStore.updateScrollPosition(props.tab.id, v.scrollDOM.scrollTop, v.scrollDOM.scrollLeft)
}

function restoreEditorState() {
  const v = editorView.value
  if (!v) return
  const doc = v.state.doc
  const savedLine = props.tab.cursorPosition?.line || 1
  const savedCol = props.tab.cursorPosition?.column || 1
  const line = doc.line(Math.min(savedLine, doc.lines))
  const pos = line.from + Math.min(Math.max(savedCol - 1, 0), line.length)
  v.dispatch({ selection: { anchor: pos }, scrollIntoView: true })
  if (props.tab.scrollPosition?.top) {
    v.scrollDOM.scrollTo({
      top: props.tab.scrollPosition.top,
      left: props.tab.scrollPosition.left || 0,
    })
  }
}

function updateLanguageForTab() {
  const v = editorView.value
  if (!v) return
  v.dispatch({ effects: langCompartment.reconfigure(language.getLanguageExtension(props.tab.language)) })
}

function reconfigureAppearance() {
  const v = editorView.value
  if (!v) return
  v.dispatch({
    effects: [
      appearanceCompartment.reconfigure(theme.buildAppearanceTheme()),
      syntaxHighlightCompartment.reconfigure(theme.buildSyntaxHighlight()),
      foldCompartment.reconfigure(theme.buildFoldGutter(colors.value.isDark)),
    ],
  })
}

// ==================== 主命令分发 ====================
function handleEditorCommand(e: Event) {
  const detail = (e as CustomEvent).detail
  if (!detail) return
  let cmd: string, args: any[] = []
  if (typeof detail === 'string') cmd = detail
  else { cmd = detail.cmd; args = detail.args || [] }
  if (!cmd) return
  cmd = CMD_ALIASES[cmd] || cmd

  // 查找词同步（F3 依赖）与批量高亮
  if (cmd === 'set-search-term') { setSearchTerm(String(args[0] ?? '')); return }
  if (cmd === 'highlight-all') { highlightRanges(args[0] as number[], Number(args[1] ?? 0)); return }

  // 基础编辑
  if (cmd === 'undo') { undoAction(); return }
  if (cmd === 'redo') { redoAction(); return }
  if (cmd === 'cut') { doEditorAction('cut'); return }
  if (cmd === 'copy') { doEditorAction('copy'); return }
  if (cmd === 'paste') { doEditorAction('paste'); return }
  if (cmd === 'select-all') { doEditorAction('selectAll'); return }
  if (cmd === 'delete') { deleteSelectionOrChar(); return }
  if (cmd === 'line-cut') { cutCurrentLine(); return }
  if (cmd === 'copy-line') { lineOperation('duplicate'); return }

  // 查找 / 跳转
  if (cmd === 'find') { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'find' })); return }
  if (cmd === 'replace') { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'replace' })); return }
  if (cmd === 'find-next') { doFindAction('next'); return }
  if (cmd === 'find-prev') { doFindAction('prev'); return }
  if (cmd === 'show-goto-line') { showGotoLineDialog(); return }
  if (cmd === 'goto-line' && args[0]) { gotoLine(args[0] as number); return }
  if (cmd === 'scroll-to-pos' && args[0] && editorView.value) { editorView.value.dispatch({ selection: { anchor: args[0] }, scrollIntoView: true }); return }
  if (cmd === 'scroll-to-line' && args[0]) { gotoLine(args[0] as number); return }
  if (cmd === 'scroll-to-end' && editorView.value) { editorView.value.dispatch({ selection: { anchor: editorView.value.state.doc.length }, scrollIntoView: true }); return }
  if (cmd === 'goto-bracket') { gotoBracket(); return }
  if (cmd === 'prev-position') { gotoPrevPosition(); return }
  if (cmd === 'next-position') { gotoNextPosition(); return }

  // 行 / 文本变换
  if (cmd.startsWith('case-')) { transformCase(cmd.replace('case-', '')); return }
  if (cmd.startsWith('line-')) { lineOperation(cmd.replace('line-', '')); return }
  if (cmd.startsWith('sort-')) { sortLines(cmd.replace('sort-', '')); return }
  if (cmd.startsWith('trim-')) { trimWhitespace(cmd.replace('trim-', '')); return }
  if (cmd === 'tab-to-spaces') { convertTabsSpaces('tabToSpaces'); return }
  if (cmd === 'spaces-all-to-tabs') { convertTabsSpaces('spacesAllToTabs'); return }
  if (cmd === 'spaces-leading-to-tabs') { convertTabsSpaces('spacesLeadingToTabs'); return }
  if (cmd === 'indent') { indentLines(); return }
  if (cmd === 'dedent') { dedentLines(); return }
  if (cmd === 'insert-blank-above') { insertBlankLine(true); return }
  if (cmd === 'insert-blank-below') { insertBlankLine(false); return }
  if ((cmd === 'insert-text' || cmd === 'insert-snippet') && args[0]) {
    editorView.value?.dispatch({ changes: { from: editorView.value.state.selection.main.head, insert: String(args[0]) } })
    return
  }

  // 注释
  if (cmd === 'comment-line') { if (editorView.value) toggleComment(editorView.value); return }
  if (cmd === 'comment-block') { if (editorView.value) toggleBlockComment(editorView.value); return }

  // 自动换行 / 空白 / 行尾 / web 地址 / minimap
  if (cmd === 'wordwrap-on') { toggleWordWrap(true); return }
  if (cmd === 'wordwrap-off') { toggleWordWrap(false); return }
  if (cmd === 'toggle-word-wrap') { toggleWordWrap(); return }
  if (cmd === 'show-whitespace') { toggleShowWhitespace(true); return }
  if (cmd === 'hide-whitespace') { toggleShowWhitespace(false); return }
  if (cmd === 'toggle-whitespace') {
    const on = !(settingStore.config?.editor?.showWhitespace)
    toggleShowWhitespace(on)
    if (settingStore.config) { settingStore.config.editor.showWhitespace = on; settingStore.saveConfig() }
    return
  }
  if (cmd === 'show-eol') { toggleEol(true); return }
  if (cmd === 'hide-eol') { toggleEol(false); return }
  if (cmd === 'toggle-eol') {
    const on = !(settingStore.config?.editor?.showEol)
    toggleEol(on)
    if (settingStore.config) { settingStore.config.editor.showEol = on; settingStore.saveConfig() }
    return
  }
  if (cmd === 'show-all') { toggleShowAll(); return }
  if (cmd === 'toggle-webaddr') { toggleWebAddr(); return }
  if (cmd === 'toggle-minimap') { showMinimap.value = !showMinimap.value; return }

  // 书签（useEditorBookmark）
  if (cmd === 'toggle-bookmark') { bookmark.toggleBookmark(editorView.value as any, props.tab); return }
  if (cmd === 'next-bookmark') { bookmark.gotoNextBookmark(editorView.value as any, props.tab, gotoLine); return }
  if (cmd === 'prev-bookmark') { bookmark.gotoPrevBookmark(editorView.value as any, props.tab, gotoLine); return }
  if (cmd === 'clear-bookmarks') { bookmark.clearAllBookmarks(props.tab); return }
  if (cmd === 'copy-bookmark-lines') { bookmark.copyBookmarkLines(editorView.value as any, props.tab); return }
  if (cmd === 'cut-bookmark-lines') { bookmark.cutBookmarkLines(editorView.value as any, props.tab); return }
  if (cmd === 'delete-bookmark-lines') { bookmark.deleteBookmarkLines(editorView.value as any, props.tab); return }
  if (cmd === 'delete-unbookmark-lines') { bookmark.deleteUnbookmarkLines(editorView.value as any, props.tab); return }
  if (cmd === 'paste-bookmark-lines') { bookmark.pasteBookmarkLines(editorView.value as any, props.tab); return }

  // 标记 / 高亮
  if (cmd === 'clear-mark' || cmd === 'clear-highlight' || cmd === 'clear-all-highlight') { clearWordHighlight(); return }
  if (cmd === 'clear-all-marks' || cmd === 'clear-marks') { clearAllMarks(); return }
  if (cmd === 'word-highlight') { highlightWordAtCursor(); return }
  if (cmd === 'mark-color') { highlightWordAtCursor(); return }
  if (cmd === 'mark-all') {
    if (Array.isArray(args[0])) highlightRanges(args[0] as number[], Number(args[1] ?? 0))
    else markAll(String(args[0] ?? lastSearchTerm))
    return
  }
  if (cmd === 'mark-keywords' && args[0]) { markKeywords(args[0] as string[]); return }
  if (cmd === 'mark-red') { currentMarkColor = 1; markSelectionOrWord(); return }
  if (cmd === 'mark-yellow') { currentMarkColor = 0; markSelectionOrWord(); return }
  if (cmd === 'mark-blue') { currentMarkColor = 2; markSelectionOrWord(); return }
  if (cmd === 'mark-1') { currentMarkColor = 0; markSelectionOrWord(); return }
  if (cmd === 'mark-2') { currentMarkColor = 1; markSelectionOrWord(); return }
  if (cmd === 'mark-3') { currentMarkColor = 2; markSelectionOrWord(); return }
  if (cmd === 'mark-4') { currentMarkColor = 3; markSelectionOrWord(); return }
  if (cmd === 'mark-5') { currentMarkColor = 4; markSelectionOrWord(); return }
  if (cmd === 'mark-loop') { currentMarkColor = (currentMarkColor + 1) % 5; markSelectionOrWord(); return }

  // 格式化
  if (cmd === 'format-json') { formatJsonSelection(); return }
  if (cmd === 'format-xml') { formatXmlSelection(); return }
  if (cmd === 'minify-json') { minifyJsonSelection(); return }
  if (cmd === 'validate-json') { validateJsonSelection(); return }

  // 列块 / 宏 / MD
  if (cmd === 'column-mode') { columnMode.toggleColumnMode(); return }
  if (cmd === 'column-block') { columnMode.enterColumnMode(); return }
  if (cmd === 'toggle-md-mode') { markdown.toggleMdMode(); return }

  // 列编辑插入（ColumnEditWin 派发）
  if (cmd === 'column-insert-text' && args[0]) {
    const v = editorView.value
    if (!v) return
    const text = String(args[0])
    const ranges = [...v.state.selection.ranges].sort((a, b) => b.from - a.from)
    if (ranges.length <= 1) {
      v.dispatch({ changes: { from: v.state.selection.main.head, insert: text } })
    } else {
      v.dispatch({ changes: ranges.map(r => ({ from: r.from, insert: text })) })
    }
    return
  }
  if (cmd === 'column-insert-num' && args[0]) {
    const opts = args[0]
    let val = opts.init
    const lines: string[] = []
    for (let i = 0; i < opts.repeat; i++) {
      let s = val.toString(opts.radix || 10)
      if (opts.radix === 16 && opts.capital) s = s.toUpperCase()
      lines.push((opts.prefix || '') + s)
      val += opts.inc || 1
    }
    editorView.value?.dispatch({ changes: { from: editorView.value.state.selection.main.head, insert: lines.join('\n') } })
    return
  }
}

// ==================== Minimap ====================
const showMinimap = ref(false)
const minimapViewport = ref({ scrollTop: 0, scrollHeight: 0, clientHeight: 0 })

// ==================== Tab 切换（原版 in-place swap） ====================
watch(() => props.tab.id, (newId, oldId) => {
  if (!editorView.value) {
    createEditor()
    return
  }
  if (oldId) saveEditorState()
  isInitializing = true
  editorView.value.dispatch({
    changes: { from: 0, to: editorView.value.state.doc.length, insert: props.tab.content },
  })
  updateLanguageForTab()
  restoreEditorState()
  // bug #1 修复：新 tab 的书签从 DB 同步
  bookmark.scheduleSyncFromDB(props.tab)
  nextTick(() => {
    isInitializing = false
    editorView.value?.focus()
  })
})

// 外部内容变化（tail-f / reloadAsText / 外部重载）→ 同步到编辑器（原版行为）
watch(() => props.tab.content, (nc) => {
  const v = editorView.value
  if (!v) return
  const current = v.state.doc.toString()
  if (current !== nc) {
    isInitializing = true
    v.dispatch({ changes: { from: 0, to: current.length, insert: nc } })
    queueMicrotask(() => { isInitializing = false })
  }
})

watch(() => props.tab.path, () => {
  // bug #1 修复：文件重命名 / 重新打开后重新同步书签
  bookmark.scheduleSyncFromDB(props.tab)
})

watch(() => props.tab.language, () => {
  updateLanguageForTab()
  if (props.tab.language === 'markdown') markdown.initMermaid(colors.value.isDark)
})

watch(() => settingStore.config?.theme?.currentTheme, () => reconfigureAppearance())
watch(() => config.value?.editor?.tabSize, (s) => { if (editorView.value && s) editorView.value.dispatch({ effects: tabSizeCompartment.reconfigure(EditorState.tabSize.of(s)) }) })
watch(() => config.value?.editor?.wordWrap, (w) => { if (w !== undefined) toggleWordWrap(w) })
watch(() => config.value?.editor?.fontSize, () => reconfigureAppearance())
watch(() => config.value?.editor?.fontFamily, () => reconfigureAppearance())
watch(() => config.value?.ui?.zoomLevel, () => reconfigureAppearance())
watch(() => settingStore.config?.ui?.showWebAddr, () => toggleWebAddr())

// bug #2 修复联动：snippet 数量变化 → 失效补全缓存（在 useEditorCompletion 内部 watch）

// 宏回放
watch(() => editorStore.macroState.isPlaying, (playing) => {
  if (playing) macro.playMacro(editorView.value as any)
})

// ==================== 生命周期 ====================
onMounted(() => {
  createEditor()
  document.addEventListener('editor-command', handleEditorCommand as EventListener)
})

onBeforeUnmount(() => {
  saveEditorState()
  if (editorView.value) { editorView.value.destroy(); editorView.value = null }
  document.removeEventListener('editor-command', handleEditorCommand as EventListener)
  if (posTimer) clearTimeout(posTimer)
})

// ==================== 右键菜单 dispatch ====================
function handleContextMenuDispatch(cmd: string) {
  const dispatcher = contextMenu.run((c) => handleEditorCommand(new CustomEvent('editor-command', { detail: { cmd: c } })))
  return dispatcher(cmd)
}

// ==================== 暴露给父组件 ====================
defineExpose({
  gotoLine, showGotoLineDialog, toggleWordWrap, toggleShowWhitespace,
  transformCase, lineOperation, convertTabsSpaces, trimWhitespace, sortLines,
  indentLines, dedentLines, getLineCount, getSelectedText,
  formatJsonSelection, formatXmlSelection, minifyJsonSelection, validateJsonSelection,
  highlightWordAtCursor, clearWordHighlight, markAll, markKeywords, highlightRanges,
  toggleMdMode: markdown.toggleMdMode,
  enterColumnMode: columnMode.enterColumnMode,
  toggleColumnMode: columnMode.toggleColumnMode,
  exportMdHtml: () => markdown.exportMdHtml(props.tab),
  handlePreviewClick: markdown.handlePreviewClick,
})
</script>

<template>
  <div class="h-full flex flex-col overflow-hidden">
    <!-- MD toolbar -->
    <div v-if="markdown.isMarkdown.value" class="flex items-center gap-1 px-3 py-1 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#2d2d2d] flex-shrink-0">
      <span class="text-xs text-gray-400 mr-2">Markdown</span>
      <button class="px-2 py-0.5 text-xs rounded" :class="markdown.mdMode.value==='edit'?'bg-blue-500 text-white':'text-gray-500 hover:bg-gray-200 dark:hover:bg-gray-600 dark:text-gray-400'" @click="markdown.mdMode.value='edit'">编辑</button>
      <button class="px-2 py-0.5 text-xs rounded" :class="markdown.mdMode.value==='split'?'bg-blue-500 text-white':'text-gray-500 hover:bg-gray-200 dark:hover:bg-gray-600 dark:text-gray-400'" @click="markdown.mdMode.value='split'">分屏</button>
      <button class="px-2 py-0.5 text-xs rounded" :class="markdown.mdMode.value==='preview'?'bg-blue-500 text-white':'text-gray-500 hover:bg-gray-200 dark:hover:bg-gray-600 dark:text-gray-400'" @click="markdown.mdMode.value='preview'">预览</button>
      <span class="flex-1"></span>
      <button class="px-2 py-0.5 text-xs rounded text-gray-500 hover:bg-gray-200 dark:hover:bg-gray-600 dark:text-gray-400" @click="markdown.exportMdHtml(props.tab)">导出 HTML</button>
    </div>

    <div class="flex-1 flex overflow-hidden" style="position:relative;" @contextmenu="contextMenu.open($event, editorView)" @click="contextMenu.close()">
      <div ref="editorContainer" :class="{'w-full':!markdown.isMarkdown.value||markdown.mdMode.value==='edit','cm-container-split border-r border-gray-200 dark:border-gray-700':markdown.isMarkdown.value&&markdown.mdMode.value==='split','hidden':markdown.isMarkdown.value&&markdown.mdMode.value==='preview','with-minimap':showMinimap}" class="cm-container" @dblclick="highlightWordAtCursor"></div>

      <!-- 右键菜单（数据驱动：ext-context-menu） -->
      <Teleport to="body">
        <div v-if="contextMenu.visible.value"
          class="fixed z-[9999] bg-white dark:bg-[#2d2d2d] rounded-lg shadow-2xl border border-gray-200 dark:border-gray-600 py-1 min-w-[200px] context-menu-panel"
          :style="{ left: contextMenu.x.value + 'px', top: contextMenu.y.value + 'px' }"
          @click.stop
          @contextmenu.prevent
        >
          <template v-for="(section, si) in CONTEXT_MENU_SECTIONS" :key="si">
            <div v-if="si > 0" class="context-menu-separator"></div>
            <button v-for="item in section.items" :key="item.cmd"
              class="context-menu-item"
              :disabled="item.needSelection && !contextMenu.hasSelection.value"
              @click="handleContextMenuDispatch(item.cmd)"
            >
              <span class="item-label">{{ item.label }}</span>
              <span v-if="item.shortcut" class="item-shortcut">{{ item.shortcut }}</span>
            </button>
          </template>
        </div>
      </Teleport>

      <!-- Goto line dialog -->
      <Teleport to="body">
        <div v-if="showGotoLine" class="fixed inset-0 z-50 flex items-center justify-center bg-black/20" @click.self="showGotoLine=false" @keydown.enter="handleGotoLine" @keydown.escape="showGotoLine=false">
          <div class="bg-white dark:bg-[#2d2d2d] rounded-lg shadow-2xl p-6 w-80 border border-gray-200 dark:border-gray-600">
            <h3 class="text-sm font-medium mb-4 dark:text-gray-200">跳转到行</h3>
            <div class="flex gap-2">
              <input v-model.number="gotoLineInput" type="number" min="1" :max="getLineCount()" placeholder="输入行号..." class="goto-line-input flex-1 px-3 py-1.5 text-sm bg-gray-50 dark:bg-[#3c3c3c] border border-gray-200 dark:border-gray-600 rounded focus:outline-none focus:border-blue-500 dark:text-gray-200" @keydown.enter="handleGotoLine" @keydown.escape="showGotoLine=false"/>
              <button class="px-3 py-1.5 text-sm bg-blue-500 text-white rounded hover:bg-blue-600" @click="handleGotoLine">跳转</button>
            </div>
            <p class="text-xs text-gray-400 mt-2">总行数: {{ getLineCount() }}</p>
          </div>
        </div>
      </Teleport>

      <!-- MD preview -->
      <div v-if="markdown.isMarkdown.value && (markdown.mdMode.value==='preview' || markdown.mdMode.value==='split')" class="overflow-auto p-4 bg-white dark:bg-[#1e1e1e] select-text" :class="markdown.mdMode.value==='split' ? 'w-1/2' : 'w-full'" @click="markdown.handlePreviewClick">
        <div class="markdown-body" v-html="markdown.renderedHtml.value"></div>
      </div>

      <!-- Minimap -->
      <Minimap v-if="showMinimap" :content="props.tab.content" :viewport="minimapViewport" @seek="gotoLine" />
    </div>
  </div>
</template>

<style scoped>
/* ---- 右键菜单样式（token 化） ---- */
.context-menu-panel {
  max-height: calc(100vh - 8px);
  overflow-y: auto;
  box-shadow: 0 8px 30px rgba(0,0,0,.18), 0 0 1px rgba(0,0,0,.12);
}
.context-menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 6px 16px;
  font-size: var(--et-text-md, 13px);
  color: var(--et-fg, #1f2328);
  background: transparent;
  border: none;
  cursor: pointer;
  text-align: left;
  transition: background 0.1s;
}
.context-menu-item:hover:not(:disabled) {
  background: var(--et-bg-hover, #e8e8e8);
}
.context-menu-item:disabled {
  color: var(--et-fg-subtle, #ccc);
  cursor: not-allowed;
}
.context-menu-item .item-label { flex: 1; }
.context-menu-item .item-shortcut {
  font-size: var(--et-text-xs, 11px);
  color: var(--et-fg-subtle, #999);
  margin-left: 20px;
}
.context-menu-separator {
  height: 1px;
  margin: 4px 0;
  background-color: var(--et-border, #e5e5e5);
}
</style>