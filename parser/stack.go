package parser

type NodeType int

const (
	BLUE NodeType = 0 << iota
	GREEN
)

type Node struct {
	nodeType NodeType
	parent   *Node
	vars     *Node
}

func NewNode(nodeType NodeType) *Node {
  return &Node{nodeType: nodeType, parent: nil, vars: nil}
}
