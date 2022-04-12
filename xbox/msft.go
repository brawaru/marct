package xbox

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/brawaru/marct/network"
	"github.com/brawaru/marct/utils"
)

// FIXME: many errors are unwrapped

func RequestDeviceAuth() (*DeviceAuthResponse, error) {
	req, err := createDeviceAuthRequest()
	if err != nil {
		return nil, err
	}

	resp, err := network.PerformRequest(req, network.WithRetries())

	if err != nil {
		return nil, err
	}

	defer utils.DClose(resp.Body)

	var r DeviceAuthResponse
	return &r, json.NewDecoder(resp.Body).Decode(&r)
}

func TokenAcquisitionLoop(req DeviceAuthResponse) (*TokenResponse, error) {
	interval := time.Second * time.Duration(req.Interval)
	r, err := createMsftTokenRequest(req.DeviceCode)
	if err != nil {
		return nil, err
	}

	for {
		resp, reqErr := network.PerformRequest(r, network.WithRetries())

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
	req, err := createMsftRefreshTokenRequest(refreshToken)
	if err != nil {
		return nil, err
	}

	resp, err := network.PerformRequest(req, network.WithRetries())
	if err != nil {
		return nil, err
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
