<script lang="ts" setup>
import { ref, computed, watch } from 'vue'
import { X, Copy, Trash2, ChevronDown } from 'lucide-vue-next'
import { ElMessage } from 'element-plus'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'close'): void }>()

// ============ 正则输入 ============
const pattern = ref('')
const flags = ref({ global: true, ignoreCase: true, multiline: false, dotAll: false })
const testText = ref('')
const replaceText = ref('')
const showReplace = ref(false)

// ============ 常用模板 ============
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

// ============ 正则构建 ============
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

// ============ 匹配结果 ============
const COLOR_PALETTE = ['#e06c75', '#61afef', '#e5c07b', '#98c379', '#c678dd', '#56b6c2', '#d19a66', '#be5046']

interface MatchResult {
  index: number
  text: string
  groups: { index: number; text: string; groupIndex: number }[]
}

const matches = computed<MatchResult[]>(() => {
  if (!regex.value || !testText.value) return []
  const r = regex.value
  const results: MatchResult[] = []
  // 使用非 global 正则获取捕获组
  const nonGlobal = new RegExp(r.source, buildFlags.value.replace('g', ''))
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
    if (idx++ > 10000) break // 安全上限
  }
  return results
})

const matchCount = computed(() => matches.value.length)

// ============ 替换预览 ============
const replacedText = computed(() => {
  if (!regex.value || !testText.value) return testText.value
  try {
    return testText.value.replace(regex.value, replaceText.value)
  } catch {
    return testText.value
  }
})

// ============ 高亮文本 ============
function highlightText(text: string): string {
  if (!regex.value || !text) return escapeHtml(text)
  const r = regex.value
  let result = ''
  let lastIdx = 0
  let match: RegExpExecArray | null
  let groupIdx = 0
  while ((match = r.exec(text)) !== null) {
    result += escapeHtml(text.slice(lastIdx, match.index))
    const color = COLOR_PALETTE[groupIdx % COLOR_PALETTE.length]
    let highlighted = escapeHtml(match[0])
    // 高亮捕获组
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
  navigator.clipboard.writeText(text).then(() => ElMessage.success('已复制'))
}

function clearAll() {
  pattern.value = ''
  testText.value = ''
  replaceText.value = ''
  showReplace.value = false
}
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-50 flex items-start justify-center pt-20 bg-black/10" @click.self="emit('close')">
      <div class="bg-white dark:bg-[#2d2d2d] border border-gray-300 dark:border-gray-600 rounded shadow-2xl" style="width: 780px; height: 600px; display: flex; flex-direction: column;">
        <!-- 标题栏 -->
        <div class="flex items-center justify-between px-3 py-2 border-b border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-[#3c3c3c]">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-200">正则表达式测试工具</span>
          <div class="flex items-center gap-2">
            <button class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-500" @click="showTemplates = !showTemplates" title="常用模板">
              <ChevronDown class="w-4 h-4" />
            </button>
            <button class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-500" @click="clearAll" title="清空">
              <Trash2 class="w-4 h-4" />
            </button>
            <button class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-500" @click="emit('close')">
              <X class="w-4 h-4" />
            </button>
          </div>
        </div>

        <!-- 模板下拉 -->
        <div v-if="showTemplates" class="border-b border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-[#3c3c3c] p-2 grid grid-cols-3 gap-1">
          <button
            v-for="t in templates" :key="t.name"
            class="text-left px-2 py-1 text-xs rounded hover:bg-blue-50 dark:hover:bg-blue-900/30 border border-transparent hover:border-blue-300 dark:hover:border-blue-600"
            :title="t.desc"
            @click="selectTemplate(t)"
          >
            <div class="font-medium text-gray-700 dark:text-gray-200">{{ t.name }}</div>
            <div class="text-gray-400 font-mono truncate">{{ t.pattern }}</div>
          </button>
        </div>

        <div class="flex-1 flex flex-col p-3 gap-3 overflow-hidden">
          <!-- 正则输入 -->
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-500 w-8">/</span>
            <input
              v-model="pattern"
              placeholder="输入正则表达式…"
              class="flex-1 px-2 py-1.5 text-sm font-mono border rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 focus:outline-none focus:border-blue-400"
              :class="regexError ? 'border-red-400' : 'border-gray-300 dark:border-gray-500'"
            />
            <span class="text-xs text-gray-500">/{{ buildFlags }}</span>
            <label class="flex items-center gap-0.5 text-xs text-gray-500 cursor-pointer" title="全局匹配">
              <input type="checkbox" v-model="flags.global" /> g
            </label>
            <label class="flex items-center gap-0.5 text-xs text-gray-500 cursor-pointer" title="忽略大小写">
              <input type="checkbox" v-model="flags.ignoreCase" /> i
            </label>
            <label class="flex items-center gap-0.5 text-xs text-gray-500 cursor-pointer" title="多行模式">
              <input type="checkbox" v-model="flags.multiline" /> m
            </label>
            <label class="flex items-center gap-0.5 text-xs text-gray-500 cursor-pointer" title="点号匹配换行">
              <input type="checkbox" v-model="flags.dotAll" /> s
            </label>
          </div>

          <!-- 错误提示 -->
          <div v-if="regexError" class="text-xs text-red-500 px-1 -mt-1">{{ regexError }}</div>

          <!-- 匹配统计 -->
          <div v-if="regex" class="flex items-center gap-4 text-xs text-gray-500">
            <span>匹配: <strong class="text-green-600">{{ matchCount }}</strong> 处</span>
            <span v-if="matches.length > 0 && matches[0].groups.length > 0" class="text-blue-500">
              捕获组: {{ matches[0].groups.length }} 个
            </span>
          </div>

          <!-- 测试文本 -->
          <div class="flex-1 flex flex-col min-h-0">
            <div class="flex items-center justify-between mb-1">
              <span class="text-xs text-gray-500">测试文本</span>
              <label class="flex items-center gap-1 text-xs text-gray-500 cursor-pointer">
                <input type="checkbox" v-model="showReplace" /> 显示替换预览
              </label>
            </div>
            <div class="flex-1 flex gap-2 min-h-0">
              <textarea
                v-model="testText"
                placeholder="在此输入测试文本…"
                class="flex-1 px-3 py-2 text-sm font-mono border border-gray-300 dark:border-gray-500 rounded resize-none bg-white dark:bg-[#1e1e1e] dark:text-gray-200 focus:outline-none focus:border-blue-400"
              ></textarea>

              <!-- 替换预览 -->
              <div v-if="showReplace" class="flex-1 flex flex-col min-h-0 gap-1">
                <input
                  v-model="replaceText"
                  placeholder="替换为…"
                  class="px-2 py-1 text-sm font-mono border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 focus:outline-none focus:border-blue-400"
                />
                <div
                  class="flex-1 px-3 py-2 text-sm font-mono border border-gray-300 dark:border-gray-500 rounded bg-gray-50 dark:bg-[#252525] text-gray-700 dark:text-gray-300 overflow-auto whitespace-pre-wrap"
                  v-html="highlightText(replacedText) || '&nbsp;'"
                ></div>
              </div>
            </div>
          </div>

          <!-- 匹配结果列表 -->
          <div v-if="matches.length > 0" class="border-t border-gray-200 dark:border-gray-600 pt-2 max-h-40 overflow-auto">
            <div class="text-xs text-gray-500 mb-1">匹配详情</div>
            <div
              v-for="(m, i) in matches.slice(0, 50)" :key="i"
              class="flex items-start gap-2 py-0.5 text-xs"
            >
              <span class="text-gray-400 w-8 text-right flex-shrink-0">{{ i + 1 }}</span>
              <span class="text-gray-500 w-16 flex-shrink-0">位置 {{ m.index }}</span>
              <code
                class="flex-1 px-1 rounded font-mono"
                :style="{ background: COLOR_PALETTE[i % COLOR_PALETTE.length] + '22', color: COLOR_PALETTE[i % COLOR_PALETTE.length] }"
              >{{ m.text || '(空)' }}</code>
              <button class="p-0.5 text-gray-400 hover:text-gray-600" @click="copyToClipboard(m.text)" title="复制">
                <Copy class="w-3 h-3" />
              </button>
            </div>
            <div v-if="matches.length > 50" class="text-xs text-gray-400 text-center py-1">
              … 仅显示前 50 条，共 {{ matches.length }} 条
            </div>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>