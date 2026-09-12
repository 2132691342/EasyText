// WARNING: 本文件由 scripts/gen-bindings.cjs 生成，请勿手改。
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
