package maven

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParsing(t *testing.T) {
	expectations := map[string]Coordinates{
		"com.example:artifact:1.0.0": {
			GroupId:      "com.example",
			ArtifactId:   "artifact",
			Version:      "1.0.0",
			VersionLabel: "",
			Packaging:    "jar",
			Classifier:   "",
		},
		"id.group:artifact-id:1.0.0-SNAPSHOT:ext:classifier": {
			GroupId:      "id.group",
			ArtifactId:   "artifact-id",
			Version:      "1.0.0",
			VersionLabel: "SNAPSHOT",
			Packaging:    "ext",
			Classifier:   "classifier",
		},
	}

	for coordinates, expectation := range expectations {
		res, err := NewCoordinates(coordinates)
		if !assert.NoErrorf(t, err, "must parse %s", coordinates) {
			return
		}

		assert.EqualValues(t, *res, expectation)
	}
}

func TestNames(t *testing.T) {
	type Expectation struct {
		FullName string
		BaseName string
		Path     string
	}

	expectations := map[string]Expectation{
		"id.group:artifact:1.0.0": {
			FullName: "artifact-1.0.0.jar",
			BaseName: "artifact-1.0.0",
			Path:     "id/group/artifact/1.0.0/artifact-1.0.0.jar",
		},
		"id.group:artifact-id:1.0.0-SNAPSHOT:ext:classifier": {
			FullName: "artifact-id-1.0.0-SNAPSHOT-classifier.ext",
			BaseName: "artifact-id-1.0.0-SNAPSHOT-classifier",
			Path:     "id/group/artifact-id/1.0.0-SNAPSHOT/artifact-id-1.0.0-SNAPSHOT-classifier.ext",
		},
	}

	for coordinates, expectation := range expectations {
		res, err := NewCoordinates(coordinates)
		if !assert.NoErrorf(t, err, "must parse %s", coordinates) {
			return
		}

		if !assert.Equal(t, res.FileName(), expectation.FullName, "full name must match") {
			return
		}

		if !assert.Equal(t, res.FileBaseName(), expectation.BaseName, "base name must match") {
			return
		}

		if !assert.Equal(t, res.Path('/'), expectation.Path, "path must match") {
			return
		}
	}
}

func TestConversionPersistence(t *testing.T) {
	expectations := map[string]string{
		"id.group:artifact:1.0.0":                            "",
		"id.group:artifact-id:1.0.0-SNAPSHOT:ext:classifier": "",
		"id.group:artifact:1.0.0:jar":                        "id.group:artifact:1.0.0",
		"id.group:artifact:1.0.0:jar:classifier":             "",
	}

	for coordinates, expectation := range expectations {
		res, err := NewCoordinates(coordinates)
		if !assert.NoErrorf(t, err, "must parse %s", coordinates) {
			return
		}

		if len(expectation) > 0 {
			if !assert.Equal(t, res.String(), expectation, "must meet expectation") {
				return
			}
		} else {
			if !assert.Equal(t, res.String(), coordinates, "must be the same as input") {
				return
			}
		}
	}
}

type JSONTestValue struct {
	A Coordinates `json:"a"`
	B Coordinates `json:"b"`
	C string      `json:"c"`
}

func TestJSON(t *testing.T) {

	input := `{
	"A": "id.group:artifact:1.0.0",
	"B": "id.group:artifact-id:1.0.0-SNAPSHOT:ext:classifier",
	"C": "id.group:artifact:1.0.0:jar:classifier"
}`

	expectation := JSONTestValue{
		A: Coordinates{
			GroupId:      "id.group",
			ArtifactId:   "artifact",
			Version:      "1.0.0",
			VersionLabel: "",
			Packaging:    "jar",
			Classifier:   "",
		},
		B: Coordinates{
			GroupId:      "id.group",
			ArtifactId:   "artifact-id",
			Version:      "1.0.0",
			VersionLabel: "SNAPSHOT",
			Packaging:    "ext",
			Classifier:   "classifier",
		},
		C: "id.group:artifact:1.0.0:jar:classifier",
	}

	var res JSONTestValue

	err := json.Unmarshal([]byte(input), &res)
	if !assert.NoError(t, err, "must parse JSON") {
		return
	}

	assert.Equal(t, res, expectation, "values must be equal")
}
