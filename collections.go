package main

type stack[T comparable] []T

func (s *stack[T]) pop() T {
	tmp := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return tmp
}
func (s *stack[T]) push(elem T) {
	*s = append(*s, elem)
}
func (s stack[T]) peek() T {
	return s[len(s)-1]
}

type set[T comparable] map[T]struct{}

func (s set[T]) has(value T) bool {
	_, ok := s[value]
	return ok
}
func (s *set[T]) add(value T) {
	(*s)[value] = struct{}{}
}
func (s *set[T]) remove(value T) {
	delete(*s, value)
}
func (s set[T]) intersection(otherSet set[T]) []T {
	result := make([]T, 0)
	for v := range s {
		if otherSet.has(v) {
			result = append(result, v)
		}
	}
	return result
}
