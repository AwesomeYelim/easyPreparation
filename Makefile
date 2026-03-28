.PHONY: dev build clean

# 개발 모드: Go 서버 + Next.js dev server 동시 실행
dev:
	@echo "Starting Go server (:8080) + Next.js dev (:3000)..."
	@. ~/.nvm/nvm.sh && nvm use 22 --silent && \
	(cd ui && npm run dev &) && \
	go run ./cmd/server/main.go

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
