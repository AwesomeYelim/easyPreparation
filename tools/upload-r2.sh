#!/bin/bash
# R2 버킷에 PDF 에셋 업로드 스크립트
# 사용법: bash tools/upload-r2.sh
#
# 사전 조건:
#   1. wrangler CLI 설치: npm install -g wrangler
#   2. wrangler login 완료
#   3. R2 버킷 생성: wrangler r2 bucket create easyprep-assets
#
# 업로드 대상: data/pdf/hymn/*.pdf, data/pdf/responsive_reading/*.pdf

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
PDF_DIR="$ROOT_DIR/data/pdf"

if [ ! -d "$PDF_DIR" ]; then
  echo "ERROR: $PDF_DIR 디렉토리를 찾을 수 없습니다."
  exit 1
fi

BUCKET="easyprep-assets"
UPLOADED=0
SKIPPED=0

for category in hymn responsive_reading; do
  dir="$PDF_DIR/$category"
  if [ ! -d "$dir" ]; then
    echo "SKIP: $dir 없음"
    continue
  fi

  for pdf in "$dir"/*.pdf; do
    [ -f "$pdf" ] || continue
    filename="$(basename "$pdf")"
    key="$category/$filename"

    echo "UPLOAD: $key"
    wrangler r2 object put "$BUCKET/$key" --file "$pdf" --content-type "application/pdf"
    UPLOADED=$((UPLOADED + 1))
  done
done

echo ""
echo "완료: ${UPLOADED}개 업로드"
