package xbox

import (
	"net/http"
	"net/url"
)

func deviceAuthRequest(client http.Client) (*http.Response, error) {
	return client.PostForm("https://login.microsoftonline.com/consumers/oauth2/v2.0/devicecode", url.Values{
		"client_id": []string{clientId},
		"scope":     []string{"XboxLive.signin XboxLive.offline_access"},
	})
}

func createMsftTokenRequest(deviceCode string) func(client http.Client) (*http.Response, error) {
	return func(client http.Client) (*http.Response, error) {
		return client.PostForm("https://login.microsoftonline.com/consumers/oauth2/v2.0/token", url.Values{
			"client_id":   []string{clientId},
			"grant_type":  []string{"urn:ietf:params:oauth:grant-type:device_code"},
			"device_code": []string{deviceCode},
		})
	}
}

func createMsftRefreshTokenRequest(refreshToken string) func(client http.Client) (*http.Response, error) {
	return func(client http.Client) (*http.Response, error) {
		return client.PostForm("https://login.microsoftonline.com/consumers/oauth2/v2.0/token", url.Values{
			"client_id":     []string{clientId},
			"grant_type":    []string{"refresh_token"},
			"refresh_token": []string{refreshToken},
			"scope":         []string{"XboxLive.signin XboxLive.offline_access"},
		})
	}
}
