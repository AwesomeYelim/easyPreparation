#!/bin/bash

# 빌드할 Go 파일 경로 설정
GO_FILE_PATH="./executor/ppt/lyrics/lyrics.go"

# 빌드 명령 실행
echo "Building Go file: ${GO_FILE_PATH}"
go build -a -trimpath "${GO_FILE_PATH}"

# 빌드 성공 여부 확인
if [ $? -eq 0 ]; then
echo "Build completed successfully!"
else
echo "Build failed."
exit 1
fi
