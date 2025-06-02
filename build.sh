#!/bin/bash

# 빌드할 경로
TARGET="./apiServer/."

# 바이너리 출력 경로
BIN_DIR="./bin"

# OS 및 아키텍처
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)

# 출력 파일명 생성
OUTPUT_NAME="apiServer_${GOOS}_${GOARCH}"
[ "$GOOS" == "windows" ] && OUTPUT_NAME="${OUTPUT_NAME}.exe"

# 출력 디렉토리 생성
mkdir -p "$BIN_DIR"

# 빌드 수행
echo "Building for $GOOS/$GOARCH..."
go build -o "${BIN_DIR}/${OUTPUT_NAME}" "$TARGET"

# 결과 확인
if [ $? -eq 0 ]; then
  echo "✅ 빌드 성공: ${BIN_DIR}/${OUTPUT_NAME}"
else
  echo "❌ 빌드 실패"
  exit 1
fi
