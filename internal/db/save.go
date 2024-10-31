package db

import (
	"easyPreparation_1.0/internal/lyrics"
	"errors"
	"fmt"
	"github.com/timshannon/bolthold"
)

// Bolthold 데이터베이스에 노래 저장 함수
func SaveSongToDB(store *bolthold.Store, revSong *lyrics.SlideData) error {
	var existingSong lyrics.SlideData
	// 제목으로 기존 노래 조회
	err := store.FindOne(&existingSong, bolthold.Where("Title").Eq(revSong.Title))
	if err == nil {
		// 기존 노래가 있는 경우, 내용을 업데이트
		existingSong.Content = append(existingSong.Content, revSong.Content...) // 기존 내용에 추가
		return store.Update(existingSong.TrackID, &existingSong)                // 업데이트
	}

	if errors.Is(err, bolthold.ErrNotFound) {
		return store.Insert(revSong.TrackID, revSong)
	}

	if err != nil {
		return fmt.Errorf("ID 생성 실패: %v", err)
	}

	return store.Insert(revSong.TrackID, revSong)
}
