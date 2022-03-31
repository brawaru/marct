package launcher

import (
	_ "embed"
	"github.com/brawaru/marct/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

//go:embed test_assets/assets_1.6.json
var assetsPre16 []byte

func TestParsing(t *testing.T) {
	var res AssetIndex

	if err := utils.StrictDecodeJSON(assetsPre16, &res); !assert.NoError(t, err, "must parse without error") {
		return
	}

	assert.Equal(t, *res.MapToResources, true, "mapToResources must be true")

	object, objectExists := res.Objects["READ_ME_I_AM_VERY_IMPORTANT"]
	if !assert.True(t, objectExists, "object READ_ME_I_AM_VERY_IMPORTANT must exist") {
		return
	}

	assert.Equal(t, object.URL(), "https://resources.download.minecraft.net/0d/0d000710b71ca9aafabd8f587768431d0b560b32")
}
