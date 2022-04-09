package launcher

import (
	"fmt"
	"os"

	"github.com/99designs/keyring"
	"github.com/BurntSushi/toml"
	"github.com/brawaru/marct/utils/slices"
	"github.com/mitchellh/mapstructure"
)

type Versioned struct {
	Version int `mapstructure:"version"`
}

type Settings struct {
	Versioned `mapstructure:",squash"`
	Keyring   struct {
		Backend *keyring.BackendType `mapstructure:"backend"`
		PassCmd string               `mapstructure:"pass-cmd"`
		PassDir string               `mapstructure:"pass-dir"`
	} `mapstructure:"keyring"`
}

type SettingsFile struct {
	Settings        // Actual settings.
	fp       string // Settings file path.
}

func (s *SettingsFile) Save() error {
	f, err := os.Create(s.fp)
	if err != nil {
		return fmt.Errorf("create %s: %w", s.fp, err)
	}

	err = toml.NewEncoder(f).Encode(s.Settings)
	if err != nil {
		return fmt.Errorf("encode %s: %w", s.fp, err)
	}

	return err
}

type settingsV1 struct {
	Versioned `mapstructure:",squash"`
	Keyring   struct {
		BanBackends []keyring.BackendType `mapstructure:"ban-backends"`
		PassCmd     string                `mapstructure:"pass-cmd"`
		PassDir     string                `mapstructure:"pass-dir"`
	} `mapstructure:"keyring"`
}

func runMigrations(m map[string]any) (migrated bool, err error) {
	initalVersion := -1
	for {
		var v Versioned
		err = mapstructure.Decode(m, &v)
		if err != nil {
			return
		}

		if initalVersion == -1 {
			initalVersion = v.Version
		} else {
			migrated = v.Version != initalVersion
		}

		switch v.Version {
		case 0:
			fallthrough // 0 = unset, thus 1
		case 1:
			var f settingsV1

			if err = mapstructure.Decode(m, &f); err != nil {
				err = fmt.Errorf("cannot decode v1 settings file: %w", err)
				return
			}

			r := &Settings{}
			for _, b := range keyring.AvailableBackends() {
				if !slices.Includes(f.Keyring.BanBackends, b) {
					r.Keyring.Backend = &b
					break
				}
			}

			r.Keyring.PassCmd = f.Keyring.PassCmd
			r.Keyring.PassDir = f.Keyring.PassDir
			r.Version = 2

			if err = mapstructure.Decode(r, &m); err != nil {
				err = fmt.Errorf("cannot encode v2 settings file: %w", err)
				return
			}
		case 2:
			return
		default:
			err = fmt.Errorf("unknown settings version %d", v.Version)
			return
		}
	}
}

func (s *SettingsFile) Read() (err error) {
	m := make(map[string]any)
	_, err = toml.DecodeFile(s.fp, &m)
	if err != nil {
		err = fmt.Errorf("decode %s: %w", s.fp, err)
	}
	migrated, err := runMigrations(m)
	if err != nil {
		err = fmt.Errorf("migrate %s: %w", s.fp, err)
	}
	if err := mapstructure.Decode(m, &s.Settings); err != nil {
		err = fmt.Errorf("map %s: %w", s.fp, err)
	}
	if migrated {
		err = s.Save()
		if err != nil {
			err = fmt.Errorf("save (after migration) %s: %w", s.fp, err)
		}
	}
	return
}

func NewSettings(name string) *SettingsFile {
	s := &SettingsFile{fp: name}
	return s
}
