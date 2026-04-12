//go:build dev

package main

import "io/fs"

// 개발 모드에서는 embed 없음 — Next.js dev server 사용
func getFrontendFS() fs.FS {
	return nil
}
