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
// 1순위: 제목이 title 로 시작하고 가수에 artist 가 들어 있는 곡 (artist 를 준 경우만).
// 2순위: 제목이 title 로 시작하는 첫 곡.
// 3순위: 가사 미리보기에 title 이 그대로 들어 있는 첫 곡 (첫 소절을 제목처럼 쓰는 곡).
// 모두 없으면 빈 문자열 — 검색 첫 결과를 그냥 쓰면 절반이 다른 곡이었다.
func findSongPath(doc *goquery.Document, title, artist string) string {
	want, wantArtist := Squash(title), Squash(artist)
	if want == "" {
		return ""
	}
	var byArtist, byTitle, byLyrics string
	doc.Find("a.gz-candy-song-row").EachWithBreak(func(i int, s *goquery.Selection) bool {
		href, _ := s.Attr("href")
		main := s.Find(".gz-candy-row-main")
		if strings.HasPrefix(Squash(main.ChildrenFiltered("strong").Text()), want) {
			if byTitle == "" {
				byTitle = href
			}
			if wantArtist == "" {
				return false
			}
			if strings.Contains(Squash(main.ChildrenFiltered("em").Text()), wantArtist) {
				byArtist = href
				return false
			}
			return true
		}
		if byLyrics == "" && strings.Contains(Squash(s.Find(".gz-candy-snippet").Text()), want) {
			byLyrics = href
		}
		return true
	})
	if byArtist != "" {
		return byArtist
	}
	if byTitle != "" {
		return byTitle
	}
	return byLyrics
}

// Squash — 공백 제거 + 소문자화 (띄어쓰기·대소문자 차이를 무시하고 비교하기 위함)
func Squash(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), ""))
}

// SplitQuery — "제목 - 가수" 를 제목과 가수로 나눈다. 가수가 없으면 artist 는 빈 값.
func SplitQuery(query string) (title, artist string) {
	if i := strings.LastIndex(query, " - "); i >= 0 {
		return strings.TrimSpace(query[:i]), strings.TrimSpace(query[i+3:])
	}
	return strings.TrimSpace(query), ""
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
