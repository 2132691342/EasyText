<script lang="ts" setup>
import { computed } from 'vue'
import { useFileStore, useEditorStore } from '@/stores'
import { ReadFile } from '../wailsjs/go/main/App'
import type { TreeNode } from '@/types'
import {
  ChevronDown, ChevronRight, FileText, Folder, FolderOpen,
  FileCode, FileJson, File, Image, Database, Terminal
} from 'lucide-vue-next'

const props = defineProps<{
  node: TreeNode
  depth?: number
}>()

const fileStore = useFileStore()
const editorStore = useEditorStore()

const depth = computed(() => props.depth || 0)
const isExpanded = computed(() => fileStore.isExpanded(props.node.path))
const isSelected = computed(() => fileStore.selectedPath === props.node.path)

function toggleExpand() {
  fileStore.togglePath(props.node.path)
}

async function handleClick() {
  if (props.node.isDir) {
    toggleExpand()
  } else {
    fileStore.setSelectedPath(props.node.path)

    // Check if already open
    const existingTab = editorStore.getTabByPath(props.node.path)
    if (existingTab) {
      editorStore.activateTab(existingTab.id)
      return
    }

    // Open file
    try {
      const result = await ReadFile(props.node.path)
      if (result) {
        editorStore.createTab(props.node.path, result.content, result.info.encoding, result.info.lineEnding)
      }
    } catch (error) {
      console.error('Failed to open file:', error)
    }
  }
}

function getFileIcon(node: TreeNode) {
  if (node.isDir) {
    return isExpanded.value ? FolderOpen : Folder
  }

  const ext = node.ext?.toLowerCase() || ''
  const iconMap: Record<string, any> = {
    'js': FileCode,
    'jsx': FileCode,
    'ts': FileCode,
    'tsx': FileCode,
    'json': FileJson,
    'html': FileCode,
    'css': FileCode,
    'py': FileCode,
    'java': FileCode,
    'go': FileCode,
    'sql': Database,
    'sh': Terminal,
    'bash': Terminal,
    'md': FileText,
    'txt': FileText,
    'png': Image,
    'jpg': Image,
    'jpeg': Image,
    'gif': Image,
    'svg': Image,
  }

  return iconMap[ext] || File
}
</script>

<template>
  <div>
    <div
      class="file-tree-item flex items-center py-1"
      :style="{ paddingLeft: `${depth * 16 + 8}px` }"
      :class="{ selected: isSelected }"
      @click="handleClick"
    >
      <!-- Expand/collapse icon for directories -->
      <span v-if="node.isDir" class="w-4 h-4 mr-1 flex items-center justify-center">
        <ChevronDown v-if="isExpanded" class="w-3 h-3 text-gray-400" />
        <ChevronRight v-else class="w-3 h-3 text-gray-400" />
      </span>
      <span v-else class="w-4 h-4 mr-1"></span>

      <!-- File/folder icon -->
      <component :is="getFileIcon(node)" class="w-4 h-4 mr-2 text-gray-400" />

      <!-- Name -->
      <span class="text-sm truncate">{{ node.name }}</span>
    </div>

    <!-- Children -->
    <div v-if="node.isDir && isExpanded && node.children">
      <FileTree
        v-for="child in node.children"
        :key="child.path"
        :node="child"
        :depth="depth + 1"
      />
    </div>
  </div>
</template>

<style scoped>
.file-tree-item.selected {
  background-color: rgba(59, 130, 246, 0.1);
}

html.dark .file-tree-item.selected {
  background-color: rgba(59, 130, 246, 0.2);
}
</style>