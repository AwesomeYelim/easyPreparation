// ytsetprivacy — 기존 인증된 YouTube 계정의 영상 공개설정을 변경한다.
//
// 사용:
//
//	go run ./tools/ytsetprivacy/ <videoID> <public|unlisted|private>
package main

import (
	"fmt"
	"os"

	"easyPreparation_1.0/internal/youtube"
	yt "google.golang.org/api/youtube/v3"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "사용법: go run ./tools/ytsetprivacy/ <videoID> <public|unlisted|private>")
		os.Exit(1)
	}
	videoID := os.Args[1]
	privacy := os.Args[2]

	youtube.Init(youtube.DefaultOAuthPath(), youtube.DefaultTokenPath())
	m := youtube.Get()
	if !m.IsEnabled() {
		fmt.Fprintln(os.Stderr, "YouTube 연동이 비활성 상태입니다")
		os.Exit(1)
	}
	svc := m.GetService()

	video := &yt.Video{
		Id: videoID,
		Status: &yt.VideoStatus{
			PrivacyStatus: privacy,
		},
	}

	resp, err := svc.Videos.Update([]string{"status"}, video).Do()
	if err != nil {
		fmt.Fprintf(os.Stderr, "공개설정 변경 실패: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("변경 완료: %s → %s\n", resp.Id, resp.Status.PrivacyStatus)
}
