<script lang="ts" setup>
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{
  (e: 'close'): void
  (e: 'apply', rules: any): void
}>()

// Compare Options（三选一）
const compareMode = ref<'before' | 'back' | 'all'>('before')
// Match Options
const blankMatch = ref(true)
const equalRatio = ref<50 | 70 | 90>(50)

watch(() => props.visible, (v) => {
  if (v) {
    // 从 localStorage 恢复
    try {
      const saved = localStorage.getItem('file-cmp-rules')
      if (saved) {
        const r = JSON.parse(saved)
        compareMode.value = r.compareMode || 'before'
        blankMatch.value = r.blankMatch !== false
        equalRatio.value = r.equalRatio || 50
      }
    } catch (e) { console.warn(e) }
  }
})

function apply() {
  const rules = {
    compareMode: compareMode.value,
    blankMatch: blankMatch.value,
    equalRatio: equalRatio.value,
  }
  localStorage.setItem('file-cmp-rules', JSON.stringify(rules))
  emit('apply', rules)
  ElMessage.success('对比规则已应用')
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/30" @click.self="emit('close')">
      <div class="bg-white dark:bg-[#2d2d2d] border border-gray-300 dark:border-gray-600 rounded shadow-2xl w-[480px]">
        <div class="px-3 py-1.5 border-b border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-[#3c3c3c] text-sm font-medium text-gray-700 dark:text-gray-200">
          文件对比规则
        </div>
        <div class="p-4 space-y-3">
          <!-- Compare Options -->
          <fieldset class="border border-gray-300 dark:border-gray-600 rounded p-3">
            <legend class="text-xs text-gray-600 dark:text-gray-300 px-1">对比选项</legend>
            <div class="space-y-2">
              <label class="flex items-center gap-2 text-xs text-gray-600 dark:text-gray-300 cursor-pointer">
                <input type="radio" v-model="compareMode" value="before" /> 忽略行首空白字符
              </label>
              <label class="flex items-center gap-2 text-xs text-gray-600 dark:text-gray-300 cursor-pointer">
                <input type="radio" v-model="compareMode" value="back" /> 忽略行尾空白字符（如 Python）
              </label>
              <label class="flex items-center gap-2 text-xs text-gray-600 dark:text-gray-300 cursor-pointer">
                <input type="radio" v-model="compareMode" value="all" /> 忽略所有空白字符
              </label>
            </div>
          </fieldset>
          <!-- Match Options -->
          <fieldset class="border border-gray-300 dark:border-gray-600 rounded p-3">
            <legend class="text-xs text-gray-600 dark:text-gray-300 px-1">匹配选项</legend>
            <div class="space-y-2">
              <label class="flex items-center gap-2 text-xs text-gray-600 dark:text-gray-300 cursor-pointer">
                <input type="checkbox" v-model="blankMatch" /> 空行参与匹配
              </label>
              <div class="flex items-center gap-2">
                <span class="text-xs text-gray-600 dark:text-gray-300">相等行的匹配率:</span>
                <select v-model="equalRatio" class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-500 rounded bg-white dark:bg-[#1e1e1e] dark:text-gray-200">
                  <option :value="50">匹配 >= 50%</option>
                  <option :value="70">匹配 >= 70%</option>
                  <option :value="90">匹配 >= 90%</option>
                </select>
              </div>
            </div>
          </fieldset>
          <!-- 按钮 -->
          <div class="flex justify-center gap-2 pt-1">
            <button class="ndd-btn-primary" @click="apply">应用</button>
            <button class="ndd-btn" @click="emit('close')">取消</button>
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
html.dark .ndd-btn {
  background: #3c3c3c;
  color: #e0e0e0;
  border-color: #555;
}
html.dark .ndd-btn:hover { background: #4c4c4c; }
</style>
