/**
 * ext-context-menu — 编辑器右键菜单数据 + 状态管理
 *
 * 职责（与原 CodeEditor.vue 完全等价搬运，无行为变更）：
 *  - 右键菜单状态（visible / x / y / hasSelection）
 *  - 菜单数据（按区块分组：剪切/复制/粘贴/全选 / 撤销/重做 / 查找/替换 / 行操作 / 大小写 / JSON/XML / Tab空格）
 *  - 命令分发：容器传入 dispatcher，本模块只负责关闭菜单并把 cmd 转发过去
 *
 * 设计目标：状态 + 数据 + UI 模板（外置）三件套。模板仍留在容器里以便 Vue 的 Teleport 包裹。
 */
import { ref } from 'vue'

/** 菜单项数据 */
export interface ContextMenuItem {
  /** 编辑器命令（与 useCommands 的 cmd 字符串一致；也支持右键独有的 find/replace） */
  cmd: string
  /** 显示标签 */
  label: string
  /** 快捷键提示（无则 undefined） */
  shortcut?: string
  /** 仅在有选区时可点击 */
  needSelection?: boolean
}

export interface ContextMenuSection {
  items: ContextMenuItem[]
}

/**
 * 右键菜单静态数据（与原 CodeEditor 模板中的硬编码顺序一致）。
 * 顺序不可改：影响右键菜单的视觉排列。
 */
export const CONTEXT_MENU_SECTIONS: ContextMenuSection[] = [
  // 剪切 / 复制 / 粘贴 / 全选
  {
    items: [
      { cmd: 'cut',       label: '剪切',   shortcut: 'Ctrl+X', needSelection: true },
      { cmd: 'copy',      label: '复制',   shortcut: 'Ctrl+C', needSelection: true },
      { cmd: 'paste',     label: '粘贴',   shortcut: 'Ctrl+V' },
      { cmd: 'selectAll', label: '全选',   shortcut: 'Ctrl+A' },
    ],
  },
  // 撤销 / 重做
  {
    items: [
      { cmd: 'undo', label: '撤销', shortcut: 'Ctrl+Z' },
      { cmd: 'redo', label: '重做', shortcut: 'Ctrl+Y' },
    ],
  },
  // 查找 / 替换
  {
    items: [
      { cmd: 'find',    label: '查找', shortcut: 'Ctrl+F' },
      { cmd: 'replace', label: '替换', shortcut: 'Ctrl+H' },
    ],
  },
  // 行操作
  {
    items: [
      { cmd: 'comment-line', label: '注释/取消注释', shortcut: 'Ctrl+/' },
      { cmd: 'duplicate',    label: '复制当前行',   shortcut: 'Ctrl+D' },
      { cmd: 'delete-line',  label: '删除当前行',   shortcut: 'Ctrl+L' },
      { cmd: 'move-up',      label: '上移当前行',   shortcut: 'Ctrl+Shift+↑' },
      { cmd: 'move-down',    label: '下移当前行',   shortcut: 'Ctrl+Shift+↓' },
    ],
  },
  // 大小写转换
  {
    items: [
      { cmd: 'uppercase', label: '转为大写',     needSelection: true },
      { cmd: 'lowercase', label: '转为小写',     needSelection: true },
      { cmd: 'titlecase', label: '首字母大写',   needSelection: true },
    ],
  },
  // 格式化工具
  {
    items: [
      { cmd: 'format-json',   label: 'JSON 格式化' },
      { cmd: 'minify-json',   label: 'JSON 压缩' },
      { cmd: 'validate-json', label: 'JSON 校验' },
      { cmd: 'format-xml',    label: 'XML 格式化' },
    ],
  },
  // 空白处理
  {
    items: [
      { cmd: 'tab-to-spaces',  label: 'Tab 转空格' },
      { cmd: 'spaces-to-tabs', label: '空格转 Tab' },
      { cmd: 'trim-trailing',  label: '去除行尾空格' },
    ],
  },
]

export interface ExtContextMenu {
  /** 是否显示 */
  visible: ReturnType<typeof ref<boolean>>
  /** 菜单位置 */
  x: ReturnType<typeof ref<number>>
  y: ReturnType<typeof ref<number>>
  /** 是否有选区（用于判断 case-* 等命令的 disabled） */
  hasSelection: ReturnType<typeof ref<boolean>>
  /** 打开菜单（在容器 onContextMenu 调用） */
  open: (e: MouseEvent, view: any) => void
  /** 关闭菜单 */
  close: () => void
  /** 容器提供 dispatcher，本模块回调时调用 */
  run: (dispatcher: (cmd: string) => void) => (cmd: string) => void
}

const MENU_W = 220
const MENU_H = 400

/**
 * 右键菜单 composable。
 * 容器把模板 + 关闭逻辑都外包给本模块的 visible/x/y 状态；
 * 命令分发则由容器提供 dispatcher（避免本模块反向依赖 useCommands）。
 */
export function useEditorContextMenu(): ExtContextMenu {
  const visible = ref(false)
  const x = ref(0)
  const y = ref(0)
  const hasSelection = ref(false)

  function open(e: MouseEvent, view: any) {
    e.preventDefault()
    e.stopPropagation()
    if (!view) return
    const selection = view.state.selection.main
    hasSelection.value = selection.from !== selection.to
    // 边界检测：防止菜单超出视口右下边界
    let nx = e.clientX
    let ny = e.clientY
    if (nx + MENU_W > window.innerWidth) nx = window.innerWidth - MENU_W - 4
    if (ny + MENU_H > window.innerHeight) ny = window.innerHeight - MENU_H - 4
    x.value = Math.max(2, nx)
    y.value = Math.max(2, ny)
    visible.value = true
  }

  function close() {
    visible.value = false
  }

  /**
   * 返回一个适配容器模板的 dispatcher：
   * 点击菜单项时调用 (cmd) => ... → 关闭菜单 → 转发 cmd
   */
  function run(dispatcher: (cmd: string) => void) {
    return (cmd: string) => {
      visible.value = false
      dispatcher(cmd)
    }
  }

  return { visible, x, y, hasSelection, open, close, run }
}