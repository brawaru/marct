package mozchecker

import (
	"bytes"
	"io"
	"net/http"

	"github.com/brawaru/marct/network/concheck"
)

type mozChecker struct {
	client http.Client
}

var controlData = []byte("success\n")

func (c *mozChecker) Run() (concheck.CheckResult, error) {
	// goal: to fetch below url and receive "success"
	// http://detectportal.firefox.com/success.txt

	resp, err := c.client.Get("http://detectportal.firefox.com/success.txt")

	if err == nil {
		switch resp.StatusCode {
		case http.StatusFound:
			fallthrough
		case http.StatusTemporaryRedirect:
			var redirectURL = resp.Header.Get("Location")

			return concheck.CheckResult{
				Status:   concheck.StatusCaptive,
				Redirect: &redirectURL,
			}, nil
		case http.StatusOK:
			b, err := io.ReadAll(resp.Body)

			if err == nil && bytes.Equal(controlData, b) {
				return concheck.CheckResult{
					Status: concheck.StatusOK,
				}, nil
			}
		}
	}

	return concheck.CheckResult{
		Status: concheck.StatusUnknown,
	}, nil
}

func NewMozChecker() concheck.Checker {
	return &mozChecker{
		client: http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}
