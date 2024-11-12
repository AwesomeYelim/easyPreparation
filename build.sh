#!/bin/bash

# 빌드할 Go 파일 경로 설정
GO_FILES=(
  "전체 선택"   # 전체 선택 항목 추가
  "./executor/ppt/lyrics/lyrics.go"
  "./executor/ppt/history/history.go"
  "./executor/bulletin/bulletin.go"
)

# 바이너리 파일을 저장할 디렉토리 설정
BIN_DIR="./bin"

# bin 디렉토리가 존재하지 않으면 생성
mkdir -p "${BIN_DIR}"

# 메뉴에서 선택한 인덱스를 저장할 변수
selected=0

# 색상 코드 설정
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
RESET='\033[0m'  # 색상 리셋

# 메뉴 출력 함수
print_menu() {
  clear  # 화면을 깨끗이 지움
  echo "빌드할 Go 파일을 선택하세요 (화살표로 이동 후 엔터로 선택):"
  for i in "${!GO_FILES[@]}"; do
    if [ $i -eq $selected ]; then
      # 선택된 항목에 색상 적용
      echo -e "${YELLOW}> ${GO_FILES[i]}${RESET}"
    else
      echo "  ${GO_FILES[i]}"
    fi
  done
}

# 화살표 키 입력 처리 함수
navigate_menu() {
  while true; do
    print_menu  # 메뉴를 화면에 출력

    # 사용자의 입력을 1글자씩 읽음 (특수 키 감지 위해)
    read -rsn1 input

    case "$input" in
      $'\x1b')  # ESC (화살표 키 시작)
        read -rsn2 -t 0.1 input  # 화살표 키 방향 감지
        case "$input" in
          "[A")  # 위쪽 화살표
            ((selected--))
            if [ $selected -lt 0 ]; then
              selected=$((${#GO_FILES[@]} - 1))
            fi
            ;;
          "[B")  # 아래쪽 화살표
            ((selected++))
            if [ $selected -ge ${#GO_FILES[@]} ]; then
              selected=0
            fi
            ;;
        esac
        ;;
      "")  # 엔터 키 입력 시
        break
        ;;
    esac
  done
}

# Go 파일 빌드 함수
build_file() {
  local file="$1"
  local bin_name=$(basename "${file}" .go)

  echo "Go 파일 빌드 중: ${file}"
  go build -a -trimpath -o "${BIN_DIR}/${bin_name}" "${file}"

  # 빌드 성공 여부 확인
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}빌드 성공: ${file} -> ${BIN_DIR}/${bin_name}${RESET}"
  else
    echo -e "${RED}빌드 실패:${RESET} ${file}"
    exit 1
  fi
}

# 전체 빌드 함수
build_all_files() {
  for file in "${GO_FILES[@]:1}"; do  # 첫 항목 제외하고 빌드
    build_file "${file}"
  done
}

# 메뉴 탐색 시작
navigate_menu

# 선택된 파일 또는 전체 빌드 수행
if [ $selected -eq 0 ]; then
  build_all_files
else
  build_file "${GO_FILES[selected]}"
fi

echo -e "${BLUE}모든 빌드가 성공적으로 완료되었습니다!${RESET}"
