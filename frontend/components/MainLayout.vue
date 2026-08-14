<script lang="ts" setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useEditorStore, useFileStore, useSettingStore } from '@/stores'
import Toolbar from './Toolbar.vue'
import Sidebar from './Sidebar.vue'
import EditorArea from './EditorArea.vue'
import StatusBar from './StatusBar.vue'
import { OpenFileDialog, OpenDirectoryDialog, ReadFile, SaveFile, SaveFileDialog, ShowConfirmDialog } from '../wailsjs/go/main/App'

const editorStore = useEditorStore()
const fileStore = useFileStore()
const settingStore = useSettingStore()

const sidebarWidth = ref(250)
const isResizing = ref(false)

const showSidebar = computed(() => settingStore.showFileTree)
const showStatusBar = computed(() => settingStore.showStatusBar)

// Keyboard shortcuts
function handleKeyDown(e: KeyboardEvent) {
  const isCtrl = e.ctrlKey || e.metaKey

  if (isCtrl && e.key === 'o') {
    e.preventDefault()
    if (e.shiftKey) {
      openDirectory()
    } else {
      openFile()
    }
  } else if (isCtrl && e.key === 's') {
    e.preventDefault()
    if (e.shiftKey) {
      saveFileAs()
    } else {
      saveFile()
    }
  } else if (isCtrl && e.key === 'n') {
    e.preventDefault()
    newFile()
  } else if (isCtrl && e.key === 'w') {
    e.preventDefault()
    if (editorStore.activeTab) {
      closeTab(editorStore.activeTab.id)
    }
  } else if (isCtrl && e.key === 'Tab') {
    e.preventDefault()
    switchTab(e.shiftKey ? -1 : 1)
  }
}

async function openFile() {
  try {
    const path = await OpenFileDialog()
    if (!path) return

    // Check if already open
    const existingTab = editorStore.getTabByPath(path)
    if (existingTab) {
      editorStore.activateTab(existingTab.id)
      return
    }

    // Read file
    const result = await ReadFile(path)
    if (result) {
      editorStore.createTab(path, result.content, result.info.encoding, result.info.lineEnding)
    }
  } catch (error) {
    console.error('Failed to open file:', error)
  }
}

async function openDirectory() {
  try {
    const path = await OpenDirectoryDialog()
    if (!path) return

    fileStore.setDirectory(path)
    // File tree will be loaded by Sidebar component
  } catch (error) {
    console.error('Failed to open directory:', error)
  }
}

async function saveFile() {
  const tab = editorStore.activeTab
  if (!tab) return

  try {
    if (!tab.path) {
      await saveFileAs()
      return
    }

    await SaveFile(tab.path, tab.content, tab.encoding)
    editorStore.markTabSaved(tab.id)
  } catch (error) {
    console.error('Failed to save file:', error)
  }
}

async function saveFileAs() {
  const tab = editorStore.activeTab
  if (!tab) return

  try {
    const path = await SaveFileDialog(tab.name)
    if (!path) return

    await SaveFile(path, tab.content, tab.encoding)
    editorStore.renameTab(tab.id, path)
    editorStore.markTabSaved(tab.id)
  } catch (error) {
    console.error('Failed to save file as:', error)
  }
}

function newFile() {
  editorStore.createTab('', '', 'UTF-8', 'LF')
}

async function closeTab(tabId: string) {
  const tab = editorStore.tabs.find(t => t.id === tabId)
  if (!tab) return

  if (tab.isDirty) {
    const confirm = await ShowConfirmDialog('保存更改', '文件已修改，是否保存？')
    if (confirm) {
      await saveFile()
    }
  }

  editorStore.closeTab(tabId)
}

function switchTab(direction: number) {
  const tabs = editorStore.tabs
  if (tabs.length === 0) return

  const currentIndex = editorStore.activeTabIndex
  let newIndex = currentIndex + direction

  if (newIndex < 0) {
    newIndex = tabs.length - 1
  } else if (newIndex >= tabs.length) {
    newIndex = 0
  }

  editorStore.activateTab(tabs[newIndex].id)
}

// Resize sidebar
function startResize(e: MouseEvent) {
  isResizing.value = true
  document.addEventListener('mousemove', handleResize)
  document.addEventListener('mouseup', stopResize)
}

function handleResize(e: MouseEvent) {
  if (!isResizing.value) return
  sidebarWidth.value = Math.max(150, Math.min(500, e.clientX))
}

function stopResize() {
  isResizing.value = false
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('mouseup', stopResize)
}

onMounted(() => {
  window.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
})
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Toolbar -->
    <Toolbar @open-file="openFile" @open-directory="openDirectory" @save="saveFile" @new-file="newFile" />

    <!-- Main content -->
    <div class="flex-1 flex overflow-hidden">
      <!-- Sidebar -->
      <template v-if="showSidebar">
        <div class="flex-shrink-0 border-r border-gray-200 dark:border-gray-700" :style="{ width: `${sidebarWidth}px` }">
          <Sidebar />
        </div>
        <div class="resize-handle" @mousedown="startResize"></div>
      </template>

      <!-- Editor area -->
      <div class="flex-1 overflow-hidden">
        <EditorArea />
      </div>
    </div>

    <!-- Status bar -->
    <StatusBar v-if="showStatusBar" />
  </div>
</template>

<style scoped>
.resize-handle {
  width: 4px;
  cursor: col-resize;
  background-color: transparent;
  transition: background-color 0.2s;
}

.resize-handle:hover {
  background-color: #3b82f6;
}

html.dark .resize-handle:hover {
  background-color: #60a5fa;
}
</style>