package get

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const figmaAPIBase = "https://api.figma.com/v1"

func figmaGet(token, url string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Figma-Token", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// fetchNodes는 Figma 파일의 최상위 노드 목록과 최신 버전 ID를 반환합니다.
func fetchNodes(key, token string) ([]Node, string, error) {
	var res struct {
		Document struct {
			Children []Node `json:"children"`
		} `json:"document"`
		Version string `json:"version"`
	}
	url := fmt.Sprintf("%s/files/%s?depth=5", figmaAPIBase, key)
	if err := figmaGet(token, url, &res); err != nil {
		return nil, "", err
	}
	return res.Document.Children, res.Version, nil
}

// fetchImageURL은 특정 노드의 PNG 내보내기 URL을 반환합니다.
func fetchImageURL(key, token, nodeID string) (string, error) {
	var res struct {
		Images map[string]string `json:"images"`
		Err    string            `json:"err"`
	}
	url := fmt.Sprintf("%s/images/%s?ids=%s&format=png&scale=2", figmaAPIBase, key, nodeID)
	if err := figmaGet(token, url, &res); err != nil {
		return "", err
	}
	if res.Err != "" {
		return "", fmt.Errorf("Figma images API 오류: %s", res.Err)
	}
	imgURL, ok := res.Images[nodeID]
	if !ok || imgURL == "" {
		return "", fmt.Errorf("노드 %s의 이미지 URL을 찾을 수 없습니다", nodeID)
	}
	return imgURL, nil
}

// downloadImage는 이미지 URL에서 데이터를 다운로드합니다.
// Figma CDN은 X-Figma-Token 헤더가 필요합니다.
func downloadImage(imgURL, token string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, imgURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Figma-Token", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("이미지 다운로드 실패 (HTTP %d): %s", resp.StatusCode, imgURL)
	}
	return io.ReadAll(resp.Body)
}
