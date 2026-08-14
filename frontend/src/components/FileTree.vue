<script lang="ts" setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useFileStore, useEditorStore } from '@/stores'
import {
  ReadFile, DeleteFile, DeleteDirectory, RenameFile, CopyFile, CreateDirectory,
  SaveFile, GetDirectoryTree
} from '../../wailsjs/go/main/App'
import type { TreeNode } from '@/types'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTabViewType } from '@/utils'
import {
  ChevronDown, ChevronRight, FileText, Folder, FolderOpen,
  FileCode, FileJson, File, Image, Database, Terminal,
  FilePlus, FolderPlus, Copy, Clipboard, Pencil, Trash2
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
      } catch (error) {
        console.error('Failed to open file:', error)
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
  } catch (error) {
    console.error('Failed to refresh tree:', error)
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

// ---- New file/folder ----

function startNewFile() {
  closeContextMenu()
  if (props.node.isDir) {
    if (!isExpanded.value) {
      fileStore.expandPath(props.node.path)
    }
    isNewItem.value = true
    newItemIsDir.value = false
    newItemName.value = ''
    setTimeout(() => newItemInput.value?.focus(), 100)
  } else {
    // For files, we need to create in the same parent directory
    // This is handled differently — we set the state on the parent
    // But since we're a recursive component, we emit to parent
    // Simplification: just create in the file's parent directory
    const parentPath = getParentPath(props.node.path)
    const sep = getSep(props.node.path)
    const newName = 'new_file.txt'
    const fullPath = parentPath + sep + newName
    CreateDirectory(parentPath + sep + '__new_placeholder__').catch(() => {})
    // Actually, we'll show the inline input on this node's "parent level"
    // Since we can't easily traverse up, we just create at parent level
    // For now, use a simpler approach: just create at parent directly
    isNewItem.value = true
    newItemIsDir.value = false
    newItemName.value = ''
    setTimeout(() => newItemInput.value?.focus(), 100)
  }
}

function startNewFolder() {
  closeContextMenu()
  if (props.node.isDir) {
    if (!isExpanded.value) {
      fileStore.expandPath(props.node.path)
    }
    isNewItem.value = true
    newItemIsDir.value = true
    newItemName.value = ''
    setTimeout(() => newItemInput.value?.focus(), 100)
  } else {
    isNewItem.value = true
    newItemIsDir.value = true
    newItemName.value = ''
    setTimeout(() => newItemInput.value?.focus(), 100)
  }
}

async function confirmNewItem() {
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

  try {
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
  }

  isNewItem.value = false
  newItemName.value = ''
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

async function confirmRename() {
  if (!renameValue.value.trim() || renameValue.value.trim() === props.node.name) {
    isRenaming.value = false
    return
  }

  const parentPath = getParentPath(props.node.path)
  const sep = getSep(props.node.path)
  const newPath = parentPath + sep + renameValue.value.trim()

  try {
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
  }

  isRenaming.value = false
  renameValue.value = ''
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
  let newName: string
  if (dotIndex > 0) {
    newName = name.substring(0, dotIndex) + ' - 副本' + name.substring(dotIndex)
  } else {
    newName = name + ' - 副本'
  }
  const newPath = parentPath + sep + newName

  try {
    if (props.node.isDir) {
      await CreateDirectory(newPath)
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
    } catch (error) {
      console.error('Failed to open file:', error)
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

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

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
      :class="{ selected: isSelected }"
      @click="handleClick"
      @contextmenu="handleContextMenu"
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

    <!-- Children (for directories) -->
    <div v-if="node.isDir && isExpanded && node.children">
      <!-- New item input appears as first child inside the directory -->
      <div
        v-if="isNewItem"
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
          <div class="context-menu-item" @click="handleCopy">
            <Copy class="w-4 h-4 mr-2 text-gray-400" />
            <span>复制</span>
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
          <div class="context-menu-item" @click="handleCopy">
            <Copy class="w-4 h-4 mr-2 text-gray-400" />
            <span>复制</span>
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
.rename-input {
  width: 100%;
  padding: 1px 4px;
  font-size: 13px;
  line-height: 1.4;
  border: 1px solid #3b82f6;
  border-radius: 3px;
  outline: none;
  background: white;
  color: #333;
}

html.dark .rename-input {
  background: #3c3c3c;
  color: #e0e0e0;
  border-color: #60a5fa;
}
</style>