//go:build dev

package main

import "io/fs"

// 개발 모드에서는 embed 데이터 없음 — 로컬 파일 직접 사용
func getEmbeddedDataFS() fs.FS {
	return nil
}
