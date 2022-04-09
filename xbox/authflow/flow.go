package authflow

import (
	"github.com/99designs/keyring"
	"github.com/brawaru/marct/launcher/accounts"
	minecraftAccount "github.com/brawaru/marct/minecraft/account"
	minecraftAPI "github.com/brawaru/marct/minecraft/api"
	xboxAccount "github.com/brawaru/marct/xbox/account"
)

// Options represents options for Xbox authentication flow.
type Options struct {
	DeviceAuthHandler DeviceAuthHandler // DeviceAuthHandler used for when device needs to be authenticated by user.
	Keyring           keyring.Keyring   // Keyring where account data is stored.
}

type IntermediateState struct {
	*xboxAccount.AuthData
	key                        []byte                       // Key used to encrypt or decrypt account data.
	XboxAccountProperties      *xboxAccount.Properties      // Properties of the Xbox account.
	MinecraftAccountProperties *minecraftAccount.Properties // Properties of the Minecraft account.
	MsftTokenRefreshed         bool                         // Whether Microsoft token has been refreshed, requiring steps relying on it to also refresh their tokens.
}

func CreateAuthFlow(options *Options) *accounts.AuthFlow[IntermediateState] {
	flow := accounts.AuthFlow[IntermediateState]{
		AccountType: "xbox",
	}

	flow.AddStep(&ReadAccountProperties{})
	flow.AddStep(&ReadKeyringKeyStep{
		Keyring: options.Keyring,
	})
	flow.AddStep(&ReadDataStep{})
	flow.AddStep(&DeviceAuthStep{Handler: options.DeviceAuthHandler})
	flow.AddStep(&XBLAuthStep{})
	flow.AddStep(&XSTSAuthStep{})
	flow.AddStep(&MCAuthStep{})
	flow.AddStep(&EntitlementsCheckStep{
		RequiredEntitlements: []string{
			minecraftAPI.EntitlementMinecraftGame,
			minecraftAPI.EntitlementMinecraftProduct,
		},
	})
	flow.AddStep(&RefreshMCProfileStep{})
	flow.AddStep(&FlushDataStep{})
	flow.AddStep(&FlushKeyringKeyStep{
		Keyring: options.Keyring,
	})
	flow.AddStep(&FlushAccountPropertiesStep{})
	flow.AddStep(&UpdateAuthorizationStep{})

	return &flow
}
