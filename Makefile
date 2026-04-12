export PATH := /usr/local/go/bin:$(PATH)

VERSION   ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
COMMIT    ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    = -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)

.PHONY: dev restart build clean build-desktop build-desktop-macos build-desktop-windows build-desktop-linux dev-desktop build-go build-go-dev build-ui build-frontend build-landing upload-r2

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
	@echo "Copying embedded data..."
	@rm -rf cmd/server/data && mkdir -p cmd/server/data/defaults
	@cp data/bible.db cmd/server/data/bible.db
	@for f in bible_info main_worship after_worship wed_worship fri_worship; do \
		[ -f config/$${f}.json ] && cp config/$${f}.json cmd/server/data/defaults/$${f}.json || true; \
	done
	@echo "Building Go binary (with embedded frontend + data)..."
	@go build -ldflags="$(LDFLAGS)" -o bin/server ./cmd/server/
	@echo "Done. Run: bin/server"

# Go 빌드만 (프로덕션 — cmd/server/frontend/ 필요)
build-go:
	@rm -rf cmd/server/frontend && cp -r ui/out cmd/server/frontend
	@rm -rf cmd/server/data && mkdir -p cmd/server/data/defaults
	@cp data/bible.db cmd/server/data/bible.db
	@for f in bible_info main_worship after_worship wed_worship fri_worship; do \
		[ -f config/$${f}.json ] && cp config/$${f}.json cmd/server/data/defaults/$${f}.json || true; \
	done
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
	rm -rf cmd/server/data cmd/desktop/data

# 프론트엔드 + 데이터 준비 (Desktop 빌드 공통 선행 작업)
build-frontend:
	@echo "Building Next.js (static export)..."
	@. ~/.nvm/nvm.sh && nvm use 22 --silent && cd ui && npm run build
	@echo "Copying frontend to cmd/desktop/frontend/..."
	@rm -rf cmd/desktop/frontend && cp -r ui/out cmd/desktop/frontend
	@echo "Copying embedded data..."
	@rm -rf cmd/desktop/data && mkdir -p cmd/desktop/data/defaults
	@cp data/bible.db cmd/desktop/data/bible.db
	@for f in bible_info main_worship after_worship wed_worship fri_worship; do \
		[ -f config/$${f}.json ] && cp config/$${f}.json cmd/desktop/data/defaults/$${f}.json || true; \
	done

# Desktop 앱 빌드 (macOS — 현재 OS 기본 타겟) — wails.json이 cmd/desktop/에 있으므로 cd 필요
build-desktop: build-frontend
	@echo "Building Wails app (macOS)..."
	@cd cmd/desktop && export PATH="$$HOME/go/bin:/usr/local/go/bin:$$PATH" && CGO_LDFLAGS="-framework UniformTypeIdentifiers" wails build -s -ldflags="$(LDFLAGS)" -o easyPreparation
	@echo "Done: cmd/desktop/build/bin/easyPreparation.app"

# Desktop 앱 빌드 (macOS 명시적 타겟)
build-desktop-macos: build-frontend
	@echo "Building Wails app (macOS)..."
	@cd cmd/desktop && export PATH="$$HOME/go/bin:/usr/local/go/bin:$$PATH" && \
	CGO_LDFLAGS="-framework UniformTypeIdentifiers" \
	wails build -o easyPreparation -ldflags="$(LDFLAGS)"
	@echo "Done: cmd/desktop/build/bin/easyPreparation.app"

# Desktop 앱 빌드 (Windows — NSIS 인스톨러 포함)
# 주의: Windows 크로스 컴파일은 docker 환경 또는 Windows 머신에서 실행 권장
# wails build -nsis 는 wails CLI v2.x 이상 필요
build-desktop-windows:
	@echo "Building Wails app (Windows amd64)..."
	@cd cmd/desktop && export PATH="$$HOME/go/bin:/usr/local/go/bin:$$PATH" && \
	GOOS=windows GOARCH=amd64 \
	wails build -o easyPreparation.exe -nsis -ldflags="$(LDFLAGS)" -platform windows/amd64
	@echo "Done: cmd/desktop/build/bin/easyPreparation.exe (+ installer)"

# Desktop 앱 빌드 (Linux — raw binary)
# .desktop 파일은 cmd/desktop/build/linux/easyPreparation.desktop 참조
build-desktop-linux:
	@echo "Building Wails app (Linux amd64)..."
	@cd cmd/desktop && export PATH="$$HOME/go/bin:/usr/local/go/bin:$$PATH" && \
	wails build -o easyPreparation -ldflags="$(LDFLAGS)" -platform linux/amd64
	@echo "Done: cmd/desktop/build/bin/easyPreparation"

# 랜딩 페이지 빌드
build-landing:
	@echo "Building landing page..."
	@. ~/.nvm/nvm.sh && nvm use 22 --silent && cd landing && npm ci && npm run build

# R2에 PDF 에셋 업로드
upload-r2:
	bash tools/upload-r2.sh

# Desktop 개발 모드
dev-desktop:
	@-lsof -ti:3000 -ti:3001 -ti:8080 | xargs kill -9 2>/dev/null || true
	@sleep 0.5
	@cd cmd/desktop && export PATH="$$HOME/go/bin:/usr/local/go/bin:$$PATH" && CGO_LDFLAGS="-framework UniformTypeIdentifiers" wails dev
