package crdt

type Set[T comparable] struct {
	inner map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{inner: map[T]struct{}{}}
}

func (s *Set[T]) Add(item T) {
	s.inner[item] = struct{}{}
}

func (s *Set[T]) Remove(item T) {
	delete(s.inner, item)
}

func (s *Set[T]) Has(item T) bool {
	_, has := s.inner[item]
	return has
}

func (s *Set[T]) Entries() []T {
	items := make([]T, 0, len(s.inner))
	for entry := range s.inner {
		items = append(items, entry)
	}

	return items
}
