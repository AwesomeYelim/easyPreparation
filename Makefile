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

# ── Node.js bin 경로 감지 (wails dev PATH 주입용) ────────────────────────────
# 우선순위: 1) 이미 PATH에 npm 있음  2) nvm(macOS/Linux)  3) nvm-windows(APPDATA)
ifdef IS_WINDOWS
  _NPM_CMD     := $(shell where npm 2>NUL | head -1)
  NODE_BIN_DIR := $(strip $(if $(_NPM_CMD),$(dir $(_NPM_CMD)),$(shell ls -d "$$APPDATA/nvm/v22"*/ 2>/dev/null | head -1)))
else
  _NPM_CMD     := $(shell command -v npm 2>/dev/null)
  NODE_BIN_DIR := $(strip $(if $(_NPM_CMD),$(dir $(_NPM_CMD)),$(shell ls -d $(HOME)/.nvm/versions/node/v22*/bin 2>/dev/null | head -1)))
endif

.PHONY: dev restart build clean build-desktop build-desktop-macos \
        build-desktop-windows build-desktop-linux dev-desktop \
        build-go build-go-dev build-ui build-frontend build-landing upload-r2 \
        sync-ai install-hooks dev-license health

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
	@(cd ui && NEXT_PUBLIC_DEV_MODE=true $(RUN_NPM) run dev &) && \
	EASYPREP_DEV=true go run -tags dev ./cmd/server/

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
	rm -rf cmd/server/frontend cmd/server/data
	rm -rf cmd/desktop/frontend cmd/desktop/data

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
# Next.js(:3000) 백그라운드 기동 → 준비 완료 후 wails dev 시작
# Ctrl+C 시 trap으로 하위 프로세스 모두 종료
dev-desktop:
	$(call kill_ports,3000 8080)
	@echo "Starting Next.js dev (:3000) + Wails Desktop..."
	@( \
	  trap 'kill $$(jobs -p) 2>/dev/null; exit 0' EXIT INT TERM; \
	  (cd ui && NEXT_PUBLIC_DEV_MODE=true $(RUN_NPM) exec -- next dev -p 3000) & \
	  echo "Next.js 시작 대기 중..."; \
	  until curl -sf http://localhost:3000 >/dev/null 2>&1; do sleep 0.5; done; \
	  echo "Next.js 준비 완료 — Wails Desktop 시작"; \
	  cd cmd/desktop && \
	  export PATH="$$HOME/go/bin:/usr/local/go/bin:$(NODE_BIN_DIR):$$PATH" && \
	  export EASYPREP_DEV=true && \
	  wails dev -s -frontenddevserverurl http://localhost:3000; \
	)

# ── 랜딩 페이지 빌드 ──────────────────────────────────────────────────────────
build-landing:
	@echo "Building landing page..."
	@cd landing && $(RUN_NPM) ci && $(RUN_NPM) run build

# ── R2에 PDF 에셋 업로드 ──────────────────────────────────────────────────────
upload-r2:
	bash tools/upload-r2.sh

# ── ai_supporter sync ────────────────────────────────────────────────────────
# auto-runs after git pull (when install-hooks is set)
# manual: make sync-ai
sync-ai:
	@bash tools/sync-ai-supporter.sh

# ── 개발용 Pro 라이선스 생성 (data/license.json 덮어씀) ──────────────────────
# 사용: make dev-license           (Pro, 무기한)
#        make dev-license PLAN=free (Free 초기화)
#        make dev-license PLAN=enterprise
PLAN ?= pro
dev-license:
	@go run ./tools/devlicense/ $(PLAN)

# Windows 배포용 devlicense.exe 빌드 (프로덕션 머신에서 실행)
build-devlicense-windows:
	@GOOS=windows GOARCH=amd64 go build -o tools/output/devlicense.exe ./tools/devlicense/
	@echo "✅ tools/output/devlicense.exe 생성 완료"
	@echo "   Windows에서 실행: devlicense.exe      (Pro)"
	@echo "                    devlicense.exe free  (Free)"

# ── 코드 헬스 체크 (빌드 검증 + 타입 체크) ──────────────────────────────────
health:
	@echo "▶ Go vet..."
	@go vet ./cmd/server/ ./cmd/desktop/ ./internal/... ./tools/devlicense/
	@echo "▶ TypeScript typecheck..."
	@cd ui && $(RUN_NPM) run typecheck 2>/dev/null || $(RUN_NPM) exec tsc -- --noEmit
	@echo "✅ 헬스 체크 완료"

# ── Git hooks install (run once) ─────────────────────────────────────────────
# .githooks/post-merge -> auto-syncs ai_supporter on git pull
install-hooks:
	@git config core.hooksPath .githooks
	@chmod +x .githooks/post-merge tools/sync-ai-supporter.sh
	@echo "Git hooks installed."
	@echo "  'git pull' will auto-sync ai_supporter."
	@echo ""
	@echo "Set ai_supporter path (pick one):"
	@echo "  1) echo '/path/to/ai_supporter' > .ai-supporter-path"
	@echo "  2) export AI_SUPPORTER_PATH=/path/to/ai_supporter"
	@echo "  -> '../ai_supporter' sibling dir is auto-detected."
