// 函数列表面板的符号提取：基于正则的轻量实现，按语言提取函数/类/方法定义

export interface SymbolInfo {
  line: number // 1-based 行号
  name: string
  kind: 'function' | 'class'
}

const C_LIKE = ['c', 'c++', 'cpp', 'csharp', 'c#', 'java', 'kotlin', 'rust', 'swift', 'php', 'd', 'objective c', 'objective-c', 'vb', 'visual basic']

export function extractSymbols(content: string, language: string): SymbolInfo[] {
  if (!content) return []
  const lang = (language || '').toLowerCase()
  const lines = content.split('\n')
  const symbols: SymbolInfo[] = []

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]
    const ln = i + 1

    if (lang === 'python') {
      const m = line.match(/^\s*(def|class)\s+([A-Za-z_]\w*)/)
      if (m) symbols.push({ line: ln, name: m[2], kind: m[1] === 'class' ? 'class' : 'function' })
      continue
    }

    if (lang === 'go' || lang === 'golang') {
      let m = line.match(/^func\s+(?:\(\s*\w+\s+\*?\w*\s*\)\s*)?([A-Za-z_]\w*)\s*\(/)
      if (m) symbols.push({ line: ln, name: m[1], kind: 'function' })
      m = line.match(/^type\s+([A-Za-z_]\w*)\s+(struct|interface)/)
      if (m) symbols.push({ line: ln, name: m[1], kind: 'class' })
      continue
    }

    if (['javascript', 'typescript', 'js', 'ts', 'jsx', 'tsx'].includes(lang)) {
      let m = line.match(/^\s*(?:export\s+)?(?:default\s+)?(?:async\s+)?function\s+\*?\s*([A-Za-z_$]\w*)/)
      if (m) { symbols.push({ line: ln, name: m[1], kind: 'function' }); continue }
      m = line.match(/^\s*(?:export\s+)?(?:const|let|var)\s+([A-Za-z_$]\w*)\s*=\s*(?:async\s*)?(?:\([^)]*\)|[A-Za-z_$]\w*)\s*=>/)
      if (m) { symbols.push({ line: ln, name: m[1], kind: 'function' }); continue }
      m = line.match(/^\s*(?:export\s+)?(?:default\s+)?(?:abstract\s+)?class\s+([A-Za-z_$]\w*)/)
      if (m) { symbols.push({ line: ln, name: m[1], kind: 'class' }); continue }
      m = line.match(/^\s{1,}(?:async\s+)?(?:static\s+)?(?:get\s+|set\s+)?([A-Za-z_$]\w*)\s*\([^)]*\)\s*\{/)
      if (m && !/^(if|for|while|switch|catch)$/.test(m[1])) symbols.push({ line: ln, name: m[1], kind: 'function' })
      continue
    }

    // 类 C 语言 / 未知语言（保守匹配）
    if (C_LIKE.includes(lang)) {
      let m = line.match(/^\s*(?:public|private|protected|static|final|inline|virtual|export|async|override|extern|constexpr|unsafe|abstract|sealed|internal)\s+(?:[\w:<>\[\],\s\*&]+?)\s+([A-Za-z_]\w*)\s*\([^;=]*\)\s*(?:const\s*)?\{?/)
      if (m && !/^(if|for|while|switch|catch|return|new|sizeof|typeof)\b/.test(m[1])) symbols.push({ line: ln, name: m[1], kind: 'function' })
      m = line.match(/^\s*(?:public\s+|private\s+|protected\s+|internal\s+)?(class|struct|interface|enum|trait)\s+([A-Za-z_]\w*)/)
      if (m) symbols.push({ line: ln, name: m[2], kind: 'class' })
    }
  }

  return symbols
}
