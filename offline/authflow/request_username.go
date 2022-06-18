package authflow

import (
	"fmt"

	"github.com/brawaru/marct/launcher/accounts"
	"github.com/brawaru/marct/utils"
)

const StepIDRequestUsername = "request_username"

type RequestUsernameStep struct {
	UsernameRequestHandler UsernameRequestHandler
}

func (s *RequestUsernameStep) ID() string {
	return StepIDRequestUsername
}

func (s *RequestUsernameStep) Authorize(state *accounts.FlowState[IntermediateState]) error {
	username, err := s.UsernameRequestHandler()
	if err != nil {
		return fmt.Errorf("request username: %w", err)
	}
	state.IntermediateState.MinecraftAccountProperties.ID = utils.NewUUID()
	state.IntermediateState.MinecraftAccountProperties.Username = username
	return nil
}

func (s *RequestUsernameStep) Refresh(state *accounts.FlowState[IntermediateState]) error {
	if state.IntermediateState.MinecraftAccountProperties.Username == "" {
		return s.Authorize(state)
	}

	return nil
}
