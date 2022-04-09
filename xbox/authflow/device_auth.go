package authflow

import (
	"fmt"
	"time"

	"github.com/brawaru/marct/launcher/accounts"
	"github.com/brawaru/marct/xbox"
)

// DeviceAuthHandler is a function that consumes device authorisation request and returns any errors with handling it.
//
// It can be used for example, to prompt user to open the link and authorise the device.
type DeviceAuthHandler func(response xbox.DeviceAuthResponse) error

type DeviceAuthStep struct {
	Handler DeviceAuthHandler
}

const StepIDDeviceAuth = "device_auth"

func (a DeviceAuthStep) ID() string {
	return StepIDDeviceAuth
}

func (a DeviceAuthStep) Authorize(state *accounts.FlowState[IntermediateState]) error {
	is := state.IntermediateState
	deviceAuth, deviceAuthErr := xbox.RequestDeviceAuth()
	if deviceAuthErr != nil {
		return fmt.Errorf("cannot request msft device auth: %w", deviceAuthErr)
	}

	if handleErr := a.Handler(*deviceAuth); handleErr != nil {
		return fmt.Errorf("handler exit: %w", handleErr)
	}

	tokenResp, tokenAcquisitionErr := xbox.TokenAcquisitionLoop(*deviceAuth)
	if tokenAcquisitionErr != nil {
		return fmt.Errorf("cannot acquire msft token: %w", tokenAcquisitionErr)
	}

	is.MsftAccessToken = tokenResp.AccessToken
	is.MsftAccessTokenExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).Add(-1 * time.Second)
	is.MsftRefreshToken = tokenResp.RefreshToken

	return nil
}

func (a DeviceAuthStep) Refresh(state *accounts.FlowState[IntermediateState]) error {
	is := state.IntermediateState

	if is.MsftAccessTokenExpiresAt.After(time.Now()) {
		return nil
	}

	tokenResp, tokenRefreshErr := xbox.RefreshToken(is.MsftRefreshToken)
	if tokenRefreshErr != nil {
		return fmt.Errorf("cannot refresh msft token: %w", tokenRefreshErr)
	}

	is.MsftAccessToken = tokenResp.AccessToken
	is.MsftAccessTokenExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).Add(-1 * time.Second)
	is.MsftRefreshToken = tokenResp.RefreshToken

	is.MsftTokenRefreshed = true

	return nil
}
