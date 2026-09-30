<script lang="ts" setup>
/**
 * EditorArea — 编辑区容器
 *
 * 单视图：TabBar + 当前 tab 的视图（code/image/hex/log/welcome）。
 * 分屏视图（参考 notepad++ 双视图）：
 *  - tabs 是全局文档列表（Buffer），splitTabId 只是第二个视图的索引；
 *  - 主视图沿用原有单视图结构（切换分屏不重挂载，保住 undo 历史）；
 *  - 第二视图 = header（文件名 + 关闭）+ CodeEditor；
 *  - 中间可拖拽分隔条，比例持久化到 localStorage；
 *  - 点击聚焦决定 activeEditorView，编辑命令作用于聚焦视图。
 */
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { X } from 'lucide-vue-next'
import { useEditorStore } from '@/stores'
import TabBar from './TabBar.vue'
import CodeEditor from './editor/CodeEditor.vue'
import WelcomeScreen from './WelcomeScreen.vue'
import ImageViewer from './viewer/ImageViewer.vue'
import ImageEditor from './viewer/ImageEditor.vue'
import HexViewer from './viewer/HexViewer.vue'
import LogViewer from './viewer/LogViewer.vue'

const editorStore = useEditorStore()
const emit = defineEmits(['open-diff', 'open-converter'])

const hasTabs = computed(() => editorStore.tabs.length > 0)
const activeTab = computed(() => editorStore.activeTab)
const viewType = computed(() => activeTab.value?.viewType || 'code')
const splitTab = computed(() => editorStore.splitTab)

/** 分屏可用：两个视图都必须是 code 视图（其他视图类型不支持双开） */
const isSplit = computed(() =>
  !!splitTab.value
  && !!activeTab.value
  && activeTab.value.viewType === 'code'
  && splitTab.value.viewType === 'code')

// —— 分屏比例（主视图宽度百分比），拖拽后持久化 ——
const SPLIT_RATIO_KEY = 'easytext-split-ratio'
const splitRatio = ref<number>((() => {
  const saved = Number(localStorage.getItem(SPLIT_RATIO_KEY))
  return saved >= 20 && saved <= 80 ? saved : 50
})())
const resizing = ref(false)

function startSplitResize() {
  resizing.value = true
  document.addEventListener('mousemove', onSplitResize)
  document.addEventListener('mouseup', endSplitResize)
}
function onSplitResize(e: MouseEvent) {
  if (!resizing.value) return
  const root = (e.currentTarget as HTMLElement).parentElement
  if (!root) return
  const pct = (e.clientX - root.getBoundingClientRect().left) / root.clientWidth * 100
  splitRatio.value = Math.min(80, Math.max(20, pct))
}
function endSplitResize() {
  if (!resizing.value) return
  resizing.value = false
  document.removeEventListener('mousemove', onSplitResize)
  document.removeEventListener('mouseup', endSplitResize)
  try { localStorage.setItem(SPLIT_RATIO_KEY, String(Math.round(splitRatio.value))) } catch { /* ignore */ }
}

// —— 分屏命令（split-* 由菜单/快捷键经 execEd 派发到 document） ——
function onEditorCommand(e: Event) {
  const detail = (e as CustomEvent).detail
  if (!detail) return
  const cmd = typeof detail === 'string' ? detail : detail.cmd
  if (cmd === 'split-toggle') toggleSplit()
  else if (cmd === 'split-close') editorStore.closeSplitView()
  else if (cmd === 'split-swap') {
    if (!isSplit.value) { ElMessage.warning('请先开启分屏'); return }
    editorStore.swapSplitView()
  }
}
function toggleSplit() {
  if (editorStore.splitTabId) { editorStore.closeSplitView(); return }
  const t = activeTab.value
  if (!t) { ElMessage.warning('没有可分屏的文档'); return }
  if (t.viewType !== 'code') { ElMessage.warning('当前视图类型不支持分屏'); return }
  editorStore.openTabInSplit(t.id)
}

onMounted(() => document.addEventListener('editor-command', onEditorCommand as EventListener))
onUnmounted(() => document.removeEventListener('editor-command', onEditorCommand as EventListener))
</script>

<template>
  <div class="h-full flex flex-col bg-[var(--et-bg)]">
    <!-- Tab bar -->
    <TabBar v-if="hasTabs" />

    <!-- Editor or welcome screen -->
    <div class="flex-1 overflow-hidden" :class="{ 'split-root': isSplit }">
      <!-- 主视图：保持原单视图结构（v-if 链不变），分屏时仅加宽度约束 -->
      <div
        class="split-pane"
        :class="{ 'is-inactive': isSplit && editorStore.activeEditorView === 'second' }"
        :style="isSplit ? { flex: `0 0 ${splitRatio}%`, width: `${splitRatio}%` } : undefined"
        @mousedown.capture="editorStore.setActiveEditorView('main')"
      >
        <!-- Code editor (default) - keep alive across tab switches via in-place content update -->
        <CodeEditor
          v-if="activeTab && viewType === 'code'"
          :tab="activeTab"
          view-id="main"
        />

        <!-- Image viewer -->
        <ImageViewer
          v-else-if="activeTab && viewType === 'image'"
          :tab="activeTab"
        />

        <!-- Image editor (裁剪/旋转/缩放/格式转换) -->
        <ImageEditor
          v-else-if="activeTab && viewType === 'image-edit'"
          :file-path="activeTab.path"
        />

        <!-- Hex viewer -->
        <HexViewer
          v-else-if="activeTab && viewType === 'hex'"
          :tab="activeTab"
        />

        <!-- Log viewer -->
        <LogViewer
          v-else-if="activeTab && viewType === 'log'"
          :tab="activeTab"
        />

        <!-- Welcome screen (no tabs) -->
        <WelcomeScreen v-else @open-diff="emit('open-diff')" @open-converter="emit('open-converter')" />
      </div>

      <!-- 可拖拽分隔条 + 第二视图 -->
      <template v-if="isSplit && splitTab">
        <div
          class="split-resizer"
          :class="{ 'is-resizing': resizing }"
          @mousedown="startSplitResize"
        />
        <div
          class="split-pane flex-1"
          :class="{ 'is-inactive': editorStore.activeEditorView === 'main' }"
          @mousedown.capture="editorStore.setActiveEditorView('second')"
        >
          <div class="split-pane-header">
            <span v-if="splitTab.isDirty" class="et-dot et-dot-warn" />
            <span class="split-pane-name" :title="splitTab.path || splitTab.name">{{ splitTab.name }}</span>
            <button
              class="et-icon-btn-sm"
              :title="`关闭分屏 (${splitTab.name})`"
              @click="editorStore.closeSplitView()"
            >
              <X :size="14" :stroke-width="1.6" />
            </button>
          </div>
          <div class="flex-1 min-h-0 overflow-hidden">
            <CodeEditor :tab="splitTab" view-id="second" />
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
/* Loading fallback for async components */
:deep(.async-component-loading) {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}
</style>
