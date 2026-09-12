/**
 * ext-keymap — CodeMirror keymap 装配 + 命令别名
 *
 * 职责（与原 CodeEditor.vue 完全等价搬运，无行为变更）：
 *  - precKeymap：默认 keymap + history + fold + completion + lint + indentWithTab + closeBrackets
 *  - CMD_ALIASES：菜单 / 旧命令名 → 实际命令名的映射
 *
 * 注意：标点键（PunctuationKeymap）已被显式移除——原注释明确说过 CodeMirror 6
 * 已正确处理标点符号，旧实现拦截会与中文 IME 冲突。
 */
import { keymap } from '@codemirror/view'
import type { Extension } from '@codemirror/state'
import { defaultKeymap, history, historyKeymap, indentWithTab } from '@codemirror/commands'
import { closeBrackets, closeBracketsKeymap } from '@codemirror/autocomplete'
import { foldKeymap } from '@codemirror/language'
import { completionKeymap } from '@codemirror/autocomplete'
import { lintKeymap } from '@codemirror/lint'

/** 命令别名：菜单/快捷键发出的命令 → 编辑器内部的实际命令 */
export const CMD_ALIASES: Record<string, string> = {
  // 行操作
  'line-dup': 'line-duplicate',
  'line-del': 'line-remove',
  'line-delete': 'line-remove',
  'line-rmdup': 'line-removeDuplicate',
  'line-up': 'line-moveUp',
  'line-down': 'line-moveDown',
  'line-move-up': 'line-moveUp',
  'line-move-down': 'line-moveDown',
  'line-rmempty': 'line-removeEmpty',
  'line-rmblank': 'line-removeBlank',
  'line-insert-above': 'line-insertAbove',
  'line-insert-below': 'line-insertBelow',
  // Tab/空格转换
  'tab2space': 'tab-to-spaces',
  'space2tab-all': 'spaces-all-to-tabs',
  'space2tab-lead': 'spaces-leading-to-tabs',
  // 显示空格/Tab、行尾符 → 切换
  'show-spaces': 'toggle-whitespace',
  'show-eol': 'toggle-eol',
}

export interface ExtKeymap {
  /** 装配 keymap 扩展 */
  buildPrecKeymap: () => Extension
}

/**
 * 容器在 createEditor() 时调用一次 buildPrecKeymap() 注入到 EditorState。
 * 不持有可变状态——纯工厂。
 */
export function useEditorKeymap(): ExtKeymap {
  function buildPrecKeymap(): Extension {
    return [
      history(),
      keymap.of([
        ...closeBracketsKeymap,
        ...defaultKeymap,
        ...historyKeymap,
        ...foldKeymap,
        ...completionKeymap,
        ...lintKeymap,
        indentWithTab,
      ]),
      closeBrackets(),
    ]
  }
  return { buildPrecKeymap }
}