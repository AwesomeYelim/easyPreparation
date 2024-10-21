#!/bin/bash

# 빌드할 Go 파일 경로들 설정
GO_FILES=(
  "./executor/ppt/lyrics/lyrics.go"
  "./executor/ppt/history/history.go"
)

# 바이너리 파일을 저장할 디렉터리 설정
BIN_DIR="./bin"

# bin 디렉터리가 존재하지 않으면 생성
mkdir -p "${BIN_DIR}"

# 각 파일 빌드 수행
for GO_FILE in "${GO_FILES[@]}"; do
  # 파일명만 추출하여 바이너리 이름 설정
  BIN_NAME=$(basename "${GO_FILE}" .go)

  echo "Building Go file: ${GO_FILE}"
  go build -a -trimpath -o "${BIN_DIR}/${BIN_NAME}" "${GO_FILE}"

  # 빌드 성공 여부 확인
  if [ $? -eq 0 ]; then
    echo "Build completed successfully for: ${GO_FILE} -> ${BIN_DIR}/${BIN_NAME}"
  else
    echo "Build failed for: ${GO_FILE}"
    exit 1
  fi
done

echo "All builds completed successfully!"
