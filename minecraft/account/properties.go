package account

import (
	"errors"
	"github.com/brawaru/marct/launcher/accounts"
	"github.com/mitchellh/mapstructure"
)

// Properties represents Minecraft account properties.
type Properties struct {
	ID       string `mapstructure:"minecraft:id"`       // Minecraft account ID.
	Username string `mapstructure:"minecraft:username"` // Minecraft account username.
}

func ReadProperties(account *accounts.Account) (properties Properties, err error) {
	err = mapstructure.Decode(account.Properties, &properties)
	return
}

var ErrNilAccount = errors.New("account is nil")

func WriteProperties(account *accounts.Account, properties Properties) error {
	if account == nil {
		return ErrNilAccount
	}
	return mapstructure.Decode(properties, &account.Properties)
}
