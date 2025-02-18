package font

import (
	"easyPreparation_1.0/pkg"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func GetFont(name, weight string, isB bool) (fontPath string, err error) {

	gistURL := "https://gist.github.com/AwesomeYelim/6dc8fbbf1f2db5ad7d84cf95dbadafa7/raw"
	saveName := strings.Replace(name, " ", "", 1)
	saveName = fmt.Sprintf("%s-%s.ttf", saveName, weight)

	savePath := filepath.Join("./public", "font", saveName)

	if _, err = os.Stat(savePath); !os.IsNotExist(err) {
		return savePath, nil
	}

	_ = pkg.CheckDirIs(filepath.Dir(savePath))

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
			targetUrl = findTargetURL(font.Files, weight, isB)
			break
		}
	}

	targetUrl = fmt.Sprintf("https:%s", targetUrl)
	fontPath, err = downloadFont(savePath, targetUrl)

	return fontPath, err
}

// fontWeight 잘못 넣은 경우
func findTargetURL(files map[string]string, weight string, isB bool) string {
	if url, ok := files[weight]; ok && url != "" {
		return url
	}
	w, err := strconv.Atoi(weight)
	if err != nil || w <= 0 {
		w = 700
	}

	var newWeight string
	if isB {
		newWeight = strconv.Itoa(w + 100)
	} else {
		newWeight = strconv.Itoa(w - 100)
	}

	return findTargetURL(files, newWeight, isB)
}

func downloadFont(savePath, url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("폰트 다운로드 실패: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	tmpFile, err := os.Create(savePath)
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
