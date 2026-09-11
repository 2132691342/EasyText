<script lang="ts" setup>
/**
 * 工具栏（Toolbar）
 *
 * 设计目标（对标 Notepad-- / Sublime Text 的紧凑工具条）：
 *  1. 数据驱动：命令在一处声明，渲染/折叠/「更多」菜单共用同一份数据，
 *     避免此前「模板里逐个写 button」导致的重复项（两个 X、两个放大镜）与漏接线。
 *  2. 尺寸自适应：按钮尺寸随 iconSize 联动（16→26px、20→30px、24→34px），
 *     修掉此前按钮固定 28px、图标放大后溢出/裁切的显示问题。
 *  3. 溢出折叠：窗口变窄时尾部按钮自动收进「更多」下拉，不再横向撑破布局。
 *  4. 状态可见：开关类按钮（自动换行/显示空白/自动保存/日志跟踪）直接反映真实状态，
 *     不再使用与配置脱节的本地 ref（此前重启后开关显示与实际行为不一致）。
 */
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useEditorStore, useSettingStore } from '@/stores'
import {
  FileText, FolderOpen, Save, SaveAll, X, XCircle, Radio,
  Scissors, Copy, Clipboard, Undo2, Redo2,
  Search, Replace, FileSearch,
  WrapText, Eye, ZoomIn, ZoomOut, RefreshCw,
  Braces, GitCompare, TestTube, FileCode, Palette,
  Printer, Maximize, Settings, Sun, Moon, ChevronDown,
} from 'lucide-vue-next'

const props = withDefaults(defineProps<{
  /** 图标像素尺寸：16 / 20 / 24（来自设置 → 视图 → 图标大小） */
  iconSize?: number
  /** tail -f 跟踪中（由 MainLayout 的 useTailWatcher 提供真实状态） */
  tailing?: boolean
}>(), { iconSize: 18, tailing: false })

const emit = defineEmits<{ (e: 'toolbar-command', cmd: string, ...a: any[]): void }>()

const ed = useEditorStore()
const se = useSettingStore()

const hasTab = computed(() => ed.activeTab !== null)
const dirty = computed(() => !!ed.activeTab?.isDirty)

/** 自动保存开关直接绑定配置，避免本地 ref 与真实配置脱节 */
const autoSaveOn = computed(() => !!se.config?.editor?.autoSave)
/** 自动换行开关同样取自配置（CodeEditor 的 toggle-whitespace 会写回配置） */
const wrapOn = computed(() => !!se.config?.editor?.wordWrap)
const wsOn = computed(() => !!se.config?.editor?.showWhitespace)

/** 按钮边长跟随图标尺寸，保证图标永远居中不溢出 */
const btnSize = computed(() => (props.iconSize <= 16 ? 26 : props.iconSize <= 20 ? 30 : 34))

interface TItem {
  cmd: string
  icon: unknown
  title: string
  /** 需要打开文档才可用 */
  needTab?: boolean
  /** 开关类按钮的激活态 */
  active?: () => boolean
}

const GROUPS: { id: string; items: TItem[] }[] = [
  {
    id: 'file',
    items: [
      { cmd: 'new-file', icon: FileText, title: '新建 (Ctrl+T)' },
      { cmd: 'open-file', icon: FolderOpen, title: '打开 (Ctrl+O)' },
      { cmd: 'save', icon: Save, title: '保存 (Ctrl+S)', needTab: true },
      { cmd: 'save-all', icon: SaveAll, title: '全部保存 (Ctrl+Alt+S)', needTab: true },
      { cmd: 'toggle-auto-save-cycle', icon: Radio, title: '循环自动保存', active: () => autoSaveOn.value },
      { cmd: 'close-tab', icon: X, title: '关闭 (Ctrl+W)', needTab: true },
      { cmd: 'close-all', icon: XCircle, title: '关闭全部 (Ctrl+Shift+W)', needTab: true },
    ],
  },
  {
    id: 'edit',
    items: [
      { cmd: 'undo', icon: Undo2, title: '撤销 (Ctrl+Z)', needTab: true },
      { cmd: 'redo', icon: Redo2, title: '重做 (Ctrl+Y)', needTab: true },
      { cmd: 'cut', icon: Scissors, title: '剪切 (Ctrl+X)', needTab: true },
      { cmd: 'copy', icon: Copy, title: '复制 (Ctrl+C)', needTab: true },
      { cmd: 'paste', icon: Clipboard, title: '粘贴 (Ctrl+V)', needTab: true },
    ],
  },
  {
    id: 'find',
    items: [
      { cmd: 'find', icon: Search, title: '查找 (Ctrl+F)' },
      { cmd: 'replace', icon: Replace, title: '替换 (Ctrl+H)', needTab: true },
      { cmd: 'search-files', icon: FileSearch, title: '在文件中查找 (Ctrl+Shift+F)' },
    ],
  },
  {
    id: 'view',
    items: [
      { cmd: 'zoom-out', icon: ZoomOut, title: '缩小 (Ctrl+-)' },
      { cmd: 'zoom-in', icon: ZoomIn, title: '放大 (Ctrl+=)' },
      { cmd: 'toggle-wrap', icon: WrapText, title: '自动换行', active: () => wrapOn.value },
      { cmd: 'toggle-whitespace', icon: Eye, title: '显示空格/制表符', active: () => wsOn.value },
      { cmd: 'toggle-tail', icon: RefreshCw, title: '跟踪文件尾部 (tail -f)', active: () => props.tailing },
    ],
  },
  {
    id: 'tools',
    items: [
      { cmd: 'format-json', icon: Braces, title: '格式化 JSON', needTab: true },
      { cmd: 'open-diff', icon: GitCompare, title: '文档对比' },
      { cmd: 'regex-tester', icon: TestTube, title: '正则测试' },
      { cmd: 'script-manager', icon: FileCode, title: '脚本管理器' },
      { cmd: 'color-picker', icon: Palette, title: '取色器' },
    ],
  },
]

const RIGHT_ITEMS: TItem[] = [
  { cmd: 'print', icon: Printer, title: '打印 (Ctrl+P)', needTab: true },
  { cmd: 'fullscreen', icon: Maximize, title: '全屏 (F11)' },
  { cmd: 'preferences', icon: Settings, title: '设置' },
]

/** 用户可在「设置 → 工具栏」里关掉单品；toolbarItems 为空表示全部显示 */
function enabled(it: TItem) {
  const map = se.config?.ui?.toolbarItems || {}
  return map[it.cmd] !== false
}

const flat = computed(() => GROUPS.flatMap(g => g.items).filter(enabled))

// ---------------- 溢出折叠 ----------------
const wrapRef = ref<HTMLElement | null>(null)
const rightRef = ref<HTMLElement | null>(null)
const capacity = ref(flat.value.length)
let ro: ResizeObserver | null = null

function recalc() {
  const wrapW = wrapRef.value?.clientWidth ?? 0
  if (!wrapW) return
  const rightW = rightRef.value?.offsetWidth ?? 0
  const per = btnSize.value + 3          // 按钮 + 左右 margin
  const sepW = GROUPS.length * 9         // 分组分隔条占宽
  const moreW = btnSize.value + 12       // 「更多」按钮（含箭头）
  const pad = 12                         // 容器内边距
  const avail = wrapW - rightW - sepW - moreW - pad
  capacity.value = Math.max(3, Math.floor(avail / per))
}

const visible = computed(() => flat.value.slice(0, capacity.value))
const hidden = computed(() => flat.value.slice(capacity.value))

/** flat 中属于「分组末尾」的下标：渲染时在这些按钮后面插入分隔条 */
const groupEnds = computed(() => {
  const ends = new Set<number>()
  let acc = 0
  for (const g of GROUPS) {
    acc += g.items.filter(enabled).length
    if (acc > 0) ends.add(acc - 1)
  }
  ends.delete(flat.value.length - 1) // 最后一项后面不加分隔条
  return ends
})

const showMore = ref(false)
function toggleMore(e: MouseEvent) {
  e.stopPropagation()
  showMore.value = !showMore.value
}
function onDocClick(e: MouseEvent) {
  if (!(e.target as HTMLElement).closest('.tb-more-wrap')) showMore.value = false
}

onMounted(() => {
  recalc()
  if (typeof ResizeObserver !== 'undefined' && wrapRef.value) {
    ro = new ResizeObserver(() => recalc())
    ro.observe(wrapRef.value)
  }
  window.addEventListener('resize', recalc)
  document.addEventListener('click', onDocClick)
})
onBeforeUnmount(() => {
  ro?.disconnect()
  window.removeEventListener('resize', recalc)
  document.removeEventListener('click', onDocClick)
})

// ---------------- 交互 ----------------
function run(it: TItem) {
  emit('toolbar-command', it.cmd)
}

/** 开关类命令需要把目标状态一并传出去（tail / 自动保存都是 (on: boolean) 签名） */
function runToggle(it: TItem) {
  const next = !(it.active?.() ?? false)
  emit('toolbar-command', it.cmd, next)
}

function isToggle(it: TItem) {
  return it.cmd === 'toggle-tail' || it.cmd === 'toggle-auto-save-cycle'
}

function disabled(it: TItem) {
  return !!it.needTab && !hasTab.value
}

function onClick(it: TItem) {
  showMore.value = false
  if (disabled(it)) return
  if (isToggle(it)) runToggle(it)
  else run(it)
}

function toggleTheme() {
  se.toggleTheme()
}
</script>

<template>
  <div ref="wrapRef" class="toolbar" :style="{ height: (btnSize + 8) + 'px' }">
    <template v-for="(it, i) in visible" :key="it.cmd">
      <button
        class="tb"
        :class="{ 'tb-on': it.active?.(), 'tb-dirty': it.cmd === 'save' && dirty }"
        :style="{ width: btnSize + 'px', height: btnSize + 'px' }"
        :title="it.title"
        :disabled="disabled(it)"
        @click="onClick(it)"
      >
        <component :is="it.icon" :size="iconSize" :stroke-width="1.8" />
      </button>
      <span v-if="groupEnds.has(i)" class="sep" :style="{ height: Math.round(btnSize * 0.55) + 'px' }" />
    </template>

    <!-- 溢出：「更多」 -->
    <div v-if="hidden.length" class="tb-more-wrap">
      <button
        class="tb tb-more"
        :style="{ width: (btnSize + 8) + 'px', height: btnSize + 'px' }"
        title="更多命令"
        @click="toggleMore"
      >
        <component :is="ChevronDown" :size="iconSize - 2" :stroke-width="2" />
      </button>
      <div v-if="showMore" class="tb-more-menu">
        <button
          v-for="it in hidden"
          :key="it.cmd"
          class="tb-more-item"
          :class="{ 'tb-more-item-on': it.active?.() }"
          :disabled="disabled(it)"
          @click="onClick(it)"
        >
          <component :is="it.icon" :size="16" :stroke-width="1.8" />
          <span class="tb-more-label">{{ it.title }}</span>
        </button>
      </div>
    </div>

    <div class="tb-spacer" />

    <div ref="rightRef" class="tb-right">
      <button
        v-for="it in RIGHT_ITEMS"
        :key="it.cmd"
        class="tb"
        :style="{ width: btnSize + 'px', height: btnSize + 'px' }"
        :title="it.title"
        :disabled="disabled(it)"
        @click="onClick(it)"
      >
        <component :is="it.icon" :size="iconSize" :stroke-width="1.8" />
      </button>
      <span class="sep" />
      <button
        class="tb"
        :style="{ width: btnSize + 'px', height: btnSize + 'px' }"
        :title="se.isDarkMode ? '切换到亮色主题' : '切换到暗色主题'"
        @click="toggleTheme"
      >
        <component :is="se.isDarkMode ? Sun : Moon" :size="iconSize" :stroke-width="1.8" />
      </button>
    </div>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 0;
  padding: 0 6px;
  background: var(--et-bg-sunken);
  border-bottom: 1px solid var(--et-border);
  user-select: none;
  overflow: hidden;
  flex-shrink: 0;
}

.tb {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin: 0 1.5px;
  padding: 0;
  border: none;
  border-radius: var(--et-radius-sm);
  background: transparent;
  color: var(--et-fg-muted);
  cursor: pointer;
  flex-shrink: 0;
  transition: background .12s ease, color .12s ease;
}
.tb:hover:not(:disabled) {
  background: var(--et-bg-hover);
  color: var(--et-fg);
}
.tb:active:not(:disabled) {
  background: var(--et-bg-active);
}
.tb:disabled {
  opacity: .3;
  cursor: default;
}
.tb:focus-visible {
  outline: 2px solid var(--et-accent);
  outline-offset: -2px;
}
/* 开关激活态 */
.tb-on {
  background: var(--et-accent-soft);
  color: var(--et-accent);
}
/* 有未保存改动 */
.tb-dirty {
  color: var(--et-accent);
}

.sep {
  width: 1px;
  background: var(--et-border);
  margin: 0 4px;
  flex-shrink: 0;
}

.tb-spacer {
  flex: 1 1 auto;
  min-width: 8px;
}

.tb-right {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

/* ---- 更多 ---- */
.tb-more-wrap {
  position: relative;
  flex-shrink: 0;
}
.tb-more-menu {
  position: absolute;
  right: 0;
  top: calc(100% + 4px);
  z-index: 900;
  min-width: 200px;
  max-height: 60vh;
  overflow-y: auto;
  padding: 4px;
  background: var(--et-bg-elevated);
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius);
  box-shadow: var(--et-shadow-md);
}
.tb-more-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 6px 10px;
  border: none;
  border-radius: var(--et-radius-sm);
  background: transparent;
  color: var(--et-fg);
  font-size: 13px;
  text-align: left;
  cursor: pointer;
  white-space: nowrap;
}
.tb-more-item:hover:not(:disabled) {
  background: var(--et-bg-hover);
}
.tb-more-item:disabled {
  opacity: .35;
  cursor: default;
}
.tb-more-item-on {
  color: var(--et-accent);
  background: var(--et-accent-soft);
}
.tb-more-label {
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
