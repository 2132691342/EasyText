<script lang="ts" setup>
defineProps<{
  visible: boolean
  title?: string
  width?: string
  height?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()
</script>

<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="modal-backdrop"
      @click.self="emit('close')"
      @keydown.escape="emit('close')"
    >
      <div
        class="modal-container"
        :style="{ width: width || '640px', maxHeight: height || '80vh' }"
      >
        <!-- Header -->
        <div v-if="title || $slots.header" class="modal-header">
          <slot name="header">
            <h2 class="text-sm font-semibold text-gray-700 dark:text-gray-200">{{ title }}</h2>
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
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(2px);
}

.modal-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  animation: modal-in 0.15s ease-out;
}

:global(html.dark) .modal-container {
  background: #1e1e1e;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.6);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  border-bottom: 1px solid #e5e7eb;
}

:global(html.dark) .modal-header {
  border-bottom-color: #374151;
}

.modal-close-btn {
  padding: 4px 8px;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: #9ca3af;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.15s;
}

.modal-close-btn:hover {
  background: #f3f4f6;
  color: #374151;
}

:global(html.dark) .modal-close-btn:hover {
  background: #374151;
  color: #e5e7eb;
}

.modal-body {
  flex: 1;
  overflow: auto;
  padding: 16px;
}

.modal-footer {
  padding: 10px 16px;
  border-top: 1px solid #e5e7eb;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

:global(html.dark) .modal-footer {
  border-top-color: #374151;
}

@keyframes modal-in {
  from {
    opacity: 0;
    transform: scale(0.96) translateY(-8px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}
</style>
