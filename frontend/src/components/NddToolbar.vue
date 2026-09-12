<script lang="ts" setup>
/**
 * 工具栏（Toolbar）
 *
 *  - 数据驱动：命令在一处声明，渲染/折叠/「更多」菜单共用同一份数据。
 *  - 固定高度 30px，按钮 26×26px；iconSize 仅影响图标像素。
 *  - 即时 Popover：hover 200ms 弹出 label + shortcut，替代 native title 延迟。
 *  - 溢出折叠：窗口变窄时尾部按钮自动收进「更多」下拉。
 *  - 开关类按钮激活态直接绑定 store/config。
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
  /** 图标像素尺寸：14 / 16 / 18（来自设置 → 视图 → 图标大小） */
  iconSize?: number
  /** tail -f 跟踪中（由 MainLayout 的 useTailWatcher 提供真实状态） */
  tailing?: boolean
}>(), { iconSize: 14, tailing: false })

const emit = defineEmits<{ (e: 'toolbar-command', cmd: string, ...a: any[]): void }>()

const ed = useEditorStore()
const se = useSettingStore()

const hasTab = computed(() => ed.activeTab !== null)
const dirty = computed(() => !!ed.activeTab?.isDirty)

const autoSaveOn = computed(() => !!se.config?.editor?.autoSave)
const wrapOn    = computed(() => !!se.config?.editor?.wordWrap)
const wsOn      = computed(() => !!se.config?.editor?.showWhitespace)

interface TItem {
  cmd: string
  icon: unknown
  /** Popover 主标签（不含快捷键） */
  label: string
  /** 快捷键描述，例 "Ctrl+S"。空字符串则不显示 */
  key?: string
  /** 需要打开文档才可用 */
  needTab?: boolean
  /** 开关类按钮的激活态 */
  active?: () => boolean
}

const GROUPS: { id: string; items: TItem[] }[] = [
  {
    id: 'file',
    items: [
      { cmd: 'new-file', icon: FileText, label: '新建', key: 'Ctrl+T' },
      { cmd: 'open-file', icon: FolderOpen, label: '打开', key: 'Ctrl+O' },
      { cmd: 'save', icon: Save, label: '保存', key: 'Ctrl+S', needTab: true },
      { cmd: 'save-all', icon: SaveAll, label: '全部保存', key: 'Ctrl+Alt+S', needTab: true },
      { cmd: 'toggle-auto-save-cycle', icon: Radio, label: '循环自动保存', active: () => autoSaveOn.value },
      { cmd: 'close-tab', icon: X, label: '关闭', key: 'Ctrl+W', needTab: true },
      { cmd: 'close-all', icon: XCircle, label: '关闭全部', key: 'Ctrl+Shift+W', needTab: true },
    ],
  },
  {
    id: 'edit',
    items: [
      { cmd: 'undo', icon: Undo2, label: '撤销', key: 'Ctrl+Z', needTab: true },
      { cmd: 'redo', icon: Redo2, label: '重做', key: 'Ctrl+Y', needTab: true },
      { cmd: 'cut', icon: Scissors, label: '剪切', key: 'Ctrl+X', needTab: true },
      { cmd: 'copy', icon: Copy, label: '复制', key: 'Ctrl+C', needTab: true },
      { cmd: 'paste', icon: Clipboard, label: '粘贴', key: 'Ctrl+V', needTab: true },
    ],
  },
  {
    id: 'find',
    items: [
      { cmd: 'find', icon: Search, label: '查找', key: 'Ctrl+F' },
      { cmd: 'replace', icon: Replace, label: '替换', key: 'Ctrl+H', needTab: true },
      { cmd: 'search-files', icon: FileSearch, label: '在文件中查找', key: 'Ctrl+Shift+F' },
    ],
  },
  {
    id: 'view',
    items: [
      { cmd: 'zoom-out', icon: ZoomOut, label: '缩小', key: 'Ctrl+-' },
      { cmd: 'zoom-in', icon: ZoomIn, label: '放大', key: 'Ctrl+=' },
      { cmd: 'toggle-wrap', icon: WrapText, label: '自动换行', active: () => wrapOn.value },
      { cmd: 'toggle-whitespace', icon: Eye, label: '显示空格 / 制表符', active: () => wsOn.value },
      { cmd: 'toggle-tail', icon: RefreshCw, label: '跟踪文件尾部 (tail -f)', active: () => props.tailing },
    ],
  },
  {
    id: 'tools',
    items: [
      { cmd: 'format-json', icon: Braces, label: '格式化 JSON', needTab: true },
      { cmd: 'open-diff', icon: GitCompare, label: '文档对比' },
      { cmd: 'regex-tester', icon: TestTube, label: '正则测试' },
      { cmd: 'script-manager', icon: FileCode, label: '脚本管理器' },
      { cmd: 'color-picker', icon: Palette, label: '取色器' },
    ],
  },
]

const RIGHT_ITEMS: TItem[] = [
  { cmd: 'print', icon: Printer, label: '打印', key: 'Ctrl+P', needTab: true },
  { cmd: 'fullscreen', icon: Maximize, label: '全屏', key: 'F11' },
  { cmd: 'preferences', icon: Settings, label: '设置' },
]

/** 用户可在「设置 → 工具栏」里关掉单品；toolbarItems 为空表示全部显示 */
function enabled(it: TItem) {
  const map = se.config?.ui?.toolbarItems || {}
  return map[it.cmd] !== false
}

const flat = computed(() => GROUPS.flatMap(g => g.items).filter(enabled))

// ---------------- 溢出折叠 ----------------
const wrapRef  = ref<HTMLElement | null>(null)
const rightRef = ref<HTMLElement | null>(null)
const capacity = ref(flat.value.length)
let ro: ResizeObserver | null = null

/** Chrome 高度 / 按钮宽度 / 分隔条宽度均来自令牌 */
const TB_BTN = 26           // 与 .et-icon-btn width 等同
const TB_GAP = 2            // margin 0 1px × 2 边
const TB_SEP = 9            // 1px 分隔条 + 4+4 padding
const TB_MORE_BTN = 28      // 「更多」按钮略宽以容纳箭头
const TB_PAD = 12           // 容器内边距 6+6

function recalc() {
  const wrapW = wrapRef.value?.clientWidth ?? 0
  if (!wrapW) return
  const rightW = rightRef.value?.offsetWidth ?? 0
  const per = TB_BTN + TB_GAP
  const sepW = GROUPS.length * TB_SEP
  const moreW = TB_MORE_BTN + TB_PAD
  const avail = wrapW - rightW - sepW - moreW - TB_PAD
  capacity.value = Math.max(3, Math.floor(avail / per))
}

const visible = computed(() => flat.value.slice(0, capacity.value))
const hidden  = computed(() => flat.value.slice(capacity.value))

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

// ---------------- 即时 Popover（替代 native title） ----------------
interface PopState {
  it: TItem
  top: number
  left: number
}
const pop = ref<PopState | null>(null)
let popTimer: number | null = null

function showPop(e: MouseEvent | FocusEvent, it: TItem) {
  if (!it.key && !it.label) return
  if (popTimer) { window.clearTimeout(popTimer); popTimer = null }
  popTimer = window.setTimeout(() => {
    const target = e.currentTarget as HTMLElement
    if (!target) return
    const r = target.getBoundingClientRect()
    pop.value = {
      it,
      top: Math.round(r.top + r.height + 4),
      left: Math.round(r.left + r.width / 2),
    }
  }, 200)
}
function hidePop() {
  if (popTimer) { window.clearTimeout(popTimer); popTimer = null }
  pop.value = null
}
</script>

<template>
  <div ref="wrapRef" class="et-chrome-row et-chrome-tool">
    <template v-for="(it, i) in visible" :key="it.cmd">
      <button
        class="et-icon-btn"
        :class="{
          'is-on': it.active?.(),
          'is-dirty': it.cmd === 'save' && dirty,
        }"
        :disabled="disabled(it)"
        @click="onClick(it)"
        @mouseenter="showPop($event, it)"
        @mouseleave="hidePop"
        @focus="showPop($event, it)"
        @blur="hidePop"
      >
        <component :is="it.icon" :size="iconSize" :stroke-width="1.6" />
      </button>
      <span v-if="groupEnds.has(i)" class="tb-sep" />
    </template>

    <!-- 溢出：「更多」 -->
    <div v-if="hidden.length" class="tb-more-wrap">
      <button
        class="et-icon-btn tb-more-btn"
        title="更多命令"
        @click="toggleMore"
        @mouseenter="hidePop"
      >
        <component :is="ChevronDown" :size="iconSize" :stroke-width="1.6" />
      </button>
      <div v-if="showMore" class="tb-more-menu">
        <button
          v-for="it in hidden"
          :key="it.cmd"
          class="tb-more-item"
          :class="{ 'tb-more-item-on': it.active?.() }"
          :disabled="disabled(it)"
          @click="onClick(it)"
          @mouseenter="hidePop"
        >
          <component :is="it.icon" :size="14" :stroke-width="1.6" />
          <span class="tb-more-label">{{ it.label }}</span>
          <span v-if="it.key" class="tb-more-key">{{ it.key }}</span>
        </button>
      </div>
    </div>

    <div class="tb-spacer" />

    <div ref="rightRef" class="tb-right">
      <button
        v-for="it in RIGHT_ITEMS"
        :key="it.cmd"
        class="et-icon-btn"
        :disabled="disabled(it)"
        @click="onClick(it)"
        @mouseenter="showPop($event, it)"
        @mouseleave="hidePop"
        @focus="showPop($event, it)"
        @blur="hidePop"
      >
        <component :is="it.icon" :size="iconSize" :stroke-width="1.6" />
      </button>
      <span class="tb-sep" />
      <button
        class="et-icon-btn"
        :title="se.isDarkMode ? '切换到亮色主题' : '切换到暗色主题'"
        @click="toggleTheme"
        @mouseenter="hidePop"
      >
        <component :is="se.isDarkMode ? Sun : Moon" :size="iconSize" :stroke-width="1.6" />
      </button>
    </div>

    <!-- 即时 Popover：固定定位在按钮下方居中 -->
    <Teleport to="body">
      <div v-if="pop" class="et-popover" :style="{ top: pop.top + 'px', left: pop.left + 'px', transform: 'translateX(-50%)' }">
        <span>{{ pop.it.label }}</span>
        <span v-if="pop.it.key" class="et-popover-key">{{ pop.it.key }}</span>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.tb-sep {
  width: 1px;
  height: 18px;
  background: var(--et-border);
  margin: 0 4px;
  flex-shrink: 0;
}

.tb-spacer {
  flex: 1 1 auto;
  min-width: var(--et-space-1);
}

.tb-right {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

/* —— 脏文件标记：仅用于颜色提示，不依赖位置 —— */
.is-dirty {
  color: var(--et-warn);
}
.is-dirty:hover:not(:disabled) {
  color: var(--et-warn);
}

/* —— 「更多」按钮略宽，容纳箭头 —— */
.tb-more-wrap {
  position: relative;
  flex-shrink: 0;
}
.tb-more-btn {
  width: var(--et-h-control);
}

/* —— 「更多」下拉 —— */
.tb-more-menu {
  position: absolute;
  right: 0;
  top: calc(100% + var(--et-space-1));
  z-index: 900;
  min-width: 220px;
  max-height: 60vh;
  overflow-y: auto;
  padding: var(--et-space-1);
  background: var(--et-bg-elevated);
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius);
  box-shadow: var(--et-shadow-md);
  animation: tb-more-in 80ms ease-out;
}
@keyframes tb-more-in {
  from { opacity: 0; transform: translateY(-2px); }
  to   { opacity: 1; transform: translateY(0); }
}
.tb-more-item {
  display: flex;
  align-items: center;
  gap: var(--et-space-2);
  width: 100%;
  padding: var(--et-space-1) var(--et-space-2);
  border: none;
  border-radius: var(--et-radius-sm);
  background: transparent;
  color: var(--et-fg);
  font-size: var(--et-text-md);
  text-align: left;
  cursor: pointer;
  white-space: nowrap;
  transition: background-color 80ms ease;
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
  flex: 1;
}
.tb-more-key {
  color: var(--et-fg-subtle);
  font-size: var(--et-text-xs);
  font-variant-numeric: tabular-nums;
}
</style>