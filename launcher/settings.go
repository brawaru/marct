package launcher

import (
	"github.com/99designs/keyring"
	"github.com/BurntSushi/toml"
)

type Settings struct {
	Keyring struct {
		BanBackends []keyring.BackendType `toml:"ban-backends"`
		PassCmd     string                `toml:"pass-cmd"`
		PassDir     string                `toml:"pass-dir"`
	} `toml:"keyring"`
}

func (settings *Settings) LoadSettings(file string) error {
	_, err := toml.DecodeFile(file, settings)
	return err
}
