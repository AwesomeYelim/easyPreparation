package get

import (
	"easyPreparation_1.0/internal/sorted"
	"easyPreparation_1.0/pkg"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/torie/figma"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func (i *Info) GetNodes() (err error) {
	f, err := i.Client.File(*i.Key)

	i.Nodes = f.Nodes()
	if len(i.Nodes) > 0 {
		log.Printf("Got %d documents", len(i.Nodes))
	} else {
		return errors.Wrap(err, "Nothing documents")
	}
	return nil
}

func (i *Info) GetFigmaImage(path string, frameName string) {
	i.AssembledNodes = i.GetFrames(frameName)
	for index := range i.AssembledNodes {
		id := i.AssembledNodes[index].ID
		name := i.AssembledNodes[index].Name
		i.GetImage(path, id, name)
	}
}

func (i *Info) GetImage(exePath string, id string, name string) {
	img, err := i.Client.Images(*i.Key, 2, figma.ImageFormatPNG, id)
	if err != nil {
		log.Println(err)
	}
	log.Printf("Downloading %s images\n", name)

	rc, err := download(img[0])
	if err != nil {
		log.Fatal(err)
	}

	data, err := io.ReadAll(rc)
	if err != nil {
		log.Fatal(err)
	}

	switch i.FrameName {
	case "forShowing":
		name = strings.Split(name, "_")[1] // 이름만
	case "forPrint":
		name = strings.Split(name, "_")[0] // 숫자만
	default:
		name = strings.Split(name, "_")[1]
	}

	path := filepath.Join(exePath, fmt.Sprintf("%s.png", name))
	i.PathInfo[name] = path
	if err := os.WriteFile(path, data, 0666); err != nil {
		log.Fatal(err)
	}
}

func (i *Info) GetFrames(frameName string) []figma.Node {
	var res []figma.Node

	for index := range i.Nodes {
		if i.Nodes[index].Type == figma.NodeTypeFrame {
			if i.Nodes[index].Name == frameName {
				i.FrameName = frameName
				childrenFrames := (&Info{Nodes: i.Nodes[index].Children}).GetFrames("children")
				res = append(res, childrenFrames...)
			} else if frameName == "children" {
				res = append(res, i.Nodes[index])
			}
		}
	}

	log.Printf("Got %d frames", len(res))
	return res
}

// ppt 만들 resource
func (i *Info) GetResource(target string) {
	var mainContent []map[string]interface{}

	sample, _ := json.MarshalIndent(i.AssembledNodes, "", "")
	err := json.Unmarshal(sample, &mainContent)

	if err != nil {
		log.Print("err : ", err)
	}
	var newG []Children

	grouped := orgJson(mainContent, i.ExecPath, target)

	var keys []string
	for key, _ := range grouped {
		keys = append(keys, key)
	}
	sorted.ToIntSort(keys, "", "_", 1)

	for ind, key := range keys {
		var temp Children
		temp = grouped[key][0]

		// 하위 정렬인 경우 이전 obj 내용을 갖고 옴
		if strings.Contains(key, ".") && ind < len(keys)-1 {
			temp.Obj = grouped[keys[ind-1]][0].Obj
		}
		newG = append(newG, Children{
			Title:   key,
			Content: temp.Content,
			Info:    temp.Info,
			Obj:     temp.Obj,
		})
	}

	sample, _ = json.MarshalIndent(newG, "", "  ")
	_ = pkg.CheckDirIs(filepath.Join(i.ExecPath, "config"))
	_ = os.WriteFile(filepath.Join(i.ExecPath, "config", target+".json"), sample, 0644)
}
