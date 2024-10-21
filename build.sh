#!/bin/bash

# 빌드할 Go 파일 경로 설정
GO_FILES=(
  "./executor/ppt/lyrics/lyrics.go"
  "./executor/ppt/history/history.go"
)

# 바이너리 파일을 저장할 디렉토리 설정
BIN_DIR="./bin"

# bin 디렉토리가 존재하지 않으면 생성
mkdir -p "${BIN_DIR}"

# fzf가 설치되어 있는지 확인하는 함수
check_fzf_installed() {
  if ! command -v fzf &> /dev/null; then
    echo "fzf가 설치되어 있지 않습니다. 설치하려면 다음 명령어를 사용하세요:"
    echo "sudo apt install fzf  # Ubuntu/Debian 사용자"
    echo "brew install fzf       # macOS 사용자"
    exit 1
  fi
}

# fzf가 설치되어 있는지 확인
check_fzf_installed

# 빌드할 파일 선택 안내 메시지 출력
echo "빌드할 Go 파일을 선택하세요(화살표 키로 탐색하고 스페이스로 선택, 엔터로 확인):"

# fzf를 사용하여 사용자에게 파일 선택하도록 하기
SELECTED_FILES=$(printf "%s\n" "${GO_FILES[@]}" | fzf --multi --header "빌드할 Go 파일을 선택하세요:")

# 선택된 파일이 없을 경우 확인
if [[ -z "$SELECTED_FILES" ]]; then
  echo "선택된 파일이 없습니다."
  exit 1
fi

# Go 파일 빌드 함수
build_file() {
  local file="$1"
  local bin_name=$(basename "${file}" .go)

  echo "Go 파일 빌드 중: ${file}"
  go build -a -trimpath -o "${BIN_DIR}/${bin_name}" "${file}"

  # 빌드 성공 여부 확인
  if [ $? -eq 0 ]; then
    echo "빌드 성공: ${file} -> ${BIN_DIR}/${bin_name}"
  else
    echo "빌드 실패: ${file}"
    exit 1
  fi
}

# 선택된 파일 빌드 수행
while IFS= read -r file; do
  build_file "$file"
done <<< "$SELECTED_FILES"

echo "모든 빌드가 성공적으로 완료되었습니다!"
