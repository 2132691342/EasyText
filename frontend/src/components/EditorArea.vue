<script lang="ts" setup>
import { computed } from 'vue'
import { useEditorStore } from '@/stores'
import TabBar from './TabBar.vue'
import CodeEditor from './editor/CodeEditor.vue'
import WelcomeScreen from './WelcomeScreen.vue'
import ImageViewer from './viewer/ImageViewer.vue'
import ImageEditor from './viewer/ImageEditor.vue'  // 🆕 V2.0.0
import HexViewer from './viewer/HexViewer.vue'
import LogViewer from './viewer/LogViewer.vue'

const editorStore = useEditorStore()
const emit = defineEmits(['open-diff'])

const hasTabs = computed(() => editorStore.tabs.length > 0)
const activeTab = computed(() => editorStore.activeTab)
const viewType = computed(() => activeTab.value?.viewType || 'code')
</script>

<template>
  <div class="h-full flex flex-col bg-white dark:bg-[#1e1e1e]">
    <!-- Tab bar -->
    <TabBar v-if="hasTabs" />

    <!-- Editor or welcome screen -->
    <div class="flex-1 overflow-hidden">
      <!-- Code editor (default) - keep alive across tab switches via in-place content update -->
      <CodeEditor
        v-if="activeTab && viewType === 'code'"
        :tab="activeTab"
      />

      <!-- Image viewer -->
      <ImageViewer
        v-else-if="activeTab && viewType === 'image'"
        :tab="activeTab"
      />

      <!-- 🆕 V2.0.0 Image editor (裁剪/旋转/缩放/格式转换) -->
      <ImageEditor
        v-else-if="activeTab && viewType === 'image-edit'"
        :file-path="activeTab.path"
      />

      <!-- Hex viewer -->
      <HexViewer
        v-else-if="activeTab && viewType === 'hex'"
        :tab="activeTab"
      />

      <!-- 🆕 V2.0.0 Log viewer -->
      <LogViewer
        v-else-if="activeTab && viewType === 'log'"
        :tab="activeTab"
      />

      <!-- Welcome screen (no tabs) -->
      <WelcomeScreen v-else @open-diff="emit('open-diff')" />
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