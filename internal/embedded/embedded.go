//go:build !dev

// Package embedded — 서버/데스크톱 공용 embed 자산 (frontend/ = Next.js static export, data/ = 기본 데이터).
// Go embed는 패키지 디렉터리 하위만 묶을 수 있으므로 두 엔트리포인트가 여기 한 곳을 공유한다.
package embedded

import (
	"embed"
	"io/fs"
)

//go:embed all:frontend
var frontendDist embed.FS

//go:embed all:data
var dataDist embed.FS

// FrontendFS — embed된 프론트엔드 정적 파일. 개발 빌드(-tags dev)에서는 nil.
func FrontendFS() fs.FS {
	sub, err := fs.Sub(frontendDist, "frontend")
	if err != nil {
		return nil
	}
	return sub
}

// DataFS — embed된 기본 데이터(bible.db, schema.sql, defaults/ ...). 개발 빌드에서는 nil.
func DataFS() fs.FS {
	sub, err := fs.Sub(dataDist, "data")
	if err != nil {
		return nil
	}
	return sub
}
