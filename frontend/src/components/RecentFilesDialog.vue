<script lang="ts" setup>
/**
 * 最近文件 / 文件夹 v2.1 — ModalOverlay 统一
 */
import { ref, watch, onMounted } from 'vue'
import { GetRecentFiles, GetRecentFolders, OpenFileDialog, ClearRecentFiles, ClearRecentFolders, AddRecentEntry } from '../../wailsjs/go/main/App'
import { FileText, FolderOpen, FileSearch, FolderSearch, X } from 'lucide-vue-next'
import { ElMessage } from 'element-plus'
import type { RecentEntry } from '@/types'
import ModalOverlay from './ModalOverlay.vue'

const props = defineProps<{ visible: boolean; initialTab?: 'files' | 'folders' }>()
const emit = defineEmits<{
  (e: 'close'): void
  (e: 'open-file', path: string): void
  (e: 'open-folder', path: string): void
}>()

const activeTab = ref<'files' | 'folders'>(props.initialTab || 'files')
const recentFiles = ref<RecentEntry[]>([])
const recentFolders = ref<RecentEntry[]>([])
const loading = ref(true)

async function loadData() {
  loading.value = true
  try {
    const [files, folders] = await Promise.all([GetRecentFiles(), GetRecentFolders()])
    recentFiles.value = (files || []).slice(0, 50)
    recentFolders.value = (folders || []).slice(0, 50)
  } catch (e: any) {
    ElMessage.error('加载最近记录失败: ' + (e?.message || ''))
  }
  loading.value = false
}

async function openRecentFile(entry: RecentEntry) {
  emit('open-file', entry.path)
  emit('close')
}
async function openRecentFolder(entry: RecentEntry) {
  try {
    await AddRecentEntry(entry.path, true)
    emit('open-folder', entry.path)
    emit('close')
  } catch (e: any) {
    ElMessage.error('打开文件夹失败: ' + (e?.message || ''))
  }
}
async function browseFile() {
  try {
    const path = await OpenFileDialog()
    if (path) {
      emit('open-file', path)
      emit('close')
    }
  } catch (e: any) {
    ElMessage.error('打开文件失败: ' + (e?.message || ''))
  }
}
async function clearRecent(section: 'files' | 'folders') {
  try {
    if (section === 'files') {
      await ClearRecentFiles()
      recentFiles.value = []
    } else {
      await ClearRecentFolders()
      recentFolders.value = []
    }
    document.dispatchEvent(new CustomEvent('recent-updated'))
  } catch (e: any) {
    ElMessage.error('清除失败: ' + (e?.message || ''))
  }
}

watch(() => props.visible, (v) => {
  if (v) {
    activeTab.value = props.initialTab || 'files'
    loadData()
  }
})
onMounted(() => { if (props.visible) loadData() })
</script>

<template>
  <ModalOverlay :visible="visible" title="最近文件" size="md" @close="emit('close')">
    <div class="rf-root">
      <!-- Tab 切换 -->
      <div class="rf-tabs">
        <button
          class="rf-tab"
          :class="{ 'is-active': activeTab === 'files' }"
          @click="activeTab = 'files'"
        >
          <FileText :size="14" :stroke-width="1.6" />
          最近文件
          <span v-if="recentFiles.length" class="rf-count">{{ recentFiles.length }}</span>
        </button>
        <button
          class="rf-tab"
          :class="{ 'is-active': activeTab === 'folders' }"
          @click="activeTab = 'folders'"
        >
          <FolderOpen :size="14" :stroke-width="1.6" />
          最近文件夹
          <span v-if="recentFolders.length" class="rf-count">{{ recentFolders.length }}</span>
        </button>
      </div>

      <!-- 主体 -->
      <div class="rf-body">
        <div v-if="loading" class="et-empty">
          <div class="et-empty-title">加载中…</div>
        </div>

        <!-- 文件列表 -->
        <template v-else-if="activeTab === 'files'">
          <div v-if="recentFiles.length === 0" class="et-empty">
            <FileSearch :size="32" :stroke-width="1.6" class="et-empty-icon" />
            <div class="et-empty-title">暂无最近打开的文件</div>
            <button class="et-btn-sm" @click="browseFile">浏览文件…</button>
          </div>
          <div v-else class="rf-list">
            <button
              v-for="entry in recentFiles" :key="entry.path"
              class="rf-item"
              :title="entry.path"
              @click="openRecentFile(entry)"
            >
              <FileText class="rf-item-icon" :size="14" :stroke-width="1.6" />
              <div class="rf-item-text">
                <div class="rf-item-name">{{ entry.name }}</div>
                <div class="rf-item-path">{{ entry.path }}</div>
              </div>
            </button>
          </div>
          <div v-if="recentFiles.length > 0" class="rf-footer">
            <button class="rf-clear" @click="clearRecent('files')">
              <X :size="12" :stroke-width="1.6" /> 清除记录
            </button>
          </div>
        </template>

        <!-- 文件夹列表 -->
        <template v-else>
          <div v-if="recentFolders.length === 0" class="et-empty">
            <FolderSearch :size="32" :stroke-width="1.6" class="et-empty-icon" />
            <div class="et-empty-title">暂无最近打开的文件夹</div>
          </div>
          <div v-else class="rf-list">
            <button
              v-for="entry in recentFolders" :key="entry.path"
              class="rf-item"
              :title="entry.path"
              @click="openRecentFolder(entry)"
            >
              <FolderOpen class="rf-item-icon rf-item-icon-warn" :size="14" :stroke-width="1.6" />
              <div class="rf-item-text">
                <div class="rf-item-name">{{ entry.name }}</div>
                <div class="rf-item-path">{{ entry.path }}</div>
              </div>
            </button>
          </div>
          <div v-if="recentFolders.length > 0" class="rf-footer">
            <button class="rf-clear" @click="clearRecent('folders')">
              <X :size="12" :stroke-width="1.6" /> 清除记录
            </button>
          </div>
        </template>
      </div>
    </div>
  </ModalOverlay>
</template>

<style scoped>
.rf-root {
  display: flex;
  flex-direction: column;
  min-height: 360px;
}
.rf-tabs {
  display: flex;
  border-bottom: 1px solid var(--et-border);
  padding: 0 var(--et-space-3);
  margin-bottom: var(--et-space-3);
  gap: var(--et-space-2);
}
.rf-tab {
  display: inline-flex;
  align-items: center;
  gap: var(--et-space-1);
  padding: var(--et-space-2) var(--et-space-3);
  background: transparent;
  border: 0;
  border-bottom: 2px solid transparent;
  color: var(--et-fg-muted);
  font-size: var(--et-text-sm);
  font-weight: var(--et-fw-medium);
  cursor: pointer;
  transition: color 80ms ease, border-color 80ms ease;
}
.rf-tab:hover { color: var(--et-fg); }
.rf-tab.is-active {
  color: var(--et-accent);
  border-bottom-color: var(--et-accent);
}
.rf-count {
  font-size: var(--et-text-xs);
  padding: 1px 6px;
  background: var(--et-bg-active);
  border-radius: var(--et-radius-sm);
  color: var(--et-fg-muted);
}

.rf-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
.rf-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.rf-item {
  display: flex;
  align-items: center;
  gap: var(--et-space-2);
  padding: var(--et-space-2) var(--et-space-3);
  background: transparent;
  border: 0;
  text-align: left;
  cursor: pointer;
  border-radius: var(--et-radius-sm);
  transition: background-color 80ms ease;
}
.rf-item:hover { background: var(--et-accent-soft); }
.rf-item-icon { color: var(--et-fg-subtle); flex-shrink: 0; }
.rf-item-icon-warn { color: var(--et-warn); }
.rf-item-text { flex: 1; min-width: 0; }
.rf-item-name {
  font-size: var(--et-text-md);
  color: var(--et-fg);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.rf-item-path {
  font-size: var(--et-text-xs);
  color: var(--et-fg-subtle);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rf-footer {
  border-top: 1px solid var(--et-border);
  padding: var(--et-space-2) 0 0;
  margin-top: var(--et-space-2);
}
.rf-clear {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  background: transparent;
  border: 0;
  padding: 4px 6px;
  font-size: var(--et-text-xs);
  color: var(--et-fg-subtle);
  cursor: pointer;
  border-radius: var(--et-radius-sm);
  transition: color 80ms ease;
}
.rf-clear:hover { color: var(--et-danger); }
</style>