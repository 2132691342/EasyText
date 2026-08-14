<script lang="ts" setup>
import { emit } from '@/utils'
import { FileText, FolderOpen, FileJson, Diff, Settings } from 'lucide-vue-next'
import { OpenFileDialog, OpenDirectoryDialog } from '../wailsjs/go/main/App'

async function openFile() {
  try {
    const path = await OpenFileDialog()
    if (path) {
      emit('open-file', path)
    }
  } catch (error) {
    console.error('Failed to open file:', error)
  }
}

async function openFolder() {
  try {
    const path = await OpenDirectoryDialog()
    if (path) {
      emit('open-folder', path)
    }
  } catch (error) {
    console.error('Failed to open folder:', error)
  }
}
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
        class="quick-action flex flex-col items-center p-4 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors opacity-60"
      >
        <Diff class="w-8 h-8 text-blue-500 mb-2" />
        <span class="text-sm text-gray-700 dark:text-gray-200">文档对比</span>
      </button>
    </div>

    <!-- Footer -->
    <div class="mt-8 text-xs text-gray-400">
      <p>版本 1.0.0 | MIT License</p>
    </div>
  </div>
</template>

<style scoped>
.quick-action {
  min-width: 140px;
}
</style>