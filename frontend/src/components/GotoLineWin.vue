<script lang="ts" setup>
import { ref, watch, nextTick } from 'vue'
import { useEditorStore } from '@/stores'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const editorStore = useEditorStore()
const lineNum = ref('')
const side = ref<'left' | 'right'>('left')
const inputRef = ref<HTMLInputElement | null>(null)

const totalLines = ref(0)
watch(() => props.visible, async (v) => {
  if (v) {
    const tab = editorStore.activeTab
    totalLines.value = tab ? tab.content.split('\n').length : 0
    lineNum.value = ''
    side.value = 'left'
    await nextTick()
    inputRef.value?.focus()
  }
})

function ok() {
  const n = parseInt(lineNum.value)
  if (isNaN(n) || n < 1) return
  document.dispatchEvent(new CustomEvent('editor-command', {
    detail: { cmd: 'goto-line', args: [n, side.value] }
  }))
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-50 flex items-start justify-center pt-32 bg-black/10" @click.self="emit('close')">
      <div class="bg-white dark:bg-[#2d2d2d] border border-gray-300 dark:border-gray-600 rounded shadow-2xl w-64">
        <div class="px-3 py-1.5 border-b border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-[#3c3c3c] text-sm font-medium text-gray-700 dark:text-gray-200">
          转到行
        </div>
        <div class="p-3 space-y-2">
          <div class="flex items-center gap-2">
            <label class="text-xs text-gray-600 dark:text-gray-300 w-12">行号:</label>
            <input
              ref="inputRef"
              v-model="lineNum"
              type="number"
              min="1"
              :max="totalLines"
              class="flex-1 px-2 py-1 text-sm border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 focus:outline-none focus:border-blue-400"
              @keydown.enter="ok"
              @keydown.escape="emit('close')"
            />
          </div>
          <div class="flex items-center gap-4 pl-12">
            <label class="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-300 cursor-pointer">
              <input type="radio" v-model="side" value="left" /> Left
            </label>
            <label class="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-300 cursor-pointer">
              <input type="radio" v-model="side" value="right" /> Right
            </label>
          </div>
          <div class="text-xs text-gray-400 pl-12">共 {{ totalLines }} 行</div>
          <div class="flex justify-center gap-2 pt-1">
            <button class="ndd-btn-primary" @click="ok">确定</button>
            <button class="ndd-btn" @click="emit('close')">关闭</button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.ndd-btn {
  padding: 4px 16px;
  font-size: 12px;
  border: 1px solid #d1d5db;
  background: #fff;
  color: #374151;
  border-radius: 3px;
  cursor: pointer;
}
.ndd-btn:hover { background: #f3f4f6; }
.ndd-btn-primary {
  padding: 4px 16px;
  font-size: 12px;
  border: 1px solid #3b82f6;
  background: #3b82f6;
  color: #fff;
  border-radius: 3px;
  cursor: pointer;
}
.ndd-btn-primary:hover { background: #2563eb; }
html.dark .ndd-btn { background: #3c3c3c; color: #e0e0e0; border-color: #555; }
html.dark .ndd-btn:hover { background: #4c4c4c; }
</style>
