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
	Client *figma.Client
	Nodes  []figma.Node
	Token  *string
	Key    *string
}

func New(token *string, key *string) (node *Info) {
	if *token == "" || *key == "" {
		flag.Usage()
		os.Exit(-1)
	}
	return &Info{
		Client: figma.New(*token),
		Token:  token,
		Key:    key,
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

func (i *Info) GetFigmaImage() {
	i.Nodes = i.GetFrames()
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
		path := filepath.Join("./output/bulletin", fmt.Sprintf("%d.png", index+1))
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

func (i *Info) GetFrames() []figma.Node {
	var res []figma.Node
	for index := range i.Nodes {
		if i.Nodes[index].Type == figma.NodeTypeFrame {
			res = append(res, i.Nodes[index])
		}
	}
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
	_ = orgJson(mainContent)
}
