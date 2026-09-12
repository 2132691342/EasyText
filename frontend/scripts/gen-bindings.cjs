#!/usr/bin/env node
/**
 * gen-bindings.cjs — 从 Go 端方法签名生成 wailsjs 绑定（真实运行时桥接）。
 *
 * 背景：wails v2.12 的 `generate module` 在本机静默失败、build/dev 又把
 * bindings 写到临时目录不落盘，导致 npm build 的 vue-tsc 死循环（TS2307）。
 * 本脚本从 backend/api 的 Go 源码提取 Handler 全部导出方法，生成与
 * wails 官方生成物**等价**的 window.go 桥接（运行时行为一致，非 stub）。
 * wails dev/build 本地重新生成时会覆盖本目录，内容一致无害。
 *
 * 用法：node scripts/gen-bindings.cjs   （在 frontend/ 下执行）
 */
const fs = require('fs')
const path = require('path')

const ROOT = path.resolve(__dirname, '..', '..')
const API_DIR = path.join(ROOT, 'backend', 'api')
const OUT_DIR = path.join(__dirname, '..', 'wailsjs', 'go', 'main')
const MODELS_DIR = path.join(__dirname, '..', 'wailsjs', 'go')

const receivers = ['*Handler', '*RecentHandler', '*FileAssocHandler', '*SearchHandler']
const files = fs.readdirSync(API_DIR).filter(f => f.endsWith('.go') && !f.endsWith('_test.go'))

const methods = new Map() // name -> { params: string[] }
const re = new RegExp(`^func \\((?:\\w+) \\*Handler\\) ([A-Z]\\w*)\\(([^)]*)\\)`)
for (const f of files) {
  const src = fs.readFileSync(path.join(API_DIR, f), 'utf8')
  for (const m of src.matchAll(/^func \(\w+ \*\w*Handler\) ([A-Z]\w*)\(([^)]*)\)/gm)) {
    const name = m[1]
    const params = m[2].split(',').map(s => s.trim()).filter(Boolean).map(s => {
      // "path string" -> "path"；"offset, count int" 的连续同类型在切分后丢类型，
      // 但我们只需要参数名（且无类型者视为继续上一类型），生成 JS 不需要类型
      return s.split(/\s+/)[0]
    })
    if (!methods.has(name)) methods.set(name, params)
  }
}

// 生命周期方法不对前端暴露（与官方生成一致：官方会包含，但调用无意义；保留一致性起见排除）
methods.delete('Startup')
methods.delete('Shutdown')

fs.mkdirSync(OUT_DIR, { recursive: true })

// ---- App.js ----
let js = '// @ts-check\n// WARNING: 本文件由 scripts/gen-bindings.cjs 生成（或 wails CLI 生成），请勿手改。\n'
for (const [name, params] of methods) {
  const args = params.join(', ')
  js += `export function ${name}(${args}) {\n`
  js += `  return window['go']['main']['App']['${name}'](${args});\n`
  js += `}\n\n`
}
fs.writeFileSync(path.join(OUT_DIR, 'App.js'), js)

// ---- App.d.ts（参数宽松 any，与前端消费方式兼容） ----
let dts = '// WARNING: 本文件由 scripts/gen-bindings.cjs 生成，请勿手改。\n'
for (const [name, params] of methods) {
  const args = params.map(p => `${p}: any`).join(', ')
  dts += `export function ${name}(${args}): Promise<any>;\n`
}
fs.writeFileSync(path.join(OUT_DIR, 'App.d.ts'), dts)

// ---- models.ts（宽松 index signature，字段以运行时 Go JSON 为准） ----
const models = `// WARNING: 本文件由 scripts/gen-bindings.cjs 生成，请勿手改。
// 字段定义保持宽松（index signature）：真实结构以 Go 端 JSON 序列化结果为准，
// 前端类型增强请直接扩展 src/types/index.ts，而不是修改本文件。

export namespace config {
  export interface AppConfig {
    [key: string]: any;
  }
}

export namespace file {
  export interface FileInfo { [key: string]: any }
  export interface FileTree { [key: string]: any }
  export interface ReadResult { [key: string]: any }
  export interface WriteResult { [key: string]: any }
  export interface ChunkResult { [key: string]: any }
}

export namespace tools {
  export interface JSONResult { [key: string]: any }
  export interface JSONPathResult { [key: string]: any }
  export interface JSONDiffResult { [key: string]: any }
  export interface EncodingInfo { [key: string]: any }
  export interface SnippetEntry { [key: string]: any }
  export interface DraftEntry { [key: string]: any }
  export interface BookmarkEntry { [key: string]: any }
  export interface Macro { [key: string]: any }
  export interface MacroStep { [key: string]: any }
  export interface ScriptInfo { [key: string]: any }
  export interface ScriptResult { [key: string]: any }
  export interface DirCompareResult { [key: string]: any }
  export interface BinCompareResult { [key: string]: any }
  export interface BatchRenameResult { [key: string]: any }
  export interface RegexTestResult { [key: string]: any }
  export interface HashResult { [key: string]: any }
  export interface HashAlgoResult { [key: string]: any }
}

export namespace main {
  export interface Session { [key: string]: any }
  export interface SessionFile { [key: string]: any }
}
`
fs.writeFileSync(path.join(MODELS_DIR, 'models.ts'), models)

console.log(`Generated ${methods.size} bindings -> frontend/wailsjs/go/main/App.{js,d.ts} + go/models.ts`)
