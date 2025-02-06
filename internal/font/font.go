package font

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type WebFontItem struct {
	Kind         string            `json:"kind"`
	Family       string            `json:"family"`
	Category     string            `json:"category"`
	Variants     []string          `json:"variants"`
	Subsets      []string          `json:"subsets"`
	Version      string            `json:"version"`
	LastModified string            `json:"lastModified"`
	Files        map[string]string `json:"files"`
}

type WebFontList struct {
	Kind  string        `json:"kind"`
	Items []WebFontItem `json:"items"`
}

func GetFont(name, weight string) (fontPath string, err error) {

	gistURL := "https://gist.github.com/AwesomeYelim/6dc8fbbf1f2db5ad7d84cf95dbadafa7/raw"

	resp, err := http.Get(gistURL)
	if err != nil {
		log.Fatalf("Error fetching the gist: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	var webFontList WebFontList
	err = json.Unmarshal(body, &webFontList)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}

	var targetUrl string
	// JSON 데이터를 출력 (예시)
	for _, font := range webFontList.Items {
		if font.Family == name {
			targetUrl = font.Files[weight]
			//fmt.Printf("Font Family: %s\n", font.Family)
			//fmt.Printf("Category: %s\n", font.Category)
			//fmt.Printf("Variants: %v\n", font.Variants)
			//fmt.Printf("Files: %v\n", font.Files)
			break
		}
	}

	targetUrl = fmt.Sprintf("https:%s", targetUrl)
	fontPath, err = downloadFont(targetUrl)

	return fontPath, err
}

func downloadFont(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("폰트 다운로드 실패: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	tmpFile, err := os.CreateTemp("", "NanumGothic-*.ttf")
	if err != nil {
		return "", fmt.Errorf("임시 파일 생성 실패: %w", err)
	}
	defer func() {
		_ = tmpFile.Close()
	}()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("폰트 저장 실패: %w", err)
	}
	return tmpFile.Name(), nil
}
