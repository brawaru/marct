package launcher

import (
	"github.com/brawaru/marct/sdtypes"
	"github.com/go-playground/validator/v10"
)

var _ *validator.Validate

const (
	LatestSnapshotID = "latest-snapshot"
	LatestReleaseID  = "latest-release"
)

type Download struct {
	SHA1 string `json:"sha1" validate:"required"`
	Size uint64 `json:"size" validate:"required"`
	URL  string `json:"url" validate:"required"`
}

type Version struct {
	Arguments              *Arguments                      `json:"arguments,omitempty"`
	AssetIndex             *AssetIndexDescriptor           `json:"assetIndex,omitempty"`
	Assets                 *string                         `json:"assets,omitempty"`
	ComplianceLevel        *int                            `json:"complianceLevel,omitempty"`
	Downloads              map[string]Download             `json:"downloads,omitempty"`
	ID                     string                          `json:"id"`
	JavaVersion            *JavaVersionRecommendation      `json:"javaVersion,omitempty"`
	Libraries              []Library                       `json:"libraries"`
	Logging                map[string]LoggingConfiguration `json:"logging,omitempty"`
	MainClass              string                          `json:"mainClass"`
	MinimumLauncherVersion int                             `json:"minimumLauncherVersion"`
	ReleaseTime            *sdtypes.ISOTime                `json:"releaseTime"`
	Time                   *sdtypes.ISOTime                `json:"time"`
	Type                   *string                         `json:"type"`
	InheritsFrom           *string                         `json:"inheritsFrom,omitempty"`
	MinecraftArguments     *string                         `json:"minecraftArguments,omitempty"`
}

type LatestVersions struct {
	Release  string `json:"release"`
	Snapshot string `json:"snapshot"`
}

type VersionDescriptor struct {
	ID              string              `json:"id"`
	Type            string              `json:"type"`
	URL             string              `json:"url"`
	Time            sdtypes.RFC3339Time `json:"time"`
	ReleaseTime     sdtypes.RFC3339Time `json:"releaseTime"`
	SHA1            string              `json:"sha1"`
	ComplianceLevel int                 `json:"complianceLevel"`
}

type VersionsManifest struct {
	Latest   LatestVersions      `json:"latest"`
	Versions []VersionDescriptor `json:"versions"`
}

func (v *VersionsManifest) GetVersion(id string) *VersionDescriptor {
	switch id {
	case LatestSnapshotID:
		id = v.Latest.Snapshot
	case LatestReleaseID:
		id = v.Latest.Release
	}

	for _, version := range v.Versions {
		if version.ID == id {
			return &version
		}
	}

	return nil
}
