#!/bin/bash
# easyPreparation macOS 설치 스크립트
# 사용법: curl -fsSL https://github.com/AwesomeYelim/easyPreparation/releases/latest/download/install-mac.sh | bash

set -euo pipefail

APP_NAME="easyPreparation"
ZIP_NAME="${APP_NAME}_desktop_darwin_arm64.zip"
DOWNLOAD_URL="https://github.com/AwesomeYelim/easyPreparation/releases/latest/download/${ZIP_NAME}"
INSTALL_DIR="/Applications"
TMP_DIR=$(mktemp -d)

cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

echo "⬇  다운로드 중..."
curl -fSL "$DOWNLOAD_URL" -o "$TMP_DIR/$ZIP_NAME"

echo "📦  압축 해제 중..."
unzip -q "$TMP_DIR/$ZIP_NAME" -d "$TMP_DIR"

echo "🔓  Gatekeeper 해제 중..."
xattr -cr "$TMP_DIR/${APP_NAME}.app"

# 기존 앱 있으면 제거
if [ -d "$INSTALL_DIR/${APP_NAME}.app" ]; then
  echo "🗑   기존 버전 제거 중..."
  rm -rf "$INSTALL_DIR/${APP_NAME}.app"
fi

echo "📂  Applications로 이동 중..."
mv "$TMP_DIR/${APP_NAME}.app" "$INSTALL_DIR/"

echo "✅  설치 완료! Launchpad 또는 Finder에서 easyPreparation을 실행하세요."
open "$INSTALL_DIR/${APP_NAME}.app"
