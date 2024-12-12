package get

import (
	"encoding/json"
	"fmt"
	"github.com/torie/figma"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func download(i figma.Image) (io.ReadCloser, error) {
	resp, err := http.Get(i.URL)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func orgJson(argResult []map[string]interface{}, execPath string, target string) map[string][]Children {
	grouped := make(map[string][]Children)

	for _, contentResult := range argResult {
		if name, ok := contentResult["name"].(string); ok {
			switch {
			case name == target:
				if children, ok := contentResult["children"].([]interface{}); ok {
					grouped = orgJson(convertToMapSlice(children), execPath, target)
				}
			case isValidPattern(name):
				grouped[name] = extractChildren(contentResult, name)
			}
		}
	}

	return grouped
}

// 숫자_이름 형식인지 확인
func isValidPattern(name string) bool {
	for i, r := range name {
		if i == 0 && !unicode.IsDigit(r) {
			return false
		}
		if r == '_' {
			return i > 0 && len(name) > i+1
		}
	}
	return false
}

// 하위 항목에서 자식 요소 추출
func extractChildren(contentResult map[string]interface{}, name string) []Children {
	var children []Children

	if childItems, ok := contentResult["children"].([]interface{}); ok {
		for _, child := range childItems {
			if childMap, ok := child.(map[string]interface{}); ok {
				cName, nOk := childMap["name"].(string)
				characters, cOk := childMap["characters"].(string)

				if nOk && cOk {
					if strings.HasSuffix(name, "부름") || strings.HasSuffix(name, "봉독") {
						children = append(children, Children{
							Content: characters,
							Info:    cName,
							Obj:     displayQuote(characters),
						})

					} else {
						children = append(children, Children{
							Content: characters,
							Info:    cName,
						})
					}

				}

			}
		}
	}

	return children
}

func StripTags(html string) string {
	html = strings.ReplaceAll(html, "\u003c", "<")
	html = strings.ReplaceAll(html, "\u003e", ">")
	html = strings.ReplaceAll(html, "&nbsp;", " ")
	html = strings.ReplaceAll(html, "\r\n", "\n")

	lines := strings.Split(html, "\n")
	var result []string

	for _, line := range lines {
		line = removeTags(line)

		if strings.TrimSpace(line) != "" {
			result = append(result, strings.TrimSpace(line))
		}
	}

	return strings.Join(result, "\n")
}

func removeTags(input string) string {
	var output strings.Builder
	inTag := false
	for _, char := range input {
		if char == '<' {
			inTag = true
		} else if char == '>' {
			inTag = false
		} else if !inTag {
			output.WriteRune(char)
		}
	}
	return output.String()
}

func displayQuote(characters string) string {

	var dictB map[string]struct {
		Eng string `json:"eng"`
		Kor string `json:"kor"`
	}

	dict, err := os.ReadFile(filepath.Join("./config", "bible_dict.json"))
	err = json.Unmarshal(dict, &dictB)
	char := strings.Split(characters, " ")
	cover := ExpandRange(char[1])

	url := fmt.Sprintf("http://ibibles.net/quote.php?kor-%s/%s", dictB[char[0]].Eng, cover)

	// HTTP GET 요청 생성
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	defer resp.Body.Close() // 응답이 끝난 후 자원 해제

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed to fetch data. Status code:", resp.StatusCode)
		return ""
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return ""
	}

	return StripTags(string(body))
}

func ExpandRange(input string) string {
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return input
	}

	prefix := parts[0]
	rangePart := parts[1]

	rangeNumbers := strings.Split(rangePart, "-")
	if len(rangeNumbers) != 2 {
		return input
	}

	return fmt.Sprintf("%s:%s-%s:%s", prefix, rangeNumbers[0], prefix, rangeNumbers[1])
}

// []interface{} => []map[string]interface{}
func convertToMapSlice(data []interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}
	return result
}
