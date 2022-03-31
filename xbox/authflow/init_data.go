package authflow

import (
	"github.com/brawaru/marct/launcher/accounts"
	xboxAccount "github.com/brawaru/marct/xbox/account"
)

type ReadDataStep struct{}

const StepIDReadData = "read_data"

func (r *ReadDataStep) ID() string {
	return StepIDReadData
}

func (r *ReadDataStep) Authorize(state *accounts.FlowState[IntermediateState]) error {
	// when authorize is used we supposedly do not have any data yet
	i := state.IntermediateState
	i.AuthData = &xboxAccount.AuthData{}

	return nil
}

func (r *ReadDataStep) Refresh(state *accounts.FlowState[IntermediateState]) error {
	i := state.IntermediateState

	authData, readErr := xboxAccount.ReadAuthData(*state.Account, i.key)

	if readErr != nil {
		return readErr
	}

	i.AuthData = &authData

	return nil
}
