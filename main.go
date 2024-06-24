package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/unidoc/unioffice/common/license"
	"github.com/unidoc/unioffice/presentation"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

// SlideData 구조체는 슬라이드에 포함될 데이터를 나타냅니다.
type SlideData struct {
	Title   string
	Content string
}

type SongInfo struct {
	TrackID int
	Lyrics  string
}

// searchLyricsList 함수는 가사 목록을 검색합니다.
func (si *SongInfo) searchLyricsList(baseUrl, query string, isDirect bool) {
	if si.Lyrics != "" {
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

// parseTrackList 함수는 트랙 리스트를 파싱합니다.
func (si *SongInfo) parseTrackList(doc *goquery.Document) {
	doc.Find("table.trackList tbody tr[rowtype='lyrics']").Each(func(i int, s *goquery.Selection) {
		albumID, exists := s.Attr("trackid")
		if exists {
			tempNo, err := strconv.ParseInt(albumID, 10, 64)
			if err != nil {
				log.Fatalf("Failed to parse track ID: %v", err)
			}
			si.TrackID = int(tempNo)
			si.searchLyricsList("https://music.bugs.co.kr/track/%s", albumID, true)

		}

	})
}

// parseLyrics 함수는 가사를 파싱합니다.
func (si *SongInfo) parseLyrics(doc *goquery.Document) {
	doc.Find(".lyricsContainer xmp").Each(func(i int, s *goquery.Selection) {
		fmt.Println(s.Text())
		si.Lyrics = s.Text()
	})
}

func init() {
	// UniOffice 라이센스 설정
	err := license.SetMeteredKey("468eb71b0f562ed29385b487b55d413ad506b3c48950ead1de75bd736c7c17c4")
	if err != nil {
		panic(fmt.Sprintf("Failed to set UniOffice license key: %v", err))
	}
}

func main() {
	// 슬라이드 데이터 생성
	slides := []SlideData{
		{"Title 1", "Content 1"},
		{"Title 2", "Content 2"},
		{"Title 3", "Content 3"},
	}

	// 프레젠테이션 생성 및 슬라이드 추가
	createPresentation(slides, "output.pptx")

	// 가사 검색
	song := &SongInfo{}
	song.searchLyricsList("https://music.bugs.co.kr/search/lyrics?q=%s", "하나님은 너를 지키시는자", false)

	fmt.Println("Presentation saved to output.pptx")
}

// createPresentation 함수는 프레젠테이션을 생성하고 슬라이드를 추가합니다.
func createPresentation(slides []SlideData, filePath string) {
	ppt := presentation.New()
	defer ppt.Close()

	for _, slideData := range slides {
		slide := ppt.AddSlide()

		// 제목 설정
		titleBox := slide.AddTextBox()
		titlePara := titleBox.AddParagraph()
		titleRun := titlePara.AddRun()
		titleRun.SetText(slideData.Title)
		titleBox.Properties().SetPosition(50, 50)
		titleBox.Properties().SetSize(50, 50)

		// 내용 설정
		contentBox := slide.AddTextBox()
		contentPara := contentBox.AddParagraph()
		contentRun := contentPara.AddRun()
		contentRun.SetText(slideData.Content)
		contentBox.Properties().SetPosition(50, 50)
		contentBox.Properties().SetSize(50, 50)
	}

	if err := ppt.SaveToFile(filePath); err != nil {
		log.Fatalf("Error saving presentation: %v", err)
	}
}
