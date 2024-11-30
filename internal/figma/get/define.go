package get

import "github.com/torie/figma"

type Element struct {
	Name     string     `json:"name"`
	Children []Children `json:"children"`
}

type Children struct {
	ChildType  string `json:"childType"`
	Characters string `json:"characters"`
}

type Info struct {
	Client         *figma.Client
	Nodes          []figma.Node
	AssembledNodes []figma.Node
	Token          *string
	Key            *string
	ExecPath       string
}
