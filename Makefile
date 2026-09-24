SHELL := /bin/bash

.PHONY: help test desktop-dev desktop-build desktop-test docs-dev docs-build

help:
	@echo "Targets:"
	@echo "  test          - go test module gốc (bookmaker, cover, m4b, scripts/tts, CLI)"
	@echo "  --- Phần mềm tạo sách (Wails, desktop/) ---"
	@echo "  desktop-dev   - wails dev (Vite :5390, mở trình duyệt :34115 để gọi được Go)"
	@echo "  desktop-build - wails build → desktop/build/bin/"
	@echo "  desktop-test  - go test + typecheck frontend của desktop/"
	@echo "  --- Trang tài liệu (VitePress, docs/) ---"
	@echo "  docs-dev      - vitepress dev (http://localhost:5173/sano-sach-noi/)"
	@echo "  docs-build    - vitepress build → docs/.vitepress/dist/ (báo lỗi nếu có link gãy)"

test:
	go test ./...

# --- Phần mềm tạo sách (desktop/, Wails v2) ---
# Cài CLI: go install github.com/wailsapp/wails/v2/cmd/wails@latest
WAILS ?= $(shell command -v wails 2>/dev/null || echo $(HOME)/go/bin/wails)
# Phiên bản gắn vào app: file VERSION ở gốc repo nếu có, không thì "dev"
APP_VERSION ?= $(shell cat VERSION 2>/dev/null || echo dev)

desktop-dev:
	cd desktop && $(WAILS) dev

desktop-build:
	cd desktop && $(WAILS) build -clean -ldflags "-X main.version=$(APP_VERSION)"

desktop-test:
	cd desktop && go vet ./... && go test ./...
	cd desktop/frontend && npm install && npx vue-tsc --noEmit

# --- Trang tài liệu (docs/, VitePress) ---
docs/node_modules: docs/package-lock.json
	cd docs && npm ci
	@touch docs/node_modules

docs-dev: docs/node_modules
	cd docs && npm run docs:dev

docs-build: docs/node_modules
	cd docs && npm run docs:build
