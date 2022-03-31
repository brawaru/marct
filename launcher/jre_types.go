package launcher

import (
	"github.com/brawaru/marct/json_helpers"
	"github.com/brawaru/marct/launcher/java"
)

type JavaVersionRecommendation struct {
	Component    string `json:"component" validate:"required"`
	MajorVersion int    `json:"majorVersion" validate:"required"`
}

type JavaFile struct {
	Downloads  map[string]Download `json:"downloads"`  // Map of downloads per type (lzma, raw)
	Executable bool                `json:"executable"` // Whether this file must be granted executable permission on Linux
	Type       java.FileType       `json:"type"`       // File type
	Target     string              `json:"target"`     // If file is a link, target to where it links to
}

type JavaManifest struct {
	Files map[string]JavaFile `json:"files"` // All files to be extracted
}

type JavaVersion struct {
	Name     string                   `json:"name"`     // Full version of the JRE
	Released json_helpers.RFC3339Time `json:"released"` // Time when the version was released
}

type JavaVersionDescriptor struct {
	Manifest Download    `json:"manifest"` // Download for the JavaManifest file
	Version  JavaVersion `json:"version"`  // Represents Java version and release time
}

// MostRecent returns most recently released Java version from the JavaVersionDescriptors array
func (a JavaVersionDescriptors) MostRecent() (newest *JavaVersionDescriptor) {
	if a == nil {
		return nil
	}

	for _, desc := range a {
		if newest == nil {
			newest = desc
			continue
		}

		if desc.Version.Released.Time().After(newest.Version.Released.Time()) {
			newest = desc
		}
	}

	return
}

func (a JavaVersionDescriptors) ByVersion(version string) *JavaVersionDescriptor {
	if a == nil {
		return nil
	}

	for _, desc := range a {
		if desc.Version.Name == version {
			return desc
		}
	}

	return nil
}

// JavaVersionDescriptors acts like a normal array of Java versions, but provides helpful method MostRecent to retrieve
// the most recent version from the array (if more than one).
type JavaVersionDescriptors []*JavaVersionDescriptor

// JavaVersionsMap acts like a normal map of Java versions arrays, under the Mojang's identifiers.
type JavaVersionsMap map[string]JavaVersionDescriptors

// JavaRuntimesMap acts just like a normal map, fetched manifest from Mojang site is unmarshalled into this type.
// It contains maps of JRE versions mapped under the keys for different operating systems. For easy access method
// GetMatching method is provided, that will return needed map for current operating system.
type JavaRuntimesMap map[string]JavaVersionsMap

// GetMatching returns a map of versions matching the current operating system (if any), as well as the selector
// reported by GetJRESelector function for when there's a need in those (e.g. logging or debugging).
func (m JavaRuntimesMap) GetMatching() (versions JavaVersionsMap, selector string) {
	if m == nil {
		return
	}

	selector = GetJRESelector()

	if c, has := m[selector]; has {
		versions = c
	}

	return
}
