<script lang="ts" setup>
import { ref, watch, onMounted } from 'vue'
import { GetRecentFiles, GetRecentFolders, OpenFileDialog, ClearRecentFiles, ClearRecentFolders } from '../../wailsjs/go/main/App'
import { useEditorStore } from '@/stores'
import { getFileExtension, getTabViewType } from '@/utils'
import { ReadFile, AddRecentEntry } from '../../wailsjs/go/main/App'
import { Clock, FileText, FolderOpen, FileSearch, FolderSearch, X } from 'lucide-vue-next'
import { ElMessage } from 'element-plus'
import type { RecentEntry } from '@/types'

const props = defineProps<{ visible: boolean; initialTab?: 'files' | 'folders' }>()
const emit = defineEmits<{
  (e: 'close'): void
  (e: 'open-file', path: string): void
}>()
const activeTab = ref<'files' | 'folders'>(props.initialTab || 'files')

const editorStore = useEditorStore()

const recentFiles = ref<RecentEntry[]>([])
const recentFolders = ref<RecentEntry[]>([])
const loading = ref(true)

async function loadData() {
  loading.value = true
  try {
    const [files, folders] = await Promise.all([GetRecentFiles(), GetRecentFolders()])
    // 显示上限放宽到 50，让 settingStore.recentFilesLimit 自定义值生效。
    // 后端默认 10，UI 不会硬截断用户在设置里填的更大数。
    recentFiles.value = (files || []).slice(0, 50)
    recentFolders.value = (folders || []).slice(0, 50)
  } catch { /* ignore */ }
  loading.value = false
}

async function openRecentFile(entry: RecentEntry) {
  emit('open-file', entry.path)
  emit('close')
}

async function openRecentFolder(entry: RecentEntry) {
  try {
    const { useFileStore } = await import('@/stores')
    useFileStore().setDirectory(entry.path)
    await AddRecentEntry(entry.path, true)
    emit('close')
  } catch (e: any) {
    ElMessage.error('打开文件夹失败: ' + (e?.message || ''))
  }
}

async function browseFile() {
  try {
    const path = await OpenFileDialog()
    if (path) {
      emit('open-file', path)
      emit('close')
    }
  } catch (e) { console.warn(e) }
}

async function clearRecent(section: 'files' | 'folders') {
  if (section === 'files') {
    await ClearRecentFiles()
    recentFiles.value = []
  } else {
    await ClearRecentFolders()
    recentFolders.value = []
  }
}

// 对话框打开时重置标签页并加载数据
watch(() => props.visible, (v) => {
  if (v) {
    activeTab.value = props.initialTab || 'files'
    loadData()
  }
})

onMounted(() => { if (props.visible) loadData() })
</script>

<template>
  <div class="flex flex-col h-full overflow-hidden">
    <!-- Tab switcher -->
    <div class="flex border-b border-gray-200 dark:border-gray-700 px-3">
      <button
        class="px-3 py-2 text-xs font-medium border-b-2 transition-colors"
        :class="activeTab === 'files' 
          ? 'border-blue-500 text-blue-600 dark:text-blue-400' 
          : 'border-transparent text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'"
        @click="activeTab = 'files'"
      >
        <span class="flex items-center gap-1.5">
          <FileText class="w-3.5 h-3.5" />
          最近文件
          <span v-if="recentFiles.length" class="text-[10px] px-1 rounded bg-gray-100 dark:bg-gray-700">{{ recentFiles.length }}</span>
        </span>
      </button>
      <button
        class="px-3 py-2 text-xs font-medium border-b-2 transition-colors"
        :class="activeTab === 'folders' 
          ? 'border-blue-500 text-blue-600 dark:text-blue-400' 
          : 'border-transparent text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'"
        @click="activeTab = 'folders'"
      >
        <span class="flex items-center gap-1.5">
          <FolderOpen class="w-3.5 h-3.5" />
          最近文件夹
          <span v-if="recentFolders.length" class="text-[10px] px-1 rounded bg-gray-100 dark:bg-gray-700">{{ recentFolders.length }}</span>
        </span>
      </button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center flex-1 py-8">
      <span class="text-sm text-gray-400">加载中...</span>
    </div>

    <!-- File list -->
    <div v-else-if="activeTab === 'files'" class="flex-1 overflow-auto">
      <div v-if="recentFiles.length === 0" class="flex flex-col items-center justify-center py-10 text-gray-400">
        <FileSearch class="w-10 h-10 mb-3 opacity-40" />
        <p class="text-sm">暂无最近打开的文件</p>
        <button class="mt-3 px-3 py-1.5 text-xs bg-blue-500 text-white rounded hover:bg-blue-600" @click="browseFile">
          浏览文件...
        </button>
      </div>
      <div v-else class="py-1">
        <div
          v-for="entry in recentFiles" :key="entry.path"
          class="flex items-center gap-2.5 px-4 py-2 hover:bg-gray-50 dark:hover:bg-[#2a2a2a] cursor-pointer group transition-colors"
          @click="openRecentFile(entry)"
          :title="entry.path"
        >
          <FileText class="w-4 h-4 text-gray-400 flex-shrink-0" />
          <div class="flex-1 min-w-0">
            <div class="text-sm text-gray-700 dark:text-gray-200 truncate">{{ entry.name }}</div>
            <div class="text-[11px] text-gray-400 truncate">{{ entry.path }}</div>
          </div>
        </div>
      </div>
      <div v-if="recentFiles.length > 0" class="px-4 py-2 border-t border-gray-100 dark:border-gray-800">
        <button class="text-[11px] text-gray-400 hover:text-red-400 flex items-center gap-1" @click="clearRecent('files')">
          <X class="w-3 h-3" /> 清除记录
        </button>
      </div>
    </div>

    <!-- Folder list -->
    <div v-else class="flex-1 overflow-auto">
      <div v-if="recentFolders.length === 0" class="flex flex-col items-center justify-center py-10 text-gray-400">
        <FolderSearch class="w-10 h-10 mb-3 opacity-40" />
        <p class="text-sm">暂无最近打开的文件夹</p>
      </div>
      <div v-else class="py-1">
        <div
          v-for="entry in recentFolders" :key="entry.path"
          class="flex items-center gap-2.5 px-4 py-2 hover:bg-gray-50 dark:hover:bg-[#2a2a2a] cursor-pointer group transition-colors"
          @click="openRecentFolder(entry)"
          :title="entry.path"
        >
          <FolderOpen class="w-4 h-4 text-yellow-500 flex-shrink-0" />
          <div class="flex-1 min-w-0">
            <div class="text-sm text-gray-700 dark:text-gray-200 truncate">{{ entry.name }}</div>
            <div class="text-[11px] text-gray-400 truncate">{{ entry.path }}</div>
          </div>
        </div>
      </div>
      <div v-if="recentFolders.length > 0" class="px-4 py-2 border-t border-gray-100 dark:border-gray-800">
        <button class="text-[11px] text-gray-400 hover:text-red-400 flex items-center gap-1" @click="clearRecent('folders')">
          <X class="w-3 h-3" /> 清除记录
        </button>
      </div>
    </div>
  </div>
</template>
