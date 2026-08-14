<script lang="ts" setup>
import { computed } from 'vue'
import { useEditorStore, useSettingStore } from '@/stores'
import { formatFileSize } from '@/utils'
import { Sun, Moon } from 'lucide-vue-next'

const editorStore = useEditorStore()
const settingStore = useSettingStore()

const activeTab = computed(() => editorStore.activeTab)
const config = computed(() => settingStore.config)

const cursorPosition = computed(() => {
  if (!activeTab) return { line: 1, column: 1 }
  return activeTab.cursorPosition
})

const fileInfo = computed(() => {
  if (!activeTab) return null
  return {
    encoding: activeTab.encoding,
    lineEnding: activeTab.lineEnding,
    language: activeTab.language,
    lines: activeTab.content.split('\n').length,
  }
})

const lineEndingLabel = computed(() => {
  return fileInfo?.lineEnding === 'CRLF' ? 'CRLF' : 'LF'
})
</script>

<template>
  <div class="status-bar flex items-center justify-between px-4 py-1 bg-gray-100 dark:bg-[#252526] border-t border-gray-200 dark:border-gray-700">
    <!-- Left section -->
    <div class="flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
      <span v-if="cursorPosition">
        行 {{ cursorPosition.line }}，列 {{ cursorPosition.column }}
      </span>
      <span v-if="fileInfo" class="border-l border-gray-300 dark:border-gray-600 pl-4">
        {{ fileInfo.lines }} 行
      </span>
      <span v-if="activeTab?.isReadOnly" class="border-l border-gray-300 dark:border-gray-600 pl-4">
        只读
      </span>
    </div>

    <!-- Center section -->
    <div class="flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
      <span v-if="activeTab?.isDirty" class="text-orange-500">未保存</span>
    </div>

    <!-- Right section -->
    <div class="flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
      <span v-if="fileInfo" class="border-l border-gray-300 dark:border-gray-600 pl-4">
        {{ fileInfo.encoding }}
      </span>
      <span v-if="fileInfo" class="border-l border-gray-300 dark:border-gray-600 pl-4">
        {{ lineEndingLabel }}
      </span>
      <span v-if="fileInfo && fileInfo.language" class="border-l border-gray-300 dark:border-gray-600 pl-4">
        {{ fileInfo.language }}
      </span>
      <span v-if="activeTab?.content" class="border-l border-gray-300 dark:border-gray-600 pl-4">
        {{ formatFileSize(activeTab.content.length) }}
      </span>

      <!-- Theme toggle -->
      <button
        class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700"
        @click="settingStore.toggleTheme()"
      >
        <Sun v-if="settingStore.isDarkMode" class="w-3 h-3" />
        <Moon v-else class="w-3 h-3" />
      </button>
    </div>
  </div>
</template>

<style scoped>
.status-bar {
  font-size: 12px;
}
</style>