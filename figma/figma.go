package figma

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/torie/figma"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Element struct {
	Name     string     `json:"name"`
	Children []Children `json:"children"`
}

type Children struct {
	ChildType  string `json:"childType"`
	Characters string `json:"characters"`
}

type Info struct {
	Client   *figma.Client
	Nodes    []figma.Node
	Token    *string
	Key      *string
	ExecPath string
}

func New(token *string, key *string, execPath string) (node *Info) {
	if *token == "" || *key == "" {
		flag.Usage()
		os.Exit(-1)
	}
	return &Info{
		Client:   figma.New(*token),
		Token:    token,
		Key:      key,
		ExecPath: execPath,
	}
}

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

func extractLeadingNumber(name string) int {
	parts := strings.SplitN(name, "_", 2)
	if len(parts) > 0 {
		if num, err := strconv.Atoi(parts[0]); err == nil {
			return num
		}
	}
	return 0
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
