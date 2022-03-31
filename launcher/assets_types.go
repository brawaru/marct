package launcher

import (
	"github.com/brawaru/marct/validfile"
	"os"
	"path/filepath"
)

type AssetIndexDescriptor struct {
	Download
	TotalSize uint64 `json:"totalSize" validate:"required"`
	ID        string `json:"id" validate:"required"`
}

type Asset struct {
	Hash string `json:"hash"`
	Size int    `json:"size"`
}

func (a *Asset) Path() string {
	initials := []rune(a.Hash)[0:2]
	return string(initials) + "/" + a.Hash
}

func (a *Asset) URL() string {
	return resourcesURL + "/" + a.Path()
}

type AssetIndex struct {
	MapToResources *bool            `json:"map_to_resources,omitempty"`
	Objects        map[string]Asset `json:"objects"`
}

type AssetPathResolver func(asset Asset) string

func (i *AssetIndex) Virtualize(pr AssetPathResolver, virtualPath string) error {
	if exists, existsErr := validfile.DirExists(virtualPath); existsErr == nil {
		if exists {
			if removeErr := os.RemoveAll(virtualPath); removeErr != nil {
				return removeErr
			}
		}
	} else {
		return existsErr
	}

	if mkdirErr := os.Mkdir(virtualPath, os.ModePerm); mkdirErr != nil {
		return mkdirErr
	}

	for p, o := range i.Objects {
		s := pr(o)
		d := filepath.Join(virtualPath, filepath.FromSlash(p))

		if dd := filepath.Dir(d); dd != "." {
			if err := os.MkdirAll(dd, os.ModePerm); err != nil {
				return err
			}
		}

		if err := os.Link(s, d); err != nil {
			return err
		}
	}

	return nil
}
