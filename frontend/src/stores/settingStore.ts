import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { AppConfig, ThemeName, ShortcutDef } from '@/types'
import { getThemeColors } from '@/utils'

export const useSettingStore = defineStore('setting', () => {
  // ============ 默认配置 ============
  const config = ref<AppConfig>({
    version: 3,
    editor: {
      fontSize: 14,
      fontFamily: 'Consolas, Monaco, "Courier New", monospace',
      tabSize: 4,
      insertSpaces: true,
      wordWrap: false,
      lineNumbers: true,
      autoSave: false,
      autoSaveInterval: 60,
      highlightLine: true,
      bracketPairColor: true,
      minimap: true,
      showIndentGuide: true,
      showWhitespace: false,
      showEol: false,
      foldEnable: true,
      columnMode: false,
      virtualSpace: false,
      scrollPastEnd: false,
      columnModeConfig: {
        numberStart: 1,
        numberStep: 1,
        numberBase: 10,
        dateFormat: '2006-01-02',
        caseConversion: '',
      },
    },
    theme: {
      currentTheme: 'Default',
      fontSize: 14,
      fontFamily: 'Consolas, Monaco, "Courier New", monospace',
      tabSize: 4,
      autoTheme: false,
    },
    file: {
      defaultEncoding: 'UTF-8',
      autoDetectEncoding: true,
      defaultLineEnding: 'CRLF',
      ignorePatterns: ['.git', '~$*', '*.tmp', 'node_modules'],
    },
    ui: {
      language: 'zh-CN',
      showFileTree: true,
      showStatusBar: true,
      showToolBar: true,
      showFileListView: false,
      showWebAddr: false,
      fileTreeWidth: 250,
      zoomLevel: 100,
      toolbarIconSize: 24,
      favorites: [],
      lastFolder: '',
      statusBarItems: {
        cursor: true, lines: true, encoding: true,
        lineEnding: true, language: true, path: true, zoom: true,
      },
      toolbarItems: {},
      recentFilesLimit: 10,
      // 默认开托盘驻留：避免用户误关后丢失正在编辑的内容。
      closeToTray: true,
    },
  })

  // ============ 19 套主题 - 匹配 notepad-- styleset.cpp ============
  const availableThemes: { key: ThemeName; name: string; isDark: boolean }[] = [
    { key: 'Default', name: 'Default (Light)', isDark: false },
    { key: 'DarkDefault', name: 'Dark Default', isDark: true },
    { key: 'Monokai', name: 'Monokai', isDark: true },
    { key: 'OneDark', name: 'One Dark', isDark: true },
    { key: 'Dracula', name: 'Dracula', isDark: true },
    { key: 'NightOwl', name: 'Night Owl', isDark: true },
    { key: 'Material', name: 'Material', isDark: true },
    { key: 'Cobalt', name: 'Cobalt', isDark: true },
    { key: 'Solarized', name: 'Solarized Light', isDark: false },
    { key: 'SolarizedDark', name: 'Solarized Dark', isDark: true },
    { key: 'BlackBoard', name: 'BlackBoard', isDark: true },
    { key: 'Bespin', name: 'Bespin', isDark: true },
    { key: 'Eiffel', name: 'Eiffel', isDark: false },
    { key: 'Elegant', name: 'Elegant', isDark: false },
    { key: 'ErlangDark', name: 'Erlang Dark', isDark: true },
    { key: 'IDLE', name: 'IDLE', isDark: false },
    { key: 'Lazy', name: 'Lazy', isDark: false },
    { key: 'MonoIndustrial', name: 'Mono Industrial', isDark: true },
    { key: 'Neat', name: 'Neat', isDark: false },
  ]

  // ============ 计算属性 ============
  const isDarkMode = computed(() => {
    const colors = getThemeColors(config.value.theme.currentTheme)
    return colors.isDark
  })

  const currentThemeColors = computed(() => {
    return getThemeColors(config.value.theme.currentTheme)
  })

  const showFileTree = computed(() => config.value.ui.showFileTree)
  const showStatusBar = computed(() => config.value.ui.showStatusBar)

  // ============ 主题操作 ============
  function setConfig(newConfig: AppConfig) {
    config.value = newConfig
    applyTheme()
  }

  function updateTheme(theme: ThemeName) {
    config.value.theme.currentTheme = theme
    applyTheme()
  }

  function toggleTheme() {
    // 在白天/暗色主题之间循环
    const current = availableThemes.find(t => t.key === config.value.theme.currentTheme)
    const currentIsDark = current?.isDark ?? false
    const target = availableThemes.find(t => t.isDark !== currentIsDark)
    if (target) {
      config.value.theme.currentTheme = target.key
      applyTheme()
    }
  }

  function applyTheme() {
    const colors = getThemeColors(config.value.theme.currentTheme)
    const html = document.documentElement
    const root = document.documentElement.style

    // Tailwind dark class
    if (colors.isDark) html.classList.add('dark')
    else html.classList.remove('dark')

    // CSS 变量 - 匹配 notepad-- mystyle.qss
    const vars: [string, string][] = [
      ['--theme-bg', colors.bg],
      ['--theme-fg', colors.fg],
      ['--theme-gutter-bg', colors.gutterBg],
      ['--theme-gutter-fg', colors.gutterFg],
      ['--theme-selection', colors.selection],
      ['--theme-active-line', colors.activeLine],
      ['--theme-cursor', colors.cursor],
      ['--theme-accent', colors.accent],
      ['--theme-comment', colors.comment],
      ['--theme-keyword', colors.keyword],
      ['--theme-string', colors.string],
      ['--theme-number', colors.number],
      ['--theme-type', colors.type],
      ['--theme-function', colors.function],
      ['--theme-variable', colors.variable],
      ['--theme-operator', colors.operator],
      ['--theme-sidebar-bg', colors.sidebarBg],
      ['--theme-sidebar-fg', colors.sidebarFg],
      ['--theme-menu-bg', colors.menuBg],
      ['--theme-menu-fg', colors.menuFg],
      ['--theme-toolbar-bg', colors.toolbarBg],
      ['--theme-statusbar-bg', colors.statusBarBg],
      ['--theme-statusbar-fg', colors.statusBarFg],
      ['--theme-tab-active', colors.tabActiveBg],
      ['--theme-tab-inactive', colors.tabInactiveBg],
      ['--theme-find-highlight', colors.findHighlight],
      ['--theme-bookmark', colors.bookmarkFg],
      ['--theme-bracket-match', colors.bracketMatch],
    ]
    for (const [key, value] of vars) {
      root.setProperty(key, value)
    }

    // 保存配置
    saveConfig()
  }

  // ============ UI 操作 ============
  // 所有 UI 操作均自动持久化，避免重启后丢失用户偏好。

  function toggleFileTree() {
    config.value.ui.showFileTree = !config.value.ui.showFileTree
    saveConfig()
  }

  function toggleStatusBar() {
    config.value.ui.showStatusBar = !config.value.ui.showStatusBar
    saveConfig()
  }

  function toggleToolBar() {
    config.value.ui.showToolBar = !config.value.ui.showToolBar
    saveConfig()
  }

  function toggleFileListView() {
    config.value.ui.showFileListView = !config.value.ui.showFileListView
    saveConfig()
  }

  // 拖拽时高频调用，使用 debounced save（替代直接赋值）
  let fileTreeWidthTimer: ReturnType<typeof setTimeout> | null = null
  function setFileTreeWidth(width: number) {
    config.value.ui.fileTreeWidth = width
    if (fileTreeWidthTimer) clearTimeout(fileTreeWidthTimer)
    fileTreeWidthTimer = setTimeout(() => {
      saveConfig()
      fileTreeWidthTimer = null
    }, 300) // 拖拽停顿 300ms 后才落盘
  }

  function setLanguage(lang: string) {
    config.value.ui.language = lang
    saveConfig()
  }

  function setEditorFontSize(size: number) {
    config.value.editor.fontSize = size
    config.value.theme.fontSize = size
    saveConfig()
  }

  function setEditorFontFamily(family: string) {
    config.value.editor.fontFamily = family
    config.value.theme.fontFamily = family
    saveConfig()
  }

  function setTabSize(size: number) {
    config.value.editor.tabSize = size
    config.value.theme.tabSize = size
    saveConfig()
  }

  function setAutoSave(enabled: boolean, interval?: number) {
    config.value.editor.autoSave = enabled
    if (interval !== undefined) config.value.editor.autoSaveInterval = interval
    saveConfig()
  }

  // ============ 快捷键管理（notepad-- shortcutkeymgr）============
  const shortcuts = ref<ShortcutDef[]>([
    { id: 'new-file', name: '新建文件', category: '文件', defaultKey: 'Ctrl+N', currentKey: 'Ctrl+N' },
    { id: 'open-file', name: '打开文件', category: '文件', defaultKey: 'Ctrl+O', currentKey: 'Ctrl+O' },
    { id: 'save', name: '保存', category: '文件', defaultKey: 'Ctrl+S', currentKey: 'Ctrl+S' },
    { id: 'save-as', name: '另存为', category: '文件', defaultKey: 'Ctrl+Shift+S', currentKey: 'Ctrl+Shift+S' },
    { id: 'save-all', name: '全部保存', category: '文件', defaultKey: 'Ctrl+Alt+S', currentKey: 'Ctrl+Alt+S' },
    { id: 'close-tab', name: '关闭标签', category: '文件', defaultKey: 'Ctrl+W', currentKey: 'Ctrl+W' },
    { id: 'close-all', name: '关闭所有', category: '文件', defaultKey: 'Ctrl+Shift+W', currentKey: 'Ctrl+Shift+W' },
    { id: 'print', name: '打印', category: '文件', defaultKey: 'Ctrl+P', currentKey: 'Ctrl+P' },
    { id: 'exit', name: '退出', category: '文件', defaultKey: 'Ctrl+Q', currentKey: 'Ctrl+Q' },
    { id: 'undo', name: '撤销', category: '编辑', defaultKey: 'Ctrl+Z', currentKey: 'Ctrl+Z' },
    { id: 'redo', name: '重做', category: '编辑', defaultKey: 'Ctrl+Y', currentKey: 'Ctrl+Y' },
    { id: 'cut', name: '剪切', category: '编辑', defaultKey: 'Ctrl+X', currentKey: 'Ctrl+X' },
    { id: 'copy', name: '复制', category: '编辑', defaultKey: 'Ctrl+C', currentKey: 'Ctrl+C' },
    { id: 'paste', name: '粘贴', category: '编辑', defaultKey: 'Ctrl+V', currentKey: 'Ctrl+V' },
    { id: 'select-all', name: '全选', category: '编辑', defaultKey: 'Ctrl+A', currentKey: 'Ctrl+A' },
    { id: 'duplicate-line', name: '复制当前行', category: '编辑', defaultKey: 'Ctrl+D', currentKey: 'Ctrl+D' },
    { id: 'delete-line', name: '删除当前行', category: '编辑', defaultKey: 'Ctrl+L', currentKey: 'Ctrl+L' },
    { id: 'move-line-up', name: '上移当前行', category: '编辑', defaultKey: 'Ctrl+Shift+Up', currentKey: 'Ctrl+Shift+Up' },
    { id: 'move-line-down', name: '下移当前行', category: '编辑', defaultKey: 'Ctrl+Shift+Down', currentKey: 'Ctrl+Shift+Down' },
    { id: 'find', name: '查找', category: '查找', defaultKey: 'Ctrl+F', currentKey: 'Ctrl+F' },
    { id: 'find-next', name: '查找下一个', category: '查找', defaultKey: 'F3', currentKey: 'F3' },
    { id: 'find-prev', name: '查找上一个', category: '查找', defaultKey: 'Shift+F3', currentKey: 'Shift+F3' },
    { id: 'replace', name: '替换', category: '查找', defaultKey: 'Ctrl+H', currentKey: 'Ctrl+H' },
    { id: 'find-in-dir', name: '在目录中查找', category: '查找', defaultKey: 'Ctrl+Shift+F', currentKey: 'Ctrl+Shift+F' },
    { id: 'goto-line', name: '转到行', category: '查找', defaultKey: 'Ctrl+G', currentKey: 'Ctrl+G' },
    { id: 'toggle-bookmark', name: '切换书签', category: '查找', defaultKey: 'Ctrl+F2', currentKey: 'Ctrl+F2' },
    { id: 'next-bookmark', name: '下一个书签', category: '查找', defaultKey: 'F2', currentKey: 'F2' },
    { id: 'prev-bookmark', name: '上一个书签', category: '查找', defaultKey: 'Shift+F2', currentKey: 'Shift+F2' },
    { id: 'zoom-in', name: '放大', category: '视图', defaultKey: 'Ctrl+=', currentKey: 'Ctrl+=' },
    { id: 'zoom-out', name: '缩小', category: '视图', defaultKey: 'Ctrl+-', currentKey: 'Ctrl+-' },
    { id: 'fullscreen', name: '全屏', category: '视图', defaultKey: 'F11', currentKey: 'F11' },
    { id: 'toggle-wrap', name: '自动换行', category: '视图', defaultKey: '', currentKey: '' },
    { id: 'format-json', name: '格式化JSON', category: '工具', defaultKey: 'Ctrl+Shift+J', currentKey: 'Ctrl+Shift+J' },
    { id: 'word-highlight', name: '高亮单词', category: '工具', defaultKey: 'F8', currentKey: 'F8' },
    { id: 'clear-highlight', name: '清除高亮', category: '工具', defaultKey: 'F7', currentKey: 'F7' },
    { id: 'record-macro', name: '开始/停止录制宏', category: '宏', defaultKey: 'Ctrl+Shift+R', currentKey: 'Ctrl+Shift+R' },
    { id: 'play-macro', name: '播放宏', category: '宏', defaultKey: 'Ctrl+Shift+P', currentKey: 'Ctrl+Shift+P' },
  ])

  const isEditingShortcut = ref(false)
  const editingShortcutId = ref<string | null>(null)

  function getShortcut(actionId: string): string {
    return shortcuts.value.find(s => s.id === actionId)?.currentKey || ''
  }

  function updateShortcut(id: string, newKey: string) {
    const s = shortcuts.value.find(t => t.id === id)
    if (s) s.currentKey = newKey
  }

  function resetShortcuts() {
    for (const s of shortcuts.value) s.currentKey = s.defaultKey
  }

  // ============ 🆕 V2.0.0 自动保存模式 ============
  const autoSaveMode = ref<'interval' | 'blur' | 'both'>('interval')

  function setAutoSaveMode(mode: 'interval' | 'blur' | 'both') {
    autoSaveMode.value = mode
  }

  // ============ 🆕 V2.0.0 自动主题切换 ============
  const autoTheme = computed(() => config.value.theme.autoTheme)
  // 用 AbortController 替代 addEventListener/removeEventListener 引用匹配，
  // 避免 disable 时无法真正移除监听器导致"自动主题关闭后系统暗色切换仍生效"的 bug。
  let autoThemeAbort: AbortController | null = null

  function setAutoTheme(enabled: boolean) {
    config.value.theme.autoTheme = enabled
    if (enabled) {
      enableAutoTheme()
    } else {
      disableAutoTheme()
    }
  }

  function enableAutoTheme() {
    disableAutoTheme() // 先清理旧的，避免叠加监听器
    autoThemeAbort = new AbortController()
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    const handler = (e: MediaQueryListEvent) => {
      if (config.value.theme.autoTheme) {
        const target = e.matches ? 'DarkDefault' : 'Default'
        config.value.theme.currentTheme = target as ThemeName
        applyTheme()
      }
    }
    // 立即应用当前系统主题
    handler(mq as unknown as MediaQueryListEvent)
    mq.addEventListener('change', handler, { signal: autoThemeAbort.signal })
  }

  function disableAutoTheme() {
    autoThemeAbort?.abort()
    autoThemeAbort = null
  }

  // ============ 🆕 V2.0.0 状态栏/工具栏自定义 ============
  function toggleStatusBarItem(item: string) {
    config.value.ui.statusBarItems[item] = !config.value.ui.statusBarItems[item]
  }

  function toggleToolbarItem(item: string) {
    config.value.ui.toolbarItems[item] = !config.value.ui.toolbarItems[item]
  }

  // ============ 🆕 V2.0.0 忽略模式 ============
  function setIgnorePatterns(patterns: string[]) {
    config.value.file.ignorePatterns = patterns
  }

  // ============ 持久化 ============
  async function saveConfig() {
    try {
      const { UpdateConfig } = await import('../../wailsjs/go/main/App')
      // 直接传 config.value：TS 结构类型允许额外字段，wailsjs 生成的
      // config.AppConfig 与本地 AppConfig 形状兼容。
      await UpdateConfig(config.value as unknown as import('../../wailsjs/go/models').config.AppConfig)
    } catch (e) {
      console.error('Failed to save config:', e)
    }
  }

  return {
    config, availableThemes, isDarkMode, currentThemeColors,
    showFileTree, showStatusBar,
    shortcuts, isEditingShortcut, editingShortcutId,
    setConfig, updateTheme, toggleTheme, applyTheme,
    toggleFileTree, toggleStatusBar, toggleToolBar, toggleFileListView,
    setFileTreeWidth, setLanguage,
    setEditorFontSize, setEditorFontFamily, setTabSize, setAutoSave,
    saveConfig,
    getShortcut, updateShortcut, resetShortcuts,
    // 🆕 V2.0.0
    autoSaveMode, setAutoSaveMode,
    autoTheme, setAutoTheme, enableAutoTheme, disableAutoTheme,
    toggleStatusBarItem, toggleToolbarItem,
    setIgnorePatterns,
  }
})
