<script lang="ts" setup>
/**
 * 批量查找替换 v2.1 — ModalOverlay 统一
 */
import { ref, watch, computed } from 'vue'
import { useEditorStore } from '@/stores'
import { ElMessage } from 'element-plus'
import ModalOverlay from './ModalOverlay.vue'

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
  <ModalOverlay :visible="visible" title="批量查找替换" size="lg" @close="emit('close')">
    <div class="bfr">
      <!-- 顶部双栏 -->
      <div class="bfr-inputs">
        <div class="bfr-col">
          <label class="bfr-label">输入多个查找关键字，以空白字符分隔：</label>
          <textarea
            v-model="findKeywords"
            class="et-textarea bfr-textarea"
            placeholder="keyword1 keyword2 keyword3..."
          />
        </div>
        <div class="bfr-col">
          <label class="bfr-label">输入多个替换关键字，以空白字符分隔：</label>
          <textarea
            v-model="replaceKeywords"
            class="et-textarea bfr-textarea"
            placeholder="replace1 replace2 replace3..."
          />
        </div>
      </div>

      <!-- 中部表格 -->
      <div class="bfr-table-wrap">
        <div class="bfr-table-scroll">
          <table class="bfr-table">
            <thead>
              <tr>
                <th class="bfr-th bfr-th-num">#</th>
                <th class="bfr-th">关键字 (Find)</th>
                <th class="bfr-th">替换 (Replace)</th>
                <th class="bfr-th bfr-th-mark">标记</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(r, idx) in rows" :key="idx" class="bfr-tr">
                <td class="bfr-td bfr-td-num">{{ idx + 1 }}</td>
                <td class="bfr-td bfr-mono">{{ r.find }}</td>
                <td class="bfr-td bfr-mono">{{ r.replace }}</td>
                <td class="bfr-td bfr-td-center">
                  <span v-if="r.marked" class="bfr-mark-on">★</span>
                  <span v-else class="bfr-mark-off">-</span>
                </td>
              </tr>
              <tr v-if="rows.length === 0">
                <td colspan="4" class="bfr-empty">点击「刷新」按钮根据上方文本框生成规则</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 结果统计 -->
      <div v-if="results.length > 0" class="bfr-results">
        <div class="bfr-results-head">查找结果（共 {{ stats.matched }} 处匹配，涉及 {{ stats.files }} 个文件）：</div>
        <div v-for="(r, idx) in results.slice(0, 50)" :key="idx" class="bfr-result-row">
          <span class="bfr-result-loc">{{ r.file }}:{{ r.line }}</span>
          <span class="bfr-result-kw">[{{ r.keyword }}]</span>
          <span class="bfr-result-content">{{ r.content }}</span>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="bfr-footer">
        <button class="et-btn-sm" @click="fresh">刷新</button>
        <button class="et-btn-sm" @click="swap">交换</button>
        <button class="et-btn-sm et-btn-primary" @click="findAll">查找</button>
        <button class="et-btn-sm et-btn-primary" @click="replaceAll">替换</button>
        <button class="et-btn-sm" @click="mark">标记</button>
        <button class="et-btn-sm" @click="clearMark">清除标记</button>
        <button class="et-btn-sm" @click="importKeywords">导入</button>
        <button class="et-btn-sm" @click="exportKeywords">导出</button>
        <button class="et-btn-sm" style="margin-left: var(--et-space-3)" @click="emit('close')">关闭</button>
      </div>
    </template>
  </ModalOverlay>
</template>

<style scoped>
.bfr {
  display: flex;
  flex-direction: column;
  gap: var(--et-space-3);
  min-height: 540px;
}
.bfr-inputs {
  display: flex;
  gap: var(--et-space-3);
  height: 140px;
}
.bfr-col { flex: 1; display: flex; flex-direction: column; min-width: 0; }
.bfr-label {
  font-size: var(--et-text-xs);
  color: var(--et-fg-muted);
  margin-bottom: var(--et-space-1);
}
.bfr-textarea {
  flex: 1;
  font-family: var(--editor-font-family, 'Consolas', monospace);
  font-size: var(--et-text-sm);
  min-height: 0;
}

.bfr-table-wrap {
  flex: 1;
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius-sm);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
.bfr-table-scroll { overflow: auto; flex: 1; }
.bfr-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--et-text-xs);
}
.bfr-table thead { background: var(--et-bg-sunken); position: sticky; top: 0; }
.bfr-th {
  padding: 6px var(--et-space-3);
  text-align: left;
  font-weight: var(--et-fw-medium);
  color: var(--et-fg-muted);
  border-bottom: 1px solid var(--et-border);
}
.bfr-th-num { width: 40px; }
.bfr-th-mark { width: 64px; }
.bfr-tr {
  border-bottom: 1px solid var(--et-border);
  transition: background-color 80ms ease;
}
.bfr-tr:hover { background: var(--et-accent-soft); }
.bfr-td {
  padding: 4px var(--et-space-3);
  color: var(--et-fg);
}
.bfr-td-num { color: var(--et-fg-subtle); }
.bfr-td-center { text-align: center; }
.bfr-mono {
  font-family: var(--editor-font-family, 'Consolas', monospace);
}
.bfr-mark-on { color: var(--et-warn); }
.bfr-mark-off { color: var(--et-fg-subtle); }
.bfr-empty {
  padding: var(--et-space-5);
  text-align: center;
  color: var(--et-fg-subtle);
}

.bfr-results {
  font-size: var(--et-text-xs);
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius-sm);
  padding: var(--et-space-2);
  max-height: 128px;
  overflow: auto;
}
.bfr-results-head {
  font-weight: var(--et-fw-medium);
  margin-bottom: var(--et-space-1);
  color: var(--et-fg);
}
.bfr-result-row {
  display: flex;
  gap: var(--et-space-2);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.bfr-result-loc { color: var(--et-accent); flex-shrink: 0; }
.bfr-result-kw  { color: var(--et-fg-muted); flex-shrink: 0; }
.bfr-result-content { color: var(--et-fg); overflow: hidden; text-overflow: ellipsis; }

.bfr-footer {
  display: flex;
  gap: var(--et-space-1);
  flex-wrap: wrap;
  align-items: center;
}
</style>
