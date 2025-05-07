#!/bin/bash

set -e

GO_FILES=(
  "./executor/lyrics/presentation/lyricsPPT.go"
  "./executor/bulletin/bulletin.go"
)

BIN_DIR="./bin"
mkdir -p "${BIN_DIR}"

TARGETS=(
  "linux/amd64"
  "darwin/arm64"
  "windows/amd64"
)

# 각 타겟에 대해 빌드
for target in "${TARGETS[@]}"; do
  GOOS=$(echo $target | cut -d'/' -f1)
  GOARCH=$(echo $target | cut -d'/' -f2)

  for file in "${GO_FILES[@]}"; do
    bin_name=$(basename "${file}" .go)
    output_file="${BIN_DIR}/${bin_name}_${GOOS}_${GOARCH}"

    echo "Building $file for $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -a -trimpath -o "$output_file" "$file"
  done
done

echo "✅ All builds completed!"
