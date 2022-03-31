package authflow

import (
	"github.com/brawaru/marct/launcher/accounts"
)

type UpdateAuthorizationStep struct{}

const StepIDUpdateAuthorization = "update_authorization"

func (u *UpdateAuthorizationStep) ID() string {
	return StepIDUpdateAuthorization
}

// updateAuthorization updates the authorization of the account passed through state.
func (u *UpdateAuthorizationStep) updateAuthorization(state *accounts.FlowState[IntermediateState]) error {
	i := state.IntermediateState
	state.Account.Authorization = &accounts.Authorization{
		AccessToken: i.MinecraftToken,
		UserType:    "msa",
		UserName:    i.MinecraftAccountProperties.Username,
		UserUUID:    i.MinecraftAccountProperties.ID,
		DemoUser:    false,
	}
	return nil
}

func (u *UpdateAuthorizationStep) Authorize(state *accounts.FlowState[IntermediateState]) error {
	return u.updateAuthorization(state)
}

func (u *UpdateAuthorizationStep) Refresh(state *accounts.FlowState[IntermediateState]) error {
	return u.updateAuthorization(state)
}
