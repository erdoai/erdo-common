package utils

import (
	"encoding/json"
	"fmt"
)

func JSONToDict(j json.RawMessage) (map[string]any, error) {
	var d map[string]any
	err := json.Unmarshal(j, &d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func JSONToString(j json.RawMessage) (string, error) {
	var s any
	err := json.Unmarshal(j, &s)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%+v", s), nil
}

func ToJSON(v any) (*json.RawMessage, error) {
	if raw, ok := v.(json.RawMessage); ok {
		// Quick validation - if it's invalid JSON, return an error
		if len(raw) > 0 {
			var temp any
			if err := json.Unmarshal(raw, &temp); err != nil {
				return nil, fmt.Errorf("invalid JSON in RawMessage: %w", err)
			}
		}
		return &raw, nil
	}

	// Check for circular references before marshaling
	if FindCircularReferences(v) {
		return nil, fmt.Errorf("circular reference detected in value of type %T", v)
	}

	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	raw := json.RawMessage(data)
	return &raw, nil
}

func JSON(v any) json.RawMessage {
	raw, err := ToJSON(v)
	if err != nil {
		panic(fmt.Sprintf("JSON serialization failed: %v", err))
	}
	return *raw
}
