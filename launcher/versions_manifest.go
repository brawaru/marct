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
	"github.com/brawaru/marct/utils/pointers"
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
			manifest = pointers.Ref(v.CachedManifest)
			return
		}
	}

	expired := force || (validfile.NotExpired(name, versionsManifestTTL) != nil)

	if expired {
		req, e := http.NewRequest(http.MethodGet, versionsManifestURL, nil)
		if e != nil {
			err = fmt.Errorf("create request: %w", e)
			return
		}

		resp, e := network.PerformRequest(req, network.WithRetries())
		if e != nil {
			err = fmt.Errorf("send request: %w", e)
			return
		}

		defer utils.DClose(resp.Body)

		if e := json.NewDecoder(resp.Body).Decode(&manifest); e != nil {
			err = fmt.Errorf("decode response: %w", e)
			return
		}

		f, e := osfile.New(name)
		if e != nil {
			err = fmt.Errorf("create file: %w", e)
			return
		}

		defer utils.DClose(f)

		_, e = io.Copy(f, resp.Body)
		if e != nil {
			err = fmt.Errorf("write response: %w", e)
			return
		}

		if e := f.Sync(); e != nil {
			err = fmt.Errorf("sync file: %w", e)
			return
		}

		w.SetTempValue(manifestCacheKey, ManifestCache{
			CachedManifest: *manifest,
			CachedAt:       time.Now(),
		})
	} else {
		return w.ReadVersions()
	}

	return
}
