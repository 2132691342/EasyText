<script lang="ts" setup>
import { computed } from 'vue'
import { useSettingStore, useEditorStore } from '@/stores'
import {
  Sun, Moon,
} from 'lucide-vue-next'

const emit = defineEmits([
  'open-file', 'open-directory', 'save', 'save-as', 'new-file',
  'open-settings', 'toggle-find', 'format-json', 'minify-json', 'validate-json', 'open-diff', 'open-converter',
  'comment-line', 'comment-block',
  'case-upper', 'case-lower', 'case-invert', 'case-title',
  'line-duplicate', 'line-remove', 'line-move-up', 'line-move-down',
  'line-remove-empty', 'line-remove-blank',
  'trim-head', 'trim-tail', 'trim-both',
  'tabs-to-spaces', 'spaces-to-tabs-all', 'spaces-to-tabs-leading',
  'sort-asc', 'sort-desc',
  'goto-line', 'toggle-word-wrap', 'toggle-indent-guide', 'toggle-show-whitespace',
  'open-explorer', 'open-cmd',
  'format-xml', 'md5-hash',
  'reload-file',
])

const settingStore = useSettingStore()
const editorStore = useEditorStore()

const hasActiveTab = computed(() => editorStore.activeTab !== null)
const isDirty = computed(() => editorStore.activeTab?.isDirty || false)

function handleCommand(cmd: string) {
  switch (cmd) {
    case 'new-file': emit('new-file'); break
    case 'open-file': emit('open-file'); break
    case 'open-directory': emit('open-directory'); break
    case 'save': emit('save'); break
    case 'save-as': emit('save-as'); break
    case 'toggle-find': emit('toggle-find'); break
    case 'format-json': emit('format-json'); break
    case 'minify-json': emit('minify-json'); break
    case 'validate-json': emit('validate-json'); break
    case 'open-converter': emit('open-converter'); break
    case 'open-diff': emit('open-diff'); break
    case 'open-settings': emit('open-settings'); break
    // Editor operations
    case 'comment-line': emit('comment-line'); break
    case 'comment-block': emit('comment-block'); break
    case 'case-upper': emit('case-upper'); break
    case 'case-lower': emit('case-lower'); break
    case 'case-invert': emit('case-invert'); break
    case 'case-title': emit('case-title'); break
    case 'line-duplicate': emit('line-duplicate'); break
    case 'line-remove': emit('line-remove'); break
    case 'line-move-up': emit('line-move-up'); break
    case 'line-move-down': emit('line-move-down'); break
    case 'line-remove-empty': emit('line-remove-empty'); break
    case 'line-remove-blank': emit('line-remove-blank'); break
    case 'trim-head': emit('trim-head'); break
    case 'trim-tail': emit('trim-tail'); break
    case 'trim-both': emit('trim-both'); break
    case 'tabs-to-spaces': emit('tabs-to-spaces'); break
    case 'spaces-to-tabs-all': emit('spaces-to-tabs-all'); break
    case 'spaces-to-tabs-leading': emit('spaces-to-tabs-leading'); break
    case 'sort-asc': emit('sort-asc'); break
    case 'sort-desc': emit('sort-desc'); break
    case 'goto-line': emit('goto-line'); break
    case 'toggle-word-wrap': emit('toggle-word-wrap'); break
    case 'toggle-indent-guide': emit('toggle-indent-guide'); break
    case 'toggle-show-whitespace': emit('toggle-show-whitespace'); break
    case 'open-explorer': emit('open-explorer'); break
    case 'open-cmd': emit('open-cmd'); break
    case 'format-xml': emit('format-xml'); break
    case 'md5-hash': emit('md5-hash'); break
    case 'reload-file': emit('reload-file'); break
  }
}
</script>

<template>
  <div class="toolbar h-10 flex items-center px-2 border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-[#252526]">
    <!-- File menu -->
    <el-dropdown trigger="click" @command="handleCommand" popper-class="toolbar-dropdown">
      <span class="el-dropdown-link flex items-center px-3 py-1.5 text-sm rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c] cursor-pointer select-none">
        文件<el-icon class="el-icon--right"><arrow-down /></el-icon>
      </span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item command="new-file">
            <el-icon><document-add /></el-icon>
            <span>新建文件</span>
            <span class="shortcut">Ctrl+N</span>
          </el-dropdown-item>
          <el-dropdown-item command="open-file">
            <el-icon><folder-opened /></el-icon>
            <span>打开文件</span>
            <span class="shortcut">Ctrl+O</span>
          </el-dropdown-item>
          <el-dropdown-item command="open-directory">
            <el-icon><folder /></el-icon>
            <span>打开文件夹</span>
            <span class="shortcut">Ctrl+Shift+O</span>
          </el-dropdown-item>
          <el-dropdown-item divided command="save" :disabled="!hasActiveTab">
            <el-icon><coin /></el-icon>
            <span>保存</span>
            <span class="shortcut">Ctrl+S</span>
          </el-dropdown-item>
          <el-dropdown-item command="save-as" :disabled="!hasActiveTab">
            <el-icon><tickets /></el-icon>
            <span>另存为</span>
            <span class="shortcut">Ctrl+Shift+S</span>
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>

    <!-- Edit menu -->
    <el-dropdown trigger="click" @command="handleCommand" popper-class="toolbar-dropdown" class="ml-1">
      <span class="el-dropdown-link flex items-center px-3 py-1.5 text-sm rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c] cursor-pointer select-none">
        编辑<el-icon class="el-icon--right"><arrow-down /></el-icon>
      </span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item command="toggle-find">
            <span>查找替换</span>
            <span class="shortcut">Ctrl+F</span>
          </el-dropdown-item>
          <el-dropdown-item command="goto-line" :disabled="!hasActiveTab">
            <span>跳转到行...</span>
            <span class="shortcut">Ctrl+G</span>
          </el-dropdown-item>
          <el-dropdown-item divided command="comment-line" :disabled="!hasActiveTab">
            <span>注释/取消注释</span>
            <span class="shortcut">Ctrl+/</span>
          </el-dropdown-item>
          <el-dropdown-item command="comment-block" :disabled="!hasActiveTab">
            <span>块注释</span>
            <span class="shortcut">Ctrl+Shift+/</span>
          </el-dropdown-item>
          <el-dropdown-item divided command="line-duplicate" :disabled="!hasActiveTab">
            <span>复制当前行</span>
            <span class="shortcut">Ctrl+D</span>
          </el-dropdown-item>
          <el-dropdown-item command="line-remove" :disabled="!hasActiveTab">
            <span>删除当前行</span>
            <span class="shortcut">Ctrl+L</span>
          </el-dropdown-item>
          <el-dropdown-item command="line-move-up" :disabled="!hasActiveTab">
            <span>上移当前行</span>
            <span class="shortcut">Ctrl+Shift+Up</span>
          </el-dropdown-item>
          <el-dropdown-item command="line-move-down" :disabled="!hasActiveTab">
            <span>下移当前行</span>
            <span class="shortcut">Ctrl+Shift+Down</span>
          </el-dropdown-item>
          <el-dropdown-item divided command="trim-head" :disabled="!hasActiveTab">
            <span>删除行首空格</span>
          </el-dropdown-item>
          <el-dropdown-item command="trim-tail" :disabled="!hasActiveTab">
            <span>删除行尾空格</span>
          </el-dropdown-item>
          <el-dropdown-item command="trim-both" :disabled="!hasActiveTab">
            <span>删除首尾空格</span>
          </el-dropdown-item>
          <el-dropdown-item command="line-remove-empty" :disabled="!hasActiveTab">
            <span>删除空行(含空白)</span>
          </el-dropdown-item>
          <el-dropdown-item command="line-remove-blank" :disabled="!hasActiveTab">
            <span>删除空白行</span>
          </el-dropdown-item>
          <el-dropdown-item divided command="tabs-to-spaces" :disabled="!hasActiveTab">
            <span>Tab转空格</span>
          </el-dropdown-item>
          <el-dropdown-item command="spaces-to-tabs-all" :disabled="!hasActiveTab">
            <span>空格全部转Tab</span>
          </el-dropdown-item>
          <el-dropdown-item command="spaces-to-tabs-leading" :disabled="!hasActiveTab">
            <span>行首空格转Tab</span>
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>

    <!-- Search menu -->
    <el-dropdown trigger="click" @command="handleCommand" popper-class="toolbar-dropdown" class="ml-1">
      <span class="el-dropdown-link flex items-center px-3 py-1.5 text-sm rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c] cursor-pointer select-none">
        搜索<el-icon class="el-icon--right"><arrow-down /></el-icon>
      </span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item command="toggle-find">
            <span>查找...</span>
            <span class="shortcut">Ctrl+F</span>
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>

    <!-- View menu -->
    <el-dropdown trigger="click" @command="handleCommand" popper-class="toolbar-dropdown" class="ml-1">
      <span class="el-dropdown-link flex items-center px-3 py-1.5 text-sm rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c] cursor-pointer select-none">
        视图<el-icon class="el-icon--right"><arrow-down /></el-icon>
      </span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item command="toggle-word-wrap" :disabled="!hasActiveTab">
            <span>自动换行</span>
          </el-dropdown-item>
          <el-dropdown-item command="toggle-show-whitespace" :disabled="!hasActiveTab">
            <span>显示空白字符</span>
          </el-dropdown-item>
          <el-dropdown-item command="toggle-indent-guide" :disabled="!hasActiveTab">
            <span>显示缩进参考线</span>
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>

    <!-- Tools menu -->
    <el-dropdown trigger="click" @command="handleCommand" popper-class="toolbar-dropdown" class="ml-1">
      <span class="el-dropdown-link flex items-center px-3 py-1.5 text-sm rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c] cursor-pointer select-none">
        工具<el-icon class="el-icon--right"><arrow-down /></el-icon>
      </span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item command="format-json" :disabled="!hasActiveTab">
            <span>JSON 格式化</span>
            <span class="shortcut">Ctrl+Shift+F</span>
          </el-dropdown-item>
          <el-dropdown-item command="minify-json" :disabled="!hasActiveTab">
            <span>JSON 压缩</span>
            <span class="shortcut">Ctrl+Shift+M</span>
          </el-dropdown-item>
          <el-dropdown-item command="validate-json" :disabled="!hasActiveTab">
            <span>JSON 校验</span>
          </el-dropdown-item>
          <el-dropdown-item command="format-xml" :disabled="!hasActiveTab">
            <span>XML 格式化</span>
          </el-dropdown-item>
          <el-dropdown-item divided command="open-converter">
            <span>格式转换 (JSON/YAML/TOML/XML)</span>
          </el-dropdown-item>
          <el-dropdown-item command="open-diff">
            <span>文档对比 (Diff)</span>
          </el-dropdown-item>
          <el-dropdown-item divided command="md5-hash" :disabled="!hasActiveTab">
            <span>MD5/SHA 哈希计算</span>
          </el-dropdown-item>
          <el-dropdown-item divided command="open-explorer" :disabled="!hasActiveTab">
            <span>在资源管理器中打开</span>
          </el-dropdown-item>
          <el-dropdown-item command="open-cmd" :disabled="!hasActiveTab">
            <span>打开命令行</span>
          </el-dropdown-item>
          <el-dropdown-item divided command="reload-file" :disabled="!hasActiveTab">
            <span>从磁盘重新加载</span>
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>

    <!-- Case conversion menu -->
    <el-dropdown trigger="click" @command="handleCommand" popper-class="toolbar-dropdown" class="ml-1">
      <span class="el-dropdown-link flex items-center px-3 py-1.5 text-sm rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c] cursor-pointer select-none">
        大小写<el-icon class="el-icon--right"><arrow-down /></el-icon>
      </span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item command="case-upper" :disabled="!hasActiveTab">
            <span>转大写</span>
          </el-dropdown-item>
          <el-dropdown-item command="case-lower" :disabled="!hasActiveTab">
            <span>转小写</span>
          </el-dropdown-item>
          <el-dropdown-item command="case-title" :disabled="!hasActiveTab">
            <span>首字母大写</span>
          </el-dropdown-item>
          <el-dropdown-item command="case-invert" :disabled="!hasActiveTab">
            <span>反转大小写</span>
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>

    <!-- Sort menu -->
    <el-dropdown trigger="click" @command="handleCommand" popper-class="toolbar-dropdown" class="ml-1">
      <span class="el-dropdown-link flex items-center px-3 py-1.5 text-sm rounded hover:bg-gray-100 dark:hover:bg-[#3c3c3c] cursor-pointer select-none">
        排序<el-icon class="el-icon--right"><arrow-down /></el-icon>
      </span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item command="sort-asc" :disabled="!hasActiveTab">
            <span>升序排列</span>
          </el-dropdown-item>
          <el-dropdown-item command="sort-desc" :disabled="!hasActiveTab">
            <span>降序排列</span>
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>

    <!-- Quick actions -->
    <div class="flex items-center ml-auto gap-1">
      <el-tooltip content="新建文件 Ctrl+N" placement="bottom" :show-after="500">
        <el-button size="small" text @click="emit('new-file')">
          <el-icon><document-add /></el-icon>
        </el-button>
      </el-tooltip>
      <el-tooltip content="打开文件 Ctrl+O" placement="bottom" :show-after="500">
        <el-button size="small" text @click="emit('open-file')">
          <el-icon><folder-opened /></el-icon>
        </el-button>
      </el-tooltip>
      <el-tooltip content="打开文件夹 Ctrl+Shift+O" placement="bottom" :show-after="500">
        <el-button size="small" text @click="emit('open-directory')">
          <el-icon><folder /></el-icon>
        </el-button>
      </el-tooltip>
      <el-tooltip content="保存 Ctrl+S" placement="bottom" :show-after="500">
        <el-button size="small" text @click="emit('save')" :disabled="!hasActiveTab" :type="isDirty ? 'primary' : 'default'">
          <el-icon><coin /></el-icon>
        </el-button>
      </el-tooltip>

      <div class="mx-2 h-4 w-px bg-gray-200 dark:bg-gray-600"></div>

      <el-tooltip :content="settingStore.isDarkMode ? '切换亮色主题' : '切换暗色主题'" placement="bottom" :show-after="500">
        <el-button size="small" text @click="settingStore.toggleTheme()">
          <Sun v-if="settingStore.isDarkMode" class="w-4 h-4" />
          <Moon v-else class="w-4 h-4" />
        </el-button>
      </el-tooltip>

      <el-tooltip content="设置" placement="bottom" :show-after="500">
        <el-button size="small" text @click="emit('open-settings')">
          <el-icon><setting /></el-icon>
        </el-button>
      </el-tooltip>
    </div>
  </div>
</template>

<style scoped>
.toolbar-dropdown .shortcut {
  margin-left: auto;
  padding-left: 16px;
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

el-dropdown-item {
  display: flex !important;
  align-items: center;
}

el-dropdown-item .el-icon {
  margin-right: 8px;
}
</style>