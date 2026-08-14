<script lang="ts" setup>
import { ref, watch, computed } from 'vue'
import { ComputeHashWithAlgo, ComputeFileHashWithAlgo, OpenFileDialog } from '../../wailsjs/go/main/App'
import { ElMessage } from 'element-plus'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const srcText = ref('')
const hashResult = ref('')
const algorithm = ref<'md4' | 'md5' | 'sha1' | 'sha256' | 'sha3_256' | 'keccak_256'>('md5')
const loading = ref(false)

const algorithms = [
  { key: 'md4', label: 'Md4' },
  { key: 'md5', label: 'Md5' },
  { key: 'sha1', label: 'Sha1' },
  { key: 'sha256', label: 'Sha256' },
  { key: 'sha3_256', label: 'Sha3_256' },
  { key: 'keccak_256', label: 'Keccak_256' },
] as const

// 实时计算
watch([srcText, algorithm], async () => {
  if (!srcText.value) {
    hashResult.value = ''
    return
  }
  await compute()
})

async function compute() {
  if (!srcText.value) {
    hashResult.value = ''
    return
  }
  loading.value = true
  try {
    const result = await ComputeHashWithAlgo(srcText.value, algorithm.value)
    if (result.error) {
      hashResult.value = `错误: ${result.error}`
    } else {
      hashResult.value = result.hash
    }
  } catch (e) {
    hashResult.value = `计算失败: ${e}`
  }
  loading.value = false
}

async function selectFile() {
  try {
    const path = await OpenFileDialog()
    if (!path) return
    loading.value = true
    const result = await ComputeFileHashWithAlgo(path, algorithm.value)
    if (result.error) {
      hashResult.value = `错误: ${result.error}`
      ElMessage.error(result.error)
    } else {
      hashResult.value = result.hash
      srcText.value = `[文件: ${path}]`
    }
  } catch (e) {
    ElMessage.error(`选择文件失败: ${e}`)
  }
  loading.value = false
}

async function copyToClipboard() {
  if (!hashResult.value) return
  try {
    await navigator.clipboard.writeText(hashResult.value)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

watch(() => props.visible, (v) => {
  if (v) {
    srcText.value = ''
    hashResult.value = ''
    algorithm.value = 'md5'
  }
})
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-50 flex items-start justify-center pt-20 bg-black/10" @click.self="emit('close')">
      <div class="bg-white dark:bg-[#2d2d2d] border border-gray-300 dark:border-gray-600 rounded shadow-2xl w-[560px] flex flex-col" style="height: 440px;">
        <div class="px-3 py-1.5 border-b border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-[#3c3c3c] text-sm font-medium text-gray-700 dark:text-gray-200">
          MD5/SHA 哈希计算
        </div>
        <div class="p-3 flex flex-col gap-2 flex-1 overflow-hidden">
          <!-- 顶部 -->
          <div class="flex items-center justify-between">
            <span class="text-xs text-gray-600 dark:text-gray-300">复制文本或选择文件</span>
            <button class="ndd-btn-small" @click="selectFile">选择文件</button>
          </div>
          <!-- 源文本 -->
          <textarea
            v-model="srcText"
            placeholder="粘贴文本到这里，或点击右上角选择文件..."
            class="w-full flex-1 px-2 py-1.5 text-sm font-mono border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200 focus:outline-none focus:border-blue-400 resize-none"
          ></textarea>
          <!-- 算法区 -->
          <fieldset class="border border-gray-300 dark:border-gray-600 rounded p-2">
            <legend class="text-xs text-gray-600 dark:text-gray-300 px-1">算法</legend>
            <div class="grid grid-cols-3 gap-2">
              <label v-for="algo in algorithms" :key="algo.key" class="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-300 cursor-pointer">
                <input type="radio" v-model="algorithm" :value="algo.key" /> {{ algo.label }}
              </label>
            </div>
          </fieldset>
          <!-- 结果 -->
          <div>
            <label class="text-xs text-gray-600 dark:text-gray-300 block mb-1">哈希结果:</label>
            <textarea
              v-model="hashResult"
              readonly
              class="w-full px-2 py-1.5 text-xs font-mono border border-gray-300 dark:border-gray-500 rounded bg-gray-50 dark:bg-[#1e1e1e] dark:text-gray-200 resize-none"
              style="height: 70px;"
            ></textarea>
          </div>
          <!-- 按钮 -->
          <div class="flex justify-end gap-2">
            <button class="ndd-btn" :disabled="!hashResult" @click="copyToClipboard">复制到剪贴板</button>
            <button class="ndd-btn" @click="emit('close')">关闭</button>
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
.ndd-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.ndd-btn-small {
  padding: 2px 10px;
  font-size: 12px;
  border: 1px solid #d1d5db;
  background: #fff;
  color: #374151;
  border-radius: 3px;
  cursor: pointer;
}
.ndd-btn-small:hover { background: #f3f4f6; }
html.dark .ndd-btn, html.dark .ndd-btn-small {
  background: #3c3c3c;
  color: #e0e0e0;
  border-color: #555;
}
html.dark .ndd-btn:hover, html.dark .ndd-btn-small:hover { background: #4c4c4c; }
</style>
