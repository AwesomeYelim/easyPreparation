package quote

import (
	"easyPreparation_1.0/internal/parser"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/pkg"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type BibleAPIResponse struct {
	Data struct {
		Testament   string `json:"testament"`
		Bookname    string `json:"bookname"`
		BooknameAbb string `json:"bookname_abb"`
		Data        struct {
			Version1 struct {
				Version        int    `json:"version"`
				Jang           int    `json:"jang"`
				VersionName    string `json:"version_name"`
				SoundtrackYn   string `json:"soundtrack_yn"`
				TranslationIdx int    `json:"translation_idx"`
				Bookname       string `json:"bookname"`
				BooknameAbb    string `json:"bookname_abb"`
				Theme          []struct {
					StartJul int    `json:"start_jul"`
					Cont     string `json:"cont"`
				} `json:"theme"`
				Content []struct {
					Jul            int    `json:"jul"`
					Text           string `json:"text"`
					DictionaryList []struct {
						Idx   int    `json:"idx"`
						Word  string `json:"word"`
						Word2 string `json:"word2"`
						Cont  string `json:"cont"`
					} `json:"dictionaryList"`
				} `json:"content"`
			} `json:"version1"`
		} `json:"data"`
	} `json:"data"`
}

func ProcessQuote(worshipTitle string, bulletin *[]map[string]interface{}) {
	i := 0
	for i < len(*bulletin) {
		el := (*bulletin)[i]

		title, tOk := el["title"].(string)
		info, iOk := el["info"].(string)
		obj, oOk := el["obj"].(string)

		if !tOk || !oOk {
			// title 또는 obj가 string이 아닐 경우 다음으로
			i++
			continue
		}

		// "b_"로 시작하는 info 필드가 있을 때만 처리
		if iOk && strings.HasPrefix(info, "b_") {
			// "성경봉독" 제목 뒤에 "말씀내용" 항목 삽입
			if title == "성경봉독" {
				newItem := map[string]interface{}{
					"key":   fmt.Sprintf("%d.1", i),
					"title": "말씀내용",
					"info":  "c_edit",
					"obj":   "-",
				}

				*bulletin = append((*bulletin)[:i+1], append([]map[string]interface{}{newItem}, (*bulletin)[i+1:]...)...)
				i++ // 삽입했으니 인덱스 증가
			}

			var sb strings.Builder
			var objRangeParts []string

			if strings.Contains(obj, ",") {
				refs := strings.Split(obj, ",")
				for _, qObj := range refs {
					qObj = strings.TrimSpace(qObj)
					parts := strings.SplitN(qObj, "_", 2)
					if len(parts) != 2 {
						continue // 포맷 이상 시 무시
					}
					kor, forUrl := parts[0], parts[1]

					quoteText := GetQuote(forUrl)

					sb.WriteString(quoteText)
					sb.WriteString("\n")

					chapterVerse := ""
					urlParts := strings.SplitN(forUrl, "/", 2)
					if len(urlParts) == 2 {
						chapterVerse = urlParts[1]
					}
					objRangeParts = append(objRangeParts, fmt.Sprintf("%s %s", kor, parser.CompressVerse(chapterVerse)))
				}
			} else {
				parts := strings.SplitN(obj, "_", 2)
				if len(parts) == 2 {
					kor, forUrl := parts[0], parts[1]
					quoteText := GetQuote(forUrl)

					sb.WriteString(quoteText)

					chapterVerse := ""
					urlParts := strings.SplitN(forUrl, "/", 2)
					if len(urlParts) == 2 {
						chapterVerse = urlParts[1]
					}
					objRangeParts = append(objRangeParts, fmt.Sprintf("%s %s", kor, parser.CompressVerse(chapterVerse)))
				}
			}

			objRange := strings.Join(objRangeParts, ", ")
			(*bulletin)[i]["contents"] = sb.String()
			(*bulletin)[i]["obj"] = objRange
		}

		// "말씀내용"으로 끝나는 title은 바로 앞 항목의 contents 복사 (범위 검사 포함)
		if strings.HasSuffix(title, "말씀내용") {
			if i-1 >= 0 {
				if prevContents, ok := (*bulletin)[i-1]["contents"]; ok {
					(*bulletin)[i]["contents"] = prevContents
				}
			}
		}

		i++
	}

	execPath := path.ExecutePath("easyPreparation")

	sample, _ := json.MarshalIndent(bulletin, "", "  ")
	_ = pkg.CheckDirIs(filepath.Join(execPath, "config"))
	_ = os.WriteFile(filepath.Join(execPath, "config", worshipTitle+".json"), sample, 0644)
}

func GetQuote(forUrl string) string {
	var startChapter, startVerse, endChapter, endVerse int

	referBible := strings.Split(forUrl, "/")
	if len(referBible) < 2 {
		log.Fatalf("잘못된 입력 형식입니다: %s (예: 1/1:2-3)", forUrl)
	}

	bookCode := referBible[0]
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
		endChapter, endVerse = startChapter, startVerse
	}

	versesText, err := getBibleVersesAPI(bookCode, startChapter, startVerse, endChapter, endVerse)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n📖 최종 결과:\n%s\n", versesText)
	return versesText
}

func getBibleVersesAPI(bookCode string, startChapter, startVerse, endChapter, endVerse int) (string, error) {
	var result []string

	for chapter := startChapter; chapter <= endChapter; chapter++ {
		minVerse := 1
		maxVerse := 150 // 적당히 큰 값 (시편 최대절 수)
		if chapter == startChapter {
			minVerse = startVerse
		}
		if chapter == endChapter {
			maxVerse = endVerse
		}

		versesMap, err := getBibleVersesByAPI(bookCode, chapter, minVerse, maxVerse)
		if err != nil {
			continue
		}

		for i := minVerse; i <= maxVerse; i++ {
			if text, ok := versesMap[i]; ok {
				result = append(result, fmt.Sprintf("%d:%d %s", chapter, i, text))
			}
		}
	}

	if len(result) == 0 {
		return "", fmt.Errorf("구절을 찾을 수 없습니다")
	}

	return strings.Join(result, "\n"), nil
}

func getBibleVersesByAPI(bookCode string, chapter int, startVerse, endVerse int) (map[int]string, error) {
	url := fmt.Sprintf("https://goodtvbible.goodtv.co.kr/api/onlinebible/bibleread/read-all?version1=0&version2=&version3=&bible_code=%s&jang=%d", bookCode, chapter)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	fmt.Println("reqUrl : ", url)
	defer resp.Body.Close()

	var apiResp BibleAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	versesMap := make(map[int]string)
	contents := apiResp.Data.Data.Version1.Content

	for _, verse := range contents {
		if verse.Jul >= startVerse && verse.Jul <= endVerse {
			versesMap[verse.Jul] = verse.Text
		}
	}
	return versesMap, nil
}
