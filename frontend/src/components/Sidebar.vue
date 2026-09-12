<script lang="ts" setup>
/**
 * 侧栏容器（资源管理器 / 文件树）v2.1
 *
 * 设计目标（v2.1 收敛）：
 *  - 统一面板头：28px、`--et-bg-sunken` 背景、底边 1px `--et-border`、
 *    标题 13px medium、右侧操作图标 14px（来自 .et-chrome-sbh / .et-icon-btn-sm）。
 *  - 统一空状态：`.et-empty`，未来 6 个面板共用。
 *  - 硬编码色 / text-gray-N 全部 token 化。
 *  - 路径新建输入框复用 token 化样式（不再用 .rename-input 局部样式）。
 */
import { onMounted, onUnmounted, watch, ref } from 'vue'
import { useFileStore, useEditorStore } from '@/stores'
import { GetDirectoryTree, CreateDirectory, SaveFile } from '../../wailsjs/go/main/App'
import FileTree from './FileTree.vue'
import { AlertCircle, FilePlus, FolderPlus, RefreshCw, FolderOpen } from 'lucide-vue-next'
import { ElMessage } from 'element-plus'

const fileStore = useFileStore()
const editorStore = useEditorStore()

const loadError = ref<string | null>(null)

const sidebarContextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
})

const isRootNewItem = ref(false)
const rootNewItemIsDir = ref(false)
const rootNewItemName = ref('')
const rootNewItemInput = ref<HTMLInputElement | null>(null)

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

async function refreshTree() {
  const rootPath = fileStore.fileTree?.path || fileStore.currentDirectory
  if (!rootPath) return
  loadError.value = null
  try {
    const tree = await GetDirectoryTree(rootPath)
    fileStore.setFileTree(tree?.root || null)
    ElMessage.success('目录已刷新')
  } catch (error) {
    loadError.value = String(error)
    ElMessage.error('刷新目录失败')
  }
}

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
function startRootNewFile() {
  closeSidebarContextMenu()
  isRootNewItem.value = true
  rootNewItemIsDir.value = false
  rootNewItemName.value = ''
  setTimeout(() => rootNewItemInput.value?.focus(), 100)
}
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
    ElMessage.error(`创建失败: ${error}`)
  }
  isRootNewItem.value = false
  rootNewItemName.value = ''
}
function cancelRootNewItem() {
  isRootNewItem.value = false
  rootNewItemName.value = ''
}
function handleDocumentClick() {
  if (sidebarContextMenu.value.visible) closeSidebarContextMenu()
}

onMounted(() => document.addEventListener('click', handleDocumentClick))
onUnmounted(() => document.removeEventListener('click', handleDocumentClick))
</script>

<template>
  <div class="h-full flex flex-col sidebar-shell">
    <!-- Header：28px 标准面板头 -->
    <div class="et-chrome-row et-chrome-sbh sb-header">
      <span class="sb-title">资源管理器</span>
      <div v-if="fileStore.hasDirectory" class="sb-actions">
        <button class="et-icon-btn-sm" title="新建文件" @click="startRootNewFile">
          <FilePlus :size="14" :stroke-width="1.6" />
        </button>
        <button class="et-icon-btn-sm" title="新建文件夹" @click="startRootNewFolder">
          <FolderPlus :size="14" :stroke-width="1.6" />
        </button>
        <button class="et-icon-btn-sm" title="刷新" @click="refreshTree">
          <RefreshCw :size="14" :stroke-width="1.6" />
        </button>
      </div>
    </div>

    <!-- 主体 -->
    <div class="flex-1 overflow-auto" @contextmenu="handleSidebarContextMenu">
      <!-- 错误态：复用 .et-empty，icon 改为 danger 色 -->
      <div v-if="loadError" class="et-empty">
        <AlertCircle class="et-empty-icon" :size="32" :stroke-width="1.6" style="color: var(--et-danger)" />
        <div class="et-empty-title" style="color: var(--et-danger)">目录加载失败</div>
        <div class="et-empty-hint" :title="loadError">{{ loadError }}</div>
      </div>

      <FileTree v-if="fileStore.fileTree" :node="fileStore.fileTree" />

      <!-- 根级别新建输入 -->
      <div
        v-if="isRootNewItem && fileStore.fileTree"
        class="sb-new-row"
        @click.stop
      >
        <component :is="rootNewItemIsDir ? FolderPlus : FilePlus" :size="14" :stroke-width="1.6" class="sb-new-icon" />
        <input
          ref="rootNewItemInput"
          v-model="rootNewItemName"
          class="sb-new-input"
          :placeholder="rootNewItemIsDir ? '文件夹名称' : '文件名称'"
          @keydown.enter="confirmRootNewItem"
          @keydown.escape="cancelRootNewItem"
          @blur="confirmRootNewItem"
        />
      </div>

      <!-- 空状态：复用 .et-empty -->
      <div v-if="!fileStore.hasDirectory && !loadError" class="et-empty">
        <FolderOpen class="et-empty-icon" :size="32" :stroke-width="1.6" />
        <div class="et-empty-title">尚未打开文件夹</div>
        <div class="et-empty-hint">菜单「文件 → 打开目录」或资源管理器空白处右键新建</div>
      </div>
    </div>

    <!-- 右键菜单 -->
    <Teleport to="body">
      <div
        v-if="sidebarContextMenu.visible"
        class="context-menu"
        :style="{ left: `${sidebarContextMenu.x}px`, top: `${sidebarContextMenu.y}px` }"
        @click.stop
      >
        <div class="context-menu-item" @click="startRootNewFile">
          <FilePlus class="ctx-icon" :size="14" :stroke-width="1.6" />
          <span>新建文件</span>
        </div>
        <div class="context-menu-item" @click="startRootNewFolder">
          <FolderPlus class="ctx-icon" :size="14" :stroke-width="1.6" />
          <span>新建文件夹</span>
        </div>
        <div class="context-menu-divider"></div>
        <div class="context-menu-item" @click="refreshTree(); closeSidebarContextMenu()">
          <RefreshCw class="ctx-icon" :size="14" :stroke-width="1.6" />
          <span>刷新</span>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.sidebar-shell {
  background-color: var(--et-bg-sunken);
}
.sb-header {
  /* .et-chrome-row.et-chrome-sbh 已提供 28px + sunken bg + 底边 */
  padding: 0 var(--et-space-2);
  gap: var(--et-space-2);
}
.sb-title {
  font-size: var(--et-text-md);
  font-weight: var(--et-fw-medium);
  color: var(--et-fg);
  flex: 1;
}
.sb-actions {
  display: flex;
  align-items: center;
  gap: 2px;
}

/* 根级别新建输入 */
.sb-new-row {
  display: flex;
  align-items: center;
  padding: 4px var(--et-space-2) 4px 24px;
  gap: var(--et-space-1);
}
.sb-new-icon {
  color: var(--et-fg-muted);
  flex-shrink: 0;
}
.sb-new-input {
  flex: 1;
  min-width: 0;
  padding: 2px 4px;
  font-size: var(--et-text-md);
  line-height: var(--et-lh-tight);
  border: 1px solid var(--et-accent);
  border-radius: var(--et-radius-sm);
  outline: none;
  background: var(--et-bg);
  color: var(--et-fg);
}
.sb-new-input:focus-visible {
  box-shadow: 0 0 0 1px var(--et-accent);
}

/* 右键菜单图标统一 token */
.ctx-icon {
  width: 14px;
  height: 14px;
  margin-right: var(--et-space-2);
  color: var(--et-fg-muted);
  flex-shrink: 0;
}
</style>