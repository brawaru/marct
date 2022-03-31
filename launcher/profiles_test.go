package launcher

import (
	_ "embed"
	"github.com/brawaru/marct/j2n"
	"github.com/stretchr/testify/assert"
	"testing"
)

//go:embed test_assets/profiles.json
var profilesJson []byte

func TestProfilesParsing(t *testing.T) {
	var res Profiles
	err := j2n.UnmarshalJSON(profilesJson, &res)

	if err != nil {
		t.Fatalf("parsing failure %v", err)
		return
	}

	if !assert.NotEmpty(t, res) {
		return
	}

	if !assert.Equal(t, *res.Version, 3) {
		return
	}

	// All profiles must be present and not empty
	if !assert.NotEmpty(t, res.Profiles["1bb37111a3abf6183eb0b749c08527cc"], "all profiles must be present") {
		return
	}
	if !assert.NotEmpty(t, res.Profiles["2c9c9b6f31b0a0567d70ef16f9e8b504"], "all profiles must be present") {
		return
	}
	if !assert.NotEmpty(t, res.Profiles["76e976312a066717e2d0c46d653d3b18"], "all profiles must be present") {
		return
	}
	if !assert.Empty(t, res.Profiles["non_existent"], "no unknown profiles") {
		return
	}

	if !assert.NotEmpty(t, res.Unknown) {
		return
	}

	if !assert.Equal(t, res.Profiles["76e976312a066717e2d0c46d653d3b18"].Name, "Classic", "strings should parse correctly") {
		return
	}

	if !assert.Equal(t, res.Profiles["76e976312a066717e2d0c46d653d3b18"].Created.Time().UnixMilli(), int64(1644108804059), "time must parse correctly") {
		return
	}

}
