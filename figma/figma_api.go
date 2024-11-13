package figma

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/torie/figma"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func GetFigmaImage(at *string, key *string, help *bool) {
	var mainContent []map[string]interface{}

	if *at == "" || *key == "" {
		flag.Usage()
		os.Exit(-1)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	c := figma.New(*at)

	f, err := c.File(*key)
	if err != nil {
		log.Println(err)
	}

	docs := f.Nodes()
	log.Printf("Got %d documents", len(docs))

	sample, _ := json.MarshalIndent(docs, "", "")
	err = json.Unmarshal(sample, &mainContent)

	if err != nil {
		log.Print("err : ", err)
	}

	_ = orgJson(mainContent)

	//sample, _ = json.MarshalIndent(result, "", "")
	//_ = os.WriteFile("./schema.json", sample, 0644)

	frameDocs := frames(docs)
	log.Printf("Got %d frames", len(frameDocs))

	images, err := c.Images(*key, 2, figma.ImageFormatPNG, ids(frameDocs)...)
	if err != nil {
		log.Println(err)
	}

	log.Printf("Downloading %d images\n", len(images))
	//os.MkdirAll(*key, os.ModePerm)
	for i, img := range images {
		rc, err := download(img)
		if err != nil {
			log.Fatal(err)
		}

		data, err := io.ReadAll(rc)
		if err != nil {
			log.Fatal(err)
		}
		path := filepath.Join("./output/bulletin", fmt.Sprintf("%d.png", i+1))
		if err := os.WriteFile(path, data, 0666); err != nil {
			log.Fatal(err)
		}
	}
}

func ids(docs []figma.Node) []string {
	var res []string
	for i := range docs {
		res = append(res, docs[i].ID)
	}
	return res
}

func frames(docs []figma.Node) []figma.Node {
	var res []figma.Node
	for i := range docs {
		if docs[i].Type == figma.NodeTypeFrame {
			res = append(res, docs[i])
		}
	}
	return res
}

func download(i figma.Image) (io.ReadCloser, error) {
	resp, err := http.Get(i.URL)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// orgJson은 그룹화된 JSON 결과를 반환
func orgJson(argResult []map[string]interface{}) map[string][]map[string]string {
	grouped := make(map[string][]map[string]string)

	for _, contentResult := range argResult {
		if name, ok := contentResult["name"].(string); ok {
			switch {
			case name == "content_1", name == "content_2", name == "content_3":
				processContent(name, contentResult)
			case strings.HasPrefix(name, "sub_"):
				grouped[name] = extractChildren(contentResult)
			}
		}
	}

	return grouped
}

// 특정 content 이름에 따라 그룹화된 결과를 파일로 저장
func processContent(name string, contentResult map[string]interface{}) {
	if children, ok := contentResult["children"].([]interface{}); ok {
		result := orgJson(convertToMapSlice(children))
		final := createSortedResult(result)

		sample, _ := json.MarshalIndent(final, "", "  ")
		_ = os.WriteFile(filepath.Join("config", name+".json"), sample, 0644)
	}
}

// 그룹화된 결과를 정렬하여 리스트 형식으로 반환
func createSortedResult(groupedResults map[string][]map[string]string) []map[string]interface{} {
	sortedKeys := sortKeys(groupedResults)
	final := make([]map[string]interface{}, 0, len(sortedKeys))
	for _, name := range sortedKeys {
		final = append(final, map[string]interface{}{
			"name":     name,
			"children": groupedResults[name],
		})
	}
	return final
}

// sub_ 항목에서 자식 요소 추출
func extractChildren(contentResult map[string]interface{}) []map[string]string {
	var children []map[string]string

	if childItems, ok := contentResult["children"].([]interface{}); ok {
		for _, child := range childItems {
			if childMap, ok := child.(map[string]interface{}); ok {
				if cName, cOk := childMap["name"].(string); cOk {
					if characters, ok := childMap["characters"].(string); ok {
						children = append(children, map[string]string{
							"type":       cName,
							"characters": characters,
						})
					}
				}
			}
		}
	}

	return children
}

// sub_ 이름의 숫자 순서로 정렬
func sortKeys(groupedResults map[string][]map[string]string) []string {
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
