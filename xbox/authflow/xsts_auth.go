package authflow

import (
	"errors"
	"github.com/brawaru/marct/launcher/accounts"
	"github.com/brawaru/marct/xbox"
)

type XSTSAuthStep struct{}

const StepIDXSTSAuth = "xsts_auth"

func (x *XSTSAuthStep) ID() string {
	return StepIDXSTSAuth
}

func (x *XSTSAuthStep) auth(state *accounts.FlowState[IntermediateState]) error {
	authData := state.IntermediateState

	authResp, authErr := xbox.GetXSTSToken(authData.XBLToken)

	if authErr != nil {
		return authErr
	}

	authData.XSTSToken = authResp.Token

	if len(authData.UserHash) == 0 {
		userHash, userHashFound := findUserHash(authResp.DisplayClaims)

		if !userHashFound {
			return errors.New("cannot find userHash")
		}

		authData.UserHash = userHash
	}

	return nil
}

func (x *XSTSAuthStep) Authorize(state *accounts.FlowState[IntermediateState]) error {
	return x.auth(state)
}

func (x *XSTSAuthStep) Refresh(state *accounts.FlowState[IntermediateState]) error {
	is := state.IntermediateState

	if !is.MsftTokenRefreshed {
		return nil
	}

	return x.auth(state)
}
