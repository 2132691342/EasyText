<script lang="ts" setup>
import { ref } from 'vue'
import { OpenDirectoryDialog, CompareDirectories } from '../../wailsjs/go/main/App'
import { X, FolderOpen } from 'lucide-vue-next'
import { ElMessage } from 'element-plus'

defineProps<{ visible: boolean }>()
const emit = defineEmits(['close'])

const leftDir = ref('')
const rightDir = ref('')
const entries = ref<any[]>([])
const loading = ref(false)

async function pickLeft() {
  try { const p = await OpenDirectoryDialog(); if (p) { leftDir.value = p; runCompare() } } catch (e) { console.warn(e) }
}
async function pickRight() {
  try { const p = await OpenDirectoryDialog(); if (p) { rightDir.value = p; runCompare() } } catch (e) { console.warn(e) }
}

async function runCompare() {
  if (!leftDir.value || !rightDir.value) return
  loading.value = true
  try {
    const r = await CompareDirectories(leftDir.value, rightDir.value)
    entries.value = r?.entries || []
    ElMessage.success(`对比完成：${entries.value.length} 项`)
  } catch (e: any) {
    ElMessage.error('目录对比失败：' + (e?.message || ''))
  } finally {
    loading.value = false
  }
}

function rowClass(e: any) {
  if (e.leftOnly) return 'bg-blue-50 dark:bg-blue-900/20'
  if (e.rightOnly) return 'bg-green-50 dark:bg-green-900/20'
  if (e.different) return 'bg-red-50 dark:bg-red-900/20'
  return ''
}
function statusText(e: any) {
  if (e.leftOnly) return '仅左侧'
  if (e.rightOnly) return '仅右侧'
  if (e.different) return '不同'
  if (e.identical) return '相同'
  return ''
}
</script>

<template>
  <div v-if="visible" class="fixed inset-0 z-[9000] flex items-center justify-center bg-black/40" @click.self="emit('close')">
    <div class="w-[900px] max-w-[92vw] h-[640px] max-h-[90vh] bg-white dark:bg-[#1e1e1e] rounded-lg shadow-2xl flex flex-col">
      <div class="flex items-center justify-between px-4 py-3 border-b border-gray-200 dark:border-gray-700">
        <span class="font-medium">目录对比</span>
        <button class="text-gray-400 hover:text-gray-600" @click="emit('close')"><X :size="18" /></button>
      </div>

      <div class="flex gap-2 px-4 py-3 border-b border-gray-200 dark:border-gray-700">
        <div class="flex-1 flex items-center gap-2">
          <span class="text-sm text-gray-500">左侧：</span>
          <input v-model="leftDir" readonly class="flex-1 px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-transparent" placeholder="选择左侧目录" />
          <button class="px-2 py-1 text-sm border rounded hover:bg-gray-100 dark:hover:bg-gray-700" @click="pickLeft"><FolderOpen :size="14" /></button>
        </div>
        <div class="flex-1 flex items-center gap-2">
          <span class="text-sm text-gray-500">右侧：</span>
          <input v-model="rightDir" readonly class="flex-1 px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-transparent" placeholder="选择右侧目录" />
          <button class="px-2 py-1 text-sm border rounded hover:bg-gray-100 dark:hover:bg-gray-700" @click="pickRight"><FolderOpen :size="14" /></button>
        </div>
      </div>

      <div class="flex-1 overflow-auto">
        <table v-if="entries.length" class="w-full text-sm">
          <thead class="sticky top-0 bg-gray-50 dark:bg-[#2a2a2a] text-gray-500">
            <tr>
              <th class="text-left px-3 py-2 font-normal">相对路径</th>
              <th class="text-left px-3 py-2 font-normal w-20">状态</th>
              <th class="text-right px-3 py-2 font-normal w-24">左大小</th>
              <th class="text-right px-3 py-2 font-normal w-24">右大小</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="e in entries" :key="e.relPath" :class="rowClass(e)" class="border-b border-gray-100 dark:border-gray-800">
              <td class="px-3 py-1.5 break-all">{{ e.relPath }}</td>
              <td class="px-3 py-1.5">{{ statusText(e) }}</td>
              <td class="px-3 py-1.5 text-right text-gray-500">{{ e.leftSize || '' }}</td>
              <td class="px-3 py-1.5 text-right text-gray-500">{{ e.rightSize || '' }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else-if="!loading" class="h-full flex items-center justify-center text-gray-400 text-sm">
          请选择左右两个目录进行对比
        </div>
        <div v-else class="h-full flex items-center justify-center text-gray-400 text-sm">对比中...</div>
      </div>
    </div>
  </div>
</template>
