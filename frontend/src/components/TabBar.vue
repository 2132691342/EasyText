<script lang="ts" setup>
/**
 * Tab Bar v2.1
 *
 * 设计目标（对标 Notepad-- / Sublime Text）：
 *  - 高度 30px，与菜单/工具栏阶梯一致。
 *  - 关闭按钮常显（Sublime 风），hover 高亮；脏文件右上小圆点 (.et-dot-warn)。
 *  - tab 宽度 90–160（紧凑）；拖拽时源 tab 半透明 + 插入位置 2px 主色条。
 *  - tab-actions 颜色统一 token。
 */
import { ref, nextTick, onMounted, onUnmounted } from 'vue'
import { useEditorStore, useFileStore } from '@/stores'
import {
  X, FileText, XCircle, XSquare, CheckCircle2,
  ChevronsLeft, ChevronsRight,
  Save, ExternalLink, FolderOpen, FileCode, Binary,
  GitCompare, Copy, Pencil,
} from 'lucide-vue-next'
import { ElMessage } from 'element-plus'
import type { EditorTab } from '@/types'
import { SaveFile, RenameFile, GetDirectoryTree, SaveFileDialog } from '../../wailsjs/go/main/App'
import { confirmDialog, confirmSaveDiscard } from '@/utils/confirm'

const editorStore = useEditorStore()
const fileStore = useFileStore()

const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  tabId: '',
})

const renamingTabId = ref<string | null>(null)
const renameValue = ref('')
let renameInputEl: HTMLInputElement | null = null

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

function setRenameRef(el: any) {
  renameInputEl = el as HTMLInputElement | null
}

async function closeTab(tab: EditorTab, e: MouseEvent) {
  e.stopPropagation()

  if (tab.isDirty) {
    const choice = await confirmSaveDiscard(tab.name)
    if (choice === 'cancel') return
    if (choice === 'save' && tab.path) {
      try {
        await SaveFile(tab.path, tab.content, tab.encoding)
        editorStore.markTabSaved(tab.id)
      } catch {
        ElMessage.error('保存失败，已取消关闭')
        return
      }
    }
  }

  editorStore.closeTab(tab.id)
  closeContextMenu()
}

function selectTab(tab: EditorTab) {
  editorStore.activateTab(tab.id)
}

// 右键菜单
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

function closeThisTab() {
  const tab = editorStore.tabs.find(t => t.id === contextMenu.value.tabId)
  if (tab) closeTab(tab, new MouseEvent('click'))
}
function closeOtherTabs() {
  closeContextMenu()
  editorStore.closeOtherTabs(contextMenu.value.tabId)
  ElMessage.success('已关闭其他标签页')
}
function closeLeftAllTabs() {
  closeContextMenu()
  const tabId = contextMenu.value.tabId
  const tabIndex = editorStore.tabs.findIndex(t => t.id === tabId)
  if (tabIndex === -1) return
  for (const tab of editorStore.tabs.slice(0, tabIndex)) {
    editorStore.closeTab(tab.id)
  }
  ElMessage.success(`已关闭左侧 ${tabIndex} 个标签页`)
}
function closeNonCurrentTabs() {
  closeContextMenu()
  const tabId = contextMenu.value.tabId
  const others = editorStore.tabs.filter(t => t.id !== tabId)
  for (const tab of others) editorStore.closeTab(tab.id)
  editorStore.activateTab(tabId)
  ElMessage.success(`已关闭非当前 ${others.length} 个标签页`)
}
function closeTabsToRight() {
  closeContextMenu()
  const tabId = contextMenu.value.tabId
  const tabIndex = editorStore.tabs.findIndex(t => t.id === tabId)
  if (tabIndex === -1) return
  const right = editorStore.tabs.slice(tabIndex + 1)
  for (const tab of right) editorStore.closeTab(tab.id)
  ElMessage.success(`已关闭右侧 ${right.length} 个标签页`)
}
function closeAllTabs() {
  closeContextMenu()
  editorStore.closeAllTabs()
  ElMessage.success('已关闭所有标签页')
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
  if (!tab.path) {
    ElMessage.warning('该标签没有磁盘文件路径')
    return
  }
  tab.viewType = 'hex'
  ElMessage.success(`已以十六进制视图打开: ${tab.name}`)
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
async function copyTabPath() {
  closeContextMenu()
  const tab = editorStore.tabs.find(t => t.id === contextMenu.value.tabId)
  if (tab?.path) {
    try {
      await navigator.clipboard.writeText(tab.path)
      ElMessage.success('路径已复制到剪贴板')
    } catch {
      ElMessage.error('复制路径失败')
    }
  }
}
function startRenameFromMenu() {
  closeContextMenu()
  const tab = editorStore.tabs.find(t => t.id === contextMenu.value.tabId)
  if (tab) startRename(tab)
}
function startRename(tab: EditorTab) {
  if (!tab.path) return
  renamingTabId.value = tab.id
  renameValue.value = tab.name
  nextTick(() => {
    if (renameInputEl) {
      renameInputEl.focus()
      const dotIndex = tab.name.lastIndexOf('.')
      if (dotIndex > 0) renameInputEl.setSelectionRange(0, dotIndex)
      else renameInputEl.select()
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
  const sep = tab.path!.includes('\\') ? '\\' : '/'
  const parentPath = tab.path!.split(/[/\\]/).slice(0, -1).join(sep)
  const newPath = parentPath + sep + renameValue.value.trim()
  try {
    await RenameFile(tab.path!, newPath)
    editorStore.renameTab(tab.id, newPath)
    if (fileStore.currentDirectory) {
      const tree = await GetDirectoryTree(fileStore.currentDirectory)
      fileStore.setFileTree(tree?.root || null)
    }
    ElMessage.success('重命名成功')
  } catch (error) {
    ElMessage.error(`重命名失败: ${error}`)
  }
  renamingTabId.value = null
  renameValue.value = ''
}
function cancelRename() {
  renamingTabId.value = null
  renameValue.value = ''
}

async function handleCloseAllTabs() {
  if (editorStore.tabs.length === 0) return
  if (editorStore.hasUnsavedChanges) {
    const confirmed = await confirmDialog({
      title: '关闭所有',
      message: `有 ${editorStore.dirtyTabs.length} 个文件未保存，确定关闭所有标签页？`,
      confirmText: '全部关闭',
      danger: true,
    })
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
    const confirmed = await confirmDialog({
      title: '关闭其他',
      message: `其他标签页中有 ${otherDirtyTabs.length} 个文件未保存，确定关闭？`,
      confirmText: '关闭其他',
      danger: true,
    })
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
  for (const tab of savedTabs) editorStore.closeTab(tab.id)
  ElMessage.success(`已关闭 ${savedTabs.length} 个已保存的标签页`)
}

function handleClickOutside() {
  if (contextMenu.value.visible) closeContextMenu()
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onUnmounted(() => document.removeEventListener('click', handleClickOutside))
</script>

<template>
  <div class="et-chrome-row et-chrome-tab tabbar select-none">
    <!-- Tab list (scrollable) -->
    <div class="flex-1 flex items-stretch overflow-x-auto min-w-0">
      <div
        v-for="(tab, idx) in editorStore.tabs"
        :key="tab.id"
        draggable="true"
        class="tab"
        :class="{
          'is-active': tab.id === editorStore.activeTabId,
          'is-dragging': dragIndex === idx,
          'is-drop-target': dragOverIndex === idx && dragIndex !== null && dragIndex !== idx,
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
        <FileText class="tab-icon" />
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
        <span v-else class="tab-name">{{ tab.name }}</span>

        <!-- 脏标记：右上角小圆点（与 StatusBar 统一） -->
        <span v-if="tab.isDirty && renamingTabId !== tab.id" class="et-dot et-dot-warn tab-dirty-dot" />

        <button
          class="et-icon-btn-sm tab-close"
          :title="`关闭 ${tab.name}`"
          @click="closeTab(tab, $event)"
        >
          <X />
        </button>
      </div>
    </div>

    <!-- Tab action buttons (right side) -->
    <div v-if="editorStore.tabs.length > 0" class="tab-actions">
      <button class="et-icon-btn-sm tab-action" title="关闭已保存的标签页" @click="handleCloseSavedTabs">
        <CheckCircle2 />
      </button>
      <button class="et-icon-btn-sm tab-action" title="关闭其他标签页" @click="handleCloseOtherTabs">
        <XCircle />
      </button>
      <button class="et-icon-btn-sm tab-action tab-action--danger" title="关闭所有标签页" @click="handleCloseAllTabs">
        <XSquare />
      </button>
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
          <X class="ctx-icon" />
          <span>关闭当前文档</span>
          <span class="ctx-key">Ctrl+W</span>
        </div>
        <div class="context-menu-item" @click="closeNonCurrentTabs">
          <XCircle class="ctx-icon" />
          <span>关闭非当前文档</span>
        </div>
        <div class="context-menu-item" @click="closeLeftAllTabs">
          <ChevronsLeft class="ctx-icon" />
          <span>关闭左侧全部</span>
        </div>
        <div class="context-menu-item" @click="closeTabsToRight">
          <ChevronsRight class="ctx-icon" />
          <span>关闭右侧全部</span>
        </div>
        <div class="context-menu-item" @click="closeAllTabs">
          <XSquare class="ctx-icon ctx-icon-danger" />
          <span>关闭所有</span>
        </div>
        <div class="context-menu-divider"></div>
        <div class="context-menu-item" @click="copyTabPath">
          <Copy class="ctx-icon" />
          <span>复制文件路径</span>
        </div>
        <div class="context-menu-item" @click="startRenameFromMenu">
          <Pencil class="ctx-icon" />
          <span>重命名当前文档</span>
        </div>
        <div class="context-menu-item" @click="saveAsTab">
          <Save class="ctx-icon" />
          <span>当前文档另存为...</span>
        </div>
        <div class="context-menu-item is-disabled" @click="openInNewWindow">
          <ExternalLink class="ctx-icon" />
          <span>在新窗口中打开</span>
          <span class="ctx-key">未实现</span>
        </div>
        <div class="context-menu-item" @click="showInExplorer">
          <FolderOpen class="ctx-icon" />
          <span>在资源管理器中显示...</span>
        </div>
        <div class="context-menu-divider"></div>
        <div class="context-menu-item" @click="reloadAsText">
          <FileCode class="ctx-icon" />
          <span>以文本模式重载</span>
        </div>
        <div class="context-menu-item" @click="reloadAsHex">
          <Binary class="ctx-icon" />
          <span>以二进制模式重载</span>
        </div>
        <div class="context-menu-divider"></div>
        <div class="context-menu-item" @click="selectLeftCmpFile">
          <GitCompare class="ctx-icon" />
          <span>选择左侧对比文件</span>
        </div>
        <div class="context-menu-item" @click="selectRightCmpFile">
          <GitCompare class="ctx-icon ctx-icon-accent" />
          <span>选择右侧对比文件</span>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.tabbar {
  /* .et-chrome-row.et-chrome-tab 已提供：height 30px + sunken bg + 下边框 */
}

/* —— Tab 主体 —— */
.tab {
  position: relative;
  display: flex;
  align-items: center;
  gap: var(--et-space-1);
  height: 100%;
  padding: 0 6px 0 8px;
  min-width: 90px;
  max-width: 160px;
  cursor: pointer;
  border-right: 1px solid var(--et-border);
  color: var(--et-fg-muted);
  transition: background-color 80ms ease, color 80ms ease;
}
.tab:last-child { border-right: none; }
.tab:hover { background: var(--et-bg-hover); }
.tab.is-active {
  background: var(--et-bg);
  color: var(--et-fg);
}
/* 激活指示：顶部 2px 主色条 */
.tab.is-active::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 2px;
  background: var(--et-accent);
}

.tab.is-dragging {
  opacity: .4;
  cursor: grabbing;
}
/* 拖拽插入位置提示 */
.tab.is-drop-target {
  box-shadow: inset 2px 0 0 0 var(--et-accent);
}

.tab-icon {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  color: var(--et-fg-subtle);
}
.tab.is-active .tab-icon { color: var(--et-accent); }

.tab-name {
  font-size: var(--et-text-sm);
  line-height: var(--et-lh-tight);
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 脏标记：右上小圆点 */
.tab-dirty-dot {
  margin: 0 2px;
  flex-shrink: 0;
}

/* 关闭按钮：常显，hover 时高亮（Sublime 风） */
.tab-close {
  width: 18px;
  height: 18px;
  color: var(--et-fg-subtle);
}
.tab-close:hover {
  background: var(--et-bg-active);
  color: var(--et-fg);
}

/* —— Tab action 按钮 —— */
.tab-actions {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 0 var(--et-space-1);
  border-left: 1px solid var(--et-border);
  flex-shrink: 0;
}
.tab-action {
  color: var(--et-fg-subtle);
}
.tab-action:hover { color: var(--et-fg); }
.tab-action--danger:hover { color: var(--et-danger); }

/* —— 右键菜单图标 / 快捷键（统一 token） —— */
.ctx-icon {
  width: 16px;
  height: 16px;
  margin-right: var(--et-space-2);
  color: var(--et-fg-muted);
  flex-shrink: 0;
}
.ctx-icon-accent { color: var(--et-accent); }
.ctx-icon-danger  { color: var(--et-danger); }
.ctx-key {
  margin-left: auto;
  padding-left: var(--et-space-3);
  font-size: var(--et-text-xs);
  color: var(--et-fg-subtle);
  font-variant-numeric: tabular-nums;
}

/* —— 重命名输入 —— */
.tab-rename-input {
  width: 100%;
  padding: 1px 4px;
  font-size: var(--et-text-sm);
  line-height: var(--et-lh-tight);
  border: 1px solid var(--et-accent);
  border-radius: var(--et-radius-sm);
  outline: none;
  background: var(--et-bg);
  color: var(--et-fg);
}

/* 未实现项置灰（与菜单栏保持一致） */
.context-menu-item.is-disabled { color: var(--et-fg-subtle); cursor: default; }
.context-menu-item.is-disabled:hover { background: transparent; }
</style>