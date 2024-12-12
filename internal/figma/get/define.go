package get

import "github.com/torie/figma"

type Children struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Info    string `json:"info"`
	Obj     string `json:"obj"`
}

type Info struct {
	FrameName      string
	Client         *figma.Client
	Nodes          []figma.Node
	AssembledNodes []figma.Node
	Token          *string
	Key            *string
	ExecPath       string
	PathInfo       map[string]string
}
