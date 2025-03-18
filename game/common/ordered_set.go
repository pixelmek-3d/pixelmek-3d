package common

// OrderedSet is a set that preserves insertion order.
type OrderedSet[T comparable] struct {
	items map[T]bool
	order []T
}

func NewOrderedSet[T comparable]() *OrderedSet[T] {
	return &OrderedSet[T]{
		items: make(map[T]bool),
		order: []T{},
	}
}

func (s *OrderedSet[T]) Add(item T) {
	if _, exists := s.items[item]; !exists {
		s.items[item] = true
		s.order = append(s.order, item)
	}
}

func (s *OrderedSet[T]) Remove(item T) {
	if _, exists := s.items[item]; exists {
		delete(s.items, item)
		for i, v := range s.order {
			if v == item {
				s.order = append(s.order[:i], s.order[i+1:]...)
				break
			}
		}
	}
}

func (s *OrderedSet[T]) Contains(item T) bool {
	_, exists := s.items[item]
	return exists
}

func (s *OrderedSet[T]) Size() int {
	return len(s.order)
}

func (s *OrderedSet[T]) GetOrder() []T {
	return s.order
}

func (s *OrderedSet[T]) Clone() *OrderedSet[T] {
	cloned := NewOrderedSet[T]()
	for item := range s.items {
		cloned.Add(item)
	}
	return cloned
}

func (s *OrderedSet[T]) Iter() <-chan T {
	ch := make(chan T)
	go func() {
		for _, elem := range s.order {
			ch <- elem
		}
		close(ch)
	}()
	return ch
}
