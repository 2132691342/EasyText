/**
 * useEditorMacro — 宏录制 / 回放（保持与原 CodeEditor 一致）
 *
 * 职责（与原 CodeEditor.vue 完全等价搬运，无行为变更）：
 *  - recordMacroStep(update, view, isInitializing, editorStore)：在 updateListener 内
 *    同步 documentChanged 时把 insert / delete 步骤追加到 editorStore.macroState.currentMacro。
 *  - playMacro(view, editorStore)：当 isPlaying 为 true 时按顺序回放步骤（基于当前光标位置）。
 *
 * 设计目标：宏录制与回放完全依赖 editorStore.macroState，本模块不持有额外状态。
 *
 * 注意：宏 step schema 在 types/index.ts 中已支持 selection / cursor / command / find 等类型，
 * 但当前录制仅产出 insert / delete（与原 CodeEditor 行为完全一致）；
 * 扩展 schema 需要后端宏 / store 同步变更（不在本轮范围）。
 */
import type { EditorView } from '@codemirror/view'
import { useEditorStore } from '@/stores'

export interface UseEditorMacro {
  /**
   * 由 useEditorView 的 EditorView.updateListener 回调
   * （容器需要把 update / view / isInitializing 透传过来）
   */
  recordMacroStep: (update: any, view: EditorView, isInitializing: boolean) => void
  /**
   * 由容器 watch(isPlaying) 时调用
   */
  playMacro: (view: EditorView | null) => void
}

export function useEditorMacro(): UseEditorMacro {
  const editorStore = useEditorStore()

  function recordMacroStep(update: any, view: EditorView, isInitializing: boolean) {
    if (!update.docChanged || isInitializing) return
    if (!editorStore.macroState.isRecording) return
    for (const tr of update.transactions) {
      if (!tr.docChanged) continue
      let inserted = '', deleted = ''
      tr.changes.iterChanges((fromA: number, toA: number, _fromB: number, _toB: number, inserted_chunk: any) => {
        if (fromA !== toA) deleted += update.startState.sliceDoc(fromA, toA)
        if (_fromB !== _toB) inserted += inserted_chunk.toString()
      })
      if (inserted) {
        const pos = update.state.selection.main.head
        editorStore.recordMacroStep({
          type: 'insert',
          text: inserted,
          timestamp: Date.now(),
          from: pos,
          to: pos + inserted.length,
        })
      }
      if (deleted) {
        editorStore.recordMacroStep({
          type: 'delete',
          text: deleted,
          timestamp: Date.now(),
        })
      }
    }
  }

  function playMacro(view: EditorView | null) {
    if (!view) return
    const run = () => {
      if (!editorStore.macroState.isPlaying || !view) return
      const step = editorStore.getNextMacroStep()
      if (!step) return
      const head = view.state.selection.main.head
      switch (step.type) {
        case 'insert':
          if (step.text) view.dispatch({ changes: { from: head, insert: step.text } })
          break
        case 'delete': {
          // 删除当前光标前 text.length 个字符（模拟退格删除）
          const len = step.text?.length || (step.to && step.from ? step.to - step.from : 0)
          if (len > 0) view.dispatch({ changes: { from: Math.max(0, head - len), to: head } })
          break
        }
        default:
          // selection / cursor / command / find 在 schema 中保留，
          // 但当前录制仅产出 insert / delete，故 default 不做任何事（与原行为一致）。
          break
      }
      if (editorStore.macroState.isPlaying) requestAnimationFrame(run)
    }
    requestAnimationFrame(run)
  }

  return { recordMacroStep, playMacro }
}