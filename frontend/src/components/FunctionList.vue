<script lang="ts" setup>
import { computed } from 'vue'
import { extractSymbols } from '@/utils/symbols'
import { FunctionSquare, Box } from 'lucide-vue-next'

const props = defineProps<{ content: string; language: string }>()
const emit = defineEmits<{ (e: 'goto', line: number): void }>()

const symbols = computed(() => extractSymbols(props.content, props.language))
</script>

<template>
  <div class="h-full flex flex-col bg-[var(--et-bg-sunken)] text-[var(--et-fg)]">
    <div class="fl-head">
      <span class="fl-title">函数列表</span>
      <span class="fl-count">{{ symbols.length }}</span>
    </div>
    <div class="flex-1 overflow-auto">
      <div v-if="!symbols.length" class="fl-empty">
        当前文件未识别到函数或类
      </div>
      <div
        v-for="s in symbols"
        :key="s.line + s.name"
        class="fl-item"
        @click="emit('goto', s.line)"
      >
        <Box v-if="s.kind === 'class'" class="fl-icon fl-icon--class" :size="14" />
        <FunctionSquare v-else class="fl-icon fl-icon--func" :size="14" />
        <span class="truncate">{{ s.name }}</span>
        <span class="fl-line">{{ s.line }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.fl-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 6px 10px; border-bottom: 1px solid var(--et-border); flex-shrink: 0;
}
.fl-title { font-size: 13px; font-weight: 500; color: var(--et-fg); }
.fl-count { font-size: 11px; color: var(--et-fg-subtle); }
.fl-empty { padding: 16px; text-align: center; font-size: 12px; color: var(--et-fg-subtle); }
.fl-item {
  display: flex; align-items: center; gap: 8px;
  padding: 3px 10px; font-size: 13px; cursor: pointer;
}
.fl-item:hover { background: var(--et-bg-hover); }
.fl-icon { flex-shrink: 0; }
.fl-icon--class { color: #a855f7; }
.fl-icon--func { color: var(--et-accent); }
.fl-line { margin-left: auto; flex-shrink: 0; font-size: 11px; color: var(--et-fg-subtle); }
</style>
