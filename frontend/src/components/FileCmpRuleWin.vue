<script lang="ts" setup>
/**
 * 文件对比规则 v2.1 — ModalOverlay 统一
 */
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import ModalOverlay from './ModalOverlay.vue'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{
  (e: 'close'): void
  (e: 'apply', rules: any): void
}>()

const compareMode = ref<'before' | 'back' | 'all'>('before')
const blankMatch = ref(true)
const equalRatio = ref<50 | 70 | 90>(50)

watch(() => props.visible, (v) => {
  if (v) {
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
  <ModalOverlay :visible="visible" title="文件对比规则" size="sm" @close="emit('close')">
    <div class="fcmp">
      <fieldset class="fcmp-field">
        <legend>对比选项</legend>
        <label><input type="radio" v-model="compareMode" value="before" /> 忽略行首空白字符</label>
        <label><input type="radio" v-model="compareMode" value="back" /> 忽略行尾空白字符（如 Python）</label>
        <label><input type="radio" v-model="compareMode" value="all" /> 忽略所有空白字符</label>
      </fieldset>

      <fieldset class="fcmp-field">
        <legend>匹配选项</legend>
        <label><input type="checkbox" v-model="blankMatch" /> 空行参与匹配</label>
        <div class="fcmp-ratio">
          <span class="fcmp-ratio-label">相等行的匹配率</span>
          <select v-model="equalRatio" class="et-select">
            <option :value="50">匹配 &gt;= 50%</option>
            <option :value="70">匹配 &gt;= 70%</option>
            <option :value="90">匹配 &gt;= 90%</option>
          </select>
          <span class="fcmp-hint">（预留：当前为逐行精确对比）</span>
        </div>
      </fieldset>
    </div>

    <template #footer>
      <button class="et-btn" @click="emit('close')">取消</button>
      <button class="et-btn et-btn-primary" @click="apply">应用</button>
    </template>
  </ModalOverlay>
</template>

<style scoped>
.fcmp {
  display: flex;
  flex-direction: column;
  gap: var(--et-space-3);
}
.fcmp-field {
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius-sm);
  padding: var(--et-space-3);
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: var(--et-space-2);
}
.fcmp-field legend {
  padding: 0 var(--et-space-1);
  font-size: var(--et-text-xs);
  color: var(--et-fg-muted);
  font-weight: var(--et-fw-medium);
}
.fcmp-field label {
  display: inline-flex;
  align-items: center;
  gap: var(--et-space-2);
  font-size: var(--et-text-sm);
  color: var(--et-fg);
  cursor: pointer;
}
.fcmp-field input { accent-color: var(--et-accent); cursor: pointer; }
.fcmp-ratio {
  display: flex;
  align-items: center;
  gap: var(--et-space-2);
}
.fcmp-ratio-label { font-size: var(--et-text-sm); color: var(--et-fg); }
.fcmp-hint { font-size: var(--et-text-xs); color: var(--et-fg-subtle); }
</style>