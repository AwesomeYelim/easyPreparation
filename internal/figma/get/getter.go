package get

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/torie/figma"
	"io"
	"log"
	"os"
	"path/filepath"
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

func (i *Info) GetImage(createdPath string, id string, name string) {
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

	path := filepath.Join(createdPath, fmt.Sprintf("%s.png", name))
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
