package authflow

import (
	"fmt"

	"github.com/brawaru/marct/launcher/accounts"
	minecraftAccount "github.com/brawaru/marct/minecraft/account"
)

const StepIDPropertiesRead = "read_properties"

type PropertiesReadStep struct{}

// implement ID, authorize, refresh

func (s *PropertiesReadStep) ID() string {
	return StepIDPropertiesRead
}

func (s *PropertiesReadStep) Authorize(state *accounts.FlowState[IntermediateState]) error {
	state.IntermediateState.MinecraftAccountProperties = &minecraftAccount.Properties{}
	return nil
}

func (s *PropertiesReadStep) Refresh(state *accounts.FlowState[IntermediateState]) error {
	props, err := minecraftAccount.ReadProperties(state.Account)
	if err != nil {
		return fmt.Errorf("read properties: %w", err)
	}

	state.IntermediateState.MinecraftAccountProperties = &props

	return nil
}
