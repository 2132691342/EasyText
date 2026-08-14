# EasyText 顶层 Makefile — 本地开发命令
#
# Wails 应用需要 Windows 桌面环境运行 npm run dev / wails dev。
# CI 在 Linux 上跑（见 .github/workflows/ci.yml）。

.PHONY: help backend frontend test lint ci-build clean

help:
	@echo "EasyText 顶层命令:"
	@echo "  make backend     - 后端 go build/vet/test"
	@echo "  make frontend    - 前端 type-check/lint/build"
	@echo "  make test        - 后端单元测试"
	@echo "  make lint        - 前端 ESLint"
	@echo "  make ci-build    - CI 同等检查（Linux 子集）"
	@echo "  make clean       - 清理构建产物"

backend:
	cd backend && go build ./... && go vet ./...

frontend:
	cd frontend && npm install && npm run typecheck && npm run build

test:
	cd backend && go test ./... -count=1 -timeout 60s

lint:
	cd frontend && npm run lint

ci-build: backend test
	cd frontend && npm ci && npm run typecheck && npm run build

clean:
	cd frontend && rm -rf dist node_modules/.cache
	cd backend && go clean -cache