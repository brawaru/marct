package launcher

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/brawaru/marct/globstate"
	"github.com/brawaru/marct/launcher/download"
	"github.com/brawaru/marct/locales"
	"github.com/brawaru/marct/network"
	"github.com/brawaru/marct/utils/slices"
	"github.com/brawaru/marct/validfile"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"os"
	"path/filepath"
)

func (w *Instance) VersionFolderPath(id string) string {
	return filepath.Join(w.Path, "versions", id) // FIXME: non-sanitised input
}

func (w *Instance) VersionFilePath(id string, ext string) string {
	return filepath.Join(w.VersionFolderPath(id), id+"."+ext) // FIXME: non-sanitised input
}

func (w *Instance) DownloadVersionFile(descriptor VersionDescriptor) error {
	dest := w.VersionFilePath(descriptor.ID, "json")

	if dl, err := download.WithSHA1(descriptor.URL, dest, descriptor.SHA1); err == nil {
		if dlErr := dl.Download(); dlErr != nil {
			return dlErr
		}
	} else {
		return err
	}

	if globstate.VerboseLogs {
		println(locales.TranslateUsing(&i18n.LocalizeConfig{
			TemplateData: map[string]string{
				"TypeFile": dest,
			},
			DefaultMessage: &i18n.Message{
				ID:    "log.verbose.version-file-downloaded",
				Other: "downloaded version file: {{ .TypeFile }}",
			},
		}))
	}

	return nil
}

func (w *Instance) ReadVersionFile(id string) (*Version, error) {
	contents, readErr := os.ReadFile(w.VersionFilePath(id, "json"))

	if readErr != nil {
		return nil, readErr
	}

	var v Version

	if err := json.Unmarshal(contents, &v); err != nil {
		return nil, err
	}

	return &v, nil
}

func (w *Instance) ReadVersionWithInherits(id string) (*Version, error) {
	var all []Version // fabric-..., 1.18.1
	var allIDs []string

	currentID := id

	for currentID != "" {
		v, err := w.ReadVersionFile(currentID)
		if err != nil {
			return nil, fmt.Errorf("cannot read %q: %w", currentID, err)
		}

		all = append(all, *v)

		if v.InheritsFrom == nil {
			break
		}

		nextID := *v.InheritsFrom

		if nextID == currentID {
			return nil, fmt.Errorf("%q self-references itself for inheritance", currentID)
		} else if slices.Includes(allIDs, nextID) {
			var chain string
			for i, e := range allIDs {
				if i != 0 {
					chain += " -> "
				}

				chain += e
			}

			if chain != "" {
				chain += " -> "
			}

			chain += nextID

			return nil, fmt.Errorf("%q contains a circular reference: %s", currentID, chain)
		}

		allIDs = append(allIDs, currentID)
		currentID = nextID
	}

	l := len(all)

	switch {
	case l > 1:
		v := all[l-1]
		for i := l - 2; i >= 0; i-- {
			c := all[i]
			r, err := MergeVersions(v, c)
			if err != nil {
				return nil, fmt.Errorf("cannot merge %q with %q", c.ID, v.ID)
			}
			v = r
		}
		return &v, nil
	case l == 1:
		return &all[0], nil
	}

	return nil, errors.New("no versions to inherit")
}

func (w *Instance) downloadClientJar(versionFile Version) error {
	downloads := versionFile.Downloads

	if downloads == nil {
		return &DownloadUnavailableError{"client"}
	}

	clientDownload, hasClientDownload := downloads["client"]

	if !hasClientDownload {
		return &DownloadUnavailableError{"client"}
	}

	clientJarPath := w.VersionFilePath(versionFile.ID, "jar")

	shouldDownload := false

	if clientDownload.SHA1 == "" {
		exists, existsErr := validfile.FileExists(clientJarPath)

		if existsErr != nil {
			return existsErr
		}

		shouldDownload = !exists
	} else {
		validateErr := validfile.ValidateFileHex(clientJarPath, sha1.New(), clientDownload.SHA1)

		if validateErr != nil {
			var v *validfile.ValidateError

			if errors.As(validateErr, &v) && v.Mismatch() {
				shouldDownload = true
			} else {
				return validateErr
			}
		}
	}

	if shouldDownload {
		if _, err := network.Download(clientDownload.URL, clientJarPath); err != nil {
			return err
		}

		if clientDownload.SHA1 != "" {
			if err := validfile.ValidateFileHex(clientJarPath, sha1.New(), clientDownload.SHA1); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *Instance) DownloadVersion(versionFile Version) error {
	// TODO: download all inherits if there any

	clientJarDlErr := w.downloadClientJar(versionFile)
	if clientJarDlErr != nil {
		if errors.Is(clientJarDlErr, &DownloadUnavailableError{}) {
			if globstate.VerboseLogs {
				println(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"ID": versionFile.ID,
					},
					DefaultMessage: &i18n.Message{
						ID:    "logs.warn.client-jar-unavailable",
						Other: "no client jar downloads available for {{ .ID }}",
					},
				}))
			}
		} else {
			return clientJarDlErr
		}
	}

	if versionFile.Libraries != nil {
		if libDlErr := w.DownloadLibraries(versionFile.Libraries); libDlErr != nil {
			return libDlErr
		}
	}

	if versionFile.AssetIndex != nil {
		indexDesc := *versionFile.AssetIndex

		if indexDlErr := w.DownloadAssetIndex(indexDesc); indexDlErr != nil {
			return indexDlErr
		}

		index, readErr := w.ReadAssetIndex(indexDesc.ID)
		if readErr != nil {
			return readErr
		}

		if dlErr := w.DownloadAssets(*index); dlErr != nil {
			return dlErr
		}
	}

	if logConfig, hasLogConfig := versionFile.Logging["client"]; hasLogConfig {
		if logDlErr := w.DownloadLogConfig(logConfig); logDlErr != nil {
			return logDlErr
		}
	}

	return nil
}
