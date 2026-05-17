//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"easyPreparation_1.0/internal/youtube"
)

func main() {
	youtube.Init(
		filepath.Join("config", "google_oauth.json"),
		filepath.Join("data", "youtube_token.json"),
	)

	pairs := [][2]string{
		{"E4wMNUUCqK8", "data/templates/thumbnail/generated/2026-05-17_main_worship.png"},
		{"A7S2d-hDVlY", "data/templates/thumbnail/generated/2026-05-17_after_worship.png"},
	}

	for _, p := range pairs {
		videoID, thumbPath := p[0], p[1]
		if err := youtube.UploadThumbnailToBroadcast(videoID, thumbPath); err != nil {
			log.Printf("[thumbnail] %s 실패: %v", videoID, err)
			os.Exit(1)
		}
		fmt.Printf("썸네일 설정 완료: https://youtu.be/%s\n", videoID)
	}
}
