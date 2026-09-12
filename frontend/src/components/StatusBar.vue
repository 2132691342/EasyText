<script lang="ts" setup>
/**
 * Status Bar v2.1
 *
 * 设计目标（对标 Notepad-- / Sublime Text）：
 *  - 高度 24px，文字 11px，分隔条 12px 高。
 *  - **新增修改指示**：左下"● 已修改"/"○ 已保存"，与 TabBar 圆点对齐。
 *  - 路径省略：超 60 字符中部 …，hover tooltip 完整路径。
 *  - 右键菜单改为"显示项勾选"原生体验。
 */
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useEditorStore, useSettingStore } from '@/stores'

const editorStore = useEditorStore()
const settingStore = useSettingStore()
const tab = computed(() => editorStore.activeTab)
const zoomLevel = computed(() => settingStore.config?.ui?.zoomLevel || 100)

const cursorPos = ref({ line: 1, column: 1, selectionLength: 0 })
function onCursorEvent(e: Event) {
  const d = (e as CustomEvent).detail
  if (d) cursorPos.value = { line: d.line || 1, column: d.column || 1, selectionLength: d.selectionLength || 0 }
}
onMounted(() => document.addEventListener('editor-cursor-pos', onCursorEvent))
onUnmounted(() => document.removeEventListener('editor-cursor-pos', onCursorEvent))

const totalLines = computed(() => tab.value ? tab.value.content.split('\n').length : 0)
const langLabel = computed(() => tab.value?.language || 'text')
const isDirty = computed(() => !!tab.value?.isDirty)

function changeZoom(delta: number) {
  if (!settingStore.config) return
  settingStore.config.ui.zoomLevel = Math.max(50, Math.min(200, zoomLevel.value + delta))
  settingStore.saveConfig()
}

// 路径中部省略
const filePathFull = computed(() => tab.value?.path || '')
const filePathShort = computed(() => {
  const p = filePathFull.value
  if (!p) return ''
  if (p.length <= 60) return p
  const head = p.slice(0, 24)
  const tail = p.slice(-32)
  return head + '…' + tail
})

const showContextMenu = ref(false)
const contextMenuPos = ref({ x: 0, y: 0 })

const statusBarItemDefs: { key: string; label: string }[] = [
  { key: 'zoom', label: '缩放' },
  { key: 'lang', label: '语言' },
  { key: 'cursor', label: '光标位置' },
  { key: 'lines', label: '总行数' },
  { key: 'lineEnding', label: '换行符' },
  { key: 'encoding', label: '编码' },
  { key: 'filePath', label: '文件路径' },
]

function isItemVisible(key: string): boolean {
  const items = settingStore.config?.ui?.statusBarItems
  if (!items) return true
  if (items[key] === undefined) return true
  return items[key]
}
function toggleItem(key: string) {
  settingStore.toggleStatusBarItem(key)
}
function onContextMenu(e: MouseEvent) {
  e.preventDefault()
  contextMenuPos.value = { x: e.clientX, y: e.clientY }
  showContextMenu.value = true
}
function closeContextMenu() {
  showContextMenu.value = false
}
onMounted(() => document.addEventListener('click', closeContextMenu))
onUnmounted(() => document.removeEventListener('click', closeContextMenu))
</script>

<template>
  <div class="et-chrome-row et-chrome-status statusbar select-none" @contextmenu="onContextMenu">
    <!-- 左：修改指示（与 TabBar 圆点统一） -->
    <span class="status-item" :title="isDirty ? '文件已修改，未保存' : '文件已保存'">
      <span class="et-dot" :class="isDirty ? 'et-dot-warn' : 'et-dot-success'" />
      <span class="ml-1">{{ isDirty ? '已修改' : '已保存' }}</span>
    </span>
    <span class="status-sep" />

    <template v-if="tab">
      <span v-if="isItemVisible('zoom')" class="status-item" title="点击 +10%；Ctrl + 鼠标滚轮可调" @click="changeZoom(10)">缩放 {{ zoomLevel }}%</span>
      <span v-if="isItemVisible('zoom')" class="status-sep" />

      <span v-if="isItemVisible('lang')" class="status-item">语言 {{ langLabel }}</span>
      <span v-if="isItemVisible('lang')" class="status-sep" />

      <span v-if="isItemVisible('cursor')" class="status-item">
        行 {{ cursorPos.line }} · 列 {{ cursorPos.column }}<span v-if="cursorPos.selectionLength > 0"> · 选 {{ cursorPos.selectionLength }}</span>
      </span>
      <span v-if="isItemVisible('cursor')" class="status-sep" />

      <span v-if="isItemVisible('lines')" class="status-item">共 {{ totalLines }} 行</span>
      <span v-if="isItemVisible('lines')" class="status-sep" />

      <select
        v-if="isItemVisible('lineEnding')"
        class="sbar-select"
        :value="tab.lineEnding"
        @change="(e: any) => editorStore.updateTabLineEnding(tab!.id, e.target.value)"
      >
        <option value="CRLF">Windows (CR LF)</option>
        <option value="LF">Unix (LF)</option>
        <option value="CR">Mac (CR)</option>
      </select>
      <span v-if="isItemVisible('lineEnding')" class="status-sep" />

      <select
        v-if="isItemVisible('encoding')"
        class="sbar-select sbar-select-encoding"
        :value="tab.encoding"
        @change="(e: any) => editorStore.updateTabEncoding(tab!.id, e.target.value)"
      >
        <option v-for="e in ['UTF-8','UTF-8-BOM','GBK','GB18030','Big5','UTF-16LE','UTF-16BE','Shift_JIS','EUC-JP','ISO-8859-1','Windows-1252']" :key="e" :value="e">{{ e }}</option>
      </select>

      <span class="flex-1" />

      <span
        v-if="isItemVisible('filePath')"
        class="status-item status-path"
        :title="filePathFull || '未保存'"
      >{{ filePathShort || '未保存' }}</span>
    </template>

    <template v-else>
      <span class="status-item">就绪</span>
      <span class="flex-1" />
    </template>
  </div>

  <!-- 右键：自定义状态栏显示项 -->
  <Teleport to="body">
    <div
      v-if="showContextMenu"
      class="context-menu statusbar-menu"
      :style="{ left: contextMenuPos.x + 'px', top: contextMenuPos.y + 'px' }"
      @click.stop
    >
      <div class="sbar-menu-head">状态栏显示项</div>
      <label
        v-for="item in statusBarItemDefs"
        :key="item.key"
        class="context-menu-item"
      >
        <input
          type="checkbox"
          :checked="isItemVisible(item.key)"
          @change="toggleItem(item.key)"
          class="sbar-check"
        />
        {{ item.label }}
      </label>
    </div>
  </Teleport>
</template>

<style scoped>
.statusbar {
  font-size: var(--et-text-xs);
  color: var(--et-fg-muted);
  padding: 0 var(--et-space-2);
  gap: 0;
}

.status-item {
  display: inline-flex;
  align-items: center;
  height: 100%;
  padding: 0 var(--et-space-2);
  border-radius: var(--et-radius-sm);
  cursor: default;
  transition: background-color 80ms ease, color 80ms ease;
  white-space: nowrap;
}
.status-item[title]:hover,
.status-item:hover:has(+ .status-sep) {
  background: var(--et-bg-hover);
  color: var(--et-fg);
}
/* 仅可点击项给手指 */
button.status-item,
select.status-item {
  cursor: pointer;
}

.status-sep {
  display: inline-block;
  width: 1px;
  height: 12px;
  background: var(--et-border);
  flex-shrink: 0;
}

.sbar-select {
  height: 18px;
  margin: 0 var(--et-space-1);
  padding: 0 var(--et-space-1);
  font-size: var(--et-text-xs);
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--et-radius-sm);
  color: inherit;
  outline: none;
  cursor: pointer;
  transition: background-color 80ms ease, border-color 80ms ease;
}
.sbar-select-encoding { min-width: 110px; }
.sbar-select:hover {
  background: var(--et-bg-hover);
  border-color: var(--et-border);
}
.sbar-select:focus-visible {
  border-color: var(--et-accent);
}

.status-path {
  max-width: 50%;
  overflow: hidden;
  text-overflow: ellipsis;
}

.statusbar-menu { min-width: 180px; }
.sbar-menu-head {
  padding: 4px 10px;
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: .04em;
  color: var(--et-fg-subtle);
  font-weight: var(--et-fw-medium);
}
.sbar-check {
  width: 12px;
  height: 12px;
  margin-right: var(--et-space-2);
  accent-color: var(--et-accent);
  cursor: pointer;
}
</style>