<script lang="ts" setup>
/**
 * 菜单栏（Menu Bar）v2.1
 *
 * 设计目标（对标 Notepad-- / Sublime Text）：
 *  1. 顶部菜单条 28px，菜单按钮 padding: 0 10px，hover 高亮过渡。
 *  2. Lucide 图标替代 emoji 风格的 ✓/▶（Check / ChevronRight）。
 *  3. 键盘可达：菜单按钮 tabindex=0，↑/↓ 在项间移动，Enter 触发，
 *     Esc 关闭，右箭头进入子菜单。
 *  4. 状态从 store 实时取（toolbar 改 → 菜单勾选同步）。
 *  5. 子菜单位置自适应右边界。
 *  6. 过渡动画 80ms。
 */
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { Check, ChevronRight } from 'lucide-vue-next'
import { useEditorStore, useSettingStore } from '@/stores'
import { GetRecentFiles } from '../../wailsjs/go/main/App'
import { NOT_IMPLEMENTED } from '@/utils/commands'

const emit = defineEmits<{ (e: 'cmd', name: string, ...args: any[]): void }>()
const ed = useEditorStore()
const se = useSettingStore()

// ---- 语言分组（与原实现保持一致） ----
const LANG_GROUPS: Record<string, string[]> = {
  A: ['ASP','ActionScript','Assembly','AutoIt','AviSynth','ASN.1'],
  B: ['BaanC','Bash','Batch','BlitzBasic'],
  C: ['C','C++','C#','Objective C','CSS','CMake','CoffeeScript','Csound'],
  D: ['D','Diff'],
  E: ['ESCRIPT','Erlang','Edifact'],
  F: ['Fortran','Fortran77','Forth','FreeBasic'],
  G: ['Go'],
  H: ['HTML'],
  I: ['IDL','ini','Intel HEX'],
  J: ['Java','JavaScript','Jsp','JSON'],
  L: ['Lua','Lisp','LaTeX'],
  M: ['Makefile','MMIXAL','MarkDown','Matlab'],
  N: ['Nfo','NSIS','NCrontab','Nim'],
  O: ['OScript','Octave'],
  P: ['Pascal','Perl','PHP','Po','PostScript','Pov','PowerShell','Properties file','PureBasic','Python'],
  R: ['R','RC','Ruby','Rust','Registry','REBOL'],
  S: ['Sql','Shell','Scheme','Smalltalk','S-Record','Swift','SPICE'],
  T: ['TeX','Tcl','TypeScript','Txt2tags'],
  V: ['VB','Verilog','Visual Basic','VHDL','Visual Prolog'],
  其他: ['XML','YAML','TXT','UserDefine'],
}

type Item = { label?: string; cmd?: string; sub?: string; key?: string; sep?: boolean; chk?: boolean; grp?: string; st?: string }

function showLabel(s?: string) { return (s || '').replace(/\(&.\)/g, '') }

const MENUS = [
  { label: '文件(&F)', items: [
    { label: '新建(&N)', cmd: 'new-file', key: 'Ctrl+T' },
    { label: '打开(&O)...', cmd: 'open-file', key: 'Ctrl+O' },
    { label: '打开目录...', cmd: 'open-directory' },
    { label: '从收藏夹打开', sub: 'fav' },
    { label: '在视图中打开(&V)', cmd: 'open-view' },
    { sep: true },
    { label: '保存(&S)', cmd: 'save', key: 'Ctrl+S' },
    { label: '另存为(&A)...', cmd: 'save-as', key: 'Ctrl+Alt+S' },
    { label: '全部保存', cmd: 'save-all' },
    { label: '重命名...', cmd: 'rename-file' },
    { sep: true },
    { label: '关闭(&C)', cmd: 'close-tab', key: 'Ctrl+W' },
    { label: '关闭所有', cmd: 'close-all', key: 'Ctrl+Shift+W' },
    { label: '关闭其他文档', cmd: 'close-others' },
    { label: '关闭左侧文档', cmd: 'close-left' },
    { label: '关闭右侧文档', cmd: 'close-right' },
    { sep: true },
    { label: '在新窗口中打开', cmd: 'new-window' },
    { label: '重新加载(&L)', cmd: 'reload-file' },
    { label: '清空最近文件记录', cmd: 'clear-history' },
    { label: '最近打开的文件', sub: 'recent' },
    { sep: true },
    { label: '草稿管理...', cmd: 'manage-drafts' },
    { label: '最近的文件...', cmd: 'recent-files' },
    { label: '最近的文件夹...', cmd: 'recent-folders' },
    { sep: true },
    { label: '保存工作空间...', cmd: 'save-workspace' },
    { label: '打开工作空间...', cmd: 'open-workspace' },
    { sep: true },
    { label: '打印(&P)...', cmd: 'print', key: 'Ctrl+P' },
    { sep: true },
    { label: '退出(&X)', cmd: 'exit', key: 'Ctrl+Q' },
  ] as Item[] },
  { label: '编辑(&E)', items: [
    { label: '撤销(&U)', cmd: 'undo', key: 'Ctrl+Z' },
    { label: '重做(&R)', cmd: 'redo', key: 'Ctrl+Y' },
    { sep: true },
    { label: '剪切(&T)', cmd: 'cut', key: 'Ctrl+X' },
    { label: '复制(&C)', cmd: 'copy', key: 'Ctrl+C' },
    { label: '粘贴(&P)', cmd: 'paste', key: 'Ctrl+V' },
    { label: '删除(&D)', cmd: 'delete', key: 'Del' },
    { label: '全选(&A)', cmd: 'select-all', key: 'Ctrl+A' },
    { sep: true },
    { label: '查找(&F)...', cmd: 'find', key: 'Ctrl+F' },
    { label: '查找下一个', cmd: 'find-next', key: 'F3' },
    { label: '查找上一个', cmd: 'find-prev', key: 'Shift+F3' },
    { label: '替换(&H)...', cmd: 'replace', key: 'Ctrl+H' },
    { label: '转到行(&G)...', cmd: 'goto-line', key: 'Ctrl+G' },
    { label: '转到匹配括号', cmd: 'goto-bracket', key: 'Ctrl+]' },
    { sep: true },
    { label: '列编辑模式(&X)', cmd: 'column-mode', key: 'Alt+X' },
    { label: '列块插入文本...', cmd: 'column-block' },
    { label: '缩进(&I)', cmd: 'indent', key: 'Tab' },
    { label: '减少缩进(&U)', cmd: 'dedent', key: 'Shift+Tab' },
    { sep: true },
    { label: '切换书签(&B)', cmd: 'toggle-bookmark', key: 'F2' },
    { label: '下一个书签', cmd: 'next-bookmark', key: 'Shift+F2' },
    { label: '清除所有书签', cmd: 'clear-bookmarks' },
    { sep: true },
    { label: '上一位置', cmd: 'prev-position' },
    { label: '下一位置', cmd: 'next-position' },
  ] as Item[] },
  { label: '查找(&S)', items: [
    { label: '快速查找...', cmd: 'find', key: 'Ctrl+F' },
    { label: '在文件中查找...', cmd: 'search-files', key: 'Ctrl+Shift+F' },
    { label: '目录查找...', cmd: 'find-dir', key: 'Ctrl+Shift+D' },
    { label: '全局搜索...', cmd: 'find-multi' },
    { label: '批量替换...', cmd: 'replace-multi' },
    { sep: true },
    { label: '标记全部', cmd: 'mark-all' },
    { label: '清除所有标记', cmd: 'clear-all-marks' },
    { sep: true },
    { label: '正则测试器...', cmd: 'regex-tester' },
  ] as Item[] },
  { label: '视图(&V)', items: [
    { label: '缩放', sub: 'zoom' },
    { label: '图标大小', sub: 'iconsize' },
    { sep: true },
    { label: '自动换行(&W)', cmd: 'toggle-wrap', chk: true },
    { label: '显示空白字符', cmd: 'toggle-whitespace', chk: true },
    { label: '缩进参考线', cmd: 'toggle-indent-guide', chk: true },
    { label: '文件列表', cmd: 'toggle-filelist', chk: true },
    { sep: true },
    { label: '工具栏', cmd: 'toggle-toolbar', chk: true, grp: 'toolbar', st: 'show' },
    { label: '状态栏', cmd: 'toggle-statusbar', chk: true, grp: 'toolbar', st: 'status' },
    { sep: true },
    { label: '显示文件树', cmd: 'showFileTree' },
    { label: '显示片段', cmd: 'snippet-panel' },
    { label: '显示书签', cmd: 'bookmark-panel' },
    { label: '显示函数列表', cmd: 'function-list' },
    { label: '显示文件监控', cmd: 'file-monitor' },
    { sep: true },
    { label: '全屏(&F)', cmd: 'fullscreen', key: 'F11' },
  ] as Item[] },
  { label: '编码(&N)', items: [
    { label: 'ANSI / 代码页探测', cmd: 'open-with-encoding' },
    { label: '以 UTF-8 打开', cmd: 'open-as-utf8' },
    { label: '以 UTF-8 无 BOM 保存', cmd: 'save-as-utf8-nobom' },
    { label: '以 UTF-8 加 BOM 保存', cmd: 'save-as-utf8-bom' },
    { sep: true },
    { label: '转为 UTF-8(&U)', cmd: 'convert-to-utf8' },
    { label: '转为 UTF-8 无 BOM', cmd: 'convert-to-utf8-nobom' },
    { label: '转为 UTF-8 加 BOM', cmd: 'convert-to-utf8-bom' },
    { sep: true },
    { label: '转为 GB2312', cmd: 'encode-GB2312' },
    { label: '转为 Shift_JIS', cmd: 'encode-SJIS' },
    { label: '转为 Big5', cmd: 'encode-ar' },
    { label: '转为 阿拉伯 (Windows)', cmd: 'encode-baltic' },
    { label: '转为 中欧', cmd: 'encode-ce' },
    { label: '转为 西里尔', cmd: 'encode-cyrillic' },
    { label: '转为 希腊语', cmd: 'encode-greek' },
    { label: '转为 希伯来语', cmd: 'encode-hebrew' },
    { label: '转为 韩文', cmd: 'encode-korean' },
    { label: '转为 泰语', cmd: 'encode-thai' },
    { label: '转为 土耳其语', cmd: 'encode-turkish' },
    { label: '转为 越南语', cmd: 'encode-vietnamese' },
    { label: '转为 西欧', cmd: 'encode-we' },
    { sep: true },
    { label: '编码转换器...', cmd: 'batch-convert' },
  ] as Item[] },
  { label: '语言(&L)', items: [] as Item[] }, // 语言子菜单渲染时用 LANG_GROUPS
  { label: '设置(&T)', items: [
    { label: '首选项...', cmd: 'preferences' },
    { label: '主题风格', sub: 'theme' },
    { label: '语言', sub: 'uilang' },
    { sep: true },
    { label: '快捷键管理...', cmd: 'shortcut-mgr' },
    { label: '宏管理...', cmd: 'macro-manager' },
    { label: '片段管理...', cmd: 'snippet-manager' },
    { sep: true },
    { label: '关联文件类型...', cmd: 'file-assoc' },
    { label: '导入主题...', cmd: 'import-theme' },
    { label: '导出主题...', cmd: 'export-theme' },
    { label: '导入快捷键...', cmd: 'import-shortcut' },
    { label: '导出快捷键...', cmd: 'export-shortcut' },
  ] as Item[] },
  { label: '工具(&O)', items: [
    { label: 'MD5 / 哈希...', cmd: 'md5-hash' },
    { label: '编码转换器...', cmd: 'batch-convert' },
    { label: '格式化转换器...', cmd: 'open-converter' },
    { label: 'JSON 路径查询...', cmd: 'json-path' },
    { label: 'JSON 转结构体...', cmd: 'json-to-struct' },
    { label: 'JSON 结构化 Diff...', cmd: 'json-diff' },
    { sep: true },
    { label: '正则测试器...', cmd: 'regex-tester' },
    { label: '脚本管理器...', cmd: 'script-manager' },
    { label: '批量查找替换...', cmd: 'batch-find' },
    { label: '批量重命名...', cmd: 'batch-rename' },
    { sep: true },
    { label: '文件比较...', cmd: 'file-cmp' },
    { label: '目录比较...', cmd: 'dir-cmp' },
    { label: '二进制比较...', cmd: 'bin-cmp' },
    { label: '选择左侧编码', cmd: 'sel-left' },
    { label: '选择右侧编码', cmd: 'sel-right' },
    { label: '比较规则...', cmd: 'cmp-rule' },
    { label: '最近的比较', cmd: 'recent-cmp' },
    { sep: true },
    { label: '十六进制查看', cmd: 'open-hex' },
    { label: '文本查看', cmd: 'open-text' },
    { label: '图片编辑器...', cmd: 'image-editor' },
    { label: '取色器...', cmd: 'color-picker' },
    { label: '日志模式', cmd: 'log-mode' },
    { sep: true },
    { label: '剪贴板历史...', cmd: 'clipboard-history' },
    { label: '记录宏', cmd: 'record-macro' },
    { label: '停止录制', cmd: 'stop-macro' },
    { label: '播放宏', cmd: 'play-macro' },
    { label: '保存宏...', cmd: 'save-macro' },
    { label: '运行宏(批量)', cmd: 'run-macro-multi' },
  ] as Item[] },
  { label: '对比', items: [
    { label: '文档对比...', cmd: 'open-diff' },
    { label: '文件比较...', cmd: 'file-cmp' },
    { label: '目录比较...', cmd: 'dir-cmp' },
  ] as Item[] },
  { label: '关于', items: [
    { label: '关于 EasyText...', cmd: 'about' },
  ] as Item[] },
]

// ---- 状态：实时从 store 取 ----
const wrapOn       = computed(() => !!se.config?.editor?.wordWrap)
const wsOn         = computed(() => !!se.config?.editor?.showWhitespace)
const showToolbar  = computed(() => !!se.config?.ui?.showToolBar)
const showStatus   = computed(() => !!se.config?.ui?.showStatusBar)
const showFileList = computed(() => !!se.config?.ui?.showFileListView)
const indentGuide  = computed(() => false) // TODO: 与 settingStore 对接（M3 再做）
function isChecked(it: Item): boolean {
  if (!it.chk) return false
  if (it.grp === 'toolbar' && it.st === 'show')  return showToolbar.value
  if (it.grp === 'toolbar' && it.st === 'status') return showStatus.value
  if (it.cmd === 'toggle-filelist') return showFileList.value
  if (it.cmd === 'toggle-wrap') return wrapOn.value
  if (it.cmd === 'toggle-whitespace') return wsOn.value
  if (it.cmd === 'toggle-indent-guide') return indentGuide.value
  return false
}

// ---- 打开 / 关闭 / 导航 ----
const openIdx = ref<number | null>(null)        // 顶层菜单下标
const subId   = ref<string | null>(null)        // 当前子菜单 key

const recent = ref<Array<{ path: string, name: string }>>([])
async function refreshRecent() {
  try {
    const r = await GetRecentFiles()
    if (Array.isArray(r)) recent.value = r as any
  } catch { /* ignore */ }
}

function toggle(idx: number) {
  if (openIdx.value === idx) {
    close()
  } else {
    openIdx.value = idx
    subId.value = null
    if (MENUS[idx].items.some((it: Item) => it.sub === 'recent')) refreshRecent()
  }
}
function onHover(idx: number) {
  if (openIdx.value !== null && openIdx.value !== idx) {
    openIdx.value = idx
    subId.value = null
    if (MENUS[idx].items.some((it: Item) => it.sub === 'recent')) refreshRecent()
  }
}
function openSub(id: string, ev?: MouseEvent) {
  subId.value = id
  if (id === 'recent') refreshRecent()
  if (id === 'recent' && ev) {
    nextTick(() => positionSubRight(ev.currentTarget as HTMLElement))
  }
}

const subFlip = ref(false)
function positionSubRight(anchor: HTMLElement) {
  const r = anchor.getBoundingClientRect()
  subFlip.value = (window.innerWidth - (r.right + 220)) < 0
}
function close() {
  openIdx.value = null
  subId.value = null
}

function clickItem(it: Item) {
  if (it.sep) return
  if (it.cmd && NOT_IMPLEMENTED.has(it.cmd)) {
    ElMessage.info('该功能暂未实现')
    close()
    return
  }
  if (it.cmd) {
    if (it.cmd.startsWith('recent-open-')) {
      emit('cmd', 'recent-open', it.cmd.slice('recent-open-'.length))
    } else {
      emit('cmd', it.cmd)
    }
    close()
    return
  }
  // 勾选类（菜单里有 cmd 但语义是切换）：emit cmd，由 useCommands 处理
  if (it.chk && it.cmd) {
    emit('cmd', it.cmd)
    close()
  }
}

function pickZoom(v: number) { emit('cmd', 'iconsize-' + v); close() }
function pickIconSize(v: number) { emit('cmd', 'iconsize-' + v); close() }
function pickTheme(name: string) { emit('cmd', 'theme-style', name); close() }
function pickLang(name: string) { emit('cmd', 'set-lang', name); close() }
function pickUiLang(name: string) { emit('cmd', name === 'zh' ? 'lang-zh' : 'lang-en'); close() }

// ---- 键盘 ----
function onMenuKey(e: KeyboardEvent) {
  if (openIdx.value === null) return
  const idx = openIdx.value
  const items = MENUS[idx].items.filter((x: Item) => !x.sep) as Item[]
  const cur = items.findIndex((x: Item) => x.label === (e.target as HTMLElement)?.dataset?.label)
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    moveFocus(items, cur + 1)
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    moveFocus(items, cur - 1)
  } else if (e.key === 'Escape') {
    e.preventDefault()
    close()
    focusTopBtn(idx)
  } else if (e.key === 'ArrowRight') {
    e.preventDefault()
    if (idx + 1 < MENUS.length) toggle(idx + 1)
  } else if (e.key === 'ArrowLeft') {
    e.preventDefault()
    if (idx > 0) toggle(idx - 1)
  } else if (e.key === 'Enter') {
    e.preventDefault()
    const cur2 = items.findIndex((x: Item) => x.label === (e.target as HTMLElement)?.dataset?.label)
    if (cur2 >= 0) clickItem(items[cur2])
  }
}
function moveFocus(items: Item[], next: number) {
  const wrap = (n: number) => (n + items.length) % items.length
  const target = items[wrap(next)]
  const el = document.querySelector(`[data-menu-idx="${openIdx.value}"] [data-label="${CSS.escape(target.label || '')}"]`) as HTMLElement
  el?.focus()
}
function focusTopBtn(idx: number) {
  const el = document.querySelector(`[data-top-idx="${idx}"]`) as HTMLElement
  el?.focus()
}

// ---- 外部点击关闭 ----
function onDocClick(e: MouseEvent) {
  if (!(e.target as HTMLElement).closest('.mb-wrap')) close()
}

onMounted(() => {
  document.addEventListener('click', onDocClick)
})
onBeforeUnmount(() => {
  document.removeEventListener('click', onDocClick)
})

// ---- 子菜单内容 ----
const themes = ['default', 'light', 'dark', 'monokai', 'solarized-light', 'solarized-dark', 'github-light', 'github-dark']
function itemsForSub(id: string): Item[] {
  switch (id) {
    case 'zoom':     return [{ label: '放大', cmd: 'zoom-in' }, { label: '缩小', cmd: 'zoom-out' }, { label: '重置', cmd: 'zoom-reset' }]
    case 'iconsize': return [{ label: '14 px', cmd: 'iconsize-14' }, { label: '16 px', cmd: 'iconsize-16' }, { label: '18 px', cmd: 'iconsize-18' }]
    case 'theme':    return themes.map(t => ({ label: t, cmd: 'theme-style' }))
    case 'uilang':   return [{ label: '简体中文', cmd: 'lang-zh' }, { label: 'English', cmd: 'lang-en' }]
    case 'fav':      return [{ label: '管理收藏夹...', cmd: 'manage-fav' }, { label: '清空收藏夹', cmd: 'clear-favorites' }]
    case 'recent':   return recent.value.length
        ? recent.value.map(r => ({ label: r.name || r.path, cmd: 'recent-open-' + r.path }))
        : [{ label: '暂无记录', cmd: 'recent-empty' } as Item]
    default:         return []
  }
}

// 当前打开菜单的 items（用于键盘聚焦）
const openItems = computed(() => {
  if (openIdx.value === null) return [] as Item[]
  return MENUS[openIdx.value].items.filter((x: Item) => !x.sep) as Item[]
})
</script>

<template>
  <div class="et-chrome-row et-chrome-menu mb-wrap" @keydown="onMenuKey">
    <div
      v-for="(m, idx) in MENUS"
      :key="m.label"
      class="mb-top-item"
    >
      <button
        class="mb-top"
        :class="{ 'is-open': openIdx === idx }"
        :data-top-idx="idx"
        tabindex="0"
        @click.stop="toggle(idx)"
        @mouseenter="onHover(idx)"
      >{{ showLabel(m.label) }}</button>

      <div v-if="openIdx === idx" class="mb-menu" :data-menu-idx="idx">
        <template v-if="m.label.startsWith('语言')">
          <div v-for="(langs, g) in LANG_GROUPS" :key="g" class="mb-lang-group">
            <div class="mb-lang-head">{{ g }}</div>
            <button
              v-for="lang in langs"
              :key="lang"
              class="mb-item"
              :class="{ 'is-active': ed.activeTab?.language?.toLowerCase() === lang.toLowerCase() }"
              @click="pickLang(lang)"
            >
              <span class="mb-item-chk">
                <Check v-if="ed.activeTab?.language?.toLowerCase() === lang.toLowerCase()" :size="12" :stroke-width="2" />
              </span>
              <span class="mb-item-label">{{ lang }}</span>
            </button>
          </div>
        </template>

        <template v-else>
          <template v-for="(it, j) in m.items" :key="j">
            <div v-if="it.sep" class="mb-sep" />
            <button
              v-else
              class="mb-item"
              :class="{ 'is-checked': isChecked(it), 'is-disabled': it.cmd && NOT_IMPLEMENTED.has(it.cmd) }"
              :data-label="it.label"
              @click="clickItem(it)"
              @mouseenter="it.sub && openSub(it.sub, $event)"
            >
              <span class="mb-item-chk">
                <Check v-if="isChecked(it)" :size="12" :stroke-width="2" />
              </span>
              <span class="mb-item-label">{{ showLabel(it.label) }}</span>
              <span v-if="it.sub" class="mb-item-arrow">
                <ChevronRight :size="12" :stroke-width="1.6" />
              </span>
              <span v-else-if="it.key" class="mb-item-key">{{ it.key }}</span>
            </button>

            <!-- 子菜单：浮动定位 -->
            <div
              v-if="it.sub && subId === it.sub"
              class="mb-sub"
              :class="{ 'mb-sub-flip': subFlip }"
              :data-sub="it.sub"
            >
              <button
                v-for="(sub, k) in itemsForSub(it.sub)"
                :key="k"
                class="mb-item"
                @click="
                  it.sub === 'zoom' ? pickZoom(sub.cmd === 'zoom-in' ? 110 : sub.cmd === 'zoom-out' ? 90 : 100) :
                  it.sub === 'iconsize' ? pickIconSize(parseInt((sub.cmd || '').split('-')[1] || '18', 10)) :
                  it.sub === 'theme' ? pickTheme(sub.label || '') :
                  it.sub === 'uilang' ? pickUiLang((sub.cmd || '').endsWith('zh') ? 'zh' : 'en') :
                  clickItem(sub)
                "
              >
                <span class="mb-item-chk" />
                <span class="mb-item-label">{{ sub.label }}</span>
                <span v-if="sub.key" class="mb-item-key">{{ sub.key }}</span>
              </button>
            </div>
          </template>
        </template>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 顶层菜单项 wrapper：必须 position:relative，
   否则 .mb-menu 的 absolute(top:100%) 会相对视口定位，
   下拉面板渲染到屏幕外——表现为"点击菜单毫无反应"（M5 用户实测） */
.mb-top-item {
  position: relative;
  height: 100%;
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.mb-top {
  height: 100%;
  padding: 0 10px;
  background: transparent;
  border: 0;
  color: var(--et-fg);
  font-size: var(--et-text-md);
  cursor: pointer;
  border-radius: 0;
  transition: background-color 80ms ease;
  -webkit-user-select: none;
  user-select: none;
}
.mb-top:hover,
.mb-top.is-open {
  background-color: var(--et-bg-hover);
}
.mb-top:focus-visible {
  outline: 2px solid var(--et-accent);
  outline-offset: -2px;
}

/* —— 下拉主菜单 —— */
.mb-menu {
  position: absolute;
  top: 100%;
  left: 0;
  z-index: 1100;
  min-width: 240px;
  padding: var(--et-space-1);
  background: var(--et-bg-elevated);
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius);
  box-shadow: var(--et-shadow-md);
  animation: mb-menu-in 80ms ease-out;
}
@keyframes mb-menu-in {
  from { opacity: 0; transform: translateY(-2px); }
  to   { opacity: 1; transform: translateY(0); }
}
.mb-item {
  display: flex;
  align-items: center;
  gap: var(--et-space-2);
  width: 100%;
  padding: 5px var(--et-space-2);
  border: 0;
  border-radius: var(--et-radius-sm);
  background: transparent;
  color: var(--et-fg);
  font-size: var(--et-text-md);
  text-align: left;
  cursor: pointer;
  white-space: nowrap;
  transition: background-color 80ms ease;
}
.mb-item:hover:not(.is-disabled),
.mb-item:focus-visible {
  background-color: var(--et-bg-hover);
  outline: none;
}
.mb-item.is-disabled {
  color: var(--et-fg-subtle);
  cursor: not-allowed;
}
.mb-item.is-checked {
  color: var(--et-accent);
}

.mb-item-chk {
  display: inline-flex;
  width: 14px;
  align-items: center;
  justify-content: center;
  color: var(--et-accent);
  flex-shrink: 0;
}
.mb-item-label {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
}
.mb-item-key {
  color: var(--et-fg-subtle);
  font-size: var(--et-text-xs);
  font-variant-numeric: tabular-nums;
  margin-left: var(--et-space-3);
}
.mb-item-arrow {
  color: var(--et-fg-subtle);
  display: inline-flex;
  align-items: center;
}

.mb-sep {
  height: 1px;
  margin: var(--et-space-1) var(--et-space-2);
  background-color: var(--et-border);
}

/* —— 子菜单（嵌套弹层） —— */
.mb-sub {
  position: absolute;
  left: calc(100% + var(--et-space-1));
  top: -5px;
  z-index: 1100;
  min-width: 200px;
  padding: var(--et-space-1);
  background: var(--et-bg-elevated);
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius);
  box-shadow: var(--et-shadow-md);
  animation: mb-menu-in 80ms ease-out;
}
.mb-sub-flip {
  left: auto;
  right: calc(100% + var(--et-space-1));
}

/* —— 语言菜单分组 —— */
.mb-lang-group {
  display: grid;
  grid-template-columns: repeat(2, minmax(120px, 1fr));
  gap: 0;
}
.mb-lang-head {
  grid-column: 1 / -1;
  padding: 4px 10px;
  font-size: var(--et-text-xs);
  font-weight: var(--et-fw-semibold);
  color: var(--et-fg-subtle);
  text-transform: uppercase;
  letter-spacing: .04em;
}
</style>