package account

import (
	"testing"

	"github.com/brawaru/marct/launcher/accounts"
	"github.com/stretchr/testify/assert"
)

func TestReadWrite(t *testing.T) {
	account := &accounts.Account{
		Type:     "any",
		ID:       "owo",
		AuthData: nil,
		Properties: map[string]string{
			"minecraft:id":       "123456789",
			"minecraft:username": "Brawaru",
			"xbox:id":            "123",
		},
	}

	accountData, readErr := ReadProperties(account)
	if !assert.NoError(t, readErr, "must not error reading out data") {
		return
	}

	assert.Equal(t, accountData.ID, "123456789", "decoded ID must be correct")
	assert.Equal(t, accountData.Username, "Brawaru", "decoded username must be correct")

	accountData.ID = "987654321"

	writeErr := WriteProperties(account, accountData)
	if !assert.NoError(t, writeErr, "must not error writing out data") {
		return
	}

	assert.Equal(t, account.Properties["minecraft:id"], "987654321", "ID must update")
	assert.Equal(t, account.Properties["xbox:id"], "123", "extra props must be untouched")
}
