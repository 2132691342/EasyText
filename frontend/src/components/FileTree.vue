<script lang="ts" setup>
import { computed, ref, onMounted, onUnmounted, watch } from 'vue'
import { useFileStore, useEditorStore } from '@/stores'
import {
  ReadFile, DeleteFile, DeleteDirectory, RenameFile, CopyFile, CreateDirectory,
  SaveFile, GetDirectoryTree, MoveFile, CopyDirectory, CheckPathExists
} from '../../wailsjs/go/main/App'
import type { TreeNode } from '@/types'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTabViewType } from '@/utils'
import {
  ChevronDown, ChevronRight, FileText, Folder, FolderOpen,
  FileCode, FileJson, File, Image, Database, Terminal,
  FilePlus, FolderPlus, Copy, ClipboardPaste, Scissors, Clipboard, Pencil, Trash2
} from 'lucide-vue-next'

const props = defineProps<{
  node: TreeNode
  depth?: number
}>()

const fileStore = useFileStore()
const editorStore = useEditorStore()

const depth = computed(() => props.depth || 0)
const isExpanded = computed(() => fileStore.isExpanded(props.node.path))
const isSelected = computed(() => fileStore.selectedPath === props.node.path)

// Context menu state
const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  targetNode: null as TreeNode | null,
  isDir: false,
})

// Inline rename state
const isRenaming = ref(false)
const renameValue = ref('')
const renameInput = ref<HTMLInputElement | null>(null)

// Inline new file/folder state — only used when this node is a directory
const isNewItem = ref(false)
const newItemIsDir = ref(false)
const newItemName = ref('')
const newItemInput = ref<HTMLInputElement | null>(null)

function toggleExpand() {
  fileStore.togglePath(props.node.path)
}

async function handleClick() {
  if (isNewItem.value) return
  if (isRenaming.value) return
  if (props.node.isDir) {
    toggleExpand()
  } else {
    fileStore.setSelectedPath(props.node.path)

    // Check if already open
    const existingTab = editorStore.getTabByPath(props.node.path)
    if (existingTab) {
      editorStore.activateTab(existingTab.id)
      return
    }

    // Determine file type and open accordingly
    const ext = props.node.ext?.toLowerCase() || ''
    const viewType = getTabViewType(ext)

    if (viewType !== 'code') {
      // Binary/document files: open with empty content, viewer will load bytes
      editorStore.createTab(props.node.path, '', 'binary', 'LF')
    } else {
      // Text files: read and display in code editor
      try {
        const result = await ReadFile(props.node.path)
        if (result) {
          editorStore.createTab(props.node.path, result.content, result.info.encoding, result.info.lineEnding)
        }
      } catch (error: any) {
        console.error('Failed to open file:', error)
        ElMessage.error('打开文件失败: ' + (error?.message || ''))
      }
    }
  }
}

// Right-click context menu
function handleContextMenu(e: MouseEvent) {
  e.preventDefault()
  e.stopPropagation()
  contextMenu.value = {
    visible: true,
    x: e.clientX,
    y: e.clientY,
    targetNode: props.node,
    isDir: props.node.isDir,
  }
  // Close inline inputs
  isRenaming.value = false
  isNewItem.value = false
}

function closeContextMenu() {
  contextMenu.value.visible = false
}

// Refresh file tree
async function refreshTree() {
  // Prefer the tree's own root path; fall back to the open directory.
  // (After the open directory itself is deleted, currentDirectory may still
  // point at a missing path, so we must target the actual displayed root.)
  const rootPath = fileStore.fileTree?.path || fileStore.currentDirectory
  if (!rootPath) return
  try {
    const tree = await GetDirectoryTree(rootPath)
    fileStore.setFileTree(tree?.root || null)
  } catch (error: any) {
    console.error('Failed to refresh tree:', error)
    ElMessage.error('刷新目录失败: ' + (error?.message || ''))
  }
}

// Check if `ancestor` is the same as `descendant` or a parent directory of it.
function isAncestorOrEqual(ancestor: string, descendant: string | null): boolean {
  if (!ancestor || !descendant) return false
  if (ancestor === descendant) return true
  const sep = ancestor.includes('\\') ? '\\' : '/'
  return descendant.startsWith(ancestor + sep)
}

// Close all editor tabs whose path lives under the deleted node.
function closeTabsUnder(deletedPath: string) {
  const sep = deletedPath.includes('\\') ? '\\' : '/'
  const tabs = editorStore.tabs.filter(
    t => t.path && (t.path === deletedPath || t.path.startsWith(deletedPath + sep))
  )
  for (const tab of tabs) {
    editorStore.closeTab(tab.id)
  }
}

// Get parent path of a node
function getParentPath(path: string): string {
  const sep = path.includes('\\') ? '\\' : '/'
  const parts = path.split(/[/\\]/)
  parts.pop()
  return parts.join(sep)
}

// Get path separator
function getSep(path: string): string {
  return path.includes('\\') ? '\\' : '/'
}

// ==================== 拖拽移动（树内拖到目标文件夹） ====================
// 拖拽源/目标是两个不同的组件实例，源路径放 fileStore 共享；
// 高亮目标用本实例的 dropTargetPath 即可。
const dropTargetPath = ref<string | null>(null)

function onDragStart(e: DragEvent) {
  fileStore.setTreeDragPath(props.node.path)
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', props.node.path)
    // 内部树拖拽标记：MainLayout 的 document drop 监听据此忽略，
    // 避免把节点路径当成文本插入编辑器
    e.dataTransfer.setData('application/x-easytext-node', props.node.path)
  }
}
function onDragEnd() {
  fileStore.setTreeDragPath(null)
  dropTargetPath.value = null
}
function canDropOn(target: TreeNode): boolean {
  const srcPath = fileStore.treeDragPath
  if (!srcPath || !target.isDir) return false
  if (srcPath === target.path) return false
  // 不能把节点拖进它自己的子目录
  if (isAncestorOrEqual(srcPath, target.path)) return false
  return true
}
function onDragOver(e: DragEvent) {
  if (!canDropOn(props.node)) return
  e.preventDefault()
  if (e.dataTransfer) e.dataTransfer.dropEffect = 'move'
  dropTargetPath.value = props.node.path
}
function onDragLeave() {
  if (dropTargetPath.value === props.node.path) dropTargetPath.value = null
}
async function onDrop(e: DragEvent) {
  e.preventDefault()
  e.stopPropagation() // 不冒泡到 MainLayout 的 document drop（外部文件打开）监听
  const srcPath = fileStore.treeDragPath
  fileStore.setTreeDragPath(null)
  dropTargetPath.value = null
  if (!srcPath || !canDropOn(props.node)) return
  const src = fileStore.findNode(srcPath)
  if (!src) return
  await moveNodeTo(src, props.node.path)
}

/** 把 src 移动到 targetDir 下（拖拽与剪切-粘贴共用） */
async function moveNodeTo(src: TreeNode, targetDir: string) {
  const sep = getSep(targetDir)
  const dest = targetDir + sep + src.name
  if (dest === src.path) return
  try {
    await MoveFile(src.path, dest)
    relocateTabsUnder(src.path, dest)
    fileStore.expandPath(targetDir)
    await refreshTree()
    ElMessage.success(`已移动到 ${targetDir.split(/[/\\]/).pop() || targetDir}`)
  } catch (err: any) {
    ElMessage.error('移动失败: ' + (err?.message || err))
  }
}

/** 移动/重命名后同步更新受影响 tab 的路径（保持已打开状态不丢） */
function relocateTabsUnder(oldPath: string, newPath: string) {
  const sep = getSep(oldPath)
  for (const t of editorStore.tabs) {
    if (!t.path) continue
    if (t.path === oldPath) {
      editorStore.renameTab(t.id, newPath)
    } else if (t.path.startsWith(oldPath + sep)) {
      editorStore.renameTab(t.id, newPath + sep + t.path.slice(oldPath.length + sep.length))
    }
  }
}

// ==================== 剪切 / 复制 / 粘贴 ====================

function cutNode() {
  closeContextMenu()
  fileStore.setTreeClipboard({ path: props.node.path, name: props.node.name, isDir: props.node.isDir, cut: true })
  ElMessage.success('已剪切，请到目标文件夹粘贴')
}
function copyNode() {
  closeContextMenu()
  fileStore.setTreeClipboard({ path: props.node.path, name: props.node.name, isDir: props.node.isDir, cut: false })
  ElMessage.success('已复制，请到目标文件夹粘贴')
}
async function pasteNode() {
  closeContextMenu()
  const clip = fileStore.treeClipboard
  if (!clip) return
  const targetDir = props.node.isDir ? props.node.path : getParentPath(props.node.path)
  const sep = getSep(props.node.path)
  const dest = targetDir + sep + clip.name
  if (clip.path === dest || isAncestorOrEqual(clip.path, dest)) {
    ElMessage.error('不能粘贴到自身或其子目录')
    return
  }
  try {
    // 重名预检：CopyFile 后端不拦截同名，统一在前端拦，避免静默覆盖
    if (await CheckPathExists(dest)) {
      ElMessage.error(`目标已存在：${clip.name}`)
      return
    }
    if (clip.cut) {
      await MoveFile(clip.path, dest)
      relocateTabsUnder(clip.path, dest)
      fileStore.setTreeClipboard(null)
    } else if (clip.isDir) {
      await CopyDirectory(clip.path, dest)
    } else {
      await CopyFile(clip.path, dest)
    }
    fileStore.expandPath(targetDir)
    await refreshTree()
    ElMessage.success(clip.cut ? '粘贴成功（已移动）' : '粘贴成功')
  } catch (err: any) {
    ElMessage.error('粘贴失败: ' + (err?.message || err))
  }
}

// ---- New file/folder ----

/** 展开目录（若是目录）并聚焦内联输入框，同时保证输入框滚进可视区 */
function focusNewItemInput(isDir: boolean) {
  if (props.node.isDir && !isExpanded.value) {
    fileStore.expandPath(props.node.path)
  }
  isNewItem.value = true
  newItemIsDir.value = isDir
  newItemName.value = ''
  setTimeout(() => {
    newItemInput.value?.focus()
    newItemInput.value?.scrollIntoView({ block: 'nearest' })
  }, 100)
}

function startNewFile() {
  closeContextMenu()
  if (props.node.isDir) {
    focusNewItemInput(false)
  } else {
    // 在文件节点上新建：目标目录是该文件的父目录（confirmNewItem 会据此拼路径）。
    // 这里只打开内联输入框，不写磁盘——早期版本会先创建一个 __new_placeholder__
    // 目录，属于残留调试代码，会在用户磁盘上留下垃圾目录。
    focusNewItemInput(false)
  }
}

function startNewFolder() {
  closeContextMenu()
  focusNewItemInput(true)
}

let creating = false
async function confirmNewItem() {
  // Enter 与 blur 会先后各触发一次，creating 防止重复创建
  if (creating) return
  if (!newItemName.value.trim()) {
    isNewItem.value = false
    return
  }

  // Determine the parent directory
  let parentPath: string
  if (props.node.isDir) {
    parentPath = props.node.path
  } else {
    parentPath = getParentPath(props.node.path)
  }

  const sep = getSep(props.node.path)
  const fullPath = parentPath + sep + newItemName.value.trim()

  creating = true
  try {
    // 重名预检：SaveFile 对同名文件会静默覆盖，必须先拦
    if (await CheckPathExists(fullPath)) {
      ElMessage.error(`同名${newItemIsDir.value ? '文件夹' : '文件'}已存在：${newItemName.value.trim()}`)
      return
    }
    if (newItemIsDir.value) {
      await CreateDirectory(fullPath)
    } else {
      // Create empty file
      await SaveFile(fullPath, '', 'UTF-8')
    }
    await refreshTree()

    // If it's a file, open it in editor
    if (!newItemIsDir.value) {
      const existingTab = editorStore.getTabByPath(fullPath)
      if (existingTab) {
        editorStore.activateTab(existingTab.id)
      } else {
        editorStore.createTab(fullPath, '', 'UTF-8', 'LF')
      }
    }
    ElMessage.success(`${newItemIsDir.value ? '文件夹' : '文件'}创建成功`)
  } catch (error) {
    console.error('Failed to create item:', error)
    ElMessage.error(`创建失败: ${error}`)
  } finally {
    creating = false
    isNewItem.value = false
    newItemName.value = ''
  }
}

function cancelNewItem() {
  isNewItem.value = false
  newItemName.value = ''
}

// ---- Rename ----

function startRename() {
  closeContextMenu()
  isRenaming.value = true
  renameValue.value = props.node.name
  setTimeout(() => {
    if (renameInput.value) {
      renameInput.value.focus()
      // Select name without extension for files
      if (!props.node.isDir) {
        const dotIndex = props.node.name.lastIndexOf('.')
        if (dotIndex > 0) {
          renameInput.value.setSelectionRange(0, dotIndex)
        } else {
          renameInput.value.select()
        }
      } else {
        renameInput.value.select()
      }
    }
  }, 50)
}

let renamingBusy = false
async function confirmRename() {
  // Enter 与 blur 会先后各触发一次，防止第二次对已不存在的源路径报错
  if (renamingBusy) return
  if (!renameValue.value.trim() || renameValue.value.trim() === props.node.name) {
    isRenaming.value = false
    return
  }

  const parentPath = getParentPath(props.node.path)
  const sep = getSep(props.node.path)
  const newPath = parentPath + sep + renameValue.value.trim()

  renamingBusy = true
  try {
    if (await CheckPathExists(newPath)) {
      ElMessage.error(`同名文件/文件夹已存在：${renameValue.value.trim()}`)
      return
    }
    await RenameFile(props.node.path, newPath)
    // Update open tab if this file is open
    const tab = editorStore.getTabByPath(props.node.path)
    if (tab) {
      editorStore.renameTab(tab.id, newPath)
    }
    await refreshTree()
    ElMessage.success('重命名成功')
  } catch (error) {
    console.error('Failed to rename:', error)
    ElMessage.error(`重命名失败: ${error}`)
  } finally {
    renamingBusy = false
    isRenaming.value = false
    renameValue.value = ''
  }
}

function cancelRename() {
  isRenaming.value = false
  renameValue.value = ''
}

// ---- Delete ----

async function handleDelete() {
  closeContextMenu()
  try {
    await ElMessageBox.confirm(
      `确定要删除 ${props.node.isDir ? '文件夹' : '文件'} "${props.node.name}" 吗？${props.node.isDir ? '文件夹内所有内容将被删除。' : ''}`,
      '确认删除',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
        confirmButtonClass: 'el-button--danger',
      }
    )

    if (props.node.isDir) {
      await DeleteDirectory(props.node.path)
    } else {
      await DeleteFile(props.node.path)
    }
    // Close any open tabs that lived under the deleted path
    closeTabsUnder(props.node.path)
    // If we just deleted the directory that is currently open (or its ancestor /
    // the tree root), clear the tree instead of refreshing against a missing path.
    if (props.node.path === fileStore.fileTree?.path || isAncestorOrEqual(props.node.path, fileStore.currentDirectory)) {
      fileStore.setDirectory(null)
      fileStore.setFileTree(null)
    } else {
      await refreshTree()
    }
    ElMessage.success('删除成功')
  } catch (error: any) {
    if (error === 'cancel' || error === 'close') return
    console.error('Failed to delete:', error)
    ElMessage.error(`删除失败: ${error?.message || error}`)
  }
}

// ---- Copy path ----

async function copyPath() {
  closeContextMenu()
  try {
    await navigator.clipboard.writeText(props.node.path)
    ElMessage.success('路径已复制到剪贴板')
  } catch (error) {
    console.error('Failed to copy path:', error)
    ElMessage.error('复制路径失败')
  }
}

// ---- Copy (duplicate) file/folder ----

async function handleCopy() {
  closeContextMenu()
  const parentPath = getParentPath(props.node.path)
  const sep = getSep(props.node.path)
  const name = props.node.name
  const dotIndex = name.lastIndexOf('.')
  const base = (i: number) => {
    const tag = i <= 1 ? '' : String(i)
    return dotIndex > 0
      ? name.substring(0, dotIndex) + ' - 副本' + tag + name.substring(dotIndex)
      : name + ' - 副本' + tag
  }
  // 重名自动加序号：副本、副本2、副本3…
  let newName = base(1)
  for (let i = 2; i < 100; i++) {
    let exists = false
    try { exists = await CheckPathExists(parentPath + sep + newName) } catch { break }
    if (!exists) break
    newName = base(i)
  }
  const newPath = parentPath + sep + newName

  try {
    if (props.node.isDir) {
      // 此前目录"复制"只 CreateDirectory 一个空目录（内容全丢），
      // 现走后端递归复制
      await CopyDirectory(props.node.path, newPath)
    } else {
      await CopyFile(props.node.path, newPath)
    }
    await refreshTree()
    ElMessage.success('复制成功')
  } catch (error) {
    console.error('Failed to copy:', error)
    ElMessage.error(`复制失败: ${error}`)
  }
}

// ---- Open in editor ----

async function openInEditor() {
  closeContextMenu()
  if (props.node.isDir) return

  fileStore.setSelectedPath(props.node.path)
  const existingTab = editorStore.getTabByPath(props.node.path)
  if (existingTab) {
    editorStore.activateTab(existingTab.id)
    return
  }

  const ext = props.node.ext?.toLowerCase() || ''
  const viewType = getTabViewType(ext)

  if (viewType !== 'code') {
    editorStore.createTab(props.node.path, '', 'binary', 'LF')
  } else {
    try {
      const result = await ReadFile(props.node.path)
      if (result) {
        editorStore.createTab(props.node.path, result.content, result.info.encoding, result.info.lineEnding)
      }
    } catch (error: any) {
      console.error('Failed to open file:', error)
      ElMessage.error('打开文件失败: ' + (error?.message || ''))
    }
  }
}

// ---- File icon ----

function getFileIcon(node: TreeNode) {
  if (node.isDir) {
    return isExpanded.value ? FolderOpen : Folder
  }

  const ext = node.ext?.toLowerCase() || ''
  const iconMap: Record<string, any> = {
    'js': FileCode,
    'jsx': FileCode,
    'ts': FileCode,
    'tsx': FileCode,
    'json': FileJson,
    'html': FileCode,
    'css': FileCode,
    'py': FileCode,
    'java': FileCode,
    'go': FileCode,
    'sql': Database,
    'sh': Terminal,
    'bash': Terminal,
    'md': FileText,
    'txt': FileText,
    'png': Image,
    'jpg': Image,
    'jpeg': Image,
    'gif': Image,
    'svg': Image,
    'webp': Image,
    'bmp': Image,
    'pdf': FileText,
    'doc': FileText,
    'docx': FileText,
    'xls': FileJson,
    'xlsx': FileJson,
    'csv': FileJson,
    'ppt': FileText,
    'pptx': FileText,
  }

  return iconMap[ext] || File
}

// ---- Click outside handler ----

function handleClickOutside() {
  if (contextMenu.value.visible) {
    closeContextMenu()
  }
}

// 全局 click 监听只在右键菜单打开期间挂载。
// 原实现让每个树节点实例在 onMounted 都注册一个 document 监听，
// 展开的节点越多，每次点击要跑的回调就越多（N 个节点 = N 个监听）。
watch(
  () => contextMenu.value.visible,
  (v) => {
    if (v) document.addEventListener('click', handleClickOutside)
    else document.removeEventListener('click', handleClickOutside)
  },
)

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div>
    <!-- Node row -->
    <div
      class="file-tree-item flex items-center py-1"
      :style="{ paddingLeft: `${depth * 16 + 8}px` }"
      :class="{
        selected: isSelected,
        'is-cut': fileStore.treeClipboard?.cut && fileStore.treeClipboard.path === node.path,
        'drop-target': dropTargetPath === node.path,
      }"
      draggable="true"
      @click="handleClick"
      @contextmenu="handleContextMenu"
      @dragstart="onDragStart"
      @dragend="onDragEnd"
      @dragover="onDragOver"
      @dragleave="onDragLeave"
      @drop="onDrop"
    >
      <!-- Expand/collapse icon for directories -->
      <span v-if="node.isDir" class="w-4 h-4 mr-1 flex items-center justify-center">
        <ChevronDown v-if="isExpanded" class="w-3 h-3 text-gray-400" />
        <ChevronRight v-else class="w-3 h-3 text-gray-400" />
      </span>
      <span v-else class="w-4 h-4 mr-1"></span>

      <!-- File/folder icon -->
      <component :is="getFileIcon(node)" class="w-4 h-4 mr-2 text-gray-400 flex-shrink-0" />

      <!-- Name or rename input -->
      <div v-if="isRenaming" class="flex-1 min-w-0" @click.stop>
        <input
          ref="renameInput"
          v-model="renameValue"
          class="rename-input"
          @keydown.enter="confirmRename"
          @keydown.escape="cancelRename"
          @blur="confirmRename"
        />
      </div>
      <span v-else class="text-sm truncate">{{ node.name }}</span>
    </div>

    <!-- 新建输入行（目录）：独立于 children 渲染——
         此前藏在 `isExpanded && node.children` 里，空目录 / children 未加载时
         输入框根本不渲染，表现为"新建一直挂起却看不见" -->
    <div
      v-if="isNewItem && node.isDir && isExpanded"
      class="flex items-center py-1"
      :style="{ paddingLeft: `${(depth + 1) * 16 + 8}px` }"
      @click.stop
    >
      <span class="w-4 h-4 mr-1"></span>
      <component :is="newItemIsDir ? Folder : FileText" class="w-4 h-4 mr-2 text-gray-400 flex-shrink-0" />
      <input
        ref="newItemInput"
        v-model="newItemName"
        class="rename-input"
        :placeholder="newItemIsDir ? '文件夹名称' : '文件名称'"
        @keydown.enter="confirmNewItem"
        @keydown.escape="cancelNewItem"
        @blur="confirmNewItem"
      />
    </div>

    <!-- Children (for directories) -->
    <div v-if="node.isDir && isExpanded && node.children">
      <FileTree
        v-for="child in node.children"
        :key="child.path"
        :node="child"
        :depth="depth + 1"
      />
    </div>

    <!-- For file nodes, new item input appears at the same depth level -->
    <div
      v-if="isNewItem && !node.isDir"
      class="flex items-center py-1"
      :style="{ paddingLeft: `${depth * 16 + 8}px` }"
      @click.stop
    >
      <span class="w-4 h-4 mr-1"></span>
      <component :is="newItemIsDir ? Folder : FileText" class="w-4 h-4 mr-2 text-gray-400 flex-shrink-0" />
      <input
        ref="newItemInput"
        v-model="newItemName"
        class="rename-input"
        :placeholder="newItemIsDir ? '文件夹名称' : '文件名称'"
        @keydown.enter="confirmNewItem"
        @keydown.escape="cancelNewItem"
        @blur="confirmNewItem"
      />
    </div>

    <!-- Context Menu (teleported to body) -->
    <Teleport to="body">
      <div
        v-if="contextMenu.visible"
        class="context-menu"
        :style="{ left: `${contextMenu.x}px`, top: `${contextMenu.y}px` }"
        @click.stop
      >
        <!-- Directory context menu -->
        <template v-if="contextMenu.isDir">
          <div class="context-menu-item" @click="startNewFile">
            <FilePlus class="w-4 h-4 mr-2 text-gray-400" />
            <span>新建文件</span>
          </div>
          <div class="context-menu-item" @click="startNewFolder">
            <FolderPlus class="w-4 h-4 mr-2 text-gray-400" />
            <span>新建文件夹</span>
          </div>
          <div class="context-menu-divider"></div>
          <div class="context-menu-item" @click="cutNode">
            <Scissors class="w-4 h-4 mr-2 text-gray-400" />
            <span>剪切</span>
          </div>
          <div class="context-menu-item" @click="copyNode">
            <Copy class="w-4 h-4 mr-2 text-gray-400" />
            <span>复制</span>
          </div>
          <div v-if="fileStore.treeClipboard" class="context-menu-item" @click="pasteNode">
            <ClipboardPaste class="w-4 h-4 mr-2 text-gray-400" />
            <span>粘贴到「{{ node.name }}」</span>
          </div>
          <div class="context-menu-item" @click="handleCopy">
            <Copy class="w-4 h-4 mr-2 text-gray-400" />
            <span>创建副本</span>
          </div>
          <div class="context-menu-item" @click="startRename">
            <Pencil class="w-4 h-4 mr-2 text-gray-400" />
            <span>重命名</span>
          </div>
          <div class="context-menu-item" @click="copyPath">
            <Clipboard class="w-4 h-4 mr-2 text-gray-400" />
            <span>复制路径</span>
          </div>
          <div class="context-menu-divider"></div>
          <div class="context-menu-item danger" @click="handleDelete">
            <Trash2 class="w-4 h-4 mr-2" />
            <span>删除</span>
          </div>
        </template>

        <!-- File context menu -->
        <template v-else>
          <div class="context-menu-item" @click="openInEditor">
            <FileText class="w-4 h-4 mr-2 text-gray-400" />
            <span>打开文件</span>
          </div>
          <div class="context-menu-divider"></div>
          <div class="context-menu-item" @click="cutNode">
            <Scissors class="w-4 h-4 mr-2 text-gray-400" />
            <span>剪切</span>
          </div>
          <div class="context-menu-item" @click="copyNode">
            <Copy class="w-4 h-4 mr-2 text-gray-400" />
            <span>复制</span>
          </div>
          <div v-if="fileStore.treeClipboard" class="context-menu-item" @click="pasteNode">
            <ClipboardPaste class="w-4 h-4 mr-2 text-gray-400" />
            <span>粘贴到「{{ getParentPath(node.path).split(/[/\\]/).pop() }}」</span>
          </div>
          <div class="context-menu-item" @click="handleCopy">
            <Copy class="w-4 h-4 mr-2 text-gray-400" />
            <span>创建副本</span>
          </div>
          <div class="context-menu-item" @click="startRename">
            <Pencil class="w-4 h-4 mr-2 text-gray-400" />
            <span>重命名</span>
          </div>
          <div class="context-menu-item" @click="copyPath">
            <Clipboard class="w-4 h-4 mr-2 text-gray-400" />
            <span>复制路径</span>
          </div>
          <div class="context-menu-divider"></div>
          <div class="context-menu-item danger" @click="handleDelete">
            <Trash2 class="w-4 h-4 mr-2" />
            <span>删除</span>
          </div>
        </template>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
/* 剪切待粘贴态：半透明提示，与资源管理器一致 */
.is-cut {
  opacity: .45;
}

/* 拖拽悬停目标文件夹高亮 */
.drop-target {
  box-shadow: inset 0 0 0 1px var(--et-accent);
  background-color: var(--et-accent-soft);
  border-radius: 3px;
}

.rename-input {
  width: 100%;
  padding: 1px 4px;
  font-size: 13px;
  line-height: 1.4;
  border: 1px solid var(--et-accent, #3b82f6);
  border-radius: 3px;
  outline: none;
  background: var(--et-bg, white);
  color: var(--et-fg, #333);
}

html.dark .rename-input {
  background: #3c3c3c;
  color: #e0e0e0;
  border-color: #60a5fa;
}
</style>