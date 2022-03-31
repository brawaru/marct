package json_helpers

import (
	"encoding/json"
	"time"
)

type RFC3339Time time.Time

func (t *RFC3339Time) UnmarshalJSON(v []byte) error {
	var value string
	if err := json.Unmarshal(v, &value); err != nil {
		return err
	}

	parsedTime, err := time.Parse(time.RFC3339, value)

	if err != nil {
		return err
	}

	*t = RFC3339Time(parsedTime)

	return nil
}

func (t RFC3339Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t).Format(time.RFC3339))
}

func (t RFC3339Time) Time() time.Time {
	return time.Time(t)
}
