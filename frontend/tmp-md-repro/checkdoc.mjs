/**
 * 临时诊断：用应用的渲染管线（markdown-it html:true + 自定义 fence + hljs 缓存）
 * 渲染真实文档，检查是否有"该渲染却没渲染"的残留。
 * 运行：node frontend/tmp-md-repro/checkdoc.mjs "C:/Users/LIKX/Desktop/skills/ZCode使用手册.md"
 */
import fs from 'node:fs'
import { createRequire } from 'node:module'

const require = createRequire(new URL('../package.json', import.meta.url))
const MarkdownIt = require('markdown-it')
let hljs = require('highlight.js/lib/common')
hljs = hljs.default ?? hljs

const file = process.argv[2]
const src = fs.readFileSync(file, 'utf8')

const cache = new Map()
function highlightCode(code, lang) {
  const key = `${lang}\u0000${code}`
  const hit = cache.get(key)
  if (hit !== undefined) return hit
  let html
  try {
    html = lang && hljs.getLanguage(lang)
      ? hljs.highlight(code, { language: lang, ignoreIllegals: true }).value
      : hljs.highlightAuto(code).value
  } catch { html = md.utils.escapeHtml(code) }
  cache.set(key, html)
  return html
}

const md = new MarkdownIt({ html: true, linkify: true, typographer: true })
md.renderer.rules.fence = (tokens, idx) => {
  const token = tokens[idx]
  const code = token.content
  const encoded = encodeURIComponent(code)
  const lang = token.info.trim().split(/\s+/)[0]
  const label = lang ? `<span class="code-lang-label">${md.utils.escapeHtml(lang)}</span>` : ''
  const body = lang === 'mermaid' ? md.utils.escapeHtml(code) : highlightCode(code, lang)
  return `<div class="code-block-wrapper"><div class="code-block-header">${label}<button class="code-copy-btn" data-code="${encoded}">复制</button></div><pre><code class="hljs">${body}</code></pre></div>`
}

const out = md.render(src)
const count = (s, re) => (s.match(re) || []).length

// —— 应用里的 KaTeX 步骤（原样复制）——
const katex = require('katex')
function renderMath(text) {
  text = text.replace(/\$\$([\s\S]*?)\$\$/g, (_m, formula) => {
    try { return katex.renderToString(formula.trim(), { displayMode: true, throwOnError: false }) }
    catch { return `<pre class="katex-error">${formula}</pre>` }
  })
  text = text.replace(/(?<!\$)\$(?!\$)([^$]+?)\$(?!\$)/g, (_m, formula) => {
    try { return katex.renderToString(formula.trim(), { displayMode: false, throwOnError: false }) }
    catch { return `$${formula}$` }
  })
  return text
}

const decode = (s) => s.replace(/&quot;/g, '"').replace(/&#39;/g, "'").replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&amp;/g, '&')
const visible = decode(out.replace(/<[^>]*>/g, ''))

console.log('===== 源文件 =====')
console.log(`字符 ${src.length}，行 ${count(src, /\n/g) + 1}`)
const fenceLines = count(src, /^```/gm)
console.log('围栏行数 ' + fenceLines + '  → 代码块约 ' + Math.floor(fenceLines / 2) + ' 个')
console.log(`标题行 ${count(src, /^#{1,6} /gm)}，表格行 ${count(src, /^\|/gm)}，图片 ${count(src, /!\[/g)}，公式 $ ${count(src, /\$/g)}`)

console.log('\n===== 渲染结果 =====')
console.log(`<table> ${count(out, /<table>/g)}，<h1-6> ${count(out, /<h[1-6][ >]/g)}，代码块 wrapper ${count(out, /code-block-wrapper/g)}`)
console.log(`<a href> ${count(out, /<a href/g)}，<img> ${count(out, /<img/g)}，mermaid 容器 ${count(out, /mermaid-chart/g)}`)
console.log(`渲染 HTML 长度 ${out.length}`)

console.log('\n===== 可见文本里的残留（应为 0）=====')
const leftovers = [
  ['未解析的井号标题行', /(?:^|\n)#{1,6} \S/g],
  ['未解析的表格行', /(?:^|\n)\|/g],
  ['未解析的围栏', /```/g],
  ['未解析的加粗', /\*\*[^*\n]+\*\*/g],
  ['未解析的引用块', /(?:^|\n)> \S/g],
  ['表格分隔行 |---|', /\|[\s-]*-{3,}[\s-]*\|/g],
]
for (const [label, re] of leftovers) {
  const m = visible.match(re) || []
  console.log(`${label}: ${m.length}${m.length ? '  样例: ' + JSON.stringify(m.slice(0, 3).join(' / ')) : ''}`)
}

// —— KaTeX 步骤前后对比：$ 被当成公式会不会吃掉正文 ——
const afterMath = renderMath(out)
const visibleAfter = decode(afterMath.replace(/<[^>]*>/g, ''))
const katexCount = count(afterMath, /class="katex/g)
console.log('\n===== KaTeX 步骤 =====')
console.log(`渲染出的 katex 元素：${katexCount} 个（源文件里 $ 共 33 个，多为 shell 变量与 $ARGUMENTS 这类行内代码）`)
let i = 0
while (i < visible.length && visible[i] === visibleAfter[i]) i++
console.log(`可见文本是否被改动：${visible === visibleAfter ? '否' : '是'}（首个差异位置 ${i} / 共 ${visible.length} 字符）`)
if (visible !== visibleAfter) {
  console.log('--- KaTeX 前 ---')
  console.log(JSON.stringify(visible.slice(Math.max(0, i - 60), i + 120)))
  console.log('--- KaTeX 后 ---')
  console.log(JSON.stringify(visibleAfter.slice(Math.max(0, i - 60), i + 120)))
}

// 旧算法（整段 HTML 上跑正则）到底吞了多少正文
let paired = 0
let swallowed = 0
let invalid = 0
out.replace(/(?<!\$)\$(?!\$)([^$]+?)\$(?!\$)/g, (m, f) => {
  paired++
  swallowed += f.length
  if (/[\u4e00-\u9fff]/.test(f)) invalid++
  return m
})
console.log('\n===== 旧算法的损坏规模 =====')
console.log(`跨标签配对的“公式”：${paired} 处，吞掉正文 ${swallowed} 字符，其中含中文的误判 ${invalid} 处`)
console.log(`源文件里的 $ 一共 33 个，真正是数学公式的：0 个`)

console.log('\n===== 可见文本开头 25 行 =====')
console.log(visible.split('\n').slice(0, 25).join('\n'))

const tableIdx = visible.indexOf('添加附件')
console.log('\n===== 表格区域（原文 71-76 行的渲染输出）=====')
console.log(visible.slice(Math.max(0, tableIdx - 200), tableIdx + 260))

const fenceIdx = visible.indexOf('code-block-wrapper') >= 0 ? 0 : 0
console.log('\n===== 代码块区域样例（第一个 code-block）=====')
const first = visible.indexOf('复制')
console.log(visible.slice(Math.max(0, first - 300), first + 400))
