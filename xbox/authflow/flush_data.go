package authflow

import (
	"github.com/brawaru/marct/launcher/accounts"
	xboxAccount "github.com/brawaru/marct/xbox/account"
)

type FlushDataStep struct{}

const StepIDFlushData = "flush_data"

func (s *FlushDataStep) ID() string {
	return StepIDFlushData
}

func (s *FlushDataStep) flush(state *accounts.FlowState[IntermediateState]) error {
	is := state.IntermediateState

	if writeErr := xboxAccount.WriteAuthData(state.Account, *is.AuthData, is.key); writeErr != nil {
		return writeErr
	}

	return nil
}

func (s *FlushDataStep) Authorize(state *accounts.FlowState[IntermediateState]) error {
	return s.flush(state)
}

func (s *FlushDataStep) Refresh(state *accounts.FlowState[IntermediateState]) error {
	is := state.IntermediateState

	if is.MsftTokenRefreshed {
		return s.flush(state)
	}

	return nil
}
