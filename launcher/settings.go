package launcher

import (
	"fmt"
	"os"

	"github.com/99designs/keyring"
	"github.com/BurntSushi/toml"
)

type settings struct {
	Keyring struct {
		Backend *keyring.BackendType `toml:"ban-backends"`
		PassCmd string               `toml:"pass-cmd"`
		PassDir string               `toml:"pass-dir"`
	} `toml:"keyring"`
}

type Settings struct {
	settings        // Actual settings.
	fp       string // Settings file path.
}

func (s *Settings) Save() error {
	f, err := os.Create(s.fp)
	if err != nil {
		return fmt.Errorf("create %s: %w", s.fp, err)
	}

	err = toml.NewEncoder(f).Encode(s.settings)
	if err != nil {
		return fmt.Errorf("encode %s: %w", s.fp, err)
	}

	return err
}

func (s *Settings) Read() (err error) {
	_, err = toml.DecodeFile(s.fp, s.settings)
	if err != nil {
		err = fmt.Errorf("decode %s: %w", s.fp, err)
	}
	return
}

func NewSettings(name string) (*Settings, error) {
	s := &Settings{fp: name}
	if err := s.Read(); err != nil {
		return nil, fmt.Errorf("read %s: %w", s.fp, err)
	}
	return s, nil
}
