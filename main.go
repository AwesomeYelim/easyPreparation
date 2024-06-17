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

type songInfo struct {
	trackId int
}

func (si songInfo) searchLyricsList(recvStr string) {
	// 검색어 설정
	searchQuery := url.QueryEscape(recvStr)
	searchUrl := fmt.Sprintf("https://music.bugs.co.kr/search/lyrics?q=%s", searchQuery)
	fmt.Println(searchUrl)
	// HTTP 요청 보내기
	resp, err := http.Get(searchUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	// HTML 파싱
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// 각 tr 요소를 탐색하며 albumid와 mark 태그의 텍스트를 추출
	doc.Find("tr[rowtype='lyrics']").Each(func(i int, s *goquery.Selection) {
		albumID, exists := s.Attr("trackId")
		if exists {
			tempNo, err := strconv.ParseInt(albumID, 10, 64)
			if err != nil {
				log.Fatal(err)
			}
			si.trackId = int(tempNo)
			fmt.Printf("AlbumID: %s ", albumID)
		}
	})
}
func init() {
	// Make sure to load your metered License API key prior to using the library.
	// If you need a key, you can sign up and create a free one at https://cloud.unidoc.io
	//468eb71b0f562ed29385b487b55d413ad506b3c48950ead1de75bd736c7c17c4
	//err := license.SetMeteredKey(os.Getenv(`UNIDOC_LICENSE_API_KEY`))
	err := license.SetMeteredKey("468eb71b0f562ed29385b487b55d413ad506b3c48950ead1de75bd736c7c17c4")
	if err != nil {
		panic(err)
	}
}

func main() {
	// 슬라이드 데이터 생성
	slides := []SlideData{
		{"Title 1", "Content 1"},
		{"Title 2", "Content 2"},
		{"Title 3", "Content 3"},
	}

	// 프레젠테이션 생성
	ppt := presentation.New()

	defer ppt.Close()

	// 슬라이드 추가 및 데이터 설정
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

	// 프레젠테이션 저장
	filePath := "output.pptx"

	err := ppt.SaveToFile(filePath)
	if err != nil {
		fmt.Println("Error saving presentation:", err)
		return
	}

	sample := songInfo{}
	sample.searchLyricsList("하나님은 너를 지키시는자")

	fmt.Println("Presentation saved to", filePath)
}
