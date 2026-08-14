/**
 * 文件 I/O 与会话管理 composable。
 *
 * 抽离自 MainLayout.vue，集中以下职责：
 *   - 文件 open / save / saveAll / reopen / 重载
 *   - 目录打开
 *   - 会话 save/restore
 *   - 工作空间 save/open
 *
 * 设计要点：
 *   - 不持有任何响应式状态（pure functions + closures），
 *     避免与 setup() 反复初始化带来的隐式状态分裂。
 *   - 错误处理：用户可感知的失败统一抛给 Element Plus ElMessage 提示；
 *     兜底路径（如 AddRecentEntry、DeleteDraft 失败）静默记录 console.warn。
 *   - session 持久化用 .catch(() => {}) 容错——会话保存失败不应阻塞用户操作。
 */

import { ElMessage } from 'element-plus'
import {
  OpenFileDialog,
  OpenDirectoryDialog,
  ReadFile,
  SaveFile,
  SaveFileDialog,
  DeleteDraft,
  GetSession,
  SaveSession,
  AddRecentEntry,
  CheckDraftConflict,
  ShowConfirmDialog,
  GetDraft,
} from '../../wailsjs/go/main/App'
import type { useEditorStore } from '@/stores/editorStore'
import type { useFileStore } from '@/stores/fileStore'
import { getFileExtension, getTabViewType } from '@/utils'

type Ed = ReturnType<typeof useEditorStore>
type Fs = ReturnType<typeof useFileStore>

export interface UseFileOpsOptions {
  ed: Ed
  fs: Fs
  /** UI 状态 ref：触发侧边栏显示 */
  showFileTree: { value: boolean }
  /** 持久化最后打开的目录 */
  setLastFolder: (path: string) => void
}

export function useFileOps(opts: UseFileOpsOptions) {
  const { ed, fs } = opts

  // ============ 创建/打开 ============

  function newFile(): void {
    ed.createTab('', '', 'UTF-8', 'LF')
    saveSession()
  }

  async function openFile(): Promise<void> {
    try {
      const p = await OpenFileDialog()
      if (p) await openFilePath(p)
    } catch {
      ElMessage.error('无法打开文件')
    }
  }

  async function openFilePath(path: string): Promise<void> {
    try {
      const ext = getFileExtension(path)
      const vt = getTabViewType(ext)
      if (vt !== 'code') {
        ed.createTab(path, '', 'binary', 'LF')
      } else {
        const r = await ReadFile(path)
        if (r) {
          // 草稿冲突处理：用户选择恢复则用草稿内容
          const conflict = await CheckDraftConflict(path).catch(() => 0)
          if (conflict === 1) {
            const ok = await ShowConfirmDialog(
              '发现草稿',
              `"${path.split(/[/\\]/).pop()}" 有未保存的草稿，是否恢复草稿内容？`,
            )
            if (ok) {
              const draft = await GetDraft(path)
              if (draft) {
                ed.createTab(path, draft.content, draft.encoding, draft.lineEnding)
                if (ed.activeTab) ed.activeTab.isDirty = true
                saveSession()
                await AddRecentEntry(path, false).catch(() => {})
                document.dispatchEvent(new CustomEvent('recent-updated'))
                return
              }
            }
          }
          ed.createTab(path, r.content, r.info.encoding, r.info.lineEnding)
        }
      }
      await AddRecentEntry(path, false).catch(() => {})
      document.dispatchEvent(new CustomEvent('recent-updated'))
      saveSession()
    } catch (e: unknown) {
      // 大文件（>100MB）回退到 Hex 分页视图
      const code = (e as { code?: number; data?: { code?: number } })?.code
        ?? (e as { data?: { code?: number } })?.data?.code
      if (code === 1003) {
        ed.createTab(path, '', 'binary', 'LF')
        if (ed.activeTab) ed.activeTab.viewType = 'hex'
        ElMessage.warning('文件过大，已以二进制分页模式打开')
        saveSession()
        await AddRecentEntry(path, false).catch(() => {})
      } else {
        ElMessage.error('无法打开: ' + path)
      }
    }
  }

  async function openDir(): Promise<void> {
    try {
      const p = await OpenDirectoryDialog()
      if (p) {
        fs.setDirectory(p)
        opts.showFileTree.value = true
        opts.setLastFolder(p)
      }
    } catch (e) {
      console.warn(e)
    }
  }

  // ============ 保存 ============

  async function saveFile(): Promise<void> {
    const t = ed.activeTab
    if (!t) return
    try {
      if (!t.path) {
        await saveFileAs()
        return
      }
      await SaveFile(t.path, t.content, t.encoding)
      ed.markTabSaved(t.id)
      // 保存成功后清理草稿
      try { await DeleteDraft(t.path) } catch { /* ignore */ }
    } catch {
      ElMessage.error('保存失败')
    }
  }

  async function saveFileAs(): Promise<void> {
    const t = ed.activeTab
    if (!t) return
    try {
      const p = await SaveFileDialog(t.name)
      if (!p) return
      await SaveFile(p, t.content, t.encoding)
      ed.renameTab(t.id, p)
      ed.markTabSaved(t.id)
      saveSession()
    } catch {
      ElMessage.error('另存为失败')
    }
  }

  async function saveAll(): Promise<void> {
    for (const t of ed.dirtyTabs) {
      if (t.path) {
        try {
          await SaveFile(t.path, t.content, t.encoding)
          ed.markTabSaved(t.id)
        } catch (e) {
          console.warn(e)
        }
      }
    }
  }

  async function reloadFile(): Promise<void> {
    const t = ed.activeTab
    if (!t?.path) return
    try {
      const r = await ReadFile(t.path)
      if (r) ed.updateTabContent(t.id, r.content)
    } catch (e) {
      console.warn(e)
    }
  }

  // ============ 会话 ============

  function saveSession(): void {
    SaveSession(
      ed.tabs.filter(t => t.path).map(t => ({
        path: t.path,
        encoding: t.encoding,
        language: t.language,
      })),
      ed.activeTabId || '',
    ).catch(() => { /* 会话保存失败不应阻塞用户操作 */ })
  }

  async function restoreSession(): Promise<void> {
    try {
      const s = await GetSession()
      if (s?.files?.length) {
        // 并行读取所有会话文件，大幅减少启动恢复时间
        const results = await Promise.allSettled(
          s.files.map((f: { path: string; encoding: string }) =>
            ReadFile(f.path).then(r => ({ file: f, result: r })),
          ),
        )
        // 把恢复成功的文件路径汇总，最后统一 AddRecentEntry，避免顺序竞态
        const restored: string[] = []
        for (const item of results) {
          if (item.status === 'fulfilled') {
            const { file, result: r } = item.value
            if (r) {
              try {
                ed.createTab(file.path, r.content, file.encoding || 'UTF-8', r.info.lineEnding)
                restored.push(file.path)
              } catch (e) {
                console.warn(e)
              }
            }
          }
        }
        // 修复：会话恢复未写入最近文件，导致菜单「最近打开文件」始终空
        for (const p of restored) await AddRecentEntry(p, false).catch(() => {})
        if (restored.length) document.dispatchEvent(new CustomEvent('recent-updated'))
        if (s.activeId) ed.activateTab(s.activeId)
      }
    } catch (e) {
      console.warn(e)
    }
  }

  // ============ 工作空间 ============

  async function saveWorkspace(): Promise<void> {
    try {
      const data = {
        folder: fs.currentDirectory || '',
        files: ed.tabs.filter(t => t.path).map(t => ({
          path: t.path,
          encoding: t.encoding,
          language: t.language,
        })),
        activeId: ed.activeTabId || '',
      }
      const p = await SaveFileDialog('workspace.etws')
      if (!p) return
      await SaveFile(p, JSON.stringify(data, null, 2), 'UTF-8')
      ElMessage.success('工作空间已保存')
    } catch (e: unknown) {
      ElMessage.error('保存工作空间失败: ' + ((e as Error)?.message || String(e)))
    }
  }

  async function openWorkspace(): Promise<void> {
    try {
      const p = await OpenFileDialog()
      if (!p) return
      const r = await ReadFile(p)
      if (!r) return
      const data = JSON.parse(r.content)
      if (data.folder) {
        fs.setDirectory(data.folder)
        opts.showFileTree.value = true
      }
      const opened: string[] = []
      for (const f of (data.files || []) as { path: string; encoding: string }[]) {
        if (ed.getTabByPath(f.path)) continue
        try {
          const fr = await ReadFile(f.path)
          if (fr) {
            ed.createTab(f.path, fr.content, f.encoding || 'UTF-8', fr.info.lineEnding)
            opened.push(f.path)
          }
        } catch {
          /* 文件可能已不存在，跳过 */
        }
      }
      // 工作空间内的文件同样记录到「最近打开文件」
      for (const fp of opened) await AddRecentEntry(fp, false).catch(() => {})
      if (opened.length) document.dispatchEvent(new CustomEvent('recent-updated'))
      if (data.activeId) ed.activateTab(data.activeId)
      ElMessage.success('工作空间已打开')
    } catch (e: unknown) {
      ElMessage.error('打开工作空间失败: ' + ((e as Error)?.message || String(e)))
    }
  }

  return {
    newFile,
    openFile,
    openFilePath,
    openDir,
    saveFile,
    saveFileAs,
    saveAll,
    reloadFile,
    saveSession,
    restoreSession,
    saveWorkspace,
    openWorkspace,
  }
}