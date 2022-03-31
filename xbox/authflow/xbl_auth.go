package authflow

import (
	"github.com/brawaru/marct/launcher/accounts"
	"github.com/brawaru/marct/xbox"
)

type XBLAuthStep struct{}

const StepIDXBLAuth = "xbl_auth"

func (x *XBLAuthStep) ID() string {
	return StepIDXBLAuth
}

func (x *XBLAuthStep) auth(state *accounts.FlowState[IntermediateState]) error {
	authData := state.IntermediateState

	xblAuthResp, authErr := xbox.AuthXBLUser(authData.MsftAccessToken)
	if authErr != nil {
		return authErr
	}

	authData.XBLToken = xblAuthResp.Token

	userHash, userHashFound := findUserHash(xblAuthResp.DisplayClaims)

	if userHashFound {
		authData.UserHash = userHash
	}

	return nil
}

func (x *XBLAuthStep) Authorize(state *accounts.FlowState[IntermediateState]) error {
	return x.auth(state)
}

func (x *XBLAuthStep) Refresh(state *accounts.FlowState[IntermediateState]) error {
	is := state.IntermediateState

	if !is.MsftTokenRefreshed {
		return nil
	}

	return x.auth(state)
}
