package util

import "fmt"
import "strconv"

type ScopeTree struct {
	root  *GreenNode
	stack *Stack
}

type GreenNode struct {
	name     string
	sym      *Symbol
	parent   *GreenNode
	vars     []*BlueNode
	children []*GreenNode
	params   int
}

type BlueNode struct {
	name string
	sym  *Symbol
	size int
}

func NewScopeTree() *ScopeTree {
	newStack := NewStack()
	return &ScopeTree{stack: newStack}
}

func NewGreenNode(name string, sym *Symbol) *GreenNode {
	return &GreenNode{name: name, sym: sym, vars: make([]*BlueNode, 0), children: make([]*GreenNode, 0)}
}

func NewBlueNode(name string, sym *Symbol, size int) *BlueNode {
	return &BlueNode{name, sym, size}
}

func (scope *ScopeTree) GetTop() *GreenNode {
	return scope.stack.Peek()
}

func (scope *ScopeTree) Pop() {
	scope.stack.Pop()
}

func (scope *ScopeTree) CreateRoot(name string, sym *Symbol) {
	newGreenNode := NewGreenNode(name, sym)
	scope.root = newGreenNode
	scope.stack.Push(newGreenNode)
}

func (scope *ScopeTree) GetRoot() *GreenNode {
	return scope.root
}

func (node *GreenNode) GetMemoryOffset(list *MemoryOffsetList) {
	list.AddOffset(node.name, "FFFFFFFF")
	runningTotal := 0

	for _, blueNode := range node.vars {
		if blueNode != nil {
			nodeName := blueNode.GetSymbol().name
			nodeType := blueNode.GetSymbol().GetType()
			if nodeType == nodeType&(PGPARM|PPINT|PPAINT|PPREAL|PPAREAL) {
				list.AddOffset(nodeName, "FFFFFFFF")
			} else {
				list.AddOffset(nodeName, strconv.Itoa(runningTotal))
				runningTotal += blueNode.size
			}
		}
	}

	for _, greenNode := range node.children {
		greenNode.GetMemoryOffset(list)
	}
}

func (node *GreenNode) GetName() string {
	return node.name
}

func (node *GreenNode) AddChild(newNode *GreenNode) {
	node.children = append(node.children, newNode)
}

func (node *GreenNode) RemoveChild(removeNode *GreenNode) {
	var match int = -1
	if node != nil {
		for idx, childNode := range node.children {
			if childNode == removeNode {
				match = idx
			}
		}

		if match != -1 {
			node.children = append(node.children[:match], node.children[match+1:]...)
		}
	}
}

func (scope *ScopeTree) AddGreenNode(name string, sym *Symbol) {
	newGreenNode := NewGreenNode(name, sym)
	currentNode := scope.stack.Peek()
	scope.stack.Push(newGreenNode)
	currentNode.AddChild(newGreenNode)
	newGreenNode.parent = currentNode
}

func (node *GreenNode) FindGreenNode(name string) *GreenNode {
	for _, greenNode := range node.children {
		if greenNode.name == name {
			return greenNode
		}
	}

	if node.parent != nil {
		if node.parent.name == name {
			return node.parent
		} else {
			return node.parent.FindGreenNode(name)
		}
	}
	return nil
}

func (node *GreenNode) AddBlueNode(name string, sym *Symbol, size int) error {
	newBlueNode := NewBlueNode(name, sym, size)

	blueNode, _ := node.FindBlueNode(name)
	if blueNode != nil {
		blueNodeSym := blueNode.sym
		nodeType := blueNodeSym.GetType()

		if nodeType == sym.GetType() {
			return fmt.Errorf("Variable already exists with same type")
		}
	}

	node.vars = append(node.vars, newBlueNode)

	return nil
}

func (node *GreenNode) FindBlueNode(name string) (*BlueNode, error) {
	for _, blueNode := range node.vars {
		if blueNode != nil {
			if blueNode.name == name {
				return blueNode, nil
			}
		}
	}

	if node.parent != nil {
		blueNode, _ := node.parent.FindBlueNode(name)
		if blueNode != nil {
			return blueNode, nil
		}
	}

	return nil, fmt.Errorf("Variable not found")
}

func (node *GreenNode) GetVars() []*BlueNode {
	return node.vars
}

func (node *GreenNode) IncParam() {
	node.params++
}

func (node *GreenNode) GetNumParams() int {
	return node.params
}

func (node *GreenNode) Print() {
	fmt.Println(node.name)
	for _, blueNode := range node.vars {
		if blueNode != nil {
			fmt.Print("\t")
			fmt.Print(blueNode)
			fmt.Print("\n")
		}
	}

	for _, greenNode := range node.children {
		if greenNode != nil {
			greenNode.Print()
		}
	}
	fmt.Println()
}

func (node *BlueNode) GetSymbol() *Symbol {
	return node.sym
}
