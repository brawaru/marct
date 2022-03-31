package launcher

import (
	"errors"
	"fmt"
)

// IllegalDimensionsNumberError is an error that is reported when the dimensions number does not meet or exceeds allowed
// range (that is 2 dimensions - width and height).
type IllegalDimensionsNumberError struct {
	Count int // The number of dimensions found, or -1 if none
}

func (e *IllegalDimensionsNumberError) Error() string {
	if e.Count == -1 {
		return "no dimensions found"
	} else {
		return "too many dimensions"
	}
}

func (e *IllegalDimensionsNumberError) Is(target error) bool {
	t, ok := target.(*IllegalDimensionsNumberError)
	return ok && (t.Count == 0 || e.Count == t.Count)
}

// IllegalDimensionValueError is an error that is reported when value for specific dimension is incorrect.
type IllegalDimensionValueError struct {
	Value     string // Value that cannot be processed
	Dimension string // Either "width" or "height"
	Err       error  // Parsing error, if any
}

func (e *IllegalDimensionValueError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("\"%s\" is illegal value for dimension %s: %s", e.Value, e.Dimension, e.Err)
	} else {
		return fmt.Sprintf("\"%s\" is illegal value for dimension %s", e.Value, e.Dimension)
	}
}

func (e *IllegalDimensionValueError) Unwrap() error {
	return e.Err
}

func (e *IllegalDimensionValueError) Is(target error) bool {
	t, ok := target.(*IllegalDimensionValueError)

	return ok &&
		(t.Value == "" || e.Value == t.Value) &&
		(t.Dimension == "" || e.Dimension == t.Dimension) &&
		(t.Err == nil || errors.Is(e.Err, t.Err)) // FIXME: weak check (should only check top error)
}

// ResolutionParseError is an error that is reported when resolution cannot be parsed from string
type ResolutionParseError struct {
	Input string
	Err   error
}

func (e *ResolutionParseError) Error() string {
	return fmt.Sprintf("cannot parse resolution: %s", e.Err.Error())
}

func (e *ResolutionParseError) Is(target error) bool {
	t, ok := target.(*ResolutionParseError)

	return ok &&
		(t.Err == nil || errors.Is(e.Err, t.Err)) && // FIXME: weak check (should only check top error)
		(t.Input == "" || e.Input == t.Input)
}

func (e *ResolutionParseError) Unwrap() error {
	return e.Err
}
