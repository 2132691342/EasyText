/**
 * useEditorColumnMode — 列编辑模式（Alt+X）
 *
 * 职责（与原 CodeEditor.vue 完全等价搬运，无行为变更）：
 *  - enterColumnMode：派发 show-column-edit CustomEvent → MainLayout 打开 ColumnEditWin
 *  - toggleColumnMode：翻转 settingStore.config.editor.columnMode，
 *    rectangularSelection/crosshairCursor 已在基础扩展中，
 *    这里只切十字光标（CSS 由 setting  驱动）
 *
 * 设计目标：薄薄一层，几乎只有事件派发 + config toggle。
 */
import { useSettingStore } from '@/stores'

export interface UseEditorColumnMode {
  enterColumnMode: () => void
  toggleColumnMode: () => void
  isEnabled: () => boolean
}

export function useEditorColumnMode(): UseEditorColumnMode {
  const settingStore = useSettingStore()

  function enterColumnMode() {
    document.dispatchEvent(new CustomEvent('show-column-edit'))
  }

  function toggleColumnMode() {
    const on = !(settingStore.config?.editor?.columnMode)
    if (settingStore.config) {
      settingStore.config.editor.columnMode = on
      settingStore.saveConfig()
    }
  }

  function isEnabled() {
    return !!settingStore.config?.editor?.columnMode
  }

  return { enterColumnMode, toggleColumnMode, isEnabled }
}