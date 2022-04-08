package launcher

import (
	"path/filepath"

	"github.com/brawaru/marct/globstate"
	"github.com/brawaru/marct/launcher/download"
	"github.com/brawaru/marct/locales"
	"github.com/brawaru/marct/utils/terrgroup"
	"github.com/nicksnyder/go-i18n/v2/i18n"
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

	if err := download.FromURL(descriptor.URL, dest, download.WithSHA1(descriptor.SHA1)); err != nil {
		return err
	}

	return nil
}

func (w *Instance) ReadAssetIndex(id string) (i *AssetIndex, err error) {
	err = unmarshalJSONFile(w.AssetIndexPath(id), i)

	return
}

func (w *Instance) DownloadAssets(index AssetIndex) error {
	op := w.AssetsObjectsPath()
	g, _ := terrgroup.New(8)

	for n, a := range index.Objects {
		// Clone variables, so they don't change in async execution later
		name := n
		asset := a

		g.Go(func() error {
			p := filepath.Join(op, filepath.FromSlash(asset.Path()))

			if err := download.FromURL(asset.URL(), p, download.WithSHA1(asset.Hash)); err != nil {
				return err
			}

			if globstate.VerboseLogs {
				println(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Asset": name,
						"ID":    asset.Hash,
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
