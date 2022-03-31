package launcher

import (
	"errors"
	"fmt"
	"github.com/brawaru/marct/utils"
	"os"
	"path/filepath"
)

type Instance struct {
	Path string

	Settings Settings
	closed   bool
}

const settingsFile = "marct_settings.toml"

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

	settings := new(Settings)
	if err := settings.LoadSettings(filepath.Join(currentDir, settingsFile)); err != nil {
		if !utils.DoesNotExist(err) {
			return nil, fmt.Errorf("failed to load settings: %w", err)
		}
	}

	return &Instance{
		Path:     filepath.Join(currentDir, name),
		Settings: *settings,
	}, nil
}
