<script lang="ts" setup>
/**
 * 剪贴板历史 v2.1
 *  - 用 ModalOverlay 替代手撸 Teleport
 */
import { useEditorStore } from '@/stores'
import ModalOverlay from './ModalOverlay.vue'

defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'close'): void }>()
const ed = useEditorStore()

function insert(text: string) {
  document.dispatchEvent(new CustomEvent('editor-command', { detail: { cmd: 'insert-text', args: [text] } }))
  emit('close')
}
</script>

<template>
  <ModalOverlay :visible="visible" title="剪贴板历史记录" size="sm" @close="emit('close')">
    <div class="clipboard-list">
      <div v-if="!ed.clipboardHistory.length" class="et-empty">
        <div class="et-empty-title">暂无记录</div>
        <div class="et-empty-hint">复制或剪切过的内容会出现在这里</div>
      </div>
      <button
        v-for="(t, i) in ed.clipboardHistory"
        :key="i"
        class="clipboard-item"
        @click="insert(t)"
      >
        <div class="clipboard-item-head">#{{ i + 1 }} · {{ t.length }} 字符</div>
        <div class="clipboard-item-body">{{ t }}</div>
      </button>
    </div>
  </ModalOverlay>
</template>

<style scoped>
.clipboard-list {
  display: flex;
  flex-direction: column;
  gap: var(--et-space-1);
  max-height: 60vh;
  overflow-y: auto;
}
.clipboard-item {
  display: flex;
  flex-direction: column;
  gap: var(--et-space-1);
  padding: var(--et-space-2) var(--et-space-3);
  background: transparent;
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius-sm);
  color: var(--et-fg);
  cursor: pointer;
  text-align: left;
  transition: background-color 80ms ease, border-color 80ms ease;
}
.clipboard-item:hover {
  background: var(--et-bg-hover);
  border-color: var(--et-accent);
}
.clipboard-item-head {
  font-size: var(--et-text-xs);
  color: var(--et-fg-subtle);
}
.clipboard-item-body {
  font-size: var(--et-text-md);
  line-height: var(--et-lh-base);
  word-break: break-all;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>