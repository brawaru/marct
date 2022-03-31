package launcher

import (
	_ "embed"
	"encoding/json"
	"github.com/brawaru/marct/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

//go:embed test_assets/version_mc1.18.1.json
var bytes1181 []byte

//go:embed test_assets/version_mc1.2.5.json
var bytes125 []byte

func Test1181(t *testing.T) {
	var res118 Version

	if err := utils.StrictDecodeJSON(bytes1181, &res118); !assert.NoError(t, err, "must parse 1.18 json") {
		return
	}

	assert.Equal(t, res118.ID, "1.18.1")
	assert.Equal(t, *res118.Type, "release")

	encoded, err := json.Marshal(res118)
	if !assert.NoError(t, err, "should encode 1.18 json") {
		return
	}

	jsonStr := string(encoded)
	assert.NotEmpty(t, jsonStr, "should not be empty")
}

func Test125(t *testing.T) {
	var res12 Version

	if err := utils.StrictDecodeJSON(bytes125, &res12); !assert.NoError(t, err, "must parse 1.2 json") {
		return
	}

	assert.Equal(t, res12.ID, "1.2.5", "must decode version")
	assert.Equal(t, *res12.Type, "release", "must decode release")

	encoded, err := json.Marshal(res12)
	if !assert.NoError(t, err, "should encode 1.18 json") {
		return
	}

	jsonStr := string(encoded)
	assert.NotEmpty(t, jsonStr, "should not be empty")
}
