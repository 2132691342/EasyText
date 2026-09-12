/**
 * useEditorCompletion — 编辑器 snippet 自动完成（修闭环 bug #2）
 *
 * 职责：
 *  - snippet 缓存 + 懒加载（首次需要时才请求后端）
 *  - snippetCompletionSource（CodeMirror autocompletion.override 入口）
 *  - invalidateCompletionCache()：暴露给外部，主动失效缓存
 *
 * 关键修复（bug #2）：
 *  - 原 CodeEditor 一旦首次加载过 snippets，即使 SnippetPanel 创建了新 snippet，
 *    编辑器自动完成也不会刷新（snippetsLoaded 永久为 true）。
 *  - 修复：暴露 invalidateCompletionCache()；
 *    useEditorView 监听 editorStore.snippets 长度变化时自动调用。
 *
 * 注意：本模块持有可变内部状态（缓存 + loaded flag），
 * 所以**每个编辑器实例（tab）都需要独立调用一次 useEditorCompletion**。
 */
import { ref, watch, type Ref } from 'vue'
import { GetSnippets } from '../../../../wailsjs/go/main/App'
import { snippetCompletion } from '@codemirror/autocomplete'
import type { Snippet } from '@/types'

export interface UseEditorCompletion {
  /** 提供给 EditorState.autocompletion.override */
  snippetCompletionSource: (context: any) => any | null
  /**
   * 主动失效缓存：外部（容器 / useEditorView）应在以下时机调用：
   *  1. 创建/修改/删除 snippet（editorStore.snippets 变化）
   *  2. 切到不同语言后首次触发补全时
   *  3. 重新打开编辑器实例时
   */
  invalidateCompletionCache: () => void
  /** 当前缓存的 snippet 数量（调试用） */
  cacheSize: () => number
}

/**
 * @param language 当前 tab 的语言标识（GetSnippets 第一个参数）
 * @param externalSnippets 可选：来自 editorStore 的 snippet 列表引用；
 *   当其 length 变化时，自动 invalidate 缓存——这是 bug #2 的核心修复。
 */
export function useEditorCompletion(
  language: Ref<string>,
  externalSnippets?: Ref<Snippet[]>,
): UseEditorCompletion {
  // mutable 缓存（module-level 概念搬到 instance，避免跨实例污染）
  let snippetsCache: Snippet[] = []
  let snippetsLoaded = false

  async function loadSnippetsForCompletion() {
    if (snippetsLoaded) return
    try {
      snippetsCache = await GetSnippets(language.value || '')
      snippetsLoaded = true
    } catch {
      snippetsCache = []
    }
  }

  function invalidateCompletionCache() {
    snippetsLoaded = false
    snippetsCache = []
  }

  function snippetCompletionSource(context: any) {
    const word = context.matchBefore(/\w+/)
    if (!word || snippetsCache.length === 0) return null

    // 每次打开补全时尝试加载（首次加载完成后 cache 生效）
    loadSnippetsForCompletion()

    const options: any[] = []
    for (const snippet of snippetsCache) {
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

  // bug #2 修复：监听外部 snippet 列表长度变化，自动 invalidate
  if (externalSnippets) {
    watch(
      () => externalSnippets.value.length,
      () => invalidateCompletionCache(),
    )
  }

  return {
    snippetCompletionSource,
    invalidateCompletionCache,
    cacheSize: () => snippetsCache.length,
  }
}