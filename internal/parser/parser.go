package parser

import (
	"easyPreparation_1.0/internal/utils"
	"fmt"
	"github.com/PuerkitoBio/goquery"
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
func (si *SlideData) parseTrackList(doc *goquery.Document) error {
	var retErr error
	doc.Find("table.trackList tbody tr[rowtype='lyrics']").Each(func(i int, s *goquery.Selection) {
		if retErr != nil {
			return
		}
		albumID, exists := s.Attr("trackid")
		if !exists {
			fmt.Println("trackId가 없습니다")
			return
		}
		tempNo, err := strconv.ParseInt(albumID, 10, 64)
		if err != nil {
			retErr = fmt.Errorf("트랙 ID 파싱 실패: %v", err)
			return
		}
		si.TrackID = int(tempNo)
		if err := si.SearchLyricsList("https://music.bugs.co.kr/track/%s", albumID, true); err != nil {
			retErr = err
		}
	})
	return retErr
}

// 가사 파싱
func (si *SlideData) parseLyrics(doc *goquery.Document) {
	doc.Find(".lyricsContainer xmp").Each(func(i int, s *goquery.Selection) {
		si.Lyrics = s.Text()
		si.Content = utils.SplitTwoLines(s.Text())
	})
}
