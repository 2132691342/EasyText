<script lang="ts" setup>
/**
 * 批量编码转换 v2.1 — ModalOverlay 统一
 */
import { ref, watch } from 'vue'
import {
  OpenDirectoryDialog, GetDirectoryTree, ReadFileBytes, SaveFileBytes,
  DetectEncoding, ConvertEncoding,
} from '../../wailsjs/go/main/App'
import { ElMessage } from 'element-plus'
import ModalOverlay from './ModalOverlay.vue'

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

// 值必须与后端 SupportedEncodings 的 Name 完全一致。
// 此前写成 UTF-16-LE / UTF-8-BOM 等后端不识别的名字，选中后转换必然失败。
const targetCodeOptions = [
  { value: 'UTF-8', label: 'UTF-8' },
  { value: 'UTF-16LE', label: 'UTF-16 LE' },
  { value: 'UTF-16BE', label: 'UTF-16 BE' },
  { value: 'GBK', label: '简体中文 (GBK)' },
  { value: 'GB18030', label: '简体中文 (GB18030)' },
  { value: 'Big5', label: '繁体中文 (Big5)' },
]
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

function ecResultClass(r: string): string {
  if (r === '成功') return 'ec-ok'
  if (r.startsWith('失败')) return 'ec-fail'
  return 'ec-pend'
}
</script>

<template>
  <ModalOverlay :visible="visible" title="批量编码转换" size="lg" @close="emit('close')">
    <div class="ec">
      <!-- 文件列表 -->
      <div class="ec-table-wrap">
        <div class="ec-table-scroll">
          <table class="ec-table">
            <thead>
              <tr>
                <th class="ec-th ec-th-path">文件路径</th>
                <th class="ec-th ec-th-size">大小</th>
                <th class="ec-th ec-th-enc">原编码</th>
                <th class="ec-th ec-th-enc">目标编码</th>
                <th class="ec-th ec-th-result">转换结果</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(f, idx) in files" :key="idx" class="ec-tr">
                <td class="ec-td ec-td-path" :title="f.path">{{ f.path }}</td>
                <td class="ec-td">{{ f.size }}</td>
                <td class="ec-td">{{ f.fileCode }}</td>
                <td class="ec-td">{{ f.convertCode }}</td>
                <td class="ec-td ec-td-result" :class="ecResultClass(f.result)">{{ f.result }}</td>
              </tr>
              <tr v-if="files.length === 0">
                <td colspan="5" class="ec-empty">点击「选择目录」开始扫描文件</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 底部：选项 + 按钮 + 日志 -->
      <div class="ec-bottom">
        <div class="ec-bottom-left">
          <fieldset class="ec-field">
            <legend>转换选项</legend>
            <div class="ec-row">
              <span class="ec-label">转换到编码</span>
              <select v-model="targetCode" class="et-select">
                <option v-for="c in targetCodeOptions" :key="c.value" :value="c.value">{{ c.label }}</option>
              </select>
            </div>
            <div class="ec-row">
              <span class="ec-label">文件扩展名</span>
              <select v-model="extFilter" class="et-select">
                <option v-for="o in extOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
              </select>
            </div>
          </fieldset>
          <div class="ec-actions">
            <button class="et-btn-sm" :disabled="loading" @click="selectDir">{{ loading ? '扫描中…' : '选择目录' }}</button>
            <button class="et-btn-sm et-btn-primary" :disabled="loading || files.length === 0" @click="startConvert">开始</button>
          </div>
        </div>
        <div class="ec-bottom-right">
          <label class="ec-label">日志</label>
          <textarea v-model="logText" readonly class="et-textarea ec-log" />
        </div>
      </div>
    </div>

    <template #footer>
      <button class="et-btn" @click="emit('close')">关闭</button>
    </template>
  </ModalOverlay>
</template>

<style scoped>
.ec {
  display: flex;
  flex-direction: column;
  gap: var(--et-space-3);
  min-height: 500px;
}
.ec-table-wrap {
  flex: 1;
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius-sm);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
.ec-table-scroll { overflow: auto; flex: 1; }
.ec-table { width: 100%; border-collapse: collapse; font-size: var(--et-text-sm); }
.ec-table thead { background: var(--et-bg-sunken); position: sticky; top: 0; }
.ec-th {
  padding: 4px var(--et-space-3);
  text-align: left;
  font-weight: var(--et-fw-medium);
  color: var(--et-fg-muted);
  border-bottom: 1px solid var(--et-border);
  font-size: var(--et-text-xs);
}
.ec-th-size   { width: 80px; }
.ec-th-enc    { width: 96px; }
.ec-th-result { width: 128px; }
.ec-tr { border-bottom: 1px solid var(--et-border); transition: background-color 80ms ease; }
.ec-tr:hover { background: var(--et-accent-soft); }
.ec-td {
  padding: 4px var(--et-space-3);
  color: var(--et-fg);
}
.ec-td-path { font-family: var(--editor-font-family, 'Consolas', monospace); }
.ec-td-result.ec-ok    { color: var(--et-success); }
.ec-td-result.ec-fail  { color: var(--et-danger); }
.ec-td-result.ec-pend  { color: var(--et-fg-muted); }
.ec-empty {
  padding: var(--et-space-5);
  text-align: center;
  color: var(--et-fg-subtle);
}

.ec-bottom {
  display: flex;
  gap: var(--et-space-3);
  height: 200px;
}
.ec-bottom-left  { width: 66.6667%; display: flex; flex-direction: column; gap: var(--et-space-2); }
.ec-bottom-right { flex: 1; display: flex; flex-direction: column; }
.ec-field {
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius-sm);
  padding: var(--et-space-2);
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: var(--et-space-1);
}
.ec-field legend {
  padding: 0 var(--et-space-1);
  font-size: var(--et-text-xs);
  color: var(--et-fg-muted);
  font-weight: var(--et-fw-medium);
}
.ec-row { display: flex; align-items: center; gap: var(--et-space-2); }
.ec-label {
  width: 80px;
  flex-shrink: 0;
  font-size: var(--et-text-xs);
  color: var(--et-fg-muted);
}
.ec-actions {
  display: flex;
  justify-content: center;
  gap: var(--et-space-2);
  margin-top: var(--et-space-1);
}
.ec-log {
  flex: 1;
  font-family: var(--editor-font-family, 'Consolas', monospace);
  font-size: var(--et-text-sm);
  margin-top: var(--et-space-1);
  min-height: 0;
}
</style>
