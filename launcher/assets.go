package launcher

import (
	"encoding/json"
	"github.com/brawaru/marct/globstate"
	"github.com/brawaru/marct/launcher/download"
	"github.com/brawaru/marct/locales"
	"github.com/brawaru/marct/utils/terrgroup"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"os"
	"path/filepath"
)

func (w *Instance) AssetIndexPath(id string) string {
	return filepath.Join(w.Path, filepath.FromSlash(assetIndexesPath), id+".json") // FIXME: non-sanitised input
}

func (w *Instance) AssetsVirtualPath(id string) string {
	return filepath.Join(w.Path, filepath.FromSlash(assetsVirtualPath), id) // FIXME: non-sanitised input
}

func (w *Instance) AssetsObjectsPath() string {
	return filepath.Join(w.Path, filepath.FromSlash(assetsObjectsPath))
}

func (w *Instance) DefaultAssetsObjectResolver() AssetPathResolver {
	of := w.AssetsObjectsPath()
	return func(asset Asset) string {
		return filepath.Join(of, asset.Path())
	}
}

func (w *Instance) DownloadAssetIndex(descriptor AssetIndexDescriptor) error {
	dest := w.AssetIndexPath(descriptor.ID)

	if dl, err := download.WithSHA1(descriptor.URL, dest, descriptor.SHA1); err == nil {
		if dlErr := dl.Download(); dlErr != nil {
			return dlErr
		}
	} else {
		return err
	}

	return nil
}

func (w *Instance) ReadAssetIndex(id string) (*AssetIndex, error) {
	fp := w.AssetIndexPath(id)

	var i AssetIndex

	bytes, readErr := os.ReadFile(fp)

	if readErr != nil {
		return nil, readErr
	}

	if unmarshalErr := json.Unmarshal(bytes, &i); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &i, nil
}

func (w *Instance) DownloadAssets(index AssetIndex) error {
	op := w.AssetsObjectsPath()
	g, _ := terrgroup.New(8)

	for n, asset := range index.Objects {
		// Clone variables, so they don't change in async execution later
		name := n
		a := asset

		g.Go(func() error {
			p := filepath.Join(op, filepath.FromSlash(a.Path()))

			if dl, err := download.WithSHA1(a.URL(), p, a.Hash); err == nil {
				if dlErr := dl.Download(); dlErr != nil {
					return dlErr
				}
			} else {
				return err
			}

			if globstate.VerboseLogs {
				println(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Asset": name,
						"ID":    a.Hash,
					},
					DefaultMessage: &i18n.Message{
						ID:    "log.verbose.downloaded-asset",
						Other: "downloaded asset {{ .ID }} ({{ .Asset }})",
					},
				}))
			}

			return nil
		})
	}

	return g.Wait()
}
