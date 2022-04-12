package api

import (
	"encoding/json"
	"fmt"

	"github.com/brawaru/marct/network"
	"github.com/brawaru/marct/utils"
)

type Texture struct {
	ID    string `json:"id"`
	State string `json:"state"`
	URL   string `json:"url"`
	Alias string `json:"alias"`
}

type Skin struct {
	Texture
	Variant string
}

type Profile struct {
	ID    string    `json:"id"`
	Name  string    `json:"name"`
	Skins []Skin    `json:"skins"`
	Capes []Texture `json:"capes"`
}

const profileUrl = "https://api.minecraftservices.com/minecraft/profile"

func (a *AuthorizedAPI) GetProfile() (*Profile, error) {
	req, reqCreateErr := a.newRequest("GET", profileUrl, nil)
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

	var r Profile
	return &r, json.NewDecoder(resp.Body).Decode(&r)
}
