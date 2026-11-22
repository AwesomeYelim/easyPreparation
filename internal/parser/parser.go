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

// findSongPath — 가사집 검색 결과에서 제목이 query 로 시작하는 첫 곡의 경로("/12345")를 반환.
// 띄어쓰기는 무시하고 비교한다. 없으면 빈 문자열.
func findSongPath(doc *goquery.Document, query string) string {
	want := strings.Join(strings.Fields(query), "")
	if want == "" {
		return ""
	}
	var songPath string
	doc.Find("a.gz-candy-song-row").EachWithBreak(func(i int, s *goquery.Selection) bool {
		title := strings.Join(strings.Fields(s.Find(".gz-candy-row-main > strong").Text()), "")
		if !strings.HasPrefix(title, want) {
			return true
		}
		songPath, _ = s.Attr("href")
		return false
	})
	return songPath
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
