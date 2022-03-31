package authflow

import (
	"github.com/brawaru/marct/launcher/accounts"
	minecraftAPI "github.com/brawaru/marct/minecraft/api"
	"time"
)

type MCAuthStep struct{}

const StepIDMCAuth = "mc_auth"

func (m *MCAuthStep) ID() string {
	return StepIDMCAuth
}

func (m *MCAuthStep) auth(state *accounts.FlowState[IntermediateState]) error {
	is := state.IntermediateState

	authResp, authErr := minecraftAPI.LoginWithXbox(is.UserHash, is.XSTSToken)
	if authErr != nil {
		return authErr
	}

	is.MinecraftToken = authResp.AccessToken
	is.MinecraftTokenExpiresAt = time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second).Add(-1 * time.Second)

	return nil
}

func (m *MCAuthStep) Authorize(state *accounts.FlowState[IntermediateState]) error {
	return m.auth(state)
}

func (m *MCAuthStep) Refresh(state *accounts.FlowState[IntermediateState]) error {
	is := state.IntermediateState

	if !is.MsftTokenRefreshed && is.MinecraftTokenExpiresAt.After(time.Now()) {
		return nil
	}

	return m.auth(state)
}
