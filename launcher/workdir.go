package launcher

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/brawaru/marct/utils"
)

type Instance struct {
	// Path to the instance directory
	Path string

	closed bool
}

const settingsFile = "marct_settings.toml"

func (w *Instance) OpenSettings() (*SettingsFile, error) {
	s := NewSettings(filepath.Join(w.Path, filepath.FromSlash(settingsFile)))
	if err := s.Read(); err != nil {
		if !utils.DoesNotExist(err) {
			return nil, fmt.Errorf("failed to read settings: %w", err)
		}
	}
	return s, nil
}

func (w *Instance) Close() error {
	if w.closed {
		return errors.New("already closed")
	}

	return nil
}

func OpenInstance(name string) (*Instance, error) {
	currentDir, err := os.Getwd()

	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	return &Instance{
		Path: filepath.Join(currentDir, name),
	}, nil
}
