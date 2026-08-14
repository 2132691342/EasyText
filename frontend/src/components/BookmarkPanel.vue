<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue'
import { useEditorStore } from '@/stores'
import { GetAllBookmarks, RemoveBookmark, UpdateBookmarkNote, ReadFile } from '../../wailsjs/go/main/App'
import { ElMessage } from 'element-plus'
import { Search, Trash2, Pencil, ArrowRight, Bookmark } from 'lucide-vue-next'
import type { BookmarkEntry } from '@/types'

const editorStore = useEditorStore()

const searchQuery = ref('')
const groupByFile = ref(true)
const allBookmarks = ref<BookmarkEntry[]>([])
const editingNoteId = ref<number | null>(null)
const editingNote = ref('')

const grouped = computed(() => {
  const map = new Map<string, BookmarkEntry[]>()
  for (const b of filteredBookmarks.value) {
    const list = map.get(b.filePath) || []
    list.push(b)
    map.set(b.filePath, list)
  }
  return Array.from(map.entries())
})

const filteredBookmarks = computed(() => {
  if (!searchQuery.value) return allBookmarks.value
  const q = searchQuery.value.toLowerCase()
  return allBookmarks.value.filter(b =>
    b.filePath.toLowerCase().includes(q) ||
    b.note.toLowerCase().includes(q) ||
    b.tag.toLowerCase().includes(q)
  )
})

async function loadBookmarks() {
  try {
    const result = await GetAllBookmarks()
    const list: BookmarkEntry[] = []
    if (result) {
      for (const [filePath, entries] of Object.entries(result as Record<string, BookmarkEntry[]>)) {
        list.push(...entries)
      }
    }
    allBookmarks.value = list
  } catch (e) {
    console.error('Failed to load bookmarks:', e)
  }
}

async function jumpToBookmark(b: BookmarkEntry) {
  const tab = editorStore.getTabByPath(b.filePath)
  if (tab) {
    editorStore.activateTab(tab.id)
    document.dispatchEvent(new CustomEvent('editor-command', {
      detail: { cmd: 'scroll-to-line', args: [b.lineNumber] }
    }))
  } else {
    // 打开文件然后跳转
    try {
      const result = await ReadFile(b.filePath)
      if (result) {
        editorStore.createTab(b.filePath, result.content, result.info.encoding, result.info.lineEnding)
        setTimeout(() => {
          document.dispatchEvent(new CustomEvent('editor-command', {
            detail: { cmd: 'scroll-to-line', args: [b.lineNumber] }
          }))
        }, 200)
      }
    } catch (e: any) {
      ElMessage.error(`无法打开文件: ${b.filePath}`)
    }
  }
}

async function removeBookmark(b: BookmarkEntry) {
  try {
    await RemoveBookmark(b.id)
    ElMessage.success('书签已删除')
    await loadBookmarks()
  } catch (e: any) {
    ElMessage.error(`删除失败: ${e?.message || ''}`)
  }
}

function startEditNote(b: BookmarkEntry) {
  editingNoteId.value = b.id
  editingNote.value = b.note
}

async function saveNote(b: BookmarkEntry) {
  if (editingNote.value === b.note) {
    editingNoteId.value = null
    return
  }
  try {
    await UpdateBookmarkNote(b.id, editingNote.value)
    b.note = editingNote.value
    ElMessage.success('备注已更新')
  } catch (e: any) {
    ElMessage.error(`更新失败: ${e?.message || ''}`)
  }
  editingNoteId.value = null
}

function getFileName(path: string): string {
  return path.split(/[/\\]/).pop() || path
}

onMounted(() => {
  loadBookmarks()
})
</script>

<template>
  <div class="h-full flex flex-col bg-white dark:bg-[#1e1e1e]">
    <!-- 标题栏 -->
    <div class="flex items-center justify-between px-3 py-2 border-b border-gray-200 dark:border-gray-700">
      <span class="text-xs font-medium text-gray-600 dark:text-gray-300">书签</span>
      <div class="flex items-center gap-1">
        <button
          class="p-1 rounded text-xs"
          :class="groupByFile ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-600' : 'text-gray-400 hover:text-gray-600'"
          @click="groupByFile = !groupByFile"
          title="按文件分组"
        >
          <Bookmark class="w-3 h-3" />
        </button>
      </div>
    </div>

    <!-- 搜索 -->
    <div class="px-2 py-1.5">
      <div class="flex items-center gap-1 px-2 py-1 bg-gray-100 dark:bg-[#2d2d2d] rounded">
        <Search class="w-3 h-3 text-gray-400" />
        <input
          v-model="searchQuery"
          placeholder="搜索书签…"
          class="flex-1 text-xs bg-transparent border-none outline-none text-gray-700 dark:text-gray-200"
        />
      </div>
    </div>

    <!-- 书签列表 -->
    <div class="flex-1 overflow-auto">
      <div v-if="groupByFile">
        <div v-for="[filePath, entries] in grouped" :key="filePath">
          <div class="px-3 py-1 text-[10px] font-medium text-gray-400 bg-gray-50 dark:bg-[#252525] border-b border-gray-100 dark:border-gray-800 truncate">
            {{ getFileName(filePath) }}
          </div>
          <div
            v-for="b in entries" :key="b.id"
            class="group flex items-center px-3 py-1.5 hover:bg-gray-50 dark:hover:bg-[#2a2a2a] cursor-pointer border-b border-gray-50 dark:border-gray-800/50"
            @click="jumpToBookmark(b)"
          >
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-1.5">
                <span class="text-xs text-gray-700 dark:text-gray-200">行 {{ b.lineNumber }}</span>
                <span v-if="b.tag" class="text-[10px] px-1 rounded bg-blue-50 dark:bg-blue-900/20 text-blue-500">{{ b.tag }}</span>
              </div>
              <div v-if="editingNoteId === b.id" class="mt-0.5" @click.stop>
                <input
                  v-model="editingNote"
                  class="w-full px-1 py-0.5 text-[10px] border border-blue-400 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 focus:outline-none"
                  @keydown.enter="saveNote(b)"
                  @keydown.escape="editingNoteId = null"
                  @blur="saveNote(b)"
                />
              </div>
              <div v-else-if="b.note" class="text-[10px] text-gray-400 truncate">{{ b.note }}</div>
            </div>
            <div class="hidden group-hover:flex items-center gap-1">
              <button class="p-0.5 text-gray-400 hover:text-blue-500" @click.stop="startEditNote(b)" title="编辑备注">
                <Pencil class="w-3 h-3" />
              </button>
              <button class="p-0.5 text-gray-400 hover:text-red-500" @click.stop="removeBookmark(b)" title="删除">
                <Trash2 class="w-3 h-3" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 非分组视图 -->
      <div v-else>
        <div
          v-for="b in filteredBookmarks" :key="b.id"
          class="group flex items-center px-3 py-1.5 hover:bg-gray-50 dark:hover:bg-[#2a2a2a] cursor-pointer border-b border-gray-50 dark:border-gray-800/50"
          @click="jumpToBookmark(b)"
        >
          <div class="flex-1 min-w-0">
            <div class="text-xs text-gray-700 dark:text-gray-200 truncate">{{ getFileName(b.filePath) }}:{{ b.lineNumber }}</div>
            <div v-if="b.note" class="text-[10px] text-gray-400 truncate">{{ b.note }}</div>
          </div>
          <div class="hidden group-hover:flex items-center gap-1">
            <button class="p-0.5 text-gray-400 hover:text-red-500" @click.stop="removeBookmark(b)" title="删除">
              <Trash2 class="w-3 h-3" />
            </button>
          </div>
        </div>
      </div>

      <div v-if="allBookmarks.length === 0" class="px-3 py-8 text-center text-xs text-gray-400">
        暂无书签<br/>
        <span class="text-[10px]">在编辑器中按 Ctrl+F2 添加书签</span>
      </div>
    </div>
  </div>
</template>