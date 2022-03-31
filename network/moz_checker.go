package network

import (
	"io"
	"net/http"
	"time"
)

type MozChecker int

var networkClient = http.Client{
	Timeout: time.Second * 5,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

var controlData = []byte("success\n")

func (c MozChecker) Run() CheckResult {
	// goal: to fetch below url and receive "success"
	// http://detectportal.firefox.com/success.txt

	resp, err := networkClient.Get("http://detectportal.firefox.com/success.txt")

	if err == nil {
		switch resp.StatusCode {
		case http.StatusTemporaryRedirect:
			var redirectUrl = resp.Header.Get("Location")

			return CheckResult{
				Status:   StatusCaptive,
				Redirect: &redirectUrl,
			}
		case http.StatusOK:
			bytes, err := io.ReadAll(resp.Body)

			if err == nil && len(bytes) == len(controlData) {
				ok := true

				for i, b := range controlData {
					if bytes[i] != b {
						ok = false
						break
					}
				}

				if ok {
					return CheckResult{
						Status: StatusOk,
					}
				}
			}
		}
	}

	return CheckResult{
		Status: StatusUnknown,
	}
}
