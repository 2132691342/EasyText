import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'
import type { ThemeName } from '@/types'

// ---------- Tailwind 合并 ----------
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// ---------- 字节数据处理 ----------
export function normalizeBytes(data: any): number[] {
  if (!data) return []
  if (typeof data === 'string') {
    if (data.length === 0) return []
    try {
      const binary = atob(data)
      const bytes = new Uint8Array(binary.length)
      for (let i = 0; i < binary.length; i++) {
        bytes[i] = binary.charCodeAt(i) & 0xFF
      }
      return Array.from(bytes)
    } catch {
      const bytes = new Uint8Array(data.length)
      for (let i = 0; i < data.length; i++) bytes[i] = data.charCodeAt(i) & 0xFF
      return Array.from(bytes)
    }
  }
  if (!data.length) return []
  try {
    return Array.from(new Uint8Array(data))
  } catch {
    const result: number[] = []
    for (let i = 0; i < data.length; i++) {
      const val = data[i]
      if (typeof val === 'number') result.push(val)
      else if (typeof val === 'string') result.push(parseInt(val, 10) || 0)
      else result.push(0)
    }
    return result
  }
}

// ---------- 格式化 ----------
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

export function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit',
  })
}

export function countLines(content: string): number {
  if (!content) return 0
  return content.split('\n').length
}

// ---------- 文件扩展名 ----------
export function getFileExtension(filename: string): string {
  const ext = filename.split('.').pop()?.toLowerCase() || ''
  return ext
}

// ============================================================
// 80+ 语言扩展名映射 - 完全匹配 notepad-- extlexermanager
// 参考: notepad-- src/langstyledefine.h 及 qscilexer*.cpp
// ============================================================
export function getLanguageFromExtension(ext: string): string {
  const langMap: Record<string, string> = {
    // --- C 家族 ---
    'c': 'c', 'h': 'c',
    'cpp': 'cpp', 'cc': 'cpp', 'cxx': 'cpp', 'c++': 'cpp', 'hpp': 'cpp', 'hh': 'cpp', 'hxx': 'cpp',
    'cs': 'csharp',
    'm': 'objectivec', 'mm': 'objectivecpp',
    // --- Web 前端 ---
    'js': 'javascript', 'jsx': 'javascript', 'mjs': 'javascript', 'cjs': 'javascript',
    'ts': 'typescript', 'tsx': 'typescript',
    'html': 'html', 'htm': 'html', 'xhtml': 'html', 'shtml': 'html',
    'css': 'css', 'scss': 'css', 'sass': 'css', 'less': 'css', 'styl': 'css',
    'vue': 'vue', 'svelte': 'svelte',
    'wasm': 'wast',
    // --- 数据格式 ---
    'json': 'json', 'jsonc': 'json', 'json5': 'json',
    'xml': 'xml', 'xsl': 'xml', 'xslt': 'xml', 'xsd': 'xml', 'svg': 'xml', 'rss': 'xml', 'atom': 'xml',
    'yaml': 'yaml', 'yml': 'yaml',
    'toml': 'toml',
    'ini': 'properties', 'cfg': 'properties', 'conf': 'properties', 'properties': 'properties',
    'csv': 'csv',
    // --- 脚本语言 ---
    'py': 'python', 'pyw': 'python', 'pyx': 'python',
    'rb': 'ruby', 'rake': 'ruby', 'gemspec': 'ruby',
    'php': 'php', 'phtml': 'php', 'php3': 'php', 'php4': 'php', 'php5': 'php',
    'pl': 'perl', 'pm': 'perl',
    'lua': 'lua',
    'r': 'r', 'rmd': 'r',
    // --- JVM 生态 ---
    'java': 'java',
    'kt': 'kotlin', 'kts': 'kotlin',
    'scala': 'scala',
    'groovy': 'groovy', 'gvy': 'groovy',
    'clj': 'clojure', 'cljs': 'clojure', 'edn': 'clojure',
    // --- Go / Rust ---
    'go': 'go',
    'rs': 'rust',
    // --- Shell ---
    'sh': 'shell', 'bash': 'shell', 'zsh': 'shell', 'fish': 'shell',
    'ps1': 'powershell', 'psm1': 'powershell', 'psd1': 'powershell',
    'bat': 'batch', 'cmd': 'batch',
    // --- 数据库 ---
    'sql': 'sql',
    // --- 标记语言 ---
    'md': 'markdown', 'markdown': 'markdown', 'mkd': 'markdown', 'mdx': 'markdown',
    'tex': 'latex', 'sty': 'latex', 'cls': 'latex',
    'rst': 'restructuredtext',
    'asciidoc': 'asciidoc', 'adoc': 'asciidoc',
    // --- 配置/构建 ---
    'dockerfile': 'dockerfile',
    'cmake': 'cmake', 'cmake.in': 'cmake',
    'makefile': 'makefile', 'mk': 'makefile',
    'proto': 'protobuf',
    'graphql': 'graphql', 'gql': 'graphql',
    // --- 函数式 ---
    'hs': 'haskell', 'lhs': 'haskell',
    'erl': 'erlang', 'hrl': 'erlang',
    'ex': 'elixir', 'exs': 'elixir',
    'elm': 'elm',
    // --- 其他语言 ---
    'swift': 'swift',
    'dart': 'dart',
    'zig': 'zig',
    'nim': 'nim',
    'v': 'v',
    'jl': 'julia',
    'coffee': 'coffeescript',
    'fs': 'fsharp', 'fsi': 'fsharp', 'fsx': 'fsharp',
    'vb': 'vb', 'vbs': 'vbscript',
    'asm': 'assembly', 's': 'assembly', 'S': 'assembly',
    'pascal': 'pascal', 'pas': 'pascal',
    'fortran': 'fortran', 'f90': 'fortran', 'f95': 'fortran', 'f03': 'fortran',
    'cobol': 'cobol', 'cbl': 'cobol',
    'ada': 'ada', 'adb': 'ada', 'ads': 'ada',
    'lisp': 'commonlisp', 'el': 'commonlisp',
    'scheme': 'scheme', 'scm': 'scheme',
    'tcl': 'tcl', 'mat': 'matlab',
    'octave': 'octave',
    'd': 'd',
    'pony': 'pony',
    'solidity': 'solidity', 'sol': 'solidity',
    'twig': 'twig',
    'njk': 'nunjucks',
    'hbs': 'handlebars',
    'ejs': 'ejs',
    'pug': 'pug', 'jade': 'pug',
    'st': 'smalltalk',
    'logtalk': 'logtalk',
    'pro': 'prolog',
    // --- 日志/纯文本 ---
    'log': 'text', 'txt': 'text', 'text': 'text', 'me': 'text',
    'diff': 'diff', 'patch': 'diff',
    'nfo': 'text',
    // --- 其他文件类型 ---
    'reg': 'registry',
    'rc': 'rc',
    'nsi': 'nsis', 'nsh': 'nsis',
    'ahk': 'autohotkey',
    'autoit': 'autoit',
    'ino': 'cpp', // Arduino
    'sparql': 'sparql',
    'ttl': 'turtle',
    'nt': 'turtle',
    'nq': 'turtle',
  }
  return langMap[ext] || 'text'
}

// ---------- 二进制检测 ----------
export function isBinaryExtension(ext: string): boolean {
  const binaryExts = new Set([
    'exe', 'dll', 'so', 'dylib', 'sys',
    'png', 'jpg', 'jpeg', 'gif', 'bmp', 'ico', 'svg', 'webp',
    'pdf', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx',
    'zip', 'tar', 'gz', 'rar', '7z', 'bz2', 'xz',
    'mp3', 'mp4', 'wav', 'avi', 'mkv', 'mov', 'flv',
    'bin', 'dat', 'class', 'o', 'obj', 'lib', 'a',
    'ttf', 'otf', 'woff', 'woff2', 'eot',
  ])
  return binaryExts.has(ext.toLowerCase())
}

export function isImageExtension(ext: string): boolean {
  const imageExts = new Set(['png', 'jpg', 'jpeg', 'gif', 'bmp', 'ico', 'svg', 'webp'])
  return imageExts.has(ext.toLowerCase())
}

// ---------- Tab 视图类型 ----------
export function getTabViewType(ext: string): 'code' | 'image' | 'hex' | 'log' {
  const extLower = ext.toLowerCase()
  if (isImageExtension(extLower)) return 'image'
  // 🆕 V2.0.0: 日志文件自动识别
  if (extLower === 'log') return 'log'
  // 二进制文件（含 Office/PDF，已移除专用查看器）→ 十六进制查看
  if (isBinaryExtension(extLower)) return 'hex'
  return 'code'
}

// ---------- 生成唯一ID ----------
export function generateId(): string {
  return crypto.randomUUID?.() ?? Math.random().toString(36).substring(2, 11)
}

// ---------- 路径工具 ----------
export function normalizePath(path: string): string {
  return path.replace(/\\/g, '/')
}

export function getParentDir(path: string): string {
  const normalized = normalizePath(path)
  const parts = normalized.split('/')
  parts.pop()
  return parts.join('/')
}

export function isAbsolutePath(path: string): boolean {
  return path.startsWith('/') || /^[A-Za-z]:/.test(path)
}

export function joinPath(...parts: string[]): string {
  return parts.map(normalizePath).join('/').replace(/\/+/g, '/')
}

// ---------- 防抖/节流 ----------
export function debounce<T extends (...args: any[]) => any>(func: T, wait: number): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout> | null = null
  return (...args: Parameters<T>) => {
    if (timeout) clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }
}

export function throttle<T extends (...args: any[]) => any>(func: T, limit: number): (...args: Parameters<T>) => void {
  let inThrottle = false
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args)
      inThrottle = true
      setTimeout(() => (inThrottle = false), limit)
    }
  }
}

// ---------- HTML 编解码 ----------
export function escapeHtml(str: string): string {
  const entities: Record<string, string> = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }
  return str.replace(/[&<>"']/g, c => entities[c] || c)
}

export function unescapeHtml(str: string): string {
  const entities: Record<string, string> = { '&amp;': '&', '&lt;': '<', '&gt;': '>', '&quot;': '"', '&#39;': "'" }
  return str.replace(/&(amp|lt|gt|quot|#39);/g, e => entities[e] || e)
}

// ---------- 剪贴板 ----------
export async function copyToClipboard(text: string): Promise<boolean> {
  try { await navigator.clipboard.writeText(text); return true } catch { return false }
}

// ============================================================
// 19 套主题定义 - 完全匹配 notepad-- styleset.cpp
// 参考: notepad-- src/styleset.h/.cpp, mystyle.qss
// ============================================================
export interface ThemeColors {
  bg: string
  fg: string
  gutterBg: string
  gutterFg: string
  selection: string
  activeLine: string
  cursor: string
  accent: string
  comment: string
  keyword: string
  string: string
  number: string
  type: string
  function: string
  variable: string
  operator: string
  sidebarBg: string
  sidebarFg: string
  menuBg: string
  menuFg: string
  toolbarBg: string
  statusBarBg: string
  statusBarFg: string
  tabActiveBg: string
  tabInactiveBg: string
  findHighlight: string
  bookmarkFg: string
  bracketMatch: string
  isDark: boolean
}

export const THEME_COLORS: Record<ThemeName, ThemeColors> = {
  'Default': {
    bg: '#ffffff', fg: '#333333', gutterBg: '#f5f5f5', gutterFg: '#999999',
    selection: '#a8cfff', activeLine: '#f0f0f0', cursor: '#333333', accent: '#3b82f6',
    comment: '#6a9955', keyword: '#0000ff', string: '#a31515', number: '#098658',
    type: '#267f99', function: '#795e26', variable: '#001080', operator: '#000000',
    sidebarBg: '#f0f0f0', sidebarFg: '#333333', menuBg: '#f0f0f0', menuFg: '#333333',
    toolbarBg: '#f0f0f0', statusBarBg: '#007acc', statusBarFg: '#ffffff',
    tabActiveBg: '#ffffff', tabInactiveBg: '#e5e5e5', findHighlight: '#ffff00',
    bookmarkFg: '#007acc', bracketMatch: '#a8cfff',
    isDark: false,
  },
  'DarkDefault': {
    bg: '#1e1e1e', fg: '#d4d4d4', gutterBg: '#252526', gutterFg: '#858585',
    selection: '#264f78', activeLine: '#2d2d2d', cursor: '#d4d4d4', accent: '#60a5fa',
    comment: '#6a9955', keyword: '#569cd6', string: '#ce9178', number: '#b5cea8',
    type: '#4ec9b0', function: '#dcdcaa', variable: '#9cdcfe', operator: '#d4d4d4',
    sidebarBg: '#2d2d2d', sidebarFg: '#d4d4d4', menuBg: '#2d2d2d', menuFg: '#d4d4d4',
    toolbarBg: '#2d2d2d', statusBarBg: '#007acc', statusBarFg: '#ffffff',
    tabActiveBg: '#1e1e1e', tabInactiveBg: '#2d2d2d', findHighlight: '#ffff00',
    bookmarkFg: '#60a5fa', bracketMatch: '#264f78',
    isDark: true,
  },
  'Monokai': {
    bg: '#272822', fg: '#f8f8f2', gutterBg: '#1e1f1c', gutterFg: '#75715e',
    selection: '#49483e', activeLine: '#3e3d32', cursor: '#f8f8f0', accent: '#a6e22e',
    comment: '#75715e', keyword: '#f92672', string: '#e6db74', number: '#ae81ff',
    type: '#66d9ef', function: '#a6e22e', variable: '#f8f8f2', operator: '#f92672',
    sidebarBg: '#1e1f1c', sidebarFg: '#f8f8f2', menuBg: '#1e1f1c', menuFg: '#f8f8f2',
    toolbarBg: '#1e1f1c', statusBarBg: '#a6e22e', statusBarFg: '#272822',
    tabActiveBg: '#272822', tabInactiveBg: '#1e1f1c', findHighlight: '#ae81ff',
    bookmarkFg: '#66d9ef', bracketMatch: '#49483e',
    isDark: true,
  },
  'Solarized': {
    bg: '#fdf6e3', fg: '#657b83', gutterBg: '#eee8d5', gutterFg: '#93a1a1',
    selection: '#eee8d5', activeLine: '#f5e9cf', cursor: '#657b83', accent: '#268bd2',
    comment: '#93a1a1', keyword: '#859900', string: '#2aa198', number: '#d33682',
    type: '#b58900', function: '#268bd2', variable: '#657b83', operator: '#657b83',
    sidebarBg: '#eee8d5', sidebarFg: '#657b83', menuBg: '#eee8d5', menuFg: '#657b83',
    toolbarBg: '#eee8d5', statusBarBg: '#268bd2', statusBarFg: '#ffffff',
    tabActiveBg: '#fdf6e3', tabInactiveBg: '#eee8d5', findHighlight: '#b58900',
    bookmarkFg: '#dc322f', bracketMatch: '#eee8d5',
    isDark: false,
  },
  'SolarizedDark': {
    bg: '#002b36', fg: '#839496', gutterBg: '#073642', gutterFg: '#586e75',
    selection: '#073642', activeLine: '#073642', cursor: '#839496', accent: '#268bd2',
    comment: '#586e75', keyword: '#859900', string: '#2aa198', number: '#d33682',
    type: '#b58900', function: '#268bd2', variable: '#839496', operator: '#839496',
    sidebarBg: '#073642', sidebarFg: '#839496', menuBg: '#073642', menuFg: '#839496',
    toolbarBg: '#073642', statusBarBg: '#268bd2', statusBarFg: '#ffffff',
    tabActiveBg: '#002b36', tabInactiveBg: '#073642', findHighlight: '#b58900',
    bookmarkFg: '#dc322f', bracketMatch: '#073642',
    isDark: true,
  },
  'Bespin': {
    bg: '#28211c', fg: '#8a8986', gutterBg: '#1f1814', gutterFg: '#666666',
    selection: '#4d3b33', activeLine: '#3a2e27', cursor: '#baae9e', accent: '#cf6a4c',
    comment: '#937121', keyword: '#cf6a4c', string: '#f9ee98', number: '#cf7d34',
    type: '#9b859d', function: '#5ea6ea', variable: '#8a8986', operator: '#937121',
    sidebarBg: '#1f1814', sidebarFg: '#8a8986', menuBg: '#1f1814', menuFg: '#8a8986',
    toolbarBg: '#1f1814', statusBarBg: '#cf6a4c', statusBarFg: '#28211c',
    tabActiveBg: '#28211c', tabInactiveBg: '#1f1814', findHighlight: '#f9ee98',
    bookmarkFg: '#5ea6ea', bracketMatch: '#4d3b33',
    isDark: true,
  },
  'BlackBoard': {
    bg: '#0c1021', fg: '#f8f8f8', gutterBg: '#0a0e1a', gutterFg: '#555555',
    selection: '#253b76', activeLine: '#0a0e1a', cursor: '#f8f8f8', accent: '#8da6ce',
    comment: '#aeaeae', keyword: '#fbde2d', string: '#61ce3c', number: '#d8fa3c',
    type: '#8da6ce', function: '#ff6400', variable: '#f8f8f8', operator: '#fbde2d',
    sidebarBg: '#0a0e1a', sidebarFg: '#f8f8f8', menuBg: '#0a0e1a', menuFg: '#f8f8f8',
    toolbarBg: '#0a0e1a', statusBarBg: '#fbde2d', statusBarFg: '#0c1021',
    tabActiveBg: '#0c1021', tabInactiveBg: '#0a0e1a', findHighlight: '#d8fa3c',
    bookmarkFg: '#8da6ce', bracketMatch: '#253b76',
    isDark: true,
  },
  'Cobalt': {
    bg: '#002240', fg: '#ffffff', gutterBg: '#001b33', gutterFg: '#3e7087',
    selection: '#b36539', activeLine: '#003054', cursor: '#ffffff', accent: '#ff9d00',
    comment: '#0088ff', keyword: '#ff9d00', string: '#3ad900', number: '#ff628c',
    type: '#ff80e1', function: '#ffee80', variable: '#ffffff', operator: '#ff9d00',
    sidebarBg: '#001b33', sidebarFg: '#ffffff', menuBg: '#001b33', menuFg: '#ffffff',
    toolbarBg: '#001b33', statusBarBg: '#ff9d00', statusBarFg: '#002240',
    tabActiveBg: '#002240', tabInactiveBg: '#001b33', findHighlight: '#ffee80',
    bookmarkFg: '#3ad900', bracketMatch: '#b36539',
    isDark: true,
  },
  'Dracula': {
    bg: '#282a36', fg: '#f8f8f2', gutterBg: '#21222c', gutterFg: '#6272a4',
    selection: '#44475a', activeLine: '#343746', cursor: '#f8f8f2', accent: '#bd93f9',
    comment: '#6272a4', keyword: '#ff79c6', string: '#f1fa8c', number: '#bd93f9',
    type: '#8be9fd', function: '#50fa7b', variable: '#f8f8f2', operator: '#ff79c6',
    sidebarBg: '#21222c', sidebarFg: '#f8f8f2', menuBg: '#21222c', menuFg: '#f8f8f2',
    toolbarBg: '#21222c', statusBarBg: '#bd93f9', statusBarFg: '#282a36',
    tabActiveBg: '#282a36', tabInactiveBg: '#21222c', findHighlight: '#f1fa8c',
    bookmarkFg: '#50fa7b', bracketMatch: '#44475a',
    isDark: true,
  },
  'Eiffel': {
    bg: '#ffffff', fg: '#000000', gutterBg: '#f0f0f0', gutterFg: '#b0b0b0',
    selection: '#c3dcff', activeLine: '#f5f5f5', cursor: '#000000', accent: '#0000ff',
    comment: '#00b418', keyword: '#0100b6', string: '#cc0081', number: '#cc0081',
    type: '#0100b6', function: '#000000', variable: '#0206ff', operator: '#000000',
    sidebarBg: '#f0f0f0', sidebarFg: '#000000', menuBg: '#f0f0f0', menuFg: '#000000',
    toolbarBg: '#f0f0f0', statusBarBg: '#0100b6', statusBarFg: '#ffffff',
    tabActiveBg: '#ffffff', tabInactiveBg: '#e5e5e5', findHighlight: '#00b418',
    bookmarkFg: '#cc0081', bracketMatch: '#c3dcff',
    isDark: false,
  },
  'Elegant': {
    bg: '#ffffff', fg: '#333333', gutterBg: '#f8f8f8', gutterFg: '#bbbbbb',
    selection: '#bad6fd', activeLine: '#f5f7fa', cursor: '#333333', accent: '#174781',
    comment: '#999988', keyword: '#174781', string: '#c03030', number: '#174781',
    type: '#174781', function: '#333333', variable: '#333333', operator: '#174781',
    sidebarBg: '#f8f8f8', sidebarFg: '#333333', menuBg: '#f8f8f8', menuFg: '#333333',
    toolbarBg: '#f8f8f8', statusBarBg: '#174781', statusBarFg: '#ffffff',
    tabActiveBg: '#ffffff', tabInactiveBg: '#efefef', findHighlight: '#c03030',
    bookmarkFg: '#174781', bracketMatch: '#bad6fd',
    isDark: false,
  },
  'ErlangDark': {
    bg: '#002235', fg: '#cccccc', gutterBg: '#001a2a', gutterFg: '#555555',
    selection: '#003d5e', activeLine: '#002a40', cursor: '#ffffff', accent: '#66cccc',
    comment: '#00aaaa', keyword: '#cb4d16', string: '#00cc66', number: '#00aacc',
    type: '#66cccc', function: '#ff6699', variable: '#cccccc', operator: '#cb4d16',
    sidebarBg: '#001a2a', sidebarFg: '#cccccc', menuBg: '#001a2a', menuFg: '#cccccc',
    toolbarBg: '#001a2a', statusBarBg: '#66cccc', statusBarFg: '#002235',
    tabActiveBg: '#002235', tabInactiveBg: '#001a2a', findHighlight: '#00cc66',
    bookmarkFg: '#ff6699', bracketMatch: '#003d5e',
    isDark: true,
  },
  'IDLE': {
    bg: '#ffffff', fg: '#000000', gutterBg: '#f0f0f0', gutterFg: '#808080',
    selection: '#8690ff', activeLine: '#f5f5f5', cursor: '#000000', accent: '#8080ff',
    comment: '#919191', keyword: '#ff5600', string: '#00a33f', number: '#ff5600',
    type: '#21439c', function: '#a535ae', variable: '#000000', operator: '#ff5600',
    sidebarBg: '#f0f0f0', sidebarFg: '#000000', menuBg: '#f0f0f0', menuFg: '#000000',
    toolbarBg: '#f0f0f0', statusBarBg: '#21439c', statusBarFg: '#ffffff',
    tabActiveBg: '#ffffff', tabInactiveBg: '#e5e5e5', findHighlight: '#00a33f',
    bookmarkFg: '#ff5600', bracketMatch: '#8690ff',
    isDark: false,
  },
  'Lazy': {
    bg: '#ffffff', fg: '#000000', gutterBg: '#fafafa', gutterFg: '#a0a0a0',
    selection: '#d2e2f2', activeLine: '#f7f7f7', cursor: '#7c7c7c', accent: '#00aaff',
    comment: '#8c868f', keyword: '#3a66dd', string: '#37b349', number: '#ac54c6',
    type: '#3a66dd', function: '#61676b', variable: '#000000', operator: '#8c868f',
    sidebarBg: '#fafafa', sidebarFg: '#000000', menuBg: '#fafafa', menuFg: '#000000',
    toolbarBg: '#fafafa', statusBarBg: '#3a66dd', statusBarFg: '#ffffff',
    tabActiveBg: '#ffffff', tabInactiveBg: '#eeeeee', findHighlight: '#37b349',
    bookmarkFg: '#ac54c6', bracketMatch: '#d2e2f2',
    isDark: false,
  },
  'Material': {
    bg: '#263238', fg: '#eeffff', gutterBg: '#1e2a30', gutterFg: '#546e7a',
    selection: '#314549', activeLine: '#2c3b42', cursor: '#ffcc00', accent: '#82aaff',
    comment: '#546e7a', keyword: '#c792ea', string: '#c3e88d', number: '#f78c6c',
    type: '#82aaff', function: '#82aaff', variable: '#eeffff', operator: '#89ddff',
    sidebarBg: '#1e2a30', sidebarFg: '#eeffff', menuBg: '#1e2a30', menuFg: '#eeffff',
    toolbarBg: '#1e2a30', statusBarBg: '#c792ea', statusBarFg: '#263238',
    tabActiveBg: '#263238', tabInactiveBg: '#1e2a30', findHighlight: '#c3e88d',
    bookmarkFg: '#f78c6c', bracketMatch: '#314549',
    isDark: true,
  },
  'MonoIndustrial': {
    bg: '#222c28', fg: '#e0e0e0', gutterBg: '#1b2420', gutterFg: '#668866',
    selection: '#3a4a42', activeLine: '#283530', cursor: '#ffffff', accent: '#aaccbb',
    comment: '#556b5a', keyword: '#99bb88', string: '#88cc66', number: '#66aacc',
    type: '#aaccbb', function: '#ddaa66', variable: '#e0e0e0', operator: '#99bb88',
    sidebarBg: '#1b2420', sidebarFg: '#e0e0e0', menuBg: '#1b2420', menuFg: '#e0e0e0',
    toolbarBg: '#1b2420', statusBarBg: '#aaccbb', statusBarFg: '#222c28',
    tabActiveBg: '#222c28', tabInactiveBg: '#1b2420', findHighlight: '#88cc66',
    bookmarkFg: '#66aacc', bracketMatch: '#3a4a42',
    isDark: true,
  },
  'Neat': {
    bg: '#ffffff', fg: '#444444', gutterBg: '#f5f5f5', gutterFg: '#aaaaaa',
    selection: '#d6e5f2', activeLine: '#f4f7fa', cursor: '#444444', accent: '#2266bb',
    comment: '#888888', keyword: '#2266bb', string: '#448844', number: '#886622',
    type: '#2266bb', function: '#664488', variable: '#444444', operator: '#2266bb',
    sidebarBg: '#f5f5f5', sidebarFg: '#444444', menuBg: '#f5f5f5', menuFg: '#444444',
    toolbarBg: '#f5f5f5', statusBarBg: '#2266bb', statusBarFg: '#ffffff',
    tabActiveBg: '#ffffff', tabInactiveBg: '#eaeaea', findHighlight: '#448844',
    bookmarkFg: '#664488', bracketMatch: '#d6e5f2',
    isDark: false,
  },
  'NightOwl': {
    bg: '#011627', fg: '#d6deeb', gutterBg: '#010e1a', gutterFg: '#334455',
    selection: '#1d3b53', activeLine: '#01111f', cursor: '#80a4c2', accent: '#82aaff',
    comment: '#637777', keyword: '#c792ea', string: '#ecc48d', number: '#f78c6c',
    type: '#7fdbca', function: '#82aaff', variable: '#d6deeb', operator: '#c792ea',
    sidebarBg: '#010e1a', sidebarFg: '#d6deeb', menuBg: '#010e1a', menuFg: '#d6deeb',
    toolbarBg: '#010e1a', statusBarBg: '#c792ea', statusBarFg: '#011627',
    tabActiveBg: '#011627', tabInactiveBg: '#010e1a', findHighlight: '#ecc48d',
    bookmarkFg: '#7fdbca', bracketMatch: '#1d3b53',
    isDark: true,
  },
  'OneDark': {
    bg: '#282c34', fg: '#abb2bf', gutterBg: '#21252b', gutterFg: '#636d83',
    selection: '#404859', activeLine: '#2c313c', cursor: '#528bff', accent: '#61afef',
    comment: '#5c6370', keyword: '#c678dd', string: '#98c379', number: '#d19a66',
    type: '#e5c07b', function: '#61afef', variable: '#e06c75', operator: '#56b6c2',
    sidebarBg: '#21252b', sidebarFg: '#abb2bf', menuBg: '#21252b', menuFg: '#abb2bf',
    toolbarBg: '#21252b', statusBarBg: '#c678dd', statusBarFg: '#282c34',
    tabActiveBg: '#282c34', tabInactiveBg: '#21252b', findHighlight: '#98c379',
    bookmarkFg: '#e06c75', bracketMatch: '#404859',
    isDark: true,
  },
}

// ---------- 获取当前主题的颜色 ----------
export function getThemeColors(theme: ThemeName): ThemeColors {
  return THEME_COLORS[theme] || THEME_COLORS['Default']
}

// ---------- 获取文件图标 ----------
export function getFileIcon(ext: string, isDir: boolean): string {
  if (isDir) return 'folder'
  const iconMap: Record<string, string> = {
    'js': 'file-code', 'jsx': 'file-code', 'ts': 'file-code', 'tsx': 'file-code',
    'json': 'file-json', 'html': 'file-code', 'css': 'file-code',
    'md': 'file-text', 'markdown': 'file-text', 'txt': 'file-text',
    'py': 'file-code', 'java': 'file-code', 'go': 'file-code', 'sql': 'database',
    'xml': 'file-code', 'yaml': 'file-code', 'yml': 'file-code',
    'sh': 'terminal', 'bash': 'terminal', 'log': 'file-text',
    'png': 'image', 'jpg': 'image', 'jpeg': 'image', 'gif': 'image', 'svg': 'image',
    'pdf': 'file-text', 'zip': 'archive',
  }
  return iconMap[ext.toLowerCase()] || 'file'
}
