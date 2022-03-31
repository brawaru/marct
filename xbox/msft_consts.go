package xbox

import _ "embed"

//go:embed client_id.txt
var clientId string

var (
	AuthorizationPendingErr  = &TokenError{Type: "authorization_pending"}
	AuthorizationDeclinedErr = &TokenError{Type: "authorization_declined"}
	BadVerificationCodeErr   = &TokenError{Type: "bad_verification_code"}
	ExpiredTokenErr          = &TokenError{Type: "expired_token"}
)
