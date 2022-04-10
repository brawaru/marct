package launcher

import (
	"crypto/sha1"
	"errors"
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

type AssetsMappingMethod byte

const (
	NoMapping AssetsMappingMethod = iota // Assets should not be mapped, game reads assets manually.
	AsCopies                             // Assets should be mapped in game directory, in resources folder.
	AsLinks                              // Assets should be mapped to virtual assets directory.
)

type AssetIndex struct {
	MapToResources *bool            `json:"map_to_resources,omitempty"` // Whether the assets should be virtualized into the ${game_dir}/resources directory.
	Objects        map[string]Asset `json:"objects"`                    // All asset objects mapped under their file paths.
	Virtual        *bool            `json:"virtual,omitempty"`          // Whether the assets should be virtualized into the ${root}/assets/virtual directory.
}

func (i *AssetIndex) GetMappingMethod() AssetsMappingMethod {
	switch {
	case i.MapToResources != nil && *i.MapToResources:
		return AsCopies
	case i.Virtual != nil && *i.Virtual:
		return AsLinks
	default:
		return NoMapping
	}
}

type AssetPathResolver func(asset Asset) string

func (i *AssetIndex) Virtualize(pr AssetPathResolver, virtualPath string, method AssetsMappingMethod) error {
	if mkdirErr := os.MkdirAll(virtualPath, os.ModePerm); mkdirErr != nil {
		return mkdirErr
	}

	if len(i.Objects) == 0 {
		return nil // NOOP
	}

	for p, o := range i.Objects {
		err := i.virtualizeAsset(pr, o, virtualPath, p, method)
		if err != nil {
			return fmt.Errorf("virtualize %q: %w", p, err)
		}
	}

	return nil
}

func (*AssetIndex) virtualizeAsset(pr AssetPathResolver, o Asset, vp string, p string, m AssetsMappingMethod) error {
	s := pr(o)
	d := filepath.Join(vp, filepath.FromSlash(p))

	if dd := filepath.Dir(d); dd != "." {
		if err := os.MkdirAll(dd, os.ModePerm); err != nil {
			return err
		}
	}

	if err := validfile.ValidateFileHex(d, sha1.New(), o.Hash); err != nil {
		var ve *validfile.ValidateError
		if !errors.As(err, &ve) || !ve.Mismatch() {
			return fmt.Errorf("validate existing file %q: %w", d, err)
		}
	} else {
		return nil // NOOP
	}

	if err := validfile.ValidateFileHex(s, sha1.New(), o.Hash); err != nil {
		return fmt.Errorf("validate source file %q: %w", s, err)
	}

	switch m {
	case AsCopies:
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
	case AsLinks:
		if err := os.Link(s, d); err != nil {
			return fmt.Errorf("link %q to %q: %w", s, d, err)
		}
	default:
		return fmt.Errorf("illegal mapping method %v", m)
	}

	if err := validfile.ValidateFileHex(d, sha1.New(), o.Hash); err != nil {
		return fmt.Errorf("post-map validate destination file %q: %w", d, err)
	}

	return nil
}
