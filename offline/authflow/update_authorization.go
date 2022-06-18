package authflow

import (
	"github.com/brawaru/marct/launcher/accounts"
)

const StepIDUpdateAuthorization = "update_authorization"

type UpdateAuthorizationStep struct{}

func (s *UpdateAuthorizationStep) ID() string {
	return StepIDUpdateAuthorization
}

func (s *UpdateAuthorizationStep) update(state *accounts.FlowState[IntermediateState]) error {
	state.Account.Authorization = &accounts.Authorization{
		UserUUID:    "00000000-0000-0000-0000-000000000000",
		UserType:    "msa",
		UserName:    state.IntermediateState.MinecraftAccountProperties.Username,
		AccessToken: "",
	}

	return nil
}

func (s *UpdateAuthorizationStep) Authorize(state *accounts.FlowState[IntermediateState]) error {
	return s.update(state)
}

func (s *UpdateAuthorizationStep) Refresh(state *accounts.FlowState[IntermediateState]) error {
	return s.update(state)
}
