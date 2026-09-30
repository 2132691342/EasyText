<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { FileText, FolderOpen, FileJson, Diff, Clock, X } from 'lucide-vue-next'
import { OpenFileDialog, OpenDirectoryDialog, ReadFile, GetRecentFiles, GetRecentFolders, AddRecentEntry, ClearRecentFiles, ClearRecentFolders } from '../../wailsjs/go/main/App'
import { useEditorStore, useFileStore } from '@/stores'
import { getFileExtension, getTabViewType } from '@/utils'
import type { RecentEntry } from '@/types'

const editorStore = useEditorStore()
const fileStore = useFileStore()

const emit = defineEmits(['open-diff', 'open-converter'])

// 最近访问
const recentFiles = ref<RecentEntry[]>([])
const recentFolders = ref<RecentEntry[]>([])

async function loadRecent() {
  try {
    const [files, folders] = await Promise.all([GetRecentFiles(), GetRecentFolders()])
    recentFiles.value = files || []
    recentFolders.value = folders || []
  } catch { /* 忽略 */ }
}

async function openFile() {
  try {
    const path = await OpenFileDialog()
    if (path) {
      await openFilePath(path)
    }
  } catch (error) {
    console.error('Failed to open file:', error)
  }
}

async function openFilePath(path: string) {
  try {
    const ext = getFileExtension(path)
    const viewType = getTabViewType(ext)
    if (viewType !== 'code') {
      editorStore.createTab(path, '', 'binary', 'LF')
    } else {
      const result = await ReadFile(path)
      if (result) {
        editorStore.createTab(path, result.content, result.info.encoding, result.info.lineEnding)
      }
    }
    await AddRecentEntry(path, false)
  } catch (error) {
    console.error('Failed to open file:', error)
  }
}

async function openRecentFile(entry: RecentEntry) {
  await openFilePath(entry.path)
  await loadRecent()
}

async function openRecentFolder(entry: RecentEntry) {
  try {
    fileStore.setDirectory(entry.path)
    await AddRecentEntry(entry.path, true)
    await loadRecent()
  } catch (error) {
    console.error('Failed to open folder:', error)
  }
}

async function openFolder() {
  try {
    const path = await OpenDirectoryDialog()
    if (path) {
      fileStore.setDirectory(path)
      await AddRecentEntry(path, true)
    }
  } catch (error) {
    console.error('Failed to open folder:', error)
  }
}

async function clearRecentFiles() {
  await ClearRecentFiles()
  recentFiles.value = []
}

async function clearRecentFolders() {
  await ClearRecentFolders()
  recentFolders.value = []
}

onMounted(() => {
  loadRecent()
})
</script>

<template>
  <div class="welcome">
    <!-- 氛围层：顶部青色辉光 + 点阵网格（纯 CSS，自适应深浅色） -->
    <div class="welcome-atmo" aria-hidden="true"></div>

    <!-- Logo -->
    <div class="welcome-logo">
      <div class="welcome-logo-mark">
        <FileText :size="30" :stroke-width="1.6" />
      </div>
    </div>

    <!-- Title -->
    <h1 class="welcome-title">EasyText</h1>
    <p class="welcome-sub">轻量 · 快速 · 为效率而生的桌面文本编辑器</p>

    <!-- Quick actions -->
    <div class="welcome-actions">
      <button class="qa" @click="openFile">
        <span class="qa-icon"><FileText :size="18" :stroke-width="1.6" /></span>
        <span class="qa-text">
          <span class="qa-label">打开文件</span>
          <span class="qa-key">Ctrl+O</span>
        </span>
      </button>

      <button class="qa" @click="openFolder">
        <span class="qa-icon"><FolderOpen :size="18" :stroke-width="1.6" /></span>
        <span class="qa-text">
          <span class="qa-label">打开文件夹</span>
          <span class="qa-key">Ctrl+Shift+O</span>
        </span>
      </button>

      <button class="qa" @click="emit('open-converter')">
        <span class="qa-icon"><FileJson :size="18" :stroke-width="1.6" /></span>
        <span class="qa-text">
          <span class="qa-label">格式转换</span>
          <span class="qa-key">工具 → 格式转换</span>
        </span>
      </button>

      <button class="qa" @click="emit('open-diff')">
        <span class="qa-icon"><Diff :size="18" :stroke-width="1.6" /></span>
        <span class="qa-text">
          <span class="qa-label">文档对比</span>
          <span class="qa-key">工具 → 文档对比</span>
        </span>
      </button>
    </div>

    <!-- 最近访问 -->
    <div v-if="recentFiles.length > 0 || recentFolders.length > 0" class="welcome-recent">
      <!-- 最近文件 -->
      <div v-if="recentFiles.length > 0" class="recent-block">
        <div class="recent-head">
          <div class="recent-title">
            <Clock :size="14" />
            <span>最近打开的文件</span>
          </div>
          <button class="recent-clear" title="清除" @click="clearRecentFiles">
            <X :size="12" />
          </button>
        </div>
        <div class="recent-list">
          <div
            v-for="entry in recentFiles.slice(0, 10)" :key="entry.path"
            class="recent-item"
            :title="entry.path"
            @click="openRecentFile(entry)"
          >
            <FileText :size="14" class="recent-icon" />
            <span class="recent-name">{{ entry.name }}</span>
            <span class="recent-path">{{ entry.path }}</span>
          </div>
        </div>
      </div>

      <!-- 最近文件夹 -->
      <div v-if="recentFolders.length > 0" class="recent-block">
        <div class="recent-head">
          <div class="recent-title">
            <FolderOpen :size="14" />
            <span>最近打开的文件夹</span>
          </div>
          <button class="recent-clear" title="清除" @click="clearRecentFolders">
            <X :size="12" />
          </button>
        </div>
        <div class="recent-list">
          <div
            v-for="entry in recentFolders.slice(0, 10)" :key="entry.path"
            class="recent-item"
            :title="entry.path"
            @click="openRecentFolder(entry)"
          >
            <FolderOpen :size="14" class="recent-icon recent-icon--folder" />
            <span class="recent-name">{{ entry.name }}</span>
            <span class="recent-path">{{ entry.path }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="welcome-footer">版本 2.0.0 · MIT License</div>
  </div>
</template>

<style scoped>
/* 首屏此前整块使用 Tailwind 的 gray-* / blue-* 与 dark: 变体，
   与其余界面的语义令牌不同源，切换主题时观感割裂；这里统一到 --et-*。 */
.welcome {
  position: relative;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: var(--et-bg);
  color: var(--et-fg);
  padding: 24px;
  overflow-y: auto;
  overflow-x: hidden;
}

/* 氛围层：顶部主色辉光 + 渐隐点阵网格，给空白首屏纵深与品牌感 */
.welcome-atmo {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    radial-gradient(900px 380px at 50% -8%, var(--et-accent-soft), transparent 65%);
}
.welcome-atmo::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image: radial-gradient(var(--et-border-strong) 1px, transparent 1px);
  background-size: 26px 26px;
  opacity: .22;
  -webkit-mask-image: radial-gradient(ellipse 70% 55% at 50% 0%, #000, transparent);
  mask-image: radial-gradient(ellipse 70% 55% at 50% 0%, #000, transparent);
}

.welcome > * { position: relative; }

.welcome-logo { margin-bottom: 18px; }
.welcome-logo-mark {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: 18px;
  color: var(--et-accent-contrast);
  background: linear-gradient(145deg, var(--et-accent), var(--et-accent-hover));
  box-shadow: var(--et-shadow-md);
}
.welcome-title {
  margin: 0 0 4px;
  font-size: 26px;
  font-weight: var(--et-fw-semibold);
  letter-spacing: .02em;
}
.welcome-sub { margin: 0 0 30px; font-size: var(--et-text-md); color: var(--et-fg-muted); }

.welcome-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  width: 100%;
  max-width: 400px;
}
.qa {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius);
  background: var(--et-bg-elevated);
  cursor: pointer;
  text-align: left;
  transition: border-color .12s var(--et-ease), box-shadow .12s var(--et-ease), transform .12s var(--et-ease);
}
.qa:hover {
  border-color: var(--et-accent);
  box-shadow: var(--et-shadow-sm);
  transform: translateY(-1px);
}
.qa-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 9px;
  flex-shrink: 0;
  color: var(--et-accent);
  background: var(--et-accent-soft);
}
.qa-text { display: flex; flex-direction: column; gap: 1px; min-width: 0; }
.qa-label { font-size: var(--et-text-md); color: var(--et-fg); font-weight: var(--et-fw-medium); }
.qa-key { font-size: var(--et-text-xs); color: var(--et-fg-subtle); font-variant-numeric: tabular-nums; }

.welcome-recent { margin-top: 30px; width: 100%; max-width: 520px; }
.recent-block { margin-bottom: 14px; }
.recent-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 4px; }
.recent-title {
  display: flex; align-items: center; gap: 5px;
  font-size: var(--et-text-sm); color: var(--et-fg-subtle);
  font-weight: var(--et-fw-medium);
  text-transform: uppercase;
  letter-spacing: .05em;
}
.recent-clear {
  display: inline-flex; padding: 2px; border: none; border-radius: var(--et-radius-sm);
  background: transparent; color: var(--et-fg-subtle); cursor: pointer;
}
.recent-clear:hover { background: var(--et-bg-hover); color: var(--et-danger); }

.recent-list { display: flex; flex-direction: column; }
.recent-item {
  display: flex; align-items: center; gap: 8px;
  padding: 5px 8px; border-radius: var(--et-radius-sm);
  cursor: pointer; min-width: 0;
  transition: background-color 80ms var(--et-ease);
}
.recent-item:hover { background: var(--et-bg-hover); }
.recent-icon { flex-shrink: 0; color: var(--et-accent); opacity: .75; }
.recent-icon--folder { color: var(--et-warn); }
.recent-name {
  font-size: var(--et-text-sm); color: var(--et-fg);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis; flex-shrink: 0;
  max-width: 45%;
}
.recent-path {
  font-size: var(--et-text-xs); color: var(--et-fg-subtle);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis; flex: 1;
}

.welcome-footer { margin-top: 28px; font-size: var(--et-text-xs); color: var(--et-fg-subtle); }
</style>