<script lang="ts" setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useEditorStore } from '@/stores'
import { GetSnippets, CreateSnippet, UpdateSnippet, DeleteSnippet } from '../../wailsjs/go/main/App'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Pencil, Trash2, ChevronRight, FileCode } from 'lucide-vue-next'
import type { Snippet } from '@/types'

const editorStore = useEditorStore()

const searchQuery = ref('')
const showCreate = ref(false)
const editingId = ref<number | null>(null)

// 表单
const form = ref({
  name: '',
  prefix: '',
  body: '',
  description: '',
  language: '',
})

const languageOptions = [
  { value: '', label: '全部语言' },
  { value: 'go', label: 'Go' },
  { value: 'typescript', label: 'TypeScript' },
  { value: 'javascript', label: 'JavaScript' },
  { value: 'python', label: 'Python' },
  { value: 'java', label: 'Java' },
  { value: 'rust', label: 'Rust' },
  { value: 'html', label: 'HTML' },
  { value: 'css', label: 'CSS' },
  { value: 'json', label: 'JSON' },
  { value: 'yaml', label: 'YAML' },
  { value: 'xml', label: 'XML' },
  { value: 'sql', label: 'SQL' },
  { value: 'shell', label: 'Shell' },
  { value: 'markdown', label: 'Markdown' },
]

const filteredSnippets = computed(() => {
  let list = editorStore.snippets
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(s => s.name.toLowerCase().includes(q) || s.prefix.toLowerCase().includes(q) || s.description.toLowerCase().includes(q))
  }
  return list
})

async function loadSnippets() {
  try {
    const result = await GetSnippets('')
    editorStore.snippets = result || []
  } catch (e) {
    console.error('Failed to load snippets:', e)
  }
}

function resetForm() {
  form.value = { name: '', prefix: '', body: '', description: '', language: '' }
  editingId.value = null
  showCreate.value = false
}

async function saveSnippet() {
  if (!form.value.name || !form.value.prefix || !form.value.body) {
    ElMessage.warning('名称、触发缩写和内容不能为空')
    return
  }
  try {
    if (editingId.value) {
      await UpdateSnippet({ ...form.value, id: editingId.value, createdAt: '', updatedAt: '' } as unknown as import('../../wailsjs/go/models').tools.SnippetEntry)
      ElMessage.success('片段已更新')
    } else {
      await CreateSnippet({ ...form.value, createdAt: '', updatedAt: '' } as unknown as import('../../wailsjs/go/models').tools.SnippetEntry)
      ElMessage.success('片段已创建')
    }
    resetForm()
    await loadSnippets()
  } catch (e: any) {
    ElMessage.error(`操作失败: ${e?.message || ''}`)
  }
}

function editSnippet(s: Snippet) {
  editingId.value = s.id
  form.value = { name: s.name, prefix: s.prefix, body: s.body, description: s.description, language: s.language }
  showCreate.value = true
}

async function removeSnippet(s: Snippet) {
  try {
    await ElMessageBox.confirm(`确定删除片段「${s.name}」？`, '确认删除', { type: 'warning' })
    await DeleteSnippet(s.id)
    ElMessage.success('已删除')
    await loadSnippets()
  } catch { /* 取消 */ }
}

function insertSnippet(s: Snippet) {
  const tab = editorStore.activeTab
  if (!tab) {
    ElMessage.warning('请先打开一个文件')
    return
  }
  // 发送插入片段命令到编辑器
  document.dispatchEvent(new CustomEvent('editor-command', {
    detail: { cmd: 'insert-snippet', args: [s.body] }
  }))
  ElMessage.success(`已插入片段: ${s.name}`)
}

onMounted(() => {
  loadSnippets()
})
</script>

<template>
  <div class="h-full flex flex-col bg-white dark:bg-[#1e1e1e]">
    <!-- 标题栏 -->
    <div class="flex items-center justify-between px-3 py-2 border-b border-gray-200 dark:border-gray-700">
      <span class="text-xs font-medium text-gray-600 dark:text-gray-300">代码片段</span>
      <button
        class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-400 hover:text-blue-500"
        title="新建片段"
        @click="showCreate = true; editingId = null; resetForm()"
      >
        <Plus class="w-3.5 h-3.5" />
      </button>
    </div>

    <!-- 搜索 -->
    <div class="px-2 py-1.5">
      <div class="flex items-center gap-1 px-2 py-1 bg-gray-100 dark:bg-[#2d2d2d] rounded">
        <Search class="w-3 h-3 text-gray-400" />
        <input
          v-model="searchQuery"
          placeholder="搜索片段…"
          class="flex-1 text-xs bg-transparent border-none outline-none text-gray-700 dark:text-gray-200"
        />
      </div>
    </div>

    <!-- 创建/编辑表单 -->
    <div v-if="showCreate" class="px-2 py-2 border-b border-gray-200 dark:border-gray-700 space-y-1.5">
      <input v-model="form.name" placeholder="名称" class="w-full px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 focus:outline-none focus:border-blue-400" />
      <input v-model="form.prefix" placeholder="触发缩写 (如 imp)" class="w-full px-2 py-1 text-xs font-mono border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 focus:outline-none focus:border-blue-400" />
      <select v-model="form.language" class="w-full px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 focus:outline-none">
        <option v-for="l in languageOptions" :key="l.value" :value="l.value">{{ l.label }}</option>
      </select>
      <textarea v-model="form.body" placeholder="片段内容 (支持 $1, $2, ${1:default} 占位符)" rows="4" class="w-full px-2 py-1 text-xs font-mono border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 focus:outline-none focus:border-blue-400 resize-none"></textarea>
      <input v-model="form.description" placeholder="描述 (可选)" class="w-full px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 focus:outline-none" />
      <div class="flex gap-2">
        <button class="flex-1 px-2 py-1 text-xs bg-blue-500 text-white rounded hover:bg-blue-600" @click="saveSnippet">
          {{ editingId ? '更新' : '创建' }}
        </button>
        <button class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700" @click="resetForm">
          取消
        </button>
      </div>
    </div>

    <!-- 片段列表 -->
    <div class="flex-1 overflow-auto">
      <div
        v-for="s in filteredSnippets" :key="s.id"
        class="group flex items-center px-3 py-1.5 hover:bg-gray-50 dark:hover:bg-[#2a2a2a] cursor-pointer border-b border-gray-100 dark:border-gray-800"
        @click="insertSnippet(s)"
      >
        <FileCode class="w-3.5 h-3.5 mr-2 text-gray-400 flex-shrink-0" />
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-1.5">
            <span class="text-xs text-gray-700 dark:text-gray-200 truncate">{{ s.name }}</span>
            <span v-if="s.language" class="text-[10px] px-1 rounded bg-gray-100 dark:bg-gray-700 text-gray-400">{{ s.language }}</span>
          </div>
          <div class="text-[10px] text-gray-400 font-mono truncate">{{ s.prefix }}</div>
        </div>
        <div class="hidden group-hover:flex items-center gap-1">
          <button class="p-0.5 text-gray-400 hover:text-blue-500" @click.stop="editSnippet(s)" title="编辑">
            <Pencil class="w-3 h-3" />
          </button>
          <button class="p-0.5 text-gray-400 hover:text-red-500" @click.stop="removeSnippet(s)" title="删除">
            <Trash2 class="w-3 h-3" />
          </button>
        </div>
      </div>
      <div v-if="filteredSnippets.length === 0" class="px-3 py-8 text-center text-xs text-gray-400">
        暂无片段，点击 + 创建
      </div>
    </div>
  </div>
</template>