package xbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/brawaru/marct/network"
	"github.com/brawaru/marct/utils"
)

// xbox live authentication

func createXTokenRequest(url string, request XTokenRequest) (*http.Request, error) {
	body, marshalErr := json.Marshal(request)

	if marshalErr != nil {
		return nil, marshalErr
	}

	req, reqErr := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if reqErr != nil {
		return nil, reqErr
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func mapXTokenResponse(resp *http.Response) (*XTokenResponse, error) {
	switch resp.StatusCode {
	case http.StatusOK:
		var res XTokenResponse
		return &res, json.NewDecoder(resp.Body).Decode(&res)
	case http.StatusUnauthorized:
		var e XTokenError
		if json.NewDecoder(resp.Body).Decode(&e) == nil {
			// if we couldn't decode means the error is something else in which case it is better to fall through
			return nil, &e
		}
	}

	return nil, fmt.Errorf("invalid response code %v (%s)", resp.StatusCode, resp.Status) // FIXME: wrap this error
}

func AuthXBLUser(accessToken string) (*XTokenResponse, error) {
	req, reqCreateErr := createXTokenRequest("https://user.auth.xboxlive.com/user/authenticate", XTokenRequest{
		Properties: XPropertiesMap{
			"AuthMethod": "RPS",
			"SiteName":   "user.auth.xboxlive.com",
			"RpsTicket":  "d=" + accessToken,
		},
		RelyingParty: RpAuthXboxLive,
		TokenType:    "JWT",
	})

	if reqCreateErr != nil {
		return nil, reqCreateErr
	}

	resp, reqErr := network.Do(req, network.RetryIndefinitely)
	if reqErr != nil {
		return nil, reqErr
	}

	defer utils.DClose(resp.Body)

	return mapXTokenResponse(resp)
}

func GetXSTSToken(xblToken string) (*XTokenResponse, error) {
	req, reqCreateErr := createXTokenRequest("https://xsts.auth.xboxlive.com/xsts/authorize", XTokenRequest{
		Properties: XPropertiesMap{
			"SandboxId":  "RETAIL",
			"UserTokens": []string{xblToken},
		},
		RelyingParty: RpMinecraftServices,
		TokenType:    "JWT",
	})

	if reqCreateErr != nil {
		return nil, reqCreateErr
	}

	resp, reqErr := network.Do(req, network.RetryIndefinitely)
	if reqErr != nil {
		return nil, reqErr
	}

	defer utils.DClose(resp.Body)

	return mapXTokenResponse(resp)
}
