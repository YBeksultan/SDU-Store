package main

import "fmt"

// Yerzhanov Beksultan 200103513

type Node struct {
	Value int
	Next  *Node
}

type Stack struct {
	top *Node
}

func (stack *Stack) Pop() int {
	if stack.top == nil {
		return -1
	}
	value := stack.top.Value
	stack.top = stack.top.Next
	return value
}

func (stack *Stack) Push(value int) {
	node := Node{Value: value, Next: stack.top}
	stack.top = &node
}

func (stack *Stack) Peek() int {
	if stack.top == nil {
		return -1
	}
	value := stack.top.Value
	return value
}

func (stack *Stack) Clear() {
	for stack.top != nil {
		temp := stack.top
		stack.top = stack.top.Next
		temp.Next = nil
	}
}

func (stack *Stack) Contains(value int) bool {
	check := stack.top
	for check != nil {
		if check.Value == value {
			return true
		}
		check = check.Next
	}
	return false
}

func (stack *Stack) Increment() {
	current := stack.top
	for current != nil {
		current.Value++
		current = current.Next
	}
}

func (stack *Stack) Print() {
	if stack.top == nil {
		fmt.Print("Stack is empty")
		return
	}

	check := stack.top
	for check != nil {
		fmt.Print(check.Value)
		fmt.Print(" ")
		check = check.Next
	}
	fmt.Println()
}

func (stack *Stack) PrintReverse() {
	temp := Stack{}
	check := stack.top
	for check != nil {
		temp.Push(check.Value)
		check = check.Next
	}
	temp.Print()
}

func main() {
	node := Node{Value: 1, Next: &Node{Value: 2, Next: &Node{Value: 3, Next: &Node{Value: 4, Next: nil}}}}

	stack := Stack{top: &node}
	fmt.Printf("%d is popped", stack.Pop())
	fmt.Println()
	stack.Print()
	stack.Push(1)
	stack.Print()
	stack.PrintReverse()
	stack.Increment()
	stack.Print()
	fmt.Println(stack.Contains(6))
	stack.Clear()
	stack.Print()
}
