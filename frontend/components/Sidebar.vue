<script lang="ts" setup>
import { onMounted, computed, watch } from 'vue'
import { useFileStore, useEditorStore } from '@/stores'
import { GetDirectoryTree, ReadFile, GetRecentFiles } from '../wailsjs/go/main/App'
import FileTree from './FileTree.vue'
import { ChevronDown, ChevronRight, Clock, FileText } from 'lucide-vue-next'

const fileStore = useFileStore()
const editorStore = useEditorStore()

const showRecent = computed(() => !fileStore.hasDirectory)

onMounted(async () => {
  // Load recent files
  try {
    const files = await GetRecentFiles()
    fileStore.setRecentFiles(files || [])
  } catch (error) {
    console.error('Failed to load recent files:', error)
  }
})

// Watch for directory changes
watch(() => fileStore.currentDirectory, async (newDir) => {
  if (newDir) {
    try {
      const tree = await GetDirectoryTree(newDir)
      fileStore.setFileTree(tree?.root || null)
    } catch (error) {
      console.error('Failed to load directory tree:', error)
    }
  } else {
    fileStore.setFileTree(null)
  }
}, { immediate: true })

async function openRecentFile(path: string) {
  try {
    const existingTab = editorStore.getTabByPath(path)
    if (existingTab) {
      editorStore.activateTab(existingTab.id)
      return
    }

    const result = await ReadFile(path)
    if (result) {
      editorStore.createTab(path, result.content, result.info.encoding, result.info.lineEnding)
    }
  } catch (error) {
    console.error('Failed to open recent file:', error)
  }
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))

  if (days === 0) return '今天'
  if (days === 1) return '昨天'
  if (days < 7) return `${days}天前`
  return date.toLocaleDateString('zh-CN')
}
</script>

<template>
  <div class="h-full flex flex-col bg-white dark:bg-[#252526]">
    <!-- Header -->
    <div class="flex items-center justify-between px-3 py-2 border-b border-gray-200 dark:border-gray-700">
      <span class="text-sm font-medium text-gray-700 dark:text-gray-200">
        {{ fileStore.hasDirectory ? '资源管理器' : '最近文件' }}
      </span>
    </div>

    <!-- File tree or recent files -->
    <div class="flex-1 overflow-auto">
      <FileTree v-if="fileStore.fileTree" :node="fileStore.fileTree" />

      <!-- Recent files list -->
      <div v-else-if="showRecent" class="py-2">
        <div v-if="fileStore.recentFiles.length === 0" class="px-4 py-8 text-center text-gray-400 text-sm">
          <FileText class="w-8 h-8 mx-auto mb-2 opacity-50" />
          <p>暂无最近文件</p>
          <p class="text-xs mt-1">打开文件后将显示在这里</p>
        </div>
        <div v-else>
          <div class="px-3 py-1 text-xs text-gray-400 uppercase">最近打开</div>
          <div
            v-for="file in fileStore.recentFiles"
            :key="file.Path"
            class="file-tree-item flex items-center px-3 py-1.5"
            @click="openRecentFile(file.Path)"
          >
            <FileText class="w-4 h-4 mr-2 text-gray-400" />
            <span class="flex-1 truncate text-sm">{{ file.Name }}</span>
            <span class="text-xs text-gray-400 ml-2">{{ formatDate(file.LastOpen) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Additional styles if needed */
</style>