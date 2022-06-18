package launcher

import (
	"archive/zip"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/brawaru/marct/utils"
	"github.com/brawaru/marct/utils/unzipper"
	"github.com/brawaru/marct/validfile"
)

func (w *Instance) ExtractNatives(v Version) (string, error) {
	np := filepath.Join(w.Path, "bin", utils.NewUUID())

	if err := os.MkdirAll(np, 0777); err != nil {
		return "", fmt.Errorf("cannot create bin directory: %w", err)
	}

	lp := w.LibrariesPath()

	nativeValidator := func(e *unzipper.Unzipper, name string, dst string, f *os.File) error {
		p := name + ".sha1"

		if !e.Entries.HasKey(p) {
			return nil
		}

		ce := e.Entries.Get(p)
		rc, err := ce.Open()
		if err != nil {
			return &unzipper.AbortErr{
				Err: fmt.Errorf("cannot open %q: %w", p, err),
			}
		}
		defer utils.DClose(rc)

		var s string
		{
			buf := new(strings.Builder)
			if _, err := io.Copy(buf, rc); err != nil {
				return &unzipper.AbortErr{
					Err: fmt.Errorf("cannot read %q: %w", p, err),
				}
			}
			s = buf.String()
		}

		s = strings.TrimRight(s, "\r\n")

		err = validfile.ValidateFileHex(dst, sha1.New(), s)

		var v *validfile.ValidateError
		if errors.As(err, &v) && !v.Mismatch() {
			return &unzipper.AbortErr{
				Err: v.Err,
			}
		}

		return err
	}

	metaSkipper := func(o *ExtractOptions) unzipper.EntryProcessor {
		return func(e *unzipper.Unzipper, name string, f *zip.File, dest string) error {
			if !f.FileInfo().IsDir() {
				switch filepath.Ext(name) {
				case ".sha1":
					fallthrough
				case ".git":
					return unzipper.ErrSkip
				default:
					return nil
				}
			}

			if o == nil || o.Include(name) {
				return nil
			}

			return unzipper.ErrSkip
		}
	}

	for _, l := range v.Libraries {
		n := l.GetMatchingNatives()
		if n == nil {
			continue
		}

		ap := filepath.Join(lp, filepath.FromSlash(n.Path))

		err := unzipper.Unzip(ap, np, unzipper.WithFileValidator(nativeValidator), unzipper.WithEntryProcessor(metaSkipper(l.Extract)))
		if err != nil {
			return np, fmt.Errorf("extract native %q: %w", l.Coordinates.String(), err)
		}
	}

	return np, nil
}
