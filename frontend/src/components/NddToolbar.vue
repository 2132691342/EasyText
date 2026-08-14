<script lang="ts" setup>
import { ref, computed } from 'vue'
import { useEditorStore, useSettingStore } from '@/stores'
import { FileText, FolderOpen, Save, SaveAll, Radio, X, Scissors, Copy, Clipboard, Undo2, Redo2, Search, Replace, Highlighter, Eraser, Edit3, ZoomIn, ZoomOut, WrapText, Eye, GripHorizontal, RefreshCw, ChevronLeft, ChevronRight, ArrowRightToLine, Repeat, PencilLine, Printer, Maximize, GitCompare, Code, Settings, Sun, Moon, TestTube, Palette, Braces, FileCode } from 'lucide-vue-next'

const emit = defineEmits<{ (e: 'toolbar-command', cmd: string, ...a: any[]): void }>()
const ed = useEditorStore()
const se = useSettingStore()
const hasTab = computed(() => ed.activeTab !== null)
const dirty = computed(() => ed.activeTab?.isDirty || false)
defineProps<{ iconSize: number }>()
const tailf = ref(false)
const autoSave = ref(false)
function cmd(c: string) { emit('toolbar-command', c) }
function tglTailf() { tailf.value = !tailf.value; emit('toolbar-command', 'toggle-tail', tailf.value) }
function tglAuto() { autoSave.value = !autoSave.value; emit('toolbar-command', 'toggle-auto-save-cycle', autoSave.value) }
</script>

<template>
  <div class="toolbar flex items-center h-7 px-0.5 border-b border-gray-300 dark:border-gray-700 bg-[#f0f0f0] dark:bg-[#2d2d2d] overflow-x-auto select-none">
    <!-- 1-7 文件 -->
    <button class="tb" title="新建 (Ctrl+T)" @click="cmd('new-file')"><FileText :size="iconSize"/></button>
    <button class="tb" title="打开 (Ctrl+O)" @click="cmd('open-file')"><FolderOpen :size="iconSize"/></button>
    <button class="tb" :class="dirty?'tbd':''" title="保存 (Ctrl+S)" :disabled="!hasTab" @click="cmd('save')"><Save :size="iconSize"/></button>
    <button class="tb" title="全部保存" :disabled="!hasTab" @click="cmd('save-all')"><SaveAll :size="iconSize"/></button>
    <button class="tb" :class="autoSave?'tba':''" title="循环自动保存" @click="tglAuto()"><Radio :size="iconSize"/></button>
    <button class="tb" title="关闭 (Ctrl+W)" :disabled="!hasTab" @click="cmd('close-tab')"><X :size="iconSize"/></button>
    <button class="tb" title="关闭所有" :disabled="!hasTab" @click="cmd('close-all')"><X :size="iconSize"/></button>
    <span class="sep"></span>
    <!-- 8-10 编辑 -->
    <button class="tb" title="剪切 (Ctrl+X)" :disabled="!hasTab" @click="cmd('cut')"><Scissors :size="iconSize"/></button>
    <button class="tb" title="复制 (Ctrl+C)" :disabled="!hasTab" @click="cmd('copy')"><Copy :size="iconSize"/></button>
    <button class="tb" title="粘贴 (Ctrl+V)" :disabled="!hasTab" @click="cmd('paste')"><Clipboard :size="iconSize"/></button>
    <span class="sep"></span>
    <!-- 11-12 撤销重做 -->
    <button class="tb" title="撤销 (Ctrl+Z)" :disabled="!hasTab" @click="cmd('undo')"><Undo2 :size="iconSize"/></button>
    <button class="tb" title="重做 (Ctrl+Y)" :disabled="!hasTab" @click="cmd('redo')"><Redo2 :size="iconSize"/></button>
    <span class="sep"></span>
    <!-- 13-15 查找替换标记 -->
    <button class="tb" title="查找 (Ctrl+F)" @click="cmd('find')"><Search :size="iconSize"/></button>
    <button class="tb" title="替换 (Ctrl+H)" :disabled="!hasTab" @click="cmd('replace')"><Replace :size="iconSize"/></button>
    <button class="tb" title="标记" :disabled="!hasTab" @click="cmd('mark-color')"><Edit3 :size="iconSize"/></button>
    <span class="sep"></span>
    <!-- 16-17 高亮 -->
    <button class="tb" title="单词高亮 (F8)" :disabled="!hasTab" @click="cmd('word-highlight')"><Highlighter :size="iconSize"/></button>
    <button class="tb" title="清除高亮 (F7)" :disabled="!hasTab" @click="cmd('clear-all-highlight')"><Eraser :size="iconSize"/></button>
    <span class="sep"></span>
    <!-- 18-19 缩放 -->
    <button class="tb" title="放大" :disabled="!hasTab" @click="cmd('zoom-in')"><ZoomIn :size="iconSize"/></button>
    <button class="tb" title="缩小" :disabled="!hasTab" @click="cmd('zoom-out')"><ZoomOut :size="iconSize"/></button>
    <span class="sep"></span>
    <!-- 20-23 视图 -->
    <button class="tb" title="自动换行" :disabled="!hasTab" @click="cmd('toggle-wrap')"><WrapText :size="iconSize"/></button>
    <button class="tb" title="显示空白" :disabled="!hasTab" @click="cmd('toggle-whitespace')"><Eye :size="iconSize"/></button>
    <button class="tb" title="缩进参考线" :disabled="!hasTab" @click="cmd('toggle-indent-guide')"><GripHorizontal :size="iconSize"/></button>
    <button class="tb" :class="tailf?'tba':''" title="跟踪文件 (tailf)" :disabled="!hasTab" @click="tglTailf()"><RefreshCw :size="iconSize"/></button>
    <span class="sep"></span>
    <!-- 24-26 十六进制 -->
    <button class="tb" title="上一页(十六进制)" :disabled="!hasTab" @click="cmd('pre-hex-page')"><ChevronLeft :size="iconSize"/></button>
    <button class="tb" title="下一页(十六进制)" :disabled="!hasTab" @click="cmd('next-hex-page')"><ChevronRight :size="iconSize"/></button>
    <button class="tb" title="转到页(十六进制)" :disabled="!hasTab" @click="cmd('goto-hex-page')"><ArrowRightToLine :size="iconSize"/></button>
    <span class="sep"></span>
    <!-- 27-28 批处理 -->
    <button class="tb" title="批量编码转换" @click="cmd('batch-convert')"><Repeat :size="iconSize"/></button>
    <button class="tb" title="批量重命名" @click="cmd('batch-rename')"><PencilLine :size="iconSize"/></button>
    <span class="sep"></span>
    <!-- 🆕 V2.0.0 新工具 -->
    <button class="tb" title="正则测试" @click="cmd('regex-tester')"><TestTube :size="iconSize"/></button>
    <button class="tb" title="全局搜索" @click="cmd('search-files')"><Search :size="iconSize"/></button>
    <button class="tb" title="脚本管理器" @click="cmd('script-manager')"><FileCode :size="iconSize"/></button>
    <button class="tb" title="取色器" @click="cmd('color-picker')"><Palette :size="iconSize"/></button>
    <!-- spacer -->
    <div class="flex-1"></div>
    <!-- 右侧 -->
    <button class="tb" title="打印 (Ctrl+P)" :disabled="!hasTab" @click="cmd('print')"><Printer :size="iconSize"/></button>
    <button class="tb" title="全屏 (F11)" @click="cmd('fullscreen')"><Maximize :size="iconSize"/></button>
    <button class="tb" title="文档对比" @click="cmd('open-diff')"><GitCompare :size="iconSize"/></button>
    <button class="tb" title="JSON 格式化" :disabled="!hasTab" @click="cmd('format-json')"><Code :size="iconSize"/></button>
    <button class="tb" title="设置" @click="cmd('preferences')"><Settings :size="iconSize"/></button>
    <button class="tb" :title="se.isDarkMode?'亮色主题':'暗色主题'" @click="se.toggleTheme()"><Sun v-if="se.isDarkMode" :size="iconSize"/><Moon v-else :size="iconSize"/></button>
  </div>
</template>

<style scoped>
.toolbar{font-size:12px}
.tb{display:inline-flex;align-items:center;justify-content:center;width:24px;height:22px;border:1px solid transparent;background:transparent;color:#4b5563;border-radius:2px;cursor:pointer;padding:0;margin:0 1px;transition:all .1s;flex-shrink:0}
.tb:hover:not(:disabled){background:rgba(59,130,246,.12);border-color:rgba(59,130,246,.3);color:#3b82f6}
.tb:disabled{opacity:.3;cursor:not-allowed}
.tbd{color:#3b82f6}
.tba{background:rgba(59,130,246,.18);color:#3b82f6;border-color:rgba(59,130,246,.4)}
.sep{width:1px;height:16px;background:rgba(0,0,0,.15);margin:0 3px;flex-shrink:0}
html.dark .tb{color:#bebebe}
html.dark .tb:hover:not(:disabled){background:rgba(96,165,250,.15);border-color:rgba(96,165,250,.4);color:#60a5fa}
html.dark .sep{background:rgba(255,255,255,.12)}
</style>
