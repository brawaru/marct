package launcher

import (
	"encoding/json"
	"fmt"
	"os"
)

func unmarshalJSONFile[T any](name string, v **T) error {
	file, err := os.ReadFile(name)

	if err != nil {
		return fmt.Errorf("read file %s: %v", name, err)
	}

	if err := json.Unmarshal(file, v); err != nil {
		return fmt.Errorf("unmarshal file %s: %v", name, err)
	}

	return nil
}
