package get

import "github.com/torie/figma"

type Info struct {
	FrameName      string
	Client         *figma.Client
	Nodes          []figma.Node
	AssembledNodes []figma.Node
	Token          *string `json:"token"`
	Key            *string `json:"key"`
	ExecPath       string
	PathInfo       map[string]string
}
