package launcher

import (
	"archive/zip"
	"fmt"
	"github.com/brawaru/marct/utils"
	"github.com/brawaru/marct/utils/slices"
	"github.com/brawaru/marct/validfile"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func saneExtractPath(name string, dest string) (string, error) {
	dp := filepath.Join(dest, filepath.FromSlash(name))
	if !strings.HasPrefix(dp, filepath.Clean(dest)+string(os.PathSeparator)) {
		return "", fmt.Errorf("illegal path: %s", name)
	}
	return dp, nil
}

func extract(r *zip.ReadCloser, f *zip.File, dest string) error {
	p, err := saneExtractPath(f.Name, dest)

	if err != nil {
		return err
	}

	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(p, f.Mode()); err != nil {
			return err
		}
	} else {
		sd := filepath.Dir(p)
		if de, err := validfile.DirExists(sd); err != nil {
			return err
		} else {
			if !de {
				var dm fs.FileMode
				dm = 0644
				dn := path.Dir(f.Name)
				df := slices.Find(r.File, func(item **zip.File, index int, slice []*zip.File) bool {
					return (*item).Name == dn
				})

				if df == nil {
					dm = (*df).Mode()
				}

				if err := os.MkdirAll(sd, dm); err != nil {
					return fmt.Errorf("cannot create parent directories for %s: %w", p, err)
				}
			}
		}

		of, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("cannot create %s: %w", p, err)
		}

		defer utils.DClose(of)

		fc, err := f.Open()
		if err != nil {
			return fmt.Errorf("cannot open %s: %w", p, err)
		}

		defer utils.DClose(fc)

		if _, err := io.Copy(of, fc); err != nil {
			return fmt.Errorf("cannot write %s: %w", p, err)
		}
	}

	return nil
}

func (w *Instance) ExtractNatives(v Version) (string, error) {
	np := filepath.Join(w.Path, "bin", utils.NewUUID())

	if err := os.MkdirAll(np, 0644); err != nil {
		return "", fmt.Errorf("cannot create bin directory: %w", err)
	}

	lp := w.LibrariesPath()

	for _, l := range v.Libraries {
		n := l.GetMatchingNatives()
		if n == nil {
			continue
		}

		ap := filepath.Join(lp, filepath.FromSlash(n.Path))

		r, err := zip.OpenReader(ap)
		if err != nil {
			return np, fmt.Errorf("cannot read %s: %w", ap, err)
		}

		for _, f := range r.File {
			if !l.Extract.Include(path.Clean(f.Name)) {
				continue
			}

			err := extract(r, f, np)

			if err != nil {
				return "", fmt.Errorf("%s: %w", ap, err)
			}
		}

		utils.DClose(r)
	}

	return np, nil
}
