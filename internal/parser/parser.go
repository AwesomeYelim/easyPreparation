package parser

import (
	"easyPreparation_1.0/pkg"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strconv"
)

// SlideData 구조체는 슬라이드에 포함될 데이터를 나타냅니다.
type SlideData struct {
	Title   string
	Content []string
	TrackID int
	Lyrics  string
}

// 트랙 리스트를 파싱
func (si *SlideData) parseTrackList(doc *goquery.Document) {
	doc.Find("table.trackList tbody tr[rowtype='lyrics']").Each(func(i int, s *goquery.Selection) {
		albumID, exists := s.Attr("trackid")
		if exists {
			tempNo, err := strconv.ParseInt(albumID, 10, 64)
			if err != nil {
				log.Fatalf("Failed to parse track ID: %v", err)
			}
			si.TrackID = int(tempNo)
			si.SearchLyricsList("https://music.bugs.co.kr/track/%s", albumID, true)
		} else {
			fmt.Println("the trackId is not exist")
		}
	})
}

// 가사 파싱
func (si *SlideData) parseLyrics(doc *goquery.Document) {
	doc.Find(".lyricsContainer xmp").Each(func(i int, s *goquery.Selection) {
		si.Lyrics = s.Text()
		si.Content = pkg.SplitTwoLines(s.Text())
	})
}
