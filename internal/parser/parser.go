package parser

import (
	"strings"

	"easyPreparation_1.0/internal/utils"
	"github.com/PuerkitoBio/goquery"
)

// SlideData 구조체는 슬라이드에 포함될 데이터를 나타냅니다.
type SlideData struct {
	Title   string
	Content []string
	Lyrics  string
}

// findSongPath — 가사집 검색 결과에서 곡 경로("/12345")를 고른다. 띄어쓰기는 무시하고 비교한다.
// 1순위: 제목이 query 로 시작하는 첫 곡. 2순위: 가사 미리보기에 query 가 그대로 들어 있는 첫 곡
// (첫 소절을 제목처럼 쓰는 곡). 둘 다 없으면 빈 문자열 — 검색 첫 결과를 그냥 쓰면 절반이 다른 곡이었다.
func findSongPath(doc *goquery.Document, query string) string {
	want := squash(query)
	if want == "" {
		return ""
	}
	var byTitle, byLyrics string
	doc.Find("a.gz-candy-song-row").EachWithBreak(func(i int, s *goquery.Selection) bool {
		href, _ := s.Attr("href")
		if strings.HasPrefix(squash(s.Find(".gz-candy-row-main > strong").Text()), want) {
			byTitle = href
			return false
		}
		if byLyrics == "" && strings.Contains(squash(s.Find(".gz-candy-snippet").Text()), want) {
			byLyrics = href
		}
		return true
	})
	if byTitle != "" {
		return byTitle
	}
	return byLyrics
}

// squash — 공백을 모두 제거 (띄어쓰기 차이를 무시하고 비교하기 위함)
func squash(s string) string {
	return strings.Join(strings.Fields(s), "")
}

// parseLyrics — 가사집 곡 페이지의 #gasa-copy 에서 가사를 읽는다.
// 원문이 "<br />\n" 형태라 Text() 만으로 줄바꿈이 유지된다 (<br> 를 따로 바꾸면 줄이 두 배가 됨).
func (si *SlideData) parseLyrics(doc *goquery.Document) {
	var lines []string
	for _, line := range strings.Split(doc.Find("#gasa-copy").Text(), "\n") {
		lines = append(lines, strings.TrimSpace(line))
	}
	lyrics := strings.TrimSpace(strings.Join(lines, "\n"))
	if lyrics == "" {
		return
	}
	si.Lyrics = lyrics
	si.Content = utils.SplitTwoLines(lyrics)
}
