<script lang="ts" setup>
import { useEditorStore } from '@/stores'
import { X } from 'lucide-vue-next'

defineProps<{ visible: boolean }>()
const emit = defineEmits(['close'])
const ed = useEditorStore()

function insert(text: string) {
  document.dispatchEvent(new CustomEvent('editor-command', { detail: { cmd: 'insert-text', args: [text] } }))
  emit('close')
}
</script>

<template>
  <div v-if="visible" class="fixed inset-0 z-[9000] flex items-center justify-center bg-black/40" @click.self="emit('close')">
    <div class="w-[520px] max-w-[92vw] h-[480px] max-h-[85vh] bg-white dark:bg-[#1e1e1e] rounded-lg shadow-2xl flex flex-col">
      <div class="flex items-center justify-between px-4 py-3 border-b border-gray-200 dark:border-gray-700">
        <span class="font-medium">剪贴板历史记录</span>
        <button class="text-gray-400 hover:text-gray-600" @click="emit('close')"><X :size="18" /></button>
      </div>
      <div class="flex-1 overflow-auto p-2">
        <div v-if="!ed.clipboardHistory.length" class="h-full flex items-center justify-center text-gray-400 text-sm">暂无记录</div>
        <div v-for="(t, i) in ed.clipboardHistory" :key="i"
          class="px-3 py-2 mb-1 border border-gray-200 dark:border-gray-700 rounded cursor-pointer hover:bg-blue-50 dark:hover:bg-blue-900/20"
          @click="insert(t)">
          <div class="text-xs text-gray-400 mb-1">#{{ i + 1 }} · {{ t.length }} 字符</div>
          <div class="text-sm break-all line-clamp-3">{{ t }}</div>
        </div>
      </div>
    </div>
  </div>
</template>
