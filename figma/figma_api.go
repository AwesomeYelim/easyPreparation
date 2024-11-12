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

	result := orgJson(mainContent)

	sample, _ = json.MarshalIndent(result, "", "")
	_ = os.WriteFile("./schema.json", sample, 0644)

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

func orgJson(argResult []map[string]interface{}) (result []map[string]interface{}) {
	for _, contentResult := range argResult {
		if name, ok := contentResult["name"].(string); ok && strings.HasPrefix(name, "sub_") {
			result = append(result, contentResult)
		}

		if children, ok := contentResult["children"].([]interface{}); ok {
			childResults := orgJson(convertToMapSlice(children))
			result = append(result, childResults...)
		}
	}
	return result
}

func convertToMapSlice(data []interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}
	return result
}
