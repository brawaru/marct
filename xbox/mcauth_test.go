package xbox

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFlow(t *testing.T) {
	auth, err := RequestDeviceAuth()
	if !assert.NoError(t, err, "must not error when requesting device auth") {
		return
	}

	if !assert.NotNil(t, auth, "returned reference must be not nil") {
		return
	}

	t.Log(auth.Message)

	tokenResp, tokenErr := TokenAcquisitionLoop(*auth)

	if !assert.NoError(t, tokenErr, "must not error when requesting token") {
		return
	}

	if !assert.NotNil(t, tokenResp, "returned reference must be not nil") {
		return
	}

	xblAuthResp, xblAuthErr := AuthXBLUser(tokenResp.AccessToken)

	if !assert.NoError(t, xblAuthErr, "must not error when authenticating XBL user") {
		return
	}

	if !assert.NotNil(t, xblAuthResp, "XBL token response must not be nil") {
		return
	}

	xstsTokenResp, xstsTokenReqErr := GetXSTSToken(xblAuthResp.Token)

	if !assert.NoError(t, xstsTokenReqErr, "must not error when requesting XSTS token") {
		return
	}

	if !assert.NotNil(t, xstsTokenResp, "XSTS token response must not be nil") {
		return
	}

	// here
}
