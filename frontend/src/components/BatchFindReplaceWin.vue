<script lang="ts" setup>
import { ref, watch, computed } from 'vue'
import { useEditorStore } from '@/stores'
import { ElMessage } from 'element-plus'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const editorStore = useEditorStore()

interface Row {
  find: string
  replace: string
  marked: boolean
}

const findKeywords = ref('')
const replaceKeywords = ref('')
const rows = ref<Row[]>([])
const results = ref<{ file: string, line: number, content: string, keyword: string }[]>([])

const stats = computed(() => {
  return {
    total: rows.value.length,
    matched: results.value.length,
    files: new Set(results.value.map(r => r.file)).size,
  }
})

function fresh() {
  const finds = findKeywords.value.split(/\s+/).filter(s => s.length > 0)
  const replaces = replaceKeywords.value.split(/\s+/).filter(s => s.length > 0)
  const maxLen = Math.max(finds.length, replaces.length)
  rows.value = []
  for (let i = 0; i < maxLen; i++) {
    rows.value.push({
      find: finds[i] || '',
      replace: replaces[i] || '',
      marked: false,
    })
  }
  if (rows.value.length === 0) {
    ElMessage.info('请在文本框中输入关键字')
  } else {
    ElMessage.success(`已生成 ${rows.value.length} 条规则`)
  }
}

function swap() {
  const tmp = findKeywords.value
  findKeywords.value = replaceKeywords.value
  replaceKeywords.value = tmp
  for (const r of rows.value) {
    const t = r.find
    r.find = r.replace
    r.replace = t
  }
}

function findAll() {
  if (rows.value.length === 0) fresh()
  if (rows.value.length === 0) return
  results.value = []
  for (const tab of editorStore.tabs) {
    if (tab.viewType !== 'code') continue
    const lines = tab.content.split('\n')
    for (let i = 0; i < lines.length; i++) {
      for (const r of rows.value) {
        if (r.find && lines[i].includes(r.find)) {
          results.value.push({
            file: tab.name,
            line: i + 1,
            content: lines[i].trim().substring(0, 200),
            keyword: r.find,
          })
        }
      }
    }
  }
  ElMessage.success(`找到 ${results.value.length} 处匹配，涉及 ${stats.value.files} 个文件`)
}

function replaceAll() {
  if (rows.value.length === 0) fresh()
  if (rows.value.length === 0) return
  let total = 0
  for (const tab of editorStore.tabs) {
    if (tab.isReadOnly || tab.viewType !== 'code') continue
    let content = tab.content
    for (const r of rows.value) {
      if (r.find && r.replace) {
        const before = content
        content = content.split(r.find).join(r.replace)
        if (before !== content) total++
      }
    }
    if (content !== tab.content) {
      editorStore.updateTabContent(tab.id, content)
    }
  }
  ElMessage.success(`已批量替换 ${total} 处`)
}

function mark() {
  if (rows.value.length === 0) fresh()
  for (const r of rows.value) {
    r.marked = true
  }
  document.dispatchEvent(new CustomEvent('editor-command', {
    detail: { cmd: 'mark-keywords', args: [rows.value.map(r => r.find).filter(s => s)] }
  }))
  ElMessage.success('已标记所有关键字')
}

function clearMark() {
  for (const r of rows.value) {
    r.marked = false
  }
  document.dispatchEvent(new CustomEvent('editor-command', { detail: { cmd: 'clear-mark' } }))
  ElMessage.success('已清除所有标记')
}

async function importKeywords() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json'
  input.onchange = async () => {
    const f = input.files?.[0]
    if (!f) return
    try {
      const text = await f.text()
      const data = JSON.parse(text)
      if (Array.isArray(data)) {
        rows.value = data.map((d: any) => ({
          find: d.find || '',
          replace: d.replace || '',
          marked: false,
        }))
        findKeywords.value = rows.value.map(r => r.find).join('\n')
        replaceKeywords.value = rows.value.map(r => r.replace).join('\n')
        ElMessage.success(`已导入 ${rows.value.length} 条规则`)
      }
    } catch (e) {
      ElMessage.error(`导入失败: ${e}`)
    }
  }
  input.click()
}

async function exportKeywords() {
  const data = rows.value.map(r => ({ find: r.find, replace: r.replace }))
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'batch-find-replace.json'
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('已导出')
}

watch(() => props.visible, (v) => {
  if (v) {
    findKeywords.value = ''
    replaceKeywords.value = ''
    rows.value = []
    results.value = []
  }
})
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/30" @click.self="emit('close')">
      <div class="bg-white dark:bg-[#2d2d2d] border border-gray-300 dark:border-gray-600 rounded shadow-2xl w-[900px] flex flex-col" style="height: 700px;">
        <div class="px-3 py-1.5 border-b border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-[#3c3c3c] text-sm font-medium text-gray-700 dark:text-gray-200 flex items-center justify-between">
          <span>批量查找替换</span>
          <button class="p-0.5 hover:bg-gray-200 dark:hover:bg-gray-600 rounded" @click="emit('close')">×</button>
        </div>
        <div class="flex-1 flex flex-col p-3 gap-3 overflow-hidden">
          <!-- 顶部双栏 -->
          <div class="flex gap-3" style="height: 140px;">
            <div class="flex-1 flex flex-col">
              <label class="text-xs text-gray-600 dark:text-gray-300 mb-1">输入多个查找关键字，以空白字符分隔:</label>
              <textarea
                v-model="findKeywords"
                class="flex-1 px-2 py-1 text-xs font-mono border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 resize-none focus:outline-none focus:border-blue-400"
                placeholder="keyword1 keyword2 keyword3..."
              ></textarea>
            </div>
            <div class="flex-1 flex flex-col">
              <label class="text-xs text-gray-600 dark:text-gray-300 mb-1">输入多个替换关键字，以空白字符分隔:</label>
              <textarea
                v-model="replaceKeywords"
                class="flex-1 px-2 py-1 text-xs font-mono border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 resize-none focus:outline-none focus:border-blue-400"
                placeholder="replace1 replace2 replace3..."
              ></textarea>
            </div>
          </div>
          <!-- 中部表格 -->
          <div class="flex-1 border border-gray-200 dark:border-gray-600 rounded overflow-hidden flex flex-col min-h-0">
            <div class="overflow-auto flex-1">
              <table class="w-full text-xs">
                <thead class="bg-gray-50 dark:bg-[#2d2d2d] sticky top-0">
                  <tr class="border-b border-gray-200 dark:border-gray-600">
                    <th class="px-3 py-1.5 text-left text-gray-600 dark:text-gray-300 w-10">#</th>
                    <th class="px-3 py-1.5 text-left text-gray-600 dark:text-gray-300">关键字 (Find)</th>
                    <th class="px-3 py-1.5 text-left text-gray-600 dark:text-gray-300">替换 (Replace)</th>
                    <th class="px-3 py-1.5 text-left text-gray-600 dark:text-gray-300 w-16">标记</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(r, idx) in rows" :key="idx" class="border-b border-gray-100 dark:border-gray-700 hover:bg-blue-50 dark:hover:bg-[#3c3c3c]">
                    <td class="px-3 py-1 text-gray-400">{{ idx + 1 }}</td>
                    <td class="px-3 py-1 font-mono">{{ r.find }}</td>
                    <td class="px-3 py-1 font-mono">{{ r.replace }}</td>
                    <td class="px-3 py-1 text-center">
                      <span v-if="r.marked" class="text-yellow-500">★</span>
                      <span v-else class="text-gray-300">-</span>
                    </td>
                  </tr>
                  <tr v-if="rows.length === 0">
                    <td colspan="4" class="px-3 py-8 text-center text-gray-400">点击"刷新"按钮根据上方文本框生成规则</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          <!-- 结果统计 -->
          <div v-if="results.length > 0" class="text-xs text-gray-600 dark:text-gray-300 border border-gray-200 dark:border-gray-600 rounded p-2 max-h-32 overflow-auto">
            <div class="font-medium mb-1">查找结果 (共 {{ stats.matched }} 处匹配，涉及 {{ stats.files }} 个文件):</div>
            <div v-for="(r, idx) in results.slice(0, 50)" :key="idx" class="truncate">
              <span class="text-blue-600 dark:text-blue-400">{{ r.file }}:{{ r.line }}</span>
              <span class="text-gray-500 dark:text-gray-400 ml-2">[{{ r.keyword }}]</span>
              <span class="text-gray-600 dark:text-gray-300 ml-1">{{ r.content }}</span>
            </div>
          </div>
          <!-- 底部按钮 -->
          <div class="flex justify-center gap-2">
            <button class="ndd-btn" @click="fresh">刷新</button>
            <button class="ndd-btn" @click="swap">交换</button>
            <button class="ndd-btn-primary" @click="findAll">查找</button>
            <button class="ndd-btn-primary" @click="replaceAll">替换</button>
            <button class="ndd-btn" @click="mark">标记</button>
            <button class="ndd-btn" @click="clearMark">清除标记</button>
            <button class="ndd-btn" @click="importKeywords">导入</button>
            <button class="ndd-btn" @click="exportKeywords">导出</button>
            <button class="ndd-btn ml-4" @click="emit('close')">关闭</button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.ndd-btn {
  padding: 4px 12px;
  font-size: 12px;
  border: 1px solid #d1d5db;
  background: #fff;
  color: #374151;
  border-radius: 3px;
  cursor: pointer;
}
.ndd-btn:hover { background: #f3f4f6; }
.ndd-btn-primary {
  padding: 4px 12px;
  font-size: 12px;
  border: 1px solid #3b82f6;
  background: #3b82f6;
  color: #fff;
  border-radius: 3px;
  cursor: pointer;
}
.ndd-btn-primary:hover { background: #2563eb; }
html.dark .ndd-btn {
  background: #3c3c3c;
  color: #e0e0e0;
  border-color: #555;
}
html.dark .ndd-btn:hover { background: #4c4c4c; }
</style>
