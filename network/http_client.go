package network

import (
	"crypto/tls"
	"errors"
	"github.com/brawaru/marct/utils"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

var tr = &http.Transport{
	TLSClientConfig: &tls.Config{
		Renegotiation:      tls.RenegotiateOnceAsClient,
		InsecureSkipVerify: false,
	},
}

var HttpClient = http.Client{
	Timeout:   time.Second * 30,
	Transport: tr,
}

var RetryIndefinitely = func(error) bool { return true }

func LimitRetries(maxRetries int) func(error) bool {
	retries := 0

	return func(error) bool {
		retries += 1
		return retries < maxRetries
	}
}

func Get(url string) func(client http.Client) (*http.Response, error) {
	return func(client http.Client) (*http.Response, error) {
		return client.Get(url)
	}
}

func RequestLoop(applyRequest func(client http.Client) (*http.Response, error), handleError func(err error) bool) (*http.Response, error) {
	for {
		body, err := applyRequest(HttpClient)

		if err == nil {
			return body, nil
		}

		retry := false

		if handleError(err) {
			rootCause := utils.TraverseCauses(err, errors.Unwrap)

			switch rootCause.(type) {
			case *net.DNSError:
				retry = true
			case *syscall.Errno:
				retry = errors.Is(rootCause, syscall.ECONNRESET)
			}
		}

		if retry {
			WaitForConnection()
		} else {
			return body, err
		}
	}
}

func createFile(name string) (*os.File, error) {
	mkdirErr := os.MkdirAll(filepath.Dir(name), os.ModePerm)

	if mkdirErr != nil {
		return nil, mkdirErr
	}

	return os.Create(name)
}

func Download(url string, dest string) (written int64, err error) {
	resp, reqErr := RequestLoop(Get(url), RetryIndefinitely)

	if reqErr != nil {
		return 0, reqErr
	}

	defer utils.DClose(resp.Body)

	file, createErr := createFile(dest)

	if createErr != nil {
		return 0, createErr
	}

	defer file.Close()

	return io.Copy(file, resp.Body)
}

func Do(request *http.Request, handleError func(err error) bool) (*http.Response, error) {
	return RequestLoop(func(client http.Client) (*http.Response, error) {
		return client.Do(request)
	}, handleError)
}
