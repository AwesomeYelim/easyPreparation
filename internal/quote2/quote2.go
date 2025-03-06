package quote2

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// 성경 구절 크롤링 함수 (특정 장 크롤링)
func getChapterVerses(bookIdx string, chapter int) (map[int]string, error) {

	url := fmt.Sprintf("https://goodtvbible.goodtv.co.kr/bible.asp?bible_idx=%s&jang_idx=%d&bible_version_1=2", bookIdx, chapter)
	fmt.Println("크롤링 대상 URL:", url)

	// HTTP GET 요청
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// HTML 파싱
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	// `p#one_jang` 태그 내부의 HTML 가져오기
	selection := doc.Find("p#one_jang")

	// **HTML 구조 확인**
	htmlContent, err := selection.Html()
	if err != nil {
		return nil, fmt.Errorf("본문을 찾을 수 없습니다")
	}

	// `<br>` 기준으로 구절을 나누기
	verses := strings.Split(htmlContent, "<br/>")

	// 절 번호별로 맵핑
	versesMap := make(map[int]string)

	for _, v := range verses {
		v = strings.TrimSpace(v) // 앞뒤 공백 제거

		// **각 구절에서 숫자 절 번호를 추출하여 비교**
		if strings.HasPrefix(v, "<b>") {
			endIdx := strings.Index(v, "</b>")
			if endIdx > -1 {
				// `<b>숫자.</b>` 부분 추출 후 공백 제거
				prefix := v[3:endIdx] // "<b>1.</b>" → "1."
				prefix = strings.TrimSpace(prefix)

				// 절 번호 변환
				var verseNum int
				_, err := fmt.Sscanf(prefix, "%d.", &verseNum)
				if err == nil {
					// 절 번호가 존재하면 텍스트 저장
					text := v[endIdx+4:] // "</b>" 이후부터가 구절 내용
					text = strings.TrimSpace(text)

					// HTML 태그 제거
					textDoc, _ := goquery.NewDocumentFromReader(strings.NewReader(text))
					cleanText := textDoc.Text()
					cleanText = strings.ReplaceAll(text, "\u00A0", " ")
					versesMap[verseNum] = strings.TrimSpace(cleanText)
				}
			}
		}
	}

	return versesMap, nil
}

func getBibleVerses(bookIdx string, startChapter, startVerse, endChapter, endVerse int) (string, error) {
	var result []string

	// 현재 장부터 끝 장까지 루프 실행
	for chapter := startChapter; chapter <= endChapter; chapter++ {
		// 해당 장의 모든 절을 가져오기
		versesMap, err := getChapterVerses(bookIdx, chapter)
		if err != nil {
			continue // 에러가 발생하면 넘어감
		}

		// 시작 절과 끝 절 결정
		minVerse, maxVerse := 1, 150 // 최대절은 큰 값으로 설정
		if chapter == startChapter {
			minVerse = startVerse
		}
		if chapter == endChapter {
			maxVerse = endVerse
		}

		// 해당 범위의 절 가져오기
		for i := minVerse; i <= maxVerse; i++ {
			if verseText, exists := versesMap[i]; exists {
				result = append(result, fmt.Sprintf("%d장 %d절: %s", chapter, i, verseText))
			}
		}
	}

	if len(result) == 0 {
		return "", fmt.Errorf("구절을 찾을 수 없습니다: %s %d:%d ~ %d:%d", bookIdx, startChapter, startVerse, endChapter, endVerse)
	}

	return strings.Join(result, "\n"), nil
}

func GetQuote(forUrl string) {
	var startChapter int
	var startVerse int
	var endChapter int
	var endVerse int

	referBible := strings.Split(forUrl, "/")
	// 예제: 창세기 1장 2절 ~ 2장 10절 크롤링
	bookIdx := referBible[0]
	quoteRange := referBible[1]
	if strings.Contains(quoteRange, "-") {
		qCVs := strings.Split(quoteRange, "-")
		start := strings.Split(qCVs[0], ":")
		end := strings.Split(qCVs[1], ":")

		startChapter, _ = strconv.Atoi(start[0])
		startVerse, _ = strconv.Atoi(start[1])
		endChapter, _ = strconv.Atoi(end[0])
		endVerse, _ = strconv.Atoi(end[1])
	} else {
		start := strings.Split(quoteRange, ":")
		startChapter, _ = strconv.Atoi(start[0])
		startVerse, _ = strconv.Atoi(start[1])
	}

	// 성경 구절 가져오기
	versesText, err := getBibleVerses(bookIdx, startChapter, startVerse, endChapter, endVerse)
	if err != nil {
		log.Fatal(err)
	}

	// 결과 출력
	fmt.Printf("%s %d:%d ~ %d:%d:\n%s\n", bookIdx, startChapter, startVerse, endChapter, endVerse, versesText)
}
