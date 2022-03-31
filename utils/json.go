package utils

import (
	"bytes"
	"encoding/json"
)

func StrictDecodeJSON(data []byte, out interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
}
