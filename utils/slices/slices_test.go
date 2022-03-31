package slices

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestFind(t *testing.T) {
	seq := []string{"a", "b", "c"}
	if !assert.Equal(t, Find(seq, func(item *string, index int, slice []string) bool {
		return *item == "c"
	}), &seq[2]) {
		return
	}

	if !assert.Nil(t, Find(seq, func(item *string, index int, slice []string) bool {
		return *item == "e"
	})) {
		return
	}
}

func TestFindIndex(t *testing.T) {
	seq := []string{"a", "b", "c"}
	assert.Equal(t, FindIndex(seq, func(item *string, index int, slice []string) bool {
		return *item == "b"
	}), 1)
	assert.Equal(t, FindIndex(seq, func(item *string, index int, slice []string) bool {
		return *item == "e"
	}), NotFound)
}

func TestIncludes(t *testing.T) {
	seq := []string{"a", "b", "c"}
	if !assert.True(t, Includes(seq, "b")) {
		return
	}
	if !assert.False(t, Includes(seq, "e")) {
		return
	}
}

func TestSome(t *testing.T) {
	seq := []string{"a", "b", "C", "d"}
	if !assert.True(t, Some(seq, func(item *string, index int, slice []string) bool {
		return strings.ToUpper(*item) == *item
	})) {
		return
	}

	if !assert.False(t, Some(seq, func(item *string, index int, slice []string) bool {
		return strings.HasPrefix(*item, ":)")
	})) {
		return
	}
}

func TestPush(t *testing.T) {
	seq := []string{"a", "b"}
	Push(&seq, "c", "d")

	if !assert.Contains(t, seq, "c") {
		return
	}
	if !assert.Contains(t, seq, "d") {
		return
	}
}
