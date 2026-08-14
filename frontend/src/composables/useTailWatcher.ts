/**
 * tail -f 文件跟踪 composable。
 *
 * 抽离自 MainLayout.vue。封装：
 *   - 当前正在跟踪的标签 ID（reactive）
 *   - 启动/停止监听 Wails 后端 file:change 事件
 *   - **正确清理 EventsOn 监听器**（修复原版重复 start 时泄漏监听器的 bug）
 *
 * 设计要点：
 *   - 后端文件变化事件通过 `file:change` 全局事件广播；多次启动而不取消
 *     EventsOn 会导致每次文件变化触发多次回调——这是经典的监听器泄漏。
 *   - 监听器句柄保存在模块级 closure 中，确保 cancel 总是引用同一函数。
 */

import { ref, type Ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { useEditorStore } from '@/stores/editorStore'

type Ed = ReturnType<typeof useEditorStore>

export interface UseTailWatcherOptions {
  ed: Ed
  /** 当前活动标签变化后自动重新跟踪（可选） */
  execEd: (cmd: string, ...args: unknown[]) => void
}

export interface TailWatcher {
  tailingTabId: Ref<string>
  isTailing: Ref<boolean>
  startTail: () => Promise<void>
  stopTail: () => Promise<void>
}

export function useTailWatcher(opts: UseTailWatcherOptions): TailWatcher {
  const tailingTabId = ref('')
  const isTailing = ref(false)

  // 持有当前活跃的 Wails EventsOn 取消函数。
  // 闭包变量而非模块级 let，避免多次 useTailWatcher() 互相干扰。
  let cancelFileChange: (() => void) | null = null
  let watchedPath: string | null = null

  function clearWatcher(): void {
    if (cancelFileChange) {
      cancelFileChange()
      cancelFileChange = null
    }
    watchedPath = null
  }

  async function startTail(): Promise<void> {
    const t = opts.ed.activeTab
    if (!t?.path) {
      ElMessage.warning('当前文档无文件路径')
      return
    }

    // 先停掉旧的，避免监听器叠加
    await stopTail()

    tailingTabId.value = t.id
    watchedPath = t.path

    try {
      const { StartFileWatch } = await import('../../wailsjs/go/main/App')
      const { EventsOn } = await import('../../wailsjs/runtime/runtime')

      await StartFileWatch(t.path)

      // EventsOn 直接返回取消函数，保存引用以便 stopTail 时调用。
      // 注意：每次 startTail 必须先 stopTail 清掉旧监听器，否则 file:change
      // 事件会触发多次回调（监听器叠加泄漏）。
      const handler = async (...args: unknown[]) => {
        const evt = args[0] as { path?: string } | undefined
        const cur = opts.ed.tabs.find(x => x.id === tailingTabId.value)
        if (!cur || !cur.path || evt?.path !== cur.path) return
        try {
          const { ReadFile } = await import('../../wailsjs/go/main/App')
          const r = await ReadFile(cur.path)
          if (r) {
            opts.ed.updateTabContent(cur.id, r.content)
            setTimeout(() => opts.execEd('scroll-to-end'), 50)
          }
        } catch (err) {
          console.warn(err)
        }
      }

      cancelFileChange = EventsOn('file:change', handler)

      isTailing.value = true
      ElMessage.success('已开始跟踪文件')
    } catch (e: unknown) {
      clearWatcher()
      tailingTabId.value = ''
      ElMessage.error('跟踪失败：' + ((e as Error)?.message || String(e)))
    }
  }

  async function stopTail(): Promise<void> {
    const t = opts.ed.tabs.find(x => x.id === tailingTabId.value)
    tailingTabId.value = ''
    isTailing.value = false

    const pathToStop = watchedPath || t?.path

    clearWatcher()

    if (pathToStop) {
      try {
        const { StopFileWatch } = await import('../../wailsjs/go/main/App')
        await StopFileWatch(pathToStop)
        ElMessage.info('已停止跟踪文件')
      } catch (e) {
        console.warn(e)
      }
    }
  }

  return {
    tailingTabId,
    isTailing,
    startTail,
    stopTail,
  }
}