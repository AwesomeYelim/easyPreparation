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

// SearchLyricsList — 가사집(gasazip.com)에서 곡을 찾아 가사를 채운다.
// query 는 "제목" 또는 "제목 - 가수". 고르는 기준은 findSongPath 참고.
// 맞는 곡이 없으면 Lyrics 는 빈 값으로 남는다.
func (si *SlideData) SearchLyricsList(query string) error {
	if len(si.Content) > 0 {
		return nil
	}

	// 가사집은 검색어에 가수를 넣어도 결과가 같아서 제목으로만 검색한다.
	title, artist := SplitQuery(query)

	searchDoc, err := fetchDoc(gasazipBaseURL + "/search.html?q=" + url.QueryEscape(title))
	if err != nil {
		return err
	}
	songPath := findSongPath(searchDoc, title, artist)
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
