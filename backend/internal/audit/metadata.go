package audit

import (
	"encoding/json"
	"fmt"
)

type Metadata map[string]any

func NewMetadata() Metadata {
	return make(Metadata)
}

func (m Metadata) Set(key string, value any) {
	m[key] = value
}

func (m Metadata) Marshal() ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidMetadata, err)
	}
	return data, nil
}

func (m Metadata) Validate() error {
	if m == nil {
		return nil
	}
	for k, v := range m {
		if k == "" {
			return fmt.Errorf("%w: empty key", ErrInvalidMetadata)
		}
		switch v.(type) {
		case string, bool, float64, nil, int, int64, float32:
		default:
			return fmt.Errorf("%w: unsupported value type for key %q", ErrInvalidMetadata, k)
		}
	}
	return nil
}
