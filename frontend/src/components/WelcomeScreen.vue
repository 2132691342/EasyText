<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { FileText, FolderOpen, FileJson, Diff, Clock, X } from 'lucide-vue-next'
import { OpenFileDialog, OpenDirectoryDialog, ReadFile, GetRecentFiles, GetRecentFolders, AddRecentEntry, ClearRecentFiles, ClearRecentFolders } from '../../wailsjs/go/main/App'
import { useEditorStore, useFileStore } from '@/stores'
import { getFileExtension, getTabViewType } from '@/utils'
import type { RecentEntry } from '@/types'

const editorStore = useEditorStore()
const fileStore = useFileStore()

const emit = defineEmits(['open-diff'])

// 🆕 V2.0.0 最近访问
const recentFiles = ref<RecentEntry[]>([])
const recentFolders = ref<RecentEntry[]>([])

async function loadRecent() {
  try {
    const [files, folders] = await Promise.all([GetRecentFiles(), GetRecentFolders()])
    recentFiles.value = files || []
    recentFolders.value = folders || []
  } catch { /* 忽略 */ }
}

async function openFile() {
  try {
    const path = await OpenFileDialog()
    if (path) {
      await openFilePath(path)
    }
  } catch (error) {
    console.error('Failed to open file:', error)
  }
}

async function openFilePath(path: string) {
  try {
    const ext = getFileExtension(path)
    const viewType = getTabViewType(ext)
    if (viewType !== 'code') {
      editorStore.createTab(path, '', 'binary', 'LF')
    } else {
      const result = await ReadFile(path)
      if (result) {
        editorStore.createTab(path, result.content, result.info.encoding, result.info.lineEnding)
      }
    }
    await AddRecentEntry(path, false)
  } catch (error) {
    console.error('Failed to open file:', error)
  }
}

async function openRecentFile(entry: RecentEntry) {
  await openFilePath(entry.path)
  await loadRecent()
}

async function openRecentFolder(entry: RecentEntry) {
  try {
    fileStore.setDirectory(entry.path)
    await AddRecentEntry(entry.path, true)
    await loadRecent()
  } catch (error) {
    console.error('Failed to open folder:', error)
  }
}

async function openFolder() {
  try {
    const path = await OpenDirectoryDialog()
    if (path) {
      fileStore.setDirectory(path)
      await AddRecentEntry(path, true)
    }
  } catch (error) {
    console.error('Failed to open folder:', error)
  }
}

async function clearRecentFiles() {
  await ClearRecentFiles()
  recentFiles.value = []
}

async function clearRecentFolders() {
  await ClearRecentFolders()
  recentFolders.value = []
}

onMounted(() => {
  loadRecent()
})
</script>

<template>
  <div class="h-full flex flex-col items-center justify-center bg-white dark:bg-[#1e1e1e]">
    <!-- Logo -->
    <div class="mb-8">
      <FileText class="w-16 h-16 text-gray-400 dark:text-gray-500" />
    </div>

    <!-- Title -->
    <h1 class="text-2xl font-bold text-gray-700 dark:text-gray-200 mb-2">EasyText</h1>
    <p class="text-sm text-gray-500 dark:text-gray-400 mb-8">轻量级桌面文档编辑工具</p>

    <!-- Quick actions -->
    <div class="grid grid-cols-2 gap-4 max-w-md">
      <button
        class="quick-action flex flex-col items-center p-4 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors"
        @click="openFile"
      >
        <FileText class="w-8 h-8 text-blue-500 mb-2" />
        <span class="text-sm text-gray-700 dark:text-gray-200">打开文件</span>
        <span class="text-xs text-gray-400">Ctrl+O</span>
      </button>

      <button
        class="quick-action flex flex-col items-center p-4 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors"
        @click="openFolder"
      >
        <FolderOpen class="w-8 h-8 text-blue-500 mb-2" />
        <span class="text-sm text-gray-700 dark:text-gray-200">打开文件夹</span>
        <span class="text-xs text-gray-400">Ctrl+Shift+O</span>
      </button>

      <button
        class="quick-action flex flex-col items-center p-4 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors opacity-60"
      >
        <FileJson class="w-8 h-8 text-blue-500 mb-2" />
        <span class="text-sm text-gray-700 dark:text-gray-200">JSON 工具</span>
        <span class="text-xs text-gray-400">Ctrl+Shift+F</span>
      </button>

      <button
        class="quick-action flex flex-col items-center p-4 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors"
        @click="emit('open-diff')"
      >
        <Diff class="w-8 h-8 text-blue-500 mb-2" />
        <span class="text-sm text-gray-700 dark:text-gray-200">文档对比</span>
      </button>
    </div>

    <!-- 🆕 V2.0.0 最近访问 -->
    <div v-if="recentFiles.length > 0 || recentFolders.length > 0" class="mt-6 max-w-md w-full px-4">
      <!-- 最近文件 -->
      <div v-if="recentFiles.length > 0" class="mb-4">
        <div class="flex items-center justify-between mb-2">
          <div class="flex items-center gap-1.5 text-xs text-gray-400">
            <Clock class="w-3.5 h-3.5" />
            <span>最近打开的文件</span>
          </div>
          <button class="text-[10px] text-gray-400 hover:text-red-400" @click="clearRecentFiles" title="清除">
            <X class="w-3 h-3" />
          </button>
        </div>
        <div class="space-y-0.5">
          <div
            v-for="entry in recentFiles.slice(0, 10)" :key="entry.path"
            class="flex items-center gap-2 px-2 py-1 rounded hover:bg-gray-50 dark:hover:bg-[#2a2a2a] cursor-pointer group"
            @click="openRecentFile(entry)"
          >
            <FileText class="w-3.5 h-3.5 text-gray-400 flex-shrink-0" />
            <span class="text-xs text-gray-600 dark:text-gray-300 truncate flex-1">{{ entry.name }}</span>
            <span class="text-[10px] text-gray-400 truncate max-w-[200px] hidden group-hover:inline">{{ entry.path }}</span>
          </div>
        </div>
      </div>

      <!-- 最近文件夹 -->
      <div v-if="recentFolders.length > 0">
        <div class="flex items-center justify-between mb-2">
          <div class="flex items-center gap-1.5 text-xs text-gray-400">
            <FolderOpen class="w-3.5 h-3.5" />
            <span>最近打开的文件夹</span>
          </div>
          <button class="text-[10px] text-gray-400 hover:text-red-400" @click="clearRecentFolders" title="清除">
            <X class="w-3 h-3" />
          </button>
        </div>
        <div class="space-y-0.5">
          <div
            v-for="entry in recentFolders.slice(0, 10)" :key="entry.path"
            class="flex items-center gap-2 px-2 py-1 rounded hover:bg-gray-50 dark:hover:bg-[#2a2a2a] cursor-pointer group"
            @click="openRecentFolder(entry)"
          >
            <FolderOpen class="w-3.5 h-3.5 text-yellow-500 flex-shrink-0" />
            <span class="text-xs text-gray-600 dark:text-gray-300 truncate flex-1">{{ entry.name }}</span>
            <span class="text-[10px] text-gray-400 truncate max-w-[200px] hidden group-hover:inline">{{ entry.path }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="mt-8 text-xs text-gray-400">
      <p>版本 2.0.0 | MIT License</p>
    </div>
  </div>
</template>

<style scoped>
.quick-action {
  min-width: 140px;
}
</style>