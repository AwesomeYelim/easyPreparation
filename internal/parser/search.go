package parser

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// gasazipBaseURL — 가사 검색 출처 (가사집)
const gasazipBaseURL = "https://gasazip.com"

// lyricsHTTPClient — 가사 사이트 접속 불가 시 무한 대기 방지용 타임아웃 클라이언트
var lyricsHTTPClient = &http.Client{Timeout: 8 * time.Second}

// SearchLyricsList — 가사집(gasazip.com)에서 제목으로 곡을 찾아 가사를 채운다.
// 검색 결과 중 제목이 query 로 시작하는 첫 곡을 쓴다. 맞는 곡이 없으면 Lyrics 는 빈 값으로 남는다.
func (si *SlideData) SearchLyricsList(query string) error {
	if len(si.Content) > 0 {
		return nil
	}

	searchDoc, err := fetchDoc(gasazipBaseURL + "/search.html?q=" + url.QueryEscape(query))
	if err != nil {
		return err
	}
	songPath := findSongPath(searchDoc, query)
	if songPath == "" {
		return nil
	}

	songDoc, err := fetchDoc(gasazipBaseURL + songPath)
	if err != nil {
		return err
	}
	si.parseLyrics(songDoc)
	return nil
}

// fetchDoc — URL 을 받아 HTML 문서로 파싱
func fetchDoc(pageURL string) (*goquery.Document, error) {
	resp, err := lyricsHTTPClient.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 오류: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTML 파싱 실패: %v", err)
	}
	return doc, nil
}
