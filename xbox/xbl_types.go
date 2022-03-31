package xbox

import "time"

type XPropertiesMap map[string]interface{}

type XTokenRequest struct {
	Properties   XPropertiesMap `json:"Properties"`
	RelyingParty string         `json:"RelyingParty"`
	TokenType    string         `json:"TokenType"`
}

type XDisplayClaims map[string][]interface{}

type XTokenResponse struct {
	IssueInstant  time.Time      `json:"IssueInstant"`
	NotAfter      time.Time      `json:"NotAfter"`
	Token         string         `json:"Token"`
	DisplayClaims XDisplayClaims `json:"DisplayClaims"`
}
