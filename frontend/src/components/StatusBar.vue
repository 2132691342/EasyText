<script lang="ts" setup>
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

function changeZoom(delta: number) {
  if (!settingStore.config) return
  const n = Math.max(50, Math.min(200, zoomLevel.value + delta))
  settingStore.config.ui.zoomLevel = n
  settingStore.config.editor.fontSize = Math.round((n / 100) * 14)
  settingStore.saveConfig()
}

// 🆕 V2.0.0 状态栏右键自定义菜单
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
  // 默认全部显示
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

onMounted(() => {
  document.addEventListener('click', closeContextMenu)
})
onUnmounted(() => {
  document.removeEventListener('click', closeContextMenu)
})
</script>

<template>
  <div v-if="tab" class="ndd-statusbar flex items-center h-6 border-t border-gray-300 dark:border-gray-700 bg-[#f0f0f0] dark:bg-[#007acc] text-[11px] text-gray-700 dark:text-white select-none" @contextmenu="onContextMenu">
    <span v-if="isItemVisible('zoom')" class="status-item px-2 cursor-pointer hover:bg-blue-100 dark:hover:bg-blue-800" @click="changeZoom(10)">Zoom: {{ zoomLevel }}%</span>
    <span v-if="isItemVisible('zoom')" class="status-sep" />
    <span v-if="isItemVisible('lang')" class="status-item px-2">Lang: {{ langLabel }}</span>
    <span v-if="isItemVisible('lang')" class="status-sep" />
    <span v-if="isItemVisible('cursor')" class="status-item px-2">Ln: {{ cursorPos.line }} &nbsp; Col: {{ cursorPos.column }} &nbsp; Sel: {{ cursorPos.selectionLength }}</span>
    <span v-if="isItemVisible('cursor')" class="status-sep" />
    <span v-if="isItemVisible('lines')" class="status-item px-2">{{ totalLines }} lines</span>
    <span v-if="isItemVisible('lines')" class="status-sep" />
    <select v-if="tab && isItemVisible('lineEnding')" class="sbar-select w-[120px]" :value="tab.lineEnding" @change="(e: any) => editorStore.updateTabLineEnding(tab!.id, e.target.value)">
      <option value="CRLF">Windows (CR LF)</option>
      <option value="LF">Unix (LF)</option>
      <option value="CR">Mac (CR)</option>
    </select>
    <span v-if="isItemVisible('lineEnding')" class="status-sep" />
    <select v-if="tab && isItemVisible('encoding')" class="sbar-select w-[100px]" :value="tab.encoding" @change="(e: any) => editorStore.updateTabEncoding(tab!.id, e.target.value)">
      <option v-for="e in ['UTF-8','UTF-8-BOM','GBK','GB18030','Big5','UTF-16LE','UTF-16BE','Shift_JIS','EUC-JP','ISO-8859-1','Windows-1252']" :key="e" :value="e">{{ e }}</option>
    </select>
    <span class="flex-1" />
    <span v-if="tab.path && isItemVisible('filePath')" class="status-item px-2 truncate max-w-[50%]" :title="tab.path">{{ tab.path }}</span>
    <span v-else-if="isItemVisible('filePath')" class="status-item px-2 text-gray-500 dark:text-gray-300">untitled</span>
  </div>
  <div v-else class="ndd-statusbar flex items-center h-6 border-t border-gray-300 dark:border-gray-700 bg-[#f0f0f0] dark:bg-[#007acc] text-[11px] text-gray-600 dark:text-white" @contextmenu="onContextMenu">
    <span class="px-3">Ready</span>
  </div>

  <!-- 🆕 V2.0.0 右键自定义菜单 -->
  <Teleport to="body">
    <div
      v-if="showContextMenu"
      class="fixed z-[100] bg-white dark:bg-[#2d2d2d] border border-gray-200 dark:border-gray-600 rounded shadow-lg py-1 min-w-[160px] text-xs"
      :style="{ left: contextMenuPos.x + 'px', top: contextMenuPos.y + 'px' }"
      @click.stop
    >
      <div class="px-3 py-1 text-gray-400 text-[10px] uppercase tracking-wide">状态栏显示项</div>
      <label
        v-for="item in statusBarItemDefs"
        :key="item.key"
        class="flex items-center gap-2 px-3 py-1.5 cursor-pointer hover:bg-gray-100 dark:hover:bg-[#3c3c3c] text-gray-700 dark:text-gray-200"
      >
        <input
          type="checkbox"
          :checked="isItemVisible(item.key)"
          @change="toggleItem(item.key)"
          class="w-3 h-3"
        />
        {{ item.label }}
      </label>
    </div>
  </Teleport>
</template>

<style scoped>
.status-item { white-space: nowrap; transition: background 0.1s; }
.status-sep { width: 1px; height: 14px; background: rgba(0,0,0,0.12); flex-shrink: 0; }
html.dark .status-sep { background: rgba(255,255,255,0.2); }
.sbar-select {
  min-height: 18px; height: 18px; padding: 0 4px; font-size: 11px;
  background: transparent; border: 1px solid rgba(0,0,0,0.1); border-radius: 2px;
  color: inherit; outline: none; cursor: pointer;
}
html.dark .sbar-select { border-color: rgba(255,255,255,0.15); }
</style>
