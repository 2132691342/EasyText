<script lang="ts" setup>
import { computed } from 'vue'
import { useEditorStore } from '@/stores'
import { FileText, X, Circle } from 'lucide-vue-next'

const editorStore = useEditorStore()
const tabs = computed(() => editorStore.tabs)
const activeId = computed(() => editorStore.activeTabId)
const dirtyCount = computed(() => tabs.value.filter(t => t.isDirty).length)
const emit = defineEmits<{
  (e: 'select-tab', id: string): void
  (e: 'close-tab', id: string): void
}>()
</script>

<template>
  <div class="file-list-panel h-full flex flex-col bg-[#fafafa] dark:bg-[#252526] border-r border-gray-300 dark:border-gray-700 select-none font-[Microsoft YaHei]">
    <div class="px-2 py-1 text-[11px] font-semibold text-gray-600 dark:text-gray-300 border-b border-gray-200 dark:border-gray-700 bg-[#f0f0f0] dark:bg-[#2d2d2d] flex items-center justify-between">
      <span>文件</span>
      <span class="text-[10px] text-gray-400">{{ tabs.length }} 个{{ dirtyCount > 0 ? ` (${dirtyCount} 未保存)` : '' }}</span>
    </div>
    <div class="flex-1 overflow-auto">
      <div
        v-for="tab in tabs" :key="tab.id"
        class="file-list-item flex items-center px-2 py-0.5 cursor-pointer hover:bg-blue-50 dark:hover:bg-[#2a2d2e] group"
        :class="tab.id === activeId ? 'bg-blue-100 dark:bg-[#094771] border-l-2 border-blue-500' : 'border-l-2 border-transparent'"
        @click="emit('select-tab', tab.id)">
        <FileText class="w-3.5 h-3.5 mr-1.5 text-gray-400 flex-shrink-0" />
        <span class="flex-1 text-[12px] truncate" :title="tab.path">{{ tab.name }}</span>
        <Circle v-if="tab.isDirty" class="w-2 h-2 text-orange-500 fill-current flex-shrink-0 ml-1" />
        <button class="opacity-0 group-hover:opacity-100 p-0.5 rounded hover:bg-gray-300 dark:hover:bg-gray-600 flex-shrink-0 ml-1" @click.stop="emit('close-tab', tab.id)">
          <X class="w-3 h-3" />
        </button>
      </div>
      <div v-if="tabs.length === 0" class="p-4 text-[11px] text-gray-400 text-center">没有打开的文件</div>
    </div>
  </div>
</template>
