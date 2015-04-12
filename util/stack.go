package util

type Stack struct {
	list []*GreenNode
}

func NewStack() *Stack {
	return &Stack{make([]*GreenNode, 1)}
}

func (stack *Stack) Push(item *GreenNode) {
	stack.list = append(stack.list, item)
}

func (stack *Stack) Pop() *GreenNode {
	greenNode := stack.list[len(stack.list)-1]
	stack.list = stack.list[0 : len(stack.list)-2]
	return greenNode
}

func (stack *Stack) Peek() *GreenNode {
	greenNode := stack.list[len(stack.list)-1]
	return greenNode
}
