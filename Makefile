export PATH := /usr/local/go/bin:$(PATH)

.PHONY: dev restart build clean

# 개발 모드: Go 서버 + Next.js dev server 동시 실행
dev:
	@-lsof -ti:3000 -ti:3001 -ti:3002 | xargs kill -9 2>/dev/null || true
	@-lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@sleep 0.5
	@echo "Starting Go server (:8080) + Next.js dev (:3000)..."
	@. ~/.nvm/nvm.sh && nvm use 22 --silent && \
	(cd ui && npm run dev &) && \
	go run ./cmd/server/main.go

# 재시작: 기존 프로세스 종료 + .next 캐시 삭제 + dev
restart:
	@echo "Stopping existing processes..."
	@-lsof -ti:3000 | xargs kill -9 2>/dev/null || true
	@-lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@sleep 1
	@echo "Clearing .next cache..."
	@rm -rf ui/.next
	@$(MAKE) dev

# 프로덕션 빌드
build:
	@echo "Building Next.js..."
	@. ~/.nvm/nvm.sh && nvm use 22 --silent && cd ui && npm run build
	@echo "Building Go binary..."
	@go build -o bin/server ./cmd/server/main.go
	@echo "Done. Run: bin/server"

# Go 빌드만
build-go:
	go build -o bin/server ./cmd/server/main.go

# Next.js 빌드만
build-ui:
	. ~/.nvm/nvm.sh && nvm use 22 --silent && cd ui && npm run build

clean:
	rm -f bin/server
	rm -rf ui/.next
