package get

import (
	"easyPreparation_1.0/internal/handlers"
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
		handlers.BroadcastProgress("Get figma image elements", 1, fmt.Sprintf("Got %d documents", len(i.Nodes)))
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
		handlers.BroadcastProgress(err.Error(), -1, err.Error())
	}
	handlers.BroadcastProgress("Downloading figma images", 1, fmt.Sprintf("Downloading %s images\n", name))

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
	handlers.BroadcastProgress("Got frame", 1, fmt.Sprintf("Got %d frames", len(res)))
	return res
}
