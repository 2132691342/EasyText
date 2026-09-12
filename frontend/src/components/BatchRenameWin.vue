<script lang="ts" setup>
/**
 * 批量重命名 — ModalOverlay 统一
 */
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { RefreshCw, Play, Search, Archive } from 'lucide-vue-next'
import { ListDirectory, OpenDirectoryDialog } from '../../wailsjs/go/main/App'
import ModalOverlay from './ModalOverlay.vue'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const directory = ref('')
const pattern = ref('$N$E')
const startIndex = ref(1)
const step = ref(1)
const files = ref<any[]>([])
const previewItems = ref<{ oldPath: string; newName: string; isDir: boolean }[]>([])
const loading = ref(false)

const commonPatterns = [
  { label: '$N$E (保持原名)', value: '$N$E' },
  { label: 'prefix_$I$E (前缀+序号)', value: 'prefix_$I$E' },
  { label: 'file_$I$E (序号命名)', value: 'file_$I$E' },
  { label: '$N_backup$E (添加后缀)', value: '$N_backup$E' },
  { label: '$N_$I$E (原名+序号)', value: '$N_$I$E' },
]

async function selectDirectory() {
  try {
    const path = await OpenDirectoryDialog()
    if (path) {
      directory.value = path
      await loadFiles()
    }
  } catch {
    ElMessage.error('选择目录失败')
  }
}

async function loadFiles() {
  if (!directory.value) return
  loading.value = true
  try {
    const fileList: any[] = await ListDirectory(directory.value)
    files.value = fileList.filter((f: any) => !f.isDir)
    generatePreview()
  } catch {
    ElMessage.error('加载文件列表失败')
  } finally {
    loading.value = false
  }
}

function generatePreview() {
  previewItems.value = files.value.map((f: any, idx: number) => {
    const ext = f.name.includes('.') ? '.' + f.name.split('.').pop() : ''
    const baseName = ext ? f.name.slice(0, f.name.lastIndexOf(ext)) : f.name
    const index = startIndex.value + idx * step.value
    const newName = pattern.value
      .replace(/\$N/g, baseName)
      .replace(/\$E/g, ext)
      .replace(/\$I/g, String(index))
    return { oldPath: f.name, newName, isDir: false }
  })
}

async function executeRename() {
  if (previewItems.value.length === 0) {
    ElMessage.warning('没有文件可以重命名')
    return
  }
  let successCount = 0
  let failCount = 0
  const failed: string[] = []
  for (const item of previewItems.value) {
    if (item.oldPath === item.newName) { successCount++; continue }
    try {
      const oldPath = directory.value + '\\' + item.oldPath
      const newPath = directory.value + '\\' + item.newName
      const { RenameFile } = await import('../../wailsjs/go/main/App')
      await RenameFile(oldPath, newPath)
      successCount++
    } catch (e: any) {
      failCount++
      failed.push(`${item.oldPath} → ${item.newName}：${e?.message || '未知错误'}`)
    }
  }
  if (failCount > 0) {
    ElMessage.warning(`重命名完成：成功 ${successCount} 个，失败 ${failCount} 个\n` + failed.slice(0, 5).join('\n'))
  } else {
    ElMessage.success(`重命名完成: 成功 ${successCount} 个`)
  }
  if (successCount > 0) await loadFiles()
}

watch([pattern, startIndex, step], () => {
  if (files.value.length > 0) generatePreview()
})
</script>

<template>
  <ModalOverlay :visible="visible" title="批量重命名" size="md" @close="emit('close')">
    <!-- 目录选择 -->
    <div class="br-section">
      <div class="br-row">
        <span class="br-label">目录</span>
        <input
          v-model="directory"
          class="et-input"
          placeholder="选择目录..."
          readonly
          @click="selectDirectory"
        />
        <button class="et-btn-sm" @click="selectDirectory">
          <Search :size="12" :stroke-width="1.6" /> 浏览
        </button>
        <button class="et-btn-sm" :disabled="!directory" @click="loadFiles">
          <RefreshCw :size="12" :stroke-width="1.6" /> 刷新
        </button>
      </div>
    </div>

    <!-- 模板 -->
    <div class="br-section br-pattern">
      <div class="br-row">
        <span class="br-label">模板</span>
        <select v-model="pattern" class="et-select">
          <option v-for="p in commonPatterns" :key="p.value" :value="p.value">{{ p.label }}</option>
        </select>
        <input v-model="pattern" class="et-input br-pattern-input" placeholder="$N$E" />
      </div>
      <div class="br-row">
        <label class="br-mini">
          <span>起始</span>
          <input v-model.number="startIndex" type="number" min="0" class="et-input br-num" />
        </label>
        <label class="br-mini">
          <span>步长</span>
          <input v-model.number="step" type="number" min="1" class="et-input br-num" />
        </label>
        <span class="br-hint">变量：$N=原名 $E=扩展名 $I=序号</span>
      </div>
    </div>

    <!-- 预览 -->
    <div class="br-preview">
      <div v-if="loading" class="et-empty">
        <div class="et-empty-title">加载中…</div>
      </div>
      <div v-else-if="previewItems.length === 0" class="et-empty">
        <Archive :size="32" :stroke-width="1.6" class="et-empty-icon" />
        <div class="et-empty-title">请先选择一个目录</div>
      </div>
      <table v-else class="br-table">
        <thead>
          <tr>
            <th class="br-th br-th-num">#</th>
            <th class="br-th">原文件名</th>
            <th class="br-th br-th-arrow">→</th>
            <th class="br-th">新文件名</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(item, idx) in previewItems" :key="idx" class="br-tr">
            <td class="br-td br-td-num">{{ idx + 1 }}</td>
            <td class="br-td br-mono">{{ item.oldPath }}</td>
            <td class="br-td br-td-arrow">→</td>
            <td class="br-td br-mono" :class="item.oldPath !== item.newName ? 'br-td-new' : 'br-td-same'">
              {{ item.newName }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <template #footer>
      <span class="br-count">{{ previewItems.length }} 个文件</span>
      <button class="et-btn" @click="emit('close')">关闭</button>
      <button class="et-btn et-btn-primary" :disabled="previewItems.length === 0" @click="executeRename">
        <Play :size="12" :stroke-width="1.6" /> 执行重命名
      </button>
    </template>
  </ModalOverlay>
</template>

<style scoped>
.br-section {
  padding-bottom: var(--et-space-3);
  border-bottom: 1px solid var(--et-border);
  margin-bottom: var(--et-space-3);
}
.br-section:last-of-type { border-bottom: 0; }
.br-row {
  display: flex;
  align-items: center;
  gap: var(--et-space-2);
  margin-bottom: var(--et-space-2);
}
.br-row:last-child { margin-bottom: 0; }
.br-label {
  width: 48px;
  flex-shrink: 0;
  font-size: var(--et-text-xs);
  color: var(--et-fg-muted);
}
.br-pattern-input {
  font-family: var(--editor-font-family, 'Consolas', monospace);
  width: 192px;
  flex: 0 0 auto;
}
.br-mini {
  display: inline-flex;
  align-items: center;
  gap: var(--et-space-1);
  font-size: var(--et-text-sm);
  color: var(--et-fg-muted);
}
.br-num { width: 64px; }
.br-hint {
  font-size: var(--et-text-xs);
  color: var(--et-fg-subtle);
}

.br-preview {
  max-height: 320px;
  overflow: auto;
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius-sm);
}
.br-table { width: 100%; border-collapse: collapse; font-size: var(--et-text-sm); }
.br-table thead { background: var(--et-bg-sunken); position: sticky; top: 0; }
.br-th {
  padding: 4px var(--et-space-3);
  text-align: left;
  font-weight: var(--et-fw-medium);
  color: var(--et-fg-muted);
  border-bottom: 1px solid var(--et-border);
  font-size: var(--et-text-xs);
}
.br-th-num { width: 32px; }
.br-th-arrow { width: 32px; color: var(--et-fg-subtle); }
.br-tr { border-bottom: 1px solid var(--et-border); transition: background-color 80ms ease; }
.br-tr:hover { background: var(--et-accent-soft); }
.br-td {
  padding: 4px var(--et-space-3);
  color: var(--et-fg);
}
.br-td-num { color: var(--et-fg-subtle); }
.br-td-arrow { color: var(--et-fg-subtle); text-align: center; }
.br-td-new { color: var(--et-accent); font-weight: var(--et-fw-medium); }
.br-td-same { color: var(--et-fg-muted); }
.br-mono { font-family: var(--editor-font-family, 'Consolas', monospace); }

.br-count {
  font-size: var(--et-text-xs);
  color: var(--et-fg-muted);
  margin-right: auto;
}
</style>