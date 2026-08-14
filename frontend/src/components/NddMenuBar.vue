<script lang="ts" setup>
import { ref, reactive, onMounted, onUnmounted, watch } from 'vue'
import { useEditorStore } from '@/stores'
import { GetRecentFiles } from '../../wailsjs/go/main/App'

const emit = defineEmits<{ (e: 'cmd', name: string, ...args: any[]): void }>()
const editor = useEditorStore()

// ---- 语言分组 ----
const LANG_GROUPS: Record<string, string[]> = {
  A: ['ASP','ActionScript','Assembly','AutoIt','AviSynth','ASN.1'],
  B: ['BaanC','Bash','Batch','BlitzBasic'],
  C: ['C','C++','C#','Objective C','CSS','CMake','CoffeeScript','Csound'],
  D: ['D','Diff'],
  E: ['ESCRIPT','Erlang','Edifact'],
  F: ['Fortran','Fortran77','Forth','FreeBasic'],
  G: ['Go'],
  H: ['HTML'],
  I: ['IDL','ini','Intel HEX'],
  J: ['Java','JavaScript','Jsp','JSON'],
  L: ['Lua','Lisp','LaTeX'],
  M: ['Makefile','MMIXAL','MarkDown','Matlab'],
  N: ['Nfo','NSIS','NCrontab','Nim'],
  O: ['OScript','Octave'],
  P: ['Pascal','Perl','PHP','Po','PostScript','Pov','PowerShell','Properties file','PureBasic','Python'],
  R: ['R','RC','Ruby','Rust','Registry','REBOL'],
  S: ['Sql','Shell','Scheme','Smalltalk','S-Record','Swift','SPICE'],
  T: ['TeX','Tcl','TypeScript','Txt2tags'],
  V: ['VB','Verilog','Visual Basic','VHDL','Visual Prolog'],
  其他: ['XML','YAML','TXT','UserDefine'],
}

// ---- 类型 ----
type Item = { label?: string; cmd?: string; sub?: string; key?: string; sep?: boolean; chk?: boolean; grp?: string; st?: string }

// ---- 帮助函数 ----
function showLabel(s?: string) { return (s || '').replace(/\(&.\)/g, '') }

// ---- 菜单数据（完整对齐 notepad-- ccnotepad.ui + ccnotepad.cpp 所有 addAction）----
const MENUS = [
  { label: '文件(&F)', items: [
    { label: '新建(&N)', cmd: 'new-file', key: 'Ctrl+T' },
    { label: '打开(&O)...', cmd: 'open-file', key: 'Ctrl+O' },
    { label: '打开目录...', cmd: 'open-directory' },
    { label: '从收藏夹打开', sub: 'fav' },
    { label: '在视图中打开(&V)', cmd: 'open-view' },
    { sep: true },
    { label: '保存(&S)', cmd: 'save', key: 'Ctrl+S' },
    { label: '另存为(&A)...', cmd: 'save-as', key: 'Ctrl+Alt+S' },
    { label: '全部保存', cmd: 'save-all' },
    { label: '重命名...', cmd: 'rename-file' },
    { sep: true },
    { label: '关闭(&C)', cmd: 'close-tab', key: 'Ctrl+W' },
    { label: '关闭所有', cmd: 'close-all', key: 'Ctrl+Shift+W' },
    { label: '关闭其他文档', cmd: 'close-others' },
    { label: '关闭左侧文档', cmd: 'close-left' },
    { label: '关闭右侧文档', cmd: 'close-right' },
    { sep: true },
    { label: '在新窗口中打开', cmd: 'new-window' },
    { label: '重新加载(&L)', cmd: 'reload-file' },
    { label: '清空历史记录', cmd: 'clear-history' },
    { label: '最近打开的文件', sub: 'recent' },
    { sep: true },
    { label: '草稿管理...', cmd: 'manage-drafts' },
    { label: '最近的文件...', cmd: 'recent-files' },
    { label: '最近的文件夹...', cmd: 'recent-folders' },
    { sep: true },
    { label: '保存工作空间...', cmd: 'save-workspace' },
    { label: '打开工作空间...', cmd: 'open-workspace' },
    { sep: true },
    { label: '打印(&P)...', cmd: 'print', key: 'Ctrl+P' },
    { sep: true },
    { label: '退出(&X)', cmd: 'exit', key: 'Ctrl+Q' },
  ] as Item[] },
  { label: '编辑(&E)', items: [
    { label: '撤销(&U)', cmd: 'undo', key: 'Ctrl+Z' },
    { label: '重做(&R)', cmd: 'redo', key: 'Ctrl+Y' },
    { sep: true },
    { label: '剪切(&T)', cmd: 'cut', key: 'Ctrl+X' },
    { label: '复制(&C)', cmd: 'copy', key: 'Ctrl+C' },
    { label: '粘贴(&P)', cmd: 'paste', key: 'Ctrl+V' },
    { label: '删除(&L)', cmd: 'delete' },
    { sep: true },
    { label: '全选(&A)', cmd: 'select-all', key: 'Ctrl+A' },
    { label: '开始/停止录制', cmd: 'record-macro' },
    { label: '停止录制', cmd: 'stop-macro' },
    { label: '播放宏', cmd: 'play-macro' },
    { label: '保存当前录制的宏...', cmd: 'save-macro' },
    { label: '多次运行宏...', cmd: 'run-macro-multi' },
    { sep: true },
    { label: '剪贴板历史记录', cmd: 'clipboard-history' },
    { label: '复制到剪贴板', sub: 'clipboard' },
    { sep: true },
    { label: '缩进(&I)', cmd: 'indent' },
    { label: '取消缩进(&U)', cmd: 'dedent' },
    { label: '转为大写(&U)', cmd: 'case-upper' },
    { label: '转为小写(&L)', cmd: 'case-lower' },
    { sep: true },
    { label: '设置只读(&R)', cmd: 'toggle-readonly' },
    { label: '清空只读标记', cmd: 'clear-readonly' },
    { sep: true },
    { label: '换行符转换', sub: 'le' },
    { label: '以文本模式打开', cmd: 'open-text' },
    { label: '以二进制模式打开', cmd: 'open-hex' },
    { label: '空白字符操作', sub: 'blank' },
    { label: '转换大小写为', sub: 'case' },
    { label: '行操作', sub: 'line' },
    { label: '注释/取消注释', sub: 'comment' },
    { sep: true },
    { label: '列块模式', cmd: 'column-mode' },
    { label: '列块编辑...', cmd: 'column-block', key: 'Alt+X' },
  ] as Item[] },
  { label: '查找(&S)', items: [
    { label: '查找(&F)...', cmd: 'find', key: 'Ctrl+F' },
    { label: '查找下一个', cmd: 'find-next', key: 'F3' },
    { label: '查找上一个', cmd: 'find-prev', key: 'Shift+F3' },
    { label: '查找并替换(&R)...', cmd: 'replace', key: 'Ctrl+H' },
    { sep: true },
    { label: '在目录中查找...', cmd: 'find-dir', key: 'Ctrl+Shift+D' },
    { label: '在多个文件中查找...', cmd: 'find-multi' },
    { label: '在多个文件中替换...', cmd: 'replace-multi' },
    { sep: true },
    { label: '标记所有...', cmd: 'mark-all' },
    { label: '取消所有标记', cmd: 'clear-mark' },
    { sep: true },
    { label: '转到行(&G)...', cmd: 'goto-line', key: 'Ctrl+G' },
    { label: '转到匹配的括号', cmd: 'goto-bracket', key: 'Ctrl+B' },
    { label: '上一个位置', cmd: 'prev-position' },
    { label: '下一个位置', cmd: 'next-position' },
    { sep: true },
    { label: '书签', sub: 'bookmark' },
    { label: '标记颜色', sub: 'mark' },
  ] as Item[] },
  { label: '视图(&V)', items: [
    { label: '显示符号', sub: 'display' },
    { label: '查找结果', cmd: 'search-result' },
    { label: '图标大小', sub: 'iconsize' },
    { sep: true },
    { label: '自动换行', cmd: 'toggle-wrap', chk: true, st: 'wrap' },
    { label: '文档地图', cmd: 'toggle-minimap' },
    { label: '文件列表视图', cmd: 'toggle-filelist', chk: true, st: 'filelist' },
    { label: '显示工具栏', cmd: 'toggle-toolbar', chk: true, st: 'toolbar' },
    { label: '显示状态栏', cmd: 'toggle-statusbar', chk: true, st: 'statusbar' },
    { label: '显示网页地址(不推荐)', cmd: 'toggle-webaddr', chk: true, st: 'webaddr' },
    { sep: true },
    { label: '代码片段面板', cmd: 'snippet-panel', chk: true, st: 'snippet' },
    { label: '全局书签面板', cmd: 'bookmark-panel', chk: true, st: 'bookmark' },
    { label: '函数列表面板', cmd: 'function-list', chk: true, st: 'functionlist' },
    { label: '文件监控面板', cmd: 'file-monitor', chk: true, st: 'filemonitor' },
    { label: '日志查看模式', cmd: 'log-mode', chk: true, st: 'log' },
    { label: '自动跟随系统主题', cmd: 'toggle-auto-theme', chk: true, st: 'autotheme' },
    { sep: true },
    { label: '放大(&I)', cmd: 'zoom-in', key: 'Ctrl+=' },
    { label: '缩小(&O)', cmd: 'zoom-out', key: 'Ctrl+-' },
    { label: '恢复默认缩放', cmd: 'zoom-reset', key: 'Ctrl+/' },
    { sep: true },
    { label: '全屏', cmd: 'fullscreen', key: 'F11' },
  ] as Item[] },
  { label: '编码(&N)', items: [
    { label: '编码字符集', sub: 'enc-charset' },
    { label: '以 ANSI 编码', cmd: 'encode-ANSI', chk: true, grp: 'enc' },
    { label: '以 UTF-8 编码', cmd: 'encode-UTF-8', chk: true, grp: 'enc' },
    { label: '以 UTF-8-BOM 编码', cmd: 'encode-UTF-8-BOM', chk: true, grp: 'enc' },
    { label: '以 UCS-2 BE BOM 编码', cmd: 'encode-UCS-2-BE', chk: true, grp: 'enc' },
    { label: '以 UCS-2 LE BOM 编码', cmd: 'encode-UCS-2-LE', chk: true, grp: 'enc' },
    { sep: true },
    { label: '转换为 ANSI', cmd: 'conv-ANSI' },
    { label: '转换为 UTF-8', cmd: 'conv-UTF-8' },
    { label: '转换为 UTF-8-BOM', cmd: 'conv-UTF-8-BOM' },
    { label: '转换为 UCS-2 BE BOM', cmd: 'conv-UCS-2-BE' },
    { label: '转换为 UCS-2 LE BOM', cmd: 'conv-UCS-2-LE' },
    { sep: true },
    { label: '批量编码转换', cmd: 'batch-convert' },
  ] as Item[] },
  { label: '语言(&L)', items: [] as Item[] },
  { label: '设置(&T)', items: [
    { label: '首选项(&P)...', cmd: 'preferences' },
    { sep: true },
    { label: '导入', sub: 'import' },
    { label: '导出', sub: 'export' },
    { sep: true },
    { label: '编辑弹出菜单', cmd: 'edit-context-menu' },
    { sep: true },
    { label: '界面语言', sub: 'lang' },
    { label: '主题样式', cmd: 'theme-style' },
    { label: '自定义语言格式', cmd: 'define-lang' },
    { label: '语言文件后缀', cmd: 'lang-suffix' },
    { label: '快捷键管理器', cmd: 'shortcut-mgr' },
  ] as Item[] },
  { label: '工具(&O)', items: [
    { label: 'MD5/SHA 哈希', cmd: 'md5-hash' },
    { label: '格式化语言', sub: 'fmtlang' },
    { label: '批量查找替换', cmd: 'batch-find' },
    { sep: true },
    { label: '批量重命名', cmd: 'batch-rename' },
    { label: '格式转换', cmd: 'open-converter' },
    { label: '文档对比', cmd: 'open-diff' },
    { sep: true },
    { label: '正则测试...', cmd: 'regex-tester' },
    { label: 'JSONPath查询', cmd: 'json-path' },
    { label: 'JSON生成结构体', cmd: 'json-to-struct' },
    { label: 'JSON对比', cmd: 'json-diff' },
    { label: '全局搜索...', cmd: 'search-files', key: 'Ctrl+Shift+F' },
    { sep: true },
    { label: '脚本管理器...', cmd: 'script-manager' },
    { sep: true },
    { label: '图片编辑模式', cmd: 'image-editor' },
    { label: '取色器', cmd: 'color-picker' },
    { sep: true },
    { label: '在资源管理器中打开', cmd: 'explorer' },
    { sep: true },
    { label: '打印(&P)', cmd: 'print', key: 'Ctrl+P' },
    { label: '全屏', cmd: 'fullscreen', key: 'F11' },
  ] as Item[] },
  { label: '对比(&C)', items: [
    { label: '文件对比', cmd: 'file-cmp' },
    { label: '目录对比', cmd: 'dir-cmp' },
    { label: '二进制对比', cmd: 'bin-cmp' },
    { sep: true },
    { label: '选择左侧文件', cmd: 'sel-left' },
    { label: '选择右侧文件', cmd: 'sel-right' },
    { label: '对比规则...', cmd: 'cmp-rule' },
    { sep: true },
    { label: '最近对比', cmd: 'recent-cmp' },
  ] as Item[] },
  { label: '关于(&A)', items: [
    { label: '关于 EasyText...', cmd: 'about' },
  ] as Item[] },
]

// ---- 子菜单 ----
// 用 reactive() 包裹使 SUBS.recent 等动态子菜单能响应式刷新。
const SUBS = reactive<Record<string, Item[]>>({
  'le': [
    { label: 'Windows(CR+LF)', cmd: 'le-CRLF', chk: true, grp: 'le' },
    { label: 'Unix(LF)', cmd: 'le-LF', chk: true, grp: 'le' },
    { label: 'Mac(CR)', cmd: 'le-CR', chk: true, grp: 'le' },
  ],
  'blank': [
    { label: '删除行首空白', cmd: 'trim-head' },
    { label: '删除行尾空白', cmd: 'trim-tail' },
    { label: '删除首尾空白', cmd: 'trim-both' },
    { sep: true },
    { label: 'Tab 转空格', cmd: 'tab2space' },
    { label: '空格转 Tab(全部)', cmd: 'space2tab-all' },
    { label: '行首空格转 Tab', cmd: 'space2tab-lead' },
  ],
  'case': [
    { label: 'UPPERCASE (大写)', cmd: 'case-upper' },
    { label: 'lowercase (小写)', cmd: 'case-lower' },
    { label: 'Proper Case (首字母大写)', cmd: 'case-title' },
    { label: 'Proper Case (混合)', cmd: 'case-title-blend' },
    { label: 'Sentence case (句首大写)', cmd: 'case-sentence' },
    { label: 'Sentence case (混合)', cmd: 'case-sentence-blend' },
    { label: 'Invert Case (大小写反转)', cmd: 'case-invert' },
    { label: 'Random Case (随机)', cmd: 'case-random' },
  ],
  'line': [
    { label: '复制当前行', cmd: 'line-dup', key: 'Ctrl+D' },
    { label: '删除当前行', cmd: 'line-del', key: 'Ctrl+L' },
    { label: '移除重复行', cmd: 'line-rmdup' },
    { label: '移除连续重复行', cmd: 'line-removeConsecutiveDuplicate' },
    { label: '拆分行', cmd: 'line-split' },
    { label: '合并行', cmd: 'line-join' },
    { label: '上移当前行', cmd: 'line-up', key: 'Ctrl+Shift+Up' },
    { label: '下移当前行', cmd: 'line-down', key: 'Ctrl+Shift+Down' },
    { label: '删除空行', cmd: 'line-rmempty' },
    { label: '删除空行(含空白字符)', cmd: 'line-rmblank' },
    { label: '在上方插入空行', cmd: 'line-insert-above' },
    { label: '在下方插入空行', cmd: 'line-insert-below' },
    { label: '反转行顺序', cmd: 'line-reverse' },
    { label: '随机排列行顺序', cmd: 'line-randomize' },
    { sep: true },
    { label: '按字典升序排列', cmd: 'sort-asc' },
    { label: '按字典升序(忽略大小写)', cmd: 'sort-asc-ci' },
    { label: '按字典降序排列', cmd: 'sort-desc' },
    { label: '按字典降序(忽略大小写)', cmd: 'sort-desc-ci' },
    { label: '按整数升序排列', cmd: 'sort-int-asc' },
    { label: '按整数降序排列', cmd: 'sort-int-desc' },
    { label: '按小数(点)升序排列', cmd: 'sort-float-asc' },
    { label: '按小数(点)降序排列', cmd: 'sort-float-desc' },
    { label: '按小数(逗号)升序排列', cmd: 'sort-comma-asc' },
    { label: '按小数(逗号)降序排列', cmd: 'sort-comma-desc' },
  ],
  'bookmark': [
    { label: '设置/取消书签', cmd: 'toggle-bookmark', key: 'Ctrl+F2' },
    { label: '下一个书签', cmd: 'next-bookmark', key: 'F2' },
    { label: '上一个书签', cmd: 'prev-bookmark', key: 'Shift+F2' },
    { label: '清除所有书签', cmd: 'clear-bookmarks' },
    { sep: true },
    { label: '剪切书签行', cmd: 'cut-bookmark-lines' },
    { label: '复制书签行', cmd: 'copy-bookmark-lines' },
    { label: '粘贴到书签行', cmd: 'paste-bookmark-lines' },
    { label: '删除书签行', cmd: 'delete-bookmark-lines' },
    { label: '删除非书签行', cmd: 'delete-unbookmark-lines' },
  ],
  'mark': [
    { label: '红色标记', cmd: 'mark-red' },
    { label: '黄色标记', cmd: 'mark-yellow' },
    { label: '蓝色标记', cmd: 'mark-blue' },
    { sep: true },
    { label: '颜色 1', cmd: 'mark-1', chk: true, grp: 'mark' },
    { label: '颜色 2', cmd: 'mark-2', chk: true, grp: 'mark' },
    { label: '颜色 3', cmd: 'mark-3', chk: true, grp: 'mark' },
    { label: '颜色 4', cmd: 'mark-4', chk: true, grp: 'mark' },
    { label: '颜色 5', cmd: 'mark-5', chk: true, grp: 'mark' },
    { label: '循环颜色标记', cmd: 'mark-loop' },
    { sep: true },
    { label: '清除所有标记', cmd: 'clear-all-marks' },
  ],
  'display': [
    { label: '显示空格/Tab', cmd: 'show-spaces', chk: true, st: 'spaces' },
    { label: '显示行尾符', cmd: 'show-eol', chk: true, st: 'eol' },
    { label: '显示全部', cmd: 'show-all', chk: true, st: 'all' },
  ],
  'iconsize': [
    { label: '24x24', cmd: 'iconsize-24', chk: true, grp: 'iconsize' },
    { label: '36x36', cmd: 'iconsize-36', chk: true, grp: 'iconsize' },
    { label: '48x48', cmd: 'iconsize-48', chk: true, grp: 'iconsize' },
  ],
  'enc-charset': [
    { label: '阿拉伯语', cmd: 'encode-ar' },
    { label: '波罗的海语', cmd: 'encode-baltic' },
    { label: '中欧语', cmd: 'encode-ce' },
    { label: '简体中文(GB2312)', cmd: 'encode-GB2312' },
    { label: '繁体中文(Big5)', cmd: 'encode-Big5' },
    { label: '西里尔语', cmd: 'encode-cyrillic' },
    { label: '希腊语', cmd: 'encode-greek' },
    { label: '希伯来语', cmd: 'encode-hebrew' },
    { label: '日语(Shift-JIS)', cmd: 'encode-SJIS' },
    { label: '韩语', cmd: 'encode-korean' },
    { label: '泰语', cmd: 'encode-thai' },
    { label: '土耳其语', cmd: 'encode-turkish' },
    { label: '越南语', cmd: 'encode-vietnamese' },
    { label: '西欧语', cmd: 'encode-we' },
  ],
  'lang': [
    { label: '中文', cmd: 'lang-zh', chk: true, grp: 'lang' },
    { label: 'English', cmd: 'lang-en', chk: true, grp: 'lang' },
  ],
  'fmtlang': [
    { label: '格式化 XML', cmd: 'fmt-xml' },
    { label: '格式化 JSON', cmd: 'fmt-json' },
  ],
  'comment': [
    { label: '单行注释/取消注释', cmd: 'comment-line', key: 'Ctrl+/' },
    { label: '块注释/取消注释', cmd: 'comment-block', key: 'Ctrl+Shift+/' },
  ],
  'clipboard': [
    { label: '复制当前行', cmd: 'copy-line' },
    { label: '剪切当前行', cmd: 'cut-line' },
  ],
  'fav': [
    { label: '管理收藏夹...', cmd: 'manage-fav' },
    { sep: true },
    { label: '(空)', cmd: 'fav-empty' },
  ],
  'import': [
    { label: '导入主题...', cmd: 'import-theme' },
    { label: '导入快捷键...', cmd: 'import-shortcut' },
  ],
  'export': [
    { label: '导出主题...', cmd: 'export-theme' },
    { label: '导出快捷键...', cmd: 'export-shortcut' },
  ],
  'recent': [] as Item[],
})

// 刷新「最近打开的文件」子菜单：从后端拉取并写回 SUBS.recent，
// 修复"菜单最近打开文件一直空"的 Bug。
async function refreshRecentSubmenu() {
  try {
    const list = (await GetRecentFiles()) as Array<{ path: string; name: string }> | null
    if (!list || list.length === 0) {
      SUBS.recent.splice(0, SUBS.recent.length, { label: '(无最近文件)', cmd: 'recent-empty' })
      return
    }
    const items: Item[] = list.slice(0, 10).map((e, i) => ({
      label: `${i + 1}. ${e.name || e.path.split(/[\\/]/).pop() || e.path}`,
      cmd: `recent-open-${i}`,
      // 同时挂个 st 携带原路径，交给 onMenuCmd 解析
      st: e.path,
    }))
    SUBS.recent.splice(0, SUBS.recent.length, ...items)
  } catch (e) {
    // 后端未就绪：保持空数组，模板已经处理（空菜单=短暂空白）
    console.warn('refreshRecentSubmenu failed', e)
  }
}

onMounted(() => {
  document.addEventListener('click', (e) => {
    if (!(e.target as HTMLElement).closest('.menubar')) closeAll()
  })
  refreshRecentSubmenu()
  // 监听「最近文件已更新」事件，触发刷新（在 useFileOps.AddRecentEntry 后调用）
  document.addEventListener('recent-updated', refreshRecentSubmenu)
  // 窗口聚焦时刷新一次（应对外部修改数据库或上次崩溃残留）
  window.addEventListener('focus', refreshRecentSubmenu)
})

onUnmounted(() => {
  document.removeEventListener('recent-updated', refreshRecentSubmenu)
  window.removeEventListener('focus', refreshRecentSubmenu)
})

// ---- 交互状态 ----
const active = ref(-1)
const subKey = ref('')
const langLetter = ref('')
const checks = ref<Record<string, boolean>>({ wrap: false, filelist: false, toolbar: true, statusbar: true, webaddr: false, spaces: false, eol: false, all: false })
const radios = ref<Record<string, string>>({ le: 'CRLF', enc: 'UTF-8', iconsize: '36', lang: 'zh-CN', mark: '' })

function isOpen(i: number) { return active.value === i }
function open(i: number) { active.value = i; subKey.value = ''; langLetter.value = '' }
function toggle(i: number) { active.value === i ? closeAll() : open(i) }
function onHover(i: number) { if (active.value >= 0 && i !== active.value) open(i) }
function closeAll() { active.value = -1; subKey.value = ''; langLetter.value = '' }

function clickItem(it: Item) {
  if (it.sub) { subKey.value = it.sub; return }
  if (it.chk) {
    if (it.st) { checks.value[it.st] = !checks.value[it.st]; emit('cmd', it.cmd!) }
    else if (it.grp) { radios.value[it.grp] = it.cmd!.split('-').pop()!; emit('cmd', it.cmd!) }
    else emit('cmd', it.cmd!)
  } else if (it.cmd) {
    // 「最近打开的文件」子项的 cmd 形如 recent-open-N，附加的 st 字段携带原路径，
    // 这里把 path 作为参数传给父监听器，匹配 useCommands 中新增的 `recent-open`。
    if (it.cmd.startsWith('recent-open') && it.st) emit('cmd', 'recent-open', it.st)
    else emit('cmd', it.cmd)
  }
  closeAll()
}

function checked(it: Item): boolean {
  if (it.st) return checks.value[it.st] ?? false
  if (it.grp) return radios.value[it.grp] === it.cmd!.split('-').pop()
  return false
}

function pickSub(it: Item) {
  if (it.sub) subKey.value = it.sub
  else if (it.cmd) { emit('cmd', it.cmd); closeAll() }
}

function pickLang(lang: string) { emit('cmd', 'set-lang', lang); closeAll() }
</script>

<template>
  <div class="menubar">
    <div v-for="(menu, idx) in MENUS" :key="menu.label" class="menu-top" @mouseenter="onHover(idx)">
      <button class="menu-btn" :class="{ active: isOpen(idx) }" @click.stop="toggle(idx)">
        {{ showLabel(menu.label) }}
      </button>

      <!-- 普通下拉菜单 -->
      <div v-if="isOpen(idx) && menu.items.length > 0" class="menu-dd">
        <template v-for="(it, i) in (menu.label.startsWith('语言') ? [] : menu.items)" :key="i">
          <div v-if="it.sep" class="menu-sep" />
          <div v-else class="menu-row" :class="{ hl: subKey === it.sub }"
            @click.stop="clickItem(it)"
            @mouseenter="it.sub && (subKey = it.sub)">
            <span class="flex items-center gap-1">
              <span v-if="it.chk" class="w-3 text-center blue">{{ checked(it) ? '✓' : '' }}</span>
              <span>{{ showLabel(it.label) }}</span>
            </span>
            <span v-if="it.key" class="menu-short">{{ it.key }}</span>
            <span v-else-if="it.sub" class="menu-arrow">▶</span>
            <!-- ★ 子菜单嵌套在 menu-row 内部，跟随行位置 -->
            <div v-if="it.sub && subKey === it.sub && SUBS[it.sub]" class="menu-sub">
              <template v-for="(si, j) in SUBS[it.sub]" :key="j">
                <div v-if="si.sep" class="menu-sep" />
                <div v-else class="menu-row" @click.stop="pickSub(si)">
                  <span class="flex items-center gap-1">
                    <span v-if="si.chk" class="w-3 text-center blue">{{ checked(si) ? '✓' : '' }}</span>
                    <span>{{ showLabel(si.label) }}</span>
                  </span>
                  <span v-if="si.key" class="menu-short">{{ si.key }}</span>
                </div>
              </template>
            </div>
          </div>
        </template>
      </div>

      <!-- 语言菜单（特殊） -->
      <div v-if="isOpen(idx) && menu.label.startsWith('语言')" class="menu-dd">
        <div v-for="(_v, letter) in LANG_GROUPS" :key="letter"
          class="menu-row" :class="{ hl: langLetter === letter }"
          @click.stop="langLetter = letter"
          @mouseenter="langLetter = letter">
          {{ letter }} <span class="menu-arrow ml-auto">▶</span>
          <div v-if="langLetter === letter" class="menu-sub">
            <div v-for="lang in LANG_GROUPS[letter]" :key="lang" class="menu-row" @click.stop="pickLang(lang)">
              <span class="w-3 text-center blue">{{ editor.activeTab?.language?.toLowerCase() === lang.toLowerCase() ? '✓' : '' }}</span>
              <span class="ml-1">{{ lang }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.menubar {
  display: flex; align-items: center; height: 24px; padding: 0 2px;
  border-bottom: 1px solid #d1d5db; background: #f0f0f0; user-select: none;
  font-family: 'Microsoft YaHei', sans-serif; font-size: 12px;
}
html.dark .menubar { background: #2d2d2d; border-color: #454545; }
.menu-top { position: relative; }
.menu-btn { padding: 0 6px; height: 22px; border: none; background: none; color: #333; border-radius: 3px; cursor: pointer; font-size: 12px; white-space: nowrap; }
.menu-btn:hover, .menu-btn.active { background: #d0d0d0; }
html.dark .menu-btn { color: #d4d4d4; }
html.dark .menu-btn:hover, html.dark .menu-btn.active { background: #505050; }

.menu-dd {
  position: absolute; left: 0; top: 100%; z-index: 9999; min-width: 240px;
  background: #fff; border: 1px solid #d1d5db; box-shadow: 0 4px 16px rgba(0,0,0,.18);
  border-radius: 4px; padding: 2px 0; font-size: 12px;
}
html.dark .menu-dd { background: #1e1e1e; border-color: #454545; }

.menu-row {
  display: flex; align-items: center; justify-content: space-between; position: relative;
  padding: 4px 10px; cursor: pointer; white-space: nowrap; color: #333; min-width: 200px;
}
.menu-row:hover, .menu-row.hl { background: #e8f0fe; color: #1a73e8; }
html.dark .menu-row { color: #d4d4d4; }
html.dark .menu-row:hover, html.dark .menu-row.hl { background: #094771; color: #60a5fa; }

/* ★ 子菜单：嵌套在 .menu-row 内部，top:0 对齐当前行 */
.menu-sub {
  position: absolute; left: 100%; top: 0; z-index: 10000; min-width: 240px;
  background: #fff; border: 1px solid #d1d5db; box-shadow: 0 4px 16px rgba(0,0,0,.18);
  border-radius: 4px; padding: 2px 0; font-size: 12px;
}
html.dark .menu-sub { background: #1e1e1e; border-color: #454545; }

.menu-sep { margin: 3px 0; height: 1px; background: #e5e5e5; }
html.dark .menu-sep { background: #454545; }
.menu-short, .menu-arrow { margin-left: 24px; font-size: 11px; color: #999; }
.blue { color: #2563eb; }
</style>
