/**
 * useEditorLanguage — 编辑器语言扩展加载
 *
 * 职责（与原 CodeEditor.vue 完全等价搬运，无行为变更）：
 *  - buildStreamLanguage() ：基于关键词/类型/常量的简易 Stream lexer
 *  - streamLangs          ：30+ 种小语种定义（rust / c / cpp / lua / ruby / shell …）
 *  - getLanguageExtension()：内置 langs（json/ts/html/css/md/...）+ stream langs + JS 回退
 *
 * 设计目标：纯工厂函数，不持有可变状态。容器在 createEditor() 时调一次。
 * 语言列表属于纯数据/算法，不依赖 store；可独立测试。
 */
import { StreamLanguage, LanguageSupport } from '@codemirror/language'
import { json } from '@codemirror/lang-json'
import { javascript } from '@codemirror/lang-javascript'
import { html, autoCloseTags } from '@codemirror/lang-html'
import { css } from '@codemirror/lang-css'
import { markdown } from '@codemirror/lang-markdown'
import { xml } from '@codemirror/lang-xml'
import { yaml } from '@codemirror/lang-yaml'
import { python } from '@codemirror/lang-python'
import { java } from '@codemirror/lang-java'
import { go } from '@codemirror/lang-go'
import { sql } from '@codemirror/lang-sql'

/**
 * 基于关键词的简易 Stream lexer（与原 CodeEditor 等价搬运）。
 * 适用于 CodeMirror 无官方 language pack 的小语种。
 */
function buildStreamLanguage(
  name: string,
  keywords: string,
  types: string = '',
  constants: string = '',
  lineComment: string = '//',
  blockComment: string[] = ['/*', '*/'],
  strings: string[][] = [['"', '"']],
): LanguageSupport {
  const kwSet = new Set((keywords + ' ' + types + ' ' + constants).toLowerCase().split(/\s+/).filter(Boolean))
  return new LanguageSupport(StreamLanguage.define({
    name,
    token(stream, state: any) {
      // Skip whitespace
      if (stream.eatSpace()) return null
      // Line comment
      if (stream.match(lineComment)) { stream.skipToEnd(); return 'lineComment' }
      // Block comment
      if (blockComment.length >= 2 && stream.match(blockComment[0])) {
        while (!stream.eol()) {
          if (stream.match(blockComment[1])) return 'blockComment'
          stream.next()
        }
        return 'blockComment'
      }
      // Strings
      for (const [open, close] of strings) {
        if (stream.match(open)) {
          while (!stream.eol()) {
            if (stream.match('\\' + close)) { stream.next(); continue }
            if (stream.match(close)) return 'string'
            stream.next()
          }
          return 'string'
        }
      }
      // Numbers
      if (stream.match(/0[xX][0-9a-fA-F]+/) || stream.match(/0[oO][0-7]+/) ||
          stream.match(/0[bB][01]+/) || stream.match(/\d+\.?\d*/))
        return 'number'
      // Identifiers & keywords
      if (stream.match(/[a-zA-Z_$][\w$]*/)) {
        const word = stream.current().toLowerCase()
        if (kwSet.has(word)) return 'keyword'
        if (word === 'true' || word === 'false' || word === 'null' || word === 'nil' || word === 'none') return 'atom'
        return 'variableName'
      }
      // Operators
      if (stream.match(/[+\-*/%&|^~!=<>?:;,.\[\]{}()#@`]+/)) return 'operator'
      stream.next()
      return null
    },
    languageData: {
      commentTokens: { line: lineComment, block: blockComment.length >= 2 ? { open: blockComment[0], close: blockComment[1] } : undefined },
      indentOnInput: /^\s*[\}\]]$/,
    },
  }), [])
}

// Stream language definitions matching notepad-- lexer set
const streamLangs: Record<string, () => LanguageSupport> = {
  rust: () => buildStreamLanguage('rust',
    'as async await break const continue crate dyn else enum extern false fn for if impl in let loop macro match mod move mut pub ref return self static struct super trait true type unsafe use where while',
    'bool char f32 f64 i8 i16 i32 i64 i128 isize str String u8 u16 u32 u64 u128 usize Vec Option Result',
    'Some None Ok Err', '//', ['/*', '*/'], [['"', '"'], ["r#\"", "\"#"]]
  ),
  c: () => buildStreamLanguage('c',
    'auto break case char const continue default do double else enum extern float for goto if int long register return short signed sizeof static struct switch typedef union unsigned void volatile while',
    'int8_t int16_t int32_t int64_t uint8_t uint16_t uint32_t uint64_t size_t ssize_t bool FILE NULL',
    'true false NULL', '//', ['/*', '*/']
  ),
  cpp: () => buildStreamLanguage('cpp',
    'alignas alignof and and_eq asm auto bitand bitor bool break case catch char class compl const constexpr const_cast continue decltype default delete do double dynamic_cast else enum explicit export extern false float for friend goto if inline int long mutable namespace new noexcept not not_eq nullptr operator or or_eq override private protected public register reinterpret_cast return short signed sizeof static static_assert static_cast struct switch template this thread_local throw true try typedef typeid typename union unsigned using virtual void volatile wchar_t while xor xor_eq',
    'string vector map set pair shared_ptr unique_ptr istream ostream fstream', 'nullptr true false', '//', ['/*', '*/']
  ),
  csharp: () => buildStreamLanguage('csharp',
    'abstract as base bool break byte case catch char checked class const continue decimal default delegate do double else enum event explicit extern false finally fixed float for foreach goto if implicit in int interface internal is lock long namespace new null object operator out override params private protected public readonly ref return sbyte sealed short sizeof stackalloc static string struct switch this throw true try typeof uint ulong unchecked unsafe ushort using var virtual void volatile while',
    'string int long float double decimal bool char void object var dynamic', 'null true false',
    '//', ['/*', '*/'], [['"', '"'], ["@\"", "\""], ["'", "'"]]
  ),
  php: () => buildStreamLanguage('php',
    'abstract and array as break callable case catch class clone const continue declare default die do echo else elseif empty enddeclare endfor endforeach endif endswitch endwhile eval exit extends final finally fn for foreach function global goto if implements include include_once instanceof insteadof interface isset list match namespace new or print private protected public readonly require require_once return static switch throw trait try unset use var while xor yield',
    'int float string bool array object null mixed void callable iterable self static parent', 'true false null',
    '//', ['/*', '*/'], [['"', '"'], ["'", "'"]]
  ),
  ruby: () => buildStreamLanguage('ruby',
    'BEGIN END alias and begin break case class def defined? do else elsif end ensure false for if in module next nil not or redo rescue retry return self super then true undef unless until when while yield',
    'String Integer Float Array Hash Symbol', 'true false nil', '#', ['=begin', '=end'], [['"', '"'], ["'", "'"]]
  ),
  lua: () => buildStreamLanguage('lua',
    'and break do else elseif end false for function goto if in local nil not or repeat return then true until while',
    'string number boolean table function thread userdata', 'true false nil', '--', ['--[[', ']]'], [['"', '"'], ["'", "'"]]
  ),
  kotlin: () => buildStreamLanguage('kotlin',
    'abstract actual annotation as as? break case catch class companion const continue crossinline data delegate do dynamic else enum expect external false field file final finally for fun get if import in infix init inline inner interface internal is lateinit noinline null object open operator out override package param private property protected public receiverisinline reified return sealed set super suspend tailrec this throw true try typealias typeof val var vararg when where while',
    'Int Long Float Double Boolean String Char Byte Short Any Unit Nothing', 'true false null', '//', ['/*', '*/']
  ),
  scala: () => buildStreamLanguage('scala',
    'abstract case catch class def do else extends false final finally for forSome if implicit import lazy match new null object override package private protected return sealed super this throw trait try true type val var while with yield',
    'Int Long Float Double Boolean String Char Byte Short Any Unit Nothing Option List Map Set', 'true false null', '//', ['/*', '*/']
  ),
  clojure: () => buildStreamLanguage('clojure',
    'def defn fn let if when loop recur doseq for map reduce filter', '', 'true false nil', ';', [], [['"', '"']]
  ),
  dart: () => buildStreamLanguage('dart',
    'abstract as assert async await break case catch class const continue covariant default deferred do dynamic else enum export extends extension external factory false final finally for Function get hide if implements import in interface is late library mixin new null on operator part required rethrow return set show static super switch sync this throw true try typedef var void when while with yield',
    'int double String bool List Map Set dynamic void num', 'true false null', '//', ['/*', '*/'], [['"', '"'], ["'", "'"]]
  ),
  swift: () => buildStreamLanguage('swift',
    'associatedtype async await break case catch class continue convenience default defer deinit didSet do dynamic else enum extension fallthrough false fileprivate final for func get guard if import in indirect infix init inout internal is lazy let mutating nil nonmutating open operator optional override postfix precedence prefix private protocol public repeat required rethrows return self set some static struct subscript super switch throw throws true try typealias unowned var weak where while willSet',
    'Int UInt Float Double Bool String Character Array Dictionary Set Optional Void Any', 'true false nil', '//', ['/*', '*/']
  ),
  elixir: () => buildStreamLanguage('elixir',
    'after and catch def defmodule defp do else end false fn for if import in nil not or raise rescue try unless use when',
    'integer float boolean atom list tuple map pid port reference', 'true false nil', '#', [], [['"', '"'], ["'", "'"]]
  ),
  erlang: () => buildStreamLanguage('erlang',
    'after and andalso band begin bnot bor bsl bsr bxor case catch cond div end fun if let not of or orelse receive rem try when xor',
    'integer float atom pid port reference', 'true false', '%', [], [['"', '"']]
  ),
  haskell: () => buildStreamLanguage('haskell',
    'as case of class data default deriving do else forall foreign hiding if import in infix infixl infixr instance let mdo module newtype of qualified then type where',
    'Int Integer Float Double Bool Char String IO Maybe Either', 'True False', '--', ['{-', '-}']
  ),
  julia: () => buildStreamLanguage('julia',
    'abstract begin break catch const continue do else elseif end export finally for function global if import let local macro module mutable outer primitive quote return struct try using while',
    'Int Int8 Int16 Int32 Int64 Int128 Float16 Float32 Float64 Bool Char String Array Dict Set Tuple Symbol', 'true false nothing', '#', ['#=', '=#']
  ),
  shell: () => buildStreamLanguage('shell',
    'if then else elif fi case esac for while until do done in function return exit break continue select local export readonly declare typeset unset alias',
    '', 'true false', '#', [], [['"', '"'], ["'", "'"]]
  ),
  powershell: () => buildStreamLanguage('powershell',
    'Begin Break Catch Continue Data Do DynamicParam Else ElseIf End Exit Filter Finally For ForEach From Function If In InlineScript Param Process Return Switch Throw Trap Try Until Using Var While Workflow',
    'string int long bool array hashtable psobject scriptblock', '$true $false $null', '#', ['<#', '#>'], [['"', '"'], ["'", "'"]]
  ),
  perl: () => buildStreamLanguage('perl',
    'if else elsif unless while for foreach continue do require use my our local sub return last next redo goto die warn eval', '', '', '#', [], [['"', '"'], ["'", "'"]]
  ),
  r: () => buildStreamLanguage('r',
    'if else for while repeat break next function return in NULL NA NaN Inf TRUE FALSE', '', 'TRUE FALSE NULL NA', '#', [], [['"', '"']]
  ),
  latex: () => buildStreamLanguage('latex', '', '', '', '%', [], []),
  batch: () => buildStreamLanguage('batch',
    'call echo set if else for goto pause exit rem start cd md rd del copy move ren type cls title color path ver', '', '', 'REM', [], []
  ),
  toml: () => buildStreamLanguage('toml', '', '', 'true false', '#', [], [['"', '"']]),
  ini: () => buildStreamLanguage('ini', '', '', '', ';', [], [['"', '"']]),
  cmake: () => buildStreamLanguage('cmake',
    'if else elseif endif foreach endforeach while endwhile function endfunction macro endmacro break continue return set unset list string file message option', '', 'TRUE FALSE ON OFF YES NO', '#', [])
  ,
  protobuf: () => buildStreamLanguage('protobuf',
    'syntax package import option message enum service rpc returns repeated optional required oneof map extensions reserved', '', 'true false', '//', ['/*', '*/']
  ),
  graphql: () => buildStreamLanguage('graphql',
    'query mutation subscription fragment on implements interface union enum input type schema scalar extend directive', '', 'true false null', '#', [], [['"', '"']]
  ),
  dockerfile: () => buildStreamLanguage('dockerfile',
    'FROM RUN CMD LABEL EXPOSE ENV ADD COPY ENTRYPOINT VOLUME USER WORKDIR ARG ONBUILD STOPSIGNAL HEALTHCHECK SHELL MAINTAINER', '', '', '#', [], [['"', '"']]
  ),
  zig: () => buildStreamLanguage('zig',
    'align and anytype asm async await break catch comptime const continue defer else enum errdefer error export extern fn for if inline linksection noalias nosuspend or orelse packed pub resume return struct suspend switch test threadlocal try union unreachable usingnamespace var volatile while',
    'bool f16 f32 f64 f80 f128 i8 i16 i32 i64 i128 isize u8 u16 u32 u64 u128 usize void noreturn type anyerror comptime_int comptime_float', 'true false null undefined', '//', [], [['"', '"']]
  ),
  solidity: () => buildStreamLanguage('solidity',
    'abstract after catch contract enum event function interface is library mapping modifier override pragma private public pure returns storage struct view', 'address bool bytes int uint string mapping', 'true false', '//', ['/*', '*/']
  ),
  pascal: () => buildStreamLanguage('pascal',
    'and array begin case const div do downto else end file for function goto if in label mod nil not of or packed procedure program record repeat set then to type until var while with',
    'integer real boolean char string', 'true false nil', '//', ['{', '}'], [["'", "'"]]
  ),
  fortran: () => buildStreamLanguage('fortran',
    'allocatable allocate call case contains continue cycle deallocate do else end function if implicit in integer module none nullify only parameter pointer private program public real recursive result return save select stop subroutine then type use where while',
    '.true. .false.', '.true. .false.', '!', [], [["'", "'"]]
  ),
  cobol: () => buildStreamLanguage('cobol',
    'ACCEPT ADD CALL CLOSE COMPUTE DELETE DISPLAY DIVIDE EVALUATE GO GOBACK IF INITIALIZE MERGE MOVE MULTIPLY OPEN PERFORM READ RETURN REWRITE SEARCH SET SORT START STOP STRING SUBTRACT UNSTRING WRITE',
    'PIC X 9 A V S COMP COMP-3', '', '*', [], [['"', '"'], ["'", "'"]]
  ),
  tcl: () => buildStreamLanguage('tcl', 'if else elseif for foreach while switch proc return set upvar global variable', '', '', '#', [], [['"', '"']]),
  scheme: () => buildStreamLanguage('scheme',
    'define lambda let let* letrec if cond else begin set! quote quasiquote unquote unquote-splicing do delay force', '', '#t #f', ';', [], [['"', '"']]
  ),
  smalltalk: () => buildStreamLanguage('smalltalk',
    'self super nil true false', '', 'true false nil', '"', [], [["'", "'"]]
  ),
  prolog: () => buildStreamLanguage('prolog', ':- ! true fail not is', '', 'true fail', '%', ['/*', '*/']),
  ada: () => buildStreamLanguage('ada',
    'abort abs accept access all and array at begin body case constant declare delay delta digits do else elsif end entry exception exit for function generic goto if in is limited loop mod new not null of or others out package pragma private procedure raise range record rem renames return reverse select separate subtype task terminate then type use when while with xor',
    'Integer Float Boolean Character String Duration', 'True False', '--', [], [['"', '"']]
  ),
  nsis: () => buildStreamLanguage('nsis', 'Function FunctionEnd Section SectionEnd SetOutPath File WriteRegStr ReadRegStr DeleteRegKey IfErrors MessageBox DetailPrint StrCpy IntCmp IntCmpU StrCmp Exch Pop Push Call nsDialogs Create', '', '', ';', ['/*', '*/']),
  assembly: () => buildStreamLanguage('assembly', 'mov push pop call ret jmp je jne jg jl cmp add sub mul div inc dec xor and or not shl shr lea int nop', 'eax ebx ecx edx esi edi esp ebp rax rbx rcx rdx', '', ';', [], [['"', '"'], ["'", "'"]]),
  diff: () => new LanguageSupport(StreamLanguage.define({
    name: 'diff', token(stream) {
      if (stream.sol() && stream.match(/^---/)) { stream.skipToEnd(); return 'lineComment' }
      if (stream.sol() && stream.match(/^\+\+\+/)) { stream.skipToEnd(); return 'lineComment' }
      if (stream.sol() && stream.match(/^@@/)) { stream.skipToEnd(); return 'keyword' }
      if (stream.sol() && stream.match(/^\+/)) { stream.skipToEnd(); return 'string' }
      if (stream.sol() && stream.match(/^-/)) { stream.skipToEnd(); return 'keyword' }
      stream.skipToEnd(); return null
    },
  }), []),
}

/** 内置官方 language pack 映射 */
const builtInLangs: Record<string, () => any[]> = {
  'json': () => [json()],
  'javascript': () => [javascript()],
  'typescript': () => [javascript({ typescript: true })],
  'html': () => [html(), autoCloseTags],
  'css': () => [css()],
  'markdown': () => [markdown()],
  'xml': () => [xml(), autoCloseTags],
  'yaml': () => [yaml()],
  'python': () => [python()],
  'java': () => [java()],
  'go': () => [go()],
  'sql': () => [sql()],
  // Use javascript as fallback for C-like languages that work reasonably well
  'vue': () => [html()], // HTML handles Vue templates
  'svelte': () => [html()],
  'scss': () => [css()],
  'sass': () => [css()],
  'less': () => [css()],
  'csv': () => [], // plain text
  'text': () => [],
}

export interface UseEditorLanguage {
  /** 根据语言标识返回 CodeMirror extension 数组 */
  getLanguageExtension: (lang: string) => any[]
}

/**
 * 当前实现保持原 CodeEditor 一致的语言解析策略：
 *   1. 内置官方 pack（json / js / html / ...）
 *   2. 自定义 StreamLanguage（rust / c / cpp / lua / shell / ...）
 *   3. 未知 C-like 语言回退到 javascript
 *   4. 仍未匹配则返回 []（纯文本）
 */
export function useEditorLanguage(): UseEditorLanguage {
  function getLanguageExtension(lang: string): any[] {
    if (lang in builtInLangs) return builtInLangs[lang]()
    if (lang in streamLangs) return [streamLangs[lang]()]
    if (['objectivec', 'objectivecpp', 'd', 'nim', 'v', 'coffeescript'].includes(lang))
      return [javascript()]
    return []
  }
  return { getLanguageExtension }
}