<script lang="ts" setup>
/**
 * 正则表达式测试器 v2.1 — ModalOverlay 统一
 *
 * 注意：保留 closeOnEsc=true（默认）。此 dialog 打开时用户常常在 textarea
 * 中用 Esc 关闭 IME，但 Esc 同时会触发 ModalOverlay 关闭。这是合理设计，
 * 想保留 IME 行为的用户可以关闭该 dialog（点 ✕）。
 */
import { ref, computed } from 'vue'
import { Copy, Trash2, ChevronDown } from 'lucide-vue-next'
import { ElMessage } from 'element-plus'
import ModalOverlay from './ModalOverlay.vue'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const pattern = ref('')
const flags = ref({ global: true, ignoreCase: true, multiline: false, dotAll: false })
const testText = ref('')
const replaceText = ref('')
const showReplace = ref(false)

const templates = [
  { name: '电子邮箱', pattern: '[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}', desc: '匹配邮箱地址' },
  { name: '手机号(中国)', pattern: '1[3-9]\\d{9}', desc: '匹配中国大陆手机号' },
  { name: 'IP 地址', pattern: '\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}', desc: '匹配 IPv4 地址' },
  { name: 'URL', pattern: 'https?://[\\w\\-]+(\\.[\\w\\-]+)+([\\w\\-.,@?^=%&:/~+#]*[\\w\\-@?^=%&/~+#])?', desc: '匹配 HTTP/HTTPS URL' },
  { name: '日期 (YYYY-MM-DD)', pattern: '\\d{4}-\\d{2}-\\d{2}', desc: '匹配 ISO 日期格式' },
  { name: '身份证号', pattern: '[1-9]\\d{5}(?:19|20)\\d{2}(?:0[1-9]|1[0-2])(?:0[1-9]|[12]\\d|3[01])\\d{3}[\\dXx]', desc: '匹配中国大陆身份证号' },
  { name: '十六进制颜色', pattern: '#[0-9a-fA-F]{3,8}', desc: '匹配 HEX 颜色值' },
  { name: 'Markdown 标题', pattern: '^#{1,6}\\s+.+$', desc: '匹配 Markdown 标题行' },
  { name: '中文字符', pattern: '[\\u4e00-\\u9fff]+', desc: '匹配中文字符' },
  { name: '空白行', pattern: '^\\s*$', desc: '匹配空白行' },
  { name: 'HTML 标签', pattern: '<[^>]+>', desc: '匹配 HTML 标签' },
  { name: '数字', pattern: '-?\\d+(\\.\\d+)?', desc: '匹配整数或小数' },
]

const showTemplates = ref(false)

const buildFlags = computed(() => {
  let f = ''
  if (flags.value.global) f += 'g'
  if (flags.value.ignoreCase) f += 'i'
  if (flags.value.multiline) f += 'm'
  if (flags.value.dotAll) f += 's'
  return f
})

const regex = computed(() => {
  if (!pattern.value) return null
  try {
    return new RegExp(pattern.value, buildFlags.value)
  } catch {
    return null
  }
})

const regexError = computed(() => {
  if (!pattern.value) return ''
  try {
    new RegExp(pattern.value, buildFlags.value)
    return ''
  } catch (e: any) {
    return e.message
  }
})

const COLOR_PALETTE = ['#e06c75', '#61afef', '#e5c07b', '#98c379', '#c678dd', '#56b6c2', '#d19a66', '#be5046']

interface MatchResult {
  index: number
  text: string
  groups: { index: number; text: string; groupIndex: number }[]
}

const matches = computed<MatchResult[]>(() => {
  if (!regex.value || !testText.value) return []
  const r = new RegExp(regex.value.source, regex.value.flags)
  const results: MatchResult[] = []
  let match: RegExpExecArray | null
  let idx = 0
  while ((match = r.exec(testText.value)) !== null) {
    const groups: { index: number; text: string; groupIndex: number }[] = []
    for (let i = 1; i < match.length; i++) {
      if (match[i] !== undefined) {
        groups.push({ index: match.index + (match[0].indexOf(match[i])), text: match[i], groupIndex: i })
      }
    }
    results.push({ index: match.index, text: match[0], groups })
    if (match.index === r.lastIndex) r.lastIndex++
    if (idx++ > 10000) break
  }
  return results
})

const matchCount = computed(() => matches.value.length)

const replacedText = computed(() => {
  if (!regex.value || !testText.value) return testText.value
  try {
    return testText.value.replace(regex.value, replaceText.value)
  } catch {
    return testText.value
  }
})

function highlightText(text: string): string {
  if (!regex.value || !text) return escapeHtml(text)
  const r = new RegExp(regex.value.source, regex.value.flags)
  let result = ''
  let lastIdx = 0
  let match: RegExpExecArray | null
  let groupIdx = 0
  while ((match = r.exec(text)) !== null) {
    result += escapeHtml(text.slice(lastIdx, match.index))
    const color = COLOR_PALETTE[groupIdx % COLOR_PALETTE.length]
    let highlighted = escapeHtml(match[0])
    for (let i = 1; i < match.length; i++) {
      if (match[i] !== undefined) {
        const gColor = COLOR_PALETTE[(i - 1) % COLOR_PALETTE.length]
        highlighted = highlighted.replace(escapeHtml(match[i]), `<mark style="background:${gColor}44;border:1px solid ${gColor}">${escapeHtml(match[i])}</mark>`)
      }
    }
    result += `<mark style="background:${color}44;border:1px solid ${color}">${highlighted}</mark>`
    lastIdx = match.index + match[0].length
    groupIdx++
    if (match.index === r.lastIndex) r.lastIndex++
    if (groupIdx > 10000) break
  }
  result += escapeHtml(text.slice(lastIdx))
  return result
}

function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

function selectTemplate(t: { name: string; pattern: string }) {
  pattern.value = t.pattern
  showTemplates.value = false
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
    .then(() => ElMessage.success('已复制'))
    .catch(() => ElMessage.error('复制失败'))
}

function clearAll() {
  pattern.value = ''
  testText.value = ''
  replaceText.value = ''
  showReplace.value = false
}
</script>

<template>
  <ModalOverlay :visible="visible" title="正则表达式测试工具" size="lg" @close="emit('close')">
    <template #header-actions>
      <button class="et-icon-btn-sm" @click="showTemplates = !showTemplates" title="常用模板">
        <ChevronDown :size="14" :stroke-width="1.6" />
      </button>
      <button class="et-icon-btn-sm" @click="clearAll" title="清空">
        <Trash2 :size="14" :stroke-width="1.6" />
      </button>
    </template>

    <!-- 模板 -->
    <div v-if="showTemplates" class="rt-templates">
      <button
        v-for="t in templates" :key="t.name"
        class="rt-template"
        :title="t.desc"
        @click="selectTemplate(t)"
      >
        <div class="rt-template-name">{{ t.name }}</div>
        <div class="rt-template-pattern">{{ t.pattern }}</div>
      </button>
    </div>

    <div class="rt-body">
      <!-- 正则输入 -->
      <div class="rt-pattern-row">
        <span class="rt-slash">/</span>
        <input
          v-model="pattern"
          placeholder="输入正则表达式…"
          class="et-input rt-pattern-input"
          :class="regexError ? 'rt-input-error' : ''"
        />
        <span class="rt-slash">/{{ buildFlags }}</span>
        <label class="rt-flag"><input type="checkbox" v-model="flags.global" /> g</label>
        <label class="rt-flag"><input type="checkbox" v-model="flags.ignoreCase" /> i</label>
        <label class="rt-flag"><input type="checkbox" v-model="flags.multiline" /> m</label>
        <label class="rt-flag"><input type="checkbox" v-model="flags.dotAll" /> s</label>
      </div>

      <div v-if="regexError" class="rt-error">{{ regexError }}</div>

      <!-- 匹配统计 -->
      <div v-if="regex" class="rt-stats">
        <span>匹配: <strong class="rt-stat-ok">{{ matchCount }}</strong> 处</span>
        <span v-if="matches.length > 0 && matches[0].groups.length > 0" class="rt-stat-group">
          捕获组: {{ matches[0].groups.length }} 个
        </span>
      </div>

      <!-- 测试文本 -->
      <div class="rt-test">
        <div class="rt-test-head">
          <span class="rt-test-label">测试文本</span>
          <label class="rt-flag">
            <input type="checkbox" v-model="showReplace" /> 显示替换预览
          </label>
        </div>
        <div class="rt-test-body">
          <textarea
            v-model="testText"
            placeholder="在此输入测试文本…"
            class="et-textarea rt-test-input"
          />
          <div v-if="showReplace" class="rt-replace">
            <input v-model="replaceText" placeholder="替换为…" class="et-input rt-replace-input" />
            <div
              class="rt-replace-preview"
              v-html="highlightText(replacedText) || '&nbsp;'"
            />
          </div>
        </div>
      </div>

      <!-- 匹配结果列表 -->
      <div v-if="matches.length > 0" class="rt-matches">
        <div class="rt-matches-head">匹配详情</div>
        <div
          v-for="(m, i) in matches.slice(0, 50)" :key="i"
          class="rt-match-row"
        >
          <span class="rt-match-num">{{ i + 1 }}</span>
          <span class="rt-match-pos">位置 {{ m.index }}</span>
          <code
            class="rt-match-text"
            :style="{ background: COLOR_PALETTE[i % COLOR_PALETTE.length] + '22', color: COLOR_PALETTE[i % COLOR_PALETTE.length] }"
          >{{ m.text || '(空)' }}</code>
          <button class="et-icon-btn-sm rt-match-copy" @click="copyToClipboard(m.text)" title="复制">
            <Copy :size="12" :stroke-width="1.6" />
          </button>
        </div>
        <div v-if="matches.length > 50" class="rt-matches-more">… 仅显示前 50 条，共 {{ matches.length }} 条</div>
      </div>
    </div>
  </ModalOverlay>
</template>

<style scoped>
.rt-templates {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--et-space-1);
  padding: var(--et-space-2);
  background: var(--et-bg-sunken);
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius-sm);
  margin-bottom: var(--et-space-3);
}
.rt-template {
  display: block;
  text-align: left;
  padding: var(--et-space-1) var(--et-space-2);
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--et-radius-sm);
  cursor: pointer;
  transition: background-color 80ms ease, border-color 80ms ease;
}
.rt-template:hover {
  background: var(--et-accent-soft);
  border-color: var(--et-accent);
}
.rt-template-name {
  font-size: var(--et-text-sm);
  font-weight: var(--et-fw-medium);
  color: var(--et-fg);
}
.rt-template-pattern {
  font-family: var(--editor-font-family, 'Consolas', monospace);
  font-size: var(--et-text-xs);
  color: var(--et-fg-subtle);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rt-body {
  display: flex;
  flex-direction: column;
  gap: var(--et-space-2);
}
.rt-pattern-row {
  display: flex;
  align-items: center;
  gap: var(--et-space-2);
}
.rt-slash {
  font-size: var(--et-text-md);
  color: var(--et-fg-muted);
}
.rt-pattern-input {
  flex: 1;
  font-family: var(--editor-font-family, 'Consolas', monospace);
}
.rt-input-error { border-color: var(--et-danger); }
.rt-flag {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  font-size: var(--et-text-xs);
  color: var(--et-fg-muted);
  cursor: pointer;
}
.rt-flag input { accent-color: var(--et-accent); cursor: pointer; }
.rt-error {
  font-size: var(--et-text-xs);
  color: var(--et-danger);
}
.rt-stats {
  display: flex;
  gap: var(--et-space-3);
  font-size: var(--et-text-xs);
  color: var(--et-fg-muted);
}
.rt-stat-ok { color: var(--et-success); }
.rt-stat-group { color: var(--et-accent); }

.rt-test { display: flex; flex-direction: column; min-height: 0; }
.rt-test-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--et-space-1);
}
.rt-test-label {
  font-size: var(--et-text-xs);
  color: var(--et-fg-muted);
}
.rt-test-body {
  display: flex;
  gap: var(--et-space-2);
  flex: 1;
  min-height: 120px;
}
.rt-test-input {
  flex: 1;
  font-family: var(--editor-font-family, 'Consolas', monospace);
  resize: none;
  min-height: 0;
}
.rt-replace {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--et-space-1);
  min-width: 0;
}
.rt-replace-input {
  font-family: var(--editor-font-family, 'Consolas', monospace);
}
.rt-replace-preview {
  flex: 1;
  padding: var(--et-space-2) var(--et-space-3);
  font-family: var(--editor-font-family, 'Consolas', monospace);
  font-size: var(--et-text-sm);
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius-sm);
  background: var(--et-bg-sunken);
  color: var(--et-fg);
  overflow: auto;
  white-space: pre-wrap;
  min-height: 0;
}

.rt-matches {
  border-top: 1px solid var(--et-border);
  padding-top: var(--et-space-2);
  max-height: 160px;
  overflow: auto;
}
.rt-matches-head {
  font-size: var(--et-text-xs);
  color: var(--et-fg-muted);
  margin-bottom: var(--et-space-1);
}
.rt-match-row {
  display: flex;
  align-items: center;
  gap: var(--et-space-2);
  padding: 2px 0;
  font-size: var(--et-text-xs);
}
.rt-match-num {
  color: var(--et-fg-subtle);
  width: 32px;
  text-align: right;
  flex-shrink: 0;
}
.rt-match-pos {
  color: var(--et-fg-muted);
  width: 64px;
  flex-shrink: 0;
}
.rt-match-text {
  flex: 1;
  padding: 0 4px;
  border-radius: var(--et-radius-sm);
  font-family: var(--editor-font-family, 'Consolas', monospace);
  font-size: var(--et-text-sm);
  word-break: break-all;
}
.rt-match-copy { width: 22px; height: 22px; }
.rt-matches-more {
  font-size: var(--et-text-xs);
  color: var(--et-fg-subtle);
  text-align: center;
  padding: var(--et-space-1) 0;
}
</style>