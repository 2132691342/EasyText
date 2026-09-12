<script lang="ts" setup>
/**
 * 目录对比 — ModalOverlay 统一
 */
import { ref } from 'vue'
import { OpenDirectoryDialog, CompareDirectories } from '../../wailsjs/go/main/App'
import { FolderOpen } from 'lucide-vue-next'
import { ElMessage } from 'element-plus'
import ModalOverlay from './ModalOverlay.vue'

defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const leftDir = ref('')
const rightDir = ref('')
const entries = ref<any[]>([])
const loading = ref(false)

async function pickLeft() {
  try { const p = await OpenDirectoryDialog(); if (p) { leftDir.value = p; runCompare() } } catch (e: any) { ElMessage.error('选择目录失败：' + (e?.message || '')) }
}
async function pickRight() {
  try { const p = await OpenDirectoryDialog(); if (p) { rightDir.value = p; runCompare() } } catch (e: any) { ElMessage.error('选择目录失败：' + (e?.message || '')) }
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
function rowClass(e: any): string {
  if (e.leftOnly) return 'dc-row-l'
  if (e.rightOnly) return 'dc-row-r'
  if (e.different) return 'dc-row-d'
  return ''
}
function statusText(e: any): string {
  if (e.leftOnly) return '仅左侧'
  if (e.rightOnly) return '仅右侧'
  if (e.different) return '不同'
  if (e.identical) return '相同'
  return ''
}
</script>

<template>
  <ModalOverlay :visible="visible" title="目录对比" size="lg" @close="emit('close')">
    <div class="dc-inputs">
      <div class="dc-dir">
        <span class="dc-label">左侧</span>
        <input v-model="leftDir" readonly class="et-input" placeholder="选择左侧目录" />
        <button class="et-btn-sm" @click="pickLeft"><FolderOpen :size="14" :stroke-width="1.6" /></button>
      </div>
      <div class="dc-dir">
        <span class="dc-label">右侧</span>
        <input v-model="rightDir" readonly class="et-input" placeholder="选择右侧目录" />
        <button class="et-btn-sm" @click="pickRight"><FolderOpen :size="14" :stroke-width="1.6" /></button>
      </div>
    </div>

    <div class="dc-body">
      <table v-if="entries.length" class="dc-table">
        <thead>
          <tr>
            <th class="dc-th dc-th-path">相对路径</th>
            <th class="dc-th dc-th-status">状态</th>
            <th class="dc-th dc-th-size">左大小</th>
            <th class="dc-th dc-th-size">右大小</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="e in entries" :key="e.relPath" :class="['dc-tr', rowClass(e)]">
            <td class="dc-td dc-td-path">{{ e.relPath }}</td>
            <td class="dc-td">{{ statusText(e) }}</td>
            <td class="dc-td dc-td-right">{{ e.leftSize || '' }}</td>
            <td class="dc-td dc-td-right">{{ e.rightSize || '' }}</td>
          </tr>
        </tbody>
      </table>
      <div v-else-if="!loading" class="et-empty">
        <div class="et-empty-title">请选择左右两个目录进行对比</div>
      </div>
      <div v-else class="et-empty">
        <div class="et-empty-title">对比中…</div>
      </div>
    </div>
  </ModalOverlay>
</template>

<style scoped>
.dc-inputs {
  display: flex;
  gap: var(--et-space-3);
  padding-bottom: var(--et-space-3);
  border-bottom: 1px solid var(--et-border);
  margin-bottom: var(--et-space-3);
}
.dc-dir {
  flex: 1;
  display: flex;
  align-items: center;
  gap: var(--et-space-2);
  min-width: 0;
}
.dc-dir input { flex: 1; min-width: 0; }
.dc-label {
  width: 36px;
  flex-shrink: 0;
  font-size: var(--et-text-sm);
  color: var(--et-fg-muted);
}

.dc-body {
  min-height: 360px;
  max-height: 60vh;
  overflow: auto;
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius-sm);
}
.dc-table { width: 100%; border-collapse: collapse; font-size: var(--et-text-sm); }
.dc-table thead { background: var(--et-bg-sunken); position: sticky; top: 0; }
.dc-th {
  padding: 4px var(--et-space-3);
  text-align: left;
  font-weight: var(--et-fw-medium);
  color: var(--et-fg-muted);
  border-bottom: 1px solid var(--et-border);
  font-size: var(--et-text-xs);
}
.dc-th-status { width: 80px; }
.dc-th-size   { width: 96px; text-align: right; }
.dc-tr { border-bottom: 1px solid var(--et-border); transition: background-color 80ms ease; }
.dc-tr:hover { background: var(--et-accent-soft); }
.dc-td { padding: 4px var(--et-space-3); color: var(--et-fg); }
.dc-td-path { word-break: break-all; }
.dc-td-right { text-align: right; color: var(--et-fg-muted); }
.dc-row-l { background: var(--et-accent-soft); }
.dc-row-r { background: var(--et-success-bg); }
.dc-row-d { background: var(--et-danger-bg); }
</style>