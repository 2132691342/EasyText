<script lang="ts" setup>
import { ref, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useEditorStore } from '@/stores'
import { OpenDirectoryDialog, ReadFile, GetDirectoryTree, SearchInFiles, ReplaceInFiles } from '../../wailsjs/go/main/App'
import { ElMessage } from 'element-plus'

const props = defineProps<{ visible: boolean; mode: 'find' | 'replace' | 'files' | 'global' | 'mark' }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const ed = useEditorStore()
const tab = ref<'find' | 'replace' | 'files' | 'global' | 'mark'>('find')

watch(() => props.mode, (m) => { if (m) tab.value = m }, { immediate: true })

// 🆕 V2.0.0 全局搜索
const showGlobalSearch = ref(false)

// ==================== 查找 Tab ====================
const fText = ref('')
const fBack = ref(false)
const fWhole = ref(false)
const fCase = ref(false)
const fWrap = ref(true)
const fMode = ref<'normal' | 'extend' | 'regex'>('normal')
const fResults = ref<number[]>([])
const fIndex = ref(-1)
const fStatus = ref('就绪')

// ==================== 替换 Tab ====================
const rText = ref('')
const rWith = ref('')
const rBack = ref(false)
const rWhole = ref(false)
const rCase = ref(false)
const rWrap = ref(true)
const rMode = ref<'normal' | 'extend' | 'regex'>('normal')

// ==================== 目录查找 Tab ====================
const dText = ref('')
const dWith = ref('')
const dDir = ref('')
const dFilter = ref(false)
const dFilterVal = ref('')
const dSkipDir = ref(false)
const dSkipVal = ref('')
const dWhole = ref(false)
const dCase = ref(false)
const dMode = ref<'normal' | 'extend' | 'regex'>('normal')
const dSkipChild = ref(false)
const dSkipHide = ref(true)
const dSkipBin = ref(true)
const dSkipBig = ref(true)
const dMaxSize = ref(20)
const dResults = ref<{ file: string; line: number; content: string }[]>([])
const dLoading = ref(false)

// ==================== 标记 Tab ====================
const mText = ref('')
const mWhole = ref(false)
const mCase = ref(false)
const mMode = ref<'normal' | 'extend' | 'regex'>('normal')

// ==================== 工具函数 ====================
function processExtend(p: string): string {
  return p.replace(/\\n/g, '\n').replace(/\\r/g, '\r').replace(/\\t/g, '\t').replace(/\\0/g, '\0')
    .replace(/\\x([0-9a-fA-F]{2})/g, (_, h) => String.fromCharCode(parseInt(h, 16)))
}

function buildRegex(pattern: string, mode: string, cs: boolean, whole: boolean): RegExp | null {
  let p = pattern
  if (mode === 'extend') { p = processExtend(p); p = p.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') }
  else if (mode === 'normal') { p = p.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') }
  if (whole) p = `\\b${p}\\b`
  try { return new RegExp(p, cs ? 'g' : 'gi') } catch { return null }
}

// ==================== 查找操作 ====================
function doFind() {
  const tab = ed.activeTab
  if (!tab || !fText.value) { fResults.value = []; fIndex.value = -1; fStatus.value = '请输入查找内容'; return }
  const regex = buildRegex(fText.value, fMode.value, fCase.value, fWhole.value)
  const idxs: number[] = []
  if (regex) { let m; while ((m = regex.exec(tab.content)) !== null) { idxs.push(m.index); if (m.index === regex.lastIndex) regex.lastIndex++ } }
  fResults.value = idxs
  if (idxs.length > 0) {
    fIndex.value = fBack.value ? idxs.length - 1 : 0
    fStatus.value = `找到 ${idxs.length} 处匹配`
    dispatchEd('scroll-to-pos', idxs[fIndex.value])
  } else {
    fIndex.value = -1
    fStatus.value = `找不到 "${fText.value}"`
  }
  dispatchEd('set-search-term', fText.value)
}

function findNext() {
  if (fResults.value.length === 0) { doFind(); return }
  fIndex.value = fBack.value
    ? (fIndex.value - 1 + fResults.value.length) % fResults.value.length
    : (fIndex.value + 1) % fResults.value.length
  dispatchEd('scroll-to-pos', fResults.value[fIndex.value])
  fStatus.value = `位置 ${fIndex.value + 1}/${fResults.value.length}`
}

function findPrev() {
  if (fResults.value.length === 0) return
  fIndex.value = fBack.value
    ? (fIndex.value + 1) % fResults.value.length
    : (fIndex.value - 1 + fResults.value.length) % fResults.value.length
  dispatchEd('scroll-to-pos', fResults.value[fIndex.value])
  fStatus.value = `位置 ${fIndex.value + 1}/${fResults.value.length}`
}

function countMatches() { doFind(); fStatus.value = `共 ${fResults.value.length} 处匹配` }

function findAllCurrent() {
  doFind()
  if (fResults.value.length === 0) { fStatus.value = '当前文档无匹配'; return }
  dispatchEd('highlight-all', fResults.value, fText.value.length)
  sendFindResults(`查找: "${fText.value}"`, fResults.value, ed.activeTab)
}

function findAllOpen() {
  let total = 0
  for (const t of ed.tabs) {
    if (t.viewType !== 'code') continue
    const r = buildRegex(fText.value, fMode.value, fCase.value, fWhole.value)
    if (r) { let m; while ((m = r.exec(t.content)) !== null) { total++; if (m.index === r.lastIndex) r.lastIndex++ } }
  }
  fStatus.value = `所有打开文档共 ${total} 处匹配`
}

function copyFindResults() {
  doFind()
  if (fResults.value.length === 0) return
  const tab = ed.activeTab
  if (!tab) return
  const lines = fResults.value.map(i => {
    const ls = tab.content.lastIndexOf('\n', i) + 1
    const le = tab.content.indexOf('\n', i)
    return tab.content.substring(ls, le === -1 ? undefined : le)
  })
  navigator.clipboard.writeText(lines.join('\n'))
  fStatus.value = `已复制 ${lines.length} 行`
}

function clearFind() { fResults.value = []; fIndex.value = -1; fStatus.value = '结果已清空'; dispatchEd('clear-highlight') }

// ==================== 替换操作 ====================
function replaceFind() {
  const tab = ed.activeTab
  if (!tab || !rText.value) { fStatus.value = '请输入查找内容'; return }
  const regex = buildRegex(rText.value, rMode.value, rCase.value, rWhole.value)
  const idxs: number[] = []
  if (regex) { let m; while ((m = regex.exec(tab.content)) !== null) { idxs.push(m.index); if (m.index === regex.lastIndex) regex.lastIndex++ } }
  fResults.value = idxs
  if (idxs.length > 0) {
    fIndex.value = rBack.value ? idxs.length - 1 : 0
    dispatchEd('scroll-to-pos', idxs[fIndex.value])
    fStatus.value = `位置 ${fIndex.value + 1}/${idxs.length}`
  } else { fStatus.value = `找不到 "${rText.value}"` }
}

function replaceOne() {
  const tab = ed.activeTab
  if (!tab || fIndex.value === -1) return
  const start = fResults.value[fIndex.value]
  const regex = buildRegex(rText.value, rMode.value, rCase.value, rWhole.value)
  if (!regex) return
  regex.lastIndex = start
  const m = regex.exec(tab.content)
  if (!m) return
  ed.updateTabContent(tab.id, tab.content.slice(0, start) + rWith.value + tab.content.slice(start + m[0].length))
  fStatus.value = '已替换 1 处'
  setTimeout(replaceFind, 0)
}

function replaceAll() {
  const tab = ed.activeTab
  if (!tab) return
  const regex = buildRegex(rText.value, rMode.value, rCase.value, rWhole.value)
  if (!regex) return
  let count = 0
  const nc = tab.content.replace(regex, () => { count++; return rWith.value })
  if (count > 0) { ed.updateTabContent(tab.id, nc); fStatus.value = `已替换 ${count} 处` }
  else { fStatus.value = `找不到 "${rText.value}"` }
}

function replaceAllOpen() {
  let total = 0
  for (const t of ed.tabs) {
    if (t.viewType !== 'code') continue
    const regex = buildRegex(rText.value, rMode.value, rCase.value, rWhole.value)
    if (!regex) continue
    const nc = t.content.replace(regex, () => { total++; return rWith.value })
    if (nc !== t.content) ed.updateTabContent(t.id, nc)
  }
  fStatus.value = `所有打开文档共替换 ${total} 处`
}

// ==================== 目录查找 ====================
async function selectDir() { try { const d = await OpenDirectoryDialog(); if (d) dDir.value = d } catch (e) { console.warn(e) } }

async function dirFindAll() {
  if (!dText.value || !dDir.value) { fStatus.value = '请输入查找内容并选择目录'; return }
  dLoading.value = true; dResults.value = []; fStatus.value = '正在搜索...'
  try {
    const tree = await GetDirectoryTree(dDir.value)
    if (!tree?.root) { fStatus.value = '无法读取目录'; dLoading.value = false; return }
    const textFiles: string[] = []
    const textExts = new Set(['txt','md','json','js','ts','html','css','xml','yaml','yml','toml','ini','cfg','go','java','py','c','cpp','h','hpp','rs','sh','bat','sql','vue','svelte','php','rb','swift','kt','scala','lua','r','pl','pm','tex','log','csv','env','gitignore'])
    function walk(n: any) {
      if (!n) return
      if (!n.isDir && n.path) {
        const ext = (n.name || '').split('.').pop()?.toLowerCase() || ''
        if (dFilter.value && dFilterVal.value) { const ae = dFilterVal.value.split(':').map(s => s.replace('*.','').toLowerCase()); if (!ae.includes(ext)) return }
        if (textExts.has(ext) || ext === '') textFiles.push(n.path)
      }
      if (n.children && !dSkipChild.value) for (const c of n.children) walk(c)
    }
    walk(tree.root)
    const regex = buildRegex(dText.value, dMode.value, dCase.value, dWhole.value)
    let fc = 0
    for (const fp of textFiles.slice(0, 200)) {
      if (dResults.value.length >= 500) break
      try {
        const r = await ReadFile(fp)
        if (!r?.content) continue
        const lines = r.content.split('\n')
        for (let i = 0; i < lines.length; i++) {
          if (dResults.value.length >= 500) break
          if (regex) { regex.lastIndex = 0; if (regex.test(lines[i])) dResults.value.push({ file: fp, line: i + 1, content: lines[i].trim().substring(0, 200) }) }
          else { const l1 = dCase.value ? lines[i] : lines[i].toLowerCase(); const q1 = dCase.value ? dText.value : dText.value.toLowerCase(); if (l1.includes(q1)) dResults.value.push({ file: fp, line: i + 1, content: lines[i].trim().substring(0, 200) }) }
        }
        fc++
      } catch (e) { console.warn(e) }
    }
    fStatus.value = `搜索完成：${fc} 文件，${dResults.value.length} 处匹配`
    sendFindResults(`目录查找: "${dText.value}"`, [], undefined, dResults.value)
  } catch (e) { fStatus.value = `搜索失败: ${e}` }
  dLoading.value = false
}

async function dirReplace() {
  if (!dDir.value) return
  if (dResults.value.length === 0) { await dirFindAll() }
  const fileSet = new Set(dResults.value.map(r => r.file))
  let total = 0
  for (const fp of fileSet) {
    try {
      const r = await ReadFile(fp)
      if (!r?.content) continue
      const regex = buildRegex(dText.value, dMode.value, dCase.value, dWhole.value)
      if (!regex) continue
      const nc = r.content.replace(regex, () => { total++; return dWith.value })
      if (nc !== r.content) {
        const { SaveFile } = await import('../../wailsjs/go/main/App')
        await SaveFile(fp, nc, r.info.encoding)
      }
    } catch (e) { console.warn(e) }
  }
  fStatus.value = `${fileSet.size} 文件，${total} 处替换`
  dResults.value = []
}

// ==================== 🆕 V2.0.0 全局搜索 ====================
const gDir = ref('')
const gText = ref('')
const gReplace = ref('')
const gCase = ref(false)
const gWhole = ref(false)
const gRegex = ref(false)
const gSubdir = ref(true)
const gPattern = ref('')
const gResults = ref<{ file: string; line: number; content: string }[]>([])
const gLoading = ref(false)

async function selectGlobalDir() { try { const d = await OpenDirectoryDialog(); if (d) gDir.value = d } catch { /* ignore */ } }

async function globalSearch() {
  if (!gText.value || !gDir.value) { fStatus.value = '请输入查找内容并选择目录'; return }
  gLoading.value = true; gResults.value = []; fStatus.value = '正在搜索...'
  try {
    const tree = await GetDirectoryTree(gDir.value)
    if (!tree?.root) { fStatus.value = '无法读取目录'; gLoading.value = false; return }
    const textFiles: string[] = []
    const textExts = new Set(['txt','md','json','js','ts','html','css','xml','yaml','yml','toml','ini','cfg','go','java','py','c','cpp','h','hpp','rs','sh','bat','sql','vue','svelte','php','rb','swift','kt','scala','lua','r','pl','pm','tex','log','csv','env','gitignore','svg','Makefile','Dockerfile'])
    function walk(n: any) {
      if (!n) return
      if (!n.isDir && n.path) {
        const ext = (n.name || '').split('.').pop()?.toLowerCase() || ''
        if (gPattern.value) {
          const patterns = gPattern.value.split(',').map(s => s.trim().replace('*.',''))
          if (!patterns.some(p => (n.name || '').includes(p))) return
        }
        if (textExts.has(ext) || ext === '' || n.name === 'Makefile' || n.name === 'Dockerfile') textFiles.push(n.path)
      }
      if (n.children && gSubdir.value) for (const c of n.children) walk(c)
    }
    walk(tree.root)
    // 使用 SearchInFiles API
    const searchOpts = {
      search: gText.value,
      replace: gReplace.value,
      caseSensitive: gCase.value,
      wholeWord: gWhole.value,
      useRegex: gRegex.value,
      includeSubdir: gSubdir.value,
      filePattern: gPattern.value,
    }
    const results = await SearchInFiles(textFiles.slice(0, 500), gText.value, searchOpts as any)
    if (results) {
      for (const r of results as any[]) {
        if (r.matches) {
          for (const m of r.matches) {
            gResults.value.push({ file: r.file, line: m.line, content: m.content?.substring(0, 200) || '' })
          }
        }
      }
    }
    fStatus.value = `搜索完成：${gResults.value.length} 处匹配`
    sendFindResults(`全局搜索: "${gText.value}"`, [], undefined, gResults.value)
  } catch (e: any) {
    fStatus.value = `搜索失败: ${e?.message || ''}`
  }
  gLoading.value = false
}

async function globalReplace() {
  if (!gDir.value || !gText.value) return
  if (gResults.value.length === 0) { await globalSearch() }
  const fileSet = new Set(gResults.value.map(r => r.file))
  let total = 0
  const failed: string[] = []
  for (const fp of fileSet) {
    try {
      const r = await ReadFile(fp)
      if (!r?.content) continue
      const regex = buildRegex(gText.value, gRegex.value ? 'regex' : 'normal', gCase.value, gWhole.value)
      if (!regex) continue
      const nc = r.content.replace(regex, () => { total++; return gReplace.value })
      if (nc !== r.content) {
        const { SaveFile } = await import('../../wailsjs/go/main/App')
        await SaveFile(fp, nc, r.info.encoding)
      }
    } catch { failed.push(fp) }
  }
  fStatus.value = `${fileSet.size} 文件，${total} 处替换`
  if (failed.length > 0) fStatus.value += `，${failed.length} 个文件失败`
  gResults.value = []
}

// ==================== 标记操作 ====================
function markAll() {
  const tab = ed.activeTab
  if (!tab || !mText.value) return
  const regex = buildRegex(mText.value, mMode.value, mCase.value, mWhole.value)
  const idxs: number[] = []
  if (regex) { let m; while ((m = regex.exec(tab.content)) !== null) { idxs.push(m.index); if (m.index === regex.lastIndex) regex.lastIndex++ } }
  if (idxs.length > 0) {
    dispatchEd('mark-all', idxs, mText.value.length)
    fStatus.value = `已标记 ${idxs.length} 处`
  } else { fStatus.value = `找不到 "${mText.value}"` }
}

function markAndBookmark() { markAll(); dispatchEd('toggle-bookmark'); fStatus.value = '已标记并添加书签' }
function clearMarks() { dispatchEd('clear-mark'); fStatus.value = '标记已清除' }
function clearAllMarks() { dispatchEd('clear-all-marks'); fStatus.value = '所有标记和书签已清除' }

// ==================== 辅助 ====================
let lastSearch = ''
function dispatchEd(cmd: string, ...args: any[]) {
  document.dispatchEvent(new CustomEvent('editor-command', { detail: { cmd, args } }))
  if (cmd === 'set-search-term' && args[0]) lastSearch = args[0] as string
  if (cmd === 'find-next' || cmd === 'find-prev') {
    document.dispatchEvent(new CustomEvent('set-search-term', { detail: lastSearch }))
  }
}

function sendFindResults(title: string, positions?: number[], tabInfo?: any, dirResults?: any[]) {
  let results: any[] = []
  if (dirResults) {
    results = dirResults.slice(0, 200)
  } else if (positions && tabInfo) {
    const lines = tabInfo.content.split('\n')
    results = positions.slice(0, 200).map(idx => {
      let cc = 0
      for (let i = 0; i < lines.length; i++) { if (idx >= cc && idx <= cc + lines[i].length) { cc = i + 1; break } else cc += lines[i].length + 1 }
      return { file: tabInfo.path || tabInfo.name, line: cc, content: (lines[cc - 1] || '').substring(0, 200) }
    })
  }
  document.dispatchEvent(new CustomEvent('find-results-update', { detail: { title, results } }))
}

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
  else if (e.key === 'Enter') {
    if (tab.value === 'find') findNext()
    else if (tab.value === 'replace') replaceFind()
  }
}

const fInput = ref<HTMLInputElement | null>(null)
// 读取当前编辑器选区：优先 window.getSelection（CodeMirror 6 默认会将选区同步到原生 selection），
// 否则派发自定义事件让 CodeEditor 主动回报。修复「手动打开搜索，搜不到已选文本」的 Bug。
function readEditorSelection(): string {
  try {
    const sel = window.getSelection?.()?.toString?.() || ''
    return sel.trim()
  } catch {
    return ''
  }
}
function prefilledWithSelection() {
  const sel = readEditorSelection()
  if (!sel) return
  // 只有在用户尚未手动输入时填，避免覆盖正在键入的内容
  if (!fText.value) fText.value = sel
  if (!rText.value) rText.value = sel
  if (!mText.value) mText.value = sel
}
watch(() => props.visible, async (v) => {
  if (v) {
    prefilledWithSelection()
    await nextTick(); fInput.value?.focus()
  }
})
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="find-win-overlay" @click.self="emit('close')" @keydown="onKey">
      <div class="find-win">
        <!-- 标题栏 -->
        <div class="find-win-titlebar">
          <span class="text-sm">查找 / 替换 / 目录查找 / 标记</span>
          <button class="find-win-close" @click="emit('close')">✕</button>
        </div>

        <!-- Tab 栏 (notepad-- 风格) -->
        <div class="find-win-tabs">
          <button v-for="t in [{k:'find',l:'查找'},{k:'replace',l:'替换'},{k:'files',l:'目录查找'},{k:'global',l:'全局搜索'},{k:'mark',l:'标记'}]" :key="t.k"
            class="find-tab" :class="{ active: tab === t.k }" @click="tab = t.k as any">{{ t.l }}</button>
        </div>

        <div class="find-win-body">
          <!-- ====== 查找 ====== -->
          <div v-if="tab==='find'" class="find-form-row">
            <div class="flex-1">
              <div class="find-row">
                <label class="w-20 text-right text-xs">查找目标:</label>
                <input ref="fInput" v-model="fText" class="find-input flex-1" @keydown.enter.prevent="findNext" />
                <span class="text-xs text-gray-400 w-20 text-right">{{ fResults.length ? `${fIndex+1}/${fResults.length}` : '0/0' }}</span>
              </div>
              <div class="find-checks ml-20">
                <label class="check"><input type="checkbox" v-model="fBack" /> 向后查找</label>
                <label class="check"><input type="checkbox" v-model="fWhole" /> 全词匹配</label>
                <label class="check"><input type="checkbox" v-model="fCase" /> 区分大小写</label>
                <label class="check"><input type="checkbox" v-model="fWrap" /> 循环查找</label>
              </div>
              <fieldset class="find-mode ml-20 mt-1">
                <legend>搜索模式</legend>
                <label class="check"><input type="radio" v-model="fMode" value="normal" /> 普通</label>
                <label class="check"><input type="radio" v-model="fMode" value="extend" /> 扩展(\n,\r,\t,\0,\x...)</label>
                <label class="check"><input type="radio" v-model="fMode" value="regex" /> 正则表达式</label>
              </fieldset>
            </div>
            <div class="find-btns">
              <button class="btn primary" @click="findNext">查找下一个</button>
              <button class="btn" @click="findPrev">查找上一个</button>
              <button class="btn" @click="countMatches">计数</button>
              <button class="btn" @click="findAllCurrent">当前文档全部查找</button>
              <button class="btn" @click="findAllOpen">所有打开文档查找</button>
              <button class="btn" @click="copyFindResults">复制结果</button>
              <button class="btn" @click="clearFind">清空结果</button>
              <div class="flex-1" />
              <button class="btn" @click="emit('close')">关闭</button>
            </div>
          </div>

          <!-- ====== 替换 ====== -->
          <div v-if="tab==='replace'" class="find-form-row">
            <div class="flex-1">
              <div class="find-row">
                <label class="w-20 text-right text-xs">查找目标:</label>
                <input v-model="rText" class="find-input flex-1" @keydown.enter.prevent="replaceFind" />
              </div>
              <div class="find-row">
                <label class="w-20 text-right text-xs">替换为:</label>
                <input v-model="rWith" class="find-input flex-1" />
              </div>
              <div class="find-checks ml-20">
                <label class="check"><input type="checkbox" v-model="rBack" /> 向后查找</label>
                <label class="check"><input type="checkbox" v-model="rWhole" /> 全词匹配</label>
                <label class="check"><input type="checkbox" v-model="rCase" /> 区分大小写</label>
                <label class="check"><input type="checkbox" v-model="rWrap" /> 循环查找</label>
              </div>
              <fieldset class="find-mode ml-20 mt-1">
                <legend>搜索模式</legend>
                <label class="check"><input type="radio" v-model="rMode" value="normal" /> 普通</label>
                <label class="check"><input type="radio" v-model="rMode" value="extend" /> 扩展(\n,\r,\t,\0,\x...)</label>
                <label class="check"><input type="radio" v-model="rMode" value="regex" /> 正则表达式</label>
              </fieldset>
            </div>
            <div class="find-btns">
              <button class="btn primary" @click="replaceFind">查找下一个</button>
              <button class="btn" @click="replaceOne">替换</button>
              <button class="btn" @click="replaceAll">全部替换</button>
              <button class="btn" @click="replaceAllOpen">所有打开文档替换</button>
              <div class="flex-1" />
              <button class="btn" @click="emit('close')">关闭</button>
            </div>
          </div>

          <!-- ====== 目录查找 ====== -->
          <div v-if="tab==='files'" class="find-form-row">
            <div class="flex-1">
              <div class="find-row">
                <label class="w-20 text-right text-xs">目标目录:</label>
                <input v-model="dDir" class="find-input flex-1" readonly />
                <button class="btn" @click="selectDir">选择</button>
              </div>
              <div class="find-row">
                <label class="w-20 text-right text-xs">查找目标:</label>
                <input v-model="dText" class="find-input flex-1" />
              </div>
              <div class="find-row">
                <label class="w-20 text-right text-xs">替换为:</label>
                <input v-model="dWith" class="find-input flex-1" />
              </div>
              <div class="find-row">
                <label class="check w-20 text-right"><input type="checkbox" v-model="dFilter" /> 文件类型:</label>
                <input v-model="dFilterVal" :disabled="!dFilter" placeholder="*.c:*.cpp:*.h" class="find-input flex-1" />
              </div>
              <div class="find-row">
                <label class="check w-20 text-right"><input type="checkbox" v-model="dSkipDir" /> 跳过目录:</label>
                <input v-model="dSkipVal" :disabled="!dSkipDir" placeholder="debug:.git" class="find-input flex-1" />
              </div>
              <div class="find-checks ml-20">
                <label class="check"><input type="checkbox" v-model="dWhole" /> 全词匹配</label>
                <label class="check"><input type="checkbox" v-model="dCase" /> 区分大小写</label>
              </div>
              <div class="flex gap-2 ml-20 mt-1">
                <fieldset class="find-mode flex-1">
                  <legend>搜索模式</legend>
                  <label class="check"><input type="radio" v-model="dMode" value="normal" /> 普通</label>
                  <label class="check"><input type="radio" v-model="dMode" value="extend" /> 扩展</label>
                  <label class="check"><input type="radio" v-model="dMode" value="regex" /> 正则</label>
                </fieldset>
                <fieldset class="find-mode flex-1">
                  <legend>选项</legend>
                  <label class="check"><input type="checkbox" v-model="dSkipChild" /> 跳过子目录</label>
                  <label class="check"><input type="checkbox" v-model="dSkipHide" /> 跳过隐藏文件</label>
                  <label class="check"><input type="checkbox" v-model="dSkipBin" /> 跳过二进制</label>
                </fieldset>
              </div>
            </div>
            <div class="find-btns">
              <button class="btn primary" :disabled="dLoading" @click="dirFindAll">{{ dLoading ? '搜索中...' : '全部查找' }}</button>
              <button class="btn" @click="dirReplace">替换文件</button>
              <button class="btn" @click="dResults=[]">清空结果</button>
              <div class="flex-1" />
              <button class="btn" @click="emit('close')">关闭</button>
            </div>
          </div>

          <!-- ====== 🆕 V2.0.0 全局搜索 ====== -->
          <div v-if="tab==='global'" class="find-form-row">
            <div class="flex-1">
              <div class="find-row">
                <label class="w-20 text-right text-xs">搜索目录:</label>
                <input v-model="gDir" class="find-input flex-1" readonly />
                <button class="btn" @click="selectGlobalDir">选择</button>
              </div>
              <div class="find-row">
                <label class="w-20 text-right text-xs">查找目标:</label>
                <input v-model="gText" class="find-input flex-1" @keydown.enter.prevent="globalSearch" />
              </div>
              <div class="find-row">
                <label class="w-20 text-right text-xs">替换为:</label>
                <input v-model="gReplace" class="find-input flex-1" />
              </div>
              <div class="find-row">
                <label class="w-20 text-right text-xs">文件模式:</label>
                <input v-model="gPattern" placeholder="*.go,*.ts (可选)" class="find-input flex-1" />
              </div>
              <div class="find-checks ml-20">
                <label class="check"><input type="checkbox" v-model="gCase" /> 区分大小写</label>
                <label class="check"><input type="checkbox" v-model="gWhole" /> 全词匹配</label>
                <label class="check"><input type="checkbox" v-model="gRegex" /> 正则表达式</label>
                <label class="check"><input type="checkbox" v-model="gSubdir" /> 包含子目录</label>
              </div>
            </div>
            <div class="find-btns">
              <button class="btn primary" :disabled="gLoading" @click="globalSearch">{{ gLoading ? '搜索中...' : '全部查找' }}</button>
              <button class="btn" @click="globalReplace">替换文件</button>
              <button class="btn" @click="gResults = []">清空结果</button>
              <div class="flex-1" />
              <button class="btn" @click="emit('close')">关闭</button>
            </div>
          </div>

          <!-- ====== 标记 ====== -->
          <div v-if="tab==='mark'" class="find-form-row">
            <div class="flex-1">
              <div class="find-row">
                <label class="w-20 text-right text-xs">标记内容:</label>
                <input v-model="mText" class="find-input flex-1" />
              </div>
              <div class="find-checks ml-20">
                <label class="check"><input type="checkbox" v-model="mWhole" /> 全词匹配</label>
                <label class="check"><input type="checkbox" v-model="mCase" /> 区分大小写</label>
              </div>
              <fieldset class="find-mode ml-20 mt-1">
                <legend>搜索模式</legend>
                <label class="check"><input type="radio" v-model="mMode" value="normal" /> 普通</label>
                <label class="check"><input type="radio" v-model="mMode" value="extend" /> 扩展(\n,\r,\t,\0,\x...)</label>
                <label class="check"><input type="radio" v-model="mMode" value="regex" /> 正则表达式</label>
              </fieldset>
            </div>
            <div class="find-btns">
              <button class="btn primary" @click="markAll">全部标记</button>
              <button class="btn" @click="markAndBookmark">标记并书签</button>
              <button class="btn" @click="clearMarks">清除标记</button>
              <button class="btn" @click="clearAllMarks">清除全部</button>
              <div class="flex-1" />
              <button class="btn" @click="emit('close')">关闭</button>
            </div>
          </div>
        </div>

        <!-- 状态栏 -->
        <div class="find-win-status">{{ fStatus }}</div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.find-win-overlay {
  position: fixed; inset: 0; z-index: 9999; display: flex; align-items: flex-start; justify-content: center;
  padding-top: 80px; background: rgba(0,0,0,0.15);
}
.find-win {
  width: 720px; max-height: 500px; background: #fff; border: 1px solid #d1d5db;
  box-shadow: 0 8px 32px rgba(0,0,0,0.2); border-radius: 6px; display: flex; flex-direction: column;
  font-size: 12px;
}
html.dark .find-win { background: #2d2d2d; border-color: #454545; }

.find-win-titlebar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 4px 10px; background: #f5f5f5; border-bottom: 1px solid #ddd; border-radius: 6px 6px 0 0;
  color: #333; user-select: none;
}
html.dark .find-win-titlebar { background: #3c3c3c; border-color: #555; color: #ddd; }
.find-win-close { padding: 2px 6px; border: none; background: none; cursor: pointer; font-size: 14px; color: #666; border-radius: 3px; }
.find-win-close:hover { background: #e0e0e0; color: #333; }
html.dark .find-win-close:hover { background: #555; color: #fff; }

.find-win-tabs { display: flex; border-bottom: 1px solid #ddd; background: #f0f0f0; }
html.dark .find-win-tabs { border-color: #555; background: #353535; }
.find-tab {
  padding: 5px 16px; border: none; background: none; cursor: pointer; font-size: 12px;
  color: #555; border-right: 1px solid #ddd;
}
html.dark .find-tab { color: #aaa; border-color: #555; }
.find-tab.active { background: #fff; color: #1a73e8; font-weight: 500; }
html.dark .find-tab.active { background: #1e1e1e; color: #60a5fa; }

.find-win-body { padding: 10px; overflow-y: auto; flex: 1; }
.find-form-row { display: flex; gap: 10px; }
.find-row { display: flex; align-items: center; gap: 6px; margin-bottom: 6px; }
.find-checks { display: flex; flex-wrap: wrap; gap: 4px 12px; margin-bottom: 6px; }
.find-mode { border: 1px solid #d1d5db; border-radius: 4px; padding: 4px 8px; display: flex; gap: 8px; flex-wrap: wrap; }
html.dark .find-mode { border-color: #555; }
.find-mode legend { font-size: 11px; color: #888; padding: 0 2px; }

.check { display: flex; align-items: center; gap: 2px; font-size: 12px; color: #555; cursor: pointer; }
html.dark .check { color: #aaa; }
.check input[type="checkbox"], .check input[type="radio"] { width: 13px; height: 13px; margin: 0; cursor: pointer; }

.find-input {
  padding: 3px 6px; border: 1px solid #ccc; border-radius: 3px; font-size: 12px;
  background: #fff; color: #333; outline: none;
}
.find-input:focus { border-color: #3b82f6; }
html.dark .find-input { background: #1e1e1e; color: #d4d4d4; border-color: #555; }
.find-input:disabled { opacity: 0.4; }

.find-btns { display: flex; flex-direction: column; gap: 4px; width: 140px; flex-shrink: 0; }
.btn {
  padding: 5px 8px; border: 1px solid #ccc; border-radius: 3px; background: #fff;
  font-size: 12px; cursor: pointer; color: #333; text-align: center;
}
.btn:hover { background: #f0f0f0; }
.btn:disabled { opacity: 0.4; cursor: default; }
.btn.primary { background: #3b82f6; color: #fff; border-color: #3b82f6; }
.btn.primary:hover { background: #2563eb; }
html.dark .btn { background: #3c3c3c; color: #d4d4d4; border-color: #555; }
html.dark .btn:hover { background: #4c4c4c; }

.find-win-status {
  padding: 3px 10px; border-top: 1px solid #ddd; font-size: 11px; color: #888;
  background: #f5f5f5; border-radius: 0 0 6px 6px;
}
html.dark .find-win-status { background: #252526; border-color: #555; color: #aaa; }
</style>
