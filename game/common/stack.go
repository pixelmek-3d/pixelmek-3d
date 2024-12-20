package common

import "container/list"

// FIFOStack is a stack that is first-in, first-out
type FIFOStack[T any] struct {
	list *list.List
}

// FIFOStack returns a stack that is first-in, first-out
func NewFIFOStack[T any]() *FIFOStack[T] {
	return &FIFOStack[T]{list: list.New()}
}

// Push adds an element to the end of the stack
func (s *FIFOStack[T]) Push(value T) {
	s.list.PushBack(value)
}

// Peek returns the first element in the stack
func (s *FIFOStack[T]) Peek() *T {
	if s.list.Len() == 0 {
		return nil
	}
	element := s.list.Front()
	value := element.Value.(T)
	return &value
}

// Pop removes and returns the first element from the stack
func (s *FIFOStack[T]) Pop() *T {
	if s.list.Len() == 0 {
		return nil
	}
	element := s.list.Front()
	s.list.Remove(element)
	value := element.Value.(T)
	return &value
}

// Len returns the number of elements in the stack
func (s *FIFOStack[T]) Len() int {
	return s.list.Len()
}
