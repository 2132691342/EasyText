<script lang="ts" setup>
import { onMounted, onUnmounted, watch, ref } from 'vue'
import { useFileStore, useEditorStore } from '@/stores'
import { GetDirectoryTree, ReadFile, CreateDirectory, SaveFile } from '../../wailsjs/go/main/App'
import { getTabViewType, getFileExtension } from '@/utils'
import FileTree from './FileTree.vue'
import { AlertCircle, FilePlus, FolderPlus, RefreshCw, FolderOpen } from 'lucide-vue-next'
import { ElMessage } from 'element-plus'

const fileStore = useFileStore()
const editorStore = useEditorStore()

const loadError = ref<string | null>(null)

// Sidebar context menu (for blank area)
const sidebarContextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
})

// Inline new item in root
const isRootNewItem = ref(false)
const rootNewItemIsDir = ref(false)
const rootNewItemName = ref('')
const rootNewItemInput = ref<HTMLInputElement | null>(null)

// Watch for directory changes
watch(() => fileStore.currentDirectory, async (newDir) => {
  if (newDir) {
    loadError.value = null
    try {
      const tree = await GetDirectoryTree(newDir)
      fileStore.setFileTree(tree?.root || null)
    } catch (error) {
      loadError.value = String(error)
    }
  } else {
    fileStore.setFileTree(null)
    loadError.value = null
  }
}, { immediate: true })

// Refresh directory tree
async function refreshTree() {
  const rootPath = fileStore.fileTree?.path || fileStore.currentDirectory
  if (!rootPath) return
  loadError.value = null
  try {
    const tree = await GetDirectoryTree(rootPath)
    fileStore.setFileTree(tree?.root || null)
    ElMessage.success('目录已刷新')
  } catch (error) {
    console.error('Failed to refresh tree:', error)
    loadError.value = String(error)
    ElMessage.error('刷新目录失败')
  }
}

// Sidebar blank area context menu
function handleSidebarContextMenu(e: MouseEvent) {
  if (!fileStore.hasDirectory) return
  e.preventDefault()
  sidebarContextMenu.value = {
    visible: true,
    x: e.clientX,
    y: e.clientY,
  }
}

function closeSidebarContextMenu() {
  sidebarContextMenu.value.visible = false
}

// New file at root
function startRootNewFile() {
  closeSidebarContextMenu()
  isRootNewItem.value = true
  rootNewItemIsDir.value = false
  rootNewItemName.value = ''
  setTimeout(() => rootNewItemInput.value?.focus(), 100)
}

// New folder at root
function startRootNewFolder() {
  closeSidebarContextMenu()
  isRootNewItem.value = true
  rootNewItemIsDir.value = true
  rootNewItemName.value = ''
  setTimeout(() => rootNewItemInput.value?.focus(), 100)
}

async function confirmRootNewItem() {
  if (!rootNewItemName.value.trim() || !fileStore.currentDirectory) {
    isRootNewItem.value = false
    return
  }

  const sep = fileStore.currentDirectory.includes('\\') ? '\\' : '/'
  const fullPath = fileStore.currentDirectory + sep + rootNewItemName.value.trim()

  try {
    if (rootNewItemIsDir.value) {
      await CreateDirectory(fullPath)
    } else {
      await SaveFile(fullPath, '', 'UTF-8')
    }
    await refreshTree()

    if (!rootNewItemIsDir.value) {
      const existingTab = editorStore.getTabByPath(fullPath)
      if (existingTab) {
        editorStore.activateTab(existingTab.id)
      } else {
        editorStore.createTab(fullPath, '', 'UTF-8', 'LF')
      }
    }
    ElMessage.success(`${rootNewItemIsDir.value ? '文件夹' : '文件'}创建成功`)
  } catch (error) {
    console.error('Failed to create item:', error)
    ElMessage.error(`创建失败: ${error}`)
  }

  isRootNewItem.value = false
  rootNewItemName.value = ''
}

function cancelRootNewItem() {
  isRootNewItem.value = false
  rootNewItemName.value = ''
}

// Click outside to close sidebar context menu
function handleDocumentClick() {
  if (sidebarContextMenu.value.visible) {
    closeSidebarContextMenu()
  }
}

onMounted(() => {
  document.addEventListener('click', handleDocumentClick)
})
// 侧边栏随面板切换反复挂载/卸载，不移除会持续累积监听器
onUnmounted(() => {
  document.removeEventListener('click', handleDocumentClick)
})
</script>

<template>
  <div class="h-full flex flex-col bg-[var(--et-bg-sunken)]">
    <!-- Header -->
    <div class="sb-header">
      <span class="sb-title">资源管理器</span>
      <!-- Action buttons when directory is open -->
      <div v-if="fileStore.hasDirectory" class="flex items-center gap-1">
        <button class="sb-btn" title="新建文件" @click="startRootNewFile">
          <FilePlus class="w-4 h-4" />
        </button>
        <button class="sb-btn" title="新建文件夹" @click="startRootNewFolder">
          <FolderPlus class="w-4 h-4" />
        </button>
        <button class="sb-btn" title="刷新" @click="refreshTree">
          <RefreshCw class="w-4 h-4" />
        </button>
      </div>
    </div>

    <!-- File tree or recent files -->
    <div class="flex-1 overflow-auto" @contextmenu="handleSidebarContextMenu">
      <!-- Error message -->
      <div v-if="loadError" class="flex flex-col items-center justify-center py-8 px-4 text-center">
        <AlertCircle class="w-8 h-8 text-red-400 mb-2" />
        <p class="text-sm text-red-500 mb-1">目录加载失败</p>
        <p class="text-xs text-gray-400 break-all">{{ loadError }}</p>
      </div>

      <FileTree v-if="fileStore.fileTree" :node="fileStore.fileTree" />

      <!-- Root-level new item input -->
      <div
        v-if="isRootNewItem && fileStore.fileTree"
        class="flex items-center py-1 px-2"
        style="padding-left: 24px;"
        @click.stop
      >
        <component :is="rootNewItemIsDir ? FolderPlus : FilePlus" class="w-4 h-4 mr-2 text-gray-400 flex-shrink-0" />
        <input
          ref="rootNewItemInput"
          v-model="rootNewItemName"
          class="rename-input"
          :placeholder="rootNewItemIsDir ? '文件夹名称' : '文件名称'"
          @keydown.enter="confirmRootNewItem"
          @keydown.escape="cancelRootNewItem"
          @blur="confirmRootNewItem"
        />
      </div>

      <!-- Empty state when no folder is open -->
      <div v-if="!fileStore.hasDirectory" class="flex flex-col items-center justify-center py-10 px-4 text-center">
        <FolderOpen class="w-10 h-10 mb-3 text-[color:var(--et-fg-subtle)]" />
        <p class="text-sm text-[color:var(--et-fg-muted)]">尚未打开文件夹</p>
        <p class="text-xs mt-1 text-[color:var(--et-fg-subtle)]">菜单「文件 → 打开目录」或资源管理器空白处右键新建</p>
      </div>
    </div>

    <!-- Sidebar blank area context menu -->
    <Teleport to="body">
      <div
        v-if="sidebarContextMenu.visible"
        class="context-menu"
        :style="{ left: `${sidebarContextMenu.x}px`, top: `${sidebarContextMenu.y}px` }"
        @click.stop
      >
        <div class="context-menu-item" @click="startRootNewFile">
          <FilePlus class="w-4 h-4 mr-2 text-gray-400" />
          <span>新建文件</span>
        </div>
        <div class="context-menu-item" @click="startRootNewFolder">
          <FolderPlus class="w-4 h-4 mr-2 text-gray-400" />
          <span>新建文件夹</span>
        </div>
        <div class="context-menu-divider"></div>
        <div class="context-menu-item" @click="refreshTree(); closeSidebarContextMenu()">
          <RefreshCw class="w-4 h-4 mr-2 text-gray-400" />
          <span>刷新</span>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.sb-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 6px 10px; border-bottom: 1px solid var(--et-border);
  flex-shrink: 0;
}
.sb-title { font-size: 13px; font-weight: 500; color: var(--et-fg); }
.sb-btn {
  display: inline-flex; align-items: center; justify-content: center;
  padding: 3px; border: none; border-radius: var(--et-radius-sm);
  background: transparent; color: var(--et-fg-subtle); cursor: pointer;
  transition: background .12s ease, color .12s ease;
}
.sb-btn:hover { background: var(--et-bg-hover); color: var(--et-fg); }

.rename-input {
  width: 100%;
  padding: 1px 4px;
  font-size: 13px;
  line-height: 1.4;
  border: 1px solid var(--et-accent);
  border-radius: var(--et-radius-sm);
  outline: none;
  background: var(--et-bg);
  color: var(--et-fg);
}
</style>