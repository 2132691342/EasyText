import { ElMessage, ElMessageBox } from 'element-plus'
import { OpenFileDialog, SaveFileDialog, ReadFile, SaveFile, ReadFileBytes, ShowConfirmDialog, ConvertToUTF8, RenameFile } from '../../wailsjs/go/main/App'
import { getFileExtension } from '@/utils'
import { useFormatConverterStore } from '@/stores'

/**
 * 命令分发 composable
 * 将 MainLayout 中 150+ 命令的处理逻辑抽离到这里，
 * 减少 MainLayout.vue 的体积，提升可维护性。
 */
export function useCommands(deps: {
  ed: ReturnType<typeof import('@/stores').useEditorStore>
  fs: ReturnType<typeof import('@/stores').useFileStore>
  ss: ReturnType<typeof import('@/stores').useSettingStore>
  // 浮层开关 refs
  showFindWin: { value: boolean }
  showDiff: { value: boolean }
  showSettings: { value: boolean }
  showFileList: { value: boolean }
  showFileTree: { value: boolean }
  showMd5: { value: boolean }
  showColumnEdit: { value: boolean }
  showBatchFind: { value: boolean }
  showBatchRename: { value: boolean }
  showEncodeConvert: { value: boolean }
  showFormatConverter: { value: boolean }
  showCmpRule: { value: boolean }
  showDirCmp: { value: boolean }
  showClipboardHistory: { value: boolean }
  showRegexTester: { value: boolean }
  showSnippetPanel: { value: boolean }
  showBookmarkPanel: { value: boolean }
  showFunctionList: { value: boolean }
  showFileMonitor: { value: boolean }
  showScriptManager: { value: boolean }
  showColorPicker: { value: boolean }
  showFindResults: { value: boolean }
  findMode: { value: 'find' | 'replace' | 'files' | 'global' | 'mark' }
  findResults: { value: any[] }
  findResultsTitle: { value: string }
  diffLeftPath: { value: string }
  diffRightPath: { value: string }
  // 函数引用
  newFile: () => void
  openFile: () => Promise<void>
  openFilePath: (path: string) => Promise<void>
  openDir: () => Promise<void>
  saveFile: () => Promise<void>
  saveFileAs: () => Promise<void>
  saveAll: () => Promise<void>
  closeTab: (id: string) => Promise<void>
  closeAll: () => Promise<void>
  closeOthers: () => void
  closeLeft: () => void
  closeRight: () => void
  reloadFile: () => Promise<void>
  saveSession: () => void
  saveWorkspace: () => Promise<void>
  openWorkspace: () => Promise<void>
  restoreSession: () => Promise<void>
  toggleMacroRecording: () => void
  playCurrentMacro: () => void
  saveCurrentMacro: () => Promise<void>
  runMacroMulti: () => Promise<void>
  manageFavorites: () => Promise<void>
  importTheme: () => Promise<void>
  exportTheme: () => Promise<void>
  renameCurrentFile: () => Promise<void>
  selectCompareFile: (side: 'left' | 'right') => Promise<void>
  binaryCompare: () => Promise<void>
  reopenAs: (vt: 'code' | 'hex') => Promise<void>
  reopenWithEncoding: (enc: string) => Promise<void>
  convertTo: (enc: string) => void
  execEd: (cmd: string, ...args: any[]) => void
  startTail: () => Promise<void>
  stopTail: () => Promise<void>
  manageDrafts: () => Promise<void>
  toggleAutoTheme: () => void
  showRecentFiles: () => Promise<void>
  showRecentFolders: () => Promise<void>
  showImageEditorView: () => void
  toggleLogMode: () => void
  isMacroRecording: { value: boolean }
}) {

  const { ed, ss, fs } = deps

  function execEd(cmd: string, ...args: any[]) {
    deps.execEd(cmd, ...args)
  }

  function onMenuCmd(name: string, ...args: any[]) {
    const m: Record<string, any> = {
      'new-file': deps.newFile,
      'open-file': deps.openFile,
      'open-directory': deps.openDir,
      'save': deps.saveFile,
      'save-as': deps.saveFileAs,
      'save-all': deps.saveAll,
      'close-tab': () => ed.activeTab && deps.closeTab(ed.activeTab.id),
      'close-all': deps.closeAll,
      'close-others': deps.closeOthers,
      'close-left': deps.closeLeft,
      'close-right': deps.closeRight,
      'reload-file': deps.reloadFile,
      'save-workspace': deps.saveWorkspace,
      'open-workspace': deps.openWorkspace,
      'exit': deps.saveSession,
      'undo': () => execEd('undo'),
      'redo': () => execEd('redo'),
      'cut': () => execEd('cut'),
      'copy': () => execEd('copy'),
      'paste': () => execEd('paste'),
      'select-all': () => execEd('select-all'),
      'find': () => { deps.findMode.value = 'find'; deps.showFindWin.value = true },
      'find-next': () => execEd('find-next'),
      'find-prev': () => execEd('find-prev'),
      'find-dir': () => { deps.findMode.value = 'files'; deps.showFindWin.value = true },
      'replace': () => { deps.findMode.value = 'replace'; deps.showFindWin.value = true },
      'goto-line': () => execEd('goto-line'),
      'toggle-bookmark': () => execEd('toggle-bookmark'),
      'next-bookmark': () => execEd('next-bookmark'),
      'prev-bookmark': () => execEd('prev-bookmark'),
      'clear-bookmarks': () => execEd('clear-bookmarks'),
      'toggle-wrap': () => execEd('toggle-word-wrap'),
      'toggle-filelist': () => { deps.showFileList.value = !deps.showFileList.value },
      'toggle-toolbar': () => {
        if (ss.config) { ss.config.ui.showToolBar = !ss.config.ui.showToolBar; ss.saveConfig() }
      },
      'search-result': () => { deps.showFindResults.value = true },
      'open-text': () => deps.reopenAs('code'),
      'open-hex': () => deps.reopenAs('hex'),
      'column-mode': () => execEd('column-mode'),
      'column-block': () => { deps.showColumnEdit.value = true },
      'preferences': () => { deps.showSettings.value = true },
      'theme-style': () => { deps.showSettings.value = true },
      'define-lang': () => { deps.showSettings.value = true },
      'lang-suffix': () => { deps.showSettings.value = true },
      'shortcut-mgr': () => { deps.showSettings.value = true },
      'md5-hash': () => { deps.showMd5.value = true },
      'batch-find': () => { deps.showBatchFind.value = true },
      'batch-rename': () => { deps.showBatchRename.value = true },
      'batch-convert': () => { deps.showEncodeConvert.value = true },
      'open-converter': () => { deps.showFormatConverter.value = true },
      'explorer': () => openExplorer(),
      'print': () => {
        const t = ed.activeTab
        if (!t) return
        const w = window.open('', '_blank')!
        w.document.write(`<pre>${t.content.replace(/&/g, '&amp;').replace(/</g, '&lt;')}</pre>`)
        w.print(); w.close()
      },
      'fullscreen': async () => {
        // WebView2 里 document.requestFullscreen 无效，走 Wails 原生全屏 API
        const { WindowFullscreen, WindowUnfullscreen, WindowIsFullscreen } = await import('../../wailsjs/runtime/runtime')
        const isFs = await WindowIsFullscreen()
        if (isFs) WindowUnfullscreen()
        else WindowFullscreen()
      },
      'file-cmp': () => { deps.showDiff.value = true },
      'dir-cmp': () => { deps.showDirCmp.value = true },
      'bin-cmp': () => deps.binaryCompare(),
      'sel-left': () => deps.selectCompareFile('left'),
      'sel-right': () => deps.selectCompareFile('right'),
      'cmp-rule': () => { deps.showCmpRule.value = true },
      'recent-cmp': () => ElMessage.info('暂无最近对比记录'),
      'about': () => ElMessageBox.alert('EasyText v2.0.0\n轻量级开发者桌面编辑器', '关于', { type: 'info', confirmButtonText: '确定' }),
      'encode-GBK': () => deps.reopenWithEncoding('GBK'),
      'encode-UTF-8': () => deps.reopenWithEncoding('UTF-8'),
      'encode-UTF-8-BOM': () => deps.reopenWithEncoding('UTF-8-BOM'),
      'encode-UCS-2-BE': () => deps.reopenWithEncoding('UTF-16BE'),
      'encode-UCS-2-LE': () => deps.reopenWithEncoding('UTF-16LE'),
      'encode-Big5': () => deps.reopenWithEncoding('Big5'),
      'fmt-xml': () => execEd('format-xml'),
      'fmt-json': () => execEd('format-json'),
      'le-CRLF': () => { if (ed.activeTab) ed.updateTabLineEnding(ed.activeTab.id, 'CRLF') },
      'le-LF': () => { if (ed.activeTab) ed.updateTabLineEnding(ed.activeTab.id, 'LF') },
      'le-CR': () => { if (ed.activeTab) ed.updateTabLineEnding(ed.activeTab.id, 'CR') },
      'lang-zh': () => {
        if (ss.config) { ss.config.ui.language = 'zh-CN'; ss.saveConfig(); ElMessage.success('已切换为中文，重启后生效') }
      },
      'lang-en': () => {
        if (ss.config) { ss.config.ui.language = 'en-US'; ss.saveConfig(); ElMessage.success('Switched to English. Restart to apply.') }
      },
      'set-lang': (lang: string) => {
        if (ed.activeTab) { ed.activeTab.language = lang.toLowerCase(); ElMessage.success('语言: ' + lang) }
      },
      'mark-color': () => execEd('mark-color'),
      'word-highlight': () => execEd('word-highlight'),
      'clear-all-highlight': () => execEd('clear-all-highlight'),
      'zoom-in': () => {
        if (ss.config) { ss.config.ui.zoomLevel = (ss.config.ui.zoomLevel || 100) + 10; ss.saveConfig() }
      },
      'zoom-out': () => {
        if (ss.config) { ss.config.ui.zoomLevel = Math.max(50, (ss.config.ui.zoomLevel || 100) - 10); ss.saveConfig() }
      },
      'toggle-whitespace': () => execEd('show-whitespace'),
      'toggle-indent-guide': () => execEd('toggle-indent-guide'),
      'toggle-tail': (on: boolean) => on ? deps.startTail() : deps.stopTail(),
      'toggle-auto-save-cycle': (on: boolean) => {
        if (ss.config) { ss.config.editor.autoSave = on; ss.saveConfig() }
      },
      'toggle-webaddr': () => execEd('toggle-webaddr'),
      'delete': () => execEd('delete'),
      'record-macro': deps.toggleMacroRecording,
      'stop-macro': () => { ed.stopMacroPlayback(); ElMessage.success('已停止宏回放') },
      'play-macro': deps.playCurrentMacro,
      'save-macro': deps.saveCurrentMacro,
      'run-macro-multi': deps.runMacroMulti,
      'clipboard-history': () => { deps.showClipboardHistory.value = true },
      'copy-to-clipboard': () => execEd('copy'),
      'indent': () => execEd('indent'),
      'dedent': () => execEd('dedent'),
      'toggle-readonly': () => {
        if (ed.activeTab) {
          ed.activeTab.isReadOnly = !ed.activeTab.isReadOnly
          ElMessage.success(ed.activeTab.isReadOnly ? '只读模式' : '可编辑模式')
        }
      },
      'clear-readonly': () => {
        if (ed.activeTab) { ed.activeTab.isReadOnly = false; ElMessage.success('只读模式已清除') }
      },
      'comment-line': () => execEd('comment-line'),
      'comment-block': () => execEd('comment-block'),
      'find-multi': () => { deps.showBatchFind.value = true },
      'replace-multi': () => { deps.showBatchFind.value = true },
      'mark-all': () => execEd('mark-all'),
      'goto-bracket': () => execEd('goto-bracket'),
      'prev-position': () => execEd('prev-position'),
      'next-position': () => execEd('next-position'),
      'toggle-statusbar': () => {
        if (ss.config) { ss.config.ui.showStatusBar = !ss.config.ui.showStatusBar; ss.saveConfig() }
      },
      'zoom-reset': () => {
        if (ss.config) { ss.config.ui.zoomLevel = 100; ss.saveConfig() }
      },
      'iconsize-24': () => { if (ss.config) { ss.config.ui.toolbarIconSize = 24; ss.saveConfig() } },
      'iconsize-36': () => { if (ss.config) { ss.config.ui.toolbarIconSize = 36; ss.saveConfig() } },
      'iconsize-48': () => { if (ss.config) { ss.config.ui.toolbarIconSize = 48; ss.saveConfig() } },
      'encode-ANSI': () => deps.reopenWithEncoding('GBK'),
      'conv-ANSI': () => deps.convertTo('GBK'),
      'conv-UTF-8': () => deps.convertTo('UTF-8'),
      'conv-UTF-8-BOM': () => deps.convertTo('UTF-8-BOM'),
      'conv-UCS-2-BE': () => deps.convertTo('UTF-16BE'),
      'conv-UCS-2-LE': () => deps.convertTo('UTF-16LE'),
      'open-diff': () => { deps.showDiff.value = true },
      'edit-context-menu': () => { deps.showSettings.value = true },
      'import-plugin': () => ElMessage.info('插件功能已禁用'),
      'import-theme': deps.importTheme,
      'import-shortcut': () => ElMessage.info('快捷键导入：在设置→快捷键页操作'),
      'export-theme': deps.exportTheme,
      'export-shortcut': () => ElMessage.info('快捷键导出：在设置→快捷键页操作'),
      'encode-GB2312': () => deps.reopenWithEncoding('GB2312'),
      'encode-SJIS': () => deps.reopenWithEncoding('Shift_JIS'),
      'encode-ar': () => deps.reopenWithEncoding('ISO-8859-6'),
      'encode-baltic': () => deps.reopenWithEncoding('ISO-8859-13'),
      'encode-ce': () => deps.reopenWithEncoding('ISO-8859-2'),
      'encode-cyrillic': () => deps.reopenWithEncoding('ISO-8859-5'),
      'encode-greek': () => deps.reopenWithEncoding('ISO-8859-7'),
      'encode-hebrew': () => deps.reopenWithEncoding('ISO-8859-8'),
      'encode-korean': () => deps.reopenWithEncoding('EUC-KR'),
      'encode-thai': () => deps.reopenWithEncoding('TIS-620'),
      'encode-turkish': () => deps.reopenWithEncoding('ISO-8859-9'),
      'encode-vietnamese': () => deps.reopenWithEncoding('Windows-1258'),
      'encode-we': () => deps.reopenWithEncoding('ISO-8859-1'),
      'rename-file': deps.renameCurrentFile,
      'new-window': () => ElMessage.info('多窗口暂不支持，请启动新实例'),
      'open-view': deps.manageFavorites,
      'manage-fav': deps.manageFavorites,
      'fav-empty': () => ElMessage.info('收藏夹为空'),
      'copy-line': () => execEd('line-dup'),
      'cut-line': () => execEd('line-del'),
      'clear-all-marks': () => execEd('clear-all-marks'),
      // 🆕 V2.0.0
      'regex-tester': () => { deps.showRegexTester.value = true },
      'snippet-panel': () => { deps.showSnippetPanel.value = !deps.showSnippetPanel.value },
      'bookmark-panel': () => { deps.showBookmarkPanel.value = !deps.showBookmarkPanel.value },
      'function-list': () => { deps.showFunctionList.value = !deps.showFunctionList.value },
      'file-monitor': () => { deps.showFileMonitor.value = !deps.showFileMonitor.value },
      'json-path': () => {
        useFormatConverterStore().set('jsonpath')
        deps.showFormatConverter.value = true
      },
      'json-to-struct': () => {
        useFormatConverterStore().set('json-to-struct')
        deps.showFormatConverter.value = true
      },
      'json-diff': () => {
        useFormatConverterStore().set('json-diff')
        deps.showFormatConverter.value = true
      },
      'search-files': () => { deps.findMode.value = 'global'; deps.showFindWin.value = true },
      'manage-drafts': deps.manageDrafts,
      'toggle-auto-theme': deps.toggleAutoTheme,
      'recent-files': deps.showRecentFiles,
      'recent-folders': deps.showRecentFolders,
      // 「最近打开的文件」子菜单点击：cmd='recent-open', 第一个参数为文件路径
      'recent-open': (path: string) => {
        if (typeof path === 'string' && path) deps.openFilePath(path)
      },
      'recent-empty': () => ElMessage.info('暂无最近文件'),
      // 修复原 useCommands 'clear-history' 误命名（之前只清收藏夹）—— 现清空最近文件
      'clear-history': async () => {
        try {
          const { ClearRecentFiles } = await import('../../wailsjs/go/main/App')
          await ClearRecentFiles()
          ElMessage.success('最近文件已清空')
          document.dispatchEvent(new CustomEvent('recent-updated'))
        } catch (e: any) {
          ElMessage.error('清空失败：' + (e?.message || ''))
        }
      },
      // 同时提供清空收藏夹的动作，菜单区分清楚
      'clear-favorites': () => {
        if (ss.config) { ss.config.ui.favorites = []; ss.saveConfig(); ElMessage.success('收藏夹已清空') }
      },
      // 🆕 V2.0.0 第四阶段
      'script-manager': () => { deps.showScriptManager.value = true },
      'image-editor': deps.showImageEditorView,
      'color-picker': () => { deps.showColorPicker.value = true },
      'log-mode': deps.toggleLogMode,
    }

    const fn = m[name]
    if (fn) {
      if (fn.length >= 1 && args.length > 0) fn(...args)
      else fn()
      return
    }
    // 委托给编辑器处理的命令 (case-*, trim-*, line-*, sort-*, tab2space, space2tab-*, mark-*, clear-mark, show-*, iconsize-*, conv-*, encode-*)
    execEd(name, ...args)
  }

  async function openExplorer() {
    const t = ed.activeTab
    if (!t?.path) return
    try {
      const { BrowserOpenURL } = await import('../../wailsjs/runtime/runtime')
      BrowserOpenURL('file:///' + t.path.replace(/\\/g, '/').replace(/\/[^\/]+$/, ''))
    } catch { /* ignore */ }
  }

  return { onMenuCmd }
}
