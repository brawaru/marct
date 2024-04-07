package download

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/brawaru/marct/network"
	"github.com/brawaru/marct/utils"
	"github.com/brawaru/marct/utils/reflutils"
	"github.com/brawaru/marct/validfile"
)

type Validator func() error

type Download struct {
	DownloadURL *url.URL    // URL from where artifact is being downloaded
	Destination string      // Where must this artifact be downloaded
	Validators  []Validator // Validators to check the downloaded artifact.
}

func retrieveRemoteHash(retrievalUrl string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, retrievalUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, reqErr := network.PerformRequest(req, network.WithRetries())
	if reqErr != nil {
		return nil, fmt.Errorf("request: %w", reqErr)
	}

	defer utils.DClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request not successful, got %v %s", resp.StatusCode, resp.Status)
	}

	buf, decodeErr := io.ReadAll(hex.NewDecoder(resp.Body))

	if decodeErr != nil {
		return nil, fmt.Errorf("hash decode err: %w", decodeErr)
	}

	fmt.Printf("hash retrieval %v: %x\n", retrievalUrl, buf)

	return buf, nil
}

func (d *Download) Validate() error {
	fmt.Printf("validating %s\n", d.Destination)

	for _, v := range d.Validators {
		if err := v(); err != nil {
			return err
		}
	}

	return nil
}

func (d *Download) download() error {
	_, err := network.Download(d.DownloadURL.String(), d.Destination)

	if err != nil {
		return err
	}

	return nil
}

func (d *Download) Download() error {
	shouldDownload := false

	if validateErr := d.Validate(); validateErr != nil {
		var v *validfile.ValidateError

		if errors.As(validateErr, &v) && v.Mismatch() {
			shouldDownload = true
		} else {
			return validateErr
		}
	}

	if shouldDownload {
		if dlErr := d.download(); dlErr != nil {
			return dlErr
		}

		// we shall not fail after the downloading
		if validateErr := d.Validate(); validateErr != nil {
			return validateErr
		}
	}

	return nil
}

type Option func(*Download) error

func WithURL(u *url.URL) Option {
	return func(d *Download) error {
		if u == nil {
			return errors.New("url is nil")
		}

		d.DownloadURL = u
		return nil
	}
}

func WithSHA1(hash string) Option {
	return func(d *Download) error {
		h, err := hex.DecodeString(hash)
		if err != nil {
			return fmt.Errorf("decode %q as hex: %w", hash, err)
		}

		d.Validators = append(d.Validators, func() error {
			if validateErr := validfile.ValidateFile(d.Destination, sha1.New(), h); validateErr != nil {
				return fmt.Errorf("validate with sha1: %w", validateErr)
			}
			return nil
		})

		return nil
	}
}

func WithRemoteSHA1() Option {
	return func(d *Download) error {
		d.Validators = append(d.Validators, func() error {
			u := *d.DownloadURL
			u.Path += ".sha1"

			remoteHash, retrievalErr := retrieveRemoteHash(u.String())
			if retrievalErr != nil {
				return fmt.Errorf("retrieve remote sha1 from %s: %w", u.String(), retrievalErr)
			}

			if validateErr := validfile.ValidateFile(d.Destination, sha1.New(), remoteHash); validateErr != nil {
				return fmt.Errorf("validate with remote sha1: %w", validateErr)
			}

			return nil
		})

		return nil
	}
}

func WithMD5(hash string) Option {
	return func(d *Download) error {
		h, err := hex.DecodeString(hash)
		if err != nil {
			return fmt.Errorf("decode %q as hex: %w", hash, err)
		}

		d.Validators = append(d.Validators, func() error {
			if validateErr := validfile.ValidateFile(d.Destination, md5.New(), h); validateErr != nil {
				return fmt.Errorf("validate with md5: %w", validateErr)
			}
			return nil
		})

		return nil
	}
}

func WithRemoteMD5() Option {
	return func(d *Download) error {
		d.Validators = append(d.Validators, func() error {
			retrievalUrl := *d.DownloadURL
			retrievalUrl.Path += ".md5"

			remoteHash, err := retrieveRemoteHash(retrievalUrl.String())
			if err != nil {
				return fmt.Errorf("retrieve remote md5 from %s: %w", retrievalUrl.String(), err)
			}

			if validateErr := validfile.ValidateFile(d.Destination, md5.New(), remoteHash); validateErr != nil {
				return fmt.Errorf("verify with remote md5: %w", validateErr)
			}

			return nil
		})

		return nil
	}
}

func NewURL(rawURL string, destination string, options ...Option) (*Download, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse %q as url: %w", rawURL, err)
	}

	return New(u, destination, options...)
}

func New(u *url.URL, destination string, options ...Option) (*Download, error) {
	d := &Download{
		DownloadURL: u,
		Destination: destination,
	}

	for _, option := range options {
		if err := option(d); err != nil {
			return nil, fmt.Errorf("apply option %s: %w", reflutils.GetFunctionName(option), err)
		}
	}

	return d, nil
}

func FromURL(rawURL string, destination string, options ...Option) error {
	d, err := NewURL(rawURL, destination, options...)
	if err != nil {
		return fmt.Errorf("new url %q: %w", rawURL, err)
	}
	return d.Download()
}

func From(u *url.URL, destination string, options ...Option) error {
	d, err := New(u, destination, options...)
	if err != nil {
		return fmt.Errorf("create download: %w", err)
	}
	return d.Download()
}
