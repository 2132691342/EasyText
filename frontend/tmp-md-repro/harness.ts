/**
 * 临时验证页：真实 useEditorMarkdown 的功能探针 + 真实文档 + 公式边界 + 预览排版样式。
 * 用完即删（frontend/tmp-md-repro/），不进仓库。
 */
import { ref } from 'vue'
import '../src/style.css' // 应用令牌（--et-*）与 Tailwind preflight，保证样式探针与应用一致
import { useEditorMarkdown } from '../src/components/editor/composables/useEditorMarkdown'
import fixture from './fixture.md?raw'
import realDoc from './doc.md?raw'

declare global {
  interface Window {
    __xss?: number
    __opened?: string
    runtime?: any
  }
}

const sleep = (ms: number) => new Promise(r => setTimeout(r, ms))
const opened: string[] = []
window.runtime = { BrowserOpenURL: (url: string) => { opened.push(url); window.__opened = url } }

const checks: Array<[string, boolean]> = []
/** 用一个独立的 composable 实例把 md 渲染成 DOM（挂到页面上，便于读计算样式） */
const renderTo = (md: string) => {
  const instance = useEditorMarkdown(ref('markdown') as any, ref(md) as any, ref<any>({ isDark: false }) as any)
  const el = document.createElement('div')
  el.className = 'markdown-body'
  el.innerHTML = instance.renderedHtml.value
  document.getElementById('mount')!.appendChild(el)
  return el
}
const css = (el: Element | null, prop: string) => (el ? getComputedStyle(el).getPropertyValue(prop).trim() : '')

// ── 1. HTML / 消毒 / 防抖 / 可见性（fixture）─────────────────────────
const language = ref('markdown')
const content = ref(fixture)
const colors = ref<any>({ isDark: false })
const markdown = useEditorMarkdown(language as any, content as any, colors as any)

const out = document.getElementById('out')!
out.addEventListener('click', markdown.handlePreviewClick)
await sleep(80)
out.innerHTML = markdown.previewHtml.value

const literalTags = (out.textContent || '').match(/<\/?[a-zA-Z][^>]*>/g) || []
checks.push(
  ['初始预览已产出（previewHtml 非空）', out.innerHTML.length > 500],
  ['HTML 标题/表格已渲染', !!out.querySelector('h4') && !!out.querySelector('table')],
  ['可见文本无裸露的块级标签', !literalTags.some(t => /^<\/?(h4|p|table|tbody|tr|td|th|ul|li|blockquote|strong|code)\b/i.test(t))],
  ['复制按钮 data-code 未被消毒剥离', !!out.querySelector('.code-copy-btn[data-code]')],
  ['onerror 剥离且脚本未执行', out.querySelector('img[onerror]') === null && window.__xss === undefined],
)

const beforeDebounce = markdown.previewHtml.value
content.value = fixture + '\n\n临时改动 A。\n'
await sleep(40)
checks.push(['内容变化 40ms 内预览未重渲（防抖生效）', markdown.previewHtml.value === beforeDebounce])
await sleep(320)
checks.push(['防抖窗口后预览已更新', markdown.previewHtml.value.includes('临时改动 A')])

markdown.mdMode.value = 'edit'
await sleep(60)
const snapshot = markdown.previewHtml.value
content.value = content.value + '\n\n临时改动 B（edit 模式）。\n'
await sleep(360)
checks.push(['edit 模式下不渲染预览（零成本）', markdown.previewHtml.value === snapshot])
markdown.mdMode.value = 'split'
await sleep(80)
checks.push(['切回分屏立即重渲（含 edit 期间改动）', markdown.previewHtml.value.includes('临时改动 B')])

out.innerHTML = markdown.previewHtml.value
await markdown.initMermaid(false)
await markdown.renderMermaidDiagrams(out)
const svgBefore = out.querySelector('.mermaid-chart svg')
const errorNode = out.querySelector('.mermaid-chart.mermaid-error')
checks.push(
  ['合法 mermaid 图渲染出 svg', !!svgBefore],
  ['非法图回退 mermaid-error 且保留源码', !!errorNode && (errorNode.textContent || '').includes('这不是一个合法的 mermaid 图')],
)
await markdown.renderMermaidDiagrams(out)
checks.push(['重复调用不重绘（幂等）', out.querySelector('.mermaid-chart svg') === svgBefore])

const clickOn = (el: Element) => {
  const ev = new MouseEvent('click', { bubbles: true, cancelable: true })
  return { prevented: !el.dispatchEvent(ev) }
}
checks.push(['外链 preventDefault 且交系统浏览器', clickOn(out.querySelector('a[href="https://example.com"]')!).prevented && opened[0] === 'https://example.com/'])
checks.push(['相对路径拦截且不外开', clickOn(out.querySelector('a[href="./other.md"]')!).prevented && opened.length === 1])
checks.push(['锚点保留默认行为', !clickOn(out.querySelector('a[href^="#"]')!).prevented && opened.length === 1])
clickOn(out.querySelector('.code-copy-btn[data-code]')!)
checks.push(['复制按钮不触发外开', opened.length === 1])

// ── 2. 公式边界：真公式要渲染，$ 变量/货币不能被吞 ────────────────────
const mathEl = renderTo([
  '行内公式 $E=mc^2$ 与块级公式：',
  '',
  '$$a^2+b^2=c^2$$',
  '',
  '价格 $5，变量 $HOME 未定义。',
  '',
  '行内代码 `$ARGUMENTS` 与 `$` 符号应当原样保留。',
  '',
  '```bash',
  'echo $HOME',
  'a=$1',
  '```',
].join('\n'))
const mathText = mathEl.textContent || ''
checks.push(
  ['真公式照常渲染（行内 + 块级 ≥ 2 个 katex）', mathEl.querySelectorAll('.katex').length >= 2],
  ['货币/变量文本未被误判为公式', mathText.includes('价格 $5，变量 $HOME 未定义')],
  ['行内代码里的 $ 与 $ARGUMENTS 原样保留', mathText.includes('$ARGUMENTS 与 $ 符号')],
  ['代码块里的 $ 原样保留', mathText.includes('echo $HOME') && mathText.includes('a=$1')],
)

// ── 3. 真实文档（ZCode 使用手册 158KB）──────────────────────────────
const realEl = renderTo(realDoc)
const realText = realEl.textContent || ''
checks.push(
  ['真实文档：无被 KaTeX 误吞的公式', realEl.querySelectorAll('.katex').length === 0],
  [
    '真实文档：结构完整（表格 44 / 标题 210 / 代码块 29）',
    realEl.querySelectorAll('table').length === 44
    && realEl.querySelectorAll('h1,h2,h3,h4,h5,h6').length === 210
    && realEl.querySelectorAll('.code-block-wrapper').length === 29,
  ],
  ['真实文档：行内代码 $ 保留（"通过 $ 调用技能"）', realText.includes('通过 $ 调用技能')],
  ['真实文档：无 KaTeX 转义出的标签碎片', !realText.includes('</code>') && !realText.includes('</p>') && !/<\/?h[1-6]>/.test(realText)],
  ['真实文档：$ARGUMENTS / ${user_config.键} 原样保留', realText.includes('$ARGUMENTS') && realText.includes('${user_config.键}')],
)

// ── 4. 预览排版样式（计算样式必须真的落到 DOM 上）────────────────────
const h2 = realEl.querySelector('h2')!
const td = realEl.querySelector('table td')!
const header = realEl.querySelector('.code-block-header')!
const copyBtn = realEl.querySelector('.code-copy-btn')!
checks.push(
  ['排版库已生效：h2 有下边框（1px）', css(h2, 'border-bottom-width') === '1px'],
  ['排版库已生效：表格单元格有边框（1px）', css(td, 'border-top-width') === '1px'],
  ['排版库已生效：标题字号大于正文', parseFloat(css(h2, 'font-size')) > parseFloat(css(realEl, 'font-size'))],
  ['正文字号跟随令牌（13px）', css(realEl, 'font-size') === '13px'],
  [
    '代码块头部为两端对齐（语言标签左 / 复制按钮右）',
    css(header, 'display') === 'flex' && css(header, 'justify-content') === 'space-between' && !!header.querySelector('.code-lang-label'),
  ],
  // flex 容器里的 flex-item 会被块化（inline-flex → flex），两种情况都算通过
  [`复制按钮为 flex 容器（display=${css(copyBtn, 'display')}）`, ['flex', 'inline-flex'].includes(css(copyBtn, 'display'))],
  ['代码块头部高度走控件令牌（22px）', css(header, 'height') === '22px'],
)

const lightColor = css(realEl, 'color')
document.documentElement.classList.add('dark')
await sleep(30)
const darkColor = css(realEl, 'color')
const darkTableBorder = css(td, 'border-top-color')
document.documentElement.classList.remove('dark')
checks.push(
  ['暗色主题下正文颜色跟随令牌切换', lightColor !== darkColor && darkColor !== ''],
  ['暗色主题下表格边框跟随令牌', darkTableBorder !== ''],
)

const mountLines = checks.map(([label, ok]) => `${ok ? 'PASS' : 'FAIL'}  ${label}`)
document.getElementById('probes')!.textContent = mountLines.join('\n')
console.log(mountLines.join('\n'))
