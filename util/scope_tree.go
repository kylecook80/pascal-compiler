package util

import "fmt"

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
}

type BlueNode struct {
	name string
	sym  *Symbol
}

func NewScopeTree() *ScopeTree {
	newStack := NewStack()
	return &ScopeTree{stack: newStack}
}

func NewGreenNode(name string, sym *Symbol) *GreenNode {
	return &GreenNode{name: name, sym: sym, vars: make([]*BlueNode, 0), children: make([]*GreenNode, 0)}
}

func NewBlueNode(name string, sym *Symbol) *BlueNode {
	return &BlueNode{name, sym}
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

func (scope *ScopeTree) AddGreenNode(name string, sym *Symbol) {
	newGreenNode := NewGreenNode(name, sym)
	currentNode := scope.stack.Peek()
	scope.stack.Push(newGreenNode)
	currentNode.addChild(newGreenNode)
	newGreenNode.parent = currentNode
}

func (node *GreenNode) addChild(newNode *GreenNode) {
	node.children = append(node.children, newNode)
}

func (node *GreenNode) AddBlueNode(name string, sym *Symbol) error {
	newBlueNode := NewBlueNode(name, sym)

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

func (node *BlueNode) GetSymbol() *Symbol {
	return node.sym
}
