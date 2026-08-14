<script lang="ts" setup>
import { ref, reactive, computed, watch } from 'vue'
import { useSettingStore, useEditorStore } from '@/stores'
import { UpdateConfig, ImportSnippets, ExportSnippets, DeleteSnippet, GetSnippets, OpenFileDialog, SaveFileDialog, SaveFile, ReadFile } from '../../wailsjs/go/main/App'
import { RegisterFileAssoc, UnregisterFileAssoc, IsFileAssocRegistered } from '../../wailsjs/go/main/App'
import { ElMessage } from 'element-plus'
import type { ThemeName } from '@/types'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits(['close'])

const settingStore = useSettingStore()
const editorStore = useEditorStore()
const activePage = ref('editor')

const config = computed(() => settingStore.config)

const form = reactive({
  fontSize: config.value.editor.fontSize,
  fontFamily: config.value.editor.fontFamily,
  tabSize: config.value.editor.tabSize,
  insertSpaces: config.value.editor.insertSpaces,
  wordWrap: config.value.editor.wordWrap,
  lineNumbers: config.value.editor.lineNumbers,
  autoSave: config.value.editor.autoSave,
  autoSaveInterval: config.value.editor.autoSaveInterval,
  highlightLine: config.value.editor.highlightLine,
  bracketPairColor: config.value.editor.bracketPairColor,
  showIndentGuide: config.value.editor.showIndentGuide,
  showWhitespace: config.value.editor.showWhitespace,
  showEol: config.value.editor.showEol,
  bigTextSizeLimit: 100,
  currentTheme: config.value.theme.currentTheme,
  defaultEncoding: config.value.file.defaultEncoding,
  autoDetectEncoding: config.value.file.autoDetectEncoding,
  defaultLineEnding: config.value.file.defaultLineEnding,
  restoreFilesOnClose: true,
  clearOpenFileRecord: false,
  language: config.value.ui.language,
  showFileTree: config.value.ui.showFileTree,
  showStatusBar: config.value.ui.showStatusBar,
  showToolBar: config.value.ui.showToolBar,
  showFileListView: config.value.ui.showFileListView,
  zoomLevel: config.value.ui.zoomLevel,
  autoIndent: false,
  // 🆕 V2.0.0
  autoTheme: config.value.theme.autoTheme ?? false,
  autoSaveMode: settingStore.autoSaveMode || 'interval',
  ignorePatterns: (config.value.file.ignorePatterns || ['.git', '~$*', '*.tmp', 'node_modules']).join(', '),
  // 🆕 关闭行为：默认开托盘驻留
  closeToTray: config.value.ui.closeToTray ?? true,
  // 🆕 最近文件滚动保留条数（默认 10）
  recentFilesLimit: config.value.ui.recentFilesLimit || 10,
})

const pages = [
  { key: 'editor', label: '编辑器 / 字体' },
  { key: 'theme', label: '主题样式' },
  { key: 'file', label: '文件' },
  { key: 'ui', label: '界面' },
  { key: 'shortcut', label: '快捷键管理' },
  { key: 'macro', label: '宏管理' },
  { key: 'snippets', label: '代码片段' },
]

const fontOptions = [
  { label: 'Consolas', value: 'Consolas, Monaco, "Courier New", monospace' },
  { label: 'Fira Code', value: '"Fira Code", Consolas, monospace' },
  { label: 'JetBrains Mono', value: '"JetBrains Mono", Consolas, monospace' },
  { label: 'Source Code Pro', value: '"Source Code Pro", Consolas, monospace' },
  { label: 'Monaco', value: 'Monaco, Consolas, monospace' },
]

const encodingOptions = ['UTF-8', 'UTF-16LE', 'UTF-16BE', 'GB18030', 'GBK', 'Big5', 'ISO-8859-1', 'Windows-1252']
const lineEndingOptions = [
  { value: 'CRLF', label: 'CRLF (Windows)' },
  { value: 'LF', label: 'LF (Unix)' },
  { value: 'CR', label: 'CR (Mac)' },
]
const languageOptions = [
  { value: 'zh-CN', label: '简体中文' },
  { value: 'en-US', label: 'English' },
]

// 19 套主题 - 来自 settingStore
const fileAssocRegistered = ref(false)
async function refreshFileAssocState() {
  try { fileAssocRegistered.value = await IsFileAssocRegistered() }
  catch { fileAssocRegistered.value = false }
}
async function registerFileAssoc() {
  try {
    const list = await RegisterFileAssoc()
    ElMessage.success(`已注册 ${list.length} 个文本扩展名到 EasyText「打开方式」菜单`)
    await refreshFileAssocState()
  } catch (e: any) {
    ElMessage.error('注册失败：' + (e?.message || String(e)))
  }
}
async function unregisterFileAssoc() {
  try {
    await UnregisterFileAssoc()
    ElMessage.success('已取消文件关联')
    await refreshFileAssocState()
  } catch (e: any) {
    ElMessage.error('取消失败：' + (e?.message || String(e)))
  }
}

// 19 套主题 - 来自 settingStore
const themeOptions = computed(() => settingStore.availableThemes.map(t => ({
  value: t.key,
  label: t.name + (t.isDark ? ' (暗色)' : ' (亮色)'),
})))

// 宏管理
const macros = computed(() => editorStore.macroState.savedMacros)
const editingMacroId = ref('')
const editingMacroName = ref('')

function startEditingMacro(macroId: string, currentName: string) {
  editingMacroId.value = macroId
  editingMacroName.value = currentName
}

function saveMacroName() {
  if (editingMacroId.value && editingMacroName.value.trim()) {
    editorStore.renameMacro(editingMacroId.value, editingMacroName.value.trim())
    editorStore.saveMacros()
    editingMacroId.value = ''
    ElMessage.success('宏名称已更新')
  }
}

function deleteMacroById(macroId: string) {
  editorStore.deleteMacro(macroId)
  editorStore.saveMacros()
  ElMessage.success('宏已删除')
}

// 🆕 V2.0.0 代码片段管理
async function importSnippets() {
  try {
    const path = await OpenFileDialog()
    if (!path) return
    const result = await ReadFile(path)
    if (!result) return
    const count = await ImportSnippets(result.content)
    ElMessage.success(`已导入 ${count} 个片段`)
    // 刷新列表
    const list = await GetSnippets('')
    editorStore.snippets = list || []
  } catch (e: any) {
    ElMessage.error(`导入失败: ${e?.message || ''}`)
  }
}

async function exportSnippets() {
  try {
    const path = await SaveFileDialog('snippets.json')
    if (!path) return
    const json = await ExportSnippets()
    await SaveFile(path, json, 'UTF-8')
    ElMessage.success('片段已导出')
  } catch (e: any) {
    ElMessage.error(`导出失败: ${e?.message || ''}`)
  }
}

async function deleteSnippetItem(id: number) {
  try {
    await DeleteSnippet(id)
    const list = await GetSnippets('')
    editorStore.snippets = list || []
    ElMessage.success('片段已删除')
  } catch (e: any) {
    ElMessage.error(`删除失败: ${e?.message || ''}`)
  }
}

async function saveSettings() {
  const newConfig = {
    version: 3, // 🆕 V2.0.0: v3 schema
    editor: {
      fontSize: form.fontSize, fontFamily: form.fontFamily,
      tabSize: form.tabSize, insertSpaces: form.insertSpaces,
      wordWrap: form.wordWrap, lineNumbers: form.lineNumbers,
      autoSave: form.autoSave, autoSaveInterval: form.autoSaveInterval,
      highlightLine: form.highlightLine, bracketPairColor: form.bracketPairColor,
      minimap: config.value.editor.minimap,
      showIndentGuide: form.showIndentGuide, showWhitespace: form.showWhitespace,
      showEol: form.showEol, foldEnable: config.value.editor.foldEnable,
      columnMode: config.value.editor.columnMode,
      virtualSpace: config.value.editor.virtualSpace, scrollPastEnd: config.value.editor.scrollPastEnd,
      columnModeConfig: config.value.editor.columnModeConfig,
    },
    theme: {
      currentTheme: form.currentTheme as ThemeName, fontSize: form.fontSize,
      fontFamily: form.fontFamily, tabSize: form.tabSize,
      autoTheme: form.autoTheme, // 🆕 V2.0.0
    },
    file: {
      defaultEncoding: form.defaultEncoding, autoDetectEncoding: form.autoDetectEncoding,
      defaultLineEnding: form.defaultLineEnding,
      ignorePatterns: form.ignorePatterns.split(',').map((s: string) => s.trim()).filter(Boolean), // 🆕 V2.0.0
    },
    ui: {
      language: form.language, showFileTree: form.showFileTree,
      showStatusBar: form.showStatusBar, showToolBar: form.showToolBar,
      showFileListView: form.showFileListView, showWebAddr: config.value.ui.showWebAddr,
      fileTreeWidth: config.value.ui.fileTreeWidth, zoomLevel: form.zoomLevel,
      toolbarIconSize: config.value.ui.toolbarIconSize, favorites: config.value.ui.favorites || [],
      statusBarItems: config.value.ui.statusBarItems || {}, // 🆕 V2.0.0
      toolbarItems: config.value.ui.toolbarItems || {}, // 🆕 V2.0.0
      recentFilesLimit: form.recentFilesLimit, // 🆕 用户可调，默认 10
      lastFolder: config.value.ui.lastFolder || '', // 🆕 V2.0.0
      closeToTray: form.closeToTray, // 🆕 关闭时最小化到托盘
    },
  }
  try {
    await UpdateConfig(newConfig as unknown as import('../../wailsjs/go/models').config.AppConfig)
    settingStore.setConfig(newConfig)
    settingStore.applyTheme()
    // 🆕 V2.0.0 保存自动保存模式
    settingStore.setAutoSaveMode(form.autoSaveMode as 'interval' | 'blur' | 'both')
    ElMessage.success('设置已保存')
    emit('close')
  } catch (error) {
    ElMessage.error(`保存失败: ${error}`)
  }
}

// 快捷键编辑
const shortcutFilter = ref('')
const waitForKey = ref(false)
const waitShortcutId = ref('')
const filteredShortcuts = computed(() => {
  const f = shortcutFilter.value.toLowerCase()
  if (!f) return settingStore.shortcuts
  return settingStore.shortcuts.filter(s =>
    s.name.toLowerCase().includes(f) || s.id.toLowerCase().includes(f) || s.category.toLowerCase().includes(f)
  )
})

const categories = computed(() => [...new Set(settingStore.shortcuts.map(s => s.category))])

function startEditKey(id: string) {
  waitForKey.value = true
  waitShortcutId.value = id
  settingStore.editingShortcutId = id
}

function handleKeyCapture(e: KeyboardEvent) {
  if (!waitForKey.value) return
  e.preventDefault(); e.stopPropagation()
  let key = ''
  if (e.ctrlKey || e.metaKey) key += 'Ctrl+'
  if (e.shiftKey) key += 'Shift+'
  if (e.altKey) key += 'Alt+'
  const k = e.key
  if (k === 'Control' || k === 'Shift' || k === 'Alt' || k === 'Meta') return
  if (k === ' ') key += 'Space'
  else if (k.length === 1) key += k.toUpperCase()
  else key += k
  settingStore.updateShortcut(waitShortcutId.value, key)
  waitForKey.value = false
  waitShortcutId.value = ''
  settingStore.editingShortcutId = null
  ElMessage.success(`快捷键已更新: ${key}`)
}

watch(() => props.visible, (v) => {
  if (v) {
    activePage.value = 'editor'
    refreshFileAssocState()
    Object.assign(form, {
      fontSize: config.value.editor.fontSize, fontFamily: config.value.editor.fontFamily,
      tabSize: config.value.editor.tabSize, insertSpaces: config.value.editor.insertSpaces,
      wordWrap: config.value.editor.wordWrap, lineNumbers: config.value.editor.lineNumbers,
      autoSave: config.value.editor.autoSave, autoSaveInterval: config.value.editor.autoSaveInterval,
      highlightLine: config.value.editor.highlightLine, bracketPairColor: config.value.editor.bracketPairColor,
      showIndentGuide: config.value.editor.showIndentGuide,
      showWhitespace: config.value.editor.showWhitespace, showEol: config.value.editor.showEol,
      currentTheme: config.value.theme.currentTheme,
      defaultEncoding: config.value.file.defaultEncoding, autoDetectEncoding: config.value.file.autoDetectEncoding,
      defaultLineEnding: config.value.file.defaultLineEnding,
      language: config.value.ui.language, showFileTree: config.value.ui.showFileTree,
      showStatusBar: config.value.ui.showStatusBar, showToolBar: config.value.ui.showToolBar,
      showFileListView: config.value.ui.showFileListView, zoomLevel: config.value.ui.zoomLevel,
      // 🆕 V2.0.0
      autoTheme: config.value.theme.autoTheme ?? false,
      autoSaveMode: settingStore.autoSaveMode,
      ignorePatterns: (config.value.file.ignorePatterns || ['.git', '~$*', '*.tmp', 'node_modules']).join(', '),
      closeToTray: config.value.ui.closeToTray ?? true,
      recentFilesLimit: config.value.ui.recentFilesLimit || 10,
    })
  }
})
</script>

<template>
  <el-dialog :model-value="visible" title="首选项" width="780px" :close-on-click-modal="false"
    @update:model-value="(val: boolean) => { if (!val) emit('close') }"
    @keydown="handleKeyCapture">
    <div class="flex" style="height:460px;">
      <div class="w-44 border-r border-gray-200 dark:border-gray-700 pr-2 mr-3 overflow-y-auto">
        <div v-for="p in pages" :key="p.key"
          class="flex items-center px-3 py-2 rounded text-sm cursor-pointer mb-1"
          :class="activePage===p.key?'bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400':'hover:bg-gray-50 dark:hover:bg-gray-800 text-gray-600 dark:text-gray-400'"
          @click="activePage=p.key">{{ p.label }}</div>
      </div>
      <div class="flex-1 overflow-y-auto pr-2">
        <!-- 编辑器 -->
        <div v-show="activePage==='editor'" class="space-y-3">
          <fieldset class="border border-gray-200 dark:border-gray-600 rounded p-3">
            <legend class="text-xs text-gray-600 dark:text-gray-300 px-1">Tab 设置</legend>
            <div class="flex items-center gap-4">
              <div class="flex items-center gap-2"><label class="text-xs text-gray-600 dark:text-gray-300">Tab 长度:</label><el-input-number v-model="form.tabSize" :min="1" :max="16" size="small"/></div>
              <el-switch v-model="form.insertSpaces" active-text="空格替换Tab" size="small"/>
            </div>
          </fieldset>
          <fieldset class="border border-gray-200 dark:border-gray-600 rounded p-3">
            <legend class="text-xs text-gray-600 dark:text-gray-300 px-1">字体设置</legend>
            <el-form label-position="top" size="small">
              <el-form-item label="字体大小"><el-input-number v-model="form.fontSize" :min="8" :max="32" :step="1" controls-position="right" style="width:100%"/></el-form-item>
              <el-form-item label="字体"><el-select v-model="form.fontFamily" style="width:100%"><el-option v-for="f in fontOptions" :key="f.value" :label="f.label" :value="f.value"/></el-select></el-form-item>
            </el-form>
          </fieldset>
          <fieldset class="border border-gray-200 dark:border-gray-600 rounded p-3">
            <legend class="text-xs text-gray-600 dark:text-gray-300 px-1">编辑器选项</legend>
            <div class="grid grid-cols-2 gap-2">
              <el-switch v-model="form.wordWrap" active-text="自动换行" size="small"/>
              <el-switch v-model="form.lineNumbers" active-text="显示行号" size="small"/>
              <el-switch v-model="form.highlightLine" active-text="高亮当前行" size="small"/>
              <el-switch v-model="form.bracketPairColor" active-text="括号配对颜色" size="small"/>
              <el-switch v-model="form.showIndentGuide" active-text="显示缩进参考线" size="small"/>
              <el-switch v-model="form.showWhitespace" active-text="显示空白字符" size="small"/>
              <el-switch v-model="form.showEol" active-text="显示行尾符" size="small"/>
              <el-switch v-model="form.autoSave" active-text="自动保存" size="small"/>
            </div>
            <div v-if="form.autoSave" class="flex items-center gap-2 mt-2">
              <label class="text-xs text-gray-600 dark:text-gray-300">间隔(秒):</label>
              <el-input-number v-model="form.autoSaveInterval" :min="10" :max="600" :step="10" size="small"/>
              <!-- 🆕 V2.0.0 自动保存模式 -->
              <label class="text-xs text-gray-600 dark:text-gray-300 ml-4">触发方式:</label>
              <select v-model="form.autoSaveMode" class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200">
                <option value="interval">定时</option>
                <option value="blur">失焦</option>
                <option value="both">定时+失焦</option>
              </select>
            </div>
          </fieldset>
        </div>
        <!-- 主题 -->
        <div v-show="activePage==='theme'" class="space-y-3">
          <el-form label-position="top" size="small">
            <el-form-item label="主题 (共19套，匹配 notepad-- styleset)">
              <el-select v-model="form.currentTheme" style="width:100%">
                <el-option v-for="t in themeOptions" :key="t.value" :label="t.label" :value="t.value"/>
              </el-select>
            </el-form-item>
            <!-- 🆕 V2.0.0 自动主题切换 -->
            <el-form-item>
              <el-switch v-model="form.autoTheme" active-text="自动跟随系统主题" size="small"/>
              <div class="text-xs text-gray-400 mt-1">启用后，编辑器将根据 Windows 系统深浅色模式自动切换主题</div>
            </el-form-item>
            <el-form-item>
              <div class="text-xs text-gray-400 mt-2">
                支持 19 套主题：Default, DarkDefault, Monokai, OneDark, Dracula, NightOwl, Material, Cobalt, Solarized 等
              </div>
            </el-form-item>
          </el-form>
        </div>
        <!-- 文件 -->
        <div v-show="activePage==='file'" class="space-y-3">
          <el-form label-position="top" size="small">
            <el-form-item label="默认编码"><el-select v-model="form.defaultEncoding" style="width:100%"><el-option v-for="enc in encodingOptions" :key="enc" :label="enc" :value="enc"/></el-select></el-form-item>
            <el-form-item><el-switch v-model="form.autoDetectEncoding" active-text="自动检测编码"/></el-form-item>
            <el-form-item label="默认换行符"><el-select v-model="form.defaultLineEnding" style="width:100%"><el-option v-for="le in lineEndingOptions" :key="le.value" :label="le.label" :value="le.value"/></el-select></el-form-item>
            <!-- 🆕 V2.0.0 忽略文件模式 -->
            <el-form-item label="文件树忽略模式">
              <input v-model="form.ignorePatterns" placeholder=".git, ~$*, *.tmp, node_modules" class="w-full px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200"/>
              <div class="text-xs text-gray-400 mt-1">用逗号分隔，支持通配符 * 和 ?</div>
            </el-form-item>
            <el-form-item><el-switch v-model="form.restoreFilesOnClose" active-text="关闭时恢复文件"/></el-form-item>
          </el-form>
        </div>
        <!-- 界面 -->
        <div v-show="activePage==='ui'" class="space-y-3">
          <el-form label-position="top" size="small">
            <el-form-item label="界面语言"><el-select v-model="form.language" style="width:100%"><el-option v-for="lang in languageOptions" :key="lang.value" :label="lang.label" :value="lang.value"/></el-select></el-form-item>
            <el-form-item label="缩放百分比"><el-input-number v-model="form.zoomLevel" :min="50" :max="200" :step="10" size="small"/></el-form-item>
            <el-form-item>
              <div class="flex flex-col gap-2">
                <el-switch v-model="form.showFileTree" active-text="显示文件树"/>
                <el-switch v-model="form.showStatusBar" active-text="显示状态栏"/>
                <el-switch v-model="form.showToolBar" active-text="显示工具栏"/>
                <el-switch v-model="form.showFileListView" active-text="显示文件列表视图"/>
              </div>
            </el-form-item>
            <!-- 🆕 关闭行为：默认开托盘驻留，避免误关 -->
            <el-form-item>
              <el-switch v-model="form.closeToTray" active-text="关闭时最小化到托盘（防止误操作）"/>
              <div class="text-xs text-gray-400 mt-1">
                开启后点击关闭按钮会把窗口隐藏到任务栏托盘；关闭时直接退出进程。可在托盘菜单点击「显示主窗口」或「退出」。
              </div>
            </el-form-item>
            <!-- 🆕 最近文件保留条数 -->
            <el-form-item label="最近文件保留条数">
              <el-input-number v-model="form.recentFilesLimit" :min="3" :max="50" :step="1" controls-position="right" size="small"/>
              <div class="text-xs text-gray-400 mt-1">文件菜单「最近打开的文件」与最近文件对话框共用此上限，默认 10 条</div>
            </el-form-item>
            <!-- 🆕 便携模式文件关联（不需要管理员权限） -->
            <el-form-item label="系统集成">
              <div class="flex items-center gap-2">
                <span class="text-xs" :class="fileAssocRegistered ? 'text-green-600 dark:text-green-400' : 'text-gray-500'">
                  {{ fileAssocRegistered ? '● 已注册（出现在「打开方式」菜单）' : '○ 未注册（双击文本不会用 EasyText 打开）' }}
                </span>
                <button v-if="!fileAssocRegistered" class="ndd-btn" @click="registerFileAssoc">注册为文本编辑器</button>
                <button v-else class="ndd-btn" @click="unregisterFileAssoc">取消注册</button>
              </div>
              <div class="text-xs text-gray-400 mt-1">
                把 .txt/.log/.md/.json/.yaml/.xml/.ini/.csv 等常用文本后缀关联到 EasyText。
                注册后右键 → 打开方式 → EasyText，或选「始终用 EasyText 打开」即可接管。
                不会修改既有默认编辑器，仅写入 HKCU，无需管理员权限。
              </div>
            </el-form-item>
          </el-form>
        </div>
        <!-- 快捷键管理 -->
        <div v-show="activePage==='shortcut'" class="space-y-2">
          <div class="flex items-center gap-2 mb-2">
            <input v-model="shortcutFilter" placeholder="搜索快捷键..." class="flex-1 px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200"/>
            <button class="ndd-btn" @click="settingStore.resetShortcuts();ElMessage.success('已恢复默认')">恢复默认</button>
          </div>
          <div v-if="waitForKey" class="mb-2 px-3 py-2 text-xs bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-300 dark:border-yellow-700 rounded text-yellow-700 dark:text-yellow-300">
            正在等待按键... 请按下组合键 (如 Ctrl+Shift+K)
          </div>
          <div v-for="cat in categories" :key="cat" class="mb-3">
            <div class="text-xs font-medium text-gray-600 dark:text-gray-300 mb-1 px-1">{{ cat }}</div>
            <div v-for="s in filteredShortcuts.filter(x=>x.category===cat)" :key="s.id"
              class="flex items-center justify-between px-2 py-1 text-xs border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-[#3c3c3c]">
              <span class="text-gray-600 dark:text-gray-300">{{ s.name }}</span>
              <div class="flex items-center gap-1">
                <span v-if="settingStore.editingShortcutId===s.id" class="px-2 py-0.5 bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 rounded text-xs animate-pulse">按下新快捷键...</span>
                <span v-else class="px-2 py-0.5 bg-gray-100 dark:bg-[#3c3c3c] rounded text-xs text-gray-500 dark:text-gray-400 cursor-pointer hover:bg-gray-200 dark:hover:bg-[#4c4c4c]" @click="startEditKey(s.id)">{{ s.currentKey || '(未设置)' }}</span>
              </div>
            </div>
          </div>
        </div>
        <!-- 宏管理 -->
        <div v-show="activePage==='macro'" class="space-y-2">
          <div class="text-xs text-gray-500 mb-2">管理已录制的宏。录制宏: Ctrl+Shift+R，播放宏: Ctrl+Shift+P</div>
          <div v-if="macros.length===0" class="text-xs text-gray-400 py-8 text-center">暂无保存的宏</div>
          <div v-for="m in macros" :key="m.id" class="flex items-center justify-between px-2 py-1.5 text-xs border border-gray-200 dark:border-gray-600 rounded mb-1 hover:bg-gray-50 dark:hover:bg-[#3c3c3c]">
            <div class="flex-1">
              <template v-if="editingMacroId===m.id">
                <input v-model="editingMacroName" class="px-1 py-0.5 border border-gray-300 dark:border-gray-500 rounded text-xs w-40" @keydown.enter="saveMacroName"/>
                <button class="ml-1 text-blue-500 text-xs" @click="saveMacroName">保存</button>
              </template>
              <template v-else>
                <span class="text-gray-600 dark:text-gray-300">{{ m.name }}</span>
                <span class="ml-2 text-gray-400">({{ m.steps.length }} 步)</span>
              </template>
            </div>
            <div class="flex gap-1">
              <button class="text-blue-500 text-xs hover:underline" @click="startEditingMacro(m.id,m.name)">重命名</button>
              <button class="text-red-500 text-xs hover:underline" @click="deleteMacroById(m.id)">删除</button>
            </div>
          </div>
        </div>

        <!-- 🆕 V2.0.0 代码片段管理 -->
        <div v-show="activePage==='snippets'" class="space-y-2">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs text-gray-500">管理代码片段，支持导入/导出 JSON 格式 (兼容 VS Code)</span>
            <div class="flex gap-1">
              <button class="ndd-btn" @click="importSnippets">导入</button>
              <button class="ndd-btn" @click="exportSnippets">导出</button>
            </div>
          </div>
          <div v-if="editorStore.snippets.length===0" class="text-xs text-gray-400 py-8 text-center">暂无代码片段，在侧边栏「代码片段」面板中创建</div>
          <div v-for="s in editorStore.snippets" :key="s.id" class="flex items-center justify-between px-2 py-1.5 text-xs border border-gray-200 dark:border-gray-600 rounded mb-1 hover:bg-gray-50 dark:hover:bg-[#3c3c3c]">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-1.5">
                <span class="text-gray-600 dark:text-gray-300 font-medium">{{ s.name }}</span>
                <span v-if="s.language" class="text-[10px] px-1 rounded bg-gray-100 dark:bg-gray-700 text-gray-400">{{ s.language }}</span>
              </div>
              <div class="text-gray-400 font-mono truncate">{{ s.prefix }}</div>
            </div>
            <div class="flex gap-1">
              <button class="text-red-500 text-xs hover:underline" @click="deleteSnippetItem(s.id)">删除</button>
            </div>
          </div>
        </div>
      </div>
    </div>
    <template #footer>
      <el-button @click="emit('close')">取消</el-button>
      <el-button type="primary" @click="saveSettings">保存</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.ndd-btn { padding: 2px 8px; font-size: 12px; border: 1px solid #d1d5db; background: #fff; color: #374151; border-radius: 3px; cursor: pointer; }
.ndd-btn:hover { background: #f3f4f6; }
html.dark .ndd-btn { background: #3c3c3c; color: #e0e0e0; border-color: #555; }
::deep(.el-form-item) { margin-bottom: 12px; }
::deep(.el-form-item__label) { font-size: 13px; padding-bottom: 4px; }
</style>
