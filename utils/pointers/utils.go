package pointers

// Ref returns a pointer to the passed value.
func Ref[T any](v T) *T {
	return &v
}

// DerefOrDefault attempts to dereference the non-nil pointer.
// If the pointer is nil, then the default value for type T is returned.
func DerefOrDefault[T any](r *T) T {
	if r == nil {
		var v T
		return v
	}

	return *r
}
