<script lang="ts" setup>
import { ref, watch } from 'vue'
import {
  OpenDirectoryDialog, GetDirectoryTree, ReadFileBytes, SaveFileBytes,
  DetectEncoding, ConvertEncoding,
} from '../../wailsjs/go/main/App'
import { ElMessage } from 'element-plus'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'close'): void }>()

interface FileRow {
  path: string
  size: string
  fileCode: string
  convertCode: string
  result: string
}

const files = ref<FileRow[]>([])
const targetCode = ref('UTF-8')
const extFilter = ref('all')
const logText = ref('')
const loading = ref(false)

const targetCodeOptions = ['UTF-8', 'UTF-8-BOM', 'UTF-16-LE', 'UTF-16-BE', 'GBK']
const extOptions = [
  { value: 'all', label: '所有支持的文件扩展名' },
  { value: 'txt', label: '*.txt' },
  { value: 'cpp', label: '*.cpp;*.h' },
  { value: 'go', label: '*.go' },
  { value: 'py', label: '*.py' },
  { value: 'json', label: '*.json' },
  { value: 'xml', label: '*.xml' },
  { value: 'html', label: '*.html' },
  { value: 'css', label: '*.css' },
  { value: 'js', label: '*.js' },
  { value: 'md', label: '*.md' },
  { value: 'log', label: '*.log' },
]

const textExts = ['txt','md','json','js','ts','html','css','xml','yaml','yml','toml','ini','cfg','go','java','py','c','cpp','h','hpp','rs','sh','bat','sql','vue','svelte','php','rb','swift','kt','scala','lua','r','pl','pm','tex','log','csv','env','gitignore']

function appendLog(msg: string) {
  logText.value += msg + '\n'
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`
}

async function selectDir() {
  try {
    const dir = await OpenDirectoryDialog()
    if (!dir) return
    files.value = []
    logText.value = `扫描目录: ${dir}\n`
    loading.value = true
    
    const tree = await GetDirectoryTree(dir)
    if (!tree?.root) {
      appendLog('无法读取目录')
      loading.value = false
      return
    }
    
    const collected: FileRow[] = []
    function walk(node: any) {
      if (!node) return
      if (!node.isDir && node.path) {
        const name = node.name || ''
        const ext = (name.split('.').pop() || '').toLowerCase()
        // 扩展名过滤
        if (extFilter.value !== 'all') {
          const allowed = extOptions.find(o => o.value === extFilter.value)?.label || ''
          const allowedExts = allowed.match(/\*\.(\w+)/g)?.map(s => s.replace('*.', '').toLowerCase()) || []
          if (!allowedExts.includes(ext)) return
        } else {
          if (!textExts.includes(ext)) return
        }
        collected.push({
          path: node.path,
          size: formatSize(node.size || 0),
          fileCode: '',
          convertCode: targetCode.value,
          result: '待转换',
        })
      }
      if (node.children) {
        for (const c of node.children) walk(c)
      }
    }
    walk(tree.root)
    
    // 检测每个文件的编码
    for (const f of collected) {
      try {
        const data = await ReadFileBytes(f.path)
        const detected = await DetectEncoding(data)
        f.fileCode = detected || 'unknown'
      } catch {
        f.fileCode = 'unknown'
      }
    }
    files.value = collected
    appendLog(`扫描完成，共 ${collected.length} 个文件`)
  } catch (e) {
    appendLog(`扫描失败: ${e}`)
  }
  loading.value = false
}

async function startConvert() {
  if (files.value.length === 0) {
    ElMessage.warning('请先选择目录')
    return
  }
  loading.value = true
  appendLog(`\n开始转换为目标编码: ${targetCode.value}`)
  
  let success = 0
  let failed = 0
  let skipped = 0
  
  for (const f of files.value) {
    f.convertCode = targetCode.value
    if (f.fileCode === targetCode.value || f.fileCode === 'unknown') {
      f.result = '跳过(同编码或未知)'
      skipped++
      appendLog(`跳过: ${f.path}`)
      continue
    }
    try {
      const data = await ReadFileBytes(f.path)
      const converted = await ConvertEncoding(data, f.fileCode, targetCode.value)
      await SaveFileBytes(f.path, converted)
      f.result = '成功'
      success++
      appendLog(`成功: ${f.path}`)
    } catch (e) {
      f.result = `失败: ${e}`
      failed++
      appendLog(`失败: ${f.path} - ${e}`)
    }
  }
  
  appendLog(`\n转换完成: 成功 ${success}, 失败 ${failed}, 跳过 ${skipped}`)
  ElMessage.success(`转换完成: 成功 ${success}, 失败 ${failed}, 跳过 ${skipped}`)
  loading.value = false
}

watch(() => props.visible, (v) => {
  if (v) {
    files.value = []
    logText.value = ''
    targetCode.value = 'UTF-8'
    extFilter.value = 'all'
  }
})
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/30" @click.self="emit('close')">
      <div class="bg-white dark:bg-[#2d2d2d] border border-gray-300 dark:border-gray-600 rounded shadow-2xl w-[960px] flex flex-col" style="height: 620px;">
        <div class="px-3 py-1.5 border-b border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-[#3c3c3c] text-sm font-medium text-gray-700 dark:text-gray-200 flex items-center justify-between">
          <span>批量编码转换</span>
          <button class="p-0.5 hover:bg-gray-200 dark:hover:bg-gray-600 rounded" @click="emit('close')">×</button>
        </div>
        <div class="flex-1 flex flex-col p-3 gap-3 overflow-hidden">
          <!-- 文件列表 -->
          <div class="flex-1 border border-gray-200 dark:border-gray-600 rounded overflow-hidden flex flex-col min-h-0">
            <div class="overflow-auto">
              <table class="w-full text-xs">
                <thead class="bg-gray-50 dark:bg-[#2d2d2d] sticky top-0">
                  <tr class="border-b border-gray-200 dark:border-gray-600">
                    <th class="px-3 py-1.5 text-left text-gray-600 dark:text-gray-300">文件路径</th>
                    <th class="px-3 py-1.5 text-left text-gray-600 dark:text-gray-300 w-20">大小</th>
                    <th class="px-3 py-1.5 text-left text-gray-600 dark:text-gray-300 w-24">原编码</th>
                    <th class="px-3 py-1.5 text-left text-gray-600 dark:text-gray-300 w-24">目标编码</th>
                    <th class="px-3 py-1.5 text-left text-gray-600 dark:text-gray-300 w-32">转换结果</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(f, idx) in files" :key="idx" class="border-b border-gray-100 dark:border-gray-700 hover:bg-blue-50 dark:hover:bg-[#3c3c3c]">
                    <td class="px-3 py-1 truncate" :title="f.path">{{ f.path }}</td>
                    <td class="px-3 py-1">{{ f.size }}</td>
                    <td class="px-3 py-1">{{ f.fileCode }}</td>
                    <td class="px-3 py-1">{{ f.convertCode }}</td>
                    <td class="px-3 py-1" :class="f.result === '成功' ? 'text-green-600' : f.result.startsWith('失败') ? 'text-red-600' : 'text-gray-500'">{{ f.result }}</td>
                  </tr>
                  <tr v-if="files.length === 0">
                    <td colspan="5" class="px-3 py-8 text-center text-gray-400">点击"选择目录"开始扫描文件</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          <!-- 底部 -->
          <div class="flex gap-3" style="height: 200px;">
            <!-- 左：选项 + 按钮 -->
            <div class="w-2/3 flex flex-col gap-2">
              <fieldset class="border border-gray-300 dark:border-gray-600 rounded p-2">
                <legend class="text-xs text-gray-600 dark:text-gray-300 px-1">转换选项</legend>
                <div class="flex items-center gap-2 mb-2">
                  <label class="text-xs text-gray-600 dark:text-gray-300 w-20">转换到编码:</label>
                  <select v-model="targetCode" class="flex-1 px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200">
                    <option v-for="c in targetCodeOptions" :key="c" :value="c">{{ c }}</option>
                  </select>
                </div>
                <div class="flex items-center gap-2">
                  <label class="text-xs text-gray-600 dark:text-gray-300 w-20">文件扩展名:</label>
                  <select v-model="extFilter" class="flex-1 px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200">
                    <option v-for="o in extOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
                  </select>
                </div>
              </fieldset>
              <div class="flex justify-center gap-2">
                <button class="ndd-btn" :disabled="loading" @click="selectDir">{{ loading ? '扫描中...' : '选择目录' }}</button>
                <button class="ndd-btn-primary" :disabled="loading || files.length === 0" @click="startConvert">开始</button>
                <button class="ndd-btn" @click="emit('close')">关闭</button>
              </div>
            </div>
            <!-- 右：日志 -->
            <div class="flex-1 flex flex-col">
              <label class="text-xs text-gray-600 dark:text-gray-300 mb-1">日志:</label>
              <textarea v-model="logText" readonly class="flex-1 px-2 py-1 text-xs font-mono border border-gray-300 dark:border-gray-500 rounded bg-gray-50 dark:bg-[#1e1e1e] dark:text-gray-200 resize-none"></textarea>
            </div>
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
.ndd-btn:disabled { opacity: 0.5; cursor: not-allowed; }
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
.ndd-btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
html.dark .ndd-btn {
  background: #3c3c3c;
  color: #e0e0e0;
  border-color: #555;
}
html.dark .ndd-btn:hover { background: #4c4c4c; }
</style>
