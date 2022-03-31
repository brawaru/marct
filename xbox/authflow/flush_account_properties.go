package authflow

import (
	"fmt"
	"github.com/brawaru/marct/launcher/accounts"
	minecraftAccount "github.com/brawaru/marct/minecraft/account"
	xboxAccount "github.com/brawaru/marct/xbox/account"
)

type FlushAccountPropertiesStep struct{}

const StepIDFlushAccountProperties = "flush_account_properties"

func (s *FlushAccountPropertiesStep) ID() string {
	return StepIDFlushAccountProperties
}

func (s *FlushAccountPropertiesStep) flush(state *accounts.FlowState[IntermediateState]) error {
	i := state.IntermediateState
	if err := xboxAccount.WriteProperties(state.Account, *i.XboxAccountProperties); err != nil {
		return fmt.Errorf("cannot write xbox account properties: %w", err)
	}

	if err := minecraftAccount.WriteProperties(state.Account, *i.MinecraftAccountProperties); err != nil {
		return fmt.Errorf("cannot write minecraft account properties: %w", err)
	}

	return nil
}

func (s *FlushAccountPropertiesStep) Authorize(state *accounts.FlowState[IntermediateState]) error {
	return s.flush(state)
}

func (s *FlushAccountPropertiesStep) Refresh(state *accounts.FlowState[IntermediateState]) error {
	return s.flush(state)
}
