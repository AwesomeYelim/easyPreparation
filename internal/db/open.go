package db

import (
	"fmt"

	"github.com/timshannon/bolthold"
)

// Bolthold 데이터베이스 열기 함수
func OpenDB(path string) (*bolthold.Store, error) {
	store, err := bolthold.Open(path, 0666, nil)
	if err != nil {
		return nil, fmt.Errorf("DB 열기 오류: %w", err)
	}
	return store, nil
}
