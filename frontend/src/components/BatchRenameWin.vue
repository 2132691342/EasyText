<script lang="ts" setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { X, RefreshCw, Play, Eye, FileText, Search, Archive } from 'lucide-vue-next'
import { ListDirectory, OpenDirectoryDialog } from '../../wailsjs/go/main/App'
import type { file } from '../../wailsjs/go/models'

const props = defineProps<{
  visible: boolean
}>()
const emit = defineEmits<{
  (e: 'close'): void
}>()

// State
const directory = ref('')
const fileFilter = ref('*')
const pattern = ref('$N$E')
const startIndex = ref(1)
const step = ref(1)

// File list
const files = ref<file.FileInfo[]>([])
const previewItems = ref<{ oldPath: string; newName: string; isDir: boolean }[]>([])
const loading = ref(false)

// Common patterns
const commonPatterns = [
  { label: '$N$E (保持原名)', value: '$N$E' },
  { label: 'prefix_$I$E (前缀+序号)', value: 'prefix_$I$E' },
  { label: 'file_$I$E (序号命名)', value: 'file_$I$E' },
  { label: '$N_backup$E (添加后缀)', value: '$N_backup$E' },
  { label: '$N_$I$E (原名+序号)', value: '$N_$I$E' },
]

// Select directory
async function selectDirectory() {
  try {
    const path = await OpenDirectoryDialog()
    if (path) {
      directory.value = path
      await loadFiles()
    }
  } catch (e) {
    ElMessage.error('选择目录失败')
  }
}

// Load files
async function loadFiles() {
  if (!directory.value) return
  loading.value = true
  try {
    const fileList = await ListDirectory(directory.value)
    files.value = fileList.filter(f => !f.isDir)
    generatePreview()
  } catch (e) {
    ElMessage.error('加载文件列表失败')
  } finally {
    loading.value = false
  }
}

// Generate preview
function generatePreview() {
  previewItems.value = files.value.map((f, idx) => {
    const ext = f.name.includes('.') ? '.' + f.name.split('.').pop() : ''
    const baseName = ext ? f.name.slice(0, f.name.lastIndexOf(ext)) : f.name
    const index = startIndex.value + idx * step.value
    let newName = pattern.value
      .replace(/\$N/g, baseName)
      .replace(/\$E/g, ext)
      .replace(/\$I/g, String(index))
    return {
      oldPath: f.name,
      newName,
      isDir: false,
    }
  })
}

// Execute rename
async function executeRename() {
  if (previewItems.value.length === 0) {
    ElMessage.warning('没有文件可以重命名')
    return
  }

  let successCount = 0
  let failCount = 0

  for (const item of previewItems.value) {
    if (item.oldPath === item.newName) {
      successCount++
      continue
    }
    try {
      const oldPath = directory.value + '\\' + item.oldPath
      const newPath = directory.value + '\\' + item.newName
      const { RenameFile } = await import('../../wailsjs/go/main/App')
      await RenameFile(oldPath, newPath)
      successCount++
    } catch {
      failCount++
    }
  }

  ElMessage.success(`重命名完成: 成功 ${successCount} 个${failCount > 0 ? `, 失败 ${failCount} 个` : ''}`)
  if (successCount > 0) {
    await loadFiles()
  }
}

// Watch pattern changes
watch([pattern, startIndex, step], () => {
  if (files.value.length > 0) {
    generatePreview()
  }
})
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/30" @click.self="emit('close')">
      <div class="bg-white dark:bg-[#2d2d2d] rounded-lg shadow-2xl border border-gray-200 dark:border-gray-600 w-[700px] max-h-[80vh] flex flex-col">
        <!-- Header -->
        <div class="flex items-center justify-between px-4 py-3 border-b border-gray-200 dark:border-gray-700">
          <h3 class="text-sm font-semibold dark:text-gray-200">批量重命名</h3>
          <button class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-500" @click="emit('close')">
            <X class="w-4 h-4" />
          </button>
        </div>

        <!-- Directory selection -->
        <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700">
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-500 dark:text-gray-400 w-12">目录:</span>
            <input
              v-model="directory"
              class="flex-1 px-2 py-1 text-xs border border-gray-200 dark:border-gray-600 rounded bg-white dark:bg-[#3c3c3c] dark:text-gray-200"
              placeholder="选择目录..."
              readonly
              @click="selectDirectory"
            />
            <button class="px-2 py-1 text-xs bg-gray-100 dark:bg-gray-700 rounded hover:bg-gray-200 dark:hover:bg-gray-600 dark:text-gray-200" @click="selectDirectory">
              <Search class="w-3.5 h-3.5 inline mr-1" /> 浏览
            </button>
            <button 
              class="px-2 py-1 text-xs bg-gray-100 dark:bg-gray-700 rounded hover:bg-gray-200 dark:hover:bg-gray-600 dark:text-gray-200"
              :disabled="!directory"
              @click="loadFiles"
            >
              <RefreshCw class="w-3.5 h-3.5 inline mr-1" /> 刷新
            </button>
          </div>
        </div>

        <!-- Pattern -->
        <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 space-y-2">
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-500 dark:text-gray-400 w-12">模板:</span>
            <select
              v-model="pattern"
              class="flex-1 px-2 py-1 text-xs border border-gray-200 dark:border-gray-600 rounded bg-white dark:bg-[#3c3c3c] dark:text-gray-200"
            >
              <option v-for="p in commonPatterns" :key="p.value" :value="p.value">{{ p.label }}</option>
            </select>
            <input
              v-model="pattern"
              class="w-48 px-2 py-1 text-xs border border-gray-200 dark:border-gray-600 rounded bg-white dark:bg-[#3c3c3c] dark:text-gray-200 font-mono"
              placeholder="$N$E"
            />
          </div>
          <div class="flex items-center gap-4">
            <div class="flex items-center gap-1">
              <span class="text-xs text-gray-500 dark:text-gray-400">起始:</span>
              <input v-model.number="startIndex" type="number" min="0" class="w-16 px-2 py-1 text-xs border border-gray-200 dark:border-gray-600 rounded bg-white dark:bg-[#3c3c3c] dark:text-gray-200" />
            </div>
            <div class="flex items-center gap-1">
              <span class="text-xs text-gray-500 dark:text-gray-400">步长:</span>
              <input v-model.number="step" type="number" min="1" class="w-16 px-2 py-1 text-xs border border-gray-200 dark:border-gray-600 rounded bg-white dark:bg-[#3c3c3c] dark:text-gray-200" />
            </div>
            <span class="text-xs text-gray-400">变量: $N=原名 $E=扩展名 $I=序号</span>
          </div>
        </div>

        <!-- Preview list -->
        <div class="flex-1 overflow-auto px-4 py-2">
          <div v-if="loading" class="flex items-center justify-center py-10 text-xs text-gray-400">
            加载中...
          </div>
          <div v-else-if="previewItems.length === 0" class="flex flex-col items-center justify-center py-10 text-gray-400">
            <Archive class="w-8 h-8 mb-2" />
            <span class="text-xs">请先选择一个目录</span>
          </div>
          <table v-else class="w-full text-xs">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700 text-left text-gray-500 dark:text-gray-400">
                <th class="py-1 px-2 w-8">#</th>
                <th class="py-1 px-2">原文件名</th>
                <th class="py-1 px-2 text-gray-400">→</th>
                <th class="py-1 px-2">新文件名</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(item, idx) in previewItems"
                :key="idx"
                class="border-b border-gray-100 dark:border-gray-800 hover:bg-blue-50 dark:hover:bg-[#094771]"
                :class="{ 'text-green-600': item.oldPath === item.newName }"
              >
                <td class="py-1 px-2 text-gray-400">{{ idx + 1 }}</td>
                <td class="py-1 px-2 font-mono dark:text-gray-200">{{ item.oldPath }}</td>
                <td class="py-1 px-2 text-gray-400">→</td>
                <td class="py-1 px-2 font-mono" :class="item.oldPath !== item.newName ? 'text-blue-600 dark:text-blue-400 font-medium' : 'dark:text-gray-300'">
                  {{ item.newName }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Footer -->
        <div class="flex items-center justify-between px-4 py-3 border-t border-gray-200 dark:border-gray-700">
          <span class="text-xs text-gray-400">{{ previewItems.length }} 个文件</span>
          <div class="flex gap-2">
            <button class="px-3 py-1.5 text-xs bg-gray-100 dark:bg-gray-700 rounded hover:bg-gray-200 dark:hover:bg-gray-600 dark:text-gray-200" @click="emit('close')">
              关闭
            </button>
            <button
              class="px-3 py-1.5 text-xs bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
              :disabled="previewItems.length === 0"
              @click="executeRename"
            >
              <Play class="w-3.5 h-3.5 inline mr-1" /> 执行重命名
            </button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>
