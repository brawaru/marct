package launcher

import (
	"fmt"
	"path/filepath"

	"github.com/brawaru/marct/launcher/download"
)

func (w *Instance) LogConfigPath(logConfig LoggingConfiguration) string {
	return filepath.Join(w.Path, filepath.FromSlash(logConfigsPath), logConfig.File.ID)
}

func (w *Instance) DownloadLogConfig(logConfig LoggingConfiguration) error {
	dest := w.LogConfigPath(logConfig)

	if err := download.FromURL(logConfig.File.URL, dest, download.WithSHA1(logConfig.File.SHA1)); err != nil {
		return fmt.Errorf("download %s to %q: %s", logConfig.File.URL, dest, err)
	}

	return nil
}
