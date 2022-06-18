package osfile

import (
	"fmt"
	"os"
	"path/filepath"
)

func New(name string) (*os.File, error) {
	dir := filepath.Dir(name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("mkdirall: %w", err)
	}
	return os.Create(name)
}
