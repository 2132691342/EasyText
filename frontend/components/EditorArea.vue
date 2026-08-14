<script lang="ts" setup>
import { computed } from 'vue'
import { useEditorStore } from '@/stores'
import TabBar from './TabBar.vue'
import CodeEditor from './editor/CodeEditor.vue'
import WelcomeScreen from './WelcomeScreen.vue'

const editorStore = useEditorStore()

const hasTabs = computed(() => editorStore.tabs.length > 0)
</script>

<template>
  <div class="h-full flex flex-col bg-white dark:bg-[#1e1e1e]">
    <!-- Tab bar -->
    <TabBar v-if="hasTabs" />

    <!-- Editor or welcome screen -->
    <div class="flex-1 overflow-hidden">
      <CodeEditor v-if="editorStore.activeTab" :tab="editorStore.activeTab" />
      <WelcomeScreen v-else />
    </div>
  </div>
</template>

<style scoped>
/* No additional styles needed */
</style>