package launcher

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/brawaru/marct/utils"
	"github.com/brawaru/marct/validfile"
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

type AssetIndexMappingType byte

const (
	AsObjects   AssetIndexMappingType = iota // Assets should not be mapped, game reads assets manually.
	AsResources                              // Assets should be mapped in game directory, in resources folder.
	AsVirtual                                // Assets should be mapped to virtual assets directory.
)

type AssetIndex struct {
	MapToResources *bool            `json:"map_to_resources,omitempty"` // Whether the assets should be virtualized into the ${game_dir}/resources directory.
	Objects        map[string]Asset `json:"objects"`                    // All asset objects mapped under their file paths.
	Virtual        *bool            `json:"virtual,omitempty"`          // Whether the assets should be virtualized into the ${root}/assets/virtual directory.
}

func (i *AssetIndex) MapType() AssetIndexMappingType {
	switch {
	case i.MapToResources != nil && *i.MapToResources:
		return AsResources
	case i.Virtual != nil && *i.Virtual:
		return AsVirtual
	default:
		return AsObjects
	}
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

	if mkdirErr := os.MkdirAll(virtualPath, os.ModePerm); mkdirErr != nil {
		return mkdirErr
	}

	for p, o := range i.Objects {
		err := i.virtualizeAsset(pr, o, virtualPath, p)
		if err != nil {
			return fmt.Errorf("virtualize %q: %w", p, err)
		}

		// if err := os.Link(s, d); err != nil {
		// 	return err
		// }
	}

	return nil
}

func (*AssetIndex) virtualizeAsset(pr AssetPathResolver, o Asset, vp string, p string) error {
	s := pr(o)
	d := filepath.Join(vp, filepath.FromSlash(p))

	if dd := filepath.Dir(d); dd != "." {
		if err := os.MkdirAll(dd, os.ModePerm); err != nil {
			return err
		}
	}

	sf, err := os.OpenFile(s, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open source file %q: %w", s, err)
	}
	defer utils.DClose(sf)

	df, err := os.OpenFile(d, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open destination file %q: %w", d, err)
	}
	defer utils.DClose(df)

	_, err = io.Copy(df, sf)
	if err != nil {
		return fmt.Errorf("copy contents of %q to %q", s, d)
	}

	return nil
}
