package account

import (
	"errors"
	"github.com/brawaru/marct/launcher/accounts"
	"github.com/mitchellh/mapstructure"
)

type Properties struct {
	KID string `mapstructure:"xbox:kid"`
}

var ErrNilAccount = errors.New("account is nil")

func ReadProperties(account *accounts.Account) (properties Properties, err error) {
	if account == nil {
		err = ErrNilAccount
	} else {
		err = mapstructure.Decode(account.Properties, &properties)
	}
	return
}

func WriteProperties(account *accounts.Account, properties Properties) error {
	if account == nil {
		return ErrNilAccount
	}
	return mapstructure.Decode(properties, &account.Properties)
}
