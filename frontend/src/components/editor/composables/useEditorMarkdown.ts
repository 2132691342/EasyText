/**
 * useEditorMarkdown — 编辑器 Markdown 模式 / 预览 / 导出
 *
 * 职责（与原 CodeEditor.vue 完全等价搬运，无行为变更）：
 *  - isMarkdown / mdMode / toggleMdMode
 *  - 初始化 mermaid（initMermaid）——根据主题色板决定 dark/default
 *  - highlightCode（hljs） + KaTeX 渲染（renderMath）
 *  - renderedHtml（md.render + renderMath）
 *  - exportMdHtml（导出独立 HTML 文件） + handlePreviewClick（code-block 复制）
 *
 * 设计目标：纯状态 + 工具函数集合，模板仍留在容器（toolbar / preview div）。
 * mermaid / hljs / katex / markdown-it 都是 module-level 引用，不持有实例状态。
 */
import { computed, ref, type Ref } from 'vue'
import MarkdownIt from 'markdown-it'
import mermaid from 'mermaid'
import katex from 'katex'
import hljs from 'highlight.js/lib/common'
import 'katex/dist/katex.min.css'
import 'highlight.js/styles/github.css'
import { ElMessage } from 'element-plus'
import { SaveFileDialog, SaveFile } from '../../../../wailsjs/go/main/App'
import type { EditorTab, MdViewMode } from '@/types'
import type { ThemeColors } from './useEditorTheme'

/** mermaid 计数器（保证 DOM id 唯一） */
const mermaidCounter = { n: 0 }

/** 是否已初始化（仅初始化一次） */
let mermaidInitialized = false

const md = new MarkdownIt({ html: false, linkify: true, typographer: true })

export interface UseEditorMarkdown {
  isMarkdown: Ref<boolean>
  mdMode: Ref<MdViewMode>
  toggleMdMode: () => void
  /** 由容器在 mount 时调用一次（确保主题稳定后再 init） */
  initMermaid: (isDark: boolean) => void
  renderedHtml: Ref<string>
  exportMdHtml: (tab: EditorTab) => Promise<void>
  handlePreviewClick: (e: MouseEvent) => void
}

/**
 * @param language  当前 tab 的语言标识（用于判断是否是 markdown）
 * @param content   当前 tab 的内容（用于生成预览 HTML）
 * @param colors    主题色（仅用于 mermaid 初始化）
 */
export function useEditorMarkdown(
  language: Ref<string>,
  content: Ref<string>,
  colors: Ref<ThemeColors>,
): UseEditorMarkdown {
  const mdMode = ref<MdViewMode>('split')
  const isMarkdown = computed(() => language.value === 'markdown')

  function toggleMdMode() {
    if (!isMarkdown.value) return
    const modes: MdViewMode[] = ['edit', 'split', 'preview']
    mdMode.value = modes[(modes.indexOf(mdMode.value) + 1) % 3]
  }

  function initMermaid(isDark: boolean) {
    if (mermaidInitialized) return
    mermaid.initialize({
      startOnLoad: false,
      theme: isDark ? 'dark' : 'default',
      securityLevel: 'loose',
      fontFamily: 'inherit',
    })
    mermaidInitialized = true
  }

  /** 代码块语法高亮 */
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

  /** 自定义 fence 规则：mermaid 占位 + 其他代码块 hljs 高亮 */
  md.renderer.rules.fence = (tokens, idx, _options, _env, _self) => {
    const token = tokens[idx]
    const code = token.content
    const encoded = encodeURIComponent(code)
    const lang = token.info.trim()

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

  /** KaTeX 数学公式：$$...$$（block）+ $...$（inline） */
  function renderMath(text: string): string {
    text = text.replace(/\$\$([\s\S]*?)\$\$/g, (_: string, formula: string) => {
      try { return katex.renderToString(formula.trim(), { displayMode: true, throwOnError: false }) }
      catch { return `<pre class="katex-error">${formula}</pre>` }
    })
    text = text.replace(/(?<!\$)\$(?!\$)([^$]+?)\$(?!\$)/g, (_: string, formula: string) => {
      try { return katex.renderToString(formula.trim(), { displayMode: false, throwOnError: false }) }
      catch { return `$${formula}$` }
    })
    return text
  }

  const renderedHtml = computed(() => {
    if (!isMarkdown.value) return ''
    let html = md.render(content.value || '')
    html = renderMath(html)
    return html
  })

  /** 预览区 code-block 复制按钮 */
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

  async function exportMdHtml(tab: EditorTab) {
    const html = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<title>${tab.name || 'markdown'}</title>
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
      const p = await SaveFileDialog((tab.name || 'markdown').replace(/\.md$/i, '') + '.html')
      if (p) { await SaveFile(p, html, 'UTF-8'); ElMessage.success('已导出 HTML') }
    } catch (e: any) {
      ElMessage.error('导出失败：' + (e?.message || ''))
    }
  }

  return {
    isMarkdown,
    mdMode,
    toggleMdMode,
    initMermaid,
    renderedHtml,
    exportMdHtml,
    handlePreviewClick,
  }
}