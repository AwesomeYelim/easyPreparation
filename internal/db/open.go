package db

import (
	"github.com/timshannon/bolthold"
	"log"
)

// Bolthold 데이터베이스 열기 함수
func OpenDB(path string) *bolthold.Store {
	store, err := bolthold.Open(path, 0666, nil)
	if err != nil {
		log.Fatalf("DB 열기 오류: %v", err)
	}
	return store
}
