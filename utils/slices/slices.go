package slices

// Equal tells whether a and b contain the same elements.
// A nil argument is equivalent to an empty slice.
//
// Via https://yourbasic.org/golang/compare-slices/.
func Equal[I comparable](a, b []I) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

func Includes[I comparable](slice []I, item I) bool {
	if slice == nil {
		return false
	}

	for _, i := range slice {
		if i == item {
			return true
		}
	}

	return false
}

func ExcludeFrom[I comparable](slice *[]I, item I) bool {
	s, ok := Exclude(*slice, item)
	if ok {
		*slice = s
	}
	return ok
}

func Exclude[I comparable](slice []I, item I) ([]I, bool) {
	if slice == nil {
		return nil, false
	}

	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...), true
		}
	}

	return slice, false
}

type Predicate[I any] func(item *I, index int, slice []I) bool

// NotFound is an index that is reported when no value is found.
const NotFound = -1

func FindIndex[I any](slice []I, predicate Predicate[I]) int {
	if slice == nil {
		return NotFound
	}

	for i, item := range slice {
		if predicate(&item, i, slice) {
			return i
		}
	}

	return NotFound
}

func Find[I any](slice []I, predicate Predicate[I]) *I {
	index := FindIndex(slice, predicate)

	if index == NotFound {
		return nil
	}

	return &slice[index]
}

func Push[I any](slice *[]I, values ...I) {
	for _, v := range values {
		*slice = append(*slice, v)
	}
}

func Some[I any](value []I, predicate Predicate[I]) bool {
	return FindIndex(value, predicate) != NotFound
}

func Copy[I any](v []I) []I {
	if v == nil {
		return nil
	}
	c := make([]I, len(v))
	copy(c, v)
	return c
}
