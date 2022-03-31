package utils

import (
	"errors"
	"os"
)

// DoesNotExist reports whether the error is related to absence of file.
func DoesNotExist(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
