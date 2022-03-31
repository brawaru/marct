package launcher

import (
	"github.com/brawaru/marct/json_helpers"
	"github.com/go-playground/validator/v10"
)

var _ *validator.Validate

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
	ReleaseTime            *json_helpers.ISOTime           `json:"releaseTime"`
	Time                   *json_helpers.ISOTime           `json:"time"`
	Type                   *string                         `json:"type"`
	InheritsFrom           *string                         `json:"inheritsFrom,omitempty"`
	MinecraftArguments     *string                         `json:"minecraftArguments,omitempty"`
}

type Latest struct {
	Release  string `json:"release"`
	Snapshot string `json:"snapshot"`
}

type VersionDescriptor struct {
	ID              string                   `json:"id"`
	Type            string                   `json:"type"`
	URL             string                   `json:"url"`
	Time            json_helpers.RFC3339Time `json:"time"`
	ReleaseTime     json_helpers.RFC3339Time `json:"releaseTime"`
	SHA1            string                   `json:"sha1"`
	ComplianceLevel int                      `json:"complianceLevel"`
}

type VersionsManifest struct {
	Latest   Latest              `json:"latest"`
	Versions []VersionDescriptor `json:"versions"`
}

func (v *VersionsManifest) GetLatestRelease() *VersionDescriptor {
	return v.GetVersion(v.Latest.Release)
}

func (v *VersionsManifest) GetLatestSnapshot() *VersionDescriptor {
	return v.GetVersion(v.Latest.Snapshot)
}

func (v *VersionsManifest) GetVersion(id string) *VersionDescriptor {
	for _, version := range v.Versions {
		if version.ID == id {
			return &version
		}
	}

	return nil
}
