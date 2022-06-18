package authflow

import (
	"github.com/brawaru/marct/launcher/accounts"
	minecraftAccount "github.com/brawaru/marct/minecraft/account"
	offlineAccount "github.com/brawaru/marct/offline/account"
)

type UsernameRequestHandler func() (string, error)

type Options struct {
	UsernameRequestHandler UsernameRequestHandler
}

type IntermediateState struct {
	MinecraftAccountProperties *minecraftAccount.Properties // Properties of the Minecraft account.
}

func CreateAuthFlow(options *Options) *accounts.AuthFlow[IntermediateState] {
	flow := accounts.AuthFlow[IntermediateState]{
		AccountType: offlineAccount.AccountType,
	}

	flow.AddStep(&PropertiesReadStep{})
	flow.AddStep(&RequestUsernameStep{UsernameRequestHandler: options.UsernameRequestHandler})
	flow.AddStep(&PropertiesWriteStep{})
	flow.AddStep(&UpdateAuthorizationStep{})

	return &flow
}
