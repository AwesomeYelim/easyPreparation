//go:build dev

package embedded

import "io/fs"

// 개발 모드: 프론트는 Next.js dev server, 데이터는 저장소 로컬 파일을 직접 사용한다.

func FrontendFS() fs.FS { return nil }

func DataFS() fs.FS { return nil }
