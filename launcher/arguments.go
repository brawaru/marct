package launcher

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
)

type Argument struct {
	Rules Rules    `mapstructure:"rules"`
	Value []string `mapstructure:"value"`
}

func (a *Argument) UnmarshalJSON(data []byte) error {
	var val interface{}

	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}

	switch t := val.(type) {
	case string:
		*a = Argument{
			Rules: []Rule{},
			Value: []string{val.(string)},
		}

	case map[string]interface{}:
		// I hate it here
		if flatValue, isFlatValue := t["value"].(string); isFlatValue {
			t["value"] = []string{flatValue}
		}

		if err := mapstructure.Decode(t, &a); err != nil {
			return err
		}

	default:
		return fmt.Errorf("illegal value - %v", t)
	}

	return nil
}

func (a Argument) MarshalJSON() ([]byte, error) {
	if len(a.Rules) == 0 && len(a.Value) == 1 {
		return json.Marshal(a.Value[0])
	} else {
		var mapped map[string]interface{}

		if err := mapstructure.Decode(a, &mapped); err != nil {
			return nil, err
		}

		// special flattening case for moM
		// why do they even need that, it literally saves 2 BYTES!!!
		if value, ok := mapped["value"].([]string); ok {
			if len(value) == 1 {
				mapped["value"] = value[0]
			}
		}

		return json.Marshal(mapped)
	}
}

type Arguments struct {
	Game []Argument `json:"game"`
	JVM  []Argument `json:"jvm"`
}
