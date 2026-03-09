package parser

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// SearchLyricsList 함수는 가사 목록을 검색합니다.
func (si *SlideData) SearchLyricsList(baseUrl, query string, isDirect bool) error {
	if len(si.Content) > 0 {
		return nil
	}
	searchUrl := formatSearchURL(baseUrl, query, isDirect)

	resp, err := http.Get(searchUrl)
	if err != nil {
		return fmt.Errorf("HTTP 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP 오류: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("HTML 파싱 실패: %v", err)
	}

	if isDirect {
		si.parseLyrics(doc)
	} else {
		if err := si.parseTrackList(doc); err != nil {
			return err
		}
	}
	return nil
}

// formatSearchURL 함수는 검색 URL을 생성합니다.
func formatSearchURL(baseUrl, query string, isDirect bool) string {
	if isDirect {
		return fmt.Sprintf(baseUrl, query)
	}
	searchQuery := url.QueryEscape(query)
	return fmt.Sprintf(baseUrl, searchQuery)
}
