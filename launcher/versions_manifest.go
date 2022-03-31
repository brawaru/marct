package launcher

import (
	"encoding/json"
	"github.com/brawaru/marct/network"
	"github.com/brawaru/marct/validfile"
	"os"
	"path/filepath"
)

func (w *Instance) ReadVersions() (manifest *VersionsManifest, err error) {
	name := filepath.Join(w.Path, filepath.FromSlash(versionsManifestPath))

	var bytes []byte
	bytes, err = os.ReadFile(name)

	if err == nil {
		err = json.Unmarshal(bytes, &manifest)
	}

	return
}

func (w *Instance) FetchVersions(force bool) (manifest *VersionsManifest, err error) {
	name := filepath.Join(w.Path, filepath.FromSlash(versionsManifestPath))

	expired := force || validfile.NotExpired(name, versionsManifestTTL) != nil

	if expired {
		_, err = network.Download(versionsManifestURL, name)
	}

	if err == nil {
		return w.ReadVersions()
	}

	return
}
