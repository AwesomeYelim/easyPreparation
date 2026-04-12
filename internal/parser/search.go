package parser

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// SearchLyricsList 함수는 가사 목록을 검색합니다.
func (si *SlideData) SearchLyricsList(baseUrl, query string, isDirect bool) error {
	if len(si.Content) > 0 {
		return nil
	}

	// 1차: 원본 쿼리로 가사 검색
	if err := si.doSearch(baseUrl, query, isDirect); err != nil {
		return err
	}
	if si.Lyrics != "" {
		return nil
	}

	// 2차: 가사 검색 실패 → 트랙(곡명) 검색으로 재시도 (띄어쓰기 더 유연함)
	if !isDirect {
		log.Printf("[lyrics] 가사검색 실패, 트랙검색 재시도: %q", query)
		trackURL := "https://music.bugs.co.kr/search/track?q=%s"
		if err := si.doSearch(trackURL, query, false); err != nil {
			log.Printf("[lyrics] 트랙검색도 실패: %v", err)
		}
	}

	return nil
}

// doSearch — 실제 HTTP 검색 + 파싱
func (si *SlideData) doSearch(baseUrl, query string, isDirect bool) error {
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
