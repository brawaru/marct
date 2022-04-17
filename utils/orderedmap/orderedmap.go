package orderedmap

import (
	"github.com/brawaru/marct/utils/slices"
)

type Map[K comparable, V any] interface {
	// Returns the number of elements in the map.
	Len() int

	// HasKey returns true if the map contains the given key.
	HasKey(key K) bool

	// Get returns the value associated with the given key.
	Get(key K) V

	// Put associates the given value with the given key.
	Put(key K, value V)

	// Del removes the given key and its associated value, it returns true if the key was found and deleted.
	Del(key K) bool

	// Keys returns the keys of the map.
	Keys() []K

	// Values returns the values of the map.
	Values() []V

	// 	Iter() iterator.Iterator[K, V]
}

type orderedMap[K comparable, V any] struct {
	orderedKeys []K
	values      map[K]V
	m           int // m stores a number that is changed when the map is modified to avoid concurrent map access.
}

func (m *orderedMap[K, V]) Len() int {
	return len(m.orderedKeys)
}

func (m *orderedMap[K, V]) HasKey(key K) bool {
	return slices.Includes(m.orderedKeys, key)
}

func (m *orderedMap[K, V]) Get(key K) V {
	return m.values[key]
}

func (m *orderedMap[K, V]) Put(key K, value V) {
	m.orderedKeys = append(m.orderedKeys, key)
	m.values[key] = value
	m.m++
}

func (m *orderedMap[K, V]) Del(key K) bool {
	if !m.HasKey(key) {
		return false
	}

	m.orderedKeys, _ = slices.Exclude(m.orderedKeys, key)
	delete(m.values, key)
	m.m++

	return true
}

func (m *orderedMap[K, V]) Keys() []K {
	return m.orderedKeys
}

func (m *orderedMap[K, V]) Values() []V {
	values := make([]V, len(m.orderedKeys))
	for i, key := range m.orderedKeys {
		values[i] = m.values[key]
	}
	return values
}

// type orderedMapIter[K comparable, V any] struct {
// 	m  *orderedMap[K, V] // Map that is being iterated over.
// 	c  int               // Current index.
// 	lm int               // Last map modification.
// }

// func (i *orderedMapIter[K, V]) HasNext() bool {
// 	return i.c < (i.m.Len() - 1)
// }

// func (i *orderedMapIter[K, V]) HasPrevious() bool {
// 	return i.c <= 0
// }

// func (i *orderedMapIter[K, V]) Next() (K, V) {
// 	if i.c >= i.m.Len() {
// 		panic("No next element.")
// 	}

// 	if i.lm != i.m.m {
// 		panic("Map modified during iteration.")
// 	}

// 	key := i.m.orderedKeys[i.c]
// 	value := i.m.values[key]
// 	i.c++

// 	return key, value
// }

// func (i *orderedMapIter[K, V]) Previous() (K, V) {
// 	if i.c <= 0 {
// 		panic("No previous element.")
// 	}

// 	if i.lm != i.m.m {
// 		panic("Map modified during iteration.")
// 	}

// 	i.c--
// 	key := i.m.orderedKeys[i.c]
// 	value := i.m.values[key]

// 	return key, value
// }

// func (i *orderedMapIter[K, V]) Del() {
// 	if i.c == -1 {
// 		panic("No current element.")
// 	}

// 	if i.lm != i.m.m {
// 		panic("Map modified during iteration.")
// 	}

// 	i.m.Del(i.m.orderedKeys[i.c])

// }

// func (m *orderedMap[K, V]) Iter() iterator.Iterator[K, V] {
// 	return &orderedMapIter[K, V]{m, -1, m.m}
// }

func New[K comparable, V any]() Map[K, V] {
	return &orderedMap[K, V]{
		orderedKeys: make([]K, 0),
		values:      make(map[K]V),
	}
}
