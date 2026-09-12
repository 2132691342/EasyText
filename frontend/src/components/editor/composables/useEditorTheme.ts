/**
 * useEditorTheme — 编辑器外观（主题）
 *
 * 职责（与原 CodeEditor.vue 完全等价搬运，无行为变更）：
 *  - editorFontSizePx()  ：配置字号 × 界面缩放
 *  - buildAppearanceTheme()：EditorView.theme 整体配色
 *  - buildSyntaxHighlight()：基于主题色板的 syntaxHighlighting 扩展
 *  - buildFoldGutter()    ：折叠图标（▾/▸）DOM
 *
 * 设计目标：纯函数 + 接收 store 提供的 `colors` / `config`。
 * useEditorTheme 不持有任何可变状态——它只产出 CodeMirror extension。
 */
import { computed, type ComputedRef } from 'vue'
import type { Extension } from '@codemirror/view'
import { EditorView } from '@codemirror/view'
import { syntaxHighlighting, HighlightStyle, foldGutter } from '@codemirror/language'
import { tags } from '@lezer/highlight'

/** useSettingStore.currentThemeColors 的结构（仅引用所需字段） */
export interface ThemeColors {
  bg: string
  fg: string
  gutterBg: string
  gutterFg: string
  activeLine: string
  cursor: string
  bracketMatch: string
  selection: string
  accent: string
  comment: string
  keyword: string
  string: string
  number: string
  type: string
  function: string
  variable: string
  operator: string
  isDark: boolean
}

/** AppConfig 字段（仅引用所需字段） */
export interface ThemeConfigSlice {
  editor?: {
    fontSize?: number
    fontFamily?: string
  }
  ui?: {
    zoomLevel?: number
  }
}

export interface UseEditorTheme {
  /** 构建 EditorView.theme 配置 */
  buildAppearanceTheme: () => Extension
  /** 构建 syntaxHighlighting（基于主题色板） */
  buildSyntaxHighlight: () => Extension
  /** 构建 foldGutter（isDark 决定折叠图标颜色） */
  buildFoldGutter: (isDark: boolean) => Extension
  /** 当前编辑器字号（fontSize × zoomLevel） */
  editorFontSizePx: () => number
}

/**
 * 注意：本 composable 不持有可变状态，仅是一组工厂函数。
 * 调用方（容器 / useEditorView）每次需要时即时构建。
 *
 * @param colors 主题色（store.currentThemeColors）
 * @param config AppConfig（store.config）
 */
export function useEditorTheme(
  colors: ComputedRef<ThemeColors>,
  config: ComputedRef<ThemeConfigSlice | null>,
): UseEditorTheme {
  /** 编辑器实际字号 = 配置字号 × 界面缩放 */
  function editorFontSizePx(): number {
    const base = config.value?.editor?.fontSize || 14
    const zoom = config.value?.ui?.zoomLevel || 100
    return Math.max(6, Math.round(base * zoom / 100))
  }

  function buildAppearanceTheme(): Extension {
    const c = colors.value
    return EditorView.theme({
      '&': {
        fontSize: `${editorFontSizePx()}px`,
        fontFamily: config.value?.editor?.fontFamily || 'Consolas, Monaco, "Courier New", monospace',
        backgroundColor: c.bg,
        color: c.fg,
      },
      '.cm-scroller': {
        fontFamily: config.value?.editor?.fontFamily || 'Consolas, Monaco, "Courier New", monospace',
      },
      // user-select 由 style.css 全局设置为 text（配合原生 ::selection 高亮）；
      // 此处只需禁用 WebView2 的元素拖拽（-webkit-user-drag: element 会劫持文本选择）。
      '.cm-content': { WebkitUserDrag: 'none' },
      '.cm-gutters': {
        backgroundColor: c.gutterBg,
        color: c.gutterFg,
        borderRight: `1px solid ${c.isDark ? '#404040' : '#e5e7eb'}`,
      },
      '.cm-activeLineGutter': { backgroundColor: c.activeLine },
      '.cm-activeLine': { backgroundColor: c.activeLine },
      '.cm-cursor, &.cm-focused .cm-cursor': { borderLeftColor: c.cursor },
      '.cm-matchingBracket': {
        backgroundColor: c.bracketMatch,
        outline: '1px solid ' + c.accent,
      },
      '&.cm-focused .cm-matchingBracket': { backgroundColor: c.bracketMatch },
      // 🆕 V2.0.0 标签配对高亮
      '.cm-matchingTag': {
        backgroundColor: c.bracketMatch,
        outline: '1px solid ' + c.accent,
      },
      '.cm-foldPlaceholder': {
        backgroundColor: c.isDark ? '#3c3c3c' : '#e5e7eb',
        color: c.fg,
        border: 'none',
      },
      // 原生 ::selection 高亮选中文本。
      // WebView2 下 CodeMirror 的 drawSelection（DOM 层选区）不可靠，会导致"能选中但无背景色"，
      // 故移除 drawSelection，改用原生 ::selection（配合 style.css 中的主题色变量）。
      '& .cm-content ::selection': {
        backgroundColor: c.selection,
        color: 'inherit',
      },
    })
  }

  function buildSyntaxHighlight(): Extension {
    const c = colors.value
    return syntaxHighlighting(HighlightStyle.define([
      { tag: [tags.comment, tags.lineComment, tags.blockComment, tags.docComment], color: c.comment },
      {
        tag: [tags.keyword, tags.controlKeyword, tags.moduleKeyword, tags.operatorKeyword,
              tags.definitionKeyword, tags.modifier, tags.self, tags.bool, tags.null, tags.atom],
        color: c.keyword,
      },
      {
        tag: [tags.string, tags.docString, tags.character, tags.attributeValue,
              tags.regexp, tags.escape, tags.color, tags.url],
        color: c.string,
      },
      { tag: [tags.number, tags.integer, tags.float], color: c.number },
      { tag: [tags.typeName, tags.className, tags.namespace, tags.tagName], color: c.type },
      {
        tag: [tags.function(tags.variableName), tags.function(tags.propertyName), tags.macroName],
        color: c.function,
      },
      {
        tag: [tags.variableName, tags.propertyName, tags.attributeName, tags.labelName],
        color: c.variable,
      },
      {
        tag: [tags.operator, tags.derefOperator, tags.arithmeticOperator, tags.logicOperator,
              tags.bitwiseOperator, tags.compareOperator, tags.updateOperator,
              tags.definitionOperator, tags.typeOperator, tags.controlOperator],
        color: c.operator,
      },
      { tag: [tags.definition(tags.variableName), tags.definition(tags.propertyName)], color: c.function },
    ]), { fallback: true })
  }

  function buildFoldGutter(isDark: boolean): Extension {
    return foldGutter({
      markerDOM(open) {
        const span = document.createElement('span')
        span.style.cssText = `display:inline-flex;align-items:center;justify-content:center;width:14px;height:14px;cursor:pointer;font-size:11px;color:${isDark ? '#858585' : '#999'};`
        span.textContent = open ? '▾' : '▸'
        return span
      },
    })
  }

  return {
    buildAppearanceTheme,
    buildSyntaxHighlight,
    buildFoldGutter,
    editorFontSizePx,
  }
}