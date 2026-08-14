<script lang="ts" setup>
import { ref } from 'vue'
import { useSettingStore, useEditorStore } from '@/stores'
import {
  FileText, FolderOpen, Save, Plus, Search, Settings, Moon, Sun, FormatJson, Diff, ChevronDown
} from 'lucide-vue-next'

const emit = defineEmits(['open-file', 'open-directory', 'save', 'new-file'])

const settingStore = useSettingStore()
const editorStore = useEditorStore()

const showMenu = ref(false)
const menuType = ref('')

const hasActiveTab = computed(() => editorStore.activeTab !== null)
const isDirty = computed(() => editorStore.activeTab?.isDirty || false)

function toggleMenu(type: string) {
  if (menuType.value === type) {
    showMenu.value = !showMenu.value
  } else {
    menuType.value = type
    showMenu.value = true
  }
}

function closeMenu() {
  showMenu.value = false
  menuType.value = ''
}
</script>

<template>
  <div class="toolbar h-10 flex items-center px-2 border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-[#252526]">
    <!-- File menu -->
    <div class="relative">
      <button
        class="menu-trigger flex items-center px-3 py-1.5 text-sm rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c]"
        @click="toggleMenu('file')"
      >
        文件
        <ChevronDown class="w-3 h-3 ml-1" />
      </button>
      <div v-if="showMenu && menuType === 'file'" class="menu-dropdown absolute top-full left-0 mt-1 bg-white dark:bg-[#3c3c3c] border border-gray-200 dark:border-gray-600 rounded shadow-lg z-50">
        <div class="menu-item" @click="emit('new-file'); closeMenu()">
          <Plus class="w-4 h-4 mr-2" />
          新建文件
          <span class="ml-auto text-xs text-gray-400">Ctrl+N</span>
        </div>
        <div class="menu-item" @click="emit('open-file'); closeMenu()">
          <FileText class="w-4 h-4 mr-2" />
          打开文件
          <span class="ml-auto text-xs text-gray-400">Ctrl+O</span>
        </div>
        <div class="menu-item" @click="emit('open-directory'); closeMenu()">
          <FolderOpen class="w-4 h-4 mr-2" />
          打开文件夹
          <span class="ml-auto text-xs text-gray-400">Ctrl+Shift+O</span>
        </div>
        <div class="menu-divider"></div>
        <div class="menu-item" @click="emit('save'); closeMenu()" :class="{ 'opacity-50': !hasActiveTab }">
          <Save class="w-4 h-4 mr-2" />
          保存
          <span class="ml-auto text-xs text-gray-400">Ctrl+S</span>
        </div>
      </div>
    </div>

    <!-- Edit menu -->
    <div class="relative ml-1">
      <button
        class="menu-trigger flex items-center px-3 py-1.5 text-sm rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c]"
        @click="toggleMenu('edit')"
      >
        编辑
        <ChevronDown class="w-3 h-3 ml-1" />
      </button>
      <div v-if="showMenu && menuType === 'edit'" class="menu-dropdown absolute top-full left-0 mt-1 bg-white dark:bg-[#3c3c3c] border border-gray-200 dark:border-gray-600 rounded shadow-lg z-50">
        <div class="menu-item">
          查找替换
          <span class="ml-auto text-xs text-gray-400">Ctrl+F</span>
        </div>
      </div>
    </div>

    <!-- Tools menu -->
    <div class="relative ml-1">
      <button
        class="menu-trigger flex items-center px-3 py-1.5 text-sm rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c]"
        @click="toggleMenu('tools')"
      >
        工具
        <ChevronDown class="w-3 h-3 ml-1" />
      </button>
      <div v-if="showMenu && menuType === 'tools'" class="menu-dropdown absolute top-full left-0 mt-1 bg-white dark:bg-[#3c3c3c] border border-gray-200 dark:border-gray-600 rounded shadow-lg z-50">
        <div class="menu-item">
          <FormatJson class="w-4 h-4 mr-2" />
          JSON 格式化
          <span class="ml-auto text-xs text-gray-400">Ctrl+Shift+F</span>
        </div>
        <div class="menu-item">
          <Diff class="w-4 h-4 mr-2" />
          文档对比
        </div>
      </div>
    </div>

    <!-- Quick actions -->
    <div class="flex items-center ml-auto gap-1">
      <button
        class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c] text-gray-600 dark:text-gray-300"
        title="新建文件"
        @click="emit('new-file')"
      >
        <Plus class="w-4 h-4" />
      </button>
      <button
        class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c] text-gray-600 dark:text-gray-300"
        title="打开文件"
        @click="emit('open-file')"
      >
        <FileText class="w-4 h-4" />
      </button>
      <button
        class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c] text-gray-600 dark:text-gray-300"
        title="打开文件夹"
        @click="emit('open-directory')"
      >
        <FolderOpen class="w-4 h-4" />
      </button>
      <button
        class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c] text-gray-600 dark:text-gray-300"
        title="保存"
        @click="emit('save')"
        :disabled="!hasActiveTab"
      >
        <Save class="w-4 h-4" :class="{ 'text-blue-500': isDirty }" />
      </button>

      <div class="mx-2 h-4 w-px bg-gray-200 dark:bg-gray-600"></div>

      <button
        class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c] text-gray-600 dark:text-gray-300"
        title="切换主题"
        @click="settingStore.toggleTheme()"
      >
        <Sun v-if="settingStore.isDarkMode" class="w-4 h-4" />
        <Moon v-else class="w-4 h-4" />
      </button>
    </div>
  </div>
</template>

<style scoped>
.menu-dropdown {
  min-width: 200px;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  cursor: pointer;
  transition: background-color 0.15s;
}

.menu-item:hover {
  background-color: rgba(0, 0, 0, 0.05);
}

html.dark .menu-item:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.menu-divider {
  height: 1px;
  background-color: #e5e5e5;
  margin: 4px 0;
}

html.dark .menu-divider {
  background-color: #404040;
}
</style>