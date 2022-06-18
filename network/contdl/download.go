package contdl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/brawaru/marct/network"
	"github.com/brawaru/marct/utils"
)

// RangeSupport is a byte that indicates whether the server supports range requests.
type RangeSupport byte

const (
	RangeSupportUnknown RangeSupport = iota // Range support is not yet known because no request has been made.
	RangeSupportYes                         // Byte serving is supported. When client resumes a download, it must send Range header.
	RangeSupportNo                          // Byte serving is not supported. Client has to re-download the entire file.
)

// ResumeableDownload represents an HTTP download that can be resumed in case of failure, if byte serving is supported
// by the server.
type ResumeableDownload struct {
	baseRequest  *http.Request   // The initial request used to start the download.
	dest         string          // Destination where the response body will be written.
	ctx          context.Context // Context used to cancel the download.
	f            *os.File        // File where the response body will be written.
	rangeSupport RangeSupport    // Whether byte serving is supported by the server.
	retryCount   int             // Number of times the download has been retried. Reset every time a succesful request is made.
}

func (r *ResumeableDownload) open() error {
	f, err := os.OpenFile(r.dest, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	r.f = f
	return nil
}

func (r *ResumeableDownload) Close() error {
	err := r.f.Close()
	if err != nil {
		return fmt.Errorf("close file: %w", err)
	}
	return nil
}

// ServerMisconfiguredErr is an error that is reported when the server does not follow the RFC for byte serving and
// sends an incorrect response in request to send a slice of bytes.
type ServerMisconfiguredErr struct {
	Expected int // Expected status code.
	Received int // Received status code.
}

func (e *ServerMisconfiguredErr) Error() string {
	return "server is not configured correctly to handle byte serving"
}

func (e *ServerMisconfiguredErr) Is(target error) bool {
	t, ok := target.(*ServerMisconfiguredErr)
	return ok && (t.Expected == 0 || t.Expected == e.Expected) && (t.Received == 0 || t.Received == e.Received)
}

// IOErr is an error that is reported when there is an error related to writing or reading the data.
type IOErr struct {
	Err error
}

func (e *IOErr) Error() string {
	return fmt.Sprintf("io error: %s", e.Err.Error())
}

func (e *IOErr) Unwrap() error {
	return e.Err
}

func (e *IOErr) Is(target error) bool {
	t, ok := target.(*IOErr)
	return ok && errors.Is(t.Err, e.Err)
}

// RequestErr is an error that is reported when there is an error when processing the request.
type RequestErr struct {
	Err error
}

func (e *RequestErr) Error() string {
	return fmt.Sprintf("request error: %s", e.Err.Error())
}

func (e *RequestErr) Unwrap() error {
	return e.Err
}

func (e *RequestErr) Is(target error) bool {
	t, ok := target.(*RequestErr)
	return ok && errors.Is(t.Err, e.Err)
}

func (r *ResumeableDownload) downloadLoop() error {
	// create a new request based off the base request
	req := r.baseRequest.Clone(r.ctx)

	// storing a value to check later whether we we're resuming the download
	isResuming := false

	// if the server supports byte serving, then set the Range header to number of bytes already downloaded
	// otherwise seek file to the beginning because we'll be re-downloading the entire file
	switch r.rangeSupport {
	case RangeSupportNo:
		fallthrough
	case RangeSupportUnknown:
		_, err := r.f.Seek(0, io.SeekStart)
		if err != nil {
			return &IOErr{fmt.Errorf("seek file: %w", err)}
		}
	case RangeSupportYes:
		currentPosition, err := r.f.Seek(0, io.SeekCurrent)
		if err != nil {
			return &IOErr{fmt.Errorf("getpos: %w", err)}
		}

		if currentPosition != 0 {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", currentPosition))
		}

		isResuming = true
	}

	// send request
	resp, err := network.PerformRequest(req, network.WithRetries())

	// if we have a response, then check if the server supports byte serving regardless of the error
	if resp != nil && r.rangeSupport == RangeSupportUnknown {
		if resp.Header.Get("Accept-Ranges") == "bytes" {
			r.rangeSupport = RangeSupportYes
		} else {
			r.rangeSupport = RangeSupportNo
		}
	}

	// now check for the error, if there's any, send it back to the loop
	if err != nil {
		return &RequestErr{fmt.Errorf("perform request: %w", err)}
	}

	defer utils.DClose(resp.Body)

	r.retryCount = 0

	// server that sends partial content responds with http.StatusPartialContent
	// if it does not do that, then something is very wrong, so we're throwing the error

	switch resp.StatusCode {
	case http.StatusPartialContent:
		if !isResuming {
			return &ServerMisconfiguredErr{
				Expected: http.StatusOK,
				Received: resp.StatusCode,
			}
		}
	case http.StatusOK:
		if isResuming {
			return &ServerMisconfiguredErr{
				Expected: http.StatusPartialContent,
				Received: resp.StatusCode,
			}
		}
	// TODO: handle redirect response codes?
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// write whatever we have to the file
	_, err = io.Copy(r.f, resp.Body)
	if err != nil {
		return &IOErr{fmt.Errorf("copy: %w", err)}
	}

	return nil
}

// COA is a short for Corse of Action, is the code returned by determining function based on error that occured
// during the download.
type COA int

const (
	COAAbort COA = iota // Abort the download and report error to the caller.
	COARetry            // Retry the download and dismiss error quietly.
	COAReset            // Start the download from the beginning and dismiss error quietly.
)

// determineCOA determines the COA based on the error that occured during the download taking into account the current
// state.
func (r *ResumeableDownload) determineCOA(err error) COA {
	if errors.Is(err, &ServerMisconfiguredErr{}) {
		return COAReset // We can try to reset the download, if that fails we'll just abort.
	}

	if errors.Is(err, &RequestErr{}) {
		// request error, can we allow one more retry?
		if r.retryCount < 5 { // FIXME: ideally make the maximum number configurable
			r.retryCount++
			return COARetry
		}
	}

	return COAAbort
}

func (r *ResumeableDownload) reset() error {
	_, err := r.f.Seek(0, io.SeekStart)

	if err != nil {
		return &IOErr{fmt.Errorf("seek file: %w", err)}
	}

	return nil
}

// Start starts the download loop.
func (r *ResumeableDownload) Start() error {
dlLoop:
	for {
		select {
		case <-r.ctx.Done():
			return r.ctx.Err()

		default:
			err := r.downloadLoop()

			if err != nil {
				switch r.determineCOA(err) {
				case COAAbort:
					return err
				case COARetry:
					continue
				case COAReset:
					if err := r.reset(); err != nil {
						return fmt.Errorf("reset: %w", err)
					}

					continue
				}
			}

			break dlLoop
		}
	}

	// dowload is complete, sync the file for safety sake
	if err := r.f.Sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	return nil
}

func NewResumeableDownload(req *http.Request, dest string, ctx context.Context) *ResumeableDownload {
	r := &ResumeableDownload{
		baseRequest: req.Clone(req.Context()),
		dest:        dest,
		ctx:         ctx,
	}
	r.open()
	return r
}
