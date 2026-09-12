<script lang="ts" setup>
/**
 * CodeEditor v3.0（M3 重塑后）
 *
 * 架构：
 *   - 容器（~600 行）：createEditor / dispatch / 模板
 *   - composables/：useEditorTheme / useEditorLanguage / useEditorCompletion /
 *                    useEditorBookmark / useEditorColumnMode / useEditorMacro /
 *                    useEditorMarkdown
 *   - ext/：ext-keymap / ext-context-menu
 *
 * 修复合环 bug：
 *   #1 useEditorBookmark.syncFromDB（tab 切换 / 重命名时拉取后端）
 *   #2 useEditorCompletion.invalidateCompletionCache（监听 snippets 长度变化）
 *
 * 保留原 82 个 dispatch 命令（用于主命令入口 editor-command）。
 */
import { computed, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { EditorTab, MdViewMode } from '@/types'
import { useEditorStore, useSettingStore } from '@/stores'
import { Compartment, EditorState, Prec, StateEffect, StateField, RangeSetBuilder, RangeSet } from '@codemirror/state'
import { EditorView, keymap, Decoration, gutter, GutterMarker, lineNumbers, highlightActiveLine, highlightActiveLineGutter, highlightSpecialChars, rectangularSelection, crosshairCursor, dropCursor, drawSelection } from '@codemirror/view'
import { bracketMatching, indentOnInput } from '@codemirror/language'
import { autocompletion } from '@codemirror/autocomplete'
import { history } from '@codemirror/commands'
import { FormatJSON, Convert, JsonPathQuery, JsonToStruct, JsonStructuredDiff } from '../../wailsjs/go/main/App'
import xmlFormat from 'xml-formatter'
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

// ==================== 编辑器实例与 composable 装配 ====================
const editorContainer = ref<HTMLElement | null>(null)
const editorView = shallowRef<EditorView | null>(null)
let isInitializing = false

// —— 各 composable 装配 ——
const theme = useEditorTheme(colors as any, config as any)
const language = useEditorLanguage()
const completion = useEditorCompletion(
  computed(() => props.tab.language),
  editorStore.snippets as any,
)
const bookmark = useEditorBookmark()
const columnMode = useEditorColumnMode()
const macro = useEditorMacro()
const markdown = useEditorMarkdown(
  computed(() => props.tab.language),
  computed(() => props.tab.content),
  colors as any,
)
const keymapExt = useEditorKeymap()
const contextMenu = useEditorContextMenu()

// —— 编辑器 compartments ——
const appearanceCompartment = new Compartment()
const syntaxHighlightCompartment = new Compartment()
const langCompartment = new Compartment()
const foldCompartment = new Compartment()
const wordWrapCompartment = new Compartment()
const showWhitespaceCompartment = new Compartment()
const webAddrCompartment = new Compartment()

// —— 5 色标记（M3 仍保留 in-memory currentMarkColor；store 同步保留给后续扩展）——
const MARK_COLORS = [
  'rgba(255,212,0,0.45)', 'rgba(255,120,120,0.45)', 'rgba(120,180,255,0.45)',
  'rgba(120,220,150,0.45)', 'rgba(200,140,255,0.45)',
]
let currentMarkColor = 0

// —— 多色 mark StateField ——
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
        const b = new RangeSetBuilder<Decoration>()
        const sorted = [...e.value].sort((a, b) => a.from - b.from)
        for (const r of sorted) {
          if (r.from < r.to) b.add(r.from, r.to, Decoration.mark({ attributes: { style: `background-color:${MARK_COLORS[r.color % MARK_COLORS.length]};border-radius:2px;` } }))
        }
        next = b.finish()
      } else if (e.is(addMarkRanges)) {
        const existing: MarkRange[] = []
        const iter = marks.iter()
        while (iter.value) { existing.push({ from: iter.from, to: iter.to, color: 0 }); iter.next() }
        const merged = [...existing, ...e.value].sort((a, b) => a.from - b.from)
        const b = new RangeSetBuilder<Decoration>()
        for (const r of merged) {
          if (r.from < r.to) b.add(r.from, r.to, Decoration.mark({ attributes: { style: `background-color:${MARK_COLORS[r.color % MARK_COLORS.length]};border-radius:2px;` } }))
        }
        next = b.finish()
      } else if (e.is(clearMarksEffect)) {
        next = RangeSet.empty
      }
    }
    return next
  },
  provide: f => EditorView.decorations.from(f, v => v),
})

// —— URL highlight（view→显示网页地址） ——
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
  const b = new RangeSetBuilder<Decoration>()
  const doc = state.doc.toString()
  for (const m of doc.matchAll(URL_RE)) {
    if (m.index === undefined) continue
    b.add(m.index, m.index + m[0].length, Decoration.mark({ class: 'cm-webaddr' }))
  }
  return b.finish()
}

// —— 高亮当前词（双击触发） ——
const addHighlightWord = StateEffect.define<{ from: number; to: number }>()
const clearHighlightWord = StateEffect.define()
const highlightWordField = StateField.define<RangeSet<Decoration>>({
  create() { return RangeSet.empty },
  update(h, tr) {
    for (const e of tr.effects) {
      if (e.is(addHighlightWord)) {
        const b = new RangeSetBuilder<Decoration>()
        b.add(e.value.from, e.value.to, Decoration.mark({ class: 'cm-word-highlight', attributes: { style: 'background-color:rgba(255,200,0,0.3);border-radius:2px;' } }))
        return b.finish()
      } else if (e.is(clearHighlightWord)) return RangeSet.empty
    }
    if (tr.docChanged) return RangeSet.empty
    return h
  },
  provide: f => EditorView.decorations.from(f, v => v),
})

// ==================== 工具函数 ====================
let lastSearchTerm = ''
function setSearchTerm(term: string) { lastSearchTerm = term }

function escapeRegExp(s: string) { return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') }

// —— 光标定位 ——
function getLineCount() { return editorView.value?.state.doc.lines || 0 }
function getSelectedText() {
  const v = editorView.value
  return v ? v.state.sliceDoc(v.state.selection.main.from, v.state.selection.main.to) : ''
}

function gotoLine(lineNum: number) {
  const v = editorView.value
  if (!v || lineNum < 1) return
  const line = Math.min(lineNum, v.state.doc.lines)
  const pos = v.state.doc.line(line).from
  v.dispatch({ selection: { anchor: pos, head: pos }, scrollIntoView: true })
  v.focus()
}

// —— Goto-line 弹窗 ——
const showGotoLine = ref(false)
const gotoLineInput = ref<number | null>(null)
function showGotoLineDialog() {
  showGotoLine.value = true
  setTimeout(() => {
    const el = document.querySelector('.goto-line-input') as HTMLInputElement | null
    el?.focus()
  }, 50)
}
function handleGotoLine() {
  if (gotoLineInput.value) gotoLine(gotoLineInput.value)
  showGotoLine.value = false
}

// —— 编辑操作：剪切/复制/粘贴/全选/删除 ——
function cutSelection() {
  const v = editorView.value
  if (!v) return
  const { from, to } = v.state.selection.main
  if (from === to) return
  const text = v.state.sliceDoc(from, to)
  editorStore.pushClipboard(text)
  navigator.clipboard?.writeText(text).catch(() => {})
  v.dispatch({ changes: { from, to } })
}
function copySelection() {
  const v = editorView.value
  if (!v) return
  const { from, to } = v.state.selection.main
  if (from === to) return
  editorStore.pushClipboard(v.state.sliceDoc(from, to))
  navigator.clipboard?.writeText(v.state.sliceDoc(from, to)).catch(() => {})
}
async function pasteAtCursor() {
  const v = editorView.value
  if (!v) return
  let clip = editorStore.clipboardHistory[0] || ''
  if (!clip) {
    try { clip = await navigator.clipboard.readText() } catch { clip = '' }
  }
  if (clip) v.dispatch({ changes: { from: v.state.selection.main.head, insert: clip } })
}
function selectAll() {
  const v = editorView.value
  if (!v) return
  v.dispatch({ selection: { anchor: 0, head: v.state.doc.length } })
}
function deleteSelectionOrChar() {
  const v = editorView.value
  if (!v) return
  if (v.state.selection.main.from !== v.state.selection.main.to) {
    v.dispatch({ changes: { from: v.state.selection.main.from, to: v.state.selection.main.to } })
  } else if (v.state.selection.main.head > 0) {
    v.dispatch({ changes: { from: v.state.selection.main.head - 1, to: v.state.selection.main.head } })
  }
}
function undoAction() { editorView.value?.dom.querySelector<HTMLElement>('.cm-content')?.blur(); document.execCommand?.('undo'); }
function redoAction() { document.execCommand?.('redo'); }

// —— 通用行操作 ——
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
      v.dispatch({ changes: { from: toLine.to, insert: '\n' + text } }); break
    }
    case 'remove': {
      const end = toLine.to + 1 > state.doc.length ? state.doc.length : toLine.to + 1
      v.dispatch({ changes: { from: fromLine.from, to: end } }); break
    }
    case 'moveUp': {
      if (fromLine.number <= 1) return
      const prev = state.doc.line(fromLine.number - 1)
      const text = state.sliceDoc(fromLine.from, toLine.to)
      const hasNL = toLine.to < state.doc.length && state.sliceDoc(toLine.to, toLine.to + 1) === '\n'
      v.dispatch({ changes: [{ from: fromLine.from - (prev.length + 1), to: toLine.to + (hasNL ? 1 : 0), insert: text + (hasNL ? '\n' : '') + prev.text }] }); break
    }
    case 'moveDown': {
      if (toLine.number >= state.doc.lines) return
      const next = state.doc.line(toLine.number + 1)
      const text = state.sliceDoc(fromLine.from, toLine.to)
      v.dispatch({ changes: [{ from: fromLine.from, to: next.to, insert: next.text + '\n' + text }] }); break
    }
    case 'removeEmpty': {
      const lines = state.doc.toString().split('\n').filter(l => l.trim() !== '')
      v.dispatch({ changes: { from: 0, to: state.doc.length, insert: lines.join('\n') } }); break
    }
    case 'removeBlank': {
      const lines = state.doc.toString().split('\n').filter(l => l.length > 0)
      v.dispatch({ changes: { from: 0, to: state.doc.length, insert: lines.join('\n') } }); break
    }
    case 'split': {
      if (to > from) v.dispatch({ changes: { from, to, insert: [...state.sliceDoc(from, to)].join('\n') } }); break
    }
    case 'join': {
      if (to > from) v.dispatch({ changes: { from, to, insert: state.sliceDoc(from, to).replace(/\n/g, ' ') } }); break
    }
    case 'removeDuplicate': {
      const lines = state.doc.toString().split('\n')
      const seen = new Set<string>()
      v.dispatch({ changes: { from: 0, to: state.doc.length, insert: lines.filter(l => { if (seen.has(l)) return false; seen.add(l); return true }).join('\n') } }); break
    }
    case 'removeConsecutiveDuplicate': {
      const lines = state.doc.toString().split('\n')
      const r: string[] =[]; let last = ''
      for (const l of lines) { if (l !== last) { r.push(l); last = l } }
      v.dispatch({ changes: { from: 0, to: state.doc.length, insert: r.join('\n') } }); break
    }
    case 'reverse': {
      const lines = state.doc.toString().split('\n')
      v.dispatch({ changes: { from: 0, to: state.doc.length, insert: lines.reverse().join('\n') } }); break
    }
    case 'randomize': {
      const lines = state.doc.toString().split('\n')
      for (let i = lines.length - 1; i > 0; i--) { const j = Math.floor(Math.random() * (i + 1));[lines[i], lines[j]] = [lines[j], lines[i]] }
      v.dispatch({ changes: { from: 0, to: state.doc.length, insert: lines.join('\n') } }); break
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

// —— Tab/空格转换、Trim、Sort、Indent、Case ——
function convertTabsSpaces(type: string) {
  const v = editorView.value
  if (!v) return
  const ts = config.value?.editor?.tabSize || 4
  const text = v.state.doc.toString()
  let result = text
  switch (type) {
    case 'tabToSpaces': result = text.replace(/\t/g, ' '.repeat(ts)); break
    case 'spacesAllToTabs': result = text.replace(new RegExp(` {${ts}}`, 'g'), '\t'); break
    case 'spacesLeadingToTabs':
      result = text.split('\n').map(line => {
        const m = line.match(/^ +/)
        if (!m) return line
        const spaces = m[0].length
        return '\t'.repeat(Math.floor(spaces / ts)) + ' '.repeat(spaces % ts) + line.slice(spaces)
      }).join('\n')
      break
  }
  if (result !== text) v.dispatch({ changes: { from: 0, to: v.state.doc.length, insert: result } })
}

function trimWhitespace(mode: string) {
  const v = editorView.value
  if (!v) return
  const text = v.state.doc.toString()
  let result = text
  switch (mode) {
    case 'head': result = text.split('\n').map(l => l.replace(/^[ \t]+/, '')).join('\n'); break
    case 'tail': result = text.split('\n').map(l => l.replace(/[ \t]+$/, '')).join('\n'); break
    case 'both': result = text.split('\n').map(l => l.replace(/^[ \t]+/, '').replace(/[ \t]+$/, '')).join('\n'); break
  }
  if (result !== text) v.dispatch({ changes: { from: 0, to: v.state.doc.length, insert: result } })
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
  const ts = config.value?.editor?.tabSize || 4
  const useSpaces = config.value?.editor?.insertSpaces ?? true
  const indent = useSpaces ? ' '.repeat(ts) : '\t'
  const fl = v.state.doc.lineAt(v.state.selection.main.from)
  const tl = v.state.doc.lineAt(v.state.selection.main.to)
  const changes = []
  for (let i = fl.number; i <= tl.number; i++) changes.push({ from: v.state.doc.line(i).from, insert: indent })
  v.dispatch({ changes })
}

function dedentLines() {
  const v = editorView.value
  if (!v) return
  const ts = config.value?.editor?.tabSize || 4
  const fl = v.state.doc.lineAt(v.state.selection.main.from)
  const tl = v.state.doc.lineAt(v.state.selection.main.to)
  const changes = []
  for (let i = fl.number; i <= tl.number; i++) {
    const line = v.state.doc.line(i)
    const t = line.text
    if (t.startsWith('\t')) changes.push({ from: line.from, to: line.from + 1 })
    else if (t.startsWith(' '.repeat(ts))) changes.push({ from: line.from, to: line.from + ts })
    else if (t.startsWith(' ')) { const n = t.match(/^ +/)![0].length; changes.push({ from: line.from, to: line.from + Math.min(n, ts) }) }
  }
  if (changes.length > 0) v.dispatch({ changes })
}

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

function cutCurrentLine() {
  const v = editorView.value
  if (!v) return
  const sel = v.state.selection.main
  const fromLine = v.state.doc.lineAt(sel.from)
  const toLine = v.state.doc.lineAt(sel.to)
  const text = v.state.sliceDoc(fromLine.from, toLine.to)
  editorStore.pushClipboard(text)
  const end = toLine.to + 1 > v.state.doc.length ? v.state.doc.length : toLine.to + 1
  v.dispatch({ changes: { from: fromLine.from, to: end } })
}

// —— Toggle：wordWrap / showWhitespace / eol / webAddr ——
function toggleWordWrap(enable?: boolean) {
  const v = editorView.value
  if (!v) return
  const cur = v.lineWrapping
  const w = enable !== undefined ? enable : !cur
  v.dispatch({ effects: wordWrapCompartment.reconfigure(w ? EditorView.lineWrapping :[]) })
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
function toggleShowAll() { toggleShowWhitespace(true) }
function toggleEol(show: boolean) {
  if (!editorView.value) return
  if (show) editorView.value.dom.style.setProperty('--show-eol', '1')
  else editorView.value.dom.style.removeProperty('--show-eol')
}
function toggleWebAddr() {
  const v = editorView.value
  if (!v) return
  const on = !(settingStore.config?.ui?.showWebAddr)
  if (settingStore.config) { settingStore.config.ui.showWebAddr = on; settingStore.saveConfig() }
  v.dispatch({ effects: webAddrCompartment.reconfigure(on ?[webAddrField]:[]) })
}

// —— Marks / Highlight ——
function markAll(term: string) {
  const v = editorView.value
  if (!v) return
  const re = new RegExp(escapeRegExp(term), 'gi')
  const ranges: MarkRange[] = []
  let m: RegExpExecArray | null
  while ((m = re.exec(v.state.doc.toString())) !== null) {
    if (m[0]) ranges.push({ from: m.index, to: m.index + m[0].length, color: currentMarkColor })
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
    const re = new RegExp(escapeRegExp(kw), 'gi')
    let m: RegExpExecArray | null
    while ((m = re.exec(text)) !== null) {
      if (m[0]) ranges.push({ from: m.index, to: m.index + m[0].length, color: currentMarkColor })
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

// —— JSON / XML 格式化 ——
async function formatJsonSelection() {
  const v = editorView.value
  if (!v) return
  const sel = v.state.selection.main
  const text = sel.from !== sel.to ? v.state.sliceDoc(sel.from, sel.to) : v.state.doc.toString()
  if (!text.trim()) return
  try {
    const r = await FormatJSON(JSON.stringify(JSON.parse(text)), 'json', 'json')
    const out = sel.from !== sel.to ? r : JSON.stringify(JSON.parse(text), null, 2)
    v.dispatch({ changes: { from: sel.from, to: sel.to, insert: out } })
  } catch (e: any) { ElMessage.error('JSON 格式化失败: ' + (e?.message || e)) }
}
function formatXmlSelection() {
  const v = editorView.value
  if (!v) return
  const sel = v.state.selection.main
  const text = sel.from !== sel.to ? v.state.sliceDoc(sel.from, sel.to) : v.state.doc.toString()
  if (!text.trim()) return
  try {
    const formatted = xmlFormat(text, { indentation: '  ', collapseContent: true, lineSeparator: '\n' })
    if (sel.from !== sel.to) v.dispatch({ changes: { from: sel.from, to: sel.to, insert: formatted } })
    else v.dispatch({ changes: { from: 0, to: v.state.doc.length, insert: formatted } })
  } catch (e: any) { ElMessage.error('XML 格式化失败: ' + (e?.message || e)) }
}
async function minifyJsonSelection() {
  const v = editorView.value
  if (!v) return
  const sel = v.state.selection.main
  const text = sel.from !== sel.to ? v.state.sliceDoc(sel.from, sel.to) : v.state.doc.toString()
  if (!text.trim()) return
  try {
    const r = await Convert(text, 'json', 'json')
    v.dispatch({ changes: { from: sel.from, to: sel.to, insert: r } })
  } catch (e: any) { ElMessage.error('JSON 压缩失败: ' + (e?.message || e)) }
}
async function validateJsonSelection() {
  const v = editorView.value
  if (!v) return
  const sel = v.state.selection.main
  const text = sel.from !== sel.to ? v.state.sliceDoc(sel.from, sel.to) : v.state.doc.toString()
  try { JSON.parse(text); ElMessage.success('JSON 校验通过') }
  catch (e: any) { ElMessage.error('JSON 校验失败: ' + (e?.message || e)) }
}

// —— Goto bracket / 位置历史 ——
function gotoBracket() {
  const v = editorView.value
  if (!v) return
  const sel = v.state.selection.main
  const text = v.state.doc.toString()
  const pairs: Record<string, string> = { '{': '}', '[': ']', '(': ')', '}': '{', ']': '[', ')': '(' }
  let depth = 1
  const ch = v.state.doc.sliceString(sel.head, sel.head + 1)
  const open = pairs[ch]
  if (open) {
    let i = sel.head + 1
    while (i < text.length) {
      if (text[i] === ch) depth++
      else if (text[i] === open) { depth--; if (depth === 0) { v.dispatch({ selection: { anchor: i + 1 }, scrollIntoView: true }); return } }
      i++
    }
  } else {
    let i = sel.head - 1; depth = 1
    while (i >= 0) {
      if (text[i] === ch) depth++
      else if (text[i] === open) { depth--; if (depth === 0) { v.dispatch({ selection: { anchor: i }, scrollIntoView: true }); return } }
      i--
    }
  }
}

const lastPosLine = { v: -1, l: -1 }
let posTimer: number | null = null
function recordPosition(pos: number) {
  const v = editorView.value
  if (!v) return
  const line = v.state.doc.lineAt(pos).number
  if (posTimer) window.clearTimeout(posTimer)
  posTimer = window.setTimeout(() => {
    if (Math.abs(line - lastPosLine.l) >= 5 || lastPosLine.l < 0) {
      editorStore.pushPosition(props.tab.id, pos, 0)
      lastPosLine.v = pos; lastPosLine.l = line
    }
  }, 350)
}
function gotoPrevPosition() {
  const v = editorView.value
  if (!v) return
  const p = editorStore.popPrevPosition(props.tab.id)
  if (p !== null) v.dispatch({ selection: { anchor: p.pos }, scrollIntoView: true })
}
function gotoNextPosition() {
  const v = editorView.value
  if (!v) return
  const p = editorStore.popNextPosition(props.tab.id)
  if (p !== null) v.dispatch({ selection: { anchor: p.pos }, scrollIntoView: true })
}

// —— Find next / prev ——
function doFindAction(dir: 'next' | 'prev') {
  const v = editorView.value
  if (!v || !lastSearchTerm) return
  const text = v.state.doc.toString()
  const cur = v.state.selection.main.head
  let idx: number
  try {
    if (lastSearchTerm.startsWith('/') && lastSearchTerm.endsWith('/') && lastSearchTerm.length > 2) {
      const re = new RegExp(lastSearchTerm.slice(1, -1), 'g')
      re.lastIndex = cur
      idx = dir === 'next' ? (re.exec(text)?.index ?? -1) : -1
      if (dir === 'prev') {
        const rev = new RegExp(lastSearchTerm.slice(1, -1), 'g')
        let last = -1; let m: RegExpExecArray | null
        while ((m = rev.exec(text)) !== null) { if (m.index < cur) last = m.index; else break }
        idx = last
      }
    } else {
      const lower = text.toLowerCase(); const t = lastSearchTerm.toLowerCase()
      idx = dir === 'next' ? lower.indexOf(t, cur) : lower.lastIndexOf(t, cur - 1)
    }
  } catch { return }
  if (idx >= 0) v.dispatch({ selection: { anchor: idx, head: idx + lastSearchTerm.length }, scrollIntoView: true })
}

// ==================== 编辑器创建 / 销毁 ====================
function createEditor() {
  if (!editorContainer.value) return
  isInitializing = true
  if (editorView.value) { editorView.value.destroy(); editorView.value = null }

  const isDark = colors.value.isDark
  const ed = config.value?.editor
  const langExtensions = language.getLanguageExtension(props.tab.language)

  // markdown mermaid 初始化
  if (props.tab.language === 'markdown') markdown.initMermaid(isDark)

  const state = EditorState.create({
    doc: props.tab.content,
    extensions: [
      lineNumbers(),
      highlightActiveLineGutter(),
      highlightSpecialChars(),
      history(),
      keymapExt.buildPrecKeymap(),
      foldCompartment.of(theme.buildFoldGutter(isDark)),
      dropCursor(),
      EditorState.allowMultipleSelections.of(true),
      indentOnInput(),
      syntaxHighlightCompartment.of(theme.buildSyntaxHighlight()),
      bracketMatching(),
      autocompletion({ override: [completion.snippetCompletionSource as any] }),
      rectangularSelection(),
      crosshairCursor(),
      highlightActiveLine(),
      drawSelection(),
      // Language
      langCompartment.of(langExtensions),
      // Compartment 化外观（字体 / 主题 / web addr / word wrap / show whitespace）
      appearanceCompartment.of(theme.buildAppearanceTheme()),
      showWhitespaceCompartment.of(EditorView.theme({})),
      wordWrapCompartment.of(ed?.wordWrap ? EditorView.lineWrapping :[]),
      webAddrCompartment.of(settingStore.config?.ui?.showWebAddr ?[webAddrField]:[]),
      // 多色 mark / 高亮词 StateField
      markField,
      highlightWordField,
      // updateListener（content 同步 + macro + cursor + position history）
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

  editorView.value = new EditorView({ state, parent: editorContainer.value })
  Promise.resolve().then(() => { isInitializing = false })

  // 书签同步（bug #1 修复）
  bookmark.scheduleSyncFromDB(props.tab)
}

function destroyEditor() {
  if (editorView.value) {
    editorView.value.destroy()
    editorView.value = null
  }
}

// ==================== Tab 切换 / 文件变化 ====================
function saveEditorState() {
  const v = editorView.value
  if (!v) return
  const pos = v.state.selection.main.head
  const line = v.state.doc.lineAt(pos)
  editorStore.updateCursorPosition(props.tab.id, line.number, pos - line.from + 1)
}
function restoreEditorState() {
  const v = editorView.value
  if (!v) return
  const cursor = props.tab.cursorPosition
  if (cursor) {
    const line = Math.max(1, Math.min(cursor.line, v.state.doc.lines))
    const pos = v.state.doc.line(line).from + Math.max(0, cursor.column - 1)
    v.dispatch({ selection: { anchor: pos }, scrollIntoView: true })
  }
}

function updateLanguageForTab() {
  const v = editorView.value
  if (!v) return
  v.dispatch({
    effects: langCompartment.reconfigure(language.getLanguageExtension(props.tab.language)),
  })
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

// ==================== 右键菜单 dispatch ====================
function handleContextMenuDispatch(cmd: string) {
  const dispatcher = contextMenu.run((c) => handleEditorCommand(new CustomEvent('cmd', { detail: c })))
  return dispatcher(cmd)
}

// ==================== 主命令分发（editor-command 事件） ====================
function handleEditorCommand(e: Event) {
  const detail = (e as CustomEvent).detail
  if (!detail) return
  let cmd: string, args: any[] =[]
  if (typeof detail === 'string') cmd = detail
  else { cmd = detail.cmd; args = detail.args ||[] }
  if (!cmd) return
  cmd = CMD_ALIASES[cmd] || cmd

  // 查找词同步 / 高亮（修复 F3 失灵）
  if (cmd === 'set-search-term') { setSearchTerm(String(args[0] ?? '')); return }
  if (cmd === 'highlight-all') { highlightRanges(args[0] as number[], Number(args[1] ?? 0)); return }
  if (cmd === 'show-goto-line') { showGotoLineDialog(); return }

  // 右键菜单命令
  if (cmd === 'cut') { cutSelection(); return }
  if (cmd === 'copy') { copySelection(); return }
  if (cmd === 'paste') { pasteAtCursor(); return }
  if (cmd === 'select-all') { selectAll(); return }
  if (cmd === 'undo') { undoAction(); return }
  if (cmd === 'redo') { redoAction(); return }
  if (cmd === 'find') { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'find' })); return }
  if (cmd === 'replace') { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'replace' })); return }
  if (cmd === 'duplicate') { lineOperation('duplicate'); return }
  if (cmd === 'delete-line') { lineOperation('remove'); return }
  if (cmd === 'move-up') { lineOperation('moveUp'); return }
  if (cmd === 'move-down') { lineOperation('moveDown'); return }
  if (cmd === 'uppercase') { transformCase('upper'); return }
  if (cmd === 'lowercase') { transformCase('lower'); return }
  if (cmd === 'titlecase') { transformCase('title'); return }
  if (cmd === 'format-json') { formatJsonSelection(); return }
  if (cmd === 'format-xml') { formatXmlSelection(); return }
  if (cmd === 'minify-json') { minifyJsonSelection(); return }
  if (cmd === 'validate-json') { validateJsonSelection(); return }
  if (cmd === 'tab-to-spaces') { convertTabsSpaces('tabToSpaces'); return }
  if (cmd === 'spaces-to-tabs') { convertTabsSpaces('spacesLeadingToTabs'); return }
  if (cmd === 'trim-trailing') { trimWhitespace('tail'); return }

  // 通用别名（line-*/sort-*/trim-*）
  if (cmd.startsWith('case-')) { transformCase(cmd.replace('case-', '')); return }
  if (cmd.startsWith('line-')) { lineOperation(cmd.replace('line-', '')); return }
  if (cmd.startsWith('sort-')) { sortLines(cmd.replace('sort-', '')); return }
  if (cmd.startsWith('trim-')) { trimWhitespace(cmd.replace('trim-', '')); return }

  // 编辑行为
  if (cmd === 'delete') { deleteSelectionOrChar(); return }
  if (cmd === 'line-cut') { cutCurrentLine(); return }
  if (cmd === 'copy-line') { lineOperation('duplicate'); return }
  if (cmd === 'find-next') { doFindAction('next'); return }
  if (cmd === 'find-prev') { doFindAction('prev'); return }
  if (cmd === 'wordwrap-on') { toggleWordWrap(true); return }
  if (cmd === 'wordwrap-off') { toggleWordWrap(false); return }
  if (cmd === 'show-whitespace') { toggleShowWhitespace(true); return }
  if (cmd === 'hide-whitespace') { toggleShowWhitespace(false); return }

  // 书签（通过 useEditorBookmark）
  if (cmd === 'toggle-bookmark') { bookmark.toggleBookmark(editorView.value as any, props.tab); return }
  if (cmd === 'next-bookmark') { bookmark.gotoNextBookmark(editorView.value as any, props.tab, gotoLine); return }
  if (cmd === 'prev-bookmark') { bookmark.gotoPrevBookmark(editorView.value as any, props.tab, gotoLine); return }
  if (cmd === 'clear-bookmarks') { bookmark.clearAllBookmarks(props.tab); return }
  if (cmd === 'copy-bookmark-lines') { bookmark.copyBookmarkLines(editorView.value as any, props.tab); return }
  if (cmd === 'cut-bookmark-lines') { bookmark.cutBookmarkLines(editorView.value as any, props.tab); return }
  if (cmd === 'delete-bookmark-lines') { bookmark.deleteBookmarkLines(editorView.value as any, props.tab); return }
  if (cmd === 'delete-unbookmark-lines') { bookmark.deleteUnbookmarkLines(editorView.value as any, props.tab); return }
  if (cmd === 'paste-bookmark-lines') { bookmark.pasteBookmarkLines(editorView.value as any, props.tab); return }

  // 高亮 / 标记
  if (cmd === 'clear-mark' || cmd === 'clear-highlight' || cmd === 'clear-all-highlight') { clearWordHighlight(); return }
  if (cmd === 'word-highlight') { highlightWordAtCursor(); return }
  if (cmd === 'mark-color') { highlightWordAtCursor(); return }
  if (cmd === 'mark-all') {
    if (Array.isArray(args[0])) highlightRanges(args[0] as number[], Number(args[1] ?? 0))
    else markAll(String(args[0] ?? lastSearchTerm))
    return
  }
  if (cmd === 'mark-red') { currentMarkColor = 1; markSelectionOrWord(); return }
  if (cmd === 'mark-yellow') { currentMarkColor = 0; markSelectionOrWord(); return }
  if (cmd === 'mark-blue') { currentMarkColor = 2; markSelectionOrWord(); return }
  if (cmd === 'mark-1') { currentMarkColor = 0; markSelectionOrWord(); return }
  if (cmd === 'mark-2') { currentMarkColor = 1; markSelectionOrWord(); return }
  if (cmd === 'mark-3') { currentMarkColor = 2; markSelectionOrWord(); return }
  if (cmd === 'mark-4') { currentMarkColor = 3; markSelectionOrWord(); return }
  if (cmd === 'mark-5') { currentMarkColor = 4; markSelectionOrWord(); return }
  if (cmd === 'mark-loop') { currentMarkColor = (currentMarkColor + 1) % 5; markSelectionOrWord(); return }
  if (cmd === 'mark-keywords') { markKeywords(args[0] as string[]); return }

  // 列编辑（通过 useEditorColumnMode）
  if (cmd === 'column-mode' || cmd === 'column-block') { columnMode.enterColumnMode(); return }

  // MD 模式
  if (cmd === 'toggle-md-mode') { markdown.toggleMdMode(); return }

  // 光标 / 跳转
  if (cmd === 'goto-bracket') { gotoBracket(); return }
  if (cmd === 'prev-position') { gotoPrevPosition(); return }
  if (cmd === 'next-position') { gotoNextPosition(); return }
  if ((cmd === 'insert-text' || cmd === 'insert-snippet') && args[0]) {
    editorView.value?.dispatch({ changes: { from: editorView.value.state.selection.main.head, insert: String(args[0]) } })
    return
  }
  if (cmd === 'goto-line' && args[0]) { gotoLine(args[0] as number); return }
  if (cmd === 'scroll-to-pos' && args[0] && editorView.value) {
    editorView.value.dispatch({ selection: { anchor: args[0] }, scrollIntoView: true }); return
  }
  if (cmd === 'scroll-to-line' && args[0]) { gotoLine(args[0] as number); return }
  if (cmd === 'scroll-to-end' && editorView.value) {
    editorView.value.dispatch({ selection: { anchor: editorView.value.state.doc.length }, scrollIntoView: true }); return
  }
  if (cmd === 'insert-blank-above') { insertBlankLine(true); return }
  if (cmd === 'insert-blank-below') { insertBlankLine(false); return }

  // 行操作细粒度
  if (cmd === 'line-duplicate') { lineOperation('duplicate'); return }
  if (cmd === 'line-remove') { lineOperation('remove'); return }
  if (cmd === 'line-moveUp') { lineOperation('moveUp'); return }
  if (cmd === 'line-moveDown') { lineOperation('moveDown'); return }
  if (cmd === 'line-removeEmpty') { lineOperation('removeEmpty'); return }
  if (cmd === 'line-removeEmptyCbc') { lineOperation('removeBlank'); return }
  if (cmd === 'line-reverse') { lineOperation('reverse'); return }
  if (cmd === 'line-split') { lineOperation('split'); return }
  if (cmd === 'line-join') { lineOperation('join'); return }
  if (cmd === 'line-removeDuplicate') { lineOperation('removeDuplicate'); return }

  // 缩进（CodeMirror 自带 keymap 也可触发，这里仅暴露给命令系统）
  if (cmd === 'indent') { indentLines(); return }
  if (cmd === 'dedent') { dedentLines(); return }

  // 视图切换
  if (cmd === 'show-all') { toggleShowAll(); return }
  if (cmd === 'show-eol') { toggleEol(true); return }
  if (cmd === 'hide-eol') { toggleEol(false); return }
  if (cmd === 'toggle-webaddr') { toggleWebAddr(); return }

  // 列编辑插入
  if (cmd === 'column-insert-text' && args[0]) {
    const v = editorView.value
    if (!v) return
    const text = String(args[0])
    const ranges = [...v.state.selection.ranges].sort((a, b) => b.from - a.from)
    if (ranges.length <= 1) v.dispatch({ changes: { from: v.state.selection.main.head, insert: text } })
    else v.dispatch({ changes: ranges.map(r => ({ from: r.from, insert: text })) })
    return
  }
  if (cmd === 'column-insert-num' && args[0]) {
    const opts = args[0]
    let val = opts.init
    const lines: string[] =[]
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

// ==================== Minimap & 显示 ====================
const showMinimap = ref(true)
const minimapViewport = ref<{ top: number; height: number }>({ top: 0, height: 0 })
function syncMinimap() {
  const v = editorView.value
  if (!v) return
  const el = v.scrollDOM
  const top = el.scrollTop
  const visibleH = el.clientHeight
  const totalH = el.scrollHeight
  minimapViewport.value = { top, height: visibleH }
}
let roScroll: ResizeObserver | null = null

// ==================== 监听 ====================
watch(() => props.tab.id, () => {
  saveEditorState()
  destroyEditor()
  createEditor()
  restoreEditorState()
})
watch(() => props.tab.path, () => {
  // tab 路径变化（重命名 / 新打开）→ 重新拉书签（bug #1 修复）
  bookmark.scheduleSyncFromDB(props.tab)
})
watch(() => props.tab.language, () => {
  updateLanguageForTab()
  if (props.tab.language === 'markdown') markdown.initMermaid(colors.value.isDark)
})
watch(() => colors.value.isDark, () => reconfigureAppearance())
watch(() => settingStore.config?.editor?.wordWrap, (on) => { if (on !== undefined) toggleWordWrap(on) })
watch(() => settingStore.config?.ui?.showWebAddr, () => toggleWebAddr())
watch(() => editorStore.macroState.isPlaying, (playing) => {
  if (playing) macro.playMacro(editorView.value as any)
})

// ==================== 生命周期 ====================
onMounted(() => {
  createEditor()
  restoreEditorState()
  document.addEventListener('editor-command', handleEditorCommand as EventListener)
  // 右键菜单命令（来自 ext-context-menu）
  document.addEventListener('cmd', ((e: CustomEvent) => {
    if (e.detail) handleEditorCommand(new CustomEvent('cmd', { detail: e.detail }))
  }) as EventListener)
  if (editorView.value) {
    roScroll = new ResizeObserver(() => syncMinimap())
    roScroll.observe(editorView.value.scrollDOM)
  }
})
onBeforeUnmount(() => {
  saveEditorState()
  destroyEditor()
  document.removeEventListener('editor-command', handleEditorCommand as EventListener)
  roScroll?.disconnect()
  if (posTimer) window.clearTimeout(posTimer)
})

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

    <div class="flex-1 flex overflow-hidden" style="position:relative;" @contextmenu="contextMenu.open($event as any, editorView)" @click="contextMenu.close()">
      <div ref="editorContainer" :class="{'w-full':!markdown.isMarkdown.value||markdown.mdMode.value==='edit','cm-container-split border-r border-gray-200 dark:border-gray-700':markdown.isMarkdown.value&&markdown.mdMode.value==='split','hidden':markdown.isMarkdown.value&&markdown.mdMode.value==='preview','with-minimap':showMinimap}" class="cm-container" @dblclick="highlightWordAtCursor"></div>

      <!-- 右键菜单（来自 ext-context-menu 数据） -->
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
/* ---- 右键菜单样式 ---- */
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
  font-size: 13px;
  color: #333;
  background: transparent;
  border: none;
  cursor: pointer;
  text-align: left;
  transition: background 0.1s;
}
.context-menu-item:hover:not(:disabled) { background: #e8e8e8; }
.context-menu-item:disabled { color: #ccc; cursor: not-allowed; }
.context-menu-item .item-label { flex: 1; }
.context-menu-item .item-shortcut {
  font-size: 11px;
  color: #999;
  margin-left: 20px;
}
.context-menu-separator {
  height: 1px;
  margin: 4px 0;
  background-color: #e5e5e5;
}
html.dark .context-menu-item { color: #e0e0e0; }
html.dark .context-menu-item:hover:not(:disabled) { background: #3a3a3a; }
html.dark .context-menu-item:disabled { color: #555; }
html.dark .context-menu-separator { background-color: #3a3a3a; }
</style>