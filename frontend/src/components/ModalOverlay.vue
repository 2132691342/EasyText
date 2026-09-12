<script lang="ts" setup>
/**
 * ModalOverlay v2.1 — 通用模态浮层
 *
 * 设计目标（M2 对话框统一）：
 *  - 标准化 size: sm/md/lg/xl/full（也兼容旧 width/height）
 *  - Esc 关闭（除非 closeOnEsc=false）
 *  - 点遮罩关闭（除非 closeOnBackdrop=false）
 *  - body scroll lock：打开时锁 body 滚动，防止背景跟随滚动
 *  - focus trap：Tab 在 modal 内循环聚焦
 *  - 80ms 进出动画（fade + scale）
 *  - header actions slot（右侧自定义按钮）
 *  - 关闭按钮用 Lucide X 替代 ✕ 字符
 */
import { computed, nextTick, onUnmounted, ref, watch } from 'vue'
import { X } from 'lucide-vue-next'

export type ModalSize = 'sm' | 'md' | 'lg' | 'xl' | 'full'

const props = withDefaults(defineProps<{
  visible: boolean
  title?: string
  subtitle?: string
  size?: ModalSize
  /** 兼容旧 prop：自定义宽度（如 '720px'）。优先级高于 size。 */
  width?: string
  /** 兼容旧 prop：自定义最大高度（如 '60vh'）。优先级高于 size。 */
  height?: string
  /** 是否响应 Esc 关闭（默认 true）。编辑器/正则测试器等可关闭。 */
  closeOnEsc?: boolean
  /** 是否响应点击遮罩关闭（默认 true）。 */
  closeOnBackdrop?: boolean
  /** 显示关闭按钮（默认 true） */
  showClose?: boolean
}>(), {
  size: 'md',
  closeOnEsc: true,
  closeOnBackdrop: true,
  showClose: true,
})

const emit = defineEmits<{
  (e: 'close'): void
}>()

const SIZE_WIDTH: Record<ModalSize, string> = {
  sm: '420px',
  md: '640px',
  lg: '820px',
  xl: '1080px',
  full: '95vw',
}
const SIZE_MAX_HEIGHT: Record<ModalSize, string> = {
  sm: '60vh',
  md: '80vh',
  lg: '85vh',
  xl: '90vh',
  full: '95vh',
}

const containerWidth = computed(() => props.width || SIZE_WIDTH[props.size])
const containerMaxHeight = computed(() => props.height || SIZE_MAX_HEIGHT[props.size])

const containerRef = ref<HTMLElement | null>(null)

// —— 关闭动作（暴露给 backdrop / Esc / close button 共用） ——
function requestClose() {
  emit('close')
}

// —— Esc 全局监听 ——
function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    e.stopPropagation()
    requestClose()
    return
  }
  // Focus trap：Tab 在容器内循环
  if (e.key === 'Tab' && containerRef.value) {
    const focusables = containerRef.value.querySelectorAll<HTMLElement>(
      'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])',
    )
    if (focusables.length === 0) {
      e.preventDefault()
      return
    }
    const first = focusables[0]
    const last  = focusables[focusables.length - 1]
    const active = document.activeElement as HTMLElement | null
    if (e.shiftKey && active === first) {
      e.preventDefault()
      last.focus()
    } else if (!e.shiftKey && active === last) {
      e.preventDefault()
      first.focus()
    }
  }
}

// —— 滚动锁 ——
let prevOverflow: string | null = null
function lockScroll() {
  if (prevOverflow !== null) return
  prevOverflow = document.body.style.overflow
  document.body.style.overflow = 'hidden'
}
function unlockScroll() {
  if (prevOverflow === null) return
  document.body.style.overflow = prevOverflow
  prevOverflow = null
}

watch(
  () => props.visible,
  async (v) => {
    if (v) {
      document.addEventListener('keydown', onKeydown, true)
      lockScroll()
      // 打开后把焦点放到第一个可聚焦元素（或容器本身）
      await nextTick()
      const c = containerRef.value
      if (c) {
        const first = c.querySelector<HTMLElement>(
          'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])',
        )
        ;(first ?? c).focus()
      }
    } else {
      document.removeEventListener('keydown', onKeydown, true)
      unlockScroll()
    }
  },
  { immediate: true },
)

onUnmounted(() => {
  document.removeEventListener('keydown', onKeydown, true)
  unlockScroll()
})
</script>

<template>
  <Teleport to="body">
    <Transition name="et-modal">
      <div
        v-if="visible"
        class="modal-backdrop"
        @click.self="closeOnBackdrop && requestClose()"
        @mousedown.self
      >
        <div
          ref="containerRef"
          class="modal-container"
          tabindex="-1"
          role="dialog"
          aria-modal="true"
          :style="{ width: containerWidth, maxHeight: containerMaxHeight }"
        >
          <!-- Header -->
          <div v-if="title || subtitle || $slots.header || showClose" class="modal-header">
            <slot name="header">
              <div class="modal-header-text">
                <h2 v-if="title" class="modal-title">{{ title }}</h2>
                <p v-if="subtitle" class="modal-subtitle">{{ subtitle }}</p>
              </div>
            </slot>
            <div class="modal-header-actions">
              <slot name="header-actions" />
              <button
                v-if="showClose"
                class="modal-close-btn"
                title="关闭 (Esc)"
                aria-label="关闭"
                @click="requestClose"
              >
                <X :size="14" :stroke-width="1.6" />
              </button>
            </div>
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
    </Transition>
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
  -webkit-user-select: none;
  user-select: none;
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
  outline: none;
  user-select: text;
  -webkit-user-select: text;
  min-width: 320px;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--et-space-2);
  padding: var(--et-space-2) var(--et-space-3);
  border-bottom: 1px solid var(--et-border);
  background: var(--et-bg-sunken);
}
.modal-header-text { min-width: 0; flex: 1; }
.modal-title {
  margin: 0;
  font-size: var(--et-text-md);
  font-weight: var(--et-fw-semibold);
  color: var(--et-fg);
  line-height: var(--et-lh-tight);
}
.modal-subtitle {
  margin: 2px 0 0;
  font-size: var(--et-text-xs);
  color: var(--et-fg-muted);
  line-height: var(--et-lh-tight);
}
.modal-header-actions {
  display: flex;
  align-items: center;
  gap: var(--et-space-1);
  flex-shrink: 0;
}
.modal-close-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: var(--et-h-control-sm);
  height: var(--et-h-control-sm);
  padding: 0;
  border: 0;
  border-radius: var(--et-radius-sm);
  background: transparent;
  color: var(--et-fg-subtle);
  cursor: pointer;
  transition: background-color 80ms ease, color 80ms ease;
}
.modal-close-btn:hover {
  background: var(--et-bg-hover);
  color: var(--et-fg);
}

.modal-body {
  flex: 1;
  overflow: auto;
  padding: var(--et-space-4);
}

.modal-footer {
  padding: var(--et-space-3) var(--et-space-4);
  border-top: 1px solid var(--et-border);
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: var(--et-space-2);
  background: var(--et-bg-sunken);
}

/* —— 进出动画（120ms） —— */
.et-modal-enter-active,
.et-modal-leave-active {
  transition: opacity 120ms ease-out;
}
.et-modal-enter-active .modal-container,
.et-modal-leave-active .modal-container {
  transition: transform 120ms ease-out, opacity 120ms ease-out;
}
.et-modal-enter-from,
.et-modal-leave-to {
  opacity: 0;
}
.et-modal-enter-from .modal-container,
.et-modal-leave-to .modal-container {
  transform: scale(.96) translateY(-8px);
  opacity: 0;
}
</style>