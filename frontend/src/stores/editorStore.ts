import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { EditorTab, AppConfig, EditorConfig, MacroStep, Macro, Bookmark, RecentEntry, Snippet, BookmarkEntry } from '@/types'
import { generateId, getLanguageFromExtension, getFileExtension, getTabViewType } from '@/utils'

export const useEditorStore = defineStore('editor', () => {
  // ============ 标签页状态 ============
  const tabs = ref<EditorTab[]>([])
  const activeTabId = ref<string | null>(null)
  const config = ref<AppConfig | null>(null)

  // ============ 书签（notepad-- bookmark 系统）============
  const bookmarksMap = ref<Map<string, Set<number>>>(new Map())

  // ============ 位置历史（notepad-- 跳转历史）============
  const positionHistory = ref<Map<string, { pos: number; scrollTop: number }[]>>(new Map())
  const historyIndex = ref<Map<string, number>>(new Map())

  // ============ 剪贴板历史（notepad-- 剪贴板历史记录）============
  const clipboardHistory = ref<string[]>([])
  function pushClipboard(text: string) {
    if (!text) return
    const list = clipboardHistory.value.filter(t => t !== text)
    list.unshift(text)
    if (list.length > 50) list.length = 50
    clipboardHistory.value = list
  }

  // ============ 宏系统（notepad-- 宏录制/回放）============
  const macroState = ref({
    isRecording: false,
    isPlaying: false,
    playingId: '', // 当前回放的宏 ID，原模块级 let currentPlayingMacroId 已迁移至此
    currentMacro: [] as MacroStep[],
    savedMacros: [] as Macro[],
    recordStartTime: 0,
    playIndex: 0,
    loopCount: 1,
    currentLoop: 0,
  })

  // ============ 查找状态 ============
  const findState = ref({
    isVisible: false,
    mode: 'find' as 'find' | 'replace' | 'files' | 'mark',
    options: {
      search: '', replace: '', caseSensitive: false,
      wholeWord: false, useRegex: false, wrapAround: true,
      searchDirection: 'forward' as 'forward' | 'backward',
      scope: 'current' as 'current' | 'all' | 'selection',
    },
    lastSearchTerm: '',
    currentMatchIndex: -1,
    allMatches: [] as any[],
    findInFileResults: [] as any[],
    isSearching: false,
  })

  // ============ 计算属性 ============
  const activeTab = computed(() => tabs.value.find(t => t.id === activeTabId.value) || null)
  const activeTabIndex = computed(() => tabs.value.findIndex(t => t.id === activeTabId.value))
  const dirtyTabs = computed(() => tabs.value.filter(t => t.isDirty))
  const hasUnsavedChanges = computed(() => dirtyTabs.value.length > 0)

  // ============ 标签页操作 ============
  function createTab(path: string, content: string, encoding: string = 'UTF-8', lineEnding: string = 'LF'): EditorTab {
    const name = path.split(/[/\\]/).pop() || 'Untitled'
    const ext = getFileExtension(name)
    const language = getLanguageFromExtension(ext)
    const viewType = getTabViewType(ext)

    const tab: EditorTab = {
      id: generateId(),
      path, name, content, originalContent: content,
      isDirty: false, encoding, lineEnding, language,
      cursorPosition: { line: 1, column: 1 },
      scrollPosition: { top: 0, left: 0 },
      isReadOnly: false, viewType,
    }
    tabs.value.push(tab)
    activeTabId.value = tab.id
    return tab
  }

  function closeTab(tabId: string) {
    const index = tabs.value.findIndex(t => t.id === tabId)
    if (index === -1) return
    bookmarksMap.value.delete(tabId)
    positionHistory.value.delete(tabId)
    historyIndex.value.delete(tabId)
    tabs.value.splice(index, 1)
    if (activeTabId.value === tabId) {
      if (tabs.value.length === 0) activeTabId.value = null
      else if (index >= tabs.value.length) activeTabId.value = tabs.value[tabs.value.length - 1].id
      else activeTabId.value = tabs.value[index].id
    }
  }

  function closeAllTabs() {
    tabs.value = []; activeTabId.value = null
    bookmarksMap.value.clear(); positionHistory.value.clear(); historyIndex.value.clear()
  }

  function closeOtherTabs(tabId: string) {
    const toClose = tabs.value.filter(t => t.id !== tabId)
    for (const t of toClose) {
      bookmarksMap.value.delete(t.id); positionHistory.value.delete(t.id); historyIndex.value.delete(t.id)
    }
    tabs.value = tabs.value.filter(t => t.id === tabId)
    activeTabId.value = tabId
  }

  function activateTab(tabId: string) {
    const tab = tabs.value.find(t => t.id === tabId)
    if (tab) activeTabId.value = tabId
  }

  function updateTabContent(tabId: string, content: string) {
    const tab = tabs.value.find(t => t.id === tabId)
    if (tab) { tab.content = content; tab.isDirty = content !== tab.originalContent }
  }

  function markTabSaved(tabId: string) {
    const tab = tabs.value.find(t => t.id === tabId)
    if (tab) { tab.isDirty = false; tab.originalContent = tab.content }
  }

  function updateCursorPosition(tabId: string, line: number, column: number) {
    const tab = tabs.value.find(t => t.id === tabId)
    if (tab) tab.cursorPosition = { line, column }
  }

  function updateScrollPosition(tabId: string, top: number, left: number) {
    const tab = tabs.value.find(t => t.id === tabId)
    if (tab) tab.scrollPosition = { top, left }
  }

  function updateTabEncoding(tabId: string, encoding: string) {
    const tab = tabs.value.find(t => t.id === tabId)
    if (tab) tab.encoding = encoding
  }

  function updateTabLineEnding(tabId: string, lineEnding: string) {
    const tab = tabs.value.find(t => t.id === tabId)
    if (tab) tab.lineEnding = lineEnding
  }

  function renameTab(tabId: string, newPath: string) {
    const tab = tabs.value.find(t => t.id === tabId)
    if (tab) {
      tab.path = newPath; tab.name = newPath.split(/[/\\]/).pop() || 'Untitled'
      tab.language = getLanguageFromExtension(getFileExtension(tab.name))
    }
  }

  function getTabByPath(path: string): EditorTab | undefined {
    return tabs.value.find(t => t.path === path)
  }

  // ============ 书签操作 ============
  function toggleBookmark(tabId: string, line: number) {
    if (!bookmarksMap.value.has(tabId)) bookmarksMap.value.set(tabId, new Set())
    const set = bookmarksMap.value.get(tabId)!
    if (set.has(line)) set.delete(line)
    else set.add(line)
    bookmarksMap.value = new Map(bookmarksMap.value)
  }

  function hasBookmark(tabId: string, line: number): boolean {
    return bookmarksMap.value.get(tabId)?.has(line) || false
  }

  function getBookmarks(tabId: string): number[] {
    return Array.from(bookmarksMap.value.get(tabId) || []).sort((a, b) => a - b)
  }

  function nextBookmark(tabId: string, currentLine: number): number | null {
    const list = getBookmarks(tabId)
    for (const line of list) if (line > currentLine) return line
    return list[0] ?? null
  }

  function prevBookmark(tabId: string, currentLine: number): number | null {
    const list = getBookmarks(tabId)
    for (let i = list.length - 1; i >= 0; i--) if (list[i] < currentLine) return list[i]
    return list[list.length - 1] ?? null
  }

  function clearBookmarks(tabId: string) {
    bookmarksMap.value.delete(tabId)
    bookmarksMap.value = new Map(bookmarksMap.value)
  }

  // ============ 位置历史 ============
  function pushPosition(tabId: string, pos: number, scrollTop: number = 0) {
    if (!positionHistory.value.has(tabId)) {
      positionHistory.value.set(tabId, [])
      historyIndex.value.set(tabId, -1)
    }
    const list = positionHistory.value.get(tabId)!
    const idx = historyIndex.value.get(tabId) ?? -1
    list.splice(idx + 1)
    list.push({ pos, scrollTop })
    if (list.length > 50) list.shift()
    historyIndex.value.set(tabId, list.length - 1)
    positionHistory.value = new Map(positionHistory.value)
  }

  function goBackPosition(tabId: string): { pos: number; scrollTop: number } | null {
    if (!positionHistory.value.has(tabId)) return null
    const list = positionHistory.value.get(tabId)!
    const idx = historyIndex.value.get(tabId) ?? -1
    if (idx <= 0) return null
    historyIndex.value.set(tabId, idx - 1)
    return list[idx - 1]
  }

  function goForwardPosition(tabId: string): { pos: number; scrollTop: number } | null {
    if (!positionHistory.value.has(tabId)) return null
    const list = positionHistory.value.get(tabId)!
    const idx = historyIndex.value.get(tabId) ?? -1
    if (idx >= list.length - 1) return null
    historyIndex.value.set(tabId, idx + 1)
    return list[idx + 1]
  }

  // ============ 宏操作 ============
  function startMacroRecording() {
    macroState.value.currentMacro = []
    macroState.value.isRecording = true
    macroState.value.recordStartTime = Date.now()
  }

  function stopMacroRecording() {
    macroState.value.isRecording = false
    if (macroState.value.currentMacro.length > 0) {
      const macro: Macro = {
        id: generateId(),
        name: `Macro ${macroState.value.savedMacros.length + 1}`,
        steps: [...macroState.value.currentMacro],
        createdAt: macroState.value.recordStartTime,
        modifiedAt: Date.now(),
        tabId: activeTabId.value || '',
      }
      macroState.value.savedMacros.push(macro)
    }
  }

  function recordMacroStep(step: MacroStep) {
    if (macroState.value.isRecording) {
      macroState.value.currentMacro.push(step)
    }
  }

  function playMacro(macroId: string, loopCount: number = 1) {
    const macro = macroState.value.savedMacros.find(m => m.id === macroId)
    if (!macro) return
    macroState.value.playingId = macroId
    macroState.value.isPlaying = true
    macroState.value.playIndex = 0
    macroState.value.loopCount = loopCount
    macroState.value.currentLoop = 0
  }

  function getNextMacroStep(): MacroStep | null {
    if (!macroState.value.isPlaying) return null
    const macro = macroState.value.savedMacros.find(m => m.id === macroState.value.playingId)
    if (!macro) return null
    const steps = macro.steps
    if (macroState.value.playIndex >= steps.length) {
      macroState.value.currentLoop++
      if (macroState.value.currentLoop >= macroState.value.loopCount) {
        macroState.value.isPlaying = false
        return null
      }
      macroState.value.playIndex = 0
    }
    return steps[macroState.value.playIndex++]
  }

  function stopMacroPlayback() {
    macroState.value.isPlaying = false
  }

  function deleteMacro(macroId: string) {
    macroState.value.savedMacros = macroState.value.savedMacros.filter(m => m.id !== macroId)
  }

  function renameMacro(macroId: string, newName: string) {
    const macro = macroState.value.savedMacros.find(m => m.id === macroId)
    if (macro) macro.name = newName
  }

  function saveMacros() {
    try {
      localStorage.setItem('easytext-macros', JSON.stringify(macroState.value.savedMacros))
    } catch (e) { console.warn(e) }
  }

  function loadMacros() {
    try {
      const data = localStorage.getItem('easytext-macros')
      if (data) macroState.value.savedMacros = JSON.parse(data)
    } catch (e) { console.warn(e) }
  }

  // ============ 🆕 V2.0.0 最近访问 ============
  const recentFiles = ref<RecentEntry[]>([])
  const recentFolders = ref<RecentEntry[]>([])

  // ============ 🆕 V2.0.0 代码片段 ============
  const snippets = ref<Snippet[]>([])

  // ============ 🆕 V2.0.0 持久化书签 ============
  const globalBookmarks = ref<BookmarkEntry[]>([])

  // ============ 配置 ============
  function setConfig(newConfig: AppConfig) {
    config.value = newConfig
  }

  function updateEditorConfig(editorConfig: EditorConfig) {
    if (config.value) config.value.editor = editorConfig
  }

  return {
    // State
    tabs, activeTabId, config, bookmarks: bookmarksMap, positionHistory, macroState, findState, clipboardHistory,
    // Computed
    activeTab, activeTabIndex, dirtyTabs, hasUnsavedChanges,
    // Tab actions
    createTab, closeTab, closeAllTabs, closeOtherTabs, activateTab,
    updateTabContent, markTabSaved, updateCursorPosition, updateScrollPosition,
    updateTabEncoding, updateTabLineEnding, renameTab, getTabByPath, setConfig, updateEditorConfig,
    // Bookmarks
    toggleBookmark, hasBookmark, getBookmarks, nextBookmark, prevBookmark, clearBookmarks,
    // Position history
    pushPosition, goBackPosition, goForwardPosition,
    // Clipboard history
    pushClipboard,
    // Macro
    startMacroRecording, stopMacroRecording, recordMacroStep,
    playMacro, getNextMacroStep, stopMacroPlayback,
    deleteMacro, renameMacro, saveMacros, loadMacros,
    // 🆕 V2.0.0
    recentFiles, recentFolders, snippets, globalBookmarks,
  }
})
