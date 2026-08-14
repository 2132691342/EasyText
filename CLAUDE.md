# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

EasyText is a Windows-first desktop document editor built with **Wails v2** (Go backend + system WebView2) and **Vue 3 + TypeScript** frontend, modeled on the notepad-- editor experience. The reference UI/logic being cloned lives in `demo/notepad--/` (C++ source) — `frontend/src/types/index.ts` and `backend/api/handler.go` both cite specific notepad-- source files they mirror. Single EXE, no CGO (pure-Go SQLite via glebarez/sqlite).

## Commands

```bash
# Full app — dev mode (Vite HMR + Go backend, regenerates wailsjs bindings)
wails dev

# Production build (outputs build/bin/easy-text.exe)
wails build
wails build -platform windows/amd64

# NSIS installer (~11MB) — requires NSIS installed (winget install NSIS.NSIS)
wails build -platform windows/amd64 --nsis

# Backend only (no Wails needed)
cd backend && go build ./... && go vet ./...
cd backend && go test ./... -count=1 -timeout 60s

# Frontend only (run from frontend/)
cd frontend && npm install
npm run dev          # Vite dev server
npm run build        # vue-tsc --noEmit && vite build  (type-check IS part of build)
npm run typecheck    # vue-tsc only (faster than full build)
npm run lint         # ESLint with no-explicit-any + no-empty rules
npm run lint:fix     # auto-fix

# Top-level orchestration (cross-stack local validation)
make ci-build        # mirrors .github/workflows/ci.yml without wails
```

### Quality gates

- **Backend**: `go build`, `go vet`, `go test` 全部 0 失败 + `golangci-lint` 配置见 `.golangci.yml`（errcheck / vet / staticcheck / gosec / revive）
- **Frontend**: `npm run typecheck` + `npm run lint --max-warnings=0`
- **CI**: `.github/workflows/ci.yml` 在 push/PR 时自动跑 backend + frontend 两套 pipeline

There are **6 Go tests** in the repo (`utils/errors_test.go`、`tools/script_test.go`、`tools/script_concurrency_test.go`、`api/handler_test.go`)，覆盖 AppError 错误链、Lua 超时、并发回归、Handler 启动 fail-fast 不变量。修复 bug 时**必须先写复现测试**。

## Architecture

### Wails binding pattern (the central design)

The binding surface is decoupled from logic via struct embedding:

- `main.go` embeds `frontend/dist` and binds `main.App`.
- `app.go`'s `App` struct **embeds `*api.Handler`**, so every exported method on `api.Handler` is promoted onto `App` and automatically becomes a Wails-bound callable from JS. **Never add logic to `app.go` or `main.go`** — add methods to `api.Handler` in `backend/api/` and they surface to the frontend automatically.
- Wails lifecycle hooks must be lowercase methods on `App`: `startup`/`shutdown` delegate to `Handler.Startup`/`Handler.Shutdown`.

`Handler` (`backend/api/handler.go`) holds pointers to all domain services (`fileReader`, `fileWriter`, `treeBuilder`, `jsonTool`, `diffTool`, etc.), wired in `NewHandler`. The `api_*.go` files are **thin**: each exported method validates inputs and delegates to a service. Do not put domain logic in the `api/` layer — extend the corresponding `backend/file/` or `backend/tools/` service instead.

### Adding a backend API → frontend callable

1. Add an exported (PascalCase) method to `api.Handler` in the relevant `backend/api/api_*.go`.
2. Run `wails dev` (or `wails build`) — Wails **regenerates** `frontend/wailsjs/go/main/App.d.ts` and `App.js`. These files have a Welsh "DO NOT EDIT" header; **never hand-edit them**.
3. Frontend imports call the generated binding, e.g. `import { ReadFile } from '../../wailsjs/go/main/App'`.

### `[]byte` serialization gotcha

Wails v2 serializes Go `[]byte` to JS as `[]number` (an array of ints), not a typed buffer. Consequences visible in `api_files.go`:
- `ReadFileBytes` returns `[]byte` → arrives in JS as `number[]`.
- `SaveFileBytes` accepts `[]int` (not `[]byte`) and converts element-by-element. Follow this pattern for any binary in/out API.

### Backend layering

```
backend/api/       Wails-bound handlers (thin delegation). One file per feature area (api_files/dirs/json/diff/encoding/convert/hash/macro/find/rename/session/config/dialogs).
backend/file/      File domain: reader, writer, tree, watcher, converter (Office legacy→OOXML).
backend/tools/     Pure utilities: json, diff, encoding, convert, findreplace, macro.
backend/config/    AppConfig (JSON, schema-versioned) + SQLite via GORM.
backend/utils/     AppError (numeric codes) + global logger.
```

### Error convention

`utils.AppError` carries a numeric code + Chinese `Message`. Code ranges: **1000s** file, **2000s** JSON, **3000s** encoding, **4000s** diff, **5000s** config/db. Return `utils.NewAppError(code, message, details)` for user-facing failures. User-facing strings throughout the codebase are **Simplified Chinese** — match this when adding messages.

### Config & data persistence

All under `%LOCALAPPDATA%/EasyText/` (via `os.UserConfigDir()` with Windows fallback): `config.json`, `session.json`, `easytext.db`. `AppConfig` has a `Version` field — `version >= 1` is trusted as-is; legacy/missing configs merge field-by-field with defaults (see `mergeConfig` in `backend/config/config.go`). Bump the version when the schema changes meaningfully.

### Office/PDF viewing pipeline

`backend/file/converter.go` handles legacy Office (`.doc/.ppt/.xls`) → OOXML via a **multi-level fallback**: LibreOffice CLI → PowerShell COM → OLE2 text extraction. The frontend viewers (`frontend/src/components/viewer/`) render PDF (pdfjs-dist), Word (docx-preview), Excel (xlsx, editable), PPT (jszip + XML parsing), images, and hex — all lazy-loaded via Vite `manualChunks`.

### Frontend architecture

- **State**: Pinia stores in `src/stores/` (`editorStore`, `fileStore`, `settingStore`). `editorStore` owns tabs, bookmarks, position history, and macro state — closely mirroring notepad-- `ccnotepad.h`.
- **Path alias**: `@/` → `src/` (configured in both `vite.config.ts` and `tsconfig.json`).
- **Layout**: `MainLayout.vue` is the orchestrator; child components emit `menu-command` / `toolbar-command` events rather than each binding its own handler. `handleMenuCommand` dispatches by command prefix (`line-`, `sort-`, `case-`, `encode-`, etc.).
- **Stack**: Element Plus 2.14 (zh-CN locale) + Tailwind CSS + Lucide icons + CodeMirror 6 (60+ languages).
- `vue-tsc --noEmit` runs as part of `npm run build` — type errors fail the build.

## Conventions

- Go code follows standard `gofmt`/`go vet`; comments are bilingual (Chinese intent, exported APIs documented). The Go module path is `easy-text`.
- Match the existing thin-handler style: API methods delegate, services implement.
- **Error handling**: 优先用 `utils.WrapError(code, msg, cause)` 包装错误保留链；`utils.NewAppError` 仅用于静态信息。`AppError` 实现 `Unwrap()`，`errors.Is/As` 可穿透。
- **Concurrency**: `defer L.Close()` 独占于执行 goroutine 内；goroutine 内 panic 必须用 `defer recover()` 兜底；wg.Add 必须先于 `go` 语句。
- **Startup invariant**: DB 初始化失败 → fail-fast panic；DB 依赖服务（draft/recent/snippet/bookmark）在 `Startup` 后保证非 nil，**API 层禁止 nil 守卫**（样板代码已在 I-1 清理时移除）。
- `frontend/wailsjs/` is generated — always regenerate via `wails dev`/`build` after backend API changes; don't commit manual edits to it.
- 前端 `eslint` 规则：`@typescript-eslint/no-explicit-any` 强制 `unknown` 替代；空 catch 强制 `console.warn`。
- Project docs: `EasyText—产品需求文档.md` (PRD), `EasyText—设计文档.MD` (design/API/data structures), `overview.md` (recent UI-rewrite changelog). Consult these for feature intent before large changes.
