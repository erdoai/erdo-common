package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
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
	return ToJSONWithOptions(v, false)
}

// ToJSONWithOptions serializes a value to JSON with optional safety checks
// checkSafety: if true, validates RawMessage and checks for circular references (slower but safer)
func ToJSONWithOptions(v any, checkSafety bool) (*json.RawMessage, error) {
	if raw, ok := v.(json.RawMessage); ok {
		if checkSafety {
			// Quick validation - if it's invalid JSON, return an error
			if len(raw) > 0 {
				var temp any
				if err := json.Unmarshal(raw, &temp); err != nil {
					return nil, fmt.Errorf("invalid JSON in RawMessage: %w", err)
				}
			}
		}
		return &raw, nil
	}

	if checkSafety {
		// Check for circular references before marshaling (expensive reflection-based check)
		if FindCircularReferences(v) {
			return nil, fmt.Errorf("circular reference detected in value of type %T", v)
		}
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

// SafeJSON serializes to JSON with circular reference checking and validation
// This is slower but safer - use when you're unsure about the data structure
func SafeJSON(v any) json.RawMessage {
	raw, err := ToJSONWithOptions(v, true)
	if err != nil {
		panic(fmt.Sprintf("JSON serialization failed: %v", err))
	}
	return *raw
}

// StructToMap converts a Go struct to map[string]any using struct field names (not JSON tags).
// This is useful when you want the map keys to match what Go templates see via reflection.
//
// Key differences from JSON marshaling:
// - Uses struct field names (e.g., "Dataset") not JSON tags (e.g., "dataset")
// - Recursively processes nested structs, slices, and maps
// - Automatically dereferences pointers
// - Preserves nil pointers as nil (not omitted)
//
// Example:
//
//	type Resource struct {
//	    ID      int      `json:"id"`
//	    Dataset *Dataset `json:"dataset"`
//	}
//	// StructToMap returns: {"ID": 1, "Dataset": {...}}
//	// json.Marshal returns: {"id": 1, "dataset": {...}}
func StructToMap(v any) (any, error) {
	return structToMapReflect(reflect.ValueOf(v))
}

// shouldMarshalAsPrimitive returns true if the type should be JSON-marshaled
// directly rather than converted to a map. This is true for:
// - Arrays (e.g., uuid.UUID which is [16]byte)
// - Structs that have exported fields (assumed to be value types like time.Time)
func shouldMarshalAsPrimitive(val reflect.Value) bool {
	kind := val.Kind()

	// Arrays should always be marshaled (e.g., uuid.UUID is [16]byte)
	if kind == reflect.Array {
		return true
	}

	// For structs, check if it has any exported fields
	// If it has NO exported fields or all fields are unexported, it's likely a value type
	// that implements json.Marshaler (like time.Time)
	if kind == reflect.Struct {
		typ := val.Type()
		hasExportedField := false
		for i := 0; i < typ.NumField(); i++ {
			if typ.Field(i).IsExported() {
				hasExportedField = true
				break
			}
		}
		// If no exported fields, it's a value type (like time.Time) - marshal it
		if !hasExportedField {
			return true
		}
	}

	return false
}

func structToMapReflect(val reflect.Value) (any, error) {
	// Handle invalid values
	if !val.IsValid() {
		return nil, nil
	}

	// Dereference pointers
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, nil
		}
		val = val.Elem()
	}

	// Check if this is a type that should be JSON-marshaled as a primitive
	// (not converted to a map). This includes types like uuid.UUID, time.Time, etc.
	if val.CanInterface() && shouldMarshalAsPrimitive(val) {
		data, err := json.Marshal(val.Interface())
		if err != nil {
			return nil, fmt.Errorf("marshaling primitive type: %w", err)
		}
		// Unmarshal to get the native representation (string, number, etc.)
		var result any
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("unmarshaling primitive type result: %w", err)
		}
		return result, nil
	}

	switch val.Kind() {
	case reflect.Struct:
		result := make(map[string]any)
		typ := val.Type()

		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			fieldValue := val.Field(i)

			// Skip unexported fields
			if !field.IsExported() {
				continue
			}

			// Use the struct field name (not JSON tag)
			fieldName := field.Name

			// Recursively convert the field value
			convertedValue, err := structToMapReflect(fieldValue)
			if err != nil {
				return nil, fmt.Errorf("converting field %s: %w", fieldName, err)
			}

			result[fieldName] = convertedValue
		}
		return result, nil

	case reflect.Slice:
		if val.IsNil() {
			return nil, nil
		}

		result := make([]any, val.Len())
		for i := 0; i < val.Len(); i++ {
			convertedValue, err := structToMapReflect(val.Index(i))
			if err != nil {
				return nil, fmt.Errorf("converting slice element %d: %w", i, err)
			}
			result[i] = convertedValue
		}
		return result, nil

	case reflect.Array:
		// Arrays cannot be nil, so no nil check
		result := make([]any, val.Len())
		for i := 0; i < val.Len(); i++ {
			convertedValue, err := structToMapReflect(val.Index(i))
			if err != nil {
				return nil, fmt.Errorf("converting array element %d: %w", i, err)
			}
			result[i] = convertedValue
		}
		return result, nil

	case reflect.Map:
		if val.IsNil() {
			return nil, nil
		}

		result := make(map[string]any)
		iter := val.MapRange()
		for iter.Next() {
			key := iter.Key()
			value := iter.Value()

			// Convert key to string
			keyStr := fmt.Sprintf("%v", key.Interface())

			// Recursively convert the value
			convertedValue, err := structToMapReflect(value)
			if err != nil {
				return nil, fmt.Errorf("converting map value for key %s: %w", keyStr, err)
			}

			result[keyStr] = convertedValue
		}
		return result, nil

	case reflect.Interface:
		if val.IsNil() {
			return nil, nil
		}
		// Recurse on the concrete value
		return structToMapReflect(val.Elem())

	default:
		// For primitive types, just return the value as-is
		return val.Interface(), nil
	}
}
