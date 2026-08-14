<script lang="ts" setup>
import { ref, nextTick, onMounted, onUnmounted } from 'vue'
import { useEditorStore, useFileStore } from '@/stores'
import {
  X, FileText, XCircle, XSquare, CheckCircle2, Clipboard, Pencil,
  ChevronsLeft, ChevronRight, ChevronsRight,
  Save, ExternalLink, FolderOpen, FileCode, Binary,
  GitCompare, Copy,
} from 'lucide-vue-next'
import { ElMessage } from 'element-plus'
import type { EditorTab } from '@/types'
import { ShowConfirmDialog, SaveFile, RenameFile, GetDirectoryTree, ShowMessageDialog, SaveFileDialog } from '../../wailsjs/go/main/App'

const editorStore = useEditorStore()
const fileStore = useFileStore()

// Context menu state
const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  tabId: '',
})

// Inline rename state
const renamingTabId = ref<string | null>(null)
const renameValue = ref('')
let renameInputEl: HTMLInputElement | null = null

// 🆕 V2.0.0 拖拽排序
const dragIndex = ref<number | null>(null)
const dragOverIndex = ref<number | null>(null)

function onDragStart(e: DragEvent, idx: number) {
  dragIndex.value = idx
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', String(idx))
  }
}

function onDragOver(e: DragEvent, idx: number) {
  e.preventDefault()
  if (dragIndex.value === null || dragIndex.value === idx) return
  dragOverIndex.value = idx
  if (e.dataTransfer) e.dataTransfer.dropEffect = 'move'
}

function onDragLeave() {
  dragOverIndex.value = null
}

function onDrop(e: DragEvent, idx: number) {
  e.preventDefault()
  if (dragIndex.value === null || dragIndex.value === idx) return
  const tabs = editorStore.tabs
  const [moved] = tabs.splice(dragIndex.value, 1)
  tabs.splice(idx, 0, moved)
  dragIndex.value = null
  dragOverIndex.value = null
}

function onDragEnd() {
  dragIndex.value = null
  dragOverIndex.value = null
}

// Function ref for v-for compatibility (Vue 3 collects refs in v-for into an array)
function setRenameRef(el: any) {
  renameInputEl = el as HTMLInputElement | null
}

async function closeTab(tab: EditorTab, e: MouseEvent) {
  e.stopPropagation()

  if (tab.isDirty) {
    const confirm = await ShowConfirmDialog('保存更改', `是否保存对 ${tab.name} 的更改？`)
    if (confirm) {
      await SaveFile(tab.path, tab.content, tab.encoding)
      editorStore.markTabSaved(tab.id)
    }
  }

  editorStore.closeTab(tab.id)
  closeContextMenu()
}

function selectTab(tab: EditorTab) {
  editorStore.activateTab(tab.id)
}

function getTabTitle(tab: EditorTab): string {
  return tab.name + (tab.isDirty ? ' •' : '')
}

// Right-click context menu
function handleContextMenu(e: MouseEvent, tab: EditorTab) {
  e.preventDefault()
  e.stopPropagation()
  contextMenu.value = {
    visible: true,
    x: e.clientX,
    y: e.clientY,
    tabId: tab.id,
  }
}

function closeContextMenu() {
  contextMenu.value.visible = false
}

// Context menu actions
function closeThisTab() {
  const tab = editorStore.tabs.find(t => t.id === contextMenu.value.tabId)
  if (tab) {
    closeTab(tab, new MouseEvent('click'))
  }
}

function closeOtherTabs() {
  closeContextMenu()
  const tabId = contextMenu.value.tabId
  editorStore.closeOtherTabs(tabId)
  ElMessage.success('已关闭其他标签页')
}

function closeLeftAllTabs() {
  closeContextMenu()
  const tabId = contextMenu.value.tabId
  const tabIndex = editorStore.tabs.findIndex(t => t.id === tabId)
  if (tabIndex === -1) return
  const tabsToClose = editorStore.tabs.slice(0, tabIndex)
  for (const tab of tabsToClose) {
    editorStore.closeTab(tab.id)
  }
  ElMessage.success(`已关闭左侧 ${tabsToClose.length} 个标签页`)
}

function closeNonCurrentTabs() {
  closeContextMenu()
  const tabId = contextMenu.value.tabId
  const currentTab = editorStore.tabs.find(t => t.id === tabId)
  if (!currentTab) return
  const otherTabs = editorStore.tabs.filter(t => t.id !== tabId)
  for (const tab of otherTabs) {
    editorStore.closeTab(tab.id)
  }
  editorStore.activateTab(tabId)
  ElMessage.success(`已关闭非当前 ${otherTabs.length} 个标签页`)
}

async function saveAsTab() {
  closeContextMenu()
  const tab = editorStore.tabs.find(t => t.id === contextMenu.value.tabId)
  if (!tab) return
  try {
    const path = await SaveFileDialog(tab.name)
    if (!path) return
    await SaveFile(path, tab.content, tab.encoding)
    ElMessage.success('另存为成功')
  } catch (e) {
    ElMessage.error(`另存为失败: ${e}`)
  }
}

function openInNewWindow() {
  closeContextMenu()
  ElMessage.info('新窗口打开功能开发中')
}

async function showInExplorer() {
  closeContextMenu()
  const tab = editorStore.tabs.find(t => t.id === contextMenu.value.tabId)
  if (!tab?.path) {
    ElMessage.warning('该标签无文件路径')
    return
  }
  try {
    const { BrowserOpenURL } = await import('../../wailsjs/runtime/runtime')
    const dir = tab.path.substring(0, Math.max(tab.path.lastIndexOf('\\'), tab.path.lastIndexOf('/')))
    BrowserOpenURL(`file:///${dir.replace(/\\/g, '/')}`)
  } catch {
    ElMessage.info(`目录: ${tab.path}`)
  }
}

async function reloadAsText() {
  closeContextMenu()
  const tab = editorStore.tabs.find(t => t.id === contextMenu.value.tabId)
  if (!tab?.path) return
  try {
    const { ReadFile } = await import('../../wailsjs/go/main/App')
    const result = await ReadFile(tab.path)
    if (result) {
      editorStore.updateTabContent(tab.id, result.content)
      ElMessage.success('已以文本模式重新加载')
    }
  } catch (e) {
    ElMessage.error(`重新加载失败: ${e}`)
  }
}

function reloadAsHex() {
  closeContextMenu()
  const tab = editorStore.tabs.find(t => t.id === contextMenu.value.tabId)
  if (!tab) return
  ElMessage.info(`十六进制模式重载: ${tab.name}（开发中）`)
}

function selectLeftCmpFile() {
  closeContextMenu()
  const tab = editorStore.tabs.find(t => t.id === contextMenu.value.tabId)
  if (!tab) return
  document.dispatchEvent(new CustomEvent('select-left-cmp-file', { detail: tab.path }))
  ElMessage.success(`已选择左侧对比文件: ${tab.name}`)
}

function selectRightCmpFile() {
  closeContextMenu()
  const tab = editorStore.tabs.find(t => t.id === contextMenu.value.tabId)
  if (!tab) return
  document.dispatchEvent(new CustomEvent('select-right-cmp-file', { detail: tab.path }))
  ElMessage.success(`已选择右侧对比文件: ${tab.name}`)
}

function closeTabsToRight() {
  closeContextMenu()
  const tabId = contextMenu.value.tabId
  const tabIndex = editorStore.tabs.findIndex(t => t.id === tabId)
  if (tabIndex === -1) return

  const tabsToClose = editorStore.tabs.slice(tabIndex + 1)
  for (const tab of tabsToClose) {
    editorStore.closeTab(tab.id)
  }
  ElMessage.success(`已关闭右侧 ${tabsToClose.length} 个标签页`)
}

function closeAllTabs() {
  closeContextMenu()
  editorStore.closeAllTabs()
  ElMessage.success('已关闭所有标签页')
}

async function closeSavedTabs() {
  closeContextMenu()
  const savedTabs = editorStore.tabs.filter(t => !t.isDirty)
  if (savedTabs.length === 0) {
    ElMessage.info('没有已保存的标签页')
    return
  }
  for (const tab of savedTabs) {
    editorStore.closeTab(tab.id)
  }
}

async function copyTabPath() {
  closeContextMenu()
  const tab = editorStore.tabs.find(t => t.id === contextMenu.value.tabId)
  if (tab?.path) {
    try {
      await navigator.clipboard.writeText(tab.path)
      ElMessage.success('路径已复制到剪贴板')
    } catch (error) {
      console.error('Failed to copy path:', error)
      ElMessage.error('复制路径失败')
    }
  }
}

// Double-click to rename tab (renames the actual file)
function startRenameFromMenu() {
  closeContextMenu()
  const tab = editorStore.tabs.find(t => t.id === contextMenu.value.tabId)
  if (tab) {
    startRename(tab)
  }
}

function startRename(tab: EditorTab) {
  if (!tab.path) return
  renamingTabId.value = tab.id
  renameValue.value = tab.name
  nextTick(() => {
    if (renameInputEl) {
      renameInputEl.focus()
      const dotIndex = tab.name.lastIndexOf('.')
      if (dotIndex > 0) {
        renameInputEl.setSelectionRange(0, dotIndex)
      } else {
        renameInputEl.select()
      }
    }
  })
}

async function confirmRename() {
  if (!renamingTabId.value) return

  const tab = editorStore.tabs.find(t => t.id === renamingTabId.value)
  if (!tab || !renameValue.value.trim() || renameValue.value.trim() === tab.name) {
    renamingTabId.value = null
    renameValue.value = ''
    return
  }

  const parentPath = tab.path.split(/[/\\]/).slice(0, -1).join(tab.path.includes('\\') ? '\\' : '/')
  const sep = tab.path.includes('\\') ? '\\' : '/'
  const newPath = parentPath + sep + renameValue.value.trim()

  try {
    await RenameFile(tab.path, newPath)
    editorStore.renameTab(tab.id, newPath)
    if (fileStore.currentDirectory) {
      const tree = await GetDirectoryTree(fileStore.currentDirectory)
      fileStore.setFileTree(tree?.root || null)
    }
    ElMessage.success('重命名成功')
  } catch (error) {
    console.error('Failed to rename:', error)
    ElMessage.error(`重命名失败: ${error}`)
  }

  renamingTabId.value = null
  renameValue.value = ''
}

function cancelRename() {
  renamingTabId.value = null
  renameValue.value = ''
}

// Tab bar action buttons
async function handleCloseAllTabs() {
  if (editorStore.tabs.length === 0) return
  if (editorStore.hasUnsavedChanges) {
    const confirmed = await ShowConfirmDialog('关闭所有', '部分文件未保存，确定关闭所有标签页？')
    if (!confirmed) return
  }
  editorStore.closeAllTabs()
  ElMessage.success('已关闭所有标签页')
}

async function handleCloseOtherTabs() {
  const activeTab = editorStore.activeTab
  if (!activeTab) return
  const otherDirtyTabs = editorStore.tabs.filter(t => t.id !== activeTab.id && t.isDirty)
  if (otherDirtyTabs.length > 0) {
    const confirmed = await ShowConfirmDialog('关闭其他', '其他标签页中有未保存的更改，确定关闭？')
    if (!confirmed) return
  }
  editorStore.closeOtherTabs(activeTab.id)
  ElMessage.success('已关闭其他标签页')
}

async function handleCloseSavedTabs() {
  const savedTabs = editorStore.tabs.filter(t => !t.isDirty)
  if (savedTabs.length === 0) {
    ElMessage.info('没有已保存的标签页')
    return
  }
  for (const tab of savedTabs) {
    editorStore.closeTab(tab.id)
  }
  ElMessage.success(`已关闭 ${savedTabs.length} 个已保存的标签页`)
}

// Click outside to close context menu
function handleClickOutside() {
  if (contextMenu.value.visible) {
    closeContextMenu()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div class="flex items-stretch bg-[#e8e8e8] dark:bg-[#252526] border-b border-gray-300 dark:border-gray-700 select-none">
    <!-- Tab list (scrollable) -->
    <div class="flex-1 flex items-end overflow-x-auto min-w-0">
      <div
        v-for="(tab, idx) in editorStore.tabs"
        :key="tab.id"
        draggable="true"
        class="tab flex items-center px-2.5 py-1.5 min-w-[100px] max-w-[180px] cursor-pointer border-r border-gray-300 dark:border-gray-700 relative"
        :class="{
          'bg-white dark:bg-[#1e1e1e] active': tab.id === editorStore.activeTabId,
          'bg-[#dcdcdc] dark:bg-[#2d2d2d] hover:bg-[#d4d4d4] dark:hover:bg-[#333]': tab.id !== editorStore.activeTabId,
          'opacity-50': dragIndex === idx,
          'border-l-2 border-l-blue-500': dragOverIndex === idx && dragIndex !== null && dragIndex !== idx,
        }"
        @click="selectTab(tab)"
        @contextmenu="handleContextMenu($event, tab)"
        @dblclick="startRename(tab)"
        @dragstart="onDragStart($event, idx)"
        @dragover="onDragOver($event, idx)"
        @dragleave="onDragLeave"
        @drop="onDrop($event, idx)"
        @dragend="onDragEnd"
      >
        <FileText class="w-3.5 h-3.5 mr-1.5 text-gray-400 flex-shrink-0" />
        <!-- Rename input -->
        <div v-if="renamingTabId === tab.id" class="flex-1 min-w-0" @click.stop>
          <input
            :ref="setRenameRef"
            v-model="renameValue"
            class="tab-rename-input"
            @keydown.enter="confirmRename"
            @keydown.escape="cancelRename"
            @blur="confirmRename"
          />
        </div>
        <span v-else class="text-[12px] truncate flex-1 leading-tight">{{ getTabTitle(tab) }}</span>
        <button
          class="tab-close-btn ml-1.5 p-0.5 rounded hover:bg-gray-300/80 dark:hover:bg-gray-500/60 flex-shrink-0"
          @click="closeTab(tab, $event)"
        >
          <X class="w-3 h-3" />
        </button>
      </div>
    </div>

    <!-- Tab action buttons (right side, notepad-- style) -->
    <div v-if="editorStore.tabs.length > 0" class="flex items-center px-1 border-l border-gray-300 dark:border-gray-700 bg-[#e8e8e8] dark:bg-[#252526] flex-shrink-0">
      <el-tooltip content="关闭已保存的标签页" placement="bottom" :show-after="300">
        <button
          class="p-1 rounded hover:bg-gray-300/70 dark:hover:bg-gray-600/60 text-gray-500 hover:text-green-600 dark:hover:text-green-400 transition-colors"
          @click="handleCloseSavedTabs"
        >
          <CheckCircle2 class="w-3.5 h-3.5" />
        </button>
      </el-tooltip>
      <el-tooltip content="关闭其他标签页" placement="bottom" :show-after="300">
        <button
          class="p-1 rounded hover:bg-gray-300/70 dark:hover:bg-gray-600/60 text-gray-500 hover:text-orange-600 dark:hover:text-orange-400 transition-colors"
          @click="handleCloseOtherTabs"
        >
          <XCircle class="w-3.5 h-3.5" />
        </button>
      </el-tooltip>
      <el-tooltip content="关闭所有标签页" placement="bottom" :show-after="300">
        <button
          class="p-1 rounded hover:bg-gray-300/70 dark:hover:bg-gray-600/60 text-gray-500 hover:text-red-600 dark:hover:text-red-400 transition-colors"
          @click="handleCloseAllTabs"
        >
          <XSquare class="w-3.5 h-3.5" />
        </button>
      </el-tooltip>
    </div>

    <!-- Tab context menu -->
    <Teleport to="body">
      <div
        v-if="contextMenu.visible"
        class="context-menu"
        :style="{ left: `${contextMenu.x}px`, top: `${contextMenu.y}px` }"
        @click.stop
      >
        <div class="context-menu-item" @click="closeThisTab">
          <X class="w-4 h-4 mr-2 text-gray-400" />
          <span>关闭当前文档</span>
          <span class="ml-auto text-xs text-gray-400">Ctrl+W</span>
        </div>
        <div class="context-menu-item" @click="closeNonCurrentTabs">
          <XCircle class="w-4 h-4 mr-2 text-orange-400" />
          <span>关闭非当前文档</span>
        </div>
        <div class="context-menu-item" @click="closeLeftAllTabs">
          <ChevronsLeft class="w-4 h-4 mr-2 text-gray-400" />
          <span>关闭左侧全部</span>
        </div>
        <div class="context-menu-item" @click="closeTabsToRight">
          <ChevronsRight class="w-4 h-4 mr-2 text-gray-400" />
          <span>关闭右侧全部</span>
        </div>
        <div class="context-menu-item" @click="closeAllTabs">
          <XSquare class="w-4 h-4 mr-2 text-red-400" />
          <span>关闭所有</span>
        </div>
        <div class="context-menu-divider"></div>
        <div class="context-menu-item" @click="copyTabPath">
          <Copy class="w-4 h-4 mr-2 text-gray-400" />
          <span>复制文件路径</span>
        </div>
        <div class="context-menu-item" @click="startRenameFromMenu">
          <Pencil class="w-4 h-4 mr-2 text-gray-400" />
          <span>重命名当前文档</span>
        </div>
        <div class="context-menu-item" @click="saveAsTab">
          <Save class="w-4 h-4 mr-2 text-gray-400" />
          <span>当前文档另存为...</span>
        </div>
        <div class="context-menu-item" @click="openInNewWindow">
          <ExternalLink class="w-4 h-4 mr-2 text-gray-400" />
          <span>在新窗口中打开</span>
        </div>
        <div class="context-menu-item" @click="showInExplorer">
          <FolderOpen class="w-4 h-4 mr-2 text-gray-400" />
          <span>在资源管理器中显示...</span>
        </div>
        <div class="context-menu-divider"></div>
        <div class="context-menu-item" @click="reloadAsText">
          <FileCode class="w-4 h-4 mr-2 text-gray-400" />
          <span>以文本模式重载</span>
        </div>
        <div class="context-menu-item" @click="reloadAsHex">
          <Binary class="w-4 h-4 mr-2 text-gray-400" />
          <span>以二进制模式重载</span>
        </div>
        <div class="context-menu-divider"></div>
        <div class="context-menu-item" @click="selectLeftCmpFile">
          <GitCompare class="w-4 h-4 mr-2 text-gray-400" />
          <span>选择左侧对比文件</span>
        </div>
        <div class="context-menu-item" @click="selectRightCmpFile">
          <GitCompare class="w-4 h-4 mr-2 text-blue-400" />
          <span>选择右侧对比文件</span>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.tab {
  position: relative;
  transition: background 0.1s;
}

.tab:last-child {
  border-right: none;
}

/* Active tab indicator (top blue line, like notepad--) */
.tab.active {
  box-shadow: inset 0 2px 0 0 #3b82f6;
}

.tab.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 1px;
  background-color: #3b82f6;
}

.tab-rename-input {
  width: 100%;
  padding: 0 2px;
  font-size: 12px;
  line-height: 1.4;
  border: 1px solid #3b82f6;
  border-radius: 2px;
  outline: none;
  background: white;
  color: #333;
}

html.dark .tab-rename-input {
  background: #3c3c3c;
  color: #e0e0e0;
  border-color: #60a5fa;
}
</style>