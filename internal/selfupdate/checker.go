package selfupdate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const githubRepo = "AwesomeYelim/easyPreparation"

// ReleaseAsset — GitHub Release 첨부 파일
type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
	ContentType        string `json:"content_type"`
}

// Release — GitHub Releases API 응답 구조
type Release struct {
	TagName string         `json:"tag_name"`
	HTMLURL string         `json:"html_url"`
	Body    string         `json:"body"`
	Name    string         `json:"name"`
	Assets  []ReleaseAsset `json:"assets"`
}

// FindAsset — 현재 실행 플랫폼에 맞는 바이너리 Asset을 반환합니다.
// server 바이너리를 우선 탐색하고, 없으면 desktop → legacy 이름 순으로 fallback합니다.
func (r *Release) FindAsset() *ReleaseAsset {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// 후보 이름 목록 (우선순위순)
	var candidates []string
	if goos == "windows" {
		candidates = []string{
			fmt.Sprintf("easyPreparation_desktop_%s_%s.exe", goos, goarch),  // desktop raw exe (우선)
			fmt.Sprintf("easyPreparation_server_%s_%s.exe", goos, goarch),   // server fallback
			fmt.Sprintf("easyPreparation_%s_%s.exe", goos, goarch),          // legacy
		}
	} else {
		candidates = []string{
			fmt.Sprintf("easyPreparation_server_%s_%s", goos, goarch),
			fmt.Sprintf("easyPreparation_desktop_%s_%s.zip", goos, goarch),
			fmt.Sprintf("easyPreparation_desktop_%s_%s", goos, goarch),
			fmt.Sprintf("easyPreparation_%s_%s", goos, goarch), // legacy
		}
	}

	for _, name := range candidates {
		for i := range r.Assets {
			if r.Assets[i].Name == name {
				return &r.Assets[i]
			}
		}
	}
	return nil
}

// ChecksumsAsset — checksums.txt Asset을 반환합니다.
func (r *Release) ChecksumsAsset() *ReleaseAsset {
	for i := range r.Assets {
		if r.Assets[i].Name == "checksums.txt" {
			return &r.Assets[i]
		}
	}
	return nil
}

// CheckLatest — GitHub Releases API로 최신 릴리즈를 조회합니다.
// 네트워크 오류 또는 릴리즈가 없으면 error를 반환합니다.
func CheckLatest() (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("요청 생성 실패: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub API 호출 실패: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("릴리즈를 찾을 수 없습니다 (저장소: %s)", githubRepo)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 오류: HTTP %d", resp.StatusCode)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %w", err)
	}
	if rel.TagName == "" {
		return nil, fmt.Errorf("유효하지 않은 릴리즈 응답")
	}
	return &rel, nil
}

// IsNewer — current 버전보다 latest 버전이 최신인지 판단합니다.
// semver 형식(예: v1.2.3)을 권장합니다.
func IsNewer(current, latest string) bool {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")
	// dev 빌드는 항상 업데이트 알림 표시
	if current == "dev" || current == "" {
		return true
	}
	return compareSemver(latest, current) > 0
}

// compareSemver — a > b이면 양수, 같으면 0, 작으면 음수
func compareSemver(a, b string) int {
	pa := parseSemver(a)
	pb := parseSemver(b)
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			return pa[i] - pb[i]
		}
	}
	return 0
}

func parseSemver(v string) [3]int {
	// "1.2.3-beta" → "1.2.3"
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}
	parts := strings.SplitN(v, ".", 3)
	var result [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		result[i], _ = strconv.Atoi(parts[i])
	}
	return result
}
