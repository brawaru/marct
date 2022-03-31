package download

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/brawaru/marct/globstate"
	"github.com/brawaru/marct/locales"
	"github.com/brawaru/marct/network"
	"github.com/brawaru/marct/utils"
	"github.com/brawaru/marct/validfile"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"net/http"
	"net/url"
)

type Download struct {
	DownloadURL url.URL // URL from where artifact is being downloaded
	Destination string  // Where must this artifact be downloaded
	SHA1        []byte  // SHA-1 hash for this artifact
	MD5         []byte  // MD5 hash for this artifact
	// Defines whether remote hashes must be retrieved (false) or only local should be used (true).
	//
	// Hashes are retrieved in Maven-compatible way, by applying .sha1 or .md5 file extension to download URL.
	noRemoteHashes bool
	// Defines whether validation would consist of just checking if file exists
	dummyValidation bool
}

func retrieveByteBuf(u string, buf []byte) error {
	if buf == nil {
		panic("buf is nil")
	}

	expectedLen := int64(len(buf))

	resp, reqErr := network.RequestLoop(network.Get(u), network.RetryIndefinitely)
	if reqErr != nil {
		return reqErr
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

// maybe this repetitive mess below can be unified

func (d *Download) validateViaRemoteSHA1() (retrieved bool, err error) {
	if d.SHA1 == nil {
		if d.noRemoteHashes {
			return false, errors.New("sha1 is not defined")
		}

		u := d.DownloadURL
		u.Path += ".sha1"

		d.SHA1 = make([]byte, 20)

		if retrievalErr := retrieveByteBuf(u.String(), d.SHA1); retrievalErr != nil {
			d.SHA1 = []byte{}
			return false, retrievalErr
		}
	} else if len(d.SHA1) == 0 {
		return false, errors.New("previous retrieval failed")
	}

	return true, validfile.ValidateFile(d.Destination, sha1.New(), d.SHA1)
}

func (d *Download) validateViaRemoteMD5() (retrieved bool, err error) {
	if d.MD5 == nil {
		if d.noRemoteHashes {
			return false, errors.New("md5 is not defined")
		}

		u := d.DownloadURL
		u.Path += ".sha1"

		d.MD5 = make([]byte, 16)

		if err := retrieveByteBuf(u.String(), d.MD5); err != nil {
			d.MD5 = []byte{}
			return false, err
		}
	} else if len(d.MD5) == 0 {
		return false, errors.New("previous retrieval failed")
	}

	return true, validfile.ValidateFile(d.Destination, md5.New(), d.MD5)
}

func (d *Download) Validate() error {
	if d.dummyValidation {
		return validfile.ValidateExistsFile(d.Destination)
	}

	{
		retrieved, validateErr := d.validateViaRemoteSHA1()
		if validateErr == nil {
			return nil
		} else {
			if retrieved {
				return validateErr
			} else if globstate.VerboseLogs {
				println(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"File":   d.Destination,
						"Reason": validateErr.Error(),
					},
					DefaultMessage: &i18n.Message{
						ID:    "log.warn.sha1-retrieval-fail",
						Other: "failed to validate {{ .File }} via remote sha-1, retrieval failed: {{ .Reason }}",
					},
				}))
			}
		}
	}

	{
		retrieved, validateErr := d.validateViaRemoteMD5()
		if validateErr == nil {
			return nil
		} else {
			if retrieved {
				return validateErr
			} else if globstate.VerboseLogs {
				println(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"File":   d.Destination,
						"Reason": validateErr.Error(),
					},
					DefaultMessage: &i18n.Message{
						ID:    "log.warn.md5-retrieval-fail",
						Other: "failed to validate {{ .File }} via remote md5, retrieval failed: {{ .Reason }}",
					},
				}))
			}
		}
	}

	return errors.New("no suitable validation method")
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

func WithRemoteHashes(location string, dest string) (*Download, error) {
	u, urlErr := url.Parse(location)

	if urlErr != nil {
		return nil, urlErr
	}

	return WithRemoteHashesURL(*u, dest), nil
}

func WithRemoteHashesURL(u url.URL, dest string) *Download {
	return &Download{
		DownloadURL: u,
		Destination: dest,
	}
}

func WithSHA1(location string, dest string, hash string) (*Download, error) {
	u, urlErr := url.Parse(location)

	if urlErr != nil {
		return nil, urlErr
	}

	h, hashErr := hex.DecodeString(hash)
	if hashErr != nil {
		return nil, hashErr
	}

	return &Download{
		DownloadURL:    *u,
		Destination:    dest,
		SHA1:           h,
		noRemoteHashes: true,
	}, nil
}

func AsIs(location string, dest string) (*Download, error) {
	u, urlErr := url.Parse(location)

	if urlErr != nil {
		return nil, urlErr
	}

	return &Download{
		DownloadURL:     *u,
		Destination:     dest,
		dummyValidation: true,
	}, nil
}
