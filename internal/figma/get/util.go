package get

import (
	"easyPreparation_1.0/pkg"
	"encoding/json"
	"github.com/torie/figma"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func download(i figma.Image) (io.ReadCloser, error) {
	resp, err := http.Get(i.URL)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// orgJson은 그룹화된 JSON 결과를 반환
func orgJson(argResult []map[string]interface{}, execPath string) map[string][]Children {
	grouped := make(map[string][]Children)

	for _, contentResult := range argResult {
		if name, ok := contentResult["name"].(string); ok {
			switch {
			case name == "content_1", name == "content_2", name == "content_3":
				processContent(name, contentResult, execPath)
			case strings.HasPrefix(name, "sub_"):
				grouped[name] = extractChildren(contentResult)
			}
		}
	}

	return grouped
}

// 특정 content 이름에 따라 그룹화된 결과를 파일로 저장
func processContent(name string, contentResult map[string]interface{}, execPath string) {
	if children, ok := contentResult["children"].([]interface{}); ok {
		result := orgJson(convertToMapSlice(children), execPath)
		final := createSortedResult(result)

		sample, _ := json.MarshalIndent(final, "", "  ")
		_ = pkg.CheckDirIs(filepath.Join(execPath, "config"))
		_ = os.WriteFile(filepath.Join(execPath, "config", name+".json"), sample, 0644)
	}
}

// 그룹화된 결과를 정렬하여 리스트 형식으로 반환
func createSortedResult(groupedResults map[string][]Children) []Element {
	sortedKeys := sortKeys(groupedResults)
	final := make([]Element, 0, len(sortedKeys))
	for _, name := range sortedKeys {
		final = append(final, Element{
			Name:     name,
			Children: groupedResults[name],
		})
	}
	return final
}

// sub_ 항목에서 자식 요소 추출
func extractChildren(contentResult map[string]interface{}) []Children {
	var children []Children

	if childItems, ok := contentResult["children"].([]interface{}); ok {
		for _, child := range childItems {
			if childMap, ok := child.(map[string]interface{}); ok {
				if cName, cOk := childMap["name"].(string); cOk {
					if characters, ok := childMap["characters"].(string); ok {
						children = append(children, Children{
							ChildType:  cName,
							Characters: characters,
						})
					}
				}

				// children 존재할 경우 재귀
				if nestedChildren, ok := childMap["children"].([]interface{}); ok {
					nestedResult := extractChildren(map[string]interface{}{
						"children": nestedChildren,
					})
					children = append(children, nestedResult...)
				}
			}
		}
	}

	return children
}

// sub_ 이름의 숫자 순서로 정렬
func sortKeys(groupedResults map[string][]Children) []string {
	keys := make([]string, 0, len(groupedResults))
	for name := range groupedResults {
		keys = append(keys, name)
	}

	sort.Slice(keys, func(i, j int) bool {
		numI, _ := strconv.Atoi(strings.TrimPrefix(keys[i], "sub_"))
		numJ, _ := strconv.Atoi(strings.TrimPrefix(keys[j], "sub_"))
		return numI < numJ
	})

	return keys
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
