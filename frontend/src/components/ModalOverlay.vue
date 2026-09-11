<script lang="ts" setup>
/**
 * 通用模态浮层
 *
 * 本轮修复：
 *  - ESC 关不掉：原来把 @keydown.escape 挂在 backdrop div 上，而该 div 没有
 *    tabindex、也不一定会获得焦点，键盘事件根本冒泡不到它。改为在 visible
 *    期间直接监听 document（并在卸载/关闭时移除）。
 *  - 样式令牌化：容器/头部/按钮原本是字面色 + html.dark 覆盖。
 */
import { watch, onUnmounted } from 'vue'

const props = defineProps<{
  visible: boolean
  title?: string
  width?: string
  height?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}

watch(
  () => props.visible,
  (v) => {
    if (v) document.addEventListener('keydown', onKeydown)
    else document.removeEventListener('keydown', onKeydown)
  },
  { immediate: true },
)

onUnmounted(() => document.removeEventListener('keydown', onKeydown))
</script>

<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="modal-backdrop"
      @click.self="emit('close')"
    >
      <div
        class="modal-container"
        :style="{ width: width || '640px', maxHeight: height || '80vh' }"
      >
        <!-- Header -->
        <div v-if="title || $slots.header" class="modal-header">
          <slot name="header">
            <h2 class="modal-title">{{ title }}</h2>
          </slot>
          <button
            class="modal-close-btn"
            @click="emit('close')"
            aria-label="关闭"
          >&#10005;</button>
        </div>

        <!-- Body -->
        <div class="modal-body">
          <slot />
        </div>

        <!-- Footer -->
        <div v-if="$slots.footer" class="modal-footer">
          <slot name="footer" />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, .45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(2px);
}

.modal-container {
  background: var(--et-bg-elevated);
  color: var(--et-fg);
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius);
  box-shadow: var(--et-shadow-md);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  animation: modal-in .15s ease-out;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 14px;
  border-bottom: 1px solid var(--et-border);
  background: var(--et-bg-sunken);
  border-radius: var(--et-radius) var(--et-radius) 0 0;
}
.modal-title { margin: 0; font-size: 13px; font-weight: 600; color: var(--et-fg); }

.modal-close-btn {
  padding: 4px 8px;
  border: none;
  border-radius: var(--et-radius-sm);
  background: transparent;
  color: var(--et-fg-subtle);
  font-size: 14px;
  cursor: pointer;
  transition: background .15s ease, color .15s ease;
}
.modal-close-btn:hover { background: var(--et-bg-hover); color: var(--et-fg); }

.modal-body {
  flex: 1;
  overflow: auto;
  padding: 14px;
}

.modal-footer {
  padding: 10px 14px;
  border-top: 1px solid var(--et-border);
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

@keyframes modal-in {
  from {
    opacity: 0;
    transform: scale(.96) translateY(-8px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}
</style>
