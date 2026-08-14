<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useEditorStore } from '@/stores'
import { StartFileWatch, StopFileWatch } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { ElMessage, ElSwitch } from 'element-plus'
import { Activity, FileText } from 'lucide-vue-next'

const editorStore = useEditorStore()

const monitored = ref<Set<string>>(new Set())
const events = ref<{ path: string; type: string; time: string }[]>([])

let eventsOff: (() => void) | null = null

async function toggleMonitor(path: string) {
  if (!path) return
  if (monitored.value.has(path)) {
    try { await StopFileWatch(path) } catch { /* ignore */ }
    monitored.value = new Set([...monitored.value].filter(p => p !== path))
  } else {
    try {
      await StartFileWatch(path)
      monitored.value = new Set([...monitored.value, path])
      ElMessage.success('已开始监控')
    } catch { ElMessage.error('监控失败') }
  }
}

function isMonitored(path: string): boolean {
  return monitored.value.has(path)
}

onMounted(() => {
  eventsOff = EventsOn('file:change', (evt: any) => {
    events.value.unshift({
      path: evt?.path || '',
      type: evt?.type || 'modify',
      time: new Date().toLocaleTimeString(),
    })
    if (events.value.length > 200) events.value.length = 200
  })
})

onUnmounted(() => {
  for (const p of monitored.value) { StopFileWatch(p).catch(() => {}) }
  if (eventsOff) eventsOff()
})
</script>

<template>
  <div class="h-full flex flex-col bg-white dark:bg-[#252526] text-[var(--theme-fg)]">
    <div class="flex items-center justify-between px-3 py-2 border-b border-gray-200 dark:border-gray-700">
      <span class="text-sm font-medium text-gray-700 dark:text-gray-200">文件监控</span>
      <span class="text-xs text-gray-400">{{ monitored.size }} 监控中</span>
    </div>

    <!-- 打开的文件列表 -->
    <div class="border-b border-gray-200 dark:border-gray-700 max-h-[45%] overflow-auto">
      <div class="px-3 py-1 text-[11px] text-gray-400">打开的文件</div>
      <div v-if="!editorStore.tabs.filter(t => t.path).length" class="px-3 pb-2 text-xs text-gray-400">无已打开文件</div>
      <div
        v-for="t in editorStore.tabs.filter(t => t.path)"
        :key="t.id"
        class="flex items-center gap-2 px-3 py-1 text-[13px]"
      >
        <FileText class="w-3.5 h-3.5 text-gray-400 flex-shrink-0" />
        <span class="truncate flex-1" :title="t.path">{{ t.name }}</span>
        <ElSwitch :model-value="isMonitored(t.path)" size="small" @change="toggleMonitor(t.path)" />
      </div>
    </div>

    <!-- 变更事件日志 -->
    <div class="flex-1 overflow-auto">
      <div class="px-3 py-1 text-[11px] text-gray-400 flex items-center gap-1">
        <Activity class="w-3 h-3" /> 变更事件
      </div>
      <div v-if="!events.length" class="px-3 pb-2 text-xs text-gray-400">暂无变更事件</div>
      <div v-for="(e, i) in events" :key="i" class="px-3 py-1 text-[12px] border-b border-gray-100 dark:border-gray-800">
        <span class="text-gray-400 mr-2">{{ e.time }}</span>
        <span class="text-blue-500 mr-2">{{ e.type }}</span>
        <span class="truncate" :title="e.path">{{ e.path }}</span>
      </div>
    </div>
  </div>
</template>
