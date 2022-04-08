package launcher

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/brawaru/marct/utils/slices"
	"github.com/stretchr/testify/assert"
)

//go:embed test_assets/version_mc1.18.1.json
var aBytes []byte

//go:embed test_assets/version_fabric0.12.12-1.18.1.json
var bBytes []byte

func TestMerge(t *testing.T) {
	var a Version
	var b Version

	if !assert.NoError(t, json.Unmarshal(aBytes, &a), "must parse a without errors") {
		return
	}

	if !assert.NoError(t, json.Unmarshal(bBytes, &b), "must parse b without errors") {
		return
	}

	merged, err := MergeVersions(a, b)

	if !assert.NoError(t, err, "must merge without errors") {
		return
	}

	if !assert.NotNil(t, slices.Find(merged.Libraries, func(item *Library, index int, slice []Library) bool {
		return item.Coordinates.GroupId == "net.fabricmc" && item.Coordinates.ArtifactId == "tiny-mappings-parser"
	}), "must find net:fabricmc:tiny-mapping-parser library") {
		return
	}

	if !assert.NotNil(t, slices.Find(merged.Libraries, func(item *Library, index int, slice []Library) bool {
		return item.Coordinates.GroupId == "com.mojang" && item.Coordinates.ArtifactId == "brigadier"
	}), "must find com.mojang:brigadier library") {
		return
	}

	// TODO: add more checks of correct merging
}
