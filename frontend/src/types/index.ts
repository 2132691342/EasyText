// ============================================================
// EasyText 类型定义 - 完全仿照 notepad-- 数据模型
// 参考：notepad-- src/cceditor/ccnotepad.h, src/rcglobal.h
// ============================================================

// ---------- 文件相关 ----------
export interface FileInfo {
  path: string
  name: string
  ext: string
  size: number
  modified: string
  encoding: string
  lineEnding: string
  lineCount: number
  isReadOnly: boolean
  isDir: boolean
}

export interface ReadResult {
  content: string
  info: FileInfo
  detectedEncoding: string
}

export interface WriteResult {
  path: string
  size: number
  success: boolean
}

// 文件树节点（notepad-- filemanager）
export interface TreeNode {
  path: string
  name: string
  isDir: boolean
  children?: TreeNode[]
  expanded: boolean
  ext?: string
}

export interface FileTree {
  root: TreeNode
  basePath: string
}

// ---------- 配置类型（notepad-- nddsetting）----------
export interface ColumnModeConfig {
  numberStart: number
  numberStep: number
  numberBase: number
  dateFormat: string
  caseConversion: string
}

export interface EditorConfig {
  fontSize: number
  fontFamily: string
  tabSize: number
  insertSpaces: boolean
  wordWrap: boolean
  lineNumbers: boolean
  autoSave: boolean
  autoSaveInterval: number
  highlightLine: boolean
  bracketPairColor: boolean
  minimap: boolean
  showIndentGuide: boolean
  showWhitespace: boolean
  showEol: boolean
  foldEnable: boolean
  columnMode: boolean
  virtualSpace: boolean
  scrollPastEnd: boolean
  columnModeConfig: ColumnModeConfig
}

// 19 套主题 - 完全匹配 notepad-- styleset.cpp
export type ThemeName =
  | 'Default'       // 默认亮色
  | 'DarkDefault'   // 默认暗色
  | 'Bespin'
  | 'BlackBoard'
  | 'Cobalt'
  | 'Dracula'
  | 'Eiffel'
  | 'Elegant'
  | 'ErlangDark'
  | 'IDLE'
  | 'Lazy'
  | 'Material'
  | 'Monokai'
  | 'MonoIndustrial'
  | 'Neat'
  | 'NightOwl'
  | 'OneDark'
  | 'Solarized'
  | 'SolarizedDark'

export interface ThemeConfig {
  currentTheme: ThemeName
  fontSize: number
  fontFamily: string
  tabSize: number
  autoTheme: boolean
}

export interface FileConfig {
  defaultEncoding: string
  autoDetectEncoding: boolean
  defaultLineEnding: string
  ignorePatterns: string[]
}

export interface UIConfig {
  language: string
  showFileTree: boolean
  showStatusBar: boolean
  showToolBar: boolean
  showFileListView: boolean
  showWebAddr: boolean
  fileTreeWidth: number
  zoomLevel: number
  toolbarIconSize: number
  favorites: string[]
  lastFolder: string
  statusBarItems: Record<string, boolean>
  toolbarItems: Record<string, boolean>
  recentFilesLimit: number
  // 🆕 关闭按钮行为：true 时最小化到托盘；false 时直接退出
  closeToTray?: boolean
}

export interface AppConfig {
  version: number
  editor: EditorConfig
  theme: ThemeConfig
  file: FileConfig
  ui: UIConfig
}

// ---------- 编辑器标签页（notepad-- TabInfo）----------
export type TabViewType = 'code' | 'image' | 'image-edit' | 'hex' | 'log' | 'markdown'

export interface EditorTab {
  id: string
  path: string
  name: string
  content: string
  originalContent: string
  isDirty: boolean
  encoding: string
  lineEnding: string
  language: string
  cursorPosition: { line: number; column: number }
  scrollPosition: { top: number; left: number }
  isReadOnly: boolean
  viewType: TabViewType
}

// ---------- 宏系统（notepad-- 宏录制/回放）----------
export interface MacroStep {
  type: 'insert' | 'delete' | 'replace' | 'selection' | 'cursor' | 'command' | 'find'
  text?: string
  from?: number
  to?: number
  anchor?: number
  head?: number
  timestamp: number
  command?: string
  args?: any[]
}

export interface Macro {
  id: string
  name: string
  steps: MacroStep[]
  createdAt: number
  modifiedAt: number
  tabId: string // 关联的标签页
}

export interface MacroState {
  isRecording: boolean
  isPlaying: boolean
  playingId: string   // 当前正在回放的宏 ID（响应式，替代模块级 let）
  currentMacro: MacroStep[]
  savedMacros: Macro[]
  recordStartTime: number
  playIndex: number
  loopCount: number
  currentLoop: number
}

// ---------- 查找替换（notepad-- findwin）----------
export interface FindReplaceOptions {
  search: string
  replace: string
  caseSensitive: boolean
  wholeWord: boolean
  useRegex: boolean
  wrapAround: boolean
  searchDirection: 'forward' | 'backward'
  scope: 'current' | 'all' | 'selection'
}

export interface FindMatch {
  index: number
  file: string
  line: number
  column: number
  content: string
  matchText: string
  matchLength: number
}

export interface FindInFileResult {
  file: string
  matches: FindMatch[]
  count: number
}

export interface FindState {
  isVisible: boolean
  mode: 'find' | 'replace' | 'files' | 'mark'
  options: FindReplaceOptions
  lastSearchTerm: string
  currentMatchIndex: number
  allMatches: FindMatch[]
  findInFileResults: FindInFileResult[]
  isSearching: boolean
}

// ---------- Diff 对比 ----------
export interface DiffLine {
  type: 'added' | 'removed' | 'modified' | 'unchanged'
  content: string
  oldLine: number
  newLine: number
  diffChars?: string
}

export interface DiffBlock {
  oldStart: number
  oldCount: number
  newStart: number
  newCount: number
  lines: DiffLine[]
}

export interface DiffResult {
  diffs: DiffItem[]
  added: number
  removed: number
  modified: number
  unchanged: number
}

export interface DiffItem {
  type: string
  content: string
  oldLine?: number
  newLine?: number
  position: number
}

// 对比模式状态
export interface CompareState {
  leftFile: string | null
  rightFile: string | null
  leftContent: string
  rightContent: string
  isVisible: boolean
  isDirCompare: boolean
  compareBlocks: DiffBlock[]
  configuredRules: CompareRuleConfig
  recentCompares: CompareRecord[]
}

export interface CompareRuleConfig {
  ignoreCase: boolean
  ignoreWhitespace: boolean
  ignoreEmptyLines: boolean
  detectMoved: boolean
  contextLines: number
}

export interface CompareRecord {
  leftFile: string
  rightFile: string
  timestamp: number
}

// ---------- 编码相关（notepad-- Encode.h）----------
export interface EncodingInfo {
  name: string
  displayName: string
  isUnicode: boolean
  bom: string
}

// ---------- 哈希计算 ----------
export interface HashResult {
  md5: string
  sha1: string
  sha256: string
  sha512: string
  sha3_256: string
  keccak256: string
}

// ---------- 批量重命名 ----------
export interface RenameRule {
  type: 'prefix' | 'suffix' | 'replace' | 'regex' | 'sequence' | 'case' | 'remove' | 'insert'
  value: string
  replaceFrom?: string
  replaceTo?: string
  caseType?: 'upper' | 'lower' | 'title'
  sequenceStart?: number
  sequenceStep?: number
  sequenceDigits?: number
  position?: number
}

export interface RenamePreview {
  oldName: string
  newName: string
  oldPath: string
  newPath: string
}

// ---------- 语言/语法（notepad-- langstyledefine）----------
export interface LanguageStyle {
  name: string
  extensions: string[]
  keywords: string[]
  lineComment: string
  blockCommentStart: string
  blockCommentEnd: string
  foldable: boolean
  caseSensitive: boolean
  autoIndent: boolean
}

export interface UserDefineLang {
  name: string
  ext: string
  keywords: string[]
  keywords2: string[]
  commentLine: string
  commentStart: string
  commentEnd: string
  operators: string
  delimiters: string
}

// ---------- 快捷键 ----------
export interface ShortcutDef {
  id: string
  name: string
  category: string
  defaultKey: string
  currentKey: string
}

export interface ShortcutState {
  shortcuts: ShortcutDef[]
  isEditing: boolean
  editingId: string | null
}

// ---------- 会话 ----------
export interface SessionFile {
  path: string
  encoding: string
  language: string
}

export interface Session {
  id: string
  files: SessionFile[]
  activeId: string
  timestamp: number
}

// ---------- 书签 ----------
export interface Bookmark {
  tabId: string
  line: number
  note?: string
}

// ---------- JSON 工具 ----------
export interface JSONError {
  message: string
  line: number
  column: number
  offset: number
}

export interface JSONResult {
  content?: string
  success: boolean
  error?: JSONError
}

// ---------- 通知 ----------
export interface Notification {
  id: string
  type: 'info' | 'success' | 'warning' | 'error'
  title: string
  message?: string
  duration?: number
}

// ---------- 菜单项 ----------
export interface MenuItem {
  id: string
  label: string
  shortcut?: string
  icon?: string
  action?: () => void
  submenu?: MenuItem[]
  disabled?: boolean
  checked?: boolean
}

// ---------- 列编辑 ----------
export interface ColumnEditOptions {
  type: 'text' | 'number'
  text?: string
  initNum?: number
  incNum?: number
  repeat: number
  prefix?: string
  radix?: number
  capital?: boolean
}

// ---------- 十六进制编辑 ----------
export interface HexState {
  bytes: number[]
  offset: number
  pageSize: number
  currentPage: number
  totalPages: number
  columns: number // 每行显示字节数（默认16）
  isEditing: boolean
  editPosition: number
  editNibble: 'high' | 'low'
  selection: { start: number; end: number } | null
}

// ---------- Markdown 预览 ----------
export type MdViewMode = 'edit' | 'preview' | 'split'

// ========== 🆕 V2.0.0 新增类型 ==========

// ---------- 草稿系统 ----------
export interface DraftEntry {
  id: number
  filePath: string
  content: string
  encoding: string
  lineEnding: string
  savedAt: string
  fileModtime: number
}

// ---------- 最近访问 ----------
export interface RecentEntry {
  path: string
  isFolder: boolean
  name: string
  accessedAt: string
}

// ---------- 代码片段 ----------
export interface Snippet {
  id: number
  name: string
  prefix: string
  body: string
  description: string
  language: string
  createdAt: string
  updatedAt: string
}

// ---------- 持久化书签 ----------
export interface BookmarkEntry {
  id: number
  filePath: string
  lineNumber: number
  note: string
  tag: string
  createdAt: string
}

// ---------- JSONPath 查询 ----------
export interface JSONPathResult {
  path: string
  value: any
  type: string
}

// ---------- JSON 结构化 Diff ----------
export interface JSONDiffEntry {
  path: string
  type: 'added' | 'removed' | 'modified' | 'unchanged'
  oldValue?: any
  newValue?: any
}

export interface JSONDiffResult {
  entries: JSONDiffEntry[]
  summary: {
    added: number
    removed: number
    modified: number
  }
}

// ---------- 日志查看器 ----------
export interface LogViewerConfig {
  isTailing: boolean
  levelFilter: {
    error: boolean
    warn: boolean
    info: boolean
    debug: boolean
    trace: boolean
  }
  keywords: string[]
  chunkSize: number
}

// ---------- 终端面板 ----------
export type TerminalType = 'cmd' | 'powershell'

// ---------- 全局搜索 ----------
export interface FindOptions {
  search: string
  replace: string
  caseSensitive: boolean
  wholeWord: boolean
  useRegex: boolean
  includeSubdir: boolean
  filePattern: string
}
