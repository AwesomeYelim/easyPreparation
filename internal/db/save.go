package db

import (
	"fmt"
	"github.com/timshannon/bolthold"
)

// Bolthold 데이터베이스에 노래 저장 함수
func SaveSongToDB(store *bolthold.Store, title string, content []string) error {
	var existingSong Song
	// 제목으로 기존 노래 조회
	err := store.FindOne(&existingSong, bolthold.Where("Title").Eq(title))
	if err == nil {
		// 기존 노래가 있는 경우, 내용을 업데이트
		existingSong.Content = append(existingSong.Content, content...) // 기존 내용에 추가
		return store.Update(existingSong.ID, &existingSong)             // 업데이트
	}

	// 새로운 노래 저장
	// 현재 데이터베이스의 항목 수를 기반으로 새로운 ID 생성
	count, err := store.Count(&Song{}, nil) // Song 타입의 항목 수를 계산
	if err != nil {
		return fmt.Errorf("ID 생성 실패: %v", err)
	}
	newID := count + 1 // ID는 현재 항목 수 + 1

	song := Song{ID: newID, Title: title, Content: content}
	return store.Insert(newID, &song)
}
