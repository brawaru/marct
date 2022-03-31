package api

import (
	"crypto/rsa"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/brawaru/marct/network"
	"github.com/brawaru/marct/utils"
	RSA "github.com/dvsekhvalnov/jose2go/keys/rsa"
	_ "github.com/lestrrat-go/jwx"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
)

const (
	EntitlementMinecraftProduct        = "product_minecraft"
	EntitlementMinecraftGame           = "game_minecraft"
	EntitlementMinecraftBedrockProduct = "product_minecraft_bedrock"
	EntitlementMinecraftBedrockGame    = "game_minecraft_bedrock"
)

type Entitlement struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

type EntitlementsResponse struct {
	Items     []Entitlement `json:"items"`
	Signature string        `json:"signature"`
	KeyId     string        `json:"keyId"`
	RequestId string        `json:"requestId"`
}

const entitlementsUrl = "https://api.minecraftservices.com/entitlements/license"

func (a *AuthorizedAPI) GetEntitlements() (*EntitlementsResponse, error) {
	req, reqCreateErr := a.newRequest("GET", entitlementsUrl, nil)
	if reqCreateErr != nil {
		return nil, reqCreateErr
	}

	query := req.URL.Query()
	query.Set("requestId", utils.NewUUID())
	req.URL.RawQuery = query.Encode()

	resp, respErr := network.Do(req, network.RetryIndefinitely)

	if respErr != nil {
		return nil, respErr
	}

	if resp.StatusCode != 200 {
		if resp.StatusCode == 400 {
			var e APIError
			if json.NewDecoder(resp.Body).Decode(&e) == nil {
				return nil, &e
			}
		}

		return nil, fmt.Errorf("bad response code: %v (%s)", resp.StatusCode, resp.Status)
	}

	defer utils.DClose(resp.Body)

	var e EntitlementsResponse
	return &e, json.NewDecoder(resp.Body).Decode(&e)
}

//go:embed minecraft_pubkey.txt
var mcPubKey []byte

var mcKey *rsa.PublicKey

func init() {
	var keyReadErr error
	mcKey, keyReadErr = RSA.ReadPublic(mcPubKey)

	if keyReadErr != nil {
		panic(keyReadErr)
	}
}

func (r *EntitlementsResponse) ReadSignature() (jwt.Token, error) {
	// FIXME: borked code
	key := jwk.NewRSAPublicKey()
	if rawKeyReadErr := key.FromRaw(mcKey); rawKeyReadErr != nil {
		return nil, fmt.Errorf("cannot create jwk from rsa key: %w", rawKeyReadErr)
	}

	_ = key.Set("kid", "1")

	keySet := jwk.NewSet()
	keySet.Add(key)

	token, parseErr := jwt.ParseString(r.Signature, jwt.WithKeySet(keySet))
	if parseErr != nil {
		return nil, fmt.Errorf("cannot parse jwt: %w", parseErr)
	}

	return token, nil
}
