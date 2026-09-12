<script lang="ts" setup>
/**
 * MD5 / SHA 哈希计算 v2.1
 *  - 用 ModalOverlay 替代手撸 Teleport
 *  - 所有 raw color 改为 token
 */
import { ref, watch } from 'vue'
import { ComputeHashWithAlgo, ComputeFileHashWithAlgo, OpenFileDialog } from '../../wailsjs/go/main/App'
import { ElMessage } from 'element-plus'
import ModalOverlay from './ModalOverlay.vue'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const srcText = ref('')
const hashResult = ref('')
const algorithm = ref<'md4' | 'md5' | 'sha1' | 'sha256' | 'sha3_256' | 'keccak_256'>('md5')
const loading = ref(false)

const algorithms = [
  { key: 'md4',        label: 'MD4' },
  { key: 'md5',        label: 'MD5' },
  { key: 'sha1',       label: 'SHA1' },
  { key: 'sha256',     label: 'SHA256' },
  { key: 'sha3_256',   label: 'SHA3-256' },
  { key: 'keccak_256', label: 'Keccak-256' },
] as const

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
    hashResult.value = (result as any).error ? `错误: ${(result as any).error}` : (result as any).hash
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
    if ((result as any).error) {
      hashResult.value = `错误: ${(result as any).error}`
      ElMessage.error((result as any).error)
    } else {
      hashResult.value = (result as any).hash
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
  <ModalOverlay :visible="visible" title="MD5 / SHA 哈希计算" size="sm" @close="emit('close')">
    <div class="hashwin">
      <div class="hashwin-head">
        <span class="hashwin-hint">粘贴文本或选择文件</span>
        <button class="et-btn-sm" @click="selectFile">选择文件…</button>
      </div>
      <textarea
        v-model="srcText"
        placeholder="粘贴文本到这里，或点击右上角选择文件…"
        class="et-textarea hashwin-src"
      />
      <fieldset class="hashwin-algos">
        <legend>算法</legend>
        <div class="hashwin-algo-grid">
          <label v-for="algo in algorithms" :key="algo.key" class="hashwin-algo-item">
            <input type="radio" v-model="algorithm" :value="algo.key" /> {{ algo.label }}
          </label>
        </div>
      </fieldset>
      <label class="hashwin-result-label">哈希结果</label>
      <textarea
        v-model="hashResult"
        readonly
        class="et-textarea hashwin-result"
      />
    </div>
    <template #footer>
      <button class="et-btn" :disabled="!hashResult" @click="copyToClipboard">复制到剪贴板</button>
      <button class="et-btn et-btn-primary" @click="emit('close')">关闭</button>
    </template>
  </ModalOverlay>
</template>

<style scoped>
.hashwin {
  display: flex;
  flex-direction: column;
  gap: var(--et-space-2);
  min-height: 360px;
}
.hashwin-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: var(--et-text-sm);
  color: var(--et-fg-muted);
}
.hashwin-src { flex: 1; min-height: 80px; font-family: var(--editor-font-family, 'Consolas', monospace); }
.hashwin-algos {
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius-sm);
  padding: var(--et-space-2);
}
.hashwin-algos legend {
  padding: 0 var(--et-space-1);
  font-size: var(--et-text-xs);
  color: var(--et-fg-muted);
}
.hashwin-algo-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--et-space-2);
}
.hashwin-algo-item {
  display: inline-flex;
  align-items: center;
  gap: var(--et-space-1);
  font-size: var(--et-text-sm);
  color: var(--et-fg);
  cursor: pointer;
}
.hashwin-algo-item input { accent-color: var(--et-accent); cursor: pointer; }
.hashwin-result-label {
  font-size: var(--et-text-xs);
  color: var(--et-fg-muted);
  display: block;
}
.hashwin-result {
  min-height: 70px;
  font-family: var(--editor-font-family, 'Consolas', monospace);
  font-size: var(--et-text-sm);
}
</style>