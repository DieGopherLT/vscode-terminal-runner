// Package collections provides small, self-documenting generic container types.
package collections

// Set is a collection of unique comparable elements. It replaces the ad-hoc
// map[T]bool idiom: membership is explicit, and a removed element is
// indistinguishable from one that was never present.
type Set[T comparable] struct {
	items map[T]struct{}
}

// NewSet creates an empty Set ready for use.
func NewSet[T comparable]() *Set[T] {
	return &Set[T]{items: make(map[T]struct{})}
}

// Add inserts item into the set. Adding an element already present is a no-op.
func (s *Set[T]) Add(item T) {
	s.items[item] = struct{}{}
}

// Remove deletes item from the set. Removing an absent element is a no-op.
func (s *Set[T]) Remove(item T) {
	delete(s.items, item)
}

// Contains reports whether item is currently a member of the set.
func (s *Set[T]) Contains(item T) bool {
	_, found := s.items[item]
	return found
}

// Toggle flips the membership of item: it is removed when present and added
// when absent, with a single map lookup.
func (s *Set[T]) Toggle(item T) {
	if _, found := s.items[item]; found {
		delete(s.items, item)
		return
	}
	s.items[item] = struct{}{}
}

// Len returns the number of elements currently in the set.
func (s *Set[T]) Len() int {
	return len(s.items)
}

// Clear removes every element, leaving an empty set.
func (s *Set[T]) Clear() {
	s.items = make(map[T]struct{})
}
