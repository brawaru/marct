package xbox

import (
	"encoding/json"
	"errors"
	"github.com/brawaru/marct/network"
	"github.com/brawaru/marct/utils"
	"time"
)

func RequestDeviceAuth() (*DeviceAuthResponse, error) {
	resp, reqErr := network.RequestLoop(deviceAuthRequest, network.RetryIndefinitely)

	if reqErr != nil {
		return nil, reqErr
	}

	defer utils.DClose(resp.Body)

	var r DeviceAuthResponse
	return &r, json.NewDecoder(resp.Body).Decode(&r)
}

func TokenAcquisitionLoop(req DeviceAuthResponse) (*TokenResponse, error) {
	interval := time.Second * time.Duration(req.Interval)
	request := createMsftTokenRequest(req.DeviceCode)

	for {
		resp, reqErr := network.RequestLoop(request, network.RetryIndefinitely)

		if reqErr != nil {
			return nil, reqErr
		}

		decoder := json.NewDecoder(resp.Body)

		if resp.StatusCode == 200 {
			var r TokenResponse
			//goland:noinspection GoDeferInLoop
			defer utils.DClose(resp.Body)
			return &r, decoder.Decode(&r)
		} else {
			var e *TokenError

			if decodeErr := decoder.Decode(&e); decodeErr != nil {
				return nil, decodeErr
			}

			if !errors.Is(e, AuthorizationPendingErr) {
				utils.DClose(resp.Body)
				return nil, e
			}
		}

		utils.DClose(resp.Body)
		time.Sleep(interval)
	}
}

func RefreshToken(refreshToken string) (*TokenResponse, error) {
	request := createMsftRefreshTokenRequest(refreshToken)

	resp, reqErr := network.RequestLoop(request, network.RetryIndefinitely)

	if reqErr != nil {
		return nil, reqErr
	}

	defer utils.DClose(resp.Body)

	decoder := json.NewDecoder(resp.Body)

	if resp.StatusCode == 200 {
		var r TokenResponse
		//goland:noinspection GoDeferInLoop
		return &r, decoder.Decode(&r)
	} else {
		var e *TokenError

		if decodeErr := decoder.Decode(&e); decodeErr != nil {
			return nil, decodeErr
		}

		return nil, e
	}
}
