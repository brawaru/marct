package network

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/brawaru/marct/utils"
)

var tr = &http.Transport{
	TLSClientConfig: &tls.Config{
		Renegotiation:      tls.RenegotiateOnceAsClient,
		InsecureSkipVerify: false,
	},
}

var DefaultClient = http.Client{
	Timeout:   time.Second * 30,
	Transport: tr,
}

func createFile(name string) (*os.File, error) {
	mkdirErr := os.MkdirAll(filepath.Dir(name), os.ModePerm)

	if mkdirErr != nil {
		return nil, mkdirErr
	}

	return os.Create(name)
}

// Download sends a request to a given URL and writes response body to the destination. It does not account for status
// codes and will write any body it receives without errors.
//
// It will be removed in the future when the better APIs are available. Avoid using it.
func Download(url string, dest string, options ...Option) (written int64, err error) {
	r, e := http.NewRequest("GET", url, nil)
	if e != nil {
		err = fmt.Errorf("create request: %w", e)
		return
	}

	resp, e := PerformRequest(r, options...)

	if e != nil {
		err = fmt.Errorf("perform request: %w", e)
		return
	}

	defer utils.DClose(resp.Body)

	file, createErr := createFile(dest)

	if createErr != nil {
		return 0, createErr
	}

	defer file.Close()

	return io.Copy(file, resp.Body)
}

// ErrorHandler handles errors that occur during the execution of an action. It may return an error if it needs a raise,
// or ErrRetryRequest error, if request can be repeated. If nil is returned, then the next error handler is called.
type ErrorHandler func(e error) error

type ActionOptions struct {
	// Error handlers defined in sequential order. If error occurs during the execution of an action, then the first
	// handler is called, if it returns an error that is not ErrRetryRequest, then this error is returned to the
	// caller of the action. If the error is ErrRetryRequest, then the action is repeated. If no error is returned,
	// then the next handler is called. If no handlers handle the error, then the error is returned to the caller.
	ErrorHandlers []ErrorHandler
	// HTTP client used to execute the action.
	Client *http.Client
}

type Option func(*http.Request, *ActionOptions)

var (
	// ErrRetryRequest is the error reported by Error handler if error is considered insignificant and request might be
	// re-attempted again.
	ErrRetryRequest = errors.New("retry request")
)

func PerformRequest(request *http.Request, options ...Option) (*http.Response, error) {
	o := &ActionOptions{
		ErrorHandlers: []ErrorHandler{},
		Client:        &DefaultClient,
	}

	for _, option := range options {
		option(request, o)
	}

	for {
		resp, reqErr := o.Client.Do(request)

		if reqErr != nil {
			for _, handler := range o.ErrorHandlers {
				err := handler(reqErr)

				if err != nil {
					if errors.Is(reqErr, ErrRetryRequest) {
						continue
					}

					return nil, err
				}
			}

			return nil, reqErr
		}

		return resp, nil
	}
}
