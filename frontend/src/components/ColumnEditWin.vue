<script lang="ts" setup>
import { ref, watch, nextTick } from 'vue'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'close'): void }>()

// 顶部"插入文本"组
const insertTextEnabled = ref(true)
const insertText = ref('')

// 🆕 V2.0.0 插入日期/时间
const insertDateEnabled = ref(false)
const dateFormat = ref('short') // short | long | iso | unix

// 底部"插入数字"组
const insertNumEnabled = ref(false)
const initNum = ref('1')
const incNum = ref(1)
const repeNum = ref(1)
const addPrefix = ref(false)
const prefix = ref('')
// 进制
const radix = ref<10 | 16 | 8 | 2>(10)
const capital = ref(true) // 仅 Hex 时启用

// 🆕 V2.0.0 大小写转换
const caseConversion = ref('') // '' | 'upper' | 'lower' | 'pascal' | 'camel' | 'snake' | 'kebab'

const textInputRef = ref<HTMLInputElement | null>(null)
watch(() => props.visible, async (v) => {
  if (v) {
    insertTextEnabled.value = true
    insertNumEnabled.value = false
    insertDateEnabled.value = false
    insertText.value = ''
    initNum.value = '1'
    incNum.value = 1
    repeNum.value = 1
    addPrefix.value = false
    prefix.value = ''
    radix.value = 10
    capital.value = true
    dateFormat.value = 'short'
    caseConversion.value = ''
    await nextTick()
    textInputRef.value?.focus()
  }
})

function getDateString(): string {
  const now = new Date()
  switch (dateFormat.value) {
    case 'short': return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
    case 'long': return `${now.getFullYear()}年${now.getMonth() + 1}月${now.getDate()}日 ${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`
    case 'iso': return now.toISOString()
    case 'unix': return String(Math.floor(now.getTime() / 1000))
    default: return ''
  }
}

function getCaseLabel(c: string): string {
  const map: Record<string, string> = { '': '无', 'upper': '大写', 'lower': '小写', 'pascal': 'PascalCase', 'camel': 'camelCase', 'snake': 'snake_case', 'kebab': 'kebab-case' }
  return map[c] || c
}

function ok() {
  if (insertTextEnabled.value && !insertNumEnabled.value && !insertDateEnabled.value) {
    // 插入文本
    document.dispatchEvent(new CustomEvent('editor-command', {
      detail: { cmd: 'column-insert-text', args: [insertText.value] }
    }))
  } else if (insertDateEnabled.value && !insertTextEnabled.value && !insertNumEnabled.value) {
    // 🆕 插入日期
    const dateStr = getDateString()
    document.dispatchEvent(new CustomEvent('editor-command', {
      detail: { cmd: 'column-insert-text', args: [dateStr] }
    }))
  } else if (insertNumEnabled.value && !insertTextEnabled.value && !insertDateEnabled.value) {
    // 插入数字
    const init = parseInt(initNum.value) || 0
    document.dispatchEvent(new CustomEvent('editor-command', {
      detail: {
        cmd: 'column-insert-num',
        args: [{
          init: init,
          inc: incNum.value,
          repeat: repeNum.value,
          prefix: addPrefix.value ? prefix.value : '',
          radix: radix.value,
          capital: capital.value,
        }]
      }
    }))
  }

  // 🆕 大小写转换
  if (caseConversion.value) {
    document.dispatchEvent(new CustomEvent('editor-command', {
      detail: { cmd: `case-${caseConversion.value}`, args: [] }
    }))
  }

  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-50 flex items-start justify-center pt-32 bg-black/10" @click.self="emit('close')">
      <div class="bg-white dark:bg-[#2d2d2d] border border-gray-300 dark:border-gray-600 rounded shadow-2xl w-[400px]">
        <div class="px-3 py-1.5 border-b border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-[#3c3c3c] text-sm font-medium text-gray-700 dark:text-gray-200">
          列块编辑
        </div>
        <div class="p-3 flex gap-3">
          <!-- 左侧GroupBox -->
          <div class="flex-1 space-y-2">
            <!-- 插入文本 -->
            <fieldset class="border border-gray-300 dark:border-gray-600 rounded p-2">
              <legend class="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-300 px-1">
                <input type="checkbox" v-model="insertTextEnabled" @change="insertTextEnabled && (insertNumEnabled = false, insertDateEnabled = false)" /> 插入文本
              </legend>
              <input
                ref="textInputRef"
                v-model="insertText"
                :disabled="!insertTextEnabled"
                maxlength="1024"
                class="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 disabled:opacity-50 focus:outline-none focus:border-blue-400"
                @keydown.enter="ok"
                @keydown.escape="emit('close')"
              />
            </fieldset>

            <!-- 🆕 插入日期/时间 -->
            <fieldset class="border border-gray-300 dark:border-gray-600 rounded p-2" :class="!insertDateEnabled ? 'opacity-60' : ''">
              <legend class="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-300 px-1">
                <input type="checkbox" v-model="insertDateEnabled" @change="insertDateEnabled && (insertTextEnabled = false, insertNumEnabled = false)" /> 插入日期/时间
              </legend>
              <div class="space-y-1" :class="!insertDateEnabled ? 'pointer-events-none' : ''">
                <select v-model="dateFormat" class="w-full px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200">
                  <option value="short">短日期 (2026-07-17)</option>
                  <option value="long">长日期 (2026年7月17日 HH:mm:ss)</option>
                  <option value="iso">ISO 8601</option>
                  <option value="unix">Unix 时间戳</option>
                </select>
              </div>
            </fieldset>

            <!-- 🆕 大小写转换 -->
            <fieldset class="border border-gray-300 dark:border-gray-600 rounded p-2">
              <legend class="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-300 px-1">
                大小写转换
              </legend>
              <select v-model="caseConversion" class="w-full px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200">
                <option value="">不转换</option>
                <option value="upper">大写 (UPPERCASE)</option>
                <option value="lower">小写 (lowercase)</option>
                <option value="pascal">PascalCase</option>
                <option value="camel">camelCase</option>
                <option value="snake">snake_case</option>
                <option value="kebab">kebab-case</option>
              </select>
            </fieldset>

            <!-- 插入数字 -->
            <fieldset class="border border-gray-300 dark:border-gray-600 rounded p-2" :class="!insertNumEnabled ? 'opacity-60' : ''">
              <legend class="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-300 px-1">
                <input type="checkbox" v-model="insertNumEnabled" @change="insertNumEnabled && (insertTextEnabled = false, insertDateEnabled = false)" /> 插入数字
              </legend>
              <div class="space-y-1.5" :class="!insertNumEnabled ? 'pointer-events-none' : ''">
                <div class="flex items-center gap-2">
                  <label class="text-xs text-gray-600 dark:text-gray-300 w-16">初始值:</label>
                  <input v-model="initNum" maxlength="11" class="flex-1 px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 focus:outline-none" />
                </div>
                <div class="flex items-center gap-2">
                  <label class="text-xs text-gray-600 dark:text-gray-300 w-16">步进:</label>
                  <input type="number" v-model="incNum" min="-100" class="flex-1 px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 focus:outline-none" />
                </div>
                <div class="flex items-center gap-2">
                  <label class="text-xs text-gray-600 dark:text-gray-300 w-16">重复次数:</label>
                  <input type="number" v-model="repeNum" min="1" class="flex-1 px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 focus:outline-none" />
                </div>
                <div class="flex items-center gap-2">
                  <label class="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-300 cursor-pointer">
                    <input type="checkbox" v-model="addPrefix" /> 前缀字符串:
                  </label>
                  <input v-model="prefix" :disabled="!addPrefix" class="flex-1 px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 disabled:opacity-50 focus:outline-none" />
                </div>
                <fieldset class="border border-gray-300 dark:border-gray-600 rounded p-1.5">
                  <legend class="text-xs text-gray-600 dark:text-gray-300 px-1">格式</legend>
                  <div class="grid grid-cols-2 gap-1.5">
                    <label class="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-300 cursor-pointer">
                      <input type="radio" v-model="radix" :value="10" /> 十进制
                    </label>
                    <label class="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-300 cursor-pointer">
                      <input type="radio" v-model="radix" :value="16" /> 十六进制
                    </label>
                    <label class="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-300 cursor-pointer">
                      <input type="radio" v-model="radix" :value="8" /> 八进制
                    </label>
                    <label class="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-300 cursor-pointer">
                      <input type="radio" v-model="radix" :value="2" /> 二进制
                    </label>
                  </div>
                  <label v-if="radix === 16" class="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-300 cursor-pointer mt-1">
                    <input type="checkbox" v-model="capital" /> 大写
                  </label>
                </fieldset>
              </div>
            </fieldset>
          </div>
          <!-- 右侧按钮 -->
          <div class="flex flex-col gap-2 justify-end">
            <button class="ndd-btn-primary" @click="ok">确定</button>
            <button class="ndd-btn" @click="emit('close')">关闭</button>
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
html.dark .ndd-btn { background: #3c3c3c; color: #e0e0e0; border-color: #555; }
html.dark .ndd-btn:hover { background: #4c4c4c; }
</style>
