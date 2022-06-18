package authflow

import (
	"fmt"

	"github.com/brawaru/marct/launcher/accounts"
	minecraftAccount "github.com/brawaru/marct/minecraft/account"
)

const StepIDPropertiesWrite = "write_properties"

type PropertiesWriteStep struct{}

func (s *PropertiesWriteStep) ID() string {
	return StepIDPropertiesWrite
}

func (s *PropertiesWriteStep) Authorize(state *accounts.FlowState[IntermediateState]) error {
	err := minecraftAccount.WriteProperties(state.Account, *state.IntermediateState.MinecraftAccountProperties)

	if err != nil {
		return fmt.Errorf("write properties: %w", err)
	}

	return nil
}

func (s *PropertiesWriteStep) Refresh(state *accounts.FlowState[IntermediateState]) error {
	return nil // We assume that username never updates.
}
