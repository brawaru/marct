package orderedmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderedMap(t *testing.T) {
	m := New[string, string]()

	if !assert.Equal(t, 0, m.Len(), "empty map should have 0 length") {
		return
	}

	m.Put("a", "1")
	if !assert.Equal(t, 1, m.Len(), "map should have 1 length") {
		return
	}

	if !assert.Equal(t, "1", m.Get("a"), "value of key 'a' should be '1'") {
		return
	}

	m.Put("b", "2")

	if !assert.Equal(t, []string{"a", "b"}, m.Keys(), "map should consist of keys 'a', 'b'") {
		return
	}

	if !assert.True(t, m.HasKey("a"), "map should have key 'a'") {
		return
	}

	if !assert.True(t, m.Del("a"), "Del must return true for existing key") {
		return
	}

	if !assert.False(t, m.HasKey("a"), "map should have key 'a'") {
		return
	}

	if !assert.Equal(t, []string{"2"}, m.Values(), "map should consist of values '2'") {
		return
	}
}
