<script lang="ts" setup>
import { computed } from 'vue'
import { extractSymbols } from '@/utils/symbols'
import { FunctionSquare, Box } from 'lucide-vue-next'

const props = defineProps<{ content: string; language: string }>()
const emit = defineEmits<{ (e: 'goto', line: number): void }>()

const symbols = computed(() => extractSymbols(props.content, props.language))
</script>

<template>
  <div class="h-full flex flex-col bg-white dark:bg-[#252526] text-[var(--theme-fg)]">
    <div class="flex items-center justify-between px-3 py-2 border-b border-gray-200 dark:border-gray-700">
      <span class="text-sm font-medium text-gray-700 dark:text-gray-200">函数列表</span>
      <span class="text-xs text-gray-400">{{ symbols.length }}</span>
    </div>
    <div class="flex-1 overflow-auto">
      <div v-if="!symbols.length" class="p-4 text-xs text-gray-400 text-center">
        当前文件未识别到函数或类
      </div>
      <div
        v-for="s in symbols"
        :key="s.line + s.name"
        class="flex items-center gap-2 px-3 py-1 text-[13px] cursor-pointer hover:bg-gray-100 dark:hover:bg-[#2a2d2e]"
        @click="emit('goto', s.line)"
      >
        <Box v-if="s.kind === 'class'" class="w-3.5 h-3.5 text-purple-400 flex-shrink-0" />
        <FunctionSquare v-else class="w-3.5 h-3.5 text-blue-400 flex-shrink-0" />
        <span class="truncate">{{ s.name }}</span>
        <span class="ml-auto text-xs text-gray-400 flex-shrink-0">{{ s.line }}</span>
      </div>
    </div>
  </div>
</template>
