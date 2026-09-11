/**
 * 命令常量与别名（单一数据源）
 *
 * 菜单（NddMenuBar）、工具栏（NddToolbar）、命令分发（useCommands）三处
 * 都要判断「某个命令到底能不能用」，此前各自维护一份列表，改一处漏两处。
 * 这里收敛成唯一来源。
 */

/**
 * 菜单里有入口、但当前版本尚未接线的命令。
 * 命中时统一提示「该功能暂未实现」，并在菜单里置灰，避免用户点了没反应。
 */
export const NOT_IMPLEMENTED = new Set([
  'new-window',           // 在新窗口中打开（Wails v2 单主窗口模型不支持）
  'import-shortcut',      // 快捷键导入/导出
  'export-shortcut',
  'recent-cmp',           // 最近对比记录
  'toggle-indent-guide',  // 缩进参考线
])

/**
 * 命令别名：菜单沿用的 notepad-- 命名 → CodeEditor 实际识别的命名。
 *
 * 两套命名此前从未对齐，「复制当前行」「Tab 转空格」等一批菜单点击后
 * 会被 CodeEditor 静默丢弃。所有命令都先过这张表再派发。
 */
export const CMD_ALIAS: Record<string, string> = {
  // 行操作
  'line-dup': 'line-duplicate',
  'line-del': 'line-remove',
  'line-up': 'line-moveUp',
  'line-down': 'line-moveDown',
  'line-rmdup': 'line-removeDuplicate',
  'line-rmempty': 'line-removeEmpty',
  'line-rmblank': 'line-removeBlank',
  'line-insert-above': 'insert-blank-above',
  'line-insert-below': 'insert-blank-below',
  // 空白 / 缩进转换
  'tab2space': 'tab-to-spaces',
  'space2tab-all': 'spaces-all-to-tabs',
  'space2tab-lead': 'spaces-leading-to-tabs',
  // 显示符号（菜单是 show-*，编辑器是 toggle-* 且会写回配置）
  'show-spaces': 'toggle-whitespace',
  'show-eol': 'toggle-eol',
}
