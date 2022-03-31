package launcher

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResolutionParsingHeight(t *testing.T) {
	r, err := ParseResolution("1080p")
	if assert.NoError(t, err, "must not error when parsing") {
		return
	}
	if assert.Equal(t, r.Width, 0, "width must be 0") {
		return
	}
	if assert.Equal(t, r.Height, 1080, "height must be 1080") {
		return
	}
}

func TestResolutionParsingPartial(t *testing.T) {
	r, err := ParseResolution("1920")
	if assert.NoError(t, err, "must not error when parsing") {
		return
	}
	if assert.Equal(t, r.Width, 1920, "width must be 1920") {
		return
	}
	if assert.Equal(t, r.Height, 0, "height must be 0") {
		return
	}
}

func TestResolutionParsingFull(t *testing.T) {
	r, err := ParseResolution("1280x720")
	if assert.NoError(t, err, "must not error when parsing") {
		return
	}
	if assert.Equal(t, r.Width, 1280, "width must be 1280") {
		return
	}
	if assert.Equal(t, r.Height, 720, "height must be 720") {
		return
	}
}
