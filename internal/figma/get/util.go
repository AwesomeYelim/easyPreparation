package get

import (
	"github.com/torie/figma"
	"io"
	"net/http"
)

func download(i figma.Image) (io.ReadCloser, error) {
	resp, err := http.Get(i.URL)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
