package get

// Node는 Figma 파일에서 필요한 최소 필드만 정의합니다.
// torie/figma 라이브러리의 Node는 strokeWeight를 int로 선언해
// Figma API가 1.0과 같은 float을 반환할 때 파싱 실패가 발생합니다.
type Node struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Children []Node `json:"children"`
}

type Info struct {
	FrameName      string
	Token          *string
	Key            *string
	ExecPath       string
	Nodes          []Node
	AssembledNodes []Node
	PathInfo       map[string]string
}
