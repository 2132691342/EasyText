<script lang="ts" setup>
import { ref, computed, watch } from 'vue'
import { useEditorStore } from '@/stores'
import { ChevronDown, ChevronUp, X, FolderSearch } from 'lucide-vue-next'
import { OpenDirectoryDialog } from '../../wailsjs/go/main/App'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits(['close'])

const editorStore = useEditorStore()

// Current tab: 'find' | 'files'
const activeTab = ref<'find' | 'files'>('find')

// Find & replace state
const searchQuery = ref('')
const replaceQuery = ref('')
const caseSensitive = ref(false)
const wholeWord = ref(false)
const useRegex = ref(false)

const results = ref<number[]>([])
const currentResultIndex = ref(-1)

const hasResults = computed(() => results.value.length > 0)
const resultText = computed(() => {
  if (results.value.length === 0) return '无结果'
  return `${currentResultIndex.value + 1}/${results.value.length}`
})

const activeEditorTab = computed(() => editorStore.activeTab)

// Find in files state
const findInFilesQuery = ref('')
const findInFilesDir = ref('')
const findInFilesResults = ref<{ file: string, line: number, content: string }[]>([])
const findInFilesLoading = ref(false)

// Search in current file
function performSearch() {
  if (!activeEditorTab.value || !searchQuery.value) {
    results.value = []
    currentResultIndex.value = -1
    return
  }

  const content = activeEditorTab.value.content
  const query = searchQuery.value

  results.value = []

  if (useRegex.value) {
    try {
      const regex = new RegExp(query, caseSensitive.value ? 'g' : 'gi')
      let match
      while ((match = regex.exec(content)) !== null) {
        results.value.push(match.index)
      }
    } catch (e) {
      results.value = []
    }
  } else {
    let searchContent = content
    let searchQueryLower = query

    if (!caseSensitive.value) {
      searchContent = content.toLowerCase()
      searchQueryLower = query.toLowerCase()
    }

    let pos = 0
    while (true) {
      const index = searchContent.indexOf(searchQueryLower, pos)
      if (index === -1) break

      if (wholeWord.value) {
        const beforeChar = index > 0 ? content[index - 1] : ''
        const afterChar = index + query.length < content.length ? content[index + query.length] : ''
        const isWordBoundary = !/[a-zA-Z0-9_]/.test(beforeChar) && !/[a-zA-Z0-9_]/.test(afterChar)
        if (isWordBoundary) {
          results.value.push(index)
        }
      } else {
        results.value.push(index)
      }

      pos = index + 1
    }
  }

  if (results.value.length > 0) {
    currentResultIndex.value = 0
  } else {
    currentResultIndex.value = -1
  }
}

function nextResult() {
  if (results.value.length === 0) return
  currentResultIndex.value = (currentResultIndex.value + 1) % results.value.length
}

function prevResult() {
  if (results.value.length === 0) return
  currentResultIndex.value = (currentResultIndex.value - 1 + results.value.length) % results.value.length
}

function replaceCurrent() {
  if (!activeEditorTab.value || currentResultIndex.value === -1 || !replaceQuery.value) return

  const content = activeEditorTab.value.content
  const start = results.value[currentResultIndex.value]
  const queryLength = searchQuery.value.length

  const newContent = content.substring(0, start) + replaceQuery.value + content.substring(start + queryLength)
  editorStore.updateTabContent(activeEditorTab.value.id, newContent)
  performSearch()
}

function replaceAll() {
  if (!activeEditorTab.value || results.value.length === 0 || !replaceQuery.value) return

  let content = activeEditorTab.value.content

  if (useRegex.value) {
    try {
      const regex = new RegExp(searchQuery.value, caseSensitive.value ? 'g' : 'gi')
      content = content.replace(regex, replaceQuery.value)
    } catch (e) {
      return
    }
  } else {
    if (caseSensitive.value) {
      content = content.split(searchQuery.value).join(replaceQuery.value)
    } else {
      const regex = new RegExp(searchQuery.value, 'gi')
      content = content.replace(regex, replaceQuery.value)
    }
  }

  editorStore.updateTabContent(activeEditorTab.value.id, content)
  results.value = []
  currentResultIndex.value = -1
}

// Find in files
async function openFindDir() {
  try {
    const dir = await OpenDirectoryDialog()
    if (dir) {
      findInFilesDir.value = dir
    }
  } catch {
    // ignore
  }
}

async function searchInFiles() {
  if (!findInFilesQuery.value || !findInFilesDir.value) return
  findInFilesLoading.value = true
  findInFilesResults.value = []
  
  // Client-side search using recursive directory traversal
  try {
    const { GetDirectoryTree } = await import('../../wailsjs/go/main/App')
    const tree = await GetDirectoryTree(findInFilesDir.value)
    if (!tree?.root) {
      findInFilesLoading.value = false
      return
    }

    // Collect all text files from the tree
    const textFiles: string[] = []
    function collectFiles(node: any) {
      if (!node.isDir && node.path) {
        const ext = (node.ext || '').toLowerCase()
        const textExts = ['txt','md','json','js','ts','html','css','xml','yaml','yml','toml','ini','cfg','go','java','py','c','cpp','h','hpp','rs','sh','bat','sql','vue','svelte','php','rb','swift','kt','scala','lua','r','pl','pm','tex','log','csv','env','gitignore','dockerfile']
        if (textExts.includes(ext)) {
          textFiles.push(node.path)
        }
      }
      if (node.children) {
        for (const child of node.children) {
          collectFiles(child)
        }
      }
    }
    collectFiles(tree.root)

    // Search in each file
    const query = findInFilesQuery.value.toLowerCase()
    const { ReadFile } = await import('../../wailsjs/go/main/App')
    
    for (const filePath of textFiles.slice(0, 200)) { // Limit to 200 files
      try {
        const result = await ReadFile(filePath)
        if (result?.content) {
          const lines = result.content.split('\n')
          for (let i = 0; i < lines.length; i++) {
            if (lines[i].toLowerCase().includes(query)) {
              findInFilesResults.value.push({
                file: filePath,
                line: i + 1,
                content: lines[i].trim().substring(0, 200),
              })
              if (findInFilesResults.value.length >= 500) break // Limit results
            }
          }
        }
      } catch {
        // Skip unreadable files
      }
      if (findInFilesResults.value.length >= 500) break
    }
  } catch (error) {
    console.error('Find in files error:', error)
  }
  
  findInFilesLoading.value = false
}

// Watch for search query changes
watch(searchQuery, () => {
  performSearch()
})

function close() {
  emit('close')
}

function handleKeyDown(e: KeyboardEvent) {
  if (e.key === 'Enter') {
    if (e.shiftKey) {
      prevResult()
    } else {
      nextResult()
    }
  } else if (e.key === 'Escape') {
    close()
  }
}
</script>

<template>
  <div v-if="visible" class="find-replace-panel bg-white dark:bg-[#252526] border-b border-gray-200 dark:border-gray-700" @keydown="handleKeyDown">
    <!-- Tab bar -->
    <div class="flex border-b border-gray-200 dark:border-gray-700 px-2">
      <button
        class="px-3 py-1.5 text-xs border-b-2 -mb-px transition-colors"
        :class="activeTab === 'find' ? 'border-blue-500 text-blue-600 dark:text-blue-400' : 'border-transparent text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'"
        @click="activeTab = 'find'"
      >查找替换</button>
      <button
        class="px-3 py-1.5 text-xs border-b-2 -mb-px transition-colors"
        :class="activeTab === 'files' ? 'border-blue-500 text-blue-600 dark:text-blue-400' : 'border-transparent text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'"
        @click="activeTab = 'files'"
      >在文件中查找</button>
    </div>

    <!-- Find/Replace tab -->
    <div v-if="activeTab === 'find'" class="p-2">
      <div class="flex items-center gap-2">
        <div class="flex-1 relative">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="搜索..."
            class="w-full px-3 py-1.5 text-sm bg-gray-50 dark:bg-[#3c3c3c] border border-gray-200 dark:border-gray-600 rounded focus:outline-none focus:border-blue-500"
          />
          <div class="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1 text-xs text-gray-400">
            <span v-if="hasResults">{{ resultText }}</span>
          </div>
        </div>

        <!-- Search options -->
        <button class="p-1 rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c]" :class="{ 'bg-blue-100 dark:bg-blue-900': caseSensitive }" title="区分大小写" @click="caseSensitive = !caseSensitive">Aa</button>
        <button class="p-1 rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c]" :class="{ 'bg-blue-100 dark:bg-blue-900': wholeWord }" title="全词匹配" @click="wholeWord = !wholeWord">W</button>
        <button class="p-1 rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c]" :class="{ 'bg-blue-100 dark:bg-blue-900': useRegex }" title="正则表达式" @click="useRegex = !useRegex">.*</button>

        <button class="p-1 rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c]" title="上一个" @click="prevResult" :disabled="!hasResults">
          <ChevronUp class="w-4 h-4" />
        </button>
        <button class="p-1 rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c]" title="下一个" @click="nextResult" :disabled="!hasResults">
          <ChevronDown class="w-4 h-4" />
        </button>
        <button class="p-1 rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c]" @click="close">
          <X class="w-4 h-4" />
        </button>
      </div>

      <!-- Replace input -->
      <div class="flex items-center gap-2 mt-2">
        <div class="flex-1 relative">
          <input v-model="replaceQuery" type="text" placeholder="替换..." class="w-full px-3 py-1.5 text-sm bg-gray-50 dark:bg-[#3c3c3c] border border-gray-200 dark:border-gray-600 rounded focus:outline-none focus:border-blue-500" />
        </div>
        <button class="px-2 py-1 text-sm rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c]" @click="replaceCurrent" :disabled="!hasResults">替换</button>
        <button class="px-2 py-1 text-sm rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c]" @click="replaceAll" :disabled="!hasResults">全部替换</button>
      </div>
    </div>

    <!-- Find in Files tab -->
    <div v-if="activeTab === 'files'" class="p-2">
      <div class="flex items-center gap-2">
        <div class="flex-1">
          <input v-model="findInFilesQuery" type="text" placeholder="搜索内容..." class="w-full px-3 py-1.5 text-sm bg-gray-50 dark:bg-[#3c3c3c] border border-gray-200 dark:border-gray-600 rounded focus:outline-none focus:border-blue-500" />
        </div>
        <button class="p-1 rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c]" @click="close"><X class="w-4 h-4" /></button>
      </div>
      <div class="flex items-center gap-2 mt-2">
        <div class="flex-1">
          <input v-model="findInFilesDir" type="text" readonly placeholder="选择搜索目录..." class="w-full px-3 py-1.5 text-sm bg-gray-50 dark:bg-[#3c3c3c] border border-gray-200 dark:border-gray-600 rounded focus:outline-none cursor-pointer" @click="openFindDir" />
        </div>
        <button class="px-3 py-1.5 text-sm bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50" :disabled="!findInFilesQuery || !findInFilesDir || findInFilesLoading" @click="searchInFiles">
          {{ findInFilesLoading ? '搜索中...' : '搜索' }}
        </button>
      </div>
      <!-- Results -->
      <div v-if="findInFilesResults.length > 0" class="mt-2 border border-gray-200 dark:border-gray-600 rounded overflow-hidden">
        <div class="text-xs text-gray-500 px-2 py-1 bg-gray-50 dark:bg-[#2d2d2d] border-b border-gray-200 dark:border-gray-600">
          找到 {{ findInFilesResults.length }} 处匹配
        </div>
        <div class="max-h-96 overflow-auto">
          <div v-for="(item, idx) in findInFilesResults.slice(0, 100)" :key="idx" class="px-3 py-1.5 text-xs border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-[#2d2d2d] cursor-pointer">
            <div class="text-blue-600 dark:text-blue-400 font-medium">{{ item.file.split(/[/\\]/).pop() }} : {{ item.line }}</div>
            <div class="text-gray-600 dark:text-gray-400 truncate">{{ item.content }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.find-replace-panel {
  font-size: 13px;
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>