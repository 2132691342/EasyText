<script lang="ts" setup>
import { useEditorStore } from '@/stores'
import { X, FileText } from 'lucide-vue-next'
import type { EditorTab } from '@/types'
import { ShowConfirmDialog, SaveFile } from '../wailsjs/go/main/App'

const editorStore = useEditorStore()

async function closeTab(tab: EditorTab, e: MouseEvent) {
  e.stopPropagation()

  if (tab.isDirty) {
    const confirm = await ShowConfirmDialog('保存更改', `是否保存对 ${tab.name} 的更改？`)
    if (confirm) {
      await SaveFile(tab.path, tab.content, tab.encoding)
      editorStore.markTabSaved(tab.id)
    }
  }

  editorStore.closeTab(tab.id)
}

function selectTab(tab: EditorTab) {
  editorStore.activateTab(tab.id)
}

function getTabTitle(tab: EditorTab): string {
  return tab.name + (tab.isDirty ? ' •' : '')
}
</script>

<template>
  <div class="flex items-center bg-gray-100 dark:bg-[#252526] border-b border-gray-200 dark:border-gray-700 overflow-x-auto">
    <div
      v-for="tab in editorStore.tabs"
      :key="tab.id"
      class="tab flex items-center px-3 py-2 min-w-[120px] max-w-[200px] cursor-pointer border-r border-gray-200 dark:border-gray-700"
      :class="{
        'bg-white dark:bg-[#1e1e1e]': tab.id === editorStore.activeTabId,
        'bg-gray-100 dark:bg-[#2d2d2d]': tab.id !== editorStore.activeTabId,
      }"
      @click="selectTab(tab)"
    >
      <FileText class="w-4 h-4 mr-2 text-gray-400 flex-shrink-0" />
      <span class="text-sm truncate flex-1">{{ getTabTitle(tab) }}</span>
      <button
        class="tab-close-btn ml-2 p-0.5 rounded hover:bg-gray-200 dark:hover:bg-gray-600 flex-shrink-0"
        @click="closeTab(tab, $event)"
      >
        <X class="w-3 h-3" />
      </button>
    </div>
  </div>
</template>

<style scoped>
.tab {
  position: relative;
}

.tab:last-child {
  border-right: none;
}

/* Active tab indicator */
.tab.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 2px;
  background-color: #3b82f6;
}
</style>