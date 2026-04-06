export PATH := /usr/local/go/bin:$(PATH)

VERSION   ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
COMMIT    ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    = -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)

.PHONY: dev restart build clean build-desktop dev-desktop build-go build-go-dev build-ui

# 개발 모드: Go 서버 + Next.js dev server 동시 실행
dev:
	@-lsof -ti:3000 -ti:3001 -ti:3002 | xargs kill -9 2>/dev/null || true
	@-lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@sleep 0.5
	@echo "Starting Go server (:8080) + Next.js dev (:3000)..."
	@. ~/.nvm/nvm.sh && nvm use 22 --silent && \
	(cd ui && npm run dev &) && \
	go run -tags dev ./cmd/server/

# 재시작: 기존 프로세스 종료 + .next 캐시 삭제 + dev
restart:
	@echo "Stopping existing processes..."
	@-lsof -ti:3000 | xargs kill -9 2>/dev/null || true
	@-lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@sleep 1
	@echo "Clearing .next cache..."
	@rm -rf ui/.next
	@$(MAKE) dev

# 프로덕션 빌드 (Next.js static export → Go embed)
build:
	@echo "Building Next.js (static export)..."
	@. ~/.nvm/nvm.sh && nvm use 22 --silent && cd ui && npm run build
	@echo "Copying frontend to cmd/server/frontend/..."
	@rm -rf cmd/server/frontend && cp -r ui/out cmd/server/frontend
	@echo "Building Go binary (with embedded frontend)..."
	@go build -ldflags="$(LDFLAGS)" -o bin/server ./cmd/server/
	@echo "Done. Run: bin/server"

# Go 빌드만 (프로덕션 — cmd/server/frontend/ 필요)
build-go:
	@rm -rf cmd/server/frontend && cp -r ui/out cmd/server/frontend
	go build -ldflags="$(LDFLAGS)" -o bin/server ./cmd/server/

# Go 빌드 (개발 — embed 없이)
build-go-dev:
	go build -tags dev -ldflags="$(LDFLAGS)" -o bin/server ./cmd/server/

# Next.js 빌드만
build-ui:
	. ~/.nvm/nvm.sh && nvm use 22 --silent && cd ui && npm run build

clean:
	rm -f bin/server
	rm -rf ui/.next

# Desktop 앱 빌드 (macOS) — wails.json이 cmd/desktop/에 있으므로 cd 필요
build-desktop:
	@echo "Building Next.js..."
	@. ~/.nvm/nvm.sh && nvm use 22 --silent && cd ui && npm run build
	@echo "Copying frontend..."
	@rm -rf cmd/desktop/frontend && cp -r ui/out cmd/desktop/frontend
	@echo "Building Wails app..."
	@cd cmd/desktop && export PATH="$$HOME/go/bin:/usr/local/go/bin:$$PATH" && CGO_LDFLAGS="-framework UniformTypeIdentifiers" wails build -s -ldflags="$(LDFLAGS)" -o easyPreparation
	@echo "Done: cmd/desktop/build/bin/easyPreparation.app"

# Desktop 개발 모드
dev-desktop:
	@-lsof -ti:3000 -ti:3001 -ti:8080 | xargs kill -9 2>/dev/null || true
	@sleep 0.5
	@cd cmd/desktop && export PATH="$$HOME/go/bin:/usr/local/go/bin:$$PATH" && CGO_LDFLAGS="-framework UniformTypeIdentifiers" wails dev
