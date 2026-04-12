export PATH := /usr/local/go/bin:$(PATH)

VERSION    ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS     = -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)

# ── OS 감지 ──────────────────────────────────────────────────────────────────
ifeq ($(OS),Windows_NT)
  IS_WINDOWS := 1
else
  _UNAME := $(shell uname -s 2>/dev/null)
  ifneq ($(filter MINGW% CYGWIN% MSYS%,$(_UNAME)),)
    IS_WINDOWS := 1
  endif
endif

# ── npm 명령어 ────────────────────────────────────────────────────────────────
# Windows(Git Bash): npm 직접 호출 (nvm-windows는 별도 관리)
# Unix: ~/.nvm/nvm.sh 있으면 nvm use 22, 없으면 npm 직접 호출
ifdef IS_WINDOWS
  RUN_NPM := npm
else
  RUN_NPM := $(shell [ -f "$(HOME)/.nvm/nvm.sh" ] && \
    printf '. %s/.nvm/nvm.sh && nvm use 22 --silent && npm' "$(HOME)" || echo npm)
endif

.PHONY: dev restart build clean build-desktop build-desktop-macos \
        build-desktop-windows build-desktop-linux dev-desktop \
        build-go build-go-dev build-ui build-frontend build-landing upload-r2

# ── 포트 킬 헬퍼 (크로스 플랫폼) ─────────────────────────────────────────────
# 사용: $(call kill_ports,3000 8080)
# Unix  → lsof -ti:<port> | xargs kill -9
# Windows → netstat -ano + taskkill.exe
define kill_ports
@-{ \
  PORTS="$(1)"; \
  if command -v lsof >/dev/null 2>&1; then \
    for p in $$PORTS; do lsof -ti:$$p | xargs kill -9 2>/dev/null || true; done; \
  else \
    for p in $$PORTS; do \
      for _pid in $$(netstat -ano 2>/dev/null | \
          awk '/:'"$$p"' /&&/LISTEN/{print $$NF}' | sort -u); do \
        taskkill.exe //PID $$_pid //F 2>/dev/null || true; \
      done; \
    done; \
  fi; \
} || true
endef

# ── 개발 모드: Go 서버 + Next.js dev server 동시 실행 ───────────────────────
dev:
	$(call kill_ports,3000 3001 3002 8080)
	@echo "Starting Go server (:8080) + Next.js dev (:3000)..."
	@(cd ui && $(RUN_NPM) run dev &) && \
	go run -tags dev ./cmd/server/

# ── 재시작: 기존 프로세스 종료 + .next 캐시 삭제 + dev ──────────────────────
restart:
	@echo "Stopping existing processes..."
	$(call kill_ports,3000 8080)
	@echo "Clearing .next cache..."
	@rm -rf ui/.next
	@$(MAKE) dev

# ── 프로덕션 빌드 (Next.js static export → Go embed) ─────────────────────────
build:
	@echo "Building Next.js (static export)..."
	@cd ui && $(RUN_NPM) run build
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

# ── Go 빌드만 (프로덕션 — cmd/server/frontend/ + data/ 필요) ──────────────────
build-go:
	@rm -rf cmd/server/frontend && cp -r ui/out cmd/server/frontend
	@rm -rf cmd/server/data && mkdir -p cmd/server/data/defaults
	@cp data/bible.db cmd/server/data/bible.db
	@for f in bible_info main_worship after_worship wed_worship fri_worship; do \
		[ -f config/$${f}.json ] && cp config/$${f}.json cmd/server/data/defaults/$${f}.json || true; \
	done
	go build -ldflags="$(LDFLAGS)" -o bin/server ./cmd/server/

# ── Go 빌드 (개발 — embed 없이) ───────────────────────────────────────────────
build-go-dev:
	go build -tags dev -ldflags="$(LDFLAGS)" -o bin/server ./cmd/server/

# ── Next.js 빌드만 ────────────────────────────────────────────────────────────
build-ui:
	cd ui && $(RUN_NPM) run build

clean:
	rm -f bin/server
	rm -rf ui/.next
	rm -rf cmd/server/data cmd/desktop/data

# ── 프론트엔드 + 데이터 준비 (Desktop 빌드 공통 선행 작업) ──────────────────────
# bible.db를 cmd/desktop/data/에 복사 → //go:embed all:data 에 포함됨
build-frontend:
	@echo "Building Next.js (static export)..."
	@cd ui && $(RUN_NPM) run build
	@echo "Copying frontend to cmd/desktop/frontend/..."
	@rm -rf cmd/desktop/frontend && cp -r ui/out cmd/desktop/frontend
	@echo "Copying embedded data (bible.db + config defaults)..."
	@rm -rf cmd/desktop/data && mkdir -p cmd/desktop/data/defaults
	@cp data/bible.db cmd/desktop/data/bible.db
	@for f in bible_info main_worship after_worship wed_worship fri_worship; do \
		[ -f config/$${f}.json ] && cp config/$${f}.json cmd/desktop/data/defaults/$${f}.json || true; \
	done

# ── Desktop 앱 빌드 (macOS) — wails.json이 cmd/desktop/에 있으므로 cd 필요 ──
build-desktop: build-frontend
	@echo "Building Wails app (macOS)..."
	@cd cmd/desktop && export PATH="$$HOME/go/bin:/usr/local/go/bin:$$PATH" && \
	CGO_LDFLAGS="-framework UniformTypeIdentifiers" \
	wails build -s -ldflags="$(LDFLAGS)" -o easyPreparation
	@echo "Done: cmd/desktop/build/bin/easyPreparation.app"

# ── Desktop 앱 빌드 (macOS 명시적 타겟) ──────────────────────────────────────
build-desktop-macos: build-frontend
	@echo "Building Wails app (macOS)..."
	@cd cmd/desktop && export PATH="$$HOME/go/bin:/usr/local/go/bin:$$PATH" && \
	CGO_LDFLAGS="-framework UniformTypeIdentifiers" \
	wails build -s -o easyPreparation -ldflags="$(LDFLAGS)"
	@echo "Done: cmd/desktop/build/bin/easyPreparation.app"

# ── Desktop 앱 빌드 (Windows) ─────────────────────────────────────────────────
# -s: frontend는 build-frontend에서 이미 준비됨 (wails.json frontend:build 스킵)
build-desktop-windows: build-frontend
	@echo "Building Wails app (Windows amd64)..."
	@cd cmd/desktop && export PATH="$$HOME/go/bin:/usr/local/go/bin:$$PATH" && \
	wails build -s -o easyPreparation.exe -ldflags="$(LDFLAGS)"
	@echo "Done: cmd/desktop/build/bin/easyPreparation.exe"

# ── Desktop 앱 빌드 (Linux) ───────────────────────────────────────────────────
build-desktop-linux: build-frontend
	@echo "Building Wails app (Linux amd64)..."
	@cd cmd/desktop && export PATH="$$HOME/go/bin:/usr/local/go/bin:$$PATH" && \
	wails build -s -o easyPreparation -ldflags="$(LDFLAGS)" -platform linux/amd64
	@echo "Done: cmd/desktop/build/bin/easyPreparation"

# ── Desktop 개발 모드 ─────────────────────────────────────────────────────────
# -s: wails.json frontend:build의 Unix 명령어 스킵 (Next.js는 watcher가 자동 시작)
# -frontenddevserverurl: Go 앱의 로딩화면이 리다이렉트할 URL 명시
dev-desktop:
	$(call kill_ports,3000 3001 8080)
	@cd cmd/desktop && export PATH="$$HOME/go/bin:/usr/local/go/bin:$$PATH" && \
	wails dev -s -frontenddevserverurl http://localhost:3000

# ── 랜딩 페이지 빌드 ──────────────────────────────────────────────────────────
build-landing:
	@echo "Building landing page..."
	@cd landing && $(RUN_NPM) ci && $(RUN_NPM) run build

# ── R2에 PDF 에셋 업로드 ──────────────────────────────────────────────────────
upload-r2:
	bash tools/upload-r2.sh
