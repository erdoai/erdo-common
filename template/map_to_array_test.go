package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapToArray(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       any
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:     "Convert map to array",
			template: `{{mapToArray "idMap"}}`,
			stateParams: map[string]any{
				"idMap": map[string]any{
					"old_id_1": "new_id_1",
					"old_id_2": "new_id_2",
				},
			},
			expected: []map[string]any{
				{"key": "old_id_1", "value": "new_id_1"},
				{"key": "old_id_2", "value": "new_id_2"},
			},
		},
		{
			name:     "Empty map",
			template: `{{mapToArray "emptyMap"}}`,
			stateParams: map[string]any{
				"emptyMap": map[string]any{},
			},
			expected: []map[string]any{},
		},
		{
			name:        "Non-existent map",
			template:    `{{mapToArray "nonExistentMap"}}`,
			stateParams: map[string]any{},
			expected:    []map[string]any{},
		},
		{
			name:     "Map with different value types",
			template: `{{mapToArray "mixedMap"}}`,
			stateParams: map[string]any{
				"mixedMap": map[string]any{
					"string_key": "string_value",
					"int_key":    123,
					"bool_key":   true,
					"nil_key":    nil,
				},
			},
			expected: []map[string]any{
				{"key": "bool_key", "value": true},
				{"key": "int_key", "value": 123},
				{"key": "nil_key", "value": nil},
				{"key": "string_key", "value": "string_value"},
			},
		},
		{
			name:     "Nested map values",
			template: `{{mapToArray "nestedMap"}}`,
			stateParams: map[string]any{
				"nestedMap": map[string]any{
					"key1": map[string]any{"nested": "value1"},
					"key2": map[string]any{"nested": "value2"},
				},
			},
			expected: []map[string]any{
				{"key": "key1", "value": map[string]any{"nested": "value1"}},
				{"key": "key2", "value": map[string]any{"nested": "value2"}},
			},
		},
		{
			name:     "Non-map value",
			template: `{{mapToArray "notAMap"}}`,
			stateParams: map[string]any{
				"notAMap": "this is a string",
			},
			expected: []map[string]any{},
		},
		{
			name:     "Array instead of map",
			template: `{{mapToArray "arrayValue"}}`,
			stateParams: map[string]any{
				"arrayValue": []any{"item1", "item2"},
			},
			expected: []map[string]any{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := Hydrate(test.template, &test.stateParams, nil)
			
			if test.expectedError {
				assert.Error(t, err)
				if test.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), test.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
				
				// For array results, we need to compare without considering order
				// since Go map iteration is non-deterministic
				if expectedArray, ok := test.expected.([]map[string]any); ok {
					resultArray, ok := result.([]map[string]any)
					assert.True(t, ok, "Result should be []map[string]any")
					assert.Equal(t, len(expectedArray), len(resultArray), "Arrays should have same length")
					
					// For deterministic tests (empty or single element), compare directly
					if len(expectedArray) <= 1 {
						assert.Equal(t, test.expected, result)
					} else {
						// For multiple elements, check that all expected elements exist
						for _, expectedItem := range expectedArray {
							found := false
							for _, resultItem := range resultArray {
								if expectedItem["key"] == resultItem["key"] && 
								   assert.ObjectsAreEqual(expectedItem["value"], resultItem["value"]) {
									found = true
									break
								}
							}
							assert.True(t, found, "Expected item %v not found in result", expectedItem)
						}
					}
				} else {
					assert.Equal(t, test.expected, result)
				}
			}
		})
	}
}