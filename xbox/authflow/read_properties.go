package authflow

import (
	"github.com/brawaru/marct/launcher/accounts"
	minecraftAccount "github.com/brawaru/marct/minecraft/account"
	xboxAccount "github.com/brawaru/marct/xbox/account"
)

type ReadAccountProperties struct{}

const StepIDReadAccountProperties = "read_account_properties"

func (r *ReadAccountProperties) ID() string {
	return StepIDReadAccountProperties
}

func (r *ReadAccountProperties) Authorize(state *accounts.FlowState[IntermediateState]) error {
	i := state.IntermediateState
	i.XboxAccountProperties = &xboxAccount.Properties{}
	i.MinecraftAccountProperties = &minecraftAccount.Properties{}
	return nil
}

func (r *ReadAccountProperties) Refresh(state *accounts.FlowState[IntermediateState]) error {
	i := state.IntermediateState

	xp, err := xboxAccount.ReadProperties(state.Account)
	if err != nil {
		return err
	}
	i.XboxAccountProperties = &xp

	mp, err := minecraftAccount.ReadProperties(state.Account)
	if err != nil {
		return err
	}
	i.MinecraftAccountProperties = &mp

	return nil
}
