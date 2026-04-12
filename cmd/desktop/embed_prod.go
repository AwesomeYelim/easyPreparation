//go:build !dev

package main

import (
	"embed"
	"io/fs"
)

//go:embed all:frontend
var frontendDist embed.FS

func getFrontendFS() fs.FS {
	sub, err := fs.Sub(frontendDist, "frontend")
	if err != nil {
		return nil
	}
	return sub
}

// getUIBaseURL — prod 모드에서 WebView가 리디렉션할 UI URL
func getUIBaseURL() string {
	return "http://localhost:8080"
}
