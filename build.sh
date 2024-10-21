#!/bin/bash

# 빌드할 Go 파일 경로들 설정
GO_FILES=(
  "./executor/ppt/lyrics/lyrics.go"
  "./executor/ppt/history/history.go"
)

# 각 파일 빌드 수행
for GO_FILE in "${GO_FILES[@]}"; do
  echo "Building Go file: ${GO_FILE}"
  go build -a -trimpath "${GO_FILE}"

  # 빌드 성공 여부 확인
  if [ $? -eq 0 ]; then
    echo "Build completed successfully for: ${GO_FILE}"
  else
    echo "Build failed for: ${GO_FILE}"
    exit 1
  fi
done

echo "All builds completed successfully!"