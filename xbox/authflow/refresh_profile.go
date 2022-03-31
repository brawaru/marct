package authflow

import (
	"fmt"
	"github.com/brawaru/marct/launcher/accounts"
	minecraftAPI "github.com/brawaru/marct/minecraft/api"
)

type RefreshMCProfileStep struct{}

const StepIDRefreshMCProfile = "refresh_profile"

func (s *RefreshMCProfileStep) ID() string {
	return StepIDRefreshMCProfile
}

func (s *RefreshMCProfileStep) updateProfile(state *accounts.FlowState[IntermediateState]) error {
	is := state.IntermediateState

	mcAPI := minecraftAPI.NewAuthorizedAPI(is.MinecraftToken)

	profile, profileFetchErr := mcAPI.GetProfile()

	if profileFetchErr != nil {
		return fmt.Errorf("failed to fetch minecraft account: %s", profileFetchErr)
	}

	p := is.MinecraftAccountProperties
	p.ID = profile.ID
	p.Username = profile.Name

	state.Account.ID = profile.ID

	return nil
}

func (s *RefreshMCProfileStep) Authorize(state *accounts.FlowState[IntermediateState]) error {
	return s.updateProfile(state)
}

func (s *RefreshMCProfileStep) Refresh(state *accounts.FlowState[IntermediateState]) error {
	return s.updateProfile(state)
}
