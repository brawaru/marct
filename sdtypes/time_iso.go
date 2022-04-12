package sdtypes

import (
	"encoding/json"
	"time"

	"github.com/relvacode/iso8601"
)

type ISOTime time.Time

func (t *ISOTime) UnmarshalJSON(v []byte) error {
	var value string
	if err := json.Unmarshal(v, &value); err != nil {
		return err
	}

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
