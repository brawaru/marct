package json_helpers

import (
	"encoding/json"
	"github.com/relvacode/iso8601"
	"strings"
	"time"
)

type ISOTime time.Time

func (t *ISOTime) UnmarshalJSON(v []byte) error {
	value := strings.Trim(string(v), "\"")
	parsedTime, err := iso8601.ParseString(value)

	if err != nil {
		return err
	}

	*t = ISOTime(parsedTime)

	return nil
}

func (t ISOTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time().Format(time.RFC3339))
}

func (t ISOTime) Time() time.Time {
	return time.Time(t)
}
