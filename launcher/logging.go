package launcher

import (
	"github.com/brawaru/marct/launcher/download"
	"path/filepath"
)

func (w *Instance) LogConfigPath(logConfig LoggingConfiguration) string {
	return filepath.Join(w.Path, filepath.FromSlash(logConfigsPath), logConfig.File.ID)
}

func (w *Instance) DownloadLogConfig(logConfig LoggingConfiguration) error {
	dest := w.LogConfigPath(logConfig)

	if dl, err := download.WithSHA1(logConfig.File.URL, dest, logConfig.File.SHA1); err == nil {
		if dlErr := dl.Download(); dlErr != nil {
			return dlErr
		}
	} else {
		return err
	}

	return nil
}
