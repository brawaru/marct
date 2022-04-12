package network

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func PostRequest(url string, contentType string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	return req, nil
}

func FormPostRequest(url string, form url.Values) (*http.Request, error) {
	return PostRequest(url, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
}
