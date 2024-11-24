package get

import (
	"encoding/json"
	"fmt"
	"github.com/torie/figma"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
)

func (i *Info) GetNodes() {
	f, err := i.Client.File(*i.Key)
	if err != nil {
		log.Println(err)
	}
	i.Nodes = f.Nodes()
	log.Printf("Got %d documents", len(i.Nodes))
}

func (i *Info) GetFigmaImage(path string, frameName string) {
	i.Nodes = i.GetFrames(frameName)
	ids := i.GetIds()
	images, err := i.Client.Images(*i.Key, 2, figma.ImageFormatPNG, ids...)
	if err != nil {
		log.Println(err)
	}
	log.Printf("Downloading %d images\n", len(images))

	for index, img := range images {

		rc, err := download(img)
		if err != nil {
			log.Fatal(err)
		}

		data, err := io.ReadAll(rc)
		if err != nil {
			log.Fatal(err)
		}
		path := filepath.Join(path, fmt.Sprintf("%d.png", index+1))
		if err := os.WriteFile(path, data, 0666); err != nil {
			log.Fatal(err)
		}
	}
}

func (i *Info) GetIds() []string {
	var res []string
	for index := range i.Nodes {
		res = append(res, i.Nodes[index].ID)
	}
	return res
}

func (i *Info) GetFrames(frameName string) []figma.Node {
	var res []figma.Node

	for index := range i.Nodes {
		if i.Nodes[index].Type == figma.NodeTypeFrame {
			if i.Nodes[index].Name == frameName {
				childrenFrames := (&Info{Nodes: i.Nodes[index].Children}).GetFrames("children")
				res = append(res, childrenFrames...)
			} else if frameName == "children" {
				res = append(res, i.Nodes[index])
			}
		}
	}

	// 정렬
	sort.Slice(res, func(a, b int) bool {
		numA := extractLeadingNumber(res[a].Name)
		numB := extractLeadingNumber(res[b].Name)
		return numA < numB
	})

	log.Printf("Got %d frames", len(res))
	return res
}

func (i *Info) GetContents() {
	var mainContent []map[string]interface{}

	sample, _ := json.MarshalIndent(i.Nodes, "", "")
	err := json.Unmarshal(sample, &mainContent)
	if err != nil {
		log.Print("err : ", err)
	}
	_ = orgJson(mainContent, i.ExecPath)
}
