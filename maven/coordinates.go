package maven

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	// Character responsible for separating parts of Maven coordinates
	CoordinatesSeparator = ":"
	// Standard packaging used by the Maven if none specified in coordinates
	DefaultPackaging = "jar"
	// Character responsible for separating the artefact name and the version
	FilenameSeparator = "-"
)

type Coordinates struct {
	GroupId      string
	ArtifactId   string
	Version      string
	VersionLabel string
	Packaging    string
	Classifier   string
}

func (c *Coordinates) FullVersion() string {
	if len(c.VersionLabel) == 0 {
		return c.Version
	}

	return strings.Join([]string{c.Version, FilenameSeparator, c.VersionLabel}, "")
}

func (c *Coordinates) FileBaseName() string {
	fileName := c.ArtifactId

	fileName += FilenameSeparator + c.FullVersion()

	if len(c.Classifier) > 0 {
		fileName += FilenameSeparator
		fileName += c.Classifier
	}

	return fileName
}

func (c *Coordinates) FileName() string {
	fileName := c.FileBaseName()

	if len(c.Packaging) > 0 {
		fileName += "." + c.Packaging
	}

	return fileName
}

// func (c Coordinates) URLPath() string {
// 	path := strings.Builder{}
//
// 	for _, dir := range strings.Split(c.GroupId, ".") {
// 		path.WriteString(url.PathEscape(dir))
// 		path.WriteRune('/')
// 	}
//
// 	path.WriteString(url.PathEscape(c.ArtifactId))
// 	path.WriteRune('/')
// 	path.WriteString(url.PathEscape(c.FullVersion()))
// 	path.WriteRune('/')
//
// 	path.WriteString(url.PathEscape(c.FileName()))
//
// 	return path.String()
// }

func (c *Coordinates) Path(sep rune) string {
	path := strings.Builder{}

	for _, dir := range strings.Split(c.GroupId, ".") {
		path.WriteString(dir)
		path.WriteRune(sep)
	}

	path.WriteString(c.ArtifactId)
	path.WriteRune(sep)
	path.WriteString(c.FullVersion())
	path.WriteRune(sep)

	path.WriteString(c.FileName())

	return path.String()
}

func (c *Coordinates) IsValid() error {
	if c == nil {
		return errors.New("is nil")
	}

	if len(c.ArtifactId) == 0 {
		return errors.New(".ArtifactId must not be empty")
	}

	if len(c.GroupId) == 0 {
		return errors.New(".GroupId must not be empty")
	}

	if len(c.Version) == 0 {
		return errors.New(".Version must not be empty")
	}

	if len(c.Packaging) == 0 {
		return errors.New(".Packaging must not be empty")
	}

	return nil
}

func (c *Coordinates) String() string {
	str := ""

	str += c.GroupId

	str += CoordinatesSeparator
	str += c.ArtifactId

	str += CoordinatesSeparator
	str += c.FullVersion()

	isClassifier := len(c.Classifier) > 0

	if isClassifier {
		str += CoordinatesSeparator
		str += c.Classifier
	}

	if c.Packaging != DefaultPackaging {
		str += CoordinatesSeparator
		str += c.Packaging
	}

	return str
}

func NewCoordinates(coordinates string) (*Coordinates, error) {
	parts := strings.Split(coordinates, CoordinatesSeparator)

	numOfParts := len(parts)

	if numOfParts < 3 {
		return nil, fmt.Errorf("coordinates need at least 3 parts, got only %v", numOfParts)
	}

	var coords Coordinates

	coords.GroupId = parts[0]
	coords.ArtifactId = parts[1]
	// coords.Version = parts[2]

	versionPart := parts[2]

	{
		splitIndex := strings.LastIndex(versionPart, FilenameSeparator)
		if splitIndex < 0 {
			coords.Version = versionPart
		} else {
			runes := []rune(versionPart)
			coords.Version = string(runes[0:splitIndex])
			coords.VersionLabel = string(runes[splitIndex+1:])
		}
	}

	if numOfParts > 3 {
		coords.Classifier = parts[3]
	}

	if numOfParts > 4 {
		coords.Packaging = parts[4]
	} else {
		coords.Packaging = DefaultPackaging
	}

	return &coords, nil
}

func (c *Coordinates) MarshalJSON() ([]byte, error) {
	if c == nil {
		return json.Marshal(nil)
	}

	// if err := c.IsValid(); err != nil {
	// 	return nil, err
	// }

	return json.Marshal(c.String())
}

func (c *Coordinates) UnmarshalJSON(data []byte) error {
	var str string

	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	coordinates, err := NewCoordinates(str)

	if err != nil {
		return err
	}

	*c = *coordinates

	return nil
}
