package launcher

import (
	"github.com/brawaru/marct/j2n"
	"github.com/brawaru/marct/sdtypes"
)

type Profile struct {
	// Date and time when the profile is created
	//
	// MCL compatibility: MCL does not care if this field is absent or null
	Created *sdtypes.ISOTime `json:"created,omitempty"`
	// Icon used for the profile: either launcher asset name or data:// URI
	//
	// MCL compatibility: MCL does not care if this field is absent or null (uses Furnace instead)
	Icon *string `json:"icon,omitempty"`
	// Date and time when this profile was last used
	//
	// MCL compatibility: MCL will reset this field if absent or null
	LastUsed sdtypes.ISOTime `json:"lastUsed"`
	// Version used by this profile
	//
	// MCL compatibility: MCL will ignore and delete profile if this field is absent or null
	LastVersionID string `json:"lastVersionId,omitempty"`
	// Profile name
	//
	// MCL compatibility: if this field is empty, MCL will just display "<unnamed installation>"
	Name string `json:"name"`
	// Profile type
	//
	// MCL compatibility: MCL does not care if this field is absent or null. It, however, does care if this is set to
	// "latest-release" or "latest-snapshot" in which case it will prevent user from changing the version and icon,
	// which will be set to grass block or crafting table depending on the type. Attempting to launch such version
	// also results in launching the latest release/snapshot, not the version specified in profile.
	Type string `json:"type"`
	// Arguments passed to JVM when launching this profile
	JavaArgs *string `json:"javaArgs,omitempty"`
	// Path to JavaW executable
	//
	// By unknown reason this is called JavaDir in JSON, but it expects path to JavaW executable. If this field is
	// omitted, the bundled version will be used instead.
	JavaPath *string `json:"javaDir,omitempty"`
	// Minecraft's resolution (size of the window) when launching this profile
	Resolution *Resolution `json:"resolution,omitempty"`
	// Directory where game files like resource packs and mods are stored.
	GameDir string `json:"gameDir,omitempty"`
}

type Profiles struct {
	Profiles        map[string]Profile `json:"profiles"`                  // All profiles saved in launcher
	Version         *int               `json:"version"`                   // Version of the file
	SelectedProfile *string            `json:"selectedProfile,omitempty"` // Selected profile.
	Unknown         j2n.UnknownFields  `json:"-"`                         // Support fields from Minecraft Launcher
}
