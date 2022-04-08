package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraverseCauses(t *testing.T) {
	root := errors.New("root error")

	a := NewWrappedError("wrapped error a", root)

	b := NewWrappedError("wrapped error b", a)

	ret := TraverseCauses(b, errors.Unwrap)

	assert.Equal(t, ret, root, "must return root error")
}
