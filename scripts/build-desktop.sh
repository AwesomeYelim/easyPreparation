#!/usr/bin/env bash
# Desktop 앱 빌드 (Wails v3) — Makefile 과 release.yml 이 함께 쓴다.
#
#   scripts/build-desktop.sh [darwin|windows|linux]   (생략하면 현재 OS)
#
# 환경 변수
#   LDFLAGS  추가 -ldflags (버전 정보: -X main.Version=... 등)
#   NSIS=1   Windows 설치 파일(NSIS)까지 만든다 (makensis 필요)
#   WAILS3   wails3 CLI 경로 (기본: wails3)
#
# 프론트엔드(internal/embedded/frontend)와 embed 데이터는 미리 준비돼 있어야 한다
# (make build-frontend 또는 release.yml 의 frontend-dist 아티팩트).
#
# 결과물 (cmd/desktop/bin/)
#   darwin  easyPreparation.app
#   windows easyPreparation.exe (+ NSIS=1 이면 easyPreparation-amd64-installer.exe)
#   linux   easyPreparation
set -euo pipefail

cd "$(dirname "$0")/../cmd/desktop"

TARGET_OS="${1:-$(go env GOOS)}"
WAILS3="${WAILS3:-wails3}"
APP=easyPreparation
EXTRA_LDFLAGS="${LDFLAGS:-}"

mkdir -p bin

case "$TARGET_OS" in
  windows)
    # 아이콘·버전 정보·매니페스트(DPI 인식)를 exe 에 넣는다 — 빌드 후 지운다
    SYSO="wails_windows_amd64.syso"
    trap 'rm -f "$SYSO"' EXIT
    "$WAILS3" generate syso -arch amd64 \
      -icon build/windows/icon.ico \
      -manifest build/windows/wails.exe.manifest \
      -info build/windows/info.json \
      -out "$SYSO"
    GOOS=windows GOARCH=amd64 go build -tags production -trimpath -buildvcs=false \
      -ldflags "-w -s -H windowsgui $EXTRA_LDFLAGS" -o "bin/$APP.exe" .
    echo "Done: cmd/desktop/bin/$APP.exe"

    if [ "${NSIS:-}" = "1" ]; then
      "$WAILS3" generate webview2bootstrapper -dir build/windows/nsis
      # makensis 의 File 은 역슬래시 Windows 경로만 찾는다 (D:/a/... 는 "no files found")
      EXE_WIN="$(cygpath -w "$(pwd)/bin/$APP.exe")"
      (cd build/windows/nsis && makensis -DARG_WAILS_AMD64_BINARY="$EXE_WIN" project.nsi)
      echo "Done: cmd/desktop/bin/$APP-amd64-installer.exe"
    fi
    ;;

  darwin)
    go build -tags production -trimpath -buildvcs=false \
      -ldflags "-w -s $EXTRA_LDFLAGS" -o "bin/$APP" .
    BUNDLE="bin/$APP.app"
    rm -rf "$BUNDLE"
    mkdir -p "$BUNDLE/Contents/MacOS" "$BUNDLE/Contents/Resources"
    mv "bin/$APP" "$BUNDLE/Contents/MacOS/$APP"
    cp build/darwin/Info.plist "$BUNDLE/Contents/Info.plist"
    # appicon.png → icons.icns (Info.plist CFBundleIconFile = icons)
    ICO_TMP="$(mktemp -d)"
    "$WAILS3" generate icons -input build/appicon.png \
      -macfilename "$BUNDLE/Contents/Resources/icons.icns" \
      -windowsfilename "$ICO_TMP/icon.ico"
    rm -rf "$ICO_TMP"
    echo "Done: cmd/desktop/$BUNDLE"
    ;;

  linux)
    # gtk3: v3 기본(GTK4/webkitgtk-6.0) 대신 기존과 같은 GTK3/webkit2gtk-4.1 사용
    go build -tags production,gtk3 -trimpath -buildvcs=false \
      -ldflags "-w -s $EXTRA_LDFLAGS" -o "bin/$APP" .
    echo "Done: cmd/desktop/bin/$APP"
    ;;

  *)
    echo "지원하지 않는 OS: $TARGET_OS (darwin|windows|linux)" >&2
    exit 1
    ;;
esac
