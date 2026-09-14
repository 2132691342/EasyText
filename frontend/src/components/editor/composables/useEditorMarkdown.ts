/**
 * useEditorMarkdown — 编辑器 Markdown 模式 / 预览 / 导出
 *
 * 职责：
 *  - isMarkdown / mdMode / toggleMdMode
 *  - initMermaid（主题跟随深浅色，切换时重新 initialize）
 *  - renderMermaidDiagrams（把预览里的 .mermaid-chart 渲染成图；幂等 + 失败回退源码）
 *  - highlightCode（hljs，带结果缓存） + KaTeX 渲染（renderMath，仅文本节点、跳过代码块）
 *  - renderedHtml（md.render + DOMPurify 消毒 + renderMath；导出时读）
 *  - previewHtml（renderedHtml 的防抖快照：预览不可见不渲染，连续输入只渲染一次）
 *  - exportMdHtml（导出独立 HTML 文件）
 *  - handlePreviewClick（code-block 复制 + 链接交系统浏览器打开，避免 WebView 被顶掉）
 *
 * 设计目标：纯状态 + 工具函数集合，模板仍留在容器（toolbar / preview div）。
 * mermaid / hljs / katex / markdown-it / DOMPurify 都是 module-level 引用，不持有实例状态。
 */
import { computed, onScopeDispose, ref, watch, type Ref } from 'vue'
import MarkdownIt from 'markdown-it'
import DOMPurify from 'dompurify'
import mermaid from 'mermaid'
import katex from 'katex'
import hljs from 'highlight.js/lib/common'
import 'katex/dist/katex.min.css'
import 'highlight.js/styles/github.css'
// 预览排版骨架（依赖里一直有，但此前没被引入 → 预览里的标题/表格全是默认样式）
import 'github-markdown-css/github-markdown-light.css'
// 令牌化覆盖 + 代码块头部/复制按钮样式（必须排在库样式之后）
import '../markdown-preview.css'
import { ElMessage } from 'element-plus'
import { SaveFileDialog, SaveFile } from '../../../../wailsjs/go/main/App'
import { BrowserOpenURL } from '../../../../wailsjs/runtime/runtime'
import type { EditorTab, MdViewMode } from '@/types'
import type { ThemeColors } from './useEditorTheme'

/** mermaid 计数器（保证 DOM id 唯一） */
const mermaidCounter = { n: 0 }

/** mermaid 已初始化的主题（null = 未初始化）；深浅色切换需要重新 initialize */
let mermaidTheme: 'dark' | 'default' | null = null

/** mermaid 渲染轮次：内容变化开启新一轮，旧轮次不再写回 DOM */
let mermaidRunToken = 0

/** 预览 HTML 的防抖窗口：编辑更新是逐键同步的，预览全量重渲必须合并 */
const PREVIEW_DEBOUNCE_MS = 200

/**
 * hljs 高亮结果缓存：键 = 语言 + 源码，值 = 高亮后的 HTML。
 * 实测 hljs 占整篇渲染成本的约 3/4，而编辑时绝大多数代码块内容不变，命中即省。
 * 用缓存总字符数（而非条数）做上限——单个代码块可能很大，按条数限制控不住内存。
 */
const HIGHLIGHT_CACHE_MAX_CHARS = 2 * 1024 * 1024
const highlightCache = new Map<string, string>()
let highlightCacheChars = 0

// html: true —— 从网页/导出 HTML 粘贴而来的 md 大量内嵌 <p>/<table>/<code> 等标签，
// 关闭时会被转义成纯文本（预览里满屏标签）；安全由 renderedHtml 中的 DOMPurify 消毒兜底。
const md = new MarkdownIt({ html: true, linkify: true, typographer: true })

export interface UseEditorMarkdown {
  isMarkdown: Ref<boolean>
  mdMode: Ref<MdViewMode>
  toggleMdMode: () => void
  /** 由容器在 mount / 主题切换时调用；主题未变化则跳过 */
  initMermaid: (isDark: boolean) => void
  /**
   * 渲染预览区内的 mermaid 图表；容器在 v-html 更新 / 切到预览 / 主题切换后调用。
   * force=true 时重绘已渲染节点（深浅色切换后旧 SVG 的主题已过时）。
   */
  renderMermaidDiagrams: (root: HTMLElement | null, force?: boolean) => Promise<void>
  /** 预览用的 HTML（renderedHtml 的防抖快照；预览不可见时不更新） */
  previewHtml: Ref<string>
  /** 立即值（不做防抖），供导出等一次性场景使用 */
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
  /** 预览是否可见：edit 模式或非 markdown 都不需要渲染预览 */
  const isPreviewVisible = computed(() => isMarkdown.value && mdMode.value !== 'edit')

  function toggleMdMode() {
    if (!isMarkdown.value) return
    const modes: MdViewMode[] = ['edit', 'split', 'preview']
    mdMode.value = modes[(modes.indexOf(mdMode.value) + 1) % 3]
  }

  function initMermaid(isDark: boolean) {
    const theme: 'dark' | 'default' = isDark ? 'dark' : 'default'
    if (mermaidTheme === theme) return
    mermaid.initialize({
      startOnLoad: false,
      theme,
      securityLevel: 'loose',
      fontFamily: 'inherit',
    })
    mermaidTheme = theme
  }

  /**
   * 渲染预览区内的 mermaid 图表（容器在 DOM 更新后调用）。
   * 幂等：已渲染节点带 data-processed；单张图失败只回退该图源码，不影响其他图。
   */
  async function renderMermaidDiagrams(root: HTMLElement | null, force = false) {
    if (!root) return
    if (force) root.querySelectorAll('.mermaid-chart[data-processed]').forEach(el => el.removeAttribute('data-processed'))
    const nodes = Array.from(root.querySelectorAll<HTMLElement>('.mermaid-chart:not([data-processed])'))
    if (!nodes.length) return
    const token = ++mermaidRunToken
    for (const el of nodes) {
      const code = el.textContent ?? ''
      el.setAttribute('data-processed', '1')
      try {
        const { svg, bindFunctions } = await mermaid.render(`mermaid-svg-${mermaidCounter.n++}`, code)
        // 已有更新一轮渲染（或节点已随 v-html 重建）→ 丢弃本轮结果
        if (token !== mermaidRunToken || !el.isConnected) return
        el.innerHTML = svg
        bindFunctions?.(el)
      } catch (err) {
        if (token !== mermaidRunToken || !el.isConnected) return
        el.classList.add('mermaid-error')
        el.innerHTML = `<pre class="text-xs text-red-500 whitespace-pre-wrap m-0">${md.utils.escapeHtml(code)}</pre>`
        console.warn('[mermaid] 渲染失败', err)
      }
    }
  }

  /** 代码块语法高亮（带结果缓存） */
  function highlightCode(code: string, lang: string): string {
    if (!code.trim()) return ''
    const langName = lang.trim().split(/\s+/)[0]
    const key = `${langName}\u0000${code}`
    const cached = highlightCache.get(key)
    if (cached !== undefined) return cached

    let html: string
    try {
      html = langName && hljs.getLanguage(langName)
        ? hljs.highlight(code, { language: langName, ignoreIllegals: true }).value
        : hljs.highlightAuto(code).value
    } catch {
      html = md.utils.escapeHtml(code)
    }
    // FIFO 淘汰：Map 保持插入顺序，超预算时丢最早写入的条目
    highlightCache.set(key, html)
    highlightCacheChars += key.length + html.length
    while (highlightCacheChars > HIGHLIGHT_CACHE_MAX_CHARS && highlightCache.size > 1) {
      const oldest = highlightCache.keys().next().value
      if (oldest === undefined) break
      highlightCacheChars -= oldest.length + (highlightCache.get(oldest)?.length ?? 0)
      highlightCache.delete(oldest)
    }
    return html
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

  /**
   * KaTeX 数学公式：$$...$$（块级）+ $...$（行内）。
   * 只在「文本节点」上替换，且跳过 <code>/<pre>：
   * 直接在整段 HTML 上跑正则会跨标签配对（<code>$</code> 与很远处另一个 $ 之间整段正文被塞进公式），
   * 代码块/行内代码里的 $（shell 变量、$ARGUMENTS 这类）更不能当公式。
   */
  function renderMath(html: string): string {
    const root = document.createElement('div')
    root.innerHTML = html
    const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT)
    const targets: Text[] = []
    for (let node = walker.nextNode(); node; node = walker.nextNode()) {
      const text = node as Text
      if (!text.data.includes('$')) continue
      const parent = text.parentElement
      if (parent && parent.closest('code, pre, .katex')) continue
      targets.push(text)
    }
    for (const text of targets) {
      const rendered = renderMathText(text.data)
      if (rendered === text.data) continue
      const holder = document.createElement('span')
      holder.innerHTML = rendered
      text.replaceWith(...Array.from(holder.childNodes))
    }
    return root.innerHTML
  }

  /** 单个文本节点内的公式替换；行内 $ 要求首尾不接空白，避免误伤 "$5"、"变量 $HOME" 这类正文 */
  function renderMathText(text: string): string {
    text = text.replace(/\$\$([\s\S]*?)\$\$/g, (_: string, formula: string) => {
      try { return katex.renderToString(formula.trim(), { displayMode: true, throwOnError: false }) }
      catch { return `<pre class="katex-error">${formula}</pre>` }
    })
    text = text.replace(/(?<!\$)\$(?!\s)([^$\n]+?)(?<!\s)\$(?!\$)/g, (_: string, formula: string) => {
      try { return katex.renderToString(formula.trim(), { displayMode: false, throwOnError: false }) }
      catch { return `$${formula}$` }
    })
    return text
  }

  const renderedHtml = computed(() => {
    if (!isMarkdown.value) return ''
    // 顺序固定：渲染（含内嵌 HTML）→ 消毒（DOMPurify 默认放行 data-*，复制按钮的 data-code 不受影响）
    // → 公式（KaTeX 产物由本地生成，放在消毒之后注入，避免被 DOMPurify 改写）。
    // 预览运行在 Wails WebView 内，未消毒的 onerror/script 可直接触达本地桥，必须消毒。
    const safe = DOMPurify.sanitize(md.render(content.value || ''))
    return renderMath(safe)
  })

  /**
   * 预览 HTML 的防抖快照。
   * 编辑器是逐键把内容同步进 store 的，若预览直接绑 renderedHtml，每个按键都要全量重渲
   * （md.render + hljs 高亮所有代码块 + DOMPurify + KaTeX），大文档必卡。
   * 因此：内容变化走防抖；切 tab / 切模式立即出结果；预览不可见时一次都不渲染。
   */
  const previewHtml = ref('')
  let previewTimer: ReturnType<typeof setTimeout> | null = null

  function schedulePreviewRender(delay = PREVIEW_DEBOUNCE_MS) {
    if (previewTimer) clearTimeout(previewTimer)
    previewTimer = setTimeout(() => {
      previewTimer = null
      if (!isPreviewVisible.value) return
      previewHtml.value = renderedHtml.value
    }, delay)
  }

  watch(content, () => schedulePreviewRender(), { immediate: true })
  watch([language, mdMode], () => schedulePreviewRender(0), { immediate: true })
  onScopeDispose(() => { if (previewTimer) clearTimeout(previewTimer) })

  /** 预览区点击：code-block 复制按钮 + 链接外开 */
  function handlePreviewClick(e: MouseEvent) {
    const target = e.target as HTMLElement

    const btn = target.closest('.code-copy-btn') as HTMLElement | null
    if (btn) {
      const encoded = btn.getAttribute('data-code')
      if (!encoded) return
      try {
        navigator.clipboard.writeText(decodeURIComponent(encoded)).then(() => {
          btn.classList.add('copied')
          const s = btn.querySelector('.code-copy-text')
          if (s) s.textContent = '已复制'
          setTimeout(() => { btn.classList.remove('copied'); if (s) s.textContent = '复制' }, 2000)
        })
      } catch (err) { console.warn(err) }
      return
    }

    // 链接不能在 WebView 内直接跳转（会顶掉整个应用界面）：http(s)/mailto 交系统浏览器打开；
    // 页内锚点保留默认行为；相对路径依赖本地文件位置，在 WebView 里无处可去，直接忽略。
    const link = target.closest('a[href]') as HTMLAnchorElement | null
    if (!link) return
    const href = link.getAttribute('href') || ''
    if (href.startsWith('#')) return
    e.preventDefault()
    if (/^(https?:|mailto:)/i.test(href)) BrowserOpenURL(link.href)
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
    renderMermaidDiagrams,
    previewHtml,
    renderedHtml,
    exportMdHtml,
    handlePreviewClick,
  }
}