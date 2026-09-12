<script lang="ts" setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useEditorStore, useFileStore, useSettingStore } from '@/stores'
import { getFileExtension, getTabViewType } from '@/utils'
import {
  OpenFileDialog, ReadFile, SaveFile, SaveFileDialog,
  ReadFileBytes, ConvertToUTF8,
  ConvertFromUTF8, RenameFile, AutoSaveDraft,
  GetRecentFiles, GetRecentFolders,
} from '../../wailsjs/go/main/App'
import { useCommands } from '@/composables/useCommands'
import { useFileOps } from '@/composables/useFileOps'
import { useTailWatcher } from '@/composables/useTailWatcher'
import { confirmDialog, confirmSaveDiscard } from '@/utils/confirm'
import {
  Files, FolderTree, Code2, Bookmark, ListTree, Activity,
} from 'lucide-vue-next'

// ---- Components ----
import NddMenuBar from './NddMenuBar.vue'
import NddToolbar from './NddToolbar.vue'
import EditorArea from './EditorArea.vue'
import StatusBar from './StatusBar.vue'
import FindWin from './FindWin.vue'
import FindResultsPanel from './FindResultsPanel.vue'
import SettingsView from './SettingsView.vue'
import DiffView from './DiffView.vue'
import FileListPanel from './FileListPanel.vue'
import Sidebar from './Sidebar.vue'
import Md5HashWin from './Md5HashWin.vue'
import ColumnEditWin from './ColumnEditWin.vue'
import BatchFindReplaceWin from './BatchFindReplaceWin.vue'
import BatchRenameWin from './BatchRenameWin.vue'
import EncodeConvertWin from './EncodeConvertWin.vue'
import FormatConverter from './FormatConverter.vue'
import FileCmpRuleWin from './FileCmpRuleWin.vue'
import DirCmpView from './DirCmpView.vue'
import ClipboardHistoryWin from './ClipboardHistoryWin.vue'
import RegexTester from './RegexTester.vue'
import SnippetPanel from './SnippetPanel.vue'
import BookmarkPanel from './BookmarkPanel.vue'
import FunctionList from './FunctionList.vue'
import FileMonitor from './FileMonitor.vue'
import ScriptManager from './ScriptManager.vue'
import ColorPicker from './viewer/ColorPicker.vue'
import ModalOverlay from './ModalOverlay.vue'
import RecentFilesDialog from './RecentFilesDialog.vue'

// ---- Stores ----
const ed = useEditorStore()
const fs = useFileStore()
const ss = useSettingStore()

// ---- UI state ----
const showFindWin = ref(false)
const findMode = ref<'find'|'replace'|'files'|'global'|'mark'>('find')
const showFindResults = ref(false)
const findResults = ref<any[]>([])
const findResultsTitle = ref('')
const showSettings = ref(false)
const showDiff = ref(false)
const showFileList = ref(false)
const showFileTree = ref(false)
const showSnippetPanel = ref(false)
const showBookmarkPanel = ref(false)
const showFunctionList = ref(false)
const showFileMonitor = ref(false)
// 侧边栏多面板 Tab 切换
type SidebarTab = 'files' | 'tree' | 'snippets' | 'bookmarks' | 'functions' | 'monitor'
const sidebarTab = ref<SidebarTab>('tree')
function switchSidebarTab(tab: SidebarTab) {
  sidebarTab.value = tab
  // 自动打开对应的面板
  const flagMap: Record<SidebarTab, ReturnType<typeof ref<boolean>>> = {
    files: showFileList,
    tree: showFileTree,
    snippets: showSnippetPanel,
    bookmarks: showBookmarkPanel,
    functions: showFunctionList,
    monitor: showFileMonitor,
  }
  if (!flagMap[tab].value) flagMap[tab].value = true
}
// 同步：菜单命令切换面板时自动切换 Tab
watch(showFileList, (v) => { if (v) sidebarTab.value = 'files' })
watch(showFileTree, (v) => { if (v) sidebarTab.value = 'tree' })
watch(showSnippetPanel, (v) => { if (v) sidebarTab.value = 'snippets' })
watch(showBookmarkPanel, (v) => { if (v) sidebarTab.value = 'bookmarks' })
watch(showFunctionList, (v) => { if (v) sidebarTab.value = 'functions' })
watch(showFileMonitor, (v) => { if (v) sidebarTab.value = 'monitor' })
const panelWidth = ref<number>(250)
const resizing = ref(false)

// 从持久化配置恢复侧栏宽度（与 setSidebarWidth 配对）
if (ss.config?.ui?.fileTreeWidth && typeof ss.config.ui.fileTreeWidth === 'number') {
  panelWidth.value = ss.config.ui.fileTreeWidth
}

// ---- 文件 I/O composable（抽出大量样板）----
const fileOps = useFileOps({
  ed, fs,
  showFileTree,
  setLastFolder: (p) => { if (ss.config) { ss.config.ui.lastFolder = p; ss.saveConfig() } } ,
  onFileSaved: onFileSaved,
  onFileOpened: onFileOpened,
})

// ---- 浮层对话框开关 ----
const showMd5 = ref(false)
const showColumnEdit = ref(false)
const showBatchFind = ref(false)
const showBatchRename = ref(false)
const showEncodeConvert = ref(false)
const showFormatConverter = ref(false)
const showCmpRule = ref(false)
const showDirCmp = ref(false)
const showClipboardHistory = ref(false)

const showRegexTester = ref(false)
const showScriptManager = ref(false)
const showColorPicker = ref(false)
const showRecentDialog = ref(false)
const recentDialogTab = ref<'files' | 'folders'>('files')
// ---- 对比左右文件 ----
const diffLeftPath = ref('')
const diffRightPath = ref('')
// ---- tail-f 状态 ----（已抽离到 useTailWatcher composable）
const tailWatcher = useTailWatcher({ ed, execEd })

// ---- Computed ----
const showToolbar = computed(() => ss.config?.ui?.showToolBar ?? true)

// ============================
//  File operations — 已抽离到 useFileOps composable
//  本组件仅保留依赖 UI 状态/事件的 closeTab 系列（dirty 检查 + saveSession）
// ============================
const saveSession = fileOps.saveSession

async function closeTab(id: string) {
  const t = ed.tabs.find(x => x.id === id)
  if (!t) return
  if (t.isDirty) {
    const choice = await confirmSaveDiscard(t.name)
    if (choice === 'cancel') return
    if (choice === 'save' && t.path) {
      try {
        await SaveFile(t.path, t.content, t.encoding)
      } catch {
        ElMessage.error('保存失败，已取消关闭')
        return
      }
    }
  }
  ed.closeTab(id); saveSession()
}

async function closeAll() {
  if (ed.hasUnsavedChanges) {
    const ok = await confirmDialog({
      title: '关闭所有',
      message: `有 ${ed.dirtyTabs.length} 个文件未保存，确定关闭全部？`,
      confirmText: '全部关闭',
      danger: true,
    })
    if (!ok) return
  }
  ed.closeAllTabs(); saveSession()
}

async function closeOthers() {
  if (!ed.activeTab) return
  const dirtyOthers = ed.tabs.filter(t => t.id !== ed.activeTab!.id && t.isDirty)
  if (dirtyOthers.length > 0) {
    const ok = await confirmDialog({
      title: '关闭其他',
      message: `有 ${dirtyOthers.length} 个文件未保存，是否继续关闭？`,
      confirmText: '关闭其他',
      danger: true,
    })
    if (!ok) return
  }
  ed.closeOtherTabs(ed.activeTab.id); saveSession()
}

async function closeLeft() {
  const idx = ed.activeTabIndex
  if (idx <= 0) return
  const leftTabs = ed.tabs.slice(0, idx)
  const dirtyLeft = leftTabs.filter(t => t.isDirty)
  if (dirtyLeft.length > 0) {
    const ok = await confirmDialog({
      title: '关闭左侧',
      message: `有 ${dirtyLeft.length} 个文件未保存，是否继续关闭？`,
      confirmText: '关闭左侧',
      danger: true,
    })
    if (!ok) return
  }
  for (const t of leftTabs) ed.closeTab(t.id); saveSession()
}

async function closeRight() {
  const idx = ed.activeTabIndex
  if (idx >= ed.tabs.length - 1) return
  const rightTabs = ed.tabs.slice(idx + 1)
  const dirtyRight = rightTabs.filter(t => t.isDirty)
  if (dirtyRight.length > 0) {
    const ok = await confirmDialog({
      title: '关闭右侧',
      message: `有 ${dirtyRight.length} 个文件未保存，是否继续关闭？`,
      confirmText: '关闭右侧',
      danger: true,
    })
    if (!ok) return
  }
  for (const t of rightTabs) ed.closeTab(t.id); saveSession()
}

// ---- 项目文件夹恢复：等配置加载后恢复上次打开的项目目录 ----
watch(() => ss.config?.ui?.lastFolder, (dir) => {
  if (dir && !fs.currentDirectory) {
    fs.setDirectory(dir)
    showFileTree.value = true
  }
}, { immediate: true })

// ---- Auto save ----
let autoSaveTimer: any = null
function startAutoSave() {
  const sec = (ss.config?.editor?.autoSaveInterval || 60) * 1000
  if (autoSaveTimer) clearInterval(autoSaveTimer)
  if (ss.config?.editor?.autoSave) {
    autoSaveTimer = setInterval(async () => {
      for (const t of ed.dirtyTabs) {
        if (t.path) {
          // 自动保存到草稿存储
          try { await AutoSaveDraft(t.path, t.content, t.encoding, t.lineEnding) } catch { /* ignore */ }
        }
      }
    }, sec)
  }
  // 失焦自动保存（读取 settingStore 的 autoSaveMode）
  if (ss.autoSaveMode === 'blur' || ss.autoSaveMode === 'both') {
    window.addEventListener('blur', onWindowBlur)
  }
}
function stopAutoSave() {
  if (autoSaveTimer) { clearInterval(autoSaveTimer); autoSaveTimer = null }
  window.removeEventListener('blur', onWindowBlur)
}

// 失焦自动保存
async function onWindowBlur() {
  for (const t of ed.dirtyTabs) {
    if (t.path) {
      try { await AutoSaveDraft(t.path, t.content, t.encoding, t.lineEnding) } catch { /* ignore */ }
    }
  }
}

// ---- Editor commands ----
function execEd(cmd: string, ...args: any[]) {
  document.dispatchEvent(new CustomEvent('editor-command', { detail: { cmd, args } }))
}

// ---- Macro ----
const isMacroRecording = ref(false)
function toggleMacroRecording() {
  if (ed.macroState.isRecording) {
    ed.stopMacroRecording(); isMacroRecording.value = false; ElMessage.success('宏录制已停止')
  } else {
    ed.startMacroRecording(); isMacroRecording.value = true; ElMessage.success('开始录制宏…')
  }
}
function playCurrentMacro() {
  const list = ed.macroState.savedMacros
  if (!list.length) { ElMessage.warning('暂无可用宏'); return }
  const m = list[list.length - 1]
  ed.playMacro(m.id)
  ElMessage.success('正在播放：' + m.name)
}


// ---- Find ----
function openFind() { findMode.value = 'find'; showFindWin.value = true }
function openReplace() { findMode.value = 'replace'; showFindWin.value = true }
function findInDir() { findMode.value = 'files'; showFindWin.value = true }
function onFindResults(e: Event) {
  const d = (e as CustomEvent).detail
  if (!d) return
  findResultsTitle.value = d.title || '查找结果'
  findResults.value = d.results || []
  showFindResults.value = true
}

// ---- Panel resize ----
let resizeCleanup = () => {}
function resizeStart() {
  resizing.value = true
  document.addEventListener('mousemove', resizeMove)
  document.addEventListener('mouseup', resizeEnd)
  resizeCleanup = () => {
    document.removeEventListener('mousemove', resizeMove)
    document.removeEventListener('mouseup', resizeEnd)
  }
}
function resizeMove(e: MouseEvent) {
  if (!resizing.value) return
  panelWidth.value = Math.max(150, Math.min(500, e.clientX))
}
function resizeEnd() {
  if (!resizing.value) return
  resizing.value = false
  // 拖拽结束后落盘（store 内部已 300ms debounce，再触发一次即可）
  ss.setSidebarWidth(panelWidth.value)
  resizeCleanup()
  resizeCleanup = () => {}
}

// 拖拽增强：支持文件/文件夹/文本拖放
function onDrop(e: DragEvent) {
  e.preventDefault()
  const fs = e.dataTransfer?.files
  if (fs && fs.length > 0) {
    for (let i = 0; i < fs.length; i++) {
      const f = fs[i] as any
      if (f.path) {
        fileOps.openFilePath(f.path)
      }
    }
    return
  }
  // 外部拖拽文本到编辑器
  const text = e.dataTransfer?.getData('text/plain')
  if (text && ed.activeTab) {
    document.dispatchEvent(new CustomEvent('editor-command', {
      detail: { cmd: 'insert-text', args: [text] }
    }))
    ElMessage.success('已插入文本')
  }
}
function onDrag(e: DragEvent) { e.preventDefault() }

// ---- Keyboard shortcuts ----
// CodeMirror 拦截的键通过 ndd-key 事件传过来
// document keydown 处理未被 CodeMirror 拦截的键（覆盖 LogViewer/HexViewer 等非 CodeEditor 视图，
// 修复"打开日志文件按 Ctrl+F 无效"的 Bug）
function onNddKey(e: Event) {
  const cmd = (e as CustomEvent).detail as string
  onMenuCmd(cmd)
}
function onKeyDown(e: KeyboardEvent) {
  const c = e.ctrlKey || e.metaKey
  // 全局快捷键改为读取「设置 → 快捷键管理」里的键位表（ss.matchShortcut），
  // 默认值与原先的硬编码完全一致，因此行为不变；用户在设置里改键后即时生效。
  const hit = (id: string) => ss.matchShortcut(e, id)

  // 文件操作
  if (hit('open-file')) { e.preventDefault(); fileOps.openFile() }
  else if (hit('save-as')) { e.preventDefault(); fileOps.saveFileAs() }
  else if (hit('save-all')) { e.preventDefault(); fileOps.saveAll() }
  else if (hit('save')) { e.preventDefault(); fileOps.saveFile() }
  else if (hit('new-file')) { e.preventDefault(); fileOps.newFile() }
  else if (hit('close-tab')) { e.preventDefault(); if (ed.activeTab) closeTab(ed.activeTab.id) }
  else if (hit('close-all')) { e.preventDefault(); closeAll() }
  else if (hit('exit')) { e.preventDefault(); onMenuCmd('exit') }
  else if (hit('print')) { e.preventDefault(); onMenuCmd('print') }
  // 查找 / 替换 / 跳转全局快捷键（LogViewer/HexViewer 等不挂在 CodeMirror keymap，必须在这里兜底）
  // 注：键位表中 find-in-dir 的默认键是 Ctrl+Shift+F，沿用原来的「全局搜索」行为
  else if (hit('find-in-dir')) { e.preventDefault(); onMenuCmd('search-files') }
  else if (hit('replace')) { e.preventDefault(); onMenuCmd('replace') }
  else if (hit('find')) { e.preventDefault(); onMenuCmd('find') }
  else if (hit('find-next')) { e.preventDefault(); onMenuCmd('find-next') }
  else if (hit('find-prev')) { e.preventDefault(); onMenuCmd('find-prev') }
  else if (hit('goto-line')) { e.preventDefault(); onMenuCmd('goto-line') }
  // —— 以下组合尚未进入键位表，保留硬编码兜底 ——
  else if (c && e.shiftKey && (e.key === 'O' || e.key === 'o')) { e.preventDefault(); fileOps.openDir() }
  else if (c && e.shiftKey && (e.key === 'D' || e.key === 'd')) { e.preventDefault(); onMenuCmd('find-dir') }
  else if (c && !e.shiftKey && (e.key === 't' || e.key === 'T')) { e.preventDefault(); fileOps.newFile() }
  // Tab 切换
  else if (c && e.key === 'Tab') { e.preventDefault(); const tabs = ed.tabs; if (!tabs.length) return; let i = ed.activeTabIndex + (e.shiftKey ? -1 : 1); if (i < 0) i = tabs.length - 1; else if (i >= tabs.length) i = 0; ed.activateTab(tabs[i].id) }
  else if (e.key === 'Escape') { if (showFindWin.value) showFindWin.value = false }
}

// ============================
//  Menu command dispatcher (extracted to useCommands composable)
// ============================
// useCommands 在文件末尾调用，确保所有 refs 和 helper 函数已定义

async function gotoFindResult(r: any) {
  const t = ed.getTabByPath(r.file)
  if (t) { ed.activateTab(t.id); setTimeout(() => execEd('scroll-to-line', r.line), 150) }
}

// 辅助函数
function toggleAutoTheme() {
  if (ss.config) {
    ss.config.theme.autoTheme = !ss.config.theme.autoTheme
    ss.saveConfig()
    if (ss.config.theme.autoTheme) {
      ss.enableAutoTheme()
      ElMessage.success('已启用自动跟随系统主题')
    } else {
      ss.disableAutoTheme()
      ElMessage.success('已切换为手动主题模式')
    }
  }
}

// 统一最近访问入口：通过下拉对话框展示，替代旧版文本弹窗
async function showRecentFiles() { recentDialogTab.value = 'files'; showRecentDialog.value = true }
async function showRecentFolders() { recentDialogTab.value = 'folders'; showRecentDialog.value = true }

// 打开最近文件夹：必须同时展开目录树面板，否则用户看不到任何变化
function openRecentFolder(path: string) {
  fs.setDirectory(path)
  showFileTree.value = true
  sidebarTab.value = 'tree'
  if (ss.config) { ss.config.ui.lastFolder = path; ss.saveConfig() }
}

async function manageDrafts() {
  try {
    const { ListDrafts } = await import('../../wailsjs/go/main/App')
    const drafts = await ListDrafts()
    if (!drafts || drafts.length === 0) {
      ElMessage.info('暂无草稿')
      return
    }
    const list = (drafts as any[]).map((d: any) => `${d.filePath.split(/[/\\]/).pop()} (${d.savedAt})`).join('\n')
    ElMessageBox.alert(`共 ${drafts.length} 个草稿:\n\n${list}`, '草稿管理', {
      confirmButtonText: '清除所有草稿',
      cancelButtonText: '关闭',
      showCancelButton: true,
    }).then(async () => {
      const { ClearAllDrafts } = await import('../../wailsjs/go/main/App')
      await ClearAllDrafts()
      ElMessage.success('所有草稿已清除')
    }).catch(() => {})
  } catch (e: any) {
    ElMessage.error(`草稿管理失败: ${e?.message || ''}`)
  }
}

// 第四阶段：新增功能入口
function showScriptManagerDialog() {
  showScriptManager.value = true
}

function showImageEditorView() {
  const t = ed.activeTab
  if (!t) { ElMessage.warning('请先打开一个图片文件'); return }
  const ext = getFileExtension(t.path || t.name)
  if (!['png', 'jpg', 'jpeg', 'gif', 'webp', 'bmp', 'svg', 'ico'].includes(ext.toLowerCase())) {
    ElMessage.warning('当前文件不是图片，请先打开一个图片文件')
    return
  }
  // 切换 viewType 为 image-edit，让 EditorArea 渲染 ImageEditor
  t.viewType = 'image-edit'
  ElMessage.success('已切换到图片编辑模式')
}

function showColorPickerPanel() {
  showColorPicker.value = true
}

function toggleLogMode() {
  const t = ed.activeTab
  if (!t) { ElMessage.warning('请先打开一个文件'); return }
  if (t.viewType === 'log') {
    t.viewType = 'code'
    ElMessage.success('已切换为普通模式')
  } else {
    t.viewType = 'log'
    ElMessage.success('已切换为日志查看模式')
  }
}

async function openExplorer() {
  const t = ed.activeTab
  if (!t?.path) return
  try {
    const { BrowserOpenURL } = await import('../../wailsjs/runtime/runtime')
    BrowserOpenURL('file:///' + t.path.replace(/\\/g, '/').replace(/\/[^\/]+$/, ''))
  } catch (e) { console.warn(e) }
}

// ============================
//  编码：以某编码重新解码打开 / 转换为某编码
// ============================
async function reopenWithEncoding(enc: string) {
  const t = ed.activeTab
  if (!t?.path) { ElMessage.warning('当前文档无文件路径'); return }
  try {
    const bytes = await ReadFileBytes(t.path)
    const content = await ConvertToUTF8(bytes as unknown as number[], enc)
    ed.updateTabContent(t.id, content)
    ed.updateTabEncoding(t.id, enc)
    ElMessage.success(`已以 ${enc} 编码打开`)
  } catch (e: any) { ElMessage.error(`以 ${enc} 打开失败：${e?.message || ''}`) }
}
function convertTo(enc: string) {
  const t = ed.activeTab
  if (!t) return
  ed.updateTabEncoding(t.id, enc)
  t.isDirty = true
  ElMessage.success(`将转换为 ${enc}，保存时生效`)
}

// ---- 以文本/二进制模式重新打开 ----
async function reopenAs(vt: 'code' | 'hex') {
  const t = ed.activeTab
  if (!t?.path) { ElMessage.warning('当前文档无文件路径'); return }
  try {
    if (vt === 'code') {
      const r = await ReadFile(t.path)
      if (r) {
        ed.updateTabContent(t.id, r.content)
        // 同步 originalContent，避免关闭时错误提示保存
        t.originalContent = r.content
        t.viewType = 'code'
      }
    } else { t.viewType = 'hex' }
  } catch (e: any) { ElMessage.error('重新打开失败：' + (e?.message || '')) }
}

// ---- 重命名当前文件 ----
async function renameCurrentFile() {
  const t = ed.activeTab
  if (!t?.path) { ElMessage.warning('当前文档无文件路径'); return }
  try {
    const { value: newName } = await ElMessageBox.prompt('请输入新文件名', '重命名', {
      inputValue: t.name, confirmButtonText: '确定', cancelButtonText: '取消',
    })
    if (!newName || newName === t.name) return
    const dir = t.path.replace(/[\\/][^\\/]+$/, '')
    const sep = t.path.includes('\\') ? '\\' : '/'
    const newPath = dir + sep + newName
    await RenameFile(t.path, newPath)
    ed.renameTab(t.id, newPath); saveSession()
    ElMessage.success('已重命名')
  } catch (e) { console.warn(e) }
}

// ---- 对比：选择左右文件 / 二进制对比 ----
async function selectCompareFile(side: 'left' | 'right') {
  try {
    const p = await OpenFileDialog()
    if (p) { if (side === 'left') diffLeftPath.value = p; else diffRightPath.value = p; showDiff.value = true; ElMessage.success(`已选择${side === 'left' ? '左侧' : '右侧'}文件`) }
  } catch (e) { console.warn(e) }
}

// 标签栏右键「选择左侧/右侧对比文件」：直接设置路径并打开对比视图。
function onTabCompareFile(side: 'left' | 'right') {
  return (e: Event) => {
    const path = (e as CustomEvent).detail as string
    if (!path) return
    if (side === 'left') diffLeftPath.value = path
    else diffRightPath.value = path
    showDiff.value = true
  }
}
const onSelectCmpLeft = onTabCompareFile('left')
const onSelectCmpRight = onTabCompareFile('right')

/** 对比规则已应用：通知 DiffView 用新规则重新对比 */
function onCmpRulesApplied() {
  document.dispatchEvent(new CustomEvent('cmp-rules-updated'))
}
async function binaryCompare() {
  try {
    const left = await OpenFileDialog(); if (!left) return
    const right = await OpenFileDialog(); if (!right) return
    const { BinaryCompare } = await import('../../wailsjs/go/main/App')
    const r = await BinaryCompare(left, right)
    if (r?.equal) { ElMessage.success('两文件完全相同'); return }
    ElMessageBox.alert(
      `两文件不同。\n首个差异偏移：${r?.firstDiffOffset ?? -1}\n原因：${r?.reason || ''}\n\n${r?.hexWindow || ''}`,
      '二进制对比结果', { confirmButtonText: '确定' },
    )
  } catch (e: any) { ElMessage.error('二进制对比失败：' + (e?.message || '')) }
}

// ---- 宏：保存 / 多次运行 ----
// 宏步骤在前端录制（editorStore.currentMacro），持久化走 localStorage
// （saveMacros/loadMacros），不经后端。
async function saveCurrentMacro() {
  try {
    const { value: name } = await ElMessageBox.prompt('请输入宏名称', '保存宏', {
      inputValue: `Macro ${ed.macroState.savedMacros.length + 1}`, confirmButtonText: '保存', cancelButtonText: '取消',
    })
    if (!name) return
    const last = ed.macroState.savedMacros[ed.macroState.savedMacros.length - 1]
    if (!last) { ElMessage.warning('没有可保存的宏，请先录制'); return }
    ed.renameMacro(last.id, name)
    ed.saveMacros()
    ElMessage.success('宏已保存：' + name)
  } catch (e) { console.warn(e) }
}
async function runMacroMulti() {
  const list = ed.macroState.savedMacros
  if (!list.length) { ElMessage.warning('无可用宏'); return }
  try {
    const { value } = await ElMessageBox.prompt('运行次数', '多次运行宏', { inputValue: '3', confirmButtonText: '运行', cancelButtonText: '取消' })
    const n = parseInt(value)
    if (isNaN(n) || n < 1) return
    ed.playMacro(list[list.length - 1].id, n)
    ElMessage.success(`运行宏 ×${n}`)
  } catch (e) { console.warn(e) }
}

// ---- 收藏夹 ----
async function manageFavorites() {
  const t = ed.activeTab
  if (!t?.path) { ElMessage.warning('请先打开一个文件再加入收藏夹'); return }
  if (!ss.config) return
  const favs = ss.config.ui.favorites || []
  if (favs.includes(t.path)) { ElMessage.info('已在收藏夹中'); return }
  favs.push(t.path); ss.config.ui.favorites = favs
  await ss.saveConfig()
  ElMessage.success('已加入收藏夹')
}

// ---- 主题导入/导出 ----
async function exportTheme() {
  if (!ss.config) return
  try {
    const p = await SaveFileDialog('theme.json')
    if (!p) return
    const content = JSON.stringify({ theme: ss.config.theme, colors: ss.currentThemeColors }, null, 2)
    await SaveFile(p, content, 'UTF-8')
    ElMessage.success('主题已导出')
  } catch (e: any) { ElMessage.error('导出失败：' + (e?.message || '')) }
}
async function importTheme() {
  try {
    const p = await OpenFileDialog()
    if (!p) return
    const r = await ReadFile(p)
    if (!r) return
    const data = JSON.parse(r.content)
    if (data?.theme?.currentTheme) {
      ss.config!.theme.currentTheme = data.theme.currentTheme
      ss.applyTheme()
      ElMessage.success('主题已导入：' + data.theme.currentTheme)
    } else { ElMessage.warning('无效的主题文件') }
  } catch (e: any) { ElMessage.error('导入失败：' + (e?.message || '')) }
}

// ---- tail -f 文件跟踪 ----（已抽离到 useTailWatcher composable）
const { tailingTabId, isTailing: tailingStatus, startTail, stopTail } = tailWatcher

// ---- 文件变化检测（窗口聚焦时） ----
let lastFocusCheck = 0
// 重入保护：检测过程中会 await 确认对话框，期间可能再次触发 focus 事件。
// 没有这个标志，用户点完弹窗窗口重获焦点 → 再次检测 → 弹窗死循环。
let checkingExternal = false
// 用户已选择「忽略」的外部内容快照：path -> 当时的磁盘内容。
// 只有磁盘内容再次发生变化才重新询问，避免同一变更反复打扰。
const dismissedExternal = new Map<string, string>()

/** 把磁盘内容写回标签页，并重置脏标记（重新加载后不应视为未保存）。 */
function applyExternalContent(tabId: string, diskContent: string) {
  ed.updateTabContent(tabId, diskContent)
  const tab = ed.tabs.find(x => x.id === tabId)
  if (tab) { tab.originalContent = diskContent; tab.isDirty = false }
}

async function checkExternalChanges() {
  const now = Date.now()
  // 节流：两次检测间隔至少 3 秒
  if (now - lastFocusCheck < 3000) return
  lastFocusCheck = now
  if (checkingExternal) return
  checkingExternal = true
  try {
    // 复制一份快照遍历：await 期间 tabs 可能被用户改动
    for (const t of [...ed.tabs]) {
      if (!t.path) continue
      try {
        const r = await ReadFile(t.path)
        if (!r) continue
        const disk = r.content
        // 磁盘内容与基线一致 → 没有外部变更，清除忽略记录
        if (disk === (t.originalContent ?? '')) {
          dismissedExternal.delete(t.path)
          continue
        }
        // 该版本用户已明确忽略，且内容没有再变 → 不再打扰
        if (dismissedExternal.get(t.path) === disk) continue

        if (t.isDirty) {
          const ok = await confirmDialog({
            title: '文件冲突',
            message: `"${t.name}" 在外部被修改，同时你有未保存的更改。\n重新加载会丢失本地更改，是否继续？`,
            confirmText: '重新加载（丢弃本地更改）',
            cancelText: '保留我的更改',
            danger: true,
          })
          if (ok) {
            applyExternalContent(t.id, disk)
            dismissedExternal.delete(t.path)
            ElMessage.success(`已重新加载: ${t.name}`)
          } else {
            dismissedExternal.set(t.path, disk)
          }
        } else {
          const ok = await confirmDialog({
            title: '文件已变更',
            message: `"${t.name}" 已被外部修改，是否重新加载？`,
            confirmText: '重新加载',
            cancelText: '忽略',
          })
          if (ok) {
            applyExternalContent(t.id, disk)
            dismissedExternal.delete(t.path)
            ElMessage.success(`已重新加载: ${t.name}`)
          } else {
            dismissedExternal.set(t.path, disk)
          }
        }
      } catch { /* 文件可能已被删除或无法读取，忽略 */ }
    }
  } finally {
    checkingExternal = false
  }
}

// 当文件被保存时，清除忽略状态（本地内容已成为磁盘最新内容）
function onFileSaved(filePath: string) {
  dismissedExternal.delete(filePath)
}

// 当文件被重新打开时，清除忽略状态
function onFileOpened(filePath: string) {
  dismissedExternal.delete(filePath)
}

function onWindowFocus() { checkExternalChanges() }
function onVisibilityChange() {
  if (!document.hidden) checkExternalChanges()
}

// 监听后端 file:change 事件，直接触发 checkExternalChanges，
// 不再仅依赖窗口焦点 / 可见性轮询。StartFileWatch 由 watchActiveTab() 在
// activeTab.path 变化时启动。
let cancelFileChangeListener: (() => void) | null = null
let watchedPath: string | null = null
async function watchActiveTab() {
  const { StartFileWatch, StopFileWatch } = await import('../../wailsjs/go/main/App')
  const path = ed.activeTab?.path
  if (path === watchedPath) return
  if (watchedPath) {
    try { await StopFileWatch(watchedPath) } catch { /* ignore */ }
  }
  watchedPath = path || null
  if (path) {
    try { await StartFileWatch(path) } catch { /* ignore */ }
  }
}

// 监听 file:change：后端 fsnotify 命中时直接调 checkExternalChanges
async function onFileChange() {
  // 节流由 checkExternalChanges 内部的 3s 间隔保障
  checkExternalChanges()
}

// ---- useCommands: 命令分发（在 helper 函数全部定义后调用） ----
const { onMenuCmd } = useCommands({
  ed, fs, ss,
  // 浮层开关
  showFindWin, showDiff, showSettings, showFileList, showFileTree,
  showMd5, showColumnEdit, showBatchFind, showBatchRename, showEncodeConvert,
  showFormatConverter, showCmpRule, showDirCmp, showClipboardHistory,
  showRegexTester, showSnippetPanel, showBookmarkPanel, showFunctionList, showFileMonitor,
  showScriptManager, showColorPicker,
  showFindResults, findMode, findResults, findResultsTitle,
  diffLeftPath, diffRightPath,
  // 函数引用
  newFile: fileOps.newFile,
  openFile: fileOps.openFile,
  openFilePath: fileOps.openFilePath,
  openDir: fileOps.openDir,
  saveFile: fileOps.saveFile,
  saveFileAs: fileOps.saveFileAs,
  saveAll: fileOps.saveAll,
  closeTab, closeAll, closeOthers, closeLeft, closeRight,
  reloadFile: fileOps.reloadFile,
  saveSession,
  saveWorkspace: fileOps.saveWorkspace,
  openWorkspace: fileOps.openWorkspace,
  restoreSession: fileOps.restoreSession,
  toggleMacroRecording, playCurrentMacro, saveCurrentMacro, runMacroMulti,
  manageFavorites, importTheme, exportTheme, renameCurrentFile,
  selectCompareFile, binaryCompare,
  reopenAs, reopenWithEncoding, convertTo, execEd,
  startTail, stopTail, manageDrafts, toggleAutoTheme,
  showRecentFiles, showRecentFolders,
  showImageEditorView, toggleLogMode, isMacroRecording,
})

// ---- 监听来自其他实例的文件打开请求 ----
let cancelOpenFileListener: (() => void) | null = null

async function setupOpenFileListener() {
  const { EventsOn } = await import('../../wailsjs/runtime/runtime')
  cancelOpenFileListener = EventsOn('app:open-file', async (path: string) => {
    if (path) {
      await fileOps.openFilePath(path)
    }
  })
}

function cleanupOpenFileListener() {
  if (cancelOpenFileListener) {
    cancelOpenFileListener()
    cancelOpenFileListener = null
  }
}

// ---- 关闭窗口拦截：后端关闭前询问前端，避免未保存内容被静默丢弃 ----
let cancelBeforeClose: (() => void) | null = null

async function setupBeforeCloseListener() {
  const { EventsOn, EventsEmit } = await import('../../wailsjs/runtime/runtime')
  cancelBeforeClose = EventsOn('app:before-close', async () => {
    // 无未保存内容 → 直接放行退出
    if (ed.dirtyTabs.length === 0) {
      EventsEmit('app:quit-force')
      return
    }
    try {
      await ElMessageBox.confirm(
        `有 ${ed.dirtyTabs.length} 个文件存在未保存的修改。\n` +
        `点击「保存并退出」会先保存全部更改再关闭；点击「取消」返回继续编辑。`,
        '退出确认',
        { confirmButtonText: '保存并退出', cancelButtonText: '取消', type: 'warning' },
      )
    } catch {
      return // 用户取消：窗口保持打开
    }
    await fileOps.saveAll()
    // 新建文档还没有路径，saveAll 会跳过；逐个弹出另存为，保证不丢内容
    for (const t of [...ed.dirtyTabs]) {
      if (!t.path) {
        ed.activateTab(t.id)
        await fileOps.saveFileAs()
      }
    }
    if (ed.dirtyTabs.length > 0) {
      ElMessage.warning('仍有文档未保存，已取消退出')
      return
    }
    EventsEmit('app:quit-force')
  })
}

function cleanupBeforeCloseListener() {
  if (cancelBeforeClose) {
    cancelBeforeClose()
    cancelBeforeClose = null
  }
}

// ---- Lifecycle ----
// 具名函数：add/remove 必须传同一个引用，匿名箭头函数会导致监听器永远无法移除
function onShowColumnEdit() { showColumnEdit.value = true }

onMounted(() => {
  window.addEventListener('keydown', onKeyDown)
  document.addEventListener('ndd-key', onNddKey)
  document.addEventListener('drop', onDrop)
  document.addEventListener('dragover', onDrag)
  document.addEventListener('find-results-update', onFindResults)
  document.addEventListener('show-column-edit', onShowColumnEdit)
  document.addEventListener('select-left-cmp-file', onSelectCmpLeft)
  document.addEventListener('select-right-cmp-file', onSelectCmpRight)
  document.addEventListener('visibilitychange', onVisibilityChange)
  window.addEventListener('focus', onWindowFocus)
  setupOpenFileListener()
  setupBeforeCloseListener()
  // 监听后端 file:change 事件，不再只依赖 focus/visibility 轮询
  ;(async () => {
    const { EventsOn } = await import('../../wailsjs/runtime/runtime')
    cancelFileChangeListener = EventsOn('file:change', onFileChange)
  })()
  watchActiveTab()
  ed.loadMacros() // 恢复上次保存的宏
  // 接线设置里的「关闭时恢复文件」开关
  if (ss.config?.ui?.restoreSession !== false) fileOps.restoreSession()
  startAutoSave()
})
onUnmounted(() => {
  window.removeEventListener('keydown', onKeyDown)
  document.removeEventListener('ndd-key', onNddKey)
  document.removeEventListener('drop', onDrop)
  document.removeEventListener('dragover', onDrag)
  document.removeEventListener('find-results-update', onFindResults)
  document.removeEventListener('show-column-edit', onShowColumnEdit)
  document.removeEventListener('select-left-cmp-file', onSelectCmpLeft)
  document.removeEventListener('select-right-cmp-file', onSelectCmpRight)
  document.removeEventListener('visibilitychange', onVisibilityChange)
  window.removeEventListener('focus', onWindowFocus)
  cleanupOpenFileListener()
  cleanupBeforeCloseListener()
  if (cancelFileChangeListener) { cancelFileChangeListener(); cancelFileChangeListener = null }
  stopAutoSave()
})

// 监听活动 tab 路径变化 → 重新启 watch（bug #6 修复）
watch(() => ed.activeTab?.path, () => { watchActiveTab() })
</script>

<template>
  <div class="h-full flex flex-col bg-[var(--et-bg)]">

    <!-- 菜单栏 -->
    <NddMenuBar @cmd="onMenuCmd" />

    <!-- 工具栏 -->
    <NddToolbar v-if="showToolbar" :icon-size="ss.config?.ui?.toolbarIconSize || 18" :tailing="tailingStatus" @toolbar-command="onMenuCmd" />

    <!-- 主体：侧边栏 + 编辑区 -->
    <div class="flex-1 flex overflow-hidden min-h-0">

      <!-- 侧边栏（Tab 切换：文件列表 / 目录树 / 代码片段 / 书签） -->
      <div
        v-if="showFileList || showFileTree || showSnippetPanel || showBookmarkPanel || showFunctionList || showFileMonitor"
        class="sidebar-panel flex flex-shrink-0"
        :style="{ width: panelWidth + 'px' }"
      >
        <!-- 左侧垂直 Tab 图标栏（Lucide 替代原手撸 SVG） -->
        <div class="sidebar-tab-bar">
          <button
            title="文件列表"
            class="sidebar-tab-btn"
            :class="{ active: sidebarTab === 'files' }"
            @click="switchSidebarTab('files')"
          >
            <Files :size="16" :stroke-width="1.6" />
          </button>
          <button
            title="目录树"
            class="sidebar-tab-btn"
            :class="{ active: sidebarTab === 'tree' }"
            @click="switchSidebarTab('tree')"
          >
            <FolderTree :size="16" :stroke-width="1.6" />
          </button>
          <button
            title="代码片段"
            class="sidebar-tab-btn"
            :class="{ active: sidebarTab === 'snippets' }"
            @click="switchSidebarTab('snippets')"
          >
            <Code2 :size="16" :stroke-width="1.6" />
          </button>
          <button
            title="书签"
            class="sidebar-tab-btn"
            :class="{ active: sidebarTab === 'bookmarks' }"
            @click="switchSidebarTab('bookmarks')"
          >
            <Bookmark :size="16" :stroke-width="1.6" />
          </button>
          <button
            title="函数列表"
            class="sidebar-tab-btn"
            :class="{ active: sidebarTab === 'functions' }"
            @click="switchSidebarTab('functions')"
          >
            <ListTree :size="16" :stroke-width="1.6" />
          </button>
          <button
            title="文件监控"
            class="sidebar-tab-btn"
            :class="{ active: sidebarTab === 'monitor' }"
            @click="switchSidebarTab('monitor')"
          >
            <Activity :size="16" :stroke-width="1.6" />
          </button>
        </div>
        <!-- 面板内容区域 -->
        <div class="flex-1 min-w-0 overflow-hidden">
          <FileListPanel v-if="showFileList && sidebarTab === 'files'"
            @select-tab="id => ed.activateTab(id)" @close-tab="closeTab" />
          <Sidebar v-if="showFileTree && sidebarTab === 'tree'" class="h-full" />
          <SnippetPanel v-if="showSnippetPanel && sidebarTab === 'snippets'" class="h-full" />
          <BookmarkPanel v-if="showBookmarkPanel && sidebarTab === 'bookmarks'" class="h-full" />
          <FunctionList v-if="showFunctionList && sidebarTab === 'functions'" class="h-full"
            :content="ed.activeTab?.content || ''" :language="ed.activeTab?.language || ''"
            @goto="(line: number) => execEd('scroll-to-line', line)" />
          <FileMonitor v-if="showFileMonitor && sidebarTab === 'monitor'" class="h-full" />
        </div>
      </div>

      <!-- 拖拽分隔条 -->
      <div v-if="showFileList || showFileTree || showSnippetPanel || showBookmarkPanel || showFunctionList || showFileMonitor" class="panel-resizer" @mousedown="resizeStart" />

      <!-- 编辑区 -->
      <EditorArea class="flex-1 min-w-0" @open-diff="showDiff = true" @open-converter="showFormatConverter = true" />
    </div>

    <!-- 底部查找结果面板 -->
    <FindResultsPanel
      :visible="showFindResults" :results="findResults" :title="findResultsTitle"
      @close="showFindResults = false" @clear="() => { findResults = []; showFindResults = false }"
      @goto="gotoFindResult" />

    <!-- 状态栏 -->
    <StatusBar v-if="ss.showStatusBar" />

    <!-- 浮层对话框 -->
    <FindWin :visible="showFindWin" :mode="findMode" @close="showFindWin = false" />
    <SettingsView :visible="showSettings" @close="showSettings = false" />
    <DiffView :visible="showDiff" :left-path="diffLeftPath" :right-path="diffRightPath" @close="showDiff = false" />
    <Md5HashWin :visible="showMd5" @close="showMd5 = false" />
    <ColumnEditWin :visible="showColumnEdit" @close="showColumnEdit = false" />
    <BatchFindReplaceWin :visible="showBatchFind" @close="showBatchFind = false" />
    <BatchRenameWin :visible="showBatchRename" @close="showBatchRename = false" />
    <EncodeConvertWin :visible="showEncodeConvert" @close="showEncodeConvert = false" />
    <FormatConverter :visible="showFormatConverter" @close="showFormatConverter = false" />
    <FileCmpRuleWin :visible="showCmpRule" @close="showCmpRule = false" @apply="onCmpRulesApplied" />
    <DirCmpView :visible="showDirCmp" @close="showDirCmp = false" />
    <ClipboardHistoryWin :visible="showClipboardHistory" @close="showClipboardHistory = false" />
    <!-- 新组件（浮层） -->
    <RegexTester :visible="showRegexTester" @close="showRegexTester = false" />
    <!-- 脚本管理器弹窗 -->
    <ModalOverlay :visible="showScriptManager" title="脚本管理器" width="85vw" height="80vh" @close="showScriptManager = false">
      <ScriptManager />
    </ModalOverlay>

    <!-- 取色器弹窗 -->
    <ModalOverlay :visible="showColorPicker" title="颜色选择器" width="400px" @close="showColorPicker = false">
      <ColorPicker @close="showColorPicker = false" @select="(c: string) => { /* 颜色已复制 */ }" />
    </ModalOverlay>

    <!-- 最近访问文件/文件夹对话框 -->
    <ModalOverlay :visible="showRecentDialog" title="最近访问" width="520px" height="420px" @close="showRecentDialog = false">
      <RecentFilesDialog
        :visible="showRecentDialog"
        :initial-tab="recentDialogTab"
        @close="showRecentDialog = false"
        @open-file="(path: string) => { showRecentDialog = false; fileOps.openFilePath(path) }"
        @open-folder="openRecentFolder"
      />
    </ModalOverlay>
  </div>
</template>

<style scoped>
.sidebar-panel {
  border-right: 1px solid var(--et-border);
  background: var(--et-bg);
}

.sidebar-tab-bar {
  flex-shrink: 0;
  width: 40px;
  border-right: 1px solid var(--et-border);
  background: var(--et-bg-sunken);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--et-space-2) 0;
  gap: var(--et-space-1);
}

/* 侧边栏 Tab 切换按钮 */
.sidebar-tab-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: var(--et-radius);
  background: transparent;
  color: var(--et-fg-subtle);
  cursor: pointer;
  transition: background-color 80ms ease, color 80ms ease;
  position: relative;
}
.sidebar-tab-btn:hover {
  background: var(--et-bg-hover);
  color: var(--et-fg);
}
.sidebar-tab-btn.active {
  background: var(--et-accent-soft);
  color: var(--et-accent);
}
/* active 状态左侧加 2px 主色条，与 TabBar 一致 */
.sidebar-tab-btn.active::before {
  content: '';
  position: absolute;
  left: 0; top: 50%;
  transform: translateY(-50%);
  width: 2px;
  height: 16px;
  background: var(--et-accent);
  border-radius: 0 2px 2px 0;
}

.panel-resizer {
  width: 3px;
  cursor: col-resize;
  background: transparent;
  flex-shrink: 0;
  transition: background-color 80ms ease;
}
.panel-resizer:hover {
  background: var(--et-accent);
}
</style>
