/**
 * useEditorBookmark — 编辑器书签（修闭环 bug #1）
 *
 * 职责（与原 CodeEditor.vue 完全等价搬运，无行为变更 + bug 修复）：
 *  - toggleBookmark / gotoNextBookmark / gotoPrevBookmark / clearAllBookmarks
 *  - copyBookmarkLines / cutBookmarkLines / deleteBookmarkLines /
 *    deleteUnbookmarkLines / pasteBookmarkLines
 *  - syncFromDB(tabPath)：从后端拉书签列表并写入 editorStore（bug #1 修复）
 *  - scheduleSyncFromDB(tabPath, delay)：debounced 版本（200ms），用于切换 tab 时
 *
 * 关键修复（bug #1）：
 *  - 原 CodeEditor 仅依赖 editorStore.bookmarksMap（in-memory）。
 *  - 修复：每次编辑器实例 / / tab 切换 / / 文件路径变化时，
 *    从后端 GetBookmarks(tab.path) 拉取真实数据并 merge 到 editorStore。
 *  - 这样 BookmarkPanel 创建的书签会立刻出现在 F2 导航列表里。
 */
import { useEditorStore } from '@/stores'
import type { EditorTab } from '@/types'
import { AddBookmark, RemoveBookmark, GetBookmarks } from '../../../../wailsjs/go/main/App'
import { ElMessage } from 'element-plus'

export interface UseEditorBookmark {
  toggleBookmark: (view: any, tab: EditorTab) => void
  gotoNextBookmark: (view: any, tab: EditorTab, gotoLine: (n: number) => void) => void
  gotoPrevBookmark: (view: any, tab: EditorTab, gotoLine: (n: number) => void) => void
  clearAllBookmarks: (tab: EditorTab) => void
  copyBookmarkLines: (view: any, tab: EditorTab) => void
  cutBookmarkLines: (view: any, tab: EditorTab) => void
  deleteBookmarkLines: (view: any, tab: EditorTab) => void
  deleteUnbookmarkLines: (view: any, tab: EditorTab) => void
  pasteBookmarkLines: (view: any, tab: EditorTab) => Promise<void>
  /** bug #1 修复：从后端拉书签并写入 editorStore */
  syncFromDB: (tab: EditorTab) => Promise<void>
  /** debounced 版本（200ms）——用于频繁的 tab 切换 / path 变化 */
  scheduleSyncFromDB: (tab: EditorTab, delay?: number) => void
}

export function useEditorBookmark(): UseEditorBookmark {
  const editorStore = useEditorStore()

  async function syncFromDB(tab: EditorTab) {
    if (!tab.path) return
    try {
      const entries: any[] = await GetBookmarks(tab.path)
      if (!Array.isArray(entries)) return
      // 清空该 tab 旧书签（in-memory map），然后写入后端真实数据
      editorStore.clearBookmarks(tab.id)
      for (const b of entries as any[]) {
        // 后端 lineNumber → 内存 line（结构与 BookmarkMap 一致：1-based）
        editorStore.toggleBookmark(tab.id, b.lineNumber)
      }
    } catch {
      // 后端拉失败不阻断编辑器；in-memory 已有数据
    }
  }

  // 简易 debounce：高频调用（如连续 tab 切换）只触发最后一次
  let syncTimer: number | null = null
  function scheduleSyncFromDB(tab: EditorTab, delay = 200) {
    if (syncTimer) window.clearTimeout(syncTimer)
    syncTimer = window.setTimeout(() => {
      syncFromDB(tab)
      syncTimer = null
    }, delay)
  }

  function toggleBookmark(view: any, tab: EditorTab) {
    if (!view || !tab) return
    const line = tab.cursorPosition?.line || 1
    const isBooked = editorStore.hasBookmark(tab.id, line)
    if (isBooked) {
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
      editorStore.toggleBookmark(tab.id, line)
      if (tab.path) {
        AddBookmark(tab.path, line, '', '').catch(() => {})
      }
    }
  }

  function gotoNextBookmark(view: any, tab: EditorTab, gotoLine: (n: number) => void) {
    if (!view || !tab) return
    const line = tab.cursorPosition?.line || 1
    const next = editorStore.nextBookmark(tab.id, line)
    if (next !== null) gotoLine(next)
  }

  function gotoPrevBookmark(view: any, tab: EditorTab, gotoLine: (n: number) => void) {
    if (!view || !tab) return
    const line = tab.cursorPosition?.line || 1
    const prev = editorStore.prevBookmark(tab.id, line)
    if (prev !== null) gotoLine(prev)
  }

  function clearAllBookmarks(tab: EditorTab) {
    if (!tab) return
    const bookmarks = editorStore.getBookmarks(tab.id)
    if (tab.path) {
      GetBookmarks(tab.path).then((entries: any) => {
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

  function copyBookmarkLines(view: any, tab: EditorTab) {
    if (!view) return
    const lines = editorStore.getBookmarks(tab.id)
    if (!lines.length) { ElMessage.info('没有书签行'); return }
    const doc = view.state.doc
    const text = lines.map((n: number) => doc.line(n).text).join('\n')
    editorStore.pushClipboard(text)
    navigator.clipboard?.writeText(text).catch(() => {})
    ElMessage.success(`已复制 ${lines.length} 个书签行`)
  }

  function cutBookmarkLines(view: any, tab: EditorTab) {
    if (!view) return
    const lines = editorStore.getBookmarks(tab.id)
    if (!lines.length) { ElMessage.info('没有书签行'); return }
    const doc = view.state.doc
    const text = lines.map((n: number) => doc.line(n).text).join('\n')
    editorStore.pushClipboard(text)
    navigator.clipboard?.writeText(text).catch(() => {})
    const changes = [...lines].sort((a, b) => b - a).map((n: number) => {
      const l = doc.line(n)
      return { from: l.from, to: n < doc.lines ? doc.line(n + 1).from : doc.length }
    })
    view.dispatch({ changes })
    ElMessage.success(`已剪切 ${lines.length} 个书签行`)
  }

  function deleteBookmarkLines(view: any, tab: EditorTab) {
    if (!view) return
    const lines = editorStore.getBookmarks(tab.id)
    if (!lines.length) { ElMessage.info('没有书签行'); return }
    const doc = view.state.doc
    const changes = [...lines].sort((a, b) => b - a).map((n: number) => {
      const l = doc.line(n)
      return { from: l.from, to: n < doc.lines ? doc.line(n + 1).from : doc.length }
    })
    view.dispatch({ changes })
    ElMessage.success(`已删除 ${lines.length} 个书签行`)
  }

  function deleteUnbookmarkLines(view: any, tab: EditorTab) {
    if (!view) return
    const lines = editorStore.getBookmarks(tab.id)
    if (!lines.length) { ElMessage.info('没有书签行'); return }
    const doc = view.state.doc
    const keep = new Set(lines)
    const result: string[] = []
    for (let n = 1; n <= doc.lines; n++) if (keep.has(n)) result.push(doc.line(n).text)
    view.dispatch({ changes: { from: 0, to: doc.length, insert: result.join('\n') } })
    ElMessage.success('已删除非书签行')
  }

  async function pasteBookmarkLines(view: any, tab: EditorTab) {
    if (!view) return
    const lines = editorStore.getBookmarks(tab.id)
    if (!lines.length) { ElMessage.info('没有书签行'); return }
    const doc = view.state.doc
    let clip = editorStore.clipboardHistory[0] || ''
    if (!clip) {
      try { clip = await navigator.clipboard.readText() } catch { clip = '' }
    }
    if (!clip) { ElMessage.info('剪贴板为空'); return }
    const clipLines = clip.split('\n')
    const changes = lines
      .map((n: number, i: number) => ({ line: n, insert: clipLines[i % clipLines.length] }))
      .sort((a, b) => b.line - a.line)
      .map(({ line, insert }) => { const l = doc.line(line); return { from: l.from, to: l.to, insert } })
    view.dispatch({ changes })
    ElMessage.success('已粘贴到书签行')
  }

  return {
    toggleBookmark,
    gotoNextBookmark,
    gotoPrevBookmark,
    clearAllBookmarks,
    copyBookmarkLines,
    cutBookmarkLines,
    deleteBookmarkLines,
    deleteUnbookmarkLines,
    pasteBookmarkLines,
    syncFromDB,
    scheduleSyncFromDB,
  }
}