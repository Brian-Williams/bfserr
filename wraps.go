package bfserr

// AsType finds the first error in err's tree that matches the type E, and
// if one is found, returns that error value and true. Otherwise, it
// returns the zero value of E and false.
//
// The tree consists of err itself, followed by the errors obtained by
// repeatedly calling its Unwrap() error or Unwrap() []error method. When
// err wraps multiple errors, AsType examines err followed by a
// breadth-first traversal of its children.
//
// An error err matches the type E if the type assertion err.(E) holds,
// or if the error has a method As(any) bool such that err.As(target)
// returns true when target is a non-nil *E. In the latter case, the As
// method is responsible for setting target.
//
// This is like errors.As but uses a breadth-first search instead of
// depth-first search. This is most useful when searching deep error trees
// where the desired error is likely to be near the top of the tree.
func AsType[E error](err error) (E, bool) {
	if err == nil {
		var zero E
		return zero, false
	}
	var pe *E // lazily initialized

	queue := []error{err}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if e, ok := current.(E); ok {
			return e, true
		}
		if x, ok := current.(interface{ As(any) bool }); ok {
			if pe == nil {
				pe = new(E)
			}
			if x.As(pe) {
				return *pe, true
			}
		}
		switch x := current.(type) {
		case interface{ Unwrap() error }:
			if child := x.Unwrap(); child != nil {
				queue = append(queue, child)
			}
		case interface{ Unwrap() []error }:
			for _, child := range x.Unwrap() {
				if child != nil {
					queue = append(queue, child)
				}
			}
		}
	}
	var zero E
	return zero, false
}
