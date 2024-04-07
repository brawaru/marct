package launcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/brawaru/marct/network"
	"github.com/brawaru/marct/utils"
	"github.com/brawaru/marct/utils/osfile"
	"github.com/brawaru/marct/validfile"
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

type manifestCacheKeyType int

var manifestCacheKey manifestCacheKeyType = manifestCacheKeyType(1)

type ManifestCache struct {
	CachedManifest VersionsManifest
	CachedAt       time.Time
}

func (w *Instance) FetchVersions(force bool) (manifest *VersionsManifest, err error) {
	name := filepath.Join(w.Path, filepath.FromSlash(versionsManifestPath))

	if !force {
		if v, ok := w.GetTempValue(manifestCacheKey).(ManifestCache); ok && time.Since(v.CachedAt) < versionsManifestTTL {
			manifest = &v.CachedManifest
			return
		}
	}

	expired := force || (validfile.NotExpired(name, versionsManifestTTL) != nil)

	if !expired {
		return w.ReadVersions()
	}

	req, rawErr := http.NewRequest(http.MethodGet, versionsManifestURL, nil)
	if rawErr != nil {
		err = fmt.Errorf("create request: %w", rawErr)
		return
	}

	resp, rawErr := network.PerformRequest(req, network.WithRetries())
	if rawErr != nil {
		err = fmt.Errorf("send request: %w", rawErr)
		return
	}

	defer utils.DClose(resp.Body)

	manifestFile, rawErr := osfile.New(name)
	if rawErr != nil {
		err = fmt.Errorf("create file: %w", rawErr)
		return
	}

	defer utils.DClose(manifestFile)

	bodyReadWriter := io.TeeReader(resp.Body, manifestFile)

	if rawErr := json.NewDecoder(bodyReadWriter).Decode(&manifest); rawErr != nil {
		err = fmt.Errorf("decode response: %w", rawErr)
		_ = manifestFile.Close()
		_ = os.Remove(versionsManifestPath)
		return
	}

	_, rawErr = io.Copy(manifestFile, resp.Body)
	if rawErr != nil {
		err = fmt.Errorf("write response: %w", rawErr)
		return
	}

	w.SetTempValue(manifestCacheKey, ManifestCache{
		CachedManifest: *manifest,
		CachedAt:       time.Now(),
	})

	return
}
