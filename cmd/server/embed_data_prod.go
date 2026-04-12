//go:build !dev

package main

import (
	"embed"
	"io/fs"
)

//go:embed all:data
var embeddedDataDist embed.FS

func getEmbeddedDataFS() fs.FS {
	sub, err := fs.Sub(embeddedDataDist, "data")
	if err != nil {
		return nil
	}
	return sub
}
