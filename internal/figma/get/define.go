package get

import "github.com/torie/figma"

type Children struct {
	Title string `json:"title"`
	Info  string `json:"info"`
}

type Info struct {
	Client         *figma.Client
	Nodes          []figma.Node
	AssembledNodes []figma.Node
	Token          *string
	Key            *string
	ExecPath       string
}
