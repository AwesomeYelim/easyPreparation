// ytupload — 로컬 영상 파일을 기존 인증된 YouTube 계정으로 업로드한다.
//
// 사용:
//
//	go run ./tools/ytupload/ <videoPath> <title> [description] [privacy]
//
// privacy: public / unlisted / private (기본: unlisted)
package main

import (
	"fmt"
	"os"

	"easyPreparation_1.0/internal/youtube"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "사용법: go run ./tools/ytupload/ <videoPath> <title> [description] [privacy]")
		os.Exit(1)
	}

	videoPath := os.Args[1]
	title := os.Args[2]
	description := ""
	if len(os.Args) > 3 {
		description = os.Args[3]
	}
	privacy := "unlisted"
	if len(os.Args) > 4 {
		privacy = os.Args[4]
	}

	if _, err := os.Stat(videoPath); err != nil {
		fmt.Fprintf(os.Stderr, "비디오 파일을 찾을 수 없음: %v\n", err)
		os.Exit(1)
	}

	youtube.Init(youtube.DefaultOAuthPath(), youtube.DefaultTokenPath())
	if !youtube.Get().IsEnabled() {
		fmt.Fprintln(os.Stderr, "YouTube 연동이 비활성 상태입니다 (OAuth 설정/토큰 확인 필요)")
		os.Exit(1)
	}

	fmt.Printf("업로드 시작: %s (제목: %s, 공개설정: %s)\n", videoPath, title, privacy)
	videoID, err := youtube.UploadVideo(videoPath, title, description, privacy, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "업로드 실패: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("업로드 완료: https://youtu.be/%s\n", videoID)
}
