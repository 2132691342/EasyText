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
    <!-- Logo -->
    <div class="welcome-logo">
      <FileText :size="56" :stroke-width="1.2" />
    </div>

    <!-- Title -->
    <h1 class="welcome-title">EasyText</h1>
    <p class="welcome-sub">轻量级桌面文档编辑工具</p>

    <!-- Quick actions -->
    <div class="welcome-actions">
      <button class="qa" @click="openFile">
        <FileText :size="26" :stroke-width="1.5" />
        <span class="qa-label">打开文件</span>
        <span class="qa-key">Ctrl+O</span>
      </button>

      <button class="qa" @click="openFolder">
        <FolderOpen :size="26" :stroke-width="1.5" />
        <span class="qa-label">打开文件夹</span>
        <span class="qa-key">Ctrl+Shift+O</span>
      </button>

      <button class="qa" @click="emit('open-converter')">
        <FileJson :size="26" :stroke-width="1.5" />
        <span class="qa-label">格式转换</span>
        <span class="qa-key">工具 → 格式转换</span>
      </button>

      <button class="qa" @click="emit('open-diff')">
        <Diff :size="26" :stroke-width="1.5" />
        <span class="qa-label">文档对比</span>
        <span class="qa-key">工具 → 文档对比</span>
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
    <div class="welcome-footer">版本 2.0.0 | MIT License</div>
  </div>
</template>

<style scoped>
/* 首屏此前整块使用 Tailwind 的 gray-* / blue-* 与 dark: 变体，
   与其余界面的语义令牌不同源，切换主题时观感割裂；这里统一到 --et-*。 */
.welcome {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: var(--et-bg);
  color: var(--et-fg);
  padding: 24px;
  overflow-y: auto;
}

.welcome-logo { margin-bottom: 20px; color: var(--et-fg-subtle); }
.welcome-title { margin: 0 0 4px; font-size: 24px; font-weight: 600; letter-spacing: .5px; }
.welcome-sub { margin: 0 0 28px; font-size: 13px; color: var(--et-fg-muted); }

.welcome-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  width: 100%;
  max-width: 380px;
}
.qa {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 16px 12px;
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius);
  background: var(--et-bg-elevated);
  color: var(--et-fg-muted);
  cursor: pointer;
  transition: background .14s ease, border-color .14s ease, color .14s ease;
}
.qa:hover {
  border-color: var(--et-accent);
  background: var(--et-accent-soft);
  color: var(--et-accent);
}
.qa-label { font-size: 13px; color: var(--et-fg); }
.qa-key { font-size: 11px; color: var(--et-fg-subtle); }

.welcome-recent { margin-top: 28px; width: 100%; max-width: 520px; }
.recent-block { margin-bottom: 14px; }
.recent-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 6px; }
.recent-title { display: flex; align-items: center; gap: 5px; font-size: 12px; color: var(--et-fg-subtle); }
.recent-clear {
  display: inline-flex; padding: 2px; border: none; border-radius: var(--et-radius-sm);
  background: transparent; color: var(--et-fg-subtle); cursor: pointer;
}
.recent-clear:hover { background: var(--et-bg-hover); color: #ef4444; }

.recent-list { display: flex; flex-direction: column; }
.recent-item {
  display: flex; align-items: center; gap: 8px;
  padding: 4px 8px; border-radius: var(--et-radius-sm);
  cursor: pointer; min-width: 0;
}
.recent-item:hover { background: var(--et-bg-hover); }
.recent-icon { flex-shrink: 0; color: var(--et-fg-subtle); }
.recent-icon--folder { color: #eab308; }
.recent-name {
  font-size: 12px; color: var(--et-fg);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis; flex-shrink: 0;
  max-width: 45%;
}
.recent-path {
  font-size: 11px; color: var(--et-fg-subtle);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis; flex: 1;
}

.welcome-footer { margin-top: 28px; font-size: 11px; color: var(--et-fg-subtle); }
</style>