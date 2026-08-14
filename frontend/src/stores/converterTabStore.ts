import { defineStore } from 'pinia'
import { ref } from 'vue'

/**
 * FormatConverter 子标签路由。
 *
 * useCommands.ts 写入 useFormatConverterStore().set()，
 * FormatConverter.vue 通过 consume() 读取并自动清理。
 * 用 Pinia store 替代原先模块级 let，使状态对其他 reactive watcher 也可观测。
 */

export type FormatConverterTab = 'jsonpath' | 'json-to-struct' | 'json-diff'

export const useFormatConverterStore = defineStore('formatConverterTab', () => {
  const pendingTab = ref<FormatConverterTab | null>(null)

  function set(tab: FormatConverterTab): void {
    pendingTab.value = tab
  }

  /** 读取并清空（一次性消费语义）。 */
  function consume(): FormatConverterTab | null {
    const t = pendingTab.value
    pendingTab.value = null
    return t
  }

  return { pendingTab, set, consume }
})