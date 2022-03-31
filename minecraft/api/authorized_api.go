package api

import (
	"io"
	"net/http"
)

type AuthorizedAPI struct {
	accessToken string
}

func (a *AuthorizedAPI) newRequest(method string, url string, body io.Reader) (*http.Request, error) {
	req, reqCreateErr := http.NewRequest(method, url, body)
	if reqCreateErr != nil {
		return nil, reqCreateErr
	}

	req.Header.Set("Authorization", "Bearer "+a.accessToken)

	return req, nil
}

func NewAuthorizedAPI(token string) AuthorizedAPI {
	return AuthorizedAPI{accessToken: token}
}
