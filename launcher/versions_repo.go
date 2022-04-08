package launcher

import (
	"os"
	"path/filepath"
)

type VersionRepository map[string]*Version

// IndexVersions tries to read all the versions in the versions' directory.
func (w *Instance) IndexVersions() VersionRepository {
	versions := make(VersionRepository)

	err := filepath.Walk(filepath.Join(w.Path, "versions"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()

			if validateID(name) != nil {
				return nil
			}

			version, err := w.ReadVersionFile(name)
			if err == nil {
				versions[name] = version
			}
		}

		return nil
	})

	if err != nil {
		return nil
	}

	return versions
}
