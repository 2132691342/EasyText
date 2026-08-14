<script lang="ts" setup>
import { ref, computed, watch } from 'vue'
import { useEditorStore, useFormatConverterStore } from '@/stores'
import { Convert, JsonPathQuery, JsonToStruct, JsonStructuredDiff } from '../../wailsjs/go/main/App'
import { X, ArrowRight, Copy, FileInput, FileOutput, RefreshCw, Search, FileCode, GitCompare } from 'lucide-vue-next'
import { ElMessage } from 'element-plus'
import type { tools } from '../../wailsjs/go/models'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits(['close'])

const editorStore = useEditorStore()

type Fmt = 'json' | 'yaml' | 'toml' | 'xml'
type ToolTab = 'convert' | 'jsonpath' | 'json-to-struct' | 'json-diff'

const FORMATS: { value: Fmt; label: string }[] = [
  { value: 'json', label: 'JSON' },
  { value: 'yaml', label: 'YAML' },
  { value: 'toml', label: 'TOML' },
  { value: 'xml', label: 'XML' },
]

const activeToolTab = ref<ToolTab>('convert')

// 🆕 V2.0.0 从 useCommands 接收初始 tab（专用 store 替代 window 全局）
watch(() => props.visible, (v) => {
  if (v) {
    const tab = useFormatConverterStore().consume()
    if (tab) activeToolTab.value = tab
  }
})

// ========== 格式转换 ==========
const fromFmt = ref<Fmt>('json')
const toFmt = ref<Fmt>('yaml')
const inputText = ref('')
const outputText = ref('')
const errorMsg = ref('')
const isConverting = ref(false)

const outputLang = computed(() => {
  const map: Record<Fmt, string> = { json: 'json', yaml: 'yaml', toml: 'toml', xml: 'xml' }
  return map[toFmt.value]
})

// load current tab content into input when panel opens
watch(() => props.visible, (v) => {
  if (v && editorStore.activeTab) {
    inputText.value = editorStore.activeTab.content
    const lang = editorStore.activeTab.language
    const langToFmt: Record<string, Fmt> = {
      json: 'json', yaml: 'yaml', toml: 'toml', xml: 'xml',
    }
    if (langToFmt[lang]) {
      fromFmt.value = langToFmt[lang]
      const defaults: Record<Fmt, Fmt> = { json: 'yaml', yaml: 'json', toml: 'json', xml: 'json' }
      toFmt.value = defaults[fromFmt.value]
    }
    outputText.value = ''
    errorMsg.value = ''
    // reset JSON tool states
    jpPath.value = ''
    jpResults.value = []
    jpError.value = ''
    jsLang.value = 'go'
    jsRootName.value = 'Root'
    jsOutput.value = ''
    jsError.value = ''
    jdLeft.value = ''
    jdRight.value = ''
    jdResult.value = null
    jdError.value = ''
  }
})

function swapFormats() {
  const tmp = fromFmt.value
  fromFmt.value = toFmt.value
  toFmt.value = tmp
  if (outputText.value) {
    inputText.value = outputText.value
    outputText.value = ''
  }
}

async function doConvert() {
  if (!inputText.value.trim()) {
    errorMsg.value = '请先输入要转换的内容'
    return
  }
  if (fromFmt.value === toFmt.value) {
    outputText.value = inputText.value
    errorMsg.value = ''
    return
  }
  isConverting.value = true
  errorMsg.value = ''
  try {
    const result = await Convert(inputText.value, fromFmt.value, toFmt.value)
    outputText.value = result
    ElMessage.success('转换成功')
  } catch (e: any) {
    errorMsg.value = e?.message || String(e)
    outputText.value = ''
    ElMessage.error(`转换失败: ${errorMsg.value}`)
  } finally {
    isConverting.value = false
  }
}

function loadFromEditor() {
  if (editorStore.activeTab) {
    inputText.value = editorStore.activeTab.content
    outputText.value = ''
    errorMsg.value = ''
  }
}

function applyToEditor() {
  if (!outputText.value || !editorStore.activeTab) return
  editorStore.updateTabContent(editorStore.activeTab.id, outputText.value)
  ElMessage.success('已应用到编辑器')
  emit('close')
}

async function copyOutput() {
  if (!outputText.value) return
  try {
    await navigator.clipboard.writeText(outputText.value)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

function openInNewTab() {
  if (!outputText.value) return
  const extMap: Record<Fmt, string> = { json: '.json', yaml: '.yaml', toml: '.toml', xml: '.xml' }
  editorStore.createTab('untitled' + extMap[toFmt.value], outputText.value, 'UTF-8', 'LF')
  emit('close')
}

// ========== 🆕 V2.0.0 JSONPath 查询 ==========
const jpPath = ref('')
const jpResults = ref<tools.JSONPathResult[]>([])
const jpError = ref('')
const jpLoading = ref(false)

async function doJsonPathQuery() {
  if (!inputText.value.trim()) {
    jpError.value = '请先在左侧输入 JSON 内容'
    return
  }
  jpLoading.value = true
  jpError.value = ''
  try {
    jpResults.value = await JsonPathQuery(inputText.value, jpPath.value)
    if (jpResults.value.length === 0) {
      ElMessage.info('未找到匹配结果')
    } else {
      ElMessage.success(`找到 ${jpResults.value.length} 个结果`)
    }
  } catch (e: any) {
    jpError.value = e?.message || String(e)
    jpResults.value = []
    ElMessage.error(`查询失败: ${jpError.value}`)
  } finally {
    jpLoading.value = false
  }
}

async function copyJsonPathResults() {
  try {
    await navigator.clipboard.writeText(JSON.stringify(jpResults.value, null, 2))
    ElMessage.success('已复制')
  } catch {
    ElMessage.error('复制失败')
  }
}

// ========== 🆕 V2.0.0 JSON 转结构体 ==========
const jsLang = ref('go')
const jsRootName = ref('Root')
const jsOutput = ref('')
const jsError = ref('')
const jsLoading = ref(false)

const STRUCT_LANGS = [
  { value: 'go', label: 'Go' },
  { value: 'typescript', label: 'TypeScript' },
  { value: 'java', label: 'Java' },
  { value: 'rust', label: 'Rust' },
  { value: 'python', label: 'Python' },
]

async function doJsonToStruct() {
  if (!inputText.value.trim()) {
    jsError.value = '请先在左侧输入 JSON 内容'
    return
  }
  jsLoading.value = true
  jsError.value = ''
  try {
    jsOutput.value = await JsonToStruct(inputText.value, jsLang.value, jsRootName.value)
    ElMessage.success('结构体生成成功')
  } catch (e: any) {
    jsError.value = e?.message || String(e)
    jsOutput.value = ''
    ElMessage.error(`生成失败: ${jsError.value}`)
  } finally {
    jsLoading.value = false
  }
}

async function copyStructOutput() {
  if (!jsOutput.value) return
  try {
    await navigator.clipboard.writeText(jsOutput.value)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

// ========== 🆕 V2.0.0 JSON 结构化 Diff ==========
const jdLeft = ref('')
const jdRight = ref('')
const jdResult = ref<tools.JSONDiffResult | null>(null)
const jdError = ref('')
const jdLoading = ref(false)

async function doJsonDiff() {
  if (!jdLeft.value.trim() || !jdRight.value.trim()) {
    jdError.value = '请输入两侧 JSON 内容'
    return
  }
  jdLoading.value = true
  jdError.value = ''
  try {
    jdResult.value = await JsonStructuredDiff(jdLeft.value, jdRight.value)
    ElMessage.success('对比完成')
  } catch (e: any) {
    jdError.value = e?.message || String(e)
    jdResult.value = null
    ElMessage.error(`对比失败: ${jdError.value}`)
  } finally {
    jdLoading.value = false
  }
}

function loadFromEditorToDiff(side: 'left' | 'right') {
  if (editorStore.activeTab) {
    if (side === 'left') jdLeft.value = editorStore.activeTab.content
    else jdRight.value = editorStore.activeTab.content
  }
}

function getDiffTypeClass(type: string): string {
  switch (type) {
    case 'added': return 'text-green-600 dark:text-green-400'
    case 'removed': return 'text-red-600 dark:text-red-400 line-through'
    case 'modified': return 'text-yellow-600 dark:text-yellow-400'
    default: return 'text-gray-500'
  }
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      @click.self="emit('close')"
    >
      <div class="w-[900px] max-w-[95vw] h-[80vh] flex flex-col bg-white dark:bg-[#1e1e1e] rounded-lg shadow-2xl border border-gray-200 dark:border-gray-700 overflow-hidden">
        <!-- Header -->
        <div class="flex items-center justify-between px-4 py-3 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#252526] flex-shrink-0">
          <div class="flex items-center gap-1">
            <span class="font-semibold text-sm mr-3">格式工具</span>
            <!-- 🆕 V2.0.0 Tab 切换 -->
            <div class="flex gap-0.5 bg-gray-200 dark:bg-[#3c3c3c] rounded p-0.5">
              <button
                class="px-3 py-1 text-xs rounded transition-colors"
                :class="activeToolTab === 'convert' ? 'bg-white dark:bg-[#1e1e1e] text-blue-600 dark:text-blue-400 shadow-sm' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'"
                @click="activeToolTab = 'convert'"
              >格式转换</button>
              <button
                class="px-3 py-1 text-xs rounded transition-colors flex items-center gap-1"
                :class="activeToolTab === 'jsonpath' ? 'bg-white dark:bg-[#1e1e1e] text-blue-600 dark:text-blue-400 shadow-sm' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'"
                @click="activeToolTab = 'jsonpath'"
              ><Search class="w-3 h-3" /> JSONPath</button>
              <button
                class="px-3 py-1 text-xs rounded transition-colors flex items-center gap-1"
                :class="activeToolTab === 'json-to-struct' ? 'bg-white dark:bg-[#1e1e1e] text-blue-600 dark:text-blue-400 shadow-sm' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'"
                @click="activeToolTab = 'json-to-struct'; inputText && doJsonToStruct()"
              ><FileCode class="w-3 h-3" /> 转结构体</button>
              <button
                class="px-3 py-1 text-xs rounded transition-colors flex items-center gap-1"
                :class="activeToolTab === 'json-diff' ? 'bg-white dark:bg-[#1e1e1e] text-blue-600 dark:text-blue-400 shadow-sm' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'"
                @click="activeToolTab = 'json-diff'"
              ><GitCompare class="w-3 h-3" /> JSON Diff</button>
            </div>
          </div>
          <button class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-600" @click="emit('close')">
            <X class="w-4 h-4" />
          </button>
        </div>

        <!-- ========== 格式转换 Tab ========== -->
        <template v-if="activeToolTab === 'convert'">
          <div class="flex items-center gap-2 px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#252526] flex-shrink-0">
            <div class="flex items-center gap-1">
              <span class="text-xs text-gray-500">从</span>
              <div class="flex gap-1">
                <button
                  v-for="fmt in FORMATS"
                  :key="fmt.value"
                  class="px-3 py-1 text-xs rounded font-mono border transition-colors"
                  :class="fromFmt === fmt.value
                    ? 'bg-blue-500 text-white border-blue-500'
                    : 'border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'"
                  @click="fromFmt = fmt.value"
                >{{ fmt.label }}</button>
              </div>
            </div>
            <button
              class="p-1.5 rounded hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-500 transition-transform hover:rotate-180"
              title="交换"
              @click="swapFormats"
            >
              <RefreshCw class="w-4 h-4" />
            </button>
            <div class="flex items-center gap-1">
              <span class="text-xs text-gray-500">转</span>
              <div class="flex gap-1">
                <button
                  v-for="fmt in FORMATS"
                  :key="fmt.value"
                  class="px-3 py-1 text-xs rounded font-mono border transition-colors"
                  :class="toFmt === fmt.value
                    ? 'bg-green-500 text-white border-green-500'
                    : 'border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'"
                  @click="toFmt = fmt.value"
                >{{ fmt.label }}</button>
              </div>
            </div>
            <button
              class="ml-auto flex items-center gap-1.5 px-4 py-1.5 bg-blue-500 hover:bg-blue-600 text-white text-sm rounded font-medium transition-colors disabled:opacity-50"
              :disabled="isConverting"
              @click="doConvert"
            >
              <ArrowRight class="w-4 h-4" />
              {{ isConverting ? '转换中…' : '转换' }}
            </button>
          </div>
          <div v-if="errorMsg" class="px-4 py-2 bg-red-50 dark:bg-red-900/20 border-b border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 text-xs flex-shrink-0">
            {{ errorMsg }}
          </div>
          <div class="flex-1 flex overflow-hidden">
            <div class="flex-1 flex flex-col border-r border-gray-200 dark:border-gray-700 overflow-hidden">
              <div class="flex items-center justify-between px-3 py-1.5 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#252526] flex-shrink-0">
                <span class="text-xs font-mono text-gray-500 uppercase">{{ fromFmt }}</span>
                <button
                  class="flex items-center gap-1 text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 px-1.5 py-0.5 rounded hover:bg-gray-200 dark:hover:bg-gray-700"
                  title="从编辑器加载"
                  @click="loadFromEditor"
                >
                  <FileInput class="w-3.5 h-3.5" />
                  从编辑器加载
                </button>
              </div>
              <textarea
                v-model="inputText"
                class="flex-1 w-full p-3 font-mono text-sm resize-none outline-none bg-white dark:bg-[#1e1e1e] text-gray-800 dark:text-gray-200 leading-relaxed"
                placeholder="在此粘贴或输入内容…"
                spellcheck="false"
              ></textarea>
            </div>
            <div class="flex-1 flex flex-col overflow-hidden">
              <div class="flex items-center justify-between px-3 py-1.5 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#252526] flex-shrink-0">
                <span class="text-xs font-mono text-gray-500 uppercase">{{ toFmt }}</span>
                <div class="flex items-center gap-1">
                  <button v-if="outputText" class="flex items-center gap-1 text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 px-1.5 py-0.5 rounded hover:bg-gray-200 dark:hover:bg-gray-700" @click="copyOutput"><Copy class="w-3.5 h-3.5" /> 复制</button>
                  <button v-if="outputText" class="flex items-center gap-1 text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 px-1.5 py-0.5 rounded hover:bg-gray-200 dark:hover:bg-gray-700" @click="openInNewTab" title="在新标签页打开"><FileOutput class="w-3.5 h-3.5" /> 新标签</button>
                  <button v-if="outputText && editorStore.activeTab" class="flex items-center gap-1 text-xs text-blue-500 hover:text-blue-600 px-1.5 py-0.5 rounded hover:bg-blue-50 dark:hover:bg-blue-900/20" @click="applyToEditor" title="替换当前编辑器内容">应用到编辑器</button>
                </div>
              </div>
              <textarea :value="outputText" readonly class="flex-1 w-full p-3 font-mono text-sm resize-none outline-none bg-gray-50 dark:bg-[#252526] text-gray-800 dark:text-gray-200 leading-relaxed" :placeholder="isConverting ? '转换中…' : '转换结果将显示在这里'" spellcheck="false"></textarea>
            </div>
          </div>
        </template>

        <!-- ========== 🆕 V2.0.0 JSONPath 查询 Tab ========== -->
        <template v-if="activeToolTab === 'jsonpath'">
          <div class="flex items-center gap-2 px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#252526] flex-shrink-0">
            <span class="text-xs text-gray-500">JSONPath 表达式:</span>
            <input
              v-model="jpPath"
              placeholder="如 $.store.book[*].author"
              class="flex-1 px-2 py-1 text-xs font-mono border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 focus:outline-none focus:border-blue-400"
              @keydown.enter="doJsonPathQuery"
            />
            <button
              class="flex items-center gap-1 px-3 py-1 bg-blue-500 hover:bg-blue-600 text-white text-xs rounded disabled:opacity-50"
              :disabled="jpLoading"
              @click="doJsonPathQuery"
            >
              <Search class="w-3 h-3" />
              {{ jpLoading ? '查询中…' : '查询' }}
            </button>
          </div>
          <div v-if="jpError" class="px-4 py-2 bg-red-50 dark:bg-red-900/20 border-b border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 text-xs flex-shrink-0">{{ jpError }}</div>
          <div class="flex-1 flex overflow-hidden">
            <div class="flex-1 flex flex-col border-r border-gray-200 dark:border-gray-700 overflow-hidden">
              <div class="flex items-center justify-between px-3 py-1.5 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#252526] flex-shrink-0">
                <span class="text-xs font-mono text-gray-500">JSON</span>
                <button class="flex items-center gap-1 text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 px-1.5 py-0.5 rounded hover:bg-gray-200 dark:hover:bg-gray-700" @click="loadFromEditor"><FileInput class="w-3.5 h-3.5" /> 从编辑器加载</button>
              </div>
              <textarea v-model="inputText" class="flex-1 w-full p-3 font-mono text-sm resize-none outline-none bg-white dark:bg-[#1e1e1e] text-gray-800 dark:text-gray-200 leading-relaxed" placeholder="输入 JSON 内容…" spellcheck="false"></textarea>
            </div>
            <div class="flex-1 flex flex-col overflow-hidden">
              <div class="flex items-center justify-between px-3 py-1.5 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#252526] flex-shrink-0">
                <span class="text-xs text-gray-500">结果 ({{ jpResults.length }})</span>
                <button v-if="jpResults.length > 0" class="flex items-center gap-1 text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 px-1.5 py-0.5 rounded hover:bg-gray-200 dark:hover:bg-gray-700" @click="copyJsonPathResults"><Copy class="w-3 h-3" /></button>
              </div>
              <div class="flex-1 overflow-auto p-3 font-mono text-sm">
                <div v-if="jpResults.length === 0" class="text-gray-400 text-xs">输入 JSONPath 表达式并点击查询</div>
                <div v-for="(r, i) in jpResults" :key="i" class="py-1 border-b border-gray-100 dark:border-gray-800 last:border-0">
                  <div class="text-[10px] text-gray-400">{{ r.path }}</div>
                  <pre class="text-xs text-gray-700 dark:text-gray-300 whitespace-pre-wrap">{{ JSON.stringify(r.value, null, 2) }}</pre>
                </div>
              </div>
            </div>
          </div>
        </template>

        <!-- ========== 🆕 V2.0.0 JSON 转结构体 Tab ========== -->
        <template v-if="activeToolTab === 'json-to-struct'">
          <div class="flex items-center gap-2 px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#252526] flex-shrink-0">
            <span class="text-xs text-gray-500">语言:</span>
            <select v-model="jsLang" class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200">
              <option v-for="l in STRUCT_LANGS" :key="l.value" :value="l.value">{{ l.label }}</option>
            </select>
            <span class="text-xs text-gray-500 ml-2">根类型名:</span>
            <input v-model="jsRootName" class="w-32 px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 focus:outline-none focus:border-blue-400" />
            <button
              class="ml-auto flex items-center gap-1 px-3 py-1 bg-blue-500 hover:bg-blue-600 text-white text-xs rounded disabled:opacity-50"
              :disabled="jsLoading"
              @click="doJsonToStruct"
            >
              <FileCode class="w-3 h-3" />
              {{ jsLoading ? '生成中…' : '生成' }}
            </button>
          </div>
          <div v-if="jsError" class="px-4 py-2 bg-red-50 dark:bg-red-900/20 border-b border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 text-xs flex-shrink-0">{{ jsError }}</div>
          <div class="flex-1 flex overflow-hidden">
            <div class="flex-1 flex flex-col border-r border-gray-200 dark:border-gray-700 overflow-hidden">
              <div class="flex items-center justify-between px-3 py-1.5 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#252526] flex-shrink-0">
                <span class="text-xs font-mono text-gray-500">JSON</span>
                <button class="flex items-center gap-1 text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 px-1.5 py-0.5 rounded hover:bg-gray-200 dark:hover:bg-gray-700" @click="loadFromEditor"><FileInput class="w-3.5 h-3.5" /> 从编辑器加载</button>
              </div>
              <textarea v-model="inputText" class="flex-1 w-full p-3 font-mono text-sm resize-none outline-none bg-white dark:bg-[#1e1e1e] text-gray-800 dark:text-gray-200 leading-relaxed" placeholder="输入 JSON 内容…" spellcheck="false"></textarea>
            </div>
            <div class="flex-1 flex flex-col overflow-hidden">
              <div class="flex items-center justify-between px-3 py-1.5 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#252526] flex-shrink-0">
                <span class="text-xs font-mono text-gray-500 uppercase">{{ jsLang }}</span>
                <button v-if="jsOutput" class="flex items-center gap-1 text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 px-1.5 py-0.5 rounded hover:bg-gray-200 dark:hover:bg-gray-700" @click="copyStructOutput"><Copy class="w-3.5 h-3.5" /> 复制</button>
              </div>
              <textarea :value="jsOutput" readonly class="flex-1 w-full p-3 font-mono text-sm resize-none outline-none bg-gray-50 dark:bg-[#252526] text-gray-800 dark:text-gray-200 leading-relaxed" :placeholder="jsLoading ? '生成中…' : '生成的结构体将显示在这里'" spellcheck="false"></textarea>
            </div>
          </div>
        </template>

        <!-- ========== 🆕 V2.0.0 JSON 结构化 Diff Tab ========== -->
        <template v-if="activeToolTab === 'json-diff'">
          <div class="flex items-center gap-2 px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#252526] flex-shrink-0">
            <span class="text-xs text-gray-500">按字段级对比两个 JSON 对象</span>
            <button
              class="ml-auto flex items-center gap-1 px-3 py-1 bg-blue-500 hover:bg-blue-600 text-white text-xs rounded disabled:opacity-50"
              :disabled="jdLoading"
              @click="doJsonDiff"
            >
              <GitCompare class="w-3 h-3" />
              {{ jdLoading ? '对比中…' : '对比' }}
            </button>
          </div>
          <div v-if="jdError" class="px-4 py-2 bg-red-50 dark:bg-red-900/20 border-b border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 text-xs flex-shrink-0">{{ jdError }}</div>
          <div class="flex-1 flex overflow-hidden">
            <div class="flex-1 flex flex-col border-r border-gray-200 dark:border-gray-700 overflow-hidden">
              <div class="flex items-center justify-between px-3 py-1.5 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#252526] flex-shrink-0">
                <span class="text-xs font-mono text-gray-500">原始 JSON</span>
                <button class="flex items-center gap-1 text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 px-1.5 py-0.5 rounded hover:bg-gray-200 dark:hover:bg-gray-700" @click="loadFromEditorToDiff('left')"><FileInput class="w-3.5 h-3.5" /> 从编辑器加载</button>
              </div>
              <textarea v-model="jdLeft" class="flex-1 w-full p-3 font-mono text-sm resize-none outline-none bg-white dark:bg-[#1e1e1e] text-gray-800 dark:text-gray-200 leading-relaxed" placeholder="输入原始 JSON…" spellcheck="false"></textarea>
            </div>
            <div class="flex-1 flex flex-col overflow-hidden">
              <div class="flex items-center justify-between px-3 py-1.5 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#252526] flex-shrink-0">
                <span class="text-xs font-mono text-gray-500">修改后 JSON</span>
                <button class="flex items-center gap-1 text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 px-1.5 py-0.5 rounded hover:bg-gray-200 dark:hover:bg-gray-700" @click="loadFromEditorToDiff('right')"><FileInput class="w-3.5 h-3.5" /> 从编辑器加载</button>
              </div>
              <textarea v-model="jdRight" class="flex-1 w-full p-3 font-mono text-sm resize-none outline-none bg-white dark:bg-[#1e1e1e] text-gray-800 dark:text-gray-200 leading-relaxed" placeholder="输入修改后 JSON…" spellcheck="false"></textarea>
            </div>
          </div>
          <!-- Diff 结果摘要和列表 -->
          <div v-if="jdResult" class="border-t border-gray-200 dark:border-gray-700 flex-shrink-0 max-h-[40%] overflow-auto">
            <div class="flex items-center gap-3 px-4 py-2 bg-gray-50 dark:bg-[#252526] border-b border-gray-200 dark:border-gray-700 text-xs">
              <span class="text-green-600 dark:text-green-400">+{{ jdResult.summary.added }} 新增</span>
              <span class="text-red-600 dark:text-red-400">-{{ jdResult.summary.removed }} 删除</span>
              <span class="text-yellow-600 dark:text-yellow-400">~{{ jdResult.summary.modified }} 修改</span>
            </div>
            <div v-for="(entry, i) in jdResult.entries" :key="i" class="flex items-start px-4 py-1 border-b border-gray-100 dark:border-gray-800 last:border-0 text-xs font-mono">
              <span class="w-14 flex-shrink-0 font-semibold" :class="{
                'text-green-600 dark:text-green-400': entry.type === 'added',
                'text-red-600 dark:text-red-400': entry.type === 'removed',
                'text-yellow-600 dark:text-yellow-400': entry.type === 'modified',
                'text-gray-400': entry.type === 'unchanged'
              }">{{ entry.type === 'added' ? '新增' : entry.type === 'removed' ? '删除' : entry.type === 'modified' ? '修改' : '未变' }}</span>
              <span class="text-gray-400 w-56 truncate flex-shrink-0">{{ entry.path }}</span>
              <span v-if="entry.type === 'removed'" class="text-red-500 truncate">「{{ JSON.stringify(entry.oldValue) }}」</span>
              <span v-else-if="entry.type === 'added'" class="text-green-500 truncate">「{{ JSON.stringify(entry.newValue) }}」</span>
              <span v-else-if="entry.type === 'modified'" class="text-yellow-500 truncate">「{{ JSON.stringify(entry.oldValue) }}」→「{{ JSON.stringify(entry.newValue) }}」</span>
            </div>
          </div>
        </template>
      </div>
    </div>
  </Teleport>
</template>
