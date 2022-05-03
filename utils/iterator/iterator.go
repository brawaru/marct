package iterator

type ReadOnlyIterator[K comparable, V any] interface {
	// HasNext returns true if the iterator has a next element.
	HasNext() bool

	// HasPrevious returns true if the iterator has previous elements.
	HasPrevious() bool

	// Next moves the iterator to the next element and returns it.
	Next() (K, V)

	// Previous moves the iterator to the previous element and returns it.
	Previous() (K, V)
}

type Iterator[K comparable, V any] interface {
	ReadOnlyIterator[K, V]

	// Del removes the current element from the associated map.
	Del()

	// Replace replaces the current element with the given value.
	Set(value V)
}
