package java

import (
	"encoding/json"
	"fmt"
)

type FileType string

const (
	TypeDir  FileType = "directory"
	TypeFile FileType = "file"
	TypeLink FileType = "link"
)

func (t FileType) IsDir() bool {
	return t == TypeDir
}

func (t FileType) IsFile() bool {
	return t == TypeFile
}

func (t FileType) IsLink() bool {
	return t == TypeLink
}

func (t FileType) IsValid() error {
	switch t {
	case TypeFile:
		return nil
	case TypeDir:
		return nil
	case TypeLink:
		return nil
	default:
		return fmt.Errorf("%v is not valid value for java.FileType", t)
	}
}

func (t *FileType) UnmarshalJSON(data []byte) error {
	var s string

	if unmarshalErr := json.Unmarshal(data, &s); unmarshalErr != nil {
		return unmarshalErr
	}

	r := FileType(s)

	if valErr := r.IsValid(); valErr != nil {
		return valErr
	}

	*t = r

	return nil
}

func (t FileType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}
