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
