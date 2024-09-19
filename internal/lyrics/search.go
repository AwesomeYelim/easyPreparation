package lyrics

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// searchLyricsList 함수는 가사 목록을 검색합니다.
func (si *SlideData) SearchLyricsList(baseUrl, query string, isDirect bool) {

	if len(si.Content) > 0 {
		return
	}
	searchUrl := formatSearchURL(baseUrl, query, isDirect)

	// HTTP 요청 보내기
	resp, err := http.Get(searchUrl)
	if err != nil {
		log.Fatalf("Failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("Status code error: %d %s", resp.StatusCode, resp.Status)
	}

	// HTML 파싱
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Failed to parse HTML: %v", err)
	}

	if isDirect {
		si.parseLyrics(doc)
	} else {
		si.parseTrackList(doc)
	}
}

// formatSearchURL 함수는 검색 URL을 생성합니다.
func formatSearchURL(baseUrl, query string, isDirect bool) string {
	if isDirect {
		return fmt.Sprintf(baseUrl, query)
	}
	searchQuery := url.QueryEscape(query)
	return fmt.Sprintf(baseUrl, searchQuery)
}
