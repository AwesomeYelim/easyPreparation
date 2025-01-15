package get

import (
	"easyPreparation_1.0/internal/quote"
	"github.com/torie/figma"
	"io"
	"net/http"
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
							Obj:     quote.GetQuote(characters),
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
