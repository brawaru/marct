package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/brawaru/marct/network"
	"github.com/brawaru/marct/utils"
)

type AuthResponse struct {
	Username    string        `json:"username"`
	Roles       []interface{} `json:"roles"`
	AccessToken string        `json:"access_token"`
	TokenType   string        `json:"token_type"`
	ExpiresIn   int           `json:"expires_in"`
}

type APIError struct {
	Path             string `json:"path"`
	ErrorType        string `json:"errorType"`
	ErrorCode        string `json:"error"`
	ErrorMessage     string `json:"errorMessage"`
	DeveloperMessage string `json:"developerMessage"`
}

func (e *APIError) Error() string {
	return e.ErrorMessage
}

const (
	loginWithXboxURL = "https://api.minecraftservices.com/authentication/login_with_xbox"
)

func LoginWithXbox(userHash string, xstsToken string) (*AuthResponse, error) {
	marshal, bodyMarshalErr := json.Marshal(struct {
		IdentityToken string `json:"identityToken"`
	}{
		IdentityToken: "XBL3.0 x=" + userHash + ";" + xstsToken,
	})

	if bodyMarshalErr != nil {
		return nil, bodyMarshalErr
	}

	req, reqCreateErr := http.NewRequest("POST", loginWithXboxURL, bytes.NewBuffer(marshal))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if reqCreateErr != nil {
		return nil, reqCreateErr
	}

	resp, reqErr := network.PerformRequest(req, network.WithRetries())
	if reqErr != nil {
		return nil, reqErr
	}

	defer utils.DClose(resp.Body)

	if resp.StatusCode != 200 {
		if resp.StatusCode == 400 {
			var e APIError
			if json.NewDecoder(resp.Body).Decode(&e) == nil {
				return nil, &e
			}
		}

		return nil, fmt.Errorf("bad response code: %v (%s)", resp.StatusCode, resp.Status)
	}

	var authResp AuthResponse

	if decodeErr := json.NewDecoder(resp.Body).Decode(&authResp); decodeErr != nil {
		return nil, decodeErr
	}

	return &authResp, nil
}
