<script lang="ts" setup>
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import { useEditorStore, useSettingStore } from '@/stores'
import type { EditorTab } from '@/types'
import { EditorView, basicSetup } from 'codemirror'
import { EditorState } from '@codemirror/state'
import { keymap } from '@codemirror/view'
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { searchKeymap, highlightSelectionMatches } from '@codemirror/search'
import { closeBrackets, autocompletion, closeBracketsKeymap, completionKeymap } from '@codemirror/autocomplete'
import { lintKeymap } from '@codemirror/lint'
import { bracketMatching, indentOnInput, syntaxHighlighting, defaultHighlightStyle, foldGutter, foldKeymap } from '@codemirror/language'
import { lineNumbers, highlightActiveLineGutter, highlightSpecialChars, drawSelection, dropCursor, rectangularSelection, crosshairCursor, highlightActiveLine } from '@codemirror/view'

// Language support
import { json } from '@codemirror/lang-json'
import { javascript } from '@codemirror/lang-javascript'
import { html } from '@codemirror/lang-html'
import { css } from '@codemirror/lang-css'
import { markdown } from '@codemirror/lang-markdown'
import { xml } from '@codemirror/lang-xml'
import { yaml } from '@codemirror/lang-yaml'
import { python } from '@codemirror/lang-python'
import { java } from '@codemirror/lang-java'
import { go } from '@codemirror/lang-go'
import { sql } from '@codemirror/lang-sql'

const props = defineProps<{
  tab: EditorTab
}>()

const editorStore = useEditorStore()
const settingStore = useSettingStore()

const editorContainer = ref<HTMLElement | null>(null)
let editorView: EditorView | null = null

const config = computed(() => settingStore.config)

// Get language extension
function getLanguageExtension(lang: string) {
  const langMap: Record<string, any> = {
    'json': json(),
    'javascript': javascript(),
    'typescript': javascript({ typescript: true }),
    'html': html(),
    'css': css(),
    'markdown': markdown(),
    'xml': xml(),
    'yaml': yaml(),
    'python': python(),
    'java': java(),
    'go': go(),
    'sql': sql(),
  }
  return langMap[lang] || []
}

// Create editor
function createEditor() {
  if (!editorContainer.value) return

  // Destroy existing editor
  if (editorView) {
    editorView.destroy()
  }

  const langExtension = getLanguageExtension(props.tab.language)

  // Editor theme based on dark mode
  const isDark = settingStore.isDarkMode

  // Create state
  const state = EditorState.create({
    doc: props.tab.content,
    extensions: [
      basicSetup,
      lineNumbers(),
      highlightActiveLineGutter(),
      highlightSpecialChars(),
      history(),
      foldGutter(),
      drawSelection(),
      dropCursor(),
      EditorState.allowMultipleSelections.of(true),
      indentOnInput(),
      syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
      bracketMatching(),
      closeBrackets(),
      autocompletion(),
      rectangularSelection(),
      crosshairCursor(),
      highlightActiveLine(),
      highlightSelectionMatches(),
      keymap.of([
        ...defaultKeymap,
        ...searchKeymap,
        ...historyKeymap,
        ...foldKeymap,
        ...completionKeymap,
        ...lintKeymap,
        ...closeBracketsKeymap,
      ]),
      langExtension,
      // Tab size
      EditorState.tabSize.of(config.value?.editor.tabSize || 4),
      // Line wrapping
      config.value?.editor.wordWrap ? EditorView.lineWrapping : [],
      // Font settings
      EditorView.theme({
        '&': {
          fontSize: `${config.value?.editor.fontSize || 14}px`,
          fontFamily: config.value?.editor.fontFamily || 'Consolas, Monaco, Courier New, monospace',
        },
        '.cm-scroller': {
          fontFamily: config.value?.editor.fontFamily || 'Consolas, Monaco, Courier New, monospace',
        },
      }),
      // Dark theme
      EditorView.theme({
        '&': isDark ? {
          backgroundColor: '#1e1e1e',
          color: '#d4d4d4',
        } : {},
        '.cm-gutters': isDark ? {
          backgroundColor: '#252526',
          color: '#858585',
          borderRight: '1px solid #404040',
        } : {},
        '.cm-activeLine': isDark ? {
          backgroundColor: '#2d2d2d',
        } : {},
        '.cm-selectionBackground': isDark ? {
          backgroundColor: '#264f78',
        } : {},
      }),
      // Update listener
      EditorView.updateListener.of((update) => {
        if (update.docChanged) {
          const newContent = update.state.doc.toString()
          editorStore.updateTabContent(props.tab.id, newContent)
        }
        // Track cursor position
        if (update.selectionSet) {
          const pos = update.state.selection.main.head
          const line = update.state.doc.lineAt(pos)
          editorStore.updateCursorPosition(props.tab.id, line.number, pos - line.from + 1)
        }
      }),
    ],
  })

  // Create view
  editorView = new EditorView({
    state,
    parent: editorContainer.value,
  })

  // Restore cursor position
  if (props.tab.cursorPosition.line > 1) {
    const line = state.doc.line(props.tab.cursorPosition.line)
    editorView.dispatch({
      selection: { anchor: line.from + props.tab.cursorPosition.column - 1 },
      scrollIntoView: true,
    })
  }
}

// Watch for tab changes
watch(() => props.tab.id, () => {
  createEditor()
})

watch(() => props.tab.content, (newContent) => {
  if (editorView && editorView.state.doc.toString() !== newContent) {
    editorView.dispatch({
      changes: { from: 0, to: editorView.state.doc.length, insert: newContent },
    })
  }
})

// Watch for theme changes
watch(() => settingStore.isDarkMode, () => {
  createEditor()
})

// Watch for config changes
watch(() => config.value?.editor, () => {
  createEditor()
}, { deep: true })

onMounted(() => {
  createEditor()
})

onUnmounted(() => {
  if (editorView) {
    editorView.destroy()
  }
})
</script>

<template>
  <div ref="editorContainer" class="h-full overflow-hidden"></div>
</template>

<style scoped>
/* Editor container styles */
</style>