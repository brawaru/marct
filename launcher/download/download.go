package download

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
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

func retrieveByteBuf(u string, buf []byte) error {
	if buf == nil {
		panic("buf is nil")
	}

	expectedLen := int64(len(buf))

	r, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, reqErr := network.PerformRequest(r, network.WithRetries())
	if reqErr != nil {
		return fmt.Errorf("request: %w", reqErr)
	}

	defer utils.DClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request not successful, got %v %s", resp.StatusCode, resp.Status)
	}

	if resp.ContentLength != expectedLen && resp.ContentLength != -1 {
		return errors.New("unexpected length")
	}

	readBytes, readErr := resp.Body.Read(buf)

	if readErr != nil {
		return readErr
	}

	if int64(readBytes) != expectedLen {
		return errors.New("expected to read 20 bytes")
	}

	return nil
}

func (d *Download) Validate() error {
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

type DownloadOption func(*Download) error

func WithURL(u *url.URL) DownloadOption {
	return func(d *Download) error {
		if u == nil {
			return errors.New("url is nil")
		}

		d.DownloadURL = u
		return nil
	}
}

func WithSHA1(hash string) DownloadOption {
	return func(d *Download) error {
		h, err := hex.DecodeString(hash)
		if err != nil {
			return fmt.Errorf("decode %q as hex: %w", hash, err)
		}

		d.Validators = append(d.Validators, func() error {
			return validfile.ValidateFile(d.Destination, sha1.New(), h)
		})

		return nil
	}
}

func WithRemoteSHA1() DownloadOption {
	return func(d *Download) error {
		d.Validators = append(d.Validators, func() error {
			r := make([]byte, 20)

			u := *d.DownloadURL
			u.Path += ".sha1"

			if retrievalErr := retrieveByteBuf(u.String(), r); retrievalErr != nil {
				return retrievalErr
			}

			return validfile.ValidateFile(d.Destination, sha1.New(), r)
		})

		return nil
	}
}

func WithMD5(hash string) DownloadOption {
	return func(d *Download) error {
		h, err := hex.DecodeString(hash)
		if err != nil {
			return fmt.Errorf("decode %q as hex: %w", hash, err)
		}

		d.Validators = append(d.Validators, func() error {
			return validfile.ValidateFile(d.Destination, md5.New(), h)
		})

		return nil
	}
}

func WithRemoteMD5() DownloadOption {
	return func(d *Download) error {
		d.Validators = append(d.Validators, func() error {
			r := make([]byte, 16)

			u := *d.DownloadURL
			u.Path += ".md5"

			if err := retrieveByteBuf(u.String(), r); err != nil {
				return fmt.Errorf("retrieve %s: %w", u.String(), err)
			}

			return validfile.ValidateFile(d.Destination, md5.New(), r)
		})

		return nil
	}
}

func NewURL(rawURL string, destination string, options ...DownloadOption) (*Download, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse %q as url: %w", rawURL, err)
	}

	return New(u, destination, options...)
}

func New(u *url.URL, destination string, options ...DownloadOption) (*Download, error) {
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

func FromURL(rawURL string, destination string, options ...DownloadOption) error {
	d, err := NewURL(rawURL, destination, options...)
	if err != nil {
		return fmt.Errorf("new url %q: %w", rawURL, err)
	}
	return d.Download()
}

func From(u *url.URL, destination string, options ...DownloadOption) error {
	d, err := New(u, destination, options...)
	if err != nil {
		return fmt.Errorf("create download: %w", err)
	}
	return d.Download()
}
