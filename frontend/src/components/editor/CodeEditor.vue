<script lang="ts" setup>
import { ref, onMounted, onUnmounted, watch, computed, nextTick } from 'vue'
import { useEditorStore, useSettingStore } from '@/stores'
import { ElMessage } from 'element-plus'
import type { EditorTab, MdViewMode, Snippet } from '@/types'
import { AddBookmark, RemoveBookmark, GetBookmarks, GetSnippets } from '../../../wailsjs/go/main/App'
import { EditorView } from 'codemirror'
import { keymap, gutter, GutterMarker, Decoration, highlightActiveLine, lineNumbers, highlightSpecialChars, rectangularSelection, crosshairCursor, dropCursor, highlightActiveLineGutter } from '@codemirror/view'
import { EditorState, Compartment, Prec, StateEffect, StateField, RangeSetBuilder, RangeSet } from '@codemirror/state'
import { syntaxHighlighting, defaultHighlightStyle, foldGutter, foldKeymap, indentOnInput, bracketMatching, StreamLanguage, LanguageSupport } from '@codemirror/language'
import { defaultKeymap, history, historyKeymap, indentWithTab } from '@codemirror/commands'
import { closeBrackets, closeBracketsKeymap, completionKeymap, autocompletion, snippetCompletion } from '@codemirror/autocomplete'
import { lintKeymap } from '@codemirror/lint'
import { toggleComment, toggleBlockComment, undo, redo } from '@codemirror/commands'
import MarkdownIt from 'markdown-it'
import mermaid from 'mermaid'             // 🆕 V2.0.0
import katex from 'katex'                 // 🆕 V2.0.0
import 'katex/dist/katex.min.css'         // 🆕 V2.0.0
import hljs from 'highlight.js/lib/common'                    // 🆕 V2.0.0 预览代码块语法高亮
import 'highlight.js/styles/github.css'                        // 🆕 V2.0.0 高亮主题（亮色）
import 'github-markdown-css/github-markdown-light.css'         // 🆕 V2.0.0 GitHub 风格预览样式

// ---- Language support imports ----
import { json } from '@codemirror/lang-json'
import { javascript } from '@codemirror/lang-javascript'
import { html, autoCloseTags } from '@codemirror/lang-html'
import { css } from '@codemirror/lang-css'
import { markdown } from '@codemirror/lang-markdown'
import { xml } from '@codemirror/lang-xml'
import { yaml } from '@codemirror/lang-yaml'
import { python } from '@codemirror/lang-python'
import { java } from '@codemirror/lang-java'
import { go } from '@codemirror/lang-go'
import { sql } from '@codemirror/lang-sql'
import Minimap from './Minimap.vue'

// ==================== Props & Emits ====================
const props = defineProps<{ tab: EditorTab }>()
const emit = defineEmits<{
  (e: 'update:wordWrap', value: boolean): void
  (e: 'update:indentGuide', value: boolean): void
}>()

const editorStore = useEditorStore()
const settingStore = useSettingStore()
const editorContainer = ref<HTMLElement | null>(null)
let editorView: EditorView | null = null
let isInitializing = false

// ---- Compartments ----
const tabSizeCompartment = new Compartment()
const wordWrapCompartment = new Compartment()
const appearanceCompartment = new Compartment()
const foldCompartment = new Compartment()
const indentGuideCompartment = new Compartment()
const showWhitespaceCompartment = new Compartment()
const langCompartment = new Compartment()
const webAddrCompartment = new Compartment()

// ---- Markdown state ----
const mdMode = ref<MdViewMode>('split')
const isMarkdown = computed(() => props.tab.language === 'markdown')

// ---- 文档地图（minimap）状态 ----
const showMinimap = ref(false)
const minimapViewport = ref({ scrollTop: 0, scrollHeight: 0, clientHeight: 0 })

// ---- Goto line ----
const showGotoLine = ref(false)
const gotoLineInput = ref('')

// ---- Config shortcuts ----
const config = computed(() => settingStore.config)
const colors = computed(() => settingStore.currentThemeColors)

// ==================== StreamLanguage builder (80+ languages) ====================
// 用 StreamLanguage 为没有独立 @codemirror/lang-* 包的语言创建语法高亮
function buildStreamLanguage(
  name: string,
  keywords: string,
  types: string = '',
  constants: string = '',
  lineComment: string = '//',
  blockComment: string[] = ['/*', '*/'],
  strings: string[][] = [['"', '"']],
): LanguageSupport {
  const kwSet = new Set((keywords + ' ' + types + ' ' + constants).toLowerCase().split(/\s+/).filter(Boolean))
  return new LanguageSupport(StreamLanguage.define({
    name,
    token(stream, state: any) {
      // Skip whitespace
      if (stream.eatSpace()) return null
      // Line comment
      if (stream.match(lineComment)) { stream.skipToEnd(); return 'lineComment' }
      // Block comment
      if (blockComment.length >= 2 && stream.match(blockComment[0])) {
        while (!stream.eol()) {
          if (stream.match(blockComment[1])) return 'blockComment'
          stream.next()
        }
        return 'blockComment'
      }
      // Strings
      for (const [open, close] of strings) {
        if (stream.match(open)) {
          while (!stream.eol()) {
            if (stream.match('\\' + close)) { stream.next(); continue }
            if (stream.match(close)) return 'string'
            stream.next()
          }
          return 'string'
        }
      }
      // Numbers
      if (stream.match(/0[xX][0-9a-fA-F]+/) || stream.match(/0[oO][0-7]+/) ||
          stream.match(/0[bB][01]+/) || stream.match(/\d+\.?\d*/))
        return 'number'
      // Identifiers & keywords
      if (stream.match(/[a-zA-Z_$][\w$]*/)) {
        const word = stream.current().toLowerCase()
        if (kwSet.has(word)) return 'keyword'
        if (word === 'true' || word === 'false' || word === 'null' || word === 'nil' || word === 'none') return 'atom'
        return 'variableName'
      }
      // Operators
      if (stream.match(/[+\-*/%&|^~!=<>?:;,.\[\]{}()#@`]+/)) return 'operator'
      stream.next()
      return null
    },
    languageData: {
      commentTokens: { line: lineComment, block: blockComment.length >= 2 ? { open: blockComment[0], close: blockComment[1] } : undefined },
      indentOnInput: /^\s*[\}\]]$/,
    },
  }), [])
}

// Stream language definitions matching notepad-- lexer set
const streamLangs: Record<string, () => LanguageSupport> = {
  rust: () => buildStreamLanguage('rust',
    'as async await break const continue crate dyn else enum extern false fn for if impl in let loop macro match mod move mut pub ref return self static struct super trait true type unsafe use where while',
    'bool char f32 f64 i8 i16 i32 i64 i128 isize str String u8 u16 u32 u64 u128 usize Vec Option Result',
    'Some None Ok Err', '//', ['/*', '*/'], [['"', '"'], ["r#\"", "\"#"]]
  ),
  c: () => buildStreamLanguage('c',
    'auto break case char const continue default do double else enum extern float for goto if int long register return short signed sizeof static struct switch typedef union unsigned void volatile while',
    'int8_t int16_t int32_t int64_t uint8_t uint16_t uint32_t uint64_t size_t ssize_t bool FILE NULL',
    'true false NULL', '//', ['/*', '*/']
  ),
  cpp: () => buildStreamLanguage('cpp',
    'alignas alignof and and_eq asm auto bitand bitor bool break case catch char class compl const constexpr const_cast continue decltype default delete do double dynamic_cast else enum explicit export extern false float for friend goto if inline int long mutable namespace new noexcept not not_eq nullptr operator or or_eq override private protected public register reinterpret_cast return short signed sizeof static static_assert static_cast struct switch template this thread_local throw true try typedef typeid typename union unsigned using virtual void volatile wchar_t while xor xor_eq',
    'string vector map set pair shared_ptr unique_ptr istream ostream fstream', 'nullptr true false', '//', ['/*', '*/']
  ),
  csharp: () => buildStreamLanguage('csharp',
    'abstract as base bool break byte case catch char checked class const continue decimal default delegate do double else enum event explicit extern false finally fixed float for foreach goto if implicit in int interface internal is lock long namespace new null object operator out override params private protected public readonly ref return sbyte sealed short sizeof stackalloc static string struct switch this throw true try typeof uint ulong unchecked unsafe ushort using var virtual void volatile while',
    'string int long float double decimal bool char void object var dynamic', 'null true false',
    '//', ['/*', '*/'], [['"', '"'], ["@\"", "\""], ["'", "'"]]
  ),
  php: () => buildStreamLanguage('php',
    'abstract and array as break callable case catch class clone const continue declare default die do echo else elseif empty enddeclare endfor endforeach endif endswitch endwhile eval exit extends final finally fn for foreach function global goto if implements include include_once instanceof insteadof interface isset list match namespace new or print private protected public readonly require require_once return static switch throw trait try unset use var while xor yield',
    'int float string bool array object null mixed void callable iterable self static parent', 'true false null',
    '//', ['/*', '*/'], [['"', '"'], ["'", "'"]]
  ),
  ruby: () => buildStreamLanguage('ruby',
    'BEGIN END alias and begin break case class def defined? do else elsif end ensure false for if in module next nil not or redo rescue retry return self super then true undef unless until when while yield',
    'String Integer Float Array Hash Symbol', 'true false nil', '#', ['=begin', '=end'], [['"', '"'], ["'", "'"]]
  ),
  lua: () => buildStreamLanguage('lua',
    'and break do else elseif end false for function goto if in local nil not or repeat return then true until while',
    'string number boolean table function thread userdata', 'true false nil', '--', ['--[[', ']]'], [['"', '"'], ["'", "'"]]
  ),
  kotlin: () => buildStreamLanguage('kotlin',
    'abstract actual annotation as as? break case catch class companion const continue crossinline data delegate do dynamic else enum expect external false field file final finally for fun get if import in infix init inline inner interface internal is lateinit noinline null object open operator out override package param private property protected public receiverisinline reified return sealed set super suspend tailrec this throw true try typealias typeof val var vararg when where while',
    'Int Long Float Double Boolean String Char Byte Short Any Unit Nothing', 'true false null', '//', ['/*', '*/']
  ),
  scala: () => buildStreamLanguage('scala',
    'abstract case catch class def do else extends false final finally for forSome if implicit import lazy match new null object override package private protected return sealed super this throw trait try true type val var while with yield',
    'Int Long Float Double Boolean String Char Byte Short Any Unit Nothing Option List Map Set', 'true false null', '//', ['/*', '*/']
  ),
  clojure: () => buildStreamLanguage('clojure',
    'def defn fn let if when loop recur doseq for map reduce filter', '', 'true false nil', ';', [], [['"', '"']]
  ),
  dart: () => buildStreamLanguage('dart',
    'abstract as assert async await break case catch class const continue covariant default deferred do dynamic else enum export extends extension external factory false final finally for Function get hide if implements import in interface is late library mixin new null on operator part required rethrow return set show static super switch sync this throw true try typedef var void when while with yield',
    'int double String bool List Map Set dynamic void num', 'true false null', '//', ['/*', '*/'], [['"', '"'], ["'", "'"]]
  ),
  swift: () => buildStreamLanguage('swift',
    'associatedtype async await break case catch class continue convenience default defer deinit didSet do dynamic else enum extension fallthrough false fileprivate final for func get guard if import in indirect infix init inout internal is lazy let mutating nil nonmutating open operator optional override postfix precedence prefix private protocol public repeat required rethrows return self set some static struct subscript super switch throw throws true try typealias unowned var weak where while willSet',
    'Int UInt Float Double Bool String Character Array Dictionary Set Optional Void Any', 'true false nil', '//', ['/*', '*/']
  ),
  elixir: () => buildStreamLanguage('elixir',
    'after and catch def defmodule defp do else end false fn for if import in nil not or raise rescue try unless use when',
    'integer float boolean atom list tuple map pid port reference', 'true false nil', '#', [], [['"', '"'], ["'", "'"]]
  ),
  erlang: () => buildStreamLanguage('erlang',
    'after and andalso band begin bnot bor bsl bsr bxor case catch cond div end fun if let not of or orelse receive rem try when xor',
    'integer float atom pid port reference', 'true false', '%', [], [['"', '"']]
  ),
  haskell: () => buildStreamLanguage('haskell',
    'as case of class data default deriving do else forall foreign hiding if import in infix infixl infixr instance let mdo module newtype of qualified then type where',
    'Int Integer Float Double Bool Char String IO Maybe Either', 'True False', '--', ['{-', '-}']
  ),
  julia: () => buildStreamLanguage('julia',
    'abstract begin break catch const continue do else elseif end export finally for function global if import let local macro module mutable outer primitive quote return struct try using while',
    'Int Int8 Int16 Int32 Int64 Int128 Float16 Float32 Float64 Bool Char String Array Dict Set Tuple Symbol', 'true false nothing', '#', ['#=', '=#']
  ),
  shell: () => buildStreamLanguage('shell',
    'if then else elif fi case esac for while until do done in function return exit break continue select local export readonly declare typeset unset alias',
    '', 'true false', '#', [], [['"', '"'], ["'", "'"]]
  ),
  powershell: () => buildStreamLanguage('powershell',
    'Begin Break Catch Continue Data Do DynamicParam Else ElseIf End Exit Filter Finally For ForEach From Function If In InlineScript Param Process Return Switch Throw Trap Try Until Using Var While Workflow',
    'string int long bool array hashtable psobject scriptblock', '$true $false $null', '#', ['<#', '#>'], [['"', '"'], ["'", "'"]]
  ),
  perl: () => buildStreamLanguage('perl',
    'if else elsif unless while for foreach continue do require use my our local sub return last next redo goto die warn eval', '', '', '#', [], [['"', '"'], ["'", "'"]]
  ),
  r: () => buildStreamLanguage('r',
    'if else for while repeat break next function return in NULL NA NaN Inf TRUE FALSE', '', 'TRUE FALSE NULL NA', '#', [], [['"', '"'], ["'", "'"]]
  ),
  latex: () => buildStreamLanguage('latex', '', '', '', '%', [], []),
  batch: () => buildStreamLanguage('batch',
    'call echo set if else for goto pause exit rem start cd md rd del copy move ren type cls title color path ver', '', '', 'REM', [], []
  ),
  toml: () => buildStreamLanguage('toml', '', '', 'true false', '#', [], [['"', '"'], ["'", "'"]]),
  ini: () => buildStreamLanguage('ini', '', '', '', ';', [], [['"', '"']]),
  cmake: () => buildStreamLanguage('cmake',
    'if else elseif endif foreach endforeach while endwhile function endfunction macro endmacro break continue return set unset list string file message option', '', 'TRUE FALSE ON OFF YES NO', '#', [])
  ,
  protobuf: () => buildStreamLanguage('protobuf',
    'syntax package import option message enum service rpc returns repeated optional required oneof map extensions reserved', '', 'true false', '//', ['/*', '*/']
  ),
  graphql: () => buildStreamLanguage('graphql',
    'query mutation subscription fragment on implements interface union enum input type schema scalar extend directive', '', 'true false null', '#', [], [['"', '"']]
  ),
  dockerfile: () => buildStreamLanguage('dockerfile',
    'FROM RUN CMD LABEL EXPOSE ENV ADD COPY ENTRYPOINT VOLUME USER WORKDIR ARG ONBUILD STOPSIGNAL HEALTHCHECK SHELL MAINTAINER', '', '', '#', [], [['"', '"']]
  ),
  zig: () => buildStreamLanguage('zig',
    'align and anytype asm async await break catch comptime const continue defer else enum errdefer error export extern fn for if inline linksection noalias nosuspend or orelse packed pub resume return struct suspend switch test threadlocal try union unreachable usingnamespace var volatile while',
    'bool f16 f32 f64 f80 f128 i8 i16 i32 i64 i128 isize u8 u16 u32 u64 u128 usize void noreturn type anyerror comptime_int comptime_float', 'true false null undefined', '//', [], [['"', '"']]
  ),
  solidity: () => buildStreamLanguage('solidity',
    'abstract after catch contract enum event function interface is library mapping modifier override pragma private public pure returns storage struct view', 'address bool bytes int uint string mapping', 'true false', '//', ['/*', '*/']
  ),
  pascal: () => buildStreamLanguage('pascal',
    'and array begin case const div do downto else end file for function goto if in label mod nil not of or packed procedure program record repeat set then to type until var while with',
    'integer real boolean char string', 'true false nil', '//', ['{', '}'], [["'", "'"]]
  ),
  fortran: () => buildStreamLanguage('fortran',
    'allocatable allocate call case contains continue cycle deallocate do else end function if implicit in integer module none nullify only parameter pointer private program public real recursive result return save select stop subroutine then type use where while',
    '.true. .false.', '.true. .false.', '!', [], [["'", "'"]]
  ),
  cobol: () => buildStreamLanguage('cobol',
    'ACCEPT ADD CALL CLOSE COMPUTE DELETE DISPLAY DIVIDE EVALUATE GO GOBACK IF INITIALIZE MERGE MOVE MULTIPLY OPEN PERFORM READ RETURN REWRITE SEARCH SET SORT START STOP STRING SUBTRACT UNSTRING WRITE',
    'PIC X 9 A V S COMP COMP-3', '', '*', [], [['"', '"'], ["'", "'"]]
  ),
  tcl: () => buildStreamLanguage('tcl', 'if else elseif for foreach while switch proc return set upvar global variable', '', '', '#', [], [['"', '"']]),
  scheme: () => buildStreamLanguage('scheme',
    'define lambda let let* letrec if cond else begin set! quote quasiquote unquote unquote-splicing do delay force', '', '#t #f', ';', [], [['"', '"']]
  ),
  smalltalk: () => buildStreamLanguage('smalltalk',
    'self super nil true false', '', 'true false nil', '"', [], [["'", "'"]]
  ),
  prolog: () => buildStreamLanguage('prolog', ':- ! true fail not is', '', 'true fail', '%', ['/*', '*/']),
  ada: () => buildStreamLanguage('ada',
    'abort abs accept access all and array at begin body case constant declare delay delta digits do else elsif end entry exception exit for function generic goto if in is limited loop mod new not null of or others out package pragma private procedure raise range record rem renames return reverse select separate subtype task terminate then type use when while with xor',
    'Integer Float Boolean Character String Duration', 'True False', '--', [], [['"', '"']]
  ),
  nsis: () => buildStreamLanguage('nsis', 'Function FunctionEnd Section SectionEnd SetOutPath File WriteRegStr ReadRegStr DeleteRegKey IfErrors MessageBox DetailPrint StrCpy IntCmp IntCmpU StrCmp Exch Pop Push Call nsDialogs Create', '', '', ';', ['/*', '*/']),
  assembly: () => buildStreamLanguage('assembly', 'mov push pop call ret jmp je jne jg jl cmp add sub mul div inc dec xor and or not shl shr lea int nop', 'eax ebx ecx edx esi edi esp ebp rax rbx rcx rdx', '', ';', [], [['"', '"'], ["'", "'"]]),
  diff: () => new LanguageSupport(StreamLanguage.define({
    name: 'diff', token(stream) {
      if (stream.sol() && stream.match(/^---/)) { stream.skipToEnd(); return 'lineComment' }
      if (stream.sol() && stream.match(/^\+\+\+/)) { stream.skipToEnd(); return 'lineComment' }
      if (stream.sol() && stream.match(/^@@/)) { stream.skipToEnd(); return 'keyword' }
      if (stream.sol() && stream.match(/^\+/)) { stream.skipToEnd(); return 'string' }
      if (stream.sol() && stream.match(/^-/)) { stream.skipToEnd(); return 'keyword' }
      stream.skipToEnd(); return null
    },
  }), []),
}

// ==================== Language mapper ====================
function getLanguageExtension(lang: string): any[] {
  const builtInLangs: Record<string, () => any[]> = {
    'json': () => [json()],
    'javascript': () => [javascript()],
    'typescript': () => [javascript({ typescript: true })],
    'html': () => [html(), autoCloseTags],
    'css': () => [css()],
    'markdown': () => [markdown()],
    'xml': () => [xml(), autoCloseTags],
    'yaml': () => [yaml()],
    'python': () => [python()],
    'java': () => [java()],
    'go': () => [go()],
    'sql': () => [sql()],
    // Use javascript as fallback for C-like languages that work reasonably well
    'vue': () => [html()], // HTML handles Vue templates
    'svelte': () => [html()],
    'scss': () => [css()],
    'sass': () => [css()],
    'less': () => [css()],
    'csv': () => [], // plain text
    'text': () => [],
  }
  // Check built-in first
  if (lang in builtInLangs) return builtInLangs[lang]()
  // Check stream languages
  if (lang in streamLangs) return [streamLangs[lang]()]
  // Fallback to javascript for unknown C-like
  if (['objectivec', 'objectivecpp', 'd', 'nim', 'v', 'coffeescript'].includes(lang))
    return [javascript()]
  return []
}

// ==================== Theme builder ====================
function buildAppearanceTheme() {
  const c = colors.value
  return EditorView.theme({
    '&': {
      fontSize: `${config.value?.editor?.fontSize || 14}px`,
      fontFamily: config.value?.editor?.fontFamily || 'Consolas, Monaco, "Courier New", monospace',
      backgroundColor: c.bg, color: c.fg,
    },
    '.cm-scroller': { fontFamily: config.value?.editor?.fontFamily || 'Consolas, Monaco, "Courier New", monospace' },
    '.cm-content': { userSelect: 'text', WebkitUserSelect: 'text', MozUserSelect: 'text', WebkitUserDrag: 'none' },
    '.cm-line': { userSelect: 'text', WebkitUserSelect: 'text', MozUserSelect: 'text' },
    '.cm-gutters': { backgroundColor: c.gutterBg, color: c.gutterFg, borderRight: `1px solid ${c.isDark ? '#404040' : '#e5e7eb'}` },
    '.cm-activeLineGutter': { backgroundColor: c.activeLine },
    '.cm-activeLine': { backgroundColor: c.activeLine },
    '.cm-cursor, &.cm-focused .cm-cursor': { borderLeftColor: c.cursor },
    '.cm-matchingBracket': { backgroundColor: c.bracketMatch, outline: '1px solid ' + c.accent },
    '&.cm-focused .cm-matchingBracket': { backgroundColor: c.bracketMatch },
    // 🆕 V2.0.0 标签配对高亮
    '.cm-matchingTag': { backgroundColor: c.bracketMatch, outline: '1px solid ' + c.accent },
    '.cm-foldPlaceholder': { backgroundColor: c.isDark ? '#3c3c3c' : '#e5e7eb', color: c.fg, border: 'none' },
    // Native ::selection (no drawSelection — avoids WebView2 partial-line selection bug)
    '& .cm-content ::selection': { backgroundColor: c.selection, color: 'inherit' },
  })
}

// ---- Fold gutter ----
function buildFoldGutter(isDark: boolean) {
  return foldGutter({
    markerDOM(open) {
      const span = document.createElement('span')
      span.style.cssText = `display:inline-flex;align-items:center;justify-content:center;width:14px;height:14px;cursor:pointer;font-size:11px;color:${isDark ? '#858585' : '#999'};`
      span.textContent = open ? '▾' : '▸'
      return span
    },
  })
}

// ---- Highlight word (double-click) ----
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

// ---- Multi-color marks (notepad-- mark_style_1..5) ----
// 5 种标记色 + 循环，用于"标记颜色"菜单与 mark-all
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

// ---- URL highlight (视图→显示网页地址) ----
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

// ==================== Punctuation keymap (fix Chinese IME) ====================
function makeInsertChar(ch: string) {
  return (view: EditorView) => {
    if (view.composing || view.compositionStarted) return false
    view.dispatch(view.state.replaceSelection(ch))
    return true
  }
}
const punctuationKeymap = Prec.high(keymap.of([
  { key: ',', run: makeInsertChar(',') }, { key: ';', run: makeInsertChar(';') },
  { key: '.', run: makeInsertChar('.') }, { key: ':', run: makeInsertChar(':') },
  { key: '!', run: makeInsertChar('!') }, { key: '?', run: makeInsertChar('?') },
  { key: '-', run: makeInsertChar('-') }, { key: '_', run: makeInsertChar('_') },
  { key: '~', run: makeInsertChar('~') }, { key: '@', run: makeInsertChar('@') },
  { key: '#', run: makeInsertChar('#') }, { key: '$', run: makeInsertChar('$') },
  { key: '%', run: makeInsertChar('%') }, { key: '^', run: makeInsertChar('^') },
  { key: '&', run: makeInsertChar('&') }, { key: '*', run: makeInsertChar('*') },
  { key: '+', run: makeInsertChar('+') }, { key: '=', run: makeInsertChar('=') },
  { key: '/', run: makeInsertChar('/') }, { key: '\\', run: makeInsertChar('\\') },
  { key: '|', run: makeInsertChar('|') },
]))

// ==================== Editor creation ====================
function createEditor() {
  if (!editorContainer.value) return
  isInitializing = true
  if (editorView) { editorView.destroy(); editorView = null }

  const langExtensions = getLanguageExtension(props.tab.language)
  const isDark = colors.value.isDark
  const ed = config.value?.editor

  const state = EditorState.create({
    doc: props.tab.content,
    extensions: [
      // ★ 手动组合 basicSetup，排除 searchKeymap（内置英文搜索面板）
      lineNumbers(),
      highlightActiveLineGutter(),
      highlightSpecialChars(),
      history(),
      foldGutter({}),
      dropCursor(),
      EditorState.allowMultipleSelections.of(true),
      indentOnInput(),
      syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
      bracketMatching(),
      closeBrackets(),
      // 🆕 V2.0.0 代码片段自动补全
      autocompletion({
        override: [snippetCompletionSource as any],
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
      punctuationKeymap,
      // Language
      langCompartment.of(langExtensions),
      // Folding
      foldCompartment.of(buildFoldGutter(isDark)),
      // Word highlight
      highlightWordField,
      // Multi-color marks
      markField,
      // URL highlight (toggle via compartment)
      webAddrCompartment.of([]),
      // URL click handler
      EditorView.domEventHandlers({
        click(e: MouseEvent, view) {
          if (!settingStore.config?.ui?.showWebAddr) return false
          const pos = view.posAtCoords({ x: e.clientX, y: e.clientY })
          if (pos == null) return false
          const decos = view.state.field(webAddrField, false)
          if (!decos) return false
          const cur = decos.iter()
          let url: string | null = null
          while (cur.value) {
            if (pos >= cur.from && pos <= cur.to) { url = view.state.sliceDoc(cur.from, cur.to); break }
            cur.next()
          }
          if (url) {
            e.preventDefault()
            ;(async () => {
              try { const r = await import('../../../wailsjs/runtime/runtime'); r.BrowserOpenURL(url) }
              catch { window.open(url, '_blank') }
            })()
            return true
          }
          return false
        },
      }),
      // Configurable
      tabSizeCompartment.of(EditorState.tabSize.of(ed?.tabSize || 4)),
      wordWrapCompartment.of(ed?.wordWrap ? EditorView.lineWrapping : []),
      appearanceCompartment.of(buildAppearanceTheme()),
      showWhitespaceCompartment.of(EditorView.theme({})),
      indentGuideCompartment.of(EditorView.baseTheme({})),
      // Comment keymaps
      Prec.high(keymap.of([
        { key: 'Mod-/', run: (view) => { toggleComment(view); return true }, preventDefault: true },
        { key: 'Mod-?', run: (view) => { toggleBlockComment(view); return true }, preventDefault: true },
      ])),
      // ★ notepad-- 快捷键：拦截后将事件转发给 MainLayout
      Prec.highest(keymap.of([
        { key: 'Mod-f', run: () => { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'find' })); return true } },
        { key: 'Mod-h', run: () => { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'replace' })); return true } },
        { key: 'Mod-g', run: () => { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'goto-line' })); return true } },
        { key: 'F3', run: () => { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'find-next' })); return true } },
        { key: 'Shift-F3', run: () => { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'find-prev' })); return true } },
        { key: 'F2', run: () => { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'next-bookmark' })); return true } },
        { key: 'Shift-F2', run: () => { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'prev-bookmark' })); return true } },
        { key: 'Mod-F2', run: () => { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'toggle-bookmark' })); return true } },
        { key: 'F7', run: () => { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'clear-highlight' })); return true } },
        { key: 'F8', run: () => { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'word-highlight' })); return true } },
        { key: 'F11', run: () => { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'fullscreen' })); return true } },
        { key: 'Mod-p', run: () => { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'print' })); return true } },
        { key: 'Mod-q', run: () => { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'exit' })); return true } },
        { key: 'Mod-Shift-d', run: () => { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'find-dir' })); return true } },
        { key: 'Mod-d', run: () => { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'line-duplicate' })); return true } },
        { key: 'Mod-l', run: () => { document.dispatchEvent(new CustomEvent('ndd-key', { detail: 'line-delete' })); return true } },
      ])),
      // Update listener
      EditorView.updateListener.of((update) => {
        if (update.docChanged && !isInitializing) {
          editorStore.updateTabContent(props.tab.id, update.state.doc.toString())
          // Macro recording
          if (editorStore.macroState.isRecording) {
            for (const tr of update.transactions) {
              if (tr.docChanged) {
                let inserted = '', deleted = ''
                tr.changes.iterChanges((fromA, toA, fromB, toB, inserted_chunk) => {
                  if (fromA !== toA) deleted += update.startState.sliceDoc(fromA, toA)
                  if (fromB !== toB) inserted += inserted_chunk.toString()
                })
                if (inserted) {
                  const pos = update.state.selection.main.head
                  editorStore.recordMacroStep({ type: 'insert', text: inserted, timestamp: Date.now(), from: pos, to: pos + inserted.length })
                }
                if (deleted) {
                  editorStore.recordMacroStep({ type: 'delete', text: deleted, timestamp: Date.now() })
                }
              }
            }
          }
        }
        if (update.selectionSet) {
          const pos = update.state.selection.main.head
          const line = update.state.doc.lineAt(pos)
          editorStore.updateCursorPosition(props.tab.id, line.number, pos - line.from + 1)
          // 位置历史（节流：跨行移动 >5 行 或空闲 350ms 才记录）
          recordPosition(pos)
        }
      }),
    ],
  })

  editorView = new EditorView({ state, parent: editorContainer.value })
  Promise.resolve().then(() => { isInitializing = false })

  // 文档地图：滚动时同步视口（rAF 节流）
  const scroller = editorView.scrollDOM
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

  if (props.tab.cursorPosition.line > 1) {
    try {
      const line = state.doc.line(props.tab.cursorPosition.line)
      editorView.dispatch({ selection: { anchor: line.from + props.tab.cursorPosition.column - 1 }, scrollIntoView: true })
    } catch (e) { console.warn(e) }
  }
}

function reconfigureAppearance() {
  if (!editorView) return
  editorView.dispatch({ effects: appearanceCompartment.reconfigure(buildAppearanceTheme()) })
}

// ==================== Editor operations ====================
function gotoLine(lineNum: number) {
  if (!editorView) return
  const doc = editorView.state.doc
  if (lineNum < 1 || lineNum > doc.lines) return
  const line = doc.line(lineNum)
  editorView.dispatch({ selection: { anchor: line.from }, scrollIntoView: true })
  editorView.focus()
}

function showGotoLineDialog() {
  showGotoLine.value = true; gotoLineInput.value = ''
  setTimeout(() => { const el = document.querySelector('.goto-line-input') as HTMLInputElement; el?.focus() }, 50)
}

function handleGotoLine() {
  const num = parseInt(gotoLineInput.value)
  if (isNaN(num)) return
  gotoLine(num)
  showGotoLine.value = false
}

function toggleWordWrap(enable?: boolean) {
  if (!editorView) return
  const cur = editorView.lineWrapping
  const w = enable !== undefined ? enable : !cur
  editorView.dispatch({ effects: wordWrapCompartment.reconfigure(w ? EditorView.lineWrapping : []) })
  emit('update:wordWrap', w)
}

function toggleShowWhitespace(show: boolean) {
  if (!editorView) return
  const isDark = colors.value.isDark
  const gutterFg = isDark ? '%23666' : '%23bbb'
  const specialChars = show ? EditorView.theme({
    '& .cm-space': { backgroundImage: `url("data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' width='6' height='20'><circle cx='3' cy='12' r='1' fill='${gutterFg}'/></svg>")`, backgroundRepeat: 'no-repeat', },
    '& .cm-tab': { backgroundImage: `url("data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' width='20' height='20'><path d='M2 12 L16 12' stroke='${gutterFg}' stroke-width='0.5' fill='none'/><path d='M14 9 L18 12 L14 15' stroke='${gutterFg}' stroke-width='0.5' fill='none'/></svg>")`, backgroundRepeat: 'no-repeat', },
  }) : EditorView.theme({})
  editorView.dispatch({ effects: showWhitespaceCompartment.reconfigure(specialChars) })
}

// ---- Undo/Redo ----
function undoAction() { if (editorView) undo(editorView) }
function redoAction() { if (editorView) redo(editorView) }

// ---- Case transform ----
function transformCase(type: string) {
  if (!editorView) return
  const state = editorView.state
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
  }
  if (result !== text) {
    const from = isSelection ? selection.from : state.doc.lineAt(selection.from).from
    const to = isSelection ? selection.to : state.doc.lineAt(selection.from).to
    editorView.dispatch({ changes: { from, to, insert: result } })
  }
}

// ---- Line operations ----
function lineOperation(op: string) {
  if (!editorView) return
  const state = editorView.state
  const from = state.selection.main.from
  const to = state.selection.main.to
  const fromLine = state.doc.lineAt(from)
  const toLine = state.doc.lineAt(to)
  switch (op) {
    case 'duplicate': {
      const text = state.sliceDoc(fromLine.from, toLine.to)
      editorView.dispatch({ changes: { from: toLine.to, insert: '\n' + text } })
      break
    }
    case 'remove': {
      const end = toLine.to + 1 > state.doc.length ? state.doc.length : toLine.to + 1
      editorView.dispatch({ changes: { from: fromLine.from, to: end } })
      break
    }
    case 'moveUp': {
      if (fromLine.number <= 1) return
      const prevLine = state.doc.line(fromLine.number - 1)
      const text = state.sliceDoc(fromLine.from, toLine.to)
      const hasNL = toLine.to < state.doc.length && state.sliceDoc(toLine.to, toLine.to + 1) === '\n'
      editorView.dispatch({ changes: [{ from: fromLine.from - (prevLine.length + 1), to: toLine.to + (hasNL ? 1 : 0), insert: text + (hasNL ? '\n' : '') + prevLine.text }] })
      break
    }
    case 'moveDown': {
      if (toLine.number >= state.doc.lines) return
      const nextLine = state.doc.line(toLine.number + 1)
      const text = state.sliceDoc(fromLine.from, toLine.to)
      editorView.dispatch({ changes: [{ from: fromLine.from, to: nextLine.to, insert: nextLine.text + '\n' + text }] })
      break
    }
    case 'removeEmpty': {
      const lines = state.doc.toString().split('\n').filter(l => l.trim() !== '')
      editorView.dispatch({ changes: { from: 0, to: state.doc.length, insert: lines.join('\n') } })
      break
    }
    case 'removeBlank': {
      const lines = state.doc.toString().split('\n').filter(l => l.length > 0)
      editorView.dispatch({ changes: { from: 0, to: state.doc.length, insert: lines.join('\n') } })
      break
    }
    case 'split': {
      if (to > from) {
        const text = state.sliceDoc(from, to)
        editorView.dispatch({ changes: { from, to, insert: [...text].join('\n') } })
      }
      break
    }
    case 'join': {
      if (to > from) {
        editorView.dispatch({ changes: { from, to, insert: state.sliceDoc(from, to).replace(/\n/g, ' ') } })
      }
      break
    }
    case 'removeDuplicate': {
      const lines = state.doc.toString().split('\n')
      const seen = new Set<string>()
      editorView.dispatch({ changes: { from: 0, to: state.doc.length, insert: lines.filter(l => { if (seen.has(l)) return false; seen.add(l); return true }).join('\n') } })
      break
    }
    case 'removeConsecutiveDuplicate': {
      const lines = state.doc.toString().split('\n')
      const result: string[] = []; let last = ''
      for (const l of lines) { if (l !== last) { result.push(l); last = l } }
      editorView.dispatch({ changes: { from: 0, to: state.doc.length, insert: result.join('\n') } })
      break
    }
    case 'reverse': {
      const lines = state.doc.toString().split('\n')
      editorView.dispatch({ changes: { from: 0, to: state.doc.length, insert: lines.reverse().join('\n') } })
      break
    }
    case 'randomize': {
      const lines = state.doc.toString().split('\n')
      for (let i = lines.length - 1; i > 0; i--) { const j = Math.floor(Math.random() * (i + 1)); [lines[i], lines[j]] = [lines[j], lines[i]] }
      editorView.dispatch({ changes: { from: 0, to: state.doc.length, insert: lines.join('\n') } })
      break
    }
    case 'insertAbove': insertBlankLine(true); break
    case 'insertBelow': insertBlankLine(false); break
  }
}

// ---- Tab/space conversion ----
function convertTabsSpaces(type: string) {
  if (!editorView) return
  const state = editorView.state
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
  if (result !== text) editorView.dispatch({ changes: { from: 0, to: state.doc.length, insert: result } })
}

// ---- Trim whitespace ----
function trimWhitespace(mode: string) {
  if (!editorView) return
  const state = editorView.state
  const text = state.doc.toString()
  let result = text
  switch (mode) {
    case 'head': result = text.split('\n').map(l => l.replace(/^[ \t]+/, '')).join('\n'); break
    case 'tail': result = text.split('\n').map(l => l.replace(/[ \t]+$/, '')).join('\n'); break
    case 'both': result = text.split('\n').map(l => l.replace(/^[ \t]+/, '').replace(/[ \t]+$/, '')).join('\n'); break
  }
  if (result !== text) editorView.dispatch({ changes: { from: 0, to: state.doc.length, insert: result } })
}

// ---- Sort lines ----
function sortLines(direction: string) {
  if (!editorView) return
  const lines = editorView.state.doc.toString().split('\n')
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
  editorView.dispatch({ changes: { from: 0, to: editorView.state.doc.length, insert: sorted.join('\n') } })
}

// ---- Indent/Dedent ----
function indentLines() {
  if (!editorView) return
  const state = editorView.state
  const ts = config.value?.editor?.tabSize || 4
  const useSpaces = config.value?.editor?.insertSpaces ?? true
  const indent = useSpaces ? ' '.repeat(ts) : '\t'
  const fl = state.doc.lineAt(state.selection.main.from)
  const tl = state.doc.lineAt(state.selection.main.to)
  const changes = []
  for (let i = fl.number; i <= tl.number; i++) changes.push({ from: state.doc.line(i).from, insert: indent })
  editorView.dispatch({ changes })
}

function dedentLines() {
  if (!editorView) return
  const state = editorView.state
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
  if (changes.length > 0) editorView.dispatch({ changes })
}

// ---- Insert blank line ----
function insertBlankLine(above: boolean) {
  if (!editorView) return
  const state = editorView.state
  const pos = state.selection.main.head
  const line = state.doc.lineAt(pos)
  if (above) editorView.dispatch({ changes: { from: line.from, insert: '\n' }, selection: { anchor: line.from } })
  else editorView.dispatch({ changes: { from: line.to, insert: '\n' }, selection: { anchor: line.to + 1 } })
}

// 🆕 V2.0.0 Snippets 补全源
const snippetsCache = ref<Snippet[]>([])
let snippetsLoaded = false

async function loadSnippetsForCompletion() {
  if (snippetsLoaded) return
  try {
    const lang = props.tab.language || ''
    snippetsCache.value = await GetSnippets(lang)
    snippetsLoaded = true
  } catch { snippetsCache.value = [] }
}

function snippetCompletionSource(context: any) {
  const word = context.matchBefore(/\w+/)
  if (!word || snippetsCache.value.length === 0) return null

  // 每次打开补全时尝试加载 snippet
  loadSnippetsForCompletion()

  const options: any[] = []
  for (const snippet of snippetsCache.value) {
    if (snippet.prefix && word.text && snippet.prefix.startsWith(word.text)) {
      try {
        options.push(snippetCompletion(snippet.body, {
          label: snippet.prefix,
          detail: snippet.name,
          info: snippet.description || snippet.prefix,
        }))
      } catch { /* ignore invalid snippet syntax */ }
    }
  }
  return options.length > 0 ? { from: word.from, options } : null
}

// ---- Bookmarks ----
let lastSearchTerm = ''
function setSearchTerm(term: string) { lastSearchTerm = term }

// ---- Position history (节流) ----
let lastPushedPos = -1
let lastPushedLine = -1
let posTimer: any = null
function recordPosition(pos: number) {
  if (!editorView) return
  const line = editorView.state.doc.lineAt(pos).number
  clearTimeout(posTimer)
  posTimer = setTimeout(() => {
    if (Math.abs(line - lastPushedLine) >= 5 || lastPushedLine < 0) {
      editorStore.pushPosition(props.tab.id, pos, 0)
      lastPushedPos = pos; lastPushedLine = line
    }
  }, 350)
}

function toggleBookmark() {
  if (!editorView || !editorStore.activeTab) return
  const tab = editorStore.activeTab
  const line = tab.cursorPosition?.line || 1
  const isBooked = editorStore.hasBookmark(tab.id, line)
  if (isBooked) {
    // 🆕 V2.0.0 持久化删除书签：先查找后端记录 ID
    editorStore.toggleBookmark(tab.id, line)
    if (tab.path) {
      GetBookmarks(tab.path).then((entries: any) => {
        if (entries) {
          const found = (entries as any[]).find((b: any) => b.lineNumber === line)
          if (found) RemoveBookmark(found.id).catch(() => {})
        }
      }).catch(() => {})
    }
  } else {
    // 🆕 V2.0.0 持久化添加书签
    editorStore.toggleBookmark(tab.id, line)
    if (tab.path) {
      AddBookmark(tab.path, line, '', '').catch(() => {})
    }
  }
}

function gotoNextBookmark() {
  if (!editorView || !editorStore.activeTab) return
  const line = editorStore.activeTab.cursorPosition?.line || 1
  const next = editorStore.nextBookmark(editorStore.activeTab.id, line)
  if (next !== null) gotoLine(next)
}

function gotoPrevBookmark() {
  if (!editorView || !editorStore.activeTab) return
  const line = editorStore.activeTab.cursorPosition?.line || 1
  const prev = editorStore.prevBookmark(editorStore.activeTab.id, line)
  if (prev !== null) gotoLine(prev)
}

function clearAllBookmarks() {
  if (!editorStore.activeTab) return
  const tab = editorStore.activeTab
  // 🆕 V2.0.0 持久化清除所有书签
  const bookmarks = editorStore.getBookmarks(tab.id)
  if (tab.path) {
    // 获取该文件的所有书签并逐一删除
    GetBookmarks(tab.path).then((entries) => {
      if (entries) {
        (entries as any[]).forEach((b: any) => {
          RemoveBookmark(b.id).catch(() => {})
        })
      }
    }).catch(() => {})
  }
  editorStore.clearBookmarks(tab.id)
  ElMessage.success('已清除所有书签')
}

// ---- 书签行操作（剪切/复制/粘贴/删除书签行，notepad-- 书签系统）----
function copyBookmarkLines() {
  if (!editorView) return
  const lines = editorStore.getBookmarks(props.tab.id)
  if (!lines.length) { ElMessage.info('没有书签行'); return }
  const doc = editorView.state.doc
  const text = lines.map(n => doc.line(n).text).join('\n')
  editorStore.pushClipboard(text)
  navigator.clipboard?.writeText(text).catch(() => {})
  ElMessage.success(`已复制 ${lines.length} 个书签行`)
}

function cutBookmarkLines() {
  if (!editorView) return
  const lines = editorStore.getBookmarks(props.tab.id)
  if (!lines.length) { ElMessage.info('没有书签行'); return }
  const doc = editorView.state.doc
  const text = lines.map(n => doc.line(n).text).join('\n')
  editorStore.pushClipboard(text)
  navigator.clipboard?.writeText(text).catch(() => {})
  const changes = [...lines].sort((a, b) => b - a).map(n => {
    const l = doc.line(n)
    return { from: l.from, to: n < doc.lines ? doc.line(n + 1).from : doc.length }
  })
  editorView.dispatch({ changes })
  ElMessage.success(`已剪切 ${lines.length} 个书签行`)
}

function deleteBookmarkLines() {
  if (!editorView) return
  const lines = editorStore.getBookmarks(props.tab.id)
  if (!lines.length) { ElMessage.info('没有书签行'); return }
  const doc = editorView.state.doc
  const changes = [...lines].sort((a, b) => b - a).map(n => {
    const l = doc.line(n)
    return { from: l.from, to: n < doc.lines ? doc.line(n + 1).from : doc.length }
  })
  editorView.dispatch({ changes })
  ElMessage.success(`已删除 ${lines.length} 个书签行`)
}

function deleteUnbookmarkLines() {
  if (!editorView) return
  const lines = editorStore.getBookmarks(props.tab.id)
  if (!lines.length) { ElMessage.info('没有书签行'); return }
  const doc = editorView.state.doc
  const keep = new Set(lines)
  const result: string[] = []
  for (let n = 1; n <= doc.lines; n++) if (keep.has(n)) result.push(doc.line(n).text)
  editorView.dispatch({ changes: { from: 0, to: doc.length, insert: result.join('\n') } })
  ElMessage.success('已删除非书签行')
}

async function pasteBookmarkLines() {
  if (!editorView) return
  const lines = editorStore.getBookmarks(props.tab.id)
  if (!lines.length) { ElMessage.info('没有书签行'); return }
  const doc = editorView.state.doc
  let clip = editorStore.clipboardHistory[0] || ''
  if (!clip) {
    try { clip = await navigator.clipboard.readText() } catch { clip = '' }
  }
  if (!clip) { ElMessage.info('剪贴板为空'); return }
  const clipLines = clip.split('\n')
  const changes = lines
    .map((n, i) => ({ line: n, insert: clipLines[i % clipLines.length] }))
    .sort((a, b) => b.line - a.line)
    .map(({ line, insert }) => { const l = doc.line(line); return { from: l.from, to: l.to, insert } })
  editorView.dispatch({ changes })
  ElMessage.success('已粘贴到书签行')
}

// ---- Highlight ----
function highlightWordAtCursor() {
  if (!editorView) return
  const pos = editorView.state.selection.main.head
  const word = editorView.state.wordAt(pos)
  if (word) editorView.dispatch({ effects: addHighlightWord.of({ from: word.from, to: word.to }) })
}

function clearWordHighlight() { if (editorView) editorView.dispatch({ effects: clearHighlightWord.of(null) }) }

// ---- Column mode ----
function enterColumnMode() { document.dispatchEvent(new CustomEvent('show-column-edit')) }
function doColumnBlockEdit() { document.dispatchEvent(new CustomEvent('show-column-edit')) }
function toggleColumnMode() {
  if (!editorView) return
  const on = !(settingStore.config?.editor?.columnMode)
  if (settingStore.config) { settingStore.config.editor.columnMode = on; settingStore.saveConfig() }
  // rectangularSelection/crosshairCursor 已在基础扩展中，这里仅切换光标十字样式提示
  editorView.dom.style.cursor = on ? 'crosshair' : ''
  ElMessage?.success?.(on ? '列块模式已开启（Alt+拖拽选择矩形区域）' : '列块模式已关闭')
}

// ---- 多色标记实现 ----
function markAll(term: string) {
  if (!editorView || !term) { ElMessage?.warning?.('请先输入查找内容'); return }
  const doc = editorView.state.doc.toString()
  const ranges: MarkRange[] = []
  const re = new RegExp(escapeRegExp(term), 'gi')
  for (const m of doc.matchAll(re)) {
    if (m.index === undefined) continue
    ranges.push({ from: m.index, to: m.index + m[0].length, color: currentMarkColor })
  }
  if (!ranges.length) { ElMessage?.info?.('未找到匹配'); return }
  editorView.dispatch({ effects: setMarks.of(ranges) })
}
function markSelectionOrWord() {
  if (!editorView) return
  const { from, to } = editorView.state.selection.main
  let range: { from: number; to: number }
  if (from === to) {
    const word = editorView.state.wordAt(from)
    if (!word) return
    range = word
  } else range = { from, to }
  editorView.dispatch({ effects: addMarkRanges.of([{ ...range, color: currentMarkColor }]) })
}
function markKeywords(keywords: string[]) {
  if (!editorView || !keywords.length) return
  const doc = editorView.state.doc.toString()
  const ranges: MarkRange[] = []
  for (const kw of keywords) {
    if (!kw) continue
    const re = new RegExp(escapeRegExp(kw), 'gi')
    for (const m of doc.matchAll(re)) {
      if (m.index === undefined) continue
      ranges.push({ from: m.index, to: m.index + m[0].length, color: currentMarkColor })
    }
  }
  if (ranges.length) editorView.dispatch({ effects: setMarks.of(ranges) })
}

// ---- 括号跳转 ----
function gotoBracket() {
  if (!editorView) return
  const head = editorView.state.selection.main.head
  // 利用 bracketMatching 的匹配信息：简单实现，扫描前后括号配对
  const doc = editorView.state.doc.toString()
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
        editorView.dispatch({ selection: { anchor: i }, scrollIntoView: true })
        editorView.focus()
      }
      return
    }
  }
}

// ---- 位置历史跳转 ----
function gotoPrevPosition() {
  const r = editorStore.goBackPosition(props.tab.id)
  if (r && editorView) { editorView.dispatch({ selection: { anchor: r.pos }, scrollIntoView: true }) }
}
function gotoNextPosition() {
  const r = editorStore.goForwardPosition(props.tab.id)
  if (r && editorView) { editorView.dispatch({ selection: { anchor: r.pos }, scrollIntoView: true }) }
}

// ---- 格式化选区/全文 ----
async function formatJsonSelection() {
  if (!editorView) return
  const { from, to } = editorView.state.selection.main
  const text = from === to ? editorView.state.doc.toString() : editorView.state.sliceDoc(from, to)
  try {
    const mod = await import('../../../wailsjs/go/main/App')
    const r = await mod.FormatJSON(text, 2)
    if (r && r.content) {
      if (from === to) editorView.dispatch({ changes: { from: 0, to: editorView.state.doc.length, insert: r.content } })
      else editorView.dispatch({ changes: { from, to, insert: r.content } })
    }
  } catch (e: any) { ElMessage.error(e?.message || 'JSON 格式化失败') }
}
function formatXmlSelection() {
  if (!editorView) return
  const { from, to } = editorView.state.selection.main
  const text = from === to ? editorView.state.doc.toString() : editorView.state.sliceDoc(from, to)
  const formatted = prettyPrintXml(text)
  if (!formatted) { ElMessage?.warning?.('XML 格式化失败'); return }
  if (from === to) editorView.dispatch({ changes: { from: 0, to: editorView.state.doc.length, insert: formatted } })
  else editorView.dispatch({ changes: { from, to, insert: formatted } })
}
function prettyPrintXml(xml: string): string {
  try {
    const parser = new DOMParser()
    const doc = parser.parseFromString(xml, 'text/xml')
    if (doc.getElementsByTagName('parsererror').length) return ''
    const PADDING = '  '
    const serialize = (node: Element, level: number): string => {
      const pad = PADDING.repeat(level)
      let out = `${pad}<${node.nodeName}`
      for (const attr of Array.from(node.attributes)) out += ` ${attr.name}="${attr.value}"`
      const children = Array.from(node.children)
      const text = (node.textContent || '').trim()
      if (children.length === 0 && !text) { out += ' />\n'; return out }
      if (children.length === 0) { out += `>${text}</${node.nodeName}>\n`; return out }
      out += '>\n'
      for (const c of children) out += serialize(c, level + 1)
      out += `${pad}</${node.nodeName}>\n`
      return out
    }
    const root = doc.documentElement
    return `<?xml version="1.0" encoding="UTF-8"?>\n` + serialize(root, 0)
  } catch { return '' }
}

// ---- 显示全部符号 ----
function toggleShowAll() {
  const show = !(settingStore.config?.editor?.showWhitespace)
  toggleShowWhitespace(show)
  toggleEol(show)
  if (settingStore.config) { settingStore.config.editor.showWhitespace = show; settingStore.config.editor.showEol = show; settingStore.saveConfig() }
}

// ---- URL 高亮切换 ----
function toggleWebAddr() {
  if (!editorView || !settingStore.config) return
  const on = !settingStore.config.ui.showWebAddr
  settingStore.config.ui.showWebAddr = on
  settingStore.saveConfig()
  editorView.dispatch({ effects: webAddrCompartment.reconfigure(on ? [webAddrField] : []) })
}

// ---- EOL display ----
function toggleEol(show: boolean) {
  if (!editorView) return
  if (show) editorView.dom.style.setProperty('--show-eol', "'¶'")
  else editorView.dom.style.removeProperty('--show-eol')
}

// ---- MD mode ----
function toggleMdMode() {
  if (isMarkdown.value) {
    const modes: MdViewMode[] = ['edit', 'split', 'preview']
    mdMode.value = modes[(modes.indexOf(mdMode.value) + 1) % 3]
  }
}

// ---- 导出 Markdown 为 HTML 文件 ----
async function exportMdHtml() {
  const html = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<title>${props.tab.name || 'markdown'}</title>
<style>
body{font-family:-apple-system,'Microsoft YaHei',sans-serif;max-width:820px;margin:40px auto;padding:0 16px;line-height:1.7;color:#333;}
pre{background:#f5f5f5;padding:12px;border-radius:4px;overflow:auto;}
code{font-family:Consolas,monospace;}
pre code{background:none;padding:0;}
code{background:#f5f5f5;padding:2px 4px;border-radius:3px;}
blockquote{border-left:4px solid #ddd;margin:0;padding-left:16px;color:#666;}
table{border-collapse:collapse;}
th,td{border:1px solid #ddd;padding:6px 12px;}
.code-block-header{display:flex;justify-content:space-between;align-items:center;background:#2d2d2d;color:#ccc;padding:4px 8px;border-radius:4px 4px 0 0;font-size:12px;}
.code-block-wrapper pre{margin:0;border-radius:0 0 4px 4px;}
.code-copy-btn{background:transparent;border:1px solid #555;color:#ccc;padding:2px 8px;border-radius:3px;cursor:pointer;display:flex;align-items:center;gap:4px;}
</style>
</head>
<body>
${renderedHtml.value}
</body>
</html>`
  try {
    const { SaveFileDialog, SaveFile } = await import('../../../wailsjs/go/main/App')
    const p = await SaveFileDialog((props.tab.name || 'markdown').replace(/\.md$/i, '') + '.html')
    if (p) { await SaveFile(p, html, 'UTF-8'); ElMessage.success('已导出 HTML') }
  } catch (e: any) { ElMessage.error('导出失败：' + (e?.message || '')) }
}

// ---- Get helpers ----
function getLineCount(): number { return editorView?.state.doc.lines || 0 }
function getSelectedText(): string { return editorView ? editorView.state.sliceDoc(editorView.state.selection.main.from, editorView.state.selection.main.to) : '' }

defineExpose({
  gotoLine, showGotoLineDialog, toggleWordWrap, toggleShowWhitespace,
  transformCase, lineOperation, convertTabsSpaces, trimWhitespace, sortLines,
  indentLines, dedentLines, getLineCount, getSelectedText,
  highlightWordAtCursor, clearWordHighlight, toggleBookmark,
  gotoNextBookmark, gotoPrevBookmark, clearAllBookmarks,
  enterColumnMode, doColumnBlockEdit, undoAction, redoAction,
})

// ---- Markdown renderer ----
// 🆕 V2.0.0 初始化 Mermaid（延迟到组件挂载时，根据当前主题设置）
let mermaidInitialized = false
function initMermaid() {
  const isDark = colors.value?.isDark ?? false
  mermaid.initialize({
    startOnLoad: false,
    theme: isDark ? 'dark' : 'default',
    securityLevel: 'loose',
    fontFamily: 'inherit',
  })
  mermaidInitialized = true
}

const md = new MarkdownIt({ html: false, linkify: true, typographer: true })

const mermaidCounter = { n: 0 }

// 🆕 V2.0.0 代码块语法高亮：已知语言用指定语言，未知语言自动检测，失败则转义原文
function highlightCode(code: string, lang: string): string {
  if (!code.trim()) return ''
  const langName = lang.trim().split(/\s+/)[0]
  try {
    if (langName && hljs.getLanguage(langName)) {
      return hljs.highlight(code, { language: langName, ignoreIllegals: true }).value
    }
    return hljs.highlightAuto(code).value
  } catch {
    return md.utils.escapeHtml(code)
  }
}

// 🆕 V2.0.0 自定义 fence 规则：mermaid 代码块渲染为占位 div
md.renderer.rules.fence = (tokens, idx, options, env, self) => {
  const token = tokens[idx]
  const code = token.content
  const encoded = encodeURIComponent(code)
  const lang = token.info.trim()

  // mermaid 图表
  if (lang === 'mermaid') {
    const id = `mermaid-preview-${mermaidCounter.n++}`
    return `<div class="mermaid-container my-4 p-2 bg-gray-50 dark:bg-[#1a1a2e] rounded border border-gray-200 dark:border-gray-700">
      <div class="mermaid-chart" id="${id}">${md.utils.escapeHtml(code)}</div>
      <div class="flex items-center justify-between mt-1 px-1">
        <span class="text-[10px] text-gray-400">Mermaid</span>
        <button class="code-copy-btn text-[10px] text-gray-400 hover:text-gray-600" data-code="${encoded}" title="复制代码">
          <span class="code-copy-text">复制</span>
        </button>
      </div>
    </div>`
  }

  const langName = lang.trim().split(/\s+/)[0]
  const langLabel = langName ? `<span class="code-lang-label">${md.utils.escapeHtml(langName)}</span>` : ''
  const highlighted = highlightCode(code, langName)
  const codeHtml = `<pre><code class="hljs${langName ? ' language-' + md.utils.escapeHtml(langName) : ''}">${highlighted}</code></pre>`
  return `<div class="code-block-wrapper"><div class="code-block-header">${langLabel}<button class="code-copy-btn" data-code="${encoded}" title="复制代码"><svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg><span class="code-copy-text">复制</span></button></div>${codeHtml}</div>`
}

// 🆕 V2.0.0 KaTeX math rendering
function renderMath(text: string): string {
  // Block math: $$...$$
  text = text.replace(/\$\$([\s\S]*?)\$\$/g, (_: string, formula: string) => {
    try { return katex.renderToString(formula.trim(), { displayMode: true, throwOnError: false }) }
    catch { return `<pre class="katex-error">${formula}</pre>` }
  })
  // Inline math: $...$ (but not $$)
  text = text.replace(/(?<!\$)\$(?!\$)([^$]+?)\$(?!\$)/g, (_: string, formula: string) => {
    try { return katex.renderToString(formula.trim(), { displayMode: false, throwOnError: false }) }
    catch { return `$${formula}$` }
  })
  return text
}

const renderedHtml = computed(() => {
  if (!isMarkdown.value) return ''
  let html = md.render(props.tab.content || '')
  // 🆕 V2.0.0 渲染 KaTeX 数学公式
  html = renderMath(html)
  return html
})

function handlePreviewClick(e: MouseEvent) {
  const btn = (e.target as HTMLElement).closest('.code-copy-btn') as HTMLElement | null
  if (!btn) return
  const encoded = btn.getAttribute('data-code')
  if (!encoded) return
  try {
    navigator.clipboard.writeText(decodeURIComponent(encoded)).then(() => {
      btn.classList.add('copied')
      const s = btn.querySelector('.code-copy-text')
      if (s) s.textContent = '已复制'
      setTimeout(() => { btn.classList.remove('copied'); if (s) s.textContent = '复制' }, 2000)
    })
  } catch (e) { console.warn(e) }
}

// ---- 命令别名：兼容历史遗留的多种命名，统一映射到编辑器标准命令 ----
// 菜单/键盘/工具栏曾用 line-dup / line-del / tab2space 等旧名，与编辑器标准名不一致，导致命令失效
const CMD_ALIASES: Record<string, string> = {
  // 行操作
  'line-dup': 'line-duplicate',
  'line-del': 'line-remove',
  'line-delete': 'line-remove',
  'line-rmdup': 'line-removeDuplicate',
  'line-up': 'line-moveUp',
  'line-down': 'line-moveDown',
  'line-move-up': 'line-moveUp',
  'line-move-down': 'line-moveDown',
  'line-rmempty': 'line-removeEmpty',
  'line-rmblank': 'line-removeBlank',
  'line-insert-above': 'line-insertAbove',
  'line-insert-below': 'line-insertBelow',
  // Tab/空格转换
  'tab2space': 'tab-to-spaces',
  'space2tab-all': 'spaces-all-to-tabs',
  'space2tab-lead': 'spaces-leading-to-tabs',
  // 显示空格/Tab、行尾符 → 切换
  'show-spaces': 'toggle-whitespace',
  'show-eol': 'toggle-eol',
}

// ---- Global command handler ----
function handleEditorCommand(e: Event) {
  const detail = (e as CustomEvent).detail
  if (!detail) return
  let cmd: string, args: any[] = []
  if (typeof detail === 'string') cmd = detail
  else { cmd = detail.cmd; args = detail.args || [] }
  if (!cmd) return
  cmd = CMD_ALIASES[cmd] || cmd

  // Standard operations
  if (cmd === 'undo') undoAction()
  else if (cmd === 'redo') redoAction()
  else if (cmd === 'cut') doEditorAction('cut')
  else if (cmd === 'copy') doEditorAction('copy')
  else if (cmd === 'paste') doEditorAction('paste')
  else if (cmd === 'select-all') doEditorAction('selectAll')
  else if (cmd === 'find-next') doFindAction('next')
  else if (cmd === 'find-prev') doFindAction('prev')
  else if (cmd === 'show-goto-line') showGotoLineDialog()
  else if (cmd.startsWith('case-')) transformCase(cmd.replace('case-', ''))
  else if (cmd.startsWith('line-')) lineOperation(cmd.replace('line-', ''))
  else if (cmd.startsWith('sort-')) sortLines(cmd.replace('sort-', ''))
  else if (cmd.startsWith('trim-')) trimWhitespace(cmd.replace('trim-', ''))
  else if (cmd === 'wordwrap-on') toggleWordWrap(true)
  else if (cmd === 'wordwrap-off') toggleWordWrap(false)
  else if (cmd === 'show-whitespace') toggleShowWhitespace(true)
  else if (cmd === 'hide-whitespace') toggleShowWhitespace(false)
  else if (cmd === 'toggle-bookmark') toggleBookmark()
  else if (cmd === 'next-bookmark') gotoNextBookmark()
  else if (cmd === 'prev-bookmark') gotoPrevBookmark()
  else if (cmd === 'clear-bookmarks') clearAllBookmarks()
  else if (cmd === 'clear-mark' || cmd === 'clear-highlight' || cmd === 'clear-all-highlight') clearWordHighlight()
  else if (cmd === 'word-highlight') highlightWordAtCursor()
  else if (cmd === 'mark-color') highlightWordAtCursor()
  // ---- 多色标记 ----
  else if (cmd === 'mark-all') markAll(lastSearchTerm)
  else if (cmd === 'mark-red') { currentMarkColor = 1; markSelectionOrWord() }
  else if (cmd === 'mark-yellow') { currentMarkColor = 0; markSelectionOrWord() }
  else if (cmd === 'mark-blue') { currentMarkColor = 2; markSelectionOrWord() }
  else if (cmd === 'mark-1') { currentMarkColor = 0; markSelectionOrWord() }
  else if (cmd === 'mark-2') { currentMarkColor = 1; markSelectionOrWord() }
  else if (cmd === 'mark-3') { currentMarkColor = 2; markSelectionOrWord() }
  else if (cmd === 'mark-4') { currentMarkColor = 3; markSelectionOrWord() }
  else if (cmd === 'mark-5') { currentMarkColor = 4; markSelectionOrWord() }
  else if (cmd === 'mark-loop') { currentMarkColor = (currentMarkColor + 1) % 5; markSelectionOrWord() }
  else if (cmd === 'clear-all-marks' || cmd === 'clear-marks') { if (editorView) editorView.dispatch({ effects: clearMarksEffect.of(null) }) }
  else if (cmd === 'mark-keywords' && args[0]) markKeywords(args[0] as string[])
  // ---- 括号跳转 ----
  else if (cmd === 'goto-bracket') gotoBracket()
  // ---- 位置历史 ----
  else if (cmd === 'prev-position') gotoPrevPosition()
  else if (cmd === 'next-position') gotoNextPosition()
  // ---- 格式化 ----
  else if (cmd === 'format-json') formatJsonSelection()
  else if (cmd === 'format-xml') formatXmlSelection()
  // ---- 显示全部符号 ----
  else if (cmd === 'show-all') toggleShowAll()
  // ---- URL 高亮 ----
  else if (cmd === 'toggle-webaddr') toggleWebAddr()
  // ---- 列块模式（真·多光标矩形选择）----
  else if (cmd === 'column-mode') toggleColumnMode()
  else if (cmd === 'column-block') { document.dispatchEvent(new CustomEvent('show-column-edit')) }
  else if (cmd === 'indent') indentLines()
  else if (cmd === 'dedent') dedentLines()
  else if (cmd === 'toggle-md-mode') toggleMdMode()
  else if (cmd === 'show-eol') toggleEol(true)
  else if (cmd === 'hide-eol') toggleEol(false)
  else if (cmd === 'tab-to-spaces') convertTabsSpaces('tabToSpaces')
  else if (cmd === 'spaces-all-to-tabs') convertTabsSpaces('spacesAllToTabs')
  else if (cmd === 'spaces-leading-to-tabs') convertTabsSpaces('spacesLeadingToTabs')
  // ---- 自动换行 / 显示空格 / 行尾符（真正的切换）----
  else if (cmd === 'toggle-word-wrap') toggleWordWrap()
  else if (cmd === 'toggle-whitespace') {
    const on = !(settingStore.config?.editor?.showWhitespace)
    toggleShowWhitespace(on)
    if (settingStore.config) { settingStore.config.editor.showWhitespace = on; settingStore.saveConfig() }
  }
  else if (cmd === 'toggle-eol') {
    const on = !(settingStore.config?.editor?.showEol)
    toggleEol(on)
    if (settingStore.config) { settingStore.config.editor.showEol = on; settingStore.saveConfig() }
  }
  // ---- 文档地图（minimap）切换 ----
  else if (cmd === 'toggle-minimap') showMinimap.value = !showMinimap.value
  // ---- 书签行操作（notepad-- 剪切/复制/粘贴/删除书签行）----
  else if (cmd === 'copy-bookmark-lines') copyBookmarkLines()
  else if (cmd === 'cut-bookmark-lines') cutBookmarkLines()
  else if (cmd === 'delete-bookmark-lines') deleteBookmarkLines()
  else if (cmd === 'delete-unbookmark-lines') deleteUnbookmarkLines()
  else if (cmd === 'paste-bookmark-lines') pasteBookmarkLines()
  else if (cmd === 'insert-text' && args[0])
    editorView?.dispatch({ changes: { from: editorView.state.selection.main.head, insert: String(args[0]) } })
  else if (cmd === 'goto-line' && args[0])
    gotoLine(args[0] as number)
  else if (cmd === 'scroll-to-pos' && args[0] && editorView)
    editorView.dispatch({ selection: { anchor: args[0] }, scrollIntoView: true })
  else if (cmd === 'scroll-to-line' && args[0])
    gotoLine(args[0] as number)
  else if (cmd === 'scroll-to-end' && editorView)
    editorView.dispatch({ selection: { anchor: editorView.state.doc.length }, scrollIntoView: true })
  else if (cmd === 'insert-blank-above') insertBlankLine(true)
  else if (cmd === 'insert-blank-below') insertBlankLine(false)
  else if (cmd === 'line-duplicate') lineOperation('duplicate')
  else if (cmd === 'line-delete') lineOperation('remove')
  else if (cmd === 'line-moveUp') lineOperation('moveUp')
  else if (cmd === 'line-moveDown') lineOperation('moveDown')
  else if (cmd === 'line-removeEmpty') lineOperation('removeEmpty')
  else if (cmd === 'line-removeEmptyCbc') lineOperation('removeBlank')
  else if (cmd === 'line-reverse') lineOperation('reverse')
  else if (cmd === 'line-split') lineOperation('split')
  else if (cmd === 'line-join') lineOperation('join')
  else if (cmd === 'line-removeDuplicate') lineOperation('removeDuplicate')
  else if (cmd === 'comment-line') { if (editorView) toggleComment(editorView) }
  else if (cmd === 'comment-block') { if (editorView) toggleBlockComment(editorView) }
  else if (cmd === 'column-insert-text' && args[0]) {
    // 多光标列插入：每个选区范围都插入相同文本
    const text = String(args[0])
    if (!editorView) return
    const ranges = [...editorView.state.selection.ranges].sort((a, b) => b.from - a.from)
    if (ranges.length <= 1) {
      editorView.dispatch({ changes: { from: editorView.state.selection.main.head, insert: text } })
    } else {
      const changes = ranges.map(r => ({ from: r.from, insert: text }))
      editorView.dispatch({ changes })
    }
  }
  else if (cmd === 'column-insert-num' && args[0]) {
    const opts = args[0]
    let val = opts.init
    const lines: string[] = []
    for (let i = 0; i < opts.repeat; i++) {
      let s = val.toString(opts.radix || 10)
      if (opts.radix === 16 && opts.capital) s = s.toUpperCase()
      lines.push((opts.prefix || '') + s)
      val += opts.inc || 1
    }
    editorView?.dispatch({ changes: { from: editorView.state.selection.main.head, insert: lines.join('\n') } })
  }
}

function doEditorAction(action: 'cut' | 'copy' | 'paste' | 'selectAll') {
  if (!editorView) return
  if (action === 'selectAll') editorView.dispatch({ selection: { anchor: 0, head: editorView.state.doc.length } })
  else {
    if (action === 'cut' || action === 'copy') {
      const { from, to } = editorView.state.selection.main
      if (from !== to) editorStore.pushClipboard(editorView.state.sliceDoc(from, to))
    }
    document.execCommand(action)
  }
}

function doFindAction(dir: 'next' | 'prev') {
  if (!editorView || !lastSearchTerm) return
  const state = editorView.state
  const doc = state.doc.toString()
  const from = state.selection.main.head
  const regex = new RegExp(escapeRegExp(lastSearchTerm), 'gi')
  let target: number | null = null
  if (dir === 'next') {
    const after = doc.slice(from).match(regex)
    target = after ? from + after.index! : doc.match(regex)?.index ?? null
    if (target !== null && target < from) target = doc.match(regex)?.index ?? null
  } else {
    const matches = [...doc.slice(0, from).matchAll(new RegExp(escapeRegExp(lastSearchTerm), 'gi'))]
    target = matches.length > 0 ? matches[matches.length - 1].index! : [...doc.matchAll(new RegExp(escapeRegExp(lastSearchTerm), 'gi'))].pop()?.index ?? null
  }
  if (target !== null) {
    editorView.dispatch({ selection: { anchor: target, head: target + lastSearchTerm.length }, scrollIntoView: true })
    editorView.focus()
  }
}

function escapeRegExp(s: string) { return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') }

// ---- Lifecycle ----
// 宏回放：监听 isPlaying，逐条应用步骤（以当前光标相对回放，忽略存储的绝对 from/to）
watch(() => editorStore.macroState.isPlaying, (playing) => {
  if (!playing) return
  const run = () => {
    if (!editorStore.macroState.isPlaying || !editorView) return
    const step = editorStore.getNextMacroStep()
    if (!step) return
    const head = editorView.state.selection.main.head
    switch (step.type) {
      case 'insert':
        if (step.text) editorView.dispatch({ changes: { from: head, insert: step.text } })
        break
      case 'delete': {
        // 删除当前光标前 text.length 个字符（模拟退格删除）
        const len = step.text?.length || (step.to && step.from ? step.to - step.from : 0)
        if (len > 0) editorView.dispatch({ changes: { from: Math.max(0, head - len), to: head } })
        break
      }
      default:
        break
    }
    if (editorStore.macroState.isPlaying) requestAnimationFrame(run)
  }
  requestAnimationFrame(run)
})

// ==================== Tab switching (in-place, no editor recreation) ====================
function saveEditorState() {
  if (!editorView) return
  const pos = editorView.state.selection.main.head
  const line = editorView.state.doc.lineAt(pos)
  editorStore.updateCursorPosition(props.tab.id, line.number, pos - line.from + 1)
  editorStore.updateScrollPosition(props.tab.id,
    editorView.scrollDOM.scrollTop,
    editorView.scrollDOM.scrollLeft)
}

function restoreEditorState() {
  if (!editorView) return
  const doc = editorView.state.doc
  const savedLine = props.tab.cursorPosition?.line || 1
  const savedCol = props.tab.cursorPosition?.column || 1
  const line = doc.line(Math.min(savedLine, doc.lines))
  const pos = line.from + Math.min(Math.max(savedCol - 1, 0), line.length)
  editorView.dispatch({
    selection: { anchor: pos },
    scrollIntoView: true,
  })
  if (props.tab.scrollPosition?.top) {
    editorView.scrollDOM.scrollTo({
      top: props.tab.scrollPosition.top,
      left: props.tab.scrollPosition.left || 0,
    })
  }
}

function updateLanguageForTab() {
  if (!editorView) return
  const langEx = getLanguageExtension(props.tab.language)
  editorView.dispatch({ effects: langCompartment.reconfigure(langEx) })
}

watch(() => props.tab.id, (newId, oldId) => {
  if (!editorView) {
    createEditor()
    return
  }
  // Save old tab's editor state before switching
  if (oldId) saveEditorState()
  // Switch content to new tab
  isInitializing = true
  editorView.dispatch({
    changes: { from: 0, to: editorView.state.doc.length, insert: props.tab.content },
  })
  updateLanguageForTab()
  restoreEditorState()
  nextTick(() => {
    isInitializing = false
    editorView?.focus()
  })
})

watch(() => props.tab.content, (nc) => {
  if (editorView && editorView.state.doc.toString() !== nc) {
    isInitializing = true
    editorView.dispatch({ changes: { from: 0, to: editorView.state.doc.length, insert: nc } })
    Promise.resolve().then(() => { isInitializing = false })
  }
})
watch(() => settingStore.config?.theme.currentTheme, reconfigureAppearance)
watch(() => config.value?.editor?.tabSize, (s) => { if (editorView && s) editorView.dispatch({ effects: tabSizeCompartment.reconfigure(EditorState.tabSize.of(s)) }) })
watch(() => config.value?.editor?.wordWrap, (w) => { if (editorView) editorView.dispatch({ effects: wordWrapCompartment.reconfigure(w ? EditorView.lineWrapping : []) }) })
watch(() => config.value?.editor?.fontSize, reconfigureAppearance)
watch(() => config.value?.editor?.fontFamily, reconfigureAppearance)

onMounted(() => {
  createEditor()
  window.addEventListener('editor-command', handleEditorCommand)
})

// 文档地图：内容变化后刷新 scrollHeight（避免视口指示器比例失真）
watch(() => props.tab.content, async () => {
  await nextTick()
  if (!editorView) return
  minimapViewport.value = {
    scrollTop: editorView.scrollDOM.scrollTop,
    scrollHeight: editorView.scrollDOM.scrollHeight,
    clientHeight: editorView.scrollDOM.clientHeight,
  }
})

// 🆕 V2.0.0 渲染 Mermaid 图表（DOM 更新后）
watch(renderedHtml, async () => {
  await nextTick()
  if (!isMarkdown.value) return
  initMermaid()
  const container = document.getElementById(`preview-${props.tab.id}`)
  if (!container) return
  const elements = container.querySelectorAll('.mermaid-chart')
  for (const el of elements) {
    const id = el.id
    if (!id || !el.textContent) continue
    try {
      const { svg } = await mermaid.render(id + '-svg', el.textContent.trim())
      el.innerHTML = svg
    } catch (e) {
      el.innerHTML = `<pre class="text-red-500 text-xs">Mermaid 渲染错误</pre>`
    }
  }
})

onUnmounted(() => {
  window.removeEventListener('editor-command', handleEditorCommand)
  editorView?.destroy()
})
</script>

<template>
  <div class="h-full flex flex-col overflow-hidden">
    <!-- MD toolbar -->
    <div v-if="isMarkdown" class="flex items-center gap-1 px-3 py-1 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#2d2d2d] flex-shrink-0">
      <span class="text-xs text-gray-400 mr-2">Markdown</span>
      <button class="px-2 py-0.5 text-xs rounded" :class="mdMode==='edit'?'bg-blue-500 text-white':'text-gray-500 hover:bg-gray-200 dark:hover:bg-gray-600 dark:text-gray-400'" @click="mdMode='edit'">编辑</button>
      <button class="px-2 py-0.5 text-xs rounded" :class="mdMode==='split'?'bg-blue-500 text-white':'text-gray-500 hover:bg-gray-200 dark:hover:bg-gray-600 dark:text-gray-400'" @click="mdMode='split'">分屏</button>
      <button class="px-2 py-0.5 text-xs rounded" :class="mdMode==='preview'?'bg-blue-500 text-white':'text-gray-500 hover:bg-gray-200 dark:hover:bg-gray-600 dark:text-gray-400'" @click="mdMode='preview'">预览</button>
      <span class="flex-1"></span>
      <button class="px-2 py-0.5 text-xs rounded text-gray-500 hover:bg-gray-200 dark:hover:bg-gray-600 dark:text-gray-400" @click="exportMdHtml">导出 HTML</button>
    </div>

    <div class="flex-1 flex overflow-hidden" style="position:relative;">
      <div ref="editorContainer" :class="{'w-full':!isMarkdown||mdMode==='edit','cm-container-split border-r border-gray-200 dark:border-gray-700':isMarkdown&&mdMode==='split','hidden':isMarkdown&&mdMode==='preview','with-minimap':showMinimap}" class="cm-container" @dblclick="highlightWordAtCursor"></div>

      <!-- Goto line dialog -->
      <Teleport to="body">
        <div v-if="showGotoLine" class="fixed inset-0 z-50 flex items-center justify-center bg-black/20" @click.self="showGotoLine=false" @keydown.enter="handleGotoLine" @keydown.escape="showGotoLine=false">
          <div class="bg-white dark:bg-[#2d2d2d] rounded-lg shadow-2xl p-6 w-80 border border-gray-200 dark:border-gray-600">
            <h3 class="text-sm font-medium mb-4 dark:text-gray-200">跳转到行</h3>
            <div class="flex gap-2">
              <input v-model="gotoLineInput" type="number" min="1" :max="getLineCount()" placeholder="输入行号..." class="goto-line-input flex-1 px-3 py-1.5 text-sm bg-gray-50 dark:bg-[#3c3c3c] border border-gray-200 dark:border-gray-600 rounded focus:outline-none focus:border-blue-500 dark:text-gray-200" @keydown.enter="handleGotoLine" @keydown.escape="showGotoLine=false"/>
              <button class="px-3 py-1.5 text-sm bg-blue-500 text-white rounded hover:bg-blue-600" @click="handleGotoLine">跳转</button>
            </div>
            <p class="text-xs text-gray-400 mt-2">总行数: {{ getLineCount() }}</p>
          </div>
        </div>
      </Teleport>

      <!-- MD preview -->
      <div v-if="isMarkdown&&(mdMode==='preview'||mdMode==='split')" class="overflow-auto p-4 bg-white dark:bg-[#1e1e1e] select-text" :class="mdMode==='split'?'w-1/2':'w-full'" @click="handlePreviewClick">
        <div class="markdown-body" v-html="renderedHtml"></div>
      </div>

      <!-- 文档地图（minimap） -->
      <Minimap v-if="showMinimap" :content="props.tab.content" :viewport="minimapViewport" @seek="gotoLine" />
    </div>
  </div>
</template>

<style scoped>
.cm-container { position: absolute; top: 0; left: 0; right: 0; bottom: 0; overflow: hidden; }
.cm-container.with-minimap { right: 90px; }
/* 分屏模式：编辑器脱离绝对定位回到 flex 流内占左半，避免绝对定位铺满遮住右侧预览 */
.cm-container-split { position: relative; top: auto; left: auto; right: auto; bottom: auto; width: 50%; flex-shrink: 0; }

/* ---- 代码块外壳（配合 github-markdown-css）---- */
.markdown-body :deep(.code-block-wrapper) { position: relative; margin: .8em 0; border-radius: 6px; border: 1px solid #d1d9e0; overflow: hidden; }
.markdown-body :deep(.code-block-header) { display: flex; align-items: center; justify-content: space-between; padding: 4px 12px; background: #f6f8fa; border-bottom: 1px solid #d1d9e0; font-size: 12px; }
.markdown-body :deep(.code-lang-label) { color: #59636e; font-family: monospace; font-size: 11px; }
.markdown-body :deep(.code-copy-btn) { display: inline-flex; align-items: center; gap: 4px; padding: 2px 8px; border: none; border-radius: 4px; background: transparent; color: #59636e; cursor: pointer; font-size: 12px; transition: all .15s; }
.markdown-body :deep(.code-copy-btn:hover) { background: rgba(0,0,0,.08); color: #1f2328; }
.markdown-body :deep(.code-copy-btn.copied) { color: #16a34a; background: rgba(22,163,74,.08); }

/* 代码块内部的 pre/code 与外壳对齐（覆盖 github-markdown-css 默认边框/背景） */
.markdown-body :deep(.code-block-wrapper pre) { margin: 0; border: none; border-radius: 0; background: #f6f8fa; }
.markdown-body :deep(.code-block-wrapper pre code.hljs) { display: block; padding: 1em; background: transparent; overflow-x: auto; }

/* 图片自适应 */
.markdown-body :deep(img) { max-width: 100%; border-radius: 4px; }

/* 用户可选择文本 */
.markdown-body :deep(*) { -webkit-user-select: text; user-select: text; }
.markdown-body :deep(*)::selection { background: #b3d7ff; }

/* ---- 暗色模式：覆盖 github-markdown-css 的硬编码亮色 ---- */
html.dark .markdown-body { color: #e6edf3; background-color: transparent; }
html.dark .markdown-body :deep(a) { color: #4493f8; }
html.dark .markdown-body :deep(h1),html.dark .markdown-body :deep(h2) { border-bottom-color: #30363d; }
html.dark .markdown-body :deep(h6) { color: #8b949e; }
html.dark .markdown-body :deep(code) { background-color: rgba(110,118,129,.4); }
html.dark .markdown-body :deep(blockquote) { border-left-color: #30363d; color: #8b949e; }
html.dark .markdown-body :deep(th),html.dark .markdown-body :deep(td) { border-color: #30363d; }
html.dark .markdown-body :deep(hr) { background-color: #30363d; }
html.dark .markdown-body :deep(mark) { background-color: rgba(187,128,9,.15); }

/* 代码块外壳暗色 */
html.dark .markdown-body :deep(.code-block-wrapper) { border-color: #30363d; }
html.dark .markdown-body :deep(.code-block-header) { background: #161b22; border-bottom-color: #30363d; }
html.dark .markdown-body :deep(.code-lang-label) { color: #8b949e; }
html.dark .markdown-body :deep(.code-block-wrapper pre) { background: #161b22; }
html.dark .markdown-body :deep(.code-copy-btn) { color: #8b949e; }
html.dark .markdown-body :deep(.code-copy-btn:hover) { background: rgba(255,255,255,.08); color: #e6edf3; }

/* 暗色语法高亮（GitHub dark 配色） */
html.dark .markdown-body :deep(.hljs) { color: #e6edf3; }
html.dark .markdown-body :deep(.hljs-keyword),
html.dark .markdown-body :deep(.hljs-literal),
html.dark .markdown-body :deep(.hljs-doctag) { color: #ff7b72; }
html.dark .markdown-body :deep(.hljs-string),
html.dark .markdown-body :deep(.hljs-regexp) { color: #a5d6ff; }
html.dark .markdown-body :deep(.hljs-comment),
html.dark .markdown-body :deep(.hljs-quote) { color: #8b949e; font-style: italic; }
html.dark .markdown-body :deep(.hljs-number),
html.dark .markdown-body :deep(.hljs-symbol) { color: #79c0ff; }
html.dark .markdown-body :deep(.hljs-title),
html.dark .markdown-body :deep(.hljs-function),
html.dark .markdown-body :deep(.hljs-section) { color: #d2a8ff; }
html.dark .markdown-body :deep(.hljs-built_in),
html.dark .markdown-body :deep(.hljs-type),
html.dark .markdown-body :deep(.hljs-class),
html.dark .markdown-body :deep(.hljs-attr),
html.dark .markdown-body :deep(.hljs-variable),
html.dark .markdown-body :deep(.hljs-template-variable) { color: #ffa657; }
html.dark .markdown-body :deep(.hljs-attribute),
html.dark .markdown-body :deep(.hljs-selector-tag),
html.dark .markdown-body :deep(.hljs-meta) { color: #79c0ff; }
html.dark .markdown-body :deep(.hljs-addition) { color: #aff5b4; }
html.dark .markdown-body :deep(.hljs-deletion) { color: #ffdcd7; }
html.dark .markdown-body :deep(*)::selection { background: #264f78; }
</style>
