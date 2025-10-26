package template

import (
	"fmt"
	"sort"
	"testing"
	"time"

	common "github.com/erdoai/erdo-common/types"
	"github.com/stretchr/testify/assert"
)

func TestHydrateString(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       string
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:     "Simple variable substitution",
			template: "Hello, {{name}}!",
			stateParams: map[string]any{
				"name": "World",
			},
			expected: "Hello, World!",
		},
		{
			name:     "Conditional rendering",
			template: "{{if .Data.test}}present{{end}}",
			stateParams: map[string]any{
				"test": true,
			},
			expected: "present",
		},
		{
			name:     "Conditional rendering - with additional text",
			template: "{{if (truthy \"test\" .Data)}}present{{end}} other",
			stateParams: map[string]any{
				"test": true,
			},
			expected: "present other",
		},
		{
			name:     "Conditional rendering - false condition",
			template: "{{if (truthy \"test\" .Data)}}present{{end}}",
			stateParams: map[string]any{
				"test": false,
			},
			expected: "",
		},
		{
			name:     "Conditional rendering - missing condition",
			template: "{{if (truthy \"test\" .Data)}}present{{end}}",
			stateParams: map[string]any{
				"other": false,
			},
			expected: "",
		},
		{
			name:     "Conditional rendering - missing condition with additional text",
			template: "{{if (truthy \"test\" .Data)}}present{{end}} other",
			stateParams: map[string]any{
				"other": false,
			},
			expected: " other",
		},
		{
			name:     "Conditional rendering - string value present",
			template: "{{if (truthy \"test\" .Data)}}present{{end}}",
			stateParams: map[string]any{
				"test": "yep",
			},
			expected: "present",
		},
		{
			name:     "Conditional rendering - string value empty",
			template: "{{if (truthy \"test\" .Data)}}present{{end}}",
			stateParams: map[string]any{
				"test": "",
			},
			expected: "",
		},
		{
			name: "Variable setting",
			template: `{{- $hasDerived := false -}}
{{- range $r := .Data.resources -}}
    {{- if eq $r.created_by "bot" -}}
        {{- $hasDerived = true -}}
    {{- end -}}
{{- end -}}
{{- if $hasDerived -}}
derived
{{- end }} other`,
			stateParams: map[string]any{
				"resources": []map[string]any{
					{
						"created_by": "bot",
					},
				},
			},
			expected: "derived other",
		},
		{
			name: "Variable setting - false",
			template: `{{- $hasDerived := false -}}
{{- range $r := .Data.resources -}}
    {{- if eq $r.created_by "bot" -}}
        {{- $hasDerived = true -}}
    {{- end -}}
{{- end -}}
{{- if $hasDerived -}}
derived
{{- end }} other`,
			stateParams: map[string]any{
				"resources": []map[string]any{
					{
						"created_by": "user",
					},
				},
			},
			expected: " other",
		},
		{
			name:           "Missing variable",
			template:       "Hello, {{name}}!",
			stateParams:    map[string]any{},
			expectedError:  true,
			expectedErrMsg: "info needed for keys [name]",
		},
		{
			name:     "Mix of required and optional parameters",
			template: "Required: {{required}}, Optional: {{optional?}}",
			stateParams: map[string]any{
				"required": "value",
				// optional is intentionally missing
			},
			expected: "Required: value, Optional: ",
		},
		{
			name:     "Multiple optional parameters, some missing",
			template: "Required: {{required}}, Optional1: {{optional1?}}, Optional2: {{optional2?}}",
			stateParams: map[string]any{
				"required":  "value",
				"optional1": "opt1",
				// optional2 is intentionally missing
			},
			expected: "Required: value, Optional1: opt1, Optional2: ",
		},
		{
			name:        "Noop function for whitespace removal",
			template:    "{{- noop}}Start{{- noop}} middle {{- noop}}end",
			stateParams: map[string]any{},
			expected:    "Start middleend",
		},
		{
			name:     "Noop function in JSON-like template",
			template: `{"title": "{{title}}",{{- noop}}"description": "{{description}}"}`,
			stateParams: map[string]any{
				"title":       "Test Title",
				"description": "Test Description",
			},
			expected: `{"title": "Test Title","description": "Test Description"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HydrateString(tt.template, &tt.stateParams)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestHydrateDict(t *testing.T) {
	tests := []struct {
		name           string
		template       map[string]any
		stateParams    map[string]any
		expected       map[string]any
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:        "Simple dict hydration",
			template:    map[string]any{"greeting": "Hello, {{name}}!"},
			stateParams: map[string]any{"name": "World"},
			expected:    map[string]any{"greeting": "Hello, World!"},
		},
		{
			name:        "Nested dict hydration",
			template:    map[string]any{"user": map[string]any{"name": "{{name}}", "age": "{{age}}"}},
			stateParams: map[string]any{"name": "Alice", "age": 30},
			expected:    map[string]any{"user": map[string]any{"name": "Alice", "age": 30}},
		},
		{
			name:           "Missing variable",
			template:       map[string]any{"greeting": "Hello, {{name}}!"},
			stateParams:    map[string]any{},
			expectedError:  true,
			expectedErrMsg: "info needed for keys [greeting.name]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HydrateDict(tt.template, &tt.stateParams)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestHydrateSlice(t *testing.T) {
	tests := []struct {
		name                       string
		template                   []any
		stateParams                map[string]any
		parameterHydrationBehavior *map[string]any
		expected                   []any
		expectedError              bool
		expectedErrMsg             string
		expectedPanic              bool
	}{
		{
			name:        "Simple slice hydration",
			template:    []any{"Hello, {{name}}!", "{{greeting}}"},
			stateParams: map[string]any{"name": "World", "greeting": "Hi"},
			expected:    []any{"Hello, World!", "Hi"},
		},
		{
			name:        "Slice with nested objects",
			template:    []any{map[string]any{"name": "{{name}}"}, map[string]any{"age": "{{age}}"}},
			stateParams: map[string]any{"name": "Alice", "age": 30},
			expected:    []any{map[string]any{"name": "Alice"}, map[string]any{"age": 30}},
		},
		{
			name:           "Missing variable",
			template:       []any{"Hello, {{name}}!"},
			stateParams:    map[string]any{},
			expectedError:  true,
			expectedErrMsg: "info needed for keys [[0].name]",
		},
		{
			name: "Slice with nested objects and parameterHydrationBehavior",
			template: []any{
				map[string]any{
					"name": "{{name}}",
					"raw":  "{{name}}", // Same var but will remain raw
				},
				map[string]any{
					"age":    "{{age}}",
					"city":   "{{city}}",
					"hidden": "{{hidden}}",
				},
			},
			stateParams: map[string]any{
				"name":   "Alice",
				"age":    30,
				"city":   "New York",
				"hidden": "secret",
			},
			parameterHydrationBehavior: &map[string]any{
				"raw":    common.ParameterHydrationBehaviourRaw, // Applies to all elements
				"hidden": common.ParameterHydrationBehaviourRaw, // Applies to all elements
			},
			expected: []any{
				map[string]any{
					"name": "Alice",
					"raw":  "{{name}}", // Remains unhydrated
				},
				map[string]any{
					"age":    30,
					"city":   "New York",
					"hidden": "{{hidden}}", // Remains unhydrated
				},
			},
		},
		{
			name: "Correctly preventing hydration in nested slice element fields",
			template: []any{
				map[string]any{
					"name": "{{name}}",
					"details": []any{
						map[string]any{
							"normal": "{{normalParam}}",
							"raw":    "{{rawParam}}",
						},
					},
				},
			},
			stateParams: map[string]any{
				"name":        "Alice",
				"normalParam": "Normal Param",
				"rawParam":    "Raw Param",
			},
			// To prevent hydration of fields inside nested objects in slice elements,
			// you must specify the full path to those fields
			parameterHydrationBehavior: &map[string]any{
				"details": map[string]any{
					// This Dict applies to the "details" field, which is a slice
					// The Dict will be passed down to all elements in that slice
					"raw": common.ParameterHydrationBehaviourRaw, // Applied to "raw" key in all elements of the details slice
				},
			},
			expected: []any{
				map[string]any{
					"name": "Alice",
					"details": []any{
						map[string]any{
							"normal": "Normal Param",
							"raw":    "{{rawParam}}", // Correctly remains unhydrated due to proper path specification
						},
					},
				},
			},
		},
		{
			name: "Correctly passing parameterHydrationBehaviour to slice elements",
			template: []any{
				map[string]any{
					"name": "{{name}}",
					"details": []any{
						map[string]any{
							"normal": "{{normalParam}}",
							"raw":    "{{rawParam}}", // This will be hydrated because details is processed as a Dict, not as part of the slice behavior
						},
					},
				},
			},
			stateParams: map[string]any{
				"name":        "Alice",
				"normalParam": "Normal Param",
				"rawParam":    "Raw Param",
			},
			// To prevent hydration of the raw field in the details array elements,
			// we need to specify the correct path structure
			parameterHydrationBehavior: &map[string]any{
				"details": map[string]any{
					"raw": common.ParameterHydrationBehaviourRaw, // This correctly targets the 'raw' field in all elements of the details array
				},
			},
			expected: []any{
				map[string]any{
					"name": "Alice",
					"details": []any{
						map[string]any{
							"normal": "Normal Param",
							"raw":    "{{rawParam}}", // Now it correctly remains unhydrated with the proper parameterHydrationBehaviour path
						},
					},
				},
			},
		},
		{
			name: "Slice with optional parameters and parameterHydrationBehavior",
			template: []any{
				map[string]any{
					"name":       "{{name}}",
					"middleName": "{{middleName?}}",
					"raw":        "{{rawParam}}",
				},
			},
			stateParams: map[string]any{
				"name":     "John",
				"rawParam": "Should Stay Raw",
				// middleName intentionally missing
			},
			parameterHydrationBehavior: &map[string]any{
				"raw": common.ParameterHydrationBehaviourRaw,
			},
			expected: []any{
				map[string]any{
					"name":       "John",
					"middleName": nil,            // Optional parameter becomes nil
					"raw":        "{{rawParam}}", // Remains unhydrated
				},
			},
		},
		{
			name: "Complex nested structure with specific path hydration behavior",
			template: []any{
				map[string]any{
					"name": "{{name}}",
					"nested": map[string]any{
						"normalValue": "{{nestedNormal}}",
						"rawValue":    "{{nestedRaw}}",
					},
				},
				map[string]any{
					"details": []any{
						map[string]any{
							"normal": "{{normalParam}}",
							"raw":    "{{rawParam}}",
						},
					},
				},
			},
			stateParams: map[string]any{
				"name":         "Alice",
				"nestedNormal": "Normal Value",
				"nestedRaw":    "Raw Value",
				"normalParam":  "Normal Param",
				"rawParam":     "Raw Param",
			},
			parameterHydrationBehavior: &map[string]any{
				"nested": map[string]any{
					"rawValue": common.ParameterHydrationBehaviourRaw, // Only this specific path is configured not to be hydrated
				},
				"raw": common.ParameterHydrationBehaviourRaw, // This only applies to top-level keys named "raw", not those in nested objects
			},
			expected: []any{
				map[string]any{
					"name": "Alice",
					"nested": map[string]any{
						"normalValue": "Normal Value",
						"rawValue":    "{{nestedRaw}}", // Remains unhydrated due to specific path in parameterHydrationBehavior
					},
				},
				map[string]any{
					"details": []any{
						map[string]any{
							"normal": "Normal Param",
							"raw":    "Raw Param", // Gets hydrated because the raw:ParameterHydrationBehaviourRaw only applies to top-level keys
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedPanic {
				defer func() {
					r := recover()
					assert.NotNil(t, r, "Expected panic didn't occur")
				}()
			}

			var result []any
			var err error

			if tt.parameterHydrationBehavior != nil {
				result, err = HydrateSlice(tt.template, &tt.stateParams, tt.parameterHydrationBehavior)
			} else {
				result, err = HydrateSlice(tt.template, &tt.stateParams)
			}

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else if !tt.expectedPanic {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestComplexNestedFunctionInIf(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       string
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:     "Nested function calls in if statement with coalesce and sub",
			template: `{{if lt (coalesce "tool_usage_loops?" 0) (sub (get "max_tool_calls") 1) -}}Tool calls available{{- end}}`,
			stateParams: map[string]any{
				"max_tool_calls": 5,
			},
			expected: "Tool calls available",
		},
		{
			name:     "Nested function calls in if statement - condition false",
			template: `{{if lt (coalesce "tool_usage_loops?" 0) (sub (get "max_tool_calls") 1) -}}Tool calls available{{- end}}`,
			stateParams: map[string]any{
				"tool_usage_loops": 10,
				"max_tool_calls":   5,
			},
			expected: "",
		},
		{
			name:     "Multiple nested function calls",
			template: `{{if and (gt (coalesce "counter?" 0) 2) (lt (coalesce "tool_usage_loops?" 0) (sub (get "max_tool_calls") 1)) -}}Both conditions met{{- end}}`,
			stateParams: map[string]any{
				"counter":        3,
				"max_tool_calls": 10,
			},
			expected: "Both conditions met",
		},
		{
			name:     "Nested coalesce in lt function - should not duplicate params",
			template: `{{lt (coalesce "tool_usage_loops?" 0) (sub (get "max_tool_calls") 1)}}`,
			stateParams: map[string]any{
				"max_tool_calls": 5,
			},
			expected: "true",
		},
		{
			name:        "Complex nested function with missing key - should handle gracefully",
			template:    `{{if ge (coalesce "tool_usage_loops?" 0) (sub (coalesce "max_tool_calls" 5) 1)}}true{{else}}false{{end}}`,
			stateParams: map[string]any{
				// max_tool_calls is missing - this should use default fallback of 5
			},
			expectedError: false,
			expected:      "false", // Should evaluate to false: 0 >= (5-1) = 0 >= 4 = false
		},
		{
			name:     "Complex nested function with present key - should work correctly",
			template: `{{if ge (coalesce "tool_usage_loops?" 0) (sub (coalesce "max_tool_calls" 5) 1)}}true{{else}}false{{end}}`,
			stateParams: map[string]any{
				"max_tool_calls": 5,
			},
			expected: "false", // 0 >= (5-1) = 0 >= 4 = false
		},
		{
			name:     "Complex nested function with both keys present",
			template: `{{if ge (coalesce "tool_usage_loops?" 0) (sub (coalesce "max_tool_calls" 5) 1)}}true{{else}}false{{end}}`,
			stateParams: map[string]any{
				"tool_usage_loops": 4,
				"max_tool_calls":   5,
			},
			expected: "true", // 4 >= (5-1) = 4 >= 4 = true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HydrateString(tt.template, &tt.stateParams)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFindTemplateKeys(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		includeOptional bool
		expected        []Key
	}{
		{
			name:            "Simple variable",
			input:           "Hello, {{name}}!",
			includeOptional: false,
			expected:        []Key{{Key: "name", IsOptional: false}},
		},
		{
			name:            "Optional variable",
			input:           "Hello, {{name?}}!",
			includeOptional: true,
			expected:        []Key{{Key: "name", IsOptional: true}},
		},
		{
			name:            "Multiple variables",
			input:           "{{greeting}}, {{name}}! {{farewell?}}",
			includeOptional: true,
			expected: []Key{
				{Key: "greeting", IsOptional: false},
				{Key: "name", IsOptional: false},
				{Key: "farewell", IsOptional: true},
			},
		},
		// Add more test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindTemplateKeysToHydrate(tt.input, tt.includeOptional, nil)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]any
		expected string
	}{
		{
			name:     "Simple variable",
			template: "Hello, {{name}}!",
			data:     map[string]any{"name": "World"},
			expected: "Hello, World!",
		},
		{
			name:     "Conditional rendering",
			template: "{{if .Data.show}}Visible{{else}}Hidden{{end}}",
			data:     map[string]any{"show": true},
			expected: "Visible",
		},
		{
			name:     "Nested object access",
			template: "{{.Data.user.name}} is {{.Data.user.age}} years old",
			data: map[string]any{
				"user": map[string]any{
					"name": "Alice",
					"age":  30,
				},
			},
			expected: "Alice is 30 years old",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTemplate(tt.template)
			assert.NoError(t, err)
			hydrated, err := HydrateString(result, &tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, hydrated)
		})
	}
}

func TestMapToDict(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       any
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:     "Convert string list to dict list",
			template: "{{mapToDict \"stringList\" \"key\"}}",
			stateParams: map[string]any{
				"stringList": []any{"value1", "value2", "value3"},
			},
			expected: []map[string]any{
				{"key": "value1"},
				{"key": "value2"},
				{"key": "value3"},
			},
		},
		{
			name:     "Empty list",
			template: "{{mapToDict \"emptyList\" \"key\"}}",
			stateParams: map[string]any{
				"emptyList": []any{},
			},
			expected: []map[string]any{},
		},
		{
			name:        "Non-existent list",
			template:    "{{mapToDict \"nonExistentList\" \"key\"}}",
			stateParams: map[string]any{},
			expected:    []map[string]any{},
		},
		{
			name:     "List with mixed types",
			template: "{{mapToDict \"mixedList\" \"key\"}}",
			stateParams: map[string]any{
				"mixedList": []any{"string", 123, true, nil},
			},
			expected: []map[string]any{
				{"key": "string"},
				{"key": 123}, // No longer converts to float64 since we removed JSON round-trip
				{"key": true},
				{"key": nil},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compare objects directly using Hydrate instead of using HydrateString with toJSON.
			// This avoids unnecessary JSON serialization/deserialization and makes test assertions more robust.
			result, err := Hydrate(tt.template, &tt.stateParams, nil)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestNestedTemplateFunctions(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       any
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:     "mapToDict nested function",
			template: "{{mapToDict \"stringList\" \"key\"}}",
			stateParams: map[string]any{
				"stringList": []any{"value1", "value2"},
			},
			expected: []map[string]any{
				{"key": "value1"},
				{"key": "value2"},
			},
		},
		{
			name:     "With Data variables",
			template: "{{get \"stringList.0\"}}",
			stateParams: map[string]any{
				"stringList": []any{"value1", "value2"},
			},
			expected: "value1",
		},
		{
			name:     "Non-function template",
			template: "Hello, {{name}}!",
			stateParams: map[string]any{
				"name": "World",
			},
			expected: "Hello, World!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compare objects directly using Hydrate instead of using HydrateString with toJSON.
			// This avoids unnecessary JSON serialization/deserialization and makes test assertions more robust.
			result, err := Hydrate(tt.template, &tt.stateParams, nil)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestAddkey(t *testing.T) {
	tests := []struct {
		name        string
		toObj       string
		key         string
		valueKey    string
		data        map[string]any
		missingKeys *[]string
		expected    map[string]any
	}{
		{
			name:        "Add key to existing object",
			toObj:       "object",
			key:         "newKey",
			valueKey:    "value",
			data:        map[string]any{"object": map[string]any{"existingKey": "existingValue"}, "value": "newValue"},
			missingKeys: &[]string{},
			expected:    map[string]any{"existingKey": "existingValue", "newKey": "newValue"},
		},
		{
			name:        "Add key to empty object",
			toObj:       "emptyObject",
			key:         "firstKey",
			valueKey:    "value",
			data:        map[string]any{"emptyObject": map[string]any{}, "value": "someValue"},
			missingKeys: &[]string{},
			expected:    map[string]any{"firstKey": "someValue"},
		},
		{
			name:        "Overwrite existing key",
			toObj:       "object",
			key:         "existingKey",
			valueKey:    "newValue",
			data:        map[string]any{"object": map[string]any{"existingKey": "oldValue"}, "newValue": "updatedValue"},
			missingKeys: &[]string{},
			expected:    map[string]any{"existingKey": "updatedValue"},
		},
		{
			name:        "Add nested value",
			toObj:       "object",
			key:         "nested",
			valueKey:    "nestedValue",
			data:        map[string]any{"object": map[string]any{}, "nestedValue": map[string]any{"a": 1, "b": 2}},
			missingKeys: &[]string{},
			expected:    map[string]any{"nested": map[string]any{"a": 1, "b": 2}},
		},
		{
			name:        "Object not found",
			toObj:       "nonExistentObject",
			key:         "key",
			valueKey:    "value",
			data:        map[string]any{"value": "someValue"},
			missingKeys: &[]string{},
			expected:    nil, // Should return nil since object doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Look up the value from data using valueKey
			value := get(tt.valueKey, tt.data, tt.missingKeys)
			result := addkey(tt.toObj, tt.key, value, tt.data, tt.missingKeys)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.Equal(t, tt.expected, result)

			// For the case where we're missing a key, check that it was added to missingKeys
			if tt.name == "Object not found" {
				assert.Contains(t, *tt.missingKeys, tt.toObj)
			}
		})
	}
}

// Test addkey function used inside template

func TestDedupeBy(t *testing.T) {
	t.Parallel()

	// Create test data
	testData := map[string]any{
		"simpleItems": []any{
			map[string]any{"id": "1", "name": "Item 1"},
			map[string]any{"id": "2", "name": "Item 2"},
			map[string]any{"id": "1", "name": "Item 1 Duplicate"},
			map[string]any{"id": "3", "name": "Item 3"},
		},
		"complexItems": []any{
			map[string]any{
				"ID":      "1",
				"Content": "First item",
				"CreatedAt": map[string]any{
					"Time":  time.Now().Format(time.RFC3339),
					"Valid": true,
				},
			},
			map[string]any{
				"ID":      "2",
				"Content": "Second item",
				"CreatedAt": map[string]any{
					"Time":  time.Now().Format(time.RFC3339),
					"Valid": true,
				},
			},
			map[string]any{
				"ID":      "1", // Duplicate ID
				"Content": "First item duplicate",
				"CreatedAt": map[string]any{
					"Time":  time.Now().Format(time.RFC3339),
					"Valid": true,
				},
			},
			map[string]any{
				"ID":      "3",
				"Content": "Third item",
				"CreatedAt": map[string]any{
					"Time":  time.Now().Format(time.RFC3339),
					"Valid": true,
				},
			},
		},
		"nestedItems": []any{
			map[string]any{
				"metadata": map[string]any{
					"id":   "A",
					"type": "first",
				},
				"content": "Content A",
			},
			map[string]any{
				"metadata": map[string]any{
					"id":   "B",
					"type": "second",
				},
				"content": "Content B",
			},
			map[string]any{
				"metadata": map[string]any{
					"id":   "A", // Duplicate nested ID
					"type": "third",
				},
				"content": "Content A duplicate",
			},
		},
	}

	// Test cases
	testCases := []struct {
		name          string
		arrayKey      string
		fieldKey      string
		expectedCount int
	}{
		{
			name:          "Simple deduplication by ID",
			arrayKey:      "simpleItems",
			fieldKey:      "id",
			expectedCount: 3, // 3 unique IDs (1, 2, 3)
		},
		{
			name:          "Complex object deduplication by ID",
			arrayKey:      "complexItems",
			fieldKey:      "ID",
			expectedCount: 3, // 3 unique IDs (1, 2, 3)
		},
		{
			name:          "Nested field deduplication",
			arrayKey:      "nestedItems",
			fieldKey:      "metadata.id",
			expectedCount: 2, // 2 unique IDs (A, B)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var missingKeys []string
			result := dedupeBy(tc.arrayKey, tc.fieldKey, testData, &missingKeys)

			assert.Equal(t, tc.expectedCount, len(result), "Expected %d items after deduplication", tc.expectedCount)

			// Verify no duplicates exist in result
			seen := make(map[string]bool)
			for _, item := range result {
				itemDict, ok := item.(map[string]any)
				if !ok {
					t.Fatalf("Expected Dict item, got %T", item)
				}

				// Extract the field value, handling nested fields
				var fieldValue any
				if tc.name == "Nested field deduplication" {
					metadata, ok := itemDict["metadata"].(map[string]any)
					if !ok {
						t.Fatalf("Expected metadata to be Dict, got %T", itemDict["metadata"])
					}
					fieldValue = metadata["id"]
				} else {
					fieldValue = itemDict[tc.fieldKey]
				}

				valueStr := toString(fieldValue)
				assert.False(t, seen[valueStr], "Found duplicate ID %s after deduplication", valueStr)
				seen[valueStr] = true
			}
		})
	}
}

func TestSingleVariableNoDoubleHydration(t *testing.T) {
	// This test verifies:
	// 1. Single variables with content that contains templates don't get double-hydrated
	// 2. Python-style formatting still works correctly

	// First level of parameters that will be hydrated
	stateParams := map[string]any{
		"outer_var":    "I am the outer variable with {{inner_var}}",
		"inner_var":    "this is inner content",
		"postgres_var": "I use %(name)s parameter",
		"name":         "Python",
		"steps": map[string]any{
			"code": map[string]any{
				"code": `This has a template {{variable}} that shouldn't be hydrated.`,
			},
		},
		"variable": "SHOULD NOT BE HYDRATED",
		"format":   "SHOULD BE HYDRATED",
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "Direct access to single variable with templates shouldn't double hydrate",
			template: "something {{outer_var}}",
			expected: "something I am the outer variable with {{inner_var}}",
		},
		{
			name:     "Postgres parameters should not double hydrate",
			template: "{{postgres_var}}",
			expected: "I use %(name)s parameter",
		},
		{
			name:     "Complex nested variable shouldn't double hydrate templates",
			template: "{{steps.code.code}}",
			expected: `This has a template {{variable}} that shouldn't be hydrated.`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HydrateString(tt.template, &stateParams)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParameterHydrationBehaviour(t *testing.T) {
	tests := []struct {
		name                       string
		template                   map[string]any
		stateParams                map[string]any
		parameterHydrationBehavior map[string]any
		expected                   map[string]any
	}{
		{
			name: "Basic hydration behavior - hydrate all",
			template: map[string]any{
				"hydrated":   "Value with {{param}}",
				"unmodified": "Value with {{param}}",
			},
			stateParams: map[string]any{
				"param": "test",
			},
			parameterHydrationBehavior: map[string]any{},
			expected: map[string]any{
				"hydrated":   "Value with test",
				"unmodified": "Value with test",
			},
		},
		{
			name: "Skip parameter hydration using direct raw value",
			template: map[string]any{
				"hydrated": "Value with {{param}}",
				"tools":    "This contains {{param}} value",
			},
			stateParams: map[string]any{
				"param": "test",
			},
			parameterHydrationBehavior: map[string]any{
				"tools": common.ParameterHydrationBehaviourRaw,
			},
			expected: map[string]any{
				"hydrated": "Value with test",
				"tools":    "This contains {{param}} value", // Should remain as template string
			},
		},
		{
			name: "Skip parameter hydration for nested dict",
			template: map[string]any{
				"hydrated": "Value with {{param}}",
				"tools": map[string]any{
					"parameters": map[string]any{
						"param1": "{{param}}",
						"param2": "static",
					},
				},
			},
			stateParams: map[string]any{
				"param": "test",
			},
			parameterHydrationBehavior: map[string]any{
				"tools": map[string]any{
					"parameters": common.ParameterHydrationBehaviourRaw,
				},
			},
			expected: map[string]any{
				"hydrated": "Value with test",
				"tools": map[string]any{
					"parameters": map[string]any{
						"param1": "{{param}}", // Should remain as template string
						"param2": "static",
					},
				},
			},
		},
		{
			name: "Skip parameter hydration for nested dict in slice",
			template: map[string]any{
				"hydrated": "Value with {{param}}",
				"tools": []map[string]any{
					{
						"parameters": map[string]any{
							"raw":     "{{paramDoesNotExist}}",
							"hydrate": "{{param}}",
						},
					},
				},
			},
			stateParams: map[string]any{
				"param": "test",
			},
			parameterHydrationBehavior: map[string]any{
				"tools": map[string]any{
					"parameters": map[string]any{
						"raw": common.ParameterHydrationBehaviourRaw,
					},
				},
			},
			expected: map[string]any{
				"hydrated": "Value with test",
				"tools": []map[string]any{
					{
						"parameters": map[string]any{
							"raw":     "{{paramDoesNotExist}}", // Should remain as template string
							"hydrate": "test",
						},
					},
				},
			},
		},
		{
			name: "Skip parameter hydration for nested dict in slice with optional values",
			template: map[string]any{
				"hydrated": "Value with {{param}}",
				"tools": []map[string]any{
					{
						"parameters": map[string]any{
							"param1": "{{param?}}",
							"param2": "static",
						},
					},
				},
			},
			stateParams: map[string]any{
				"param": "test",
			},
			parameterHydrationBehavior: map[string]any{
				"tools": map[string]any{
					"parameters": common.ParameterHydrationBehaviourRaw,
				},
			},
			expected: map[string]any{
				"hydrated": "Value with test",
				"tools": []map[string]any{
					{
						"parameters": map[string]any{
							"param1": "{{param?}}", // Should remain as template string
							"param2": "static",
						},
					},
				},
			},
		},
		{
			name: "Skip parameter hydration for nested value, but leaves other values alone",
			template: map[string]any{
				"hydrated": "Value with {{param}}",
				"tools": map[string]any{
					"should_hydrate": "{{param}}",
					"should_leave":   "{{param}}",
				},
			},
			stateParams: map[string]any{
				"param": "test",
			},
			parameterHydrationBehavior: map[string]any{
				"tools": map[string]any{
					"should_leave": common.ParameterHydrationBehaviourRaw,
				},
			},
			expected: map[string]any{
				"hydrated": "Value with test",
				"tools": map[string]any{
					"should_hydrate": "test",
					"should_leave":   "{{param}}",
				},
			},
		},
		{
			name: "Bots.go example - tools > parameters setting applies to parameter key in tool item",
			template: map[string]any{
				"system_prompt": "This is a prompt with {{prompt_var}}",
				"tools": []map[string]any{
					{
						"name":        "run_analysis",
						"description": "Run an analysis with {{description_var}}",
						"parameters": map[string]any{
							"param1": "{{param}}",
							"param2": "static",
						},
					},
				},
			},
			stateParams: map[string]any{
				"prompt_var":      "test_prompt",
				"description_var": "test_description",
				"param":           "test_param",
			},
			parameterHydrationBehavior: map[string]any{
				"tools": map[string]any{
					"parameters": common.ParameterHydrationBehaviourRaw,
				},
			},
			expected: map[string]any{
				"system_prompt": "This is a prompt with test_prompt",
				"tools": []map[string]any{
					{
						"name":        "run_analysis",
						"description": "Run an analysis with test_description",
						"parameters": map[string]any{
							"param1": "{{param}}", // Should remain as template string
							"param2": "static",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HydrateDict(tt.template, &tt.stateParams, &tt.parameterHydrationBehavior)
			assert.NoError(t, err)

			// Add detailed output for debugging
			t.Logf("Expected: %+v", tt.expected)
			t.Logf("Result: %+v", result)

			// Check specific fields
			if tt.name == "String should respect raw hydration behavior" {
				t.Logf("normal_string - Expected: %v, Got: %v",
					tt.expected["normal_string"], result["normal_string"])
				t.Logf("raw_string - Expected: %v, Got: %v",
					tt.expected["raw_string"], result["raw_string"])

				nestedExpected := tt.expected["nested"].(map[string]any)
				nestedResult := result["nested"].(map[string]any)
				t.Logf("nested.normal_string - Expected: %v, Got: %v",
					nestedExpected["normal_string"], nestedResult["normal_string"])
				t.Logf("nested.raw_string - Expected: %v, Got: %v",
					nestedExpected["raw_string"], nestedResult["raw_string"])
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests that FindTemplateKeyStrings respects the parameter hydration behavior
func TestFindTemplateKeyStringsWithHydrationBehaviour(t *testing.T) {
	tests := []struct {
		name                       string
		input                      any
		includeOptional            bool
		parameterHydrationBehavior *map[string]any
		expected                   []string
	}{
		{
			name:                       "Simple string without hydration behavior",
			input:                      "Hello, {{name}} and {{optional?}}",
			includeOptional:            true,
			parameterHydrationBehavior: nil,
			expected:                   []string{"name", "optional"},
		},
		{
			name:                       "Simple string with hydration behavior",
			input:                      "Hello, {{name}} and {{optional?}}",
			includeOptional:            false,
			parameterHydrationBehavior: nil,
			expected:                   []string{"name"},
		},
		{
			name: "Dictionary with nested raw fields to ignore",
			input: map[string]any{
				"normal": "Hello, {{name}}!",
				"tools": map[string]any{
					"description": "{{description}}",
					"parameters": map[string]any{
						"param1": "{{param1}}",
						"param2": "{{param2}}",
					},
				},
			},
			includeOptional: false,
			parameterHydrationBehavior: &map[string]any{
				"tools": map[string]any{
					"parameters": common.ParameterHydrationBehaviourRaw,
				},
			},
			expected: []string{"description", "name"},
		},
		{
			name: "Slice with fields that have raw behavior",
			input: []any{
				"Hello, {{name}}!",
				map[string]any{
					"normal": "{{normal}}",
					"raw":    "{{rawField}}",
				},
			},
			includeOptional: false,
			parameterHydrationBehavior: &map[string]any{
				"raw": common.ParameterHydrationBehaviourRaw,
			},
			expected: []string{"name", "normal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindTemplateKeyStringsToHydrate(tt.input, tt.includeOptional, tt.parameterHydrationBehavior)

			// Sort to ensure consistent order for comparison
			sort.Strings(result)
			sort.Strings(tt.expected)

			assert.ElementsMatch(t, tt.expected, result, "Template keys should match expected values")
		})
	}
}

func TestOptionalParametersWithHydrationBehaviour(t *testing.T) {
	tests := []struct {
		name                        string
		template                    map[string]any
		stateParams                 map[string]any
		parameterHydrationBehaviour map[string]any
		expected                    map[string]any
	}{
		{
			name: "Optional parameters should return nil when missing, even with parameterHydrationBehaviour",
			template: map[string]any{
				"loops":    "{{loops}}",
				"attempts": "{{attempts?}}",
			},
			stateParams: map[string]any{
				"loops": 1,
				// attempts is intentionally missing
			},
			parameterHydrationBehaviour: map[string]any{
				// No behavior for attempts
			},
			expected: map[string]any{
				"loops":    1,
				"attempts": nil, // Should be nil, not "{{attempts?}}"
			},
		},
		{
			name: "Optional parameters with other fields as raw",
			template: map[string]any{
				"loops":     "{{loops}}",
				"attempts":  "{{attempts?}}",
				"someParam": "{{someParam}}",
			},
			stateParams: map[string]any{
				"loops": 1,
				// attempts is intentionally missing
				"someParam": "test",
			},
			parameterHydrationBehaviour: map[string]any{
				"someParam": common.ParameterHydrationBehaviourRaw,
			},
			expected: map[string]any{
				"loops":     1,
				"attempts":  nil,             // Should be nil, not "{{attempts?}}"
				"someParam": "{{someParam}}", // Should remain as template since it's marked as raw
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HydrateDict(tt.template, &tt.stateParams, &tt.parameterHydrationBehaviour)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOptionalParametersInSlice(t *testing.T) {
	tests := []struct {
		name           string
		template       []any
		stateParams    map[string]any
		expected       []any
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "Optional parameters in slice",
			template: []any{
				"{{param1}}",
				"{{param2?}}",
				"Regular string",
			},
			stateParams: map[string]any{
				"param1": "value1",
				// param2 is intentionally missing
			},
			expected: []any{
				"value1",
				nil,
				"Regular string",
			},
		},
		{
			name: "Optional parameters in slice with parameterHydrationBehaviour",
			template: []any{
				"{{param1}}",
				"{{param2?}}",
				map[string]any{
					"nestedValue": "{{param3?}}",
				},
			},
			stateParams: map[string]any{
				"param1": "value1",
				// param2 and param3 are intentionally missing
			},
			expected: []any{
				"value1",
				nil,
				map[string]any{
					"nestedValue": nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HydrateSlice(tt.template, &tt.stateParams)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseTemplateKey(t *testing.T) {
	tests := []struct {
		name     string
		keyStr   string
		expected Key
	}{
		{
			name:   "Regular key",
			keyStr: "name",
			expected: Key{
				Key:        "name",
				IsOptional: false,
			},
		},
		{
			name:   "Optional key",
			keyStr: "name?",
			expected: Key{
				Key:        "name",
				IsOptional: true,
			},
		},
		{
			name:   "Key with .Data. prefix",
			keyStr: ".Data.name",
			expected: Key{
				Key:        "name",
				IsOptional: false,
			},
		},
		{
			name:   "Optional key with .Data. prefix",
			keyStr: ".Data.name?",
			expected: Key{
				Key:        "name",
				IsOptional: true,
			},
		},
		{
			name:   "Nested key",
			keyStr: "user.name",
			expected: Key{
				Key:        "user.name",
				IsOptional: false,
			},
		},
		{
			name:   "Optional nested key",
			keyStr: "user.name?",
			expected: Key{
				Key:        "user.name",
				IsOptional: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseTemplateKey(tt.keyStr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHydrateWithDollarSignPrefix(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       any
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:     "Direct variable access with .Data prefix",
			template: "{{.Data.user.name}}",
			stateParams: map[string]any{
				"user": map[string]any{
					"name": "John",
				},
			},
			expected: "John",
		},
		{
			name:     "Direct variable access with $.Data prefix",
			template: "{{$.Data.user.name}}",
			stateParams: map[string]any{
				"user": map[string]any{
					"name": "John",
				},
			},
			expected: "John",
		},
		{
			name:     "Direct variable in string template with .Data prefix",
			template: "Hello, {{.Data.user.name}}!",
			stateParams: map[string]any{
				"user": map[string]any{
					"name": "John",
				},
			},
			expected: "Hello, John!",
		},
		{
			name:     "Direct variable in string template with $.Data prefix",
			template: "Hello, {{$.Data.user.name}}!",
			stateParams: map[string]any{
				"user": map[string]any{
					"name": "John",
				},
			},
			expected: "Hello, John!",
		},
		{
			name:     "Template with get function using .Data",
			template: "{{get \"user.name\" .Data .MissingKeys}}",
			stateParams: map[string]any{
				"user": map[string]any{
					"name": "John",
				},
			},
			expected: "John",
		},
		{
			name:     "Template with get function using $.Data",
			template: "{{get \"user.name\" $.Data $.MissingKeys}}",
			stateParams: map[string]any{
				"user": map[string]any{
					"name": "John",
				},
			},
			expected: "John",
		},
		{
			name:     "Template with nested field access using $.Data",
			template: "{{$.Data.resource.dataset.id}}",
			stateParams: map[string]any{
				"resource": map[string]any{
					"dataset": map[string]any{
						"id": "dataset1",
						"integration": map[string]any{
							"ID": "integration1",
						},
					},
				},
			},
			expected: "dataset1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compare objects directly using Hydrate instead of using HydrateString with toJSON.
			// This avoids unnecessary JSON serialization/deserialization and makes test assertions more robust.
			result, err := Hydrate(tt.template, &tt.stateParams, nil)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFindKeysInSliceWithHydrationBehaviour(t *testing.T) {
	tests := []struct {
		name                       string
		slice                      []any
		includeOptional            bool
		parameterHydrationBehavior *map[string]any
		expected                   []Key
	}{
		{
			name: "Slice with strings - no hydration behavior",
			slice: []any{
				"{{param1}}",
				"{{param2}}",
				"{{optional?}}",
			},
			includeOptional:            false,
			parameterHydrationBehavior: nil,
			expected: []Key{
				{Key: "param1", IsOptional: false},
				{Key: "param2", IsOptional: false},
			},
		},
		{
			name: "Slice with strings - with includeOptional",
			slice: []any{
				"{{param1}}",
				"{{param2}}",
				"{{optional?}}",
			},
			includeOptional:            true,
			parameterHydrationBehavior: nil,
			expected: []Key{
				{Key: "param1", IsOptional: false},
				{Key: "param2", IsOptional: false},
				{Key: "optional", IsOptional: true},
			},
		},
		{
			name: "Slice with nested dicts - no behavior",
			slice: []any{
				map[string]any{"name": "{{param1}}"},
				map[string]any{"name": "{{param2}}"},
				map[string]any{"optional": "{{optional?}}"},
			},
			includeOptional:            false,
			parameterHydrationBehavior: nil,
			expected: []Key{
				{Key: "param1", IsOptional: false},
				{Key: "param2", IsOptional: false},
			},
		},
		{
			name: "Slice with nested dicts - parameterHydrationBehavior applies to all elements",
			slice: []any{
				map[string]any{
					"name": "{{param1}}",
					"raw":  "{{shouldNotHydrate}}",
				},
				map[string]any{
					"name": "{{param2}}",
					"raw":  "{{alsoShouldNotHydrate}}",
				},
			},
			includeOptional: false,
			parameterHydrationBehavior: &map[string]any{
				"raw": common.ParameterHydrationBehaviourRaw,
			},
			expected: []Key{
				{Key: "param1", IsOptional: false},
				{Key: "param2", IsOptional: false},
			},
		},
		{
			name: "Slice with complex nested structure",
			slice: []any{
				map[string]any{
					"name": "{{param1}}",
					"nested": map[string]any{
						"value": "{{nestedValue}}",
						"raw":   "{{rawValue}}",
					},
				},
				map[string]any{
					"name": "{{param2}}",
					"nested": map[string]any{
						"value": "{{nestedValue2}}",
						"raw":   "{{rawValue2}}",
					},
				},
			},
			includeOptional: false,
			parameterHydrationBehavior: &map[string]any{
				"nested": map[string]any{
					"raw": common.ParameterHydrationBehaviourRaw,
				},
			},
			expected: []Key{
				{Key: "param1", IsOptional: false},
				{Key: "nestedValue", IsOptional: false},
				{Key: "param2", IsOptional: false},
				{Key: "nestedValue2", IsOptional: false},
			},
		},
		{
			name: "Slice with nested slices",
			slice: []any{
				map[string]any{
					"name": "{{param1}}",
					"items": []any{
						"{{itemParam1}}",
						map[string]any{
							"raw":    "{{shouldNotHydrate}}",
							"normal": "{{shouldHydrate}}",
						},
					},
				},
				"{{param2}}",
			},
			includeOptional: false,
			parameterHydrationBehavior: &map[string]any{
				"raw": common.ParameterHydrationBehaviourRaw,
			},
			expected: []Key{
				{Key: "itemParam1", IsOptional: false},
				{Key: "param1", IsOptional: false},
				{Key: "param2", IsOptional: false},
				{Key: "shouldHydrate", IsOptional: false},
				{Key: "shouldNotHydrate", IsOptional: false}, // Current implementation doesn't respect raw at this depth
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findKeysInSlice(tt.slice, tt.includeOptional, tt.parameterHydrationBehavior)

			// Sort results and expected for stable comparison
			sortKeys := func(keys []Key) {
				sort.Slice(keys, func(i, j int) bool {
					return keys[i].Key < keys[j].Key
				})
			}

			sortKeys(result)
			sortKeys(tt.expected)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAddkeytoall(t *testing.T) {
	tests := []struct {
		name        string
		listKey     string
		key         string
		value       any
		data        map[string]any
		missingKeys *[]string
		expected    []any
	}{
		{
			name:    "Add key to list of dicts",
			listKey: "memories",
			key:     "resource_id",
			value:   "resource-123",
			data: map[string]any{
				"memories": []any{
					map[string]any{"ID": "1", "content": "memory 1"},
					map[string]any{"ID": "2", "content": "memory 2"},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				map[string]any{"ID": "1", "content": "memory 1", "resource_id": "resource-123"},
				map[string]any{"ID": "2", "content": "memory 2", "resource_id": "resource-123"},
			},
		},
		{
			name:    "Add nested key to list of dicts when parent object doesn't exist",
			listKey: "memories",
			key:     "metadata.resource_id",
			value:   "resource-123",
			data: map[string]any{
				"memories": []any{
					map[string]any{"ID": "1", "content": "memory 1"},
					map[string]any{"ID": "2", "content": "memory 2"},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				map[string]any{"ID": "1", "content": "memory 1", "metadata": map[string]any{"resource_id": "resource-123"}},
				map[string]any{"ID": "2", "content": "memory 2", "metadata": map[string]any{"resource_id": "resource-123"}},
			},
		},
		{
			name:        "List empty",
			listKey:     "nonexistent_list",
			key:         "resource_id",
			value:       "resource-123",
			data:        map[string]any{},
			missingKeys: &[]string{},
			expected:    []any{},
		},
		{
			name:    "List with non-Dict items",
			listKey: "mixed_list",
			key:     "resource_id",
			value:   "resource-123",
			data: map[string]any{
				"mixed_list": []any{
					"string",
					123,
					map[string]any{"ID": "3", "content": "memory 3"},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				"string",
				123,
				map[string]any{"ID": "3", "content": "memory 3", "resource_id": "resource-123"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s", tt.name)
			result := addkeytoall(tt.listKey, tt.key, tt.value, tt.data, tt.missingKeys)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTemplateWithAddkeytoallFunction tests using the addkeytoall function
// as part of a template string, ensuring it adds keys to items in a list correctly
func TestTemplateWithAddkeytoallFunction(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       any
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:     "Add key to list of dicts",
			template: "{{addkeytoall \"memories\" \"resource_id\" \"resource-123\"}}",
			stateParams: map[string]any{
				"memories": []any{
					map[string]any{"ID": "1", "content": "memory 1"},
					map[string]any{"ID": "2", "content": "memory 2"},
				},
			},
			expected: []any{
				map[string]any{"ID": "1", "content": "memory 1", "resource_id": "resource-123"},
				map[string]any{"ID": "2", "content": "memory 2", "resource_id": "resource-123"},
			},
		},
		{
			name:     "Add nested key to list",
			template: "{{addkeytoall \"memories\" \"metadata.resource_id\" \"resource-123\"}}",
			stateParams: map[string]any{
				"memories": []any{
					map[string]any{"ID": "1", "content": "memory 1", "metadata": map[string]any{}},
					map[string]any{"ID": "2", "content": "memory 2", "metadata": map[string]any{}},
				},
			},
			expected: []any{
				map[string]any{"ID": "1", "content": "memory 1", "metadata": map[string]any{"resource_id": "resource-123"}},
				map[string]any{"ID": "2", "content": "memory 2", "metadata": map[string]any{"resource_id": "resource-123"}},
			},
		},
		{
			name:     "Empty list",
			template: "{{addkeytoall \"empty_list\" \"resource_id\" \"resource-123\"}}",
			stateParams: map[string]any{
				"empty_list": []any{},
			},
			expected: []any{},
		},
		{
			name:        "List not found",
			template:    "{{addkeytoall \"nonexistent_list\" \"resource_id\" \"resource-123\"}}",
			stateParams: map[string]any{},
			expected:    []any{},
		},
		{
			name:     "Non-dict items in list",
			template: "{{addkeytoall \"mixed_list\" \"resource_id\" \"resource-123\"}}",
			stateParams: map[string]any{
				"mixed_list": []any{
					map[string]any{"ID": "1", "content": "memory 1"},
					"not a dict",
					123,
				},
			},
			expected: []any{
				map[string]any{"ID": "1", "content": "memory 1", "resource_id": "resource-123"},
				"not a dict",
				123, // No longer converts to float64
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compare objects directly using Hydrate instead of using HydrateString with toJSON.
			// This avoids unnecessary JSON serialization/deserialization and makes test assertions more robust.
			result, err := Hydrate(tt.template, &tt.stateParams, nil)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestExtractSlice(t *testing.T) {
	tests := []struct {
		name        string
		array       string
		field       string
		data        map[string]any
		missingKeys *[]string
		expected    []any
	}{
		{
			name:  "Extract objects from array",
			array: "items",
			field: "memory",
			data: map[string]any{
				"items": []any{
					map[string]any{
						"memory": map[string]any{
							"ID":      "1",
							"Content": "Memory 1",
						},
						"distance": 0.5,
					},
					map[string]any{
						"memory": map[string]any{
							"ID":      "2",
							"Content": "Memory 2",
						},
						"distance": 0.3,
					},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				map[string]any{
					"ID":      "1",
					"Content": "Memory 1",
				},
				map[string]any{
					"ID":      "2",
					"Content": "Memory 2",
				},
			},
		},
		{
			name:  "Extract string values",
			array: "people",
			field: "name",
			data: map[string]any{
				"people": []any{
					map[string]any{"name": "Alice", "age": 25},
					map[string]any{"name": "Bob", "age": 30},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				"Alice",
				"Bob",
			},
		},
		{
			name:  "Extract numeric values",
			array: "people",
			field: "age",
			data: map[string]any{
				"people": []any{
					map[string]any{"name": "Alice", "age": 25},
					map[string]any{"name": "Bob", "age": 30},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				25,
				30,
			},
		},
		{
			name:  "Empty array",
			array: "items",
			field: "memory",
			data: map[string]any{
				"items": []any{},
			},
			missingKeys: &[]string{},
			expected:    []any{},
		},
		{
			name:        "Array not found",
			array:       "nonexistent",
			field:       "memory",
			data:        map[string]any{},
			missingKeys: &[]string{},
			expected:    []any{},
		},
		{
			name:  "Missing field in some items",
			array: "items",
			field: "memory",
			data: map[string]any{
				"items": []any{
					map[string]any{
						"memory": map[string]any{
							"ID":      "1",
							"Content": "Memory 1",
						},
					},
					map[string]any{
						"other": "value",
					},
					map[string]any{
						"memory": map[string]any{
							"ID":      "2",
							"Content": "Memory 2",
						},
					},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				map[string]any{
					"ID":      "1",
					"Content": "Memory 1",
				},
				map[string]any{
					"ID":      "2",
					"Content": "Memory 2",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSlice(tt.array, tt.field, tt.data, tt.missingKeys)
			assert.Equal(t, tt.expected, result, "Test case: %s", tt.name)
		})
	}
}

// TestTemplateWithExtractSliceFunction tests the extractSlice function in templates
// to create new arrays with values extracted from nested objects
func TestTemplateWithExtractSliceFunction(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       any
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:     "Extract objects",
			template: "{{extractSlice \"items\" \"memory\"}}",
			stateParams: map[string]any{
				"items": []any{
					map[string]any{
						"memory": map[string]any{
							"ID":      "1",
							"Content": "Memory 1",
						},
						"distance": 0.5,
					},
					map[string]any{
						"memory": map[string]any{
							"ID":      "2",
							"Content": "Memory 2",
						},
						"distance": 0.3,
					},
				},
			},
			expected: []any{
				map[string]any{
					"ID":      "1",
					"Content": "Memory 1",
				},
				map[string]any{
					"ID":      "2",
					"Content": "Memory 2",
				},
			},
		},
		{
			name:     "Extract string values",
			template: "{{extractSlice \"people\" \"name\"}}",
			stateParams: map[string]any{
				"people": []any{
					map[string]any{"name": "Alice", "age": 25},
					map[string]any{"name": "Bob", "age": 30},
				},
			},
			expected: []any{"Alice", "Bob"},
		},
		{
			name:     "Empty array",
			template: "{{extractSlice \"empty_list\" \"memory\"}}",
			stateParams: map[string]any{
				"empty_list": []any{},
			},
			expected: []any{},
		},
		{
			name:        "Array not found",
			template:    "{{extractSlice \"nonexistent_list\" \"memory\"}}",
			stateParams: map[string]any{},
			expected:    []any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compare objects directly using Hydrate instead of using HydrateString with toJSON.
			// This avoids unnecessary JSON serialization/deserialization and makes test assertions more robust.
			result, err := Hydrate(tt.template, &tt.stateParams, nil)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSliceEndKeepFirstUserMessage(t *testing.T) {
	tests := []struct {
		name        string
		sliceKey    string
		n           int
		data        map[string]any
		missingKeys *[]string
		expected    []any
	}{
		{
			name:     "First message in slice is user - return slice as is",
			sliceKey: "messages",
			n:        2,
			data: map[string]any{
				"messages": []any{
					map[string]any{"role": "assistant", "content": "Assistant message 1"},
					map[string]any{"role": "user", "content": "User message 1"},
					map[string]any{"role": "assistant", "content": "Assistant message 2"},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				map[string]any{"role": "user", "content": "User message 1"},
				map[string]any{"role": "assistant", "content": "Assistant message 2"},
			},
		},
		{
			name:     "First message in slice is not user - should prepend user message",
			sliceKey: "messages",
			n:        3,
			data: map[string]any{
				"messages": []any{
					map[string]any{"role": "system", "content": "System message"},
					map[string]any{"role": "user", "content": "User message 1"},
					map[string]any{"role": "assistant", "content": "Assistant message 1"},
					map[string]any{"role": "user", "content": "User message 2"},
					map[string]any{"role": "assistant", "content": "Assistant message 2"},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				map[string]any{"role": "user", "content": "User message 1"},
				map[string]any{"role": "assistant", "content": "Assistant message 1"},
				map[string]any{"role": "user", "content": "User message 2"},
				map[string]any{"role": "assistant", "content": "Assistant message 2"},
			},
		},
		{
			name:     "First message in slice is not user - prepend first user message",
			sliceKey: "messages",
			n:        3,
			data: map[string]any{
				"messages": []any{
					map[string]any{"role": "user", "content": "User message 1"},
					map[string]any{"role": "assistant", "content": "Assistant message 1"},
					map[string]any{"role": "assistant", "content": "Assistant message 2"},
					map[string]any{"role": "assistant", "content": "Assistant message 3"},
					map[string]any{"role": "assistant", "content": "Assistant message 4"},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				map[string]any{"role": "user", "content": "User message 1"},
				map[string]any{"role": "assistant", "content": "Assistant message 2"},
				map[string]any{"role": "assistant", "content": "Assistant message 3"},
				map[string]any{"role": "assistant", "content": "Assistant message 4"},
			},
		},
		{
			name:     "Example from user - u3 should be prepended",
			sliceKey: "messages",
			n:        3,
			data: map[string]any{
				"messages": []any{
					map[string]any{"role": "user", "content": "u1"},
					map[string]any{"role": "user", "content": "u2"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "user", "content": "u3"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				map[string]any{"role": "user", "content": "u3"},
				map[string]any{"role": "assistant", "content": "a"},
				map[string]any{"role": "assistant", "content": "a"},
				map[string]any{"role": "assistant", "content": "a"},
			},
		},
		{
			name:     "No user messages - return slice as is",
			sliceKey: "messages",
			n:        3,
			data: map[string]any{
				"messages": []any{
					map[string]any{"role": "system", "content": "System message"},
					map[string]any{"role": "assistant", "content": "Assistant message 1"},
					map[string]any{"role": "assistant", "content": "Assistant message 2"},
					map[string]any{"role": "assistant", "content": "Assistant message 3"},
					map[string]any{"role": "assistant", "content": "Assistant message 4"},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				map[string]any{"role": "assistant", "content": "Assistant message 2"},
				map[string]any{"role": "assistant", "content": "Assistant message 3"},
				map[string]any{"role": "assistant", "content": "Assistant message 4"},
			},
		},
		{
			name:     "n >= slice length - return all messages",
			sliceKey: "messages",
			n:        10,
			data: map[string]any{
				"messages": []any{
					map[string]any{"role": "user", "content": "User message 1"},
					map[string]any{"role": "assistant", "content": "Assistant message 1"},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				map[string]any{"role": "user", "content": "User message 1"},
				map[string]any{"role": "assistant", "content": "Assistant message 1"},
			},
		},
		{
			name:     "Empty slice",
			sliceKey: "messages",
			n:        3,
			data: map[string]any{
				"messages": []any{},
			},
			missingKeys: &[]string{},
			expected:    []any{},
		},
		{
			name:        "Slice not found",
			sliceKey:    "nonexistent",
			n:           3,
			data:        map[string]any{},
			missingKeys: &[]string{},
			expected:    nil,
		},
		{
			name:     "Works with map[string]any format",
			sliceKey: "messages",
			n:        2,
			data: map[string]any{
				"messages": []any{
					map[string]any{"role": "user", "content": "User message 1"},
					map[string]any{"role": "assistant", "content": "Assistant message 1"},
					map[string]any{"role": "assistant", "content": "Assistant message 2"},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				map[string]any{"role": "user", "content": "User message 1"},
				map[string]any{"role": "assistant", "content": "Assistant message 1"},
				map[string]any{"role": "assistant", "content": "Assistant message 2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sliceEndKeepFirstUserMessage(tt.sliceKey, tt.n, tt.data, tt.missingKeys)
			assert.Equal(t, tt.expected, result, "Test case: %s", tt.name)
		})
	}
}

func TestTemplateWithSliceEndKeepFirstUserMessage(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       any
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:     "Basic template usage",
			template: "{{sliceEndKeepFirstUserMessage \"messages\" 3}}",
			stateParams: map[string]any{
				"messages": []any{
					map[string]any{"role": "user", "content": "u1"},
					map[string]any{"role": "user", "content": "u2"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "user", "content": "u3"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
					map[string]any{"role": "assistant", "content": "a"},
				},
			},
			expected: []any{
				map[string]any{"role": "user", "content": "u3"},
				map[string]any{"role": "assistant", "content": "a"},
				map[string]any{"role": "assistant", "content": "a"},
				map[string]any{"role": "assistant", "content": "a"},
			},
		},
		{
			name:     "First message in slice is already user",
			template: "{{sliceEndKeepFirstUserMessage \"messages\" 2}}",
			stateParams: map[string]any{
				"messages": []any{
					map[string]any{"role": "assistant", "content": "Assistant message 1"},
					map[string]any{"role": "user", "content": "User message 1"},
					map[string]any{"role": "assistant", "content": "Assistant message 2"},
				},
			},
			expected: []any{
				map[string]any{"role": "user", "content": "User message 1"},
				map[string]any{"role": "assistant", "content": "Assistant message 2"},
			},
		},
		{
			name:        "Array not found",
			template:    "{{sliceEndKeepFirstUserMessage \"nonexistent\" 3}}",
			stateParams: map[string]any{},
			expected:    []any(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Hydrate(tt.template, &tt.stateParams, nil)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestHydrateStringWithDataReferences tests that function arguments with $.Data or .Data prefixes
// are properly hydrated when used in functions
func TestHydrateStringWithDataReferences(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       any
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:     "Direct reference to variable using $.Data",
			template: "{{$.Data.resource_id}}",
			stateParams: map[string]any{
				"resource_id": "resource-123",
			},
			expected: "resource-123",
		},
		{
			name:     "Direct reference to variable using .Data",
			template: "{{.Data.resource_id}}",
			stateParams: map[string]any{
				"resource_id": "resource-123",
			},
			expected: "resource-123",
		},
		{
			name:     "Variable reference in string template",
			template: "The resource ID is: {{$.Data.resource_id}}",
			stateParams: map[string]any{
				"resource_id": "resource-123",
			},
			expected: "The resource ID is: resource-123",
		},
		{
			name:     "nested value using dot notation",
			template: "{{.Data.integration_query.resource_id}}",
			stateParams: map[string]any{
				"integration_query": map[string]any{
					"resource_id": "integration-resource-456",
				},
			},
			expected: "integration-resource-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Hydrate(tt.template, &tt.stateParams, nil)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestAddkeytoallWithDataReferences specifically tests the addkeytoall function
// with different data reference patterns
func TestAddkeytoallWithDataReferences(t *testing.T) {
	// Test direct function calls with concrete values
	tests := []struct {
		name     string
		listKey  string
		key      string
		value    any
		data     map[string]any
		expected []any
	}{
		{
			name:    "Add string value directly",
			listKey: "memories",
			key:     "resourceId",
			value:   "resource-123",
			data: map[string]any{
				"memories": []any{
					map[string]any{"ID": "1", "content": "memory 1"},
					map[string]any{"ID": "2", "content": "memory 2"},
				},
			},
			expected: []any{
				map[string]any{"ID": "1", "content": "memory 1", "resourceId": "resource-123"},
				map[string]any{"ID": "2", "content": "memory 2", "resourceId": "resource-123"},
			},
		},
		{
			name:    "Add nested key with dot notation",
			listKey: "memories",
			key:     "metadata.resourceId",
			value:   "resource-789",
			data: map[string]any{
				"memories": []any{
					map[string]any{"ID": "1", "content": "memory 1"},
					map[string]any{"ID": "2", "content": "memory 2"},
				},
			},
			expected: []any{
				map[string]any{"ID": "1", "content": "memory 1", "metadata": map[string]any{"resourceId": "resource-789"}},
				map[string]any{"ID": "2", "content": "memory 2", "metadata": map[string]any{"resourceId": "resource-789"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missingKeys := []string{}
			result := addkeytoall(tt.listKey, tt.key, tt.value, tt.data, &missingKeys)

			// Compare objects directly
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHydrateWithAddkeytoall tests the addkeytoall function in templates
// with different path notations and value types. This ensures the function can handle
// dot notation for nested objects and various value types without JSON conversion
func TestHydrateWithAddkeytoall(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]any
		expected []any
	}{
		{
			name:     "addkeytoall with direct string value",
			template: `{{addkeytoall "memories" "resourceId" "resource-123"}}`,
			data: map[string]any{
				"memories": []any{
					map[string]any{"ID": "1", "content": "memory 1"},
					map[string]any{"ID": "2", "content": "memory 2"},
				},
			},
			expected: []any{
				map[string]any{"ID": "1", "content": "memory 1", "resourceId": "resource-123"},
				map[string]any{"ID": "2", "content": "memory 2", "resourceId": "resource-123"},
			},
		},
		{
			name:     "addkeytoall with nested path using dot notation",
			template: `{{addkeytoall "memories" "metadata.resourceId" "resource-789"}}`,
			data: map[string]any{
				"memories": []any{
					map[string]any{"ID": "1", "content": "memory 1"},
					map[string]any{"ID": "2", "content": "memory 2"},
				},
			},
			expected: []any{
				map[string]any{"ID": "1", "content": "memory 1", "metadata": map[string]any{"resourceId": "resource-789"}},
				map[string]any{"ID": "2", "content": "memory 2", "metadata": map[string]any{"resourceId": "resource-789"}},
			},
		},
		{
			name:     "Use dot notation in real-world scenario matching the memory bot",
			template: `{{addkeytoall "memories" "memory.ResourceID" "resource-123"}}`,
			data: map[string]any{
				"memories": []any{
					map[string]any{"ID": "1", "content": "memory 1"},
					map[string]any{"ID": "2", "content": "memory 2"},
				},
			},
			expected: []any{
				map[string]any{"ID": "1", "content": "memory 1", "memory": map[string]any{"ResourceID": "resource-123"}},
				map[string]any{"ID": "2", "content": "memory 2", "memory": map[string]any{"ResourceID": "resource-123"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compare objects directly using Hydrate instead of using HydrateString with toJSON.
			// This avoids unnecessary JSON serialization/deserialization and makes test assertions more robust.
			result, err := Hydrate(tt.template, &tt.data, nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTemplateWithAddkeyFunction tests using the addkey function in templates
// to modify objects by adding new key-value pairs
func TestTemplateWithAddkeyFunction(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       any
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:     "Add key to object",
			template: "{{addkey \"object\" \"newKey\" (get \"value\")}}",
			stateParams: map[string]any{
				"object": map[string]any{"existingKey": "existingValue"},
				"value":  "newValue",
			},
			expected: map[string]any{
				"existingKey": "existingValue",
				"newKey":      "newValue",
			},
		},
		{
			name:     "Add key to empty object",
			template: "{{addkey \"emptyObject\" \"firstKey\" (get \"value\")}}",
			stateParams: map[string]any{
				"emptyObject": map[string]any{},
				"value":       "someValue",
			},
			expected: map[string]any{
				"firstKey": "someValue",
			},
		},
		// "Object not found" case is special - handled in custom assertion
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compare objects directly using Hydrate instead of using HydrateString with toJSON.
			// This avoids unnecessary JSON serialization/deserialization and makes test assertions more robust.
			result, err := Hydrate(tt.template, &tt.stateParams, nil)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}

	// Special test for "Object not found" case
	t.Run("Object not found", func(t *testing.T) {
		template := "{{addkey \"nonExistentObject\" \"key\" (get \"value\")}}"
		stateParams := map[string]any{"value": "someValue"}

		result, err := Hydrate(template, &stateParams, nil)
		assert.NoError(t, err)

		// Verify the result is nil using reflect
		assert.Nil(t, result, "Result should be nil for non-existent object")
	})
}

// TestRegexReplace tests the regexReplace function that performs regex replacement
func TestRegexReplace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		pattern     string
		replacement string
		input       any
		expected    string
	}{
		{
			name:        "Replace tabs with 2 spaces",
			pattern:     "\t",
			replacement: "  ",
			input:       "function() {\n\tif (true) {\n\t\treturn 'hello';\n\t}\n}",
			expected:    "function() {\n  if (true) {\n    return 'hello';\n  }\n}",
		},
		{
			name:        "Replace multiple spaces with single space",
			pattern:     " +",
			replacement: " ",
			input:       "hello    world     test",
			expected:    "hello world test",
		},
		{
			name:        "Replace digits with X",
			pattern:     "\\d",
			replacement: "X",
			input:       "Phone: 123-456-7890",
			expected:    "Phone: XXX-XXX-XXXX",
		},
		{
			name:        "Replace beginning of line whitespace",
			pattern:     "(?m)^\\s+",
			replacement: "",
			input:       "  line1\n    line2\n\tline3",
			expected:    "line1\nline2\nline3",
		},
		{
			name:        "Replace word boundaries",
			pattern:     "\\btest\\b",
			replacement: "TEST",
			input:       "This is a test of testing tests",
			expected:    "This is a TEST of testing tests",
		},
		{
			name:        "No match - return original",
			pattern:     "xyz",
			replacement: "abc",
			input:       "hello world",
			expected:    "hello world",
		},
		{
			name:        "Empty string input",
			pattern:     "\t",
			replacement: "  ",
			input:       "",
			expected:    "",
		},
		{
			name:        "Non-string input converted to string",
			pattern:     "123",
			replacement: "XXX",
			input:       123456,
			expected:    "XXX456",
		},
		{
			name:        "Invalid regex pattern - return original",
			pattern:     "[",
			replacement: "  ",
			input:       "hello\tworld",
			expected:    "hello\tworld",
		},
		{
			name:        "Complex replacement with capture groups",
			pattern:     "(\\w+)\\s+(\\w+)",
			replacement: "$2 $1",
			input:       "hello world",
			expected:    "world hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := regexReplace(tt.pattern, tt.replacement, tt.input)
			if result != tt.expected {
				t.Errorf("regexReplace(%q, %q, %v) = %q, expected %q",
					tt.pattern, tt.replacement, tt.input, result, tt.expected)
			}
		})
	}
}

// TestBasicFunctionInTemplate tests that basic functions (those that don't need .Data and .MissingKeys)
// work correctly when called as single function templates
func TestBasicFunctionInTemplate(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       any
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:     "regexReplace single function call",
			template: `{{regexReplace "\t" "    " "hello\tworld"}}`,
			stateParams: map[string]any{
				"code": "hello\tworld",
			},
			expected: "hello    world",
		},
		{
			name:     "regexReplace with variable from data",
			template: `{{regexReplace "\t" "    " $.Data.code}}`,
			stateParams: map[string]any{
				"code": "hello\tworld\ttest",
			},
			expected: "hello    world    test",
		},
		// Test the exact scenario from the error logs
		{
			name:     "regexReplace exact error scenario from logs",
			template: `{{regexReplace "\t" "    " $.Data.code}}`,
			stateParams: map[string]any{
				"code": "function() {\n\tif (true) {\n\t\treturn 'hello';\n\t}\n}",
			},
			expected: "function() {\n    if (true) {\n        return 'hello';\n    }\n}",
		},
		{
			name:        "toJSON single function call",
			template:    `{{toJSON "hello"}}`,
			stateParams: map[string]any{},
			expected:    `"hello"`,
		},
		{
			name:        "len single function call",
			template:    `{{len "hello"}}`,
			stateParams: map[string]any{},
			expected:    5,
		},
		{
			name:        "add two numbers",
			template:    `{{add 5 3}}`,
			stateParams: map[string]any{},
			expected:    8,
		},
		{
			name:        "nested basic functions",
			template:    `{{toJSON (add 5 3)}}`,
			stateParams: map[string]any{},
			expected:    "8",
		},
		{
			name:        "truthyValue with literal",
			template:    `{{truthyValue "hello"}}`,
			stateParams: map[string]any{},
			expected:    true,
		},
		{
			name:        "toString with number",
			template:    `{{toString 123}}`,
			stateParams: map[string]any{},
			expected:    "123",
		},
		{
			name:        "truncateString with literal",
			template:    `{{truncateString "hello world" 5}}`,
			stateParams: map[string]any{},
			expected:    "he...",
		},
		{
			name:        "genUUID function call",
			template:    `{{genUUID}}`,
			stateParams: map[string]any{},
			expected:    "", // Will be validated separately for UUID format
		},
		{
			name:        "generateUUID function call (alias)",
			template:    `{{generateUUID}}`,
			stateParams: map[string]any{},
			expected:    "", // Will be validated separately for UUID format
		},
		{
			name:        "now function call",
			template:    `{{now}}`,
			stateParams: map[string]any{},
			expected:    "", // Will be validated separately for time format
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Hydrate(tt.template, &tt.stateParams, nil)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				
				// Special validation for UUID and time functions
				switch tt.name {
				case "genUUID function call", "generateUUID function call (alias)":
					// Validate UUID format (should be 36 characters with dashes)
					resultStr, ok := result.(string)
					assert.True(t, ok, "Result should be a string")
					assert.Len(t, resultStr, 36, "UUID should be 36 characters long")
					assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, resultStr, "Should be valid UUID format")
				case "now function call":
					// Validate time format (ISO 8601: 2006-01-02T15:04:05Z)
					resultStr, ok := result.(string)
					assert.True(t, ok, "Result should be a string")
					assert.Len(t, resultStr, 20, "Time should be 20 characters long")
					assert.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`, resultStr, "Should be valid ISO 8601 format")
					// Verify it can be parsed as time
					_, parseErr := time.Parse("2006-01-02T15:04:05Z", resultStr)
					assert.NoError(t, parseErr, "Should be parseable as time")
				default:
					assert.Equal(t, tt.expected, result)
				}
			}
		})
	}
}

func TestCoalesceWithLiteralValues(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       any
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:        "Coalesce with missing key and literal number",
			template:    `{{coalesce "missing_key?" 0}}`,
			stateParams: map[string]any{
				// missing_key is intentionally not present
			},
			expected: 0, // Should return the literal 0, not try to look up "0" as a key
		},
		{
			name:        "Coalesce with missing key and literal string",
			template:    `{{coalesce "missing_key?" "default"}}`,
			stateParams: map[string]any{
				// missing_key is intentionally not present
			},
			expected: "default", // Should return the literal "default"
		},
		{
			name:     "Coalesce with existing key",
			template: `{{coalesce "existing_key?" "default"}}`,
			stateParams: map[string]any{
				"existing_key": "actual_value",
			},
			expected: "actual_value", // Should return the existing value
		},
		{
			name:        "Coalesce in condition-like scenario",
			template:    `{{coalesce "code_retry_loops?" 0}}`,
			stateParams: map[string]any{
				// code_retry_loops is intentionally not present (first failure scenario)
			},
			expected: 0, // Should return 0 for use in conditions like "0 < 2"
		},
		{
			name:     "Coalesce with existing counter",
			template: `{{coalesce "code_retry_loops?" 0}}`,
			stateParams: map[string]any{
				"code_retry_loops": 2, // After some retries
			},
			expected: 2, // Should return the existing counter value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Hydrate(tt.template, &tt.stateParams, nil)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result, "Coalesce should handle literal fallback values correctly")
			}
		})
	}
}

func TestCoalesceFunctionDirect(t *testing.T) {
	data := map[string]any{}
	missingKeys := []string{}

	// Test direct function calls with correct types
	tests := []struct {
		name         string
		key          any
		fallback     any
		expected     any
		expectedType string
	}{
		{
			name:         "String key missing, int fallback",
			key:          "missing_key?",
			fallback:     0,
			expected:     0,
			expectedType: "int",
		},
		{
			name:         "String key missing, string fallback",
			key:          "missing_key?",
			fallback:     "default",
			expected:     "default",
			expectedType: "string",
		},
		{
			name:         "String key missing, float64 fallback",
			key:          "missing_key?",
			fallback:     0.0,
			expected:     0.0,
			expectedType: "float64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := coalesce(tt.key, tt.fallback, data, &missingKeys)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedType, fmt.Sprintf("%T", result))
		})
	}
}

func TestCoalesceOriginalErrorScenario(t *testing.T) {
	// Reproduce the exact scenario from the error logs:
	// "LessThan evaluating <nil> <nil> < 2"
	// This was happening because coalesce("code_retry_loops?", 0) was returning nil
	// instead of 0 when code_retry_loops was missing

	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       any
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:        "Original error scenario - code_retry_loops missing",
			template:    `{{coalesce "code_retry_loops?" 0}}`,
			stateParams: map[string]any{
				// code_retry_loops intentionally missing (first step failure)
			},
			expected: 0, // Should return 0, not nil
		},
		{
			name:     "Original error scenario - code_retry_loops exists",
			template: `{{coalesce "code_retry_loops?" 0}}`,
			stateParams: map[string]any{
				"code_retry_loops": 1, // After one retry
			},
			expected: 1, // Should return the existing value
		},
		{
			name:        "Complete condition template scenario",
			template:    `{{coalesce "code_retry_loops?" 0}} < 2`,
			stateParams: map[string]any{
				// code_retry_loops intentionally missing
			},
			expected: "0 < 2", // Template should resolve to this string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Hydrate(tt.template, &tt.stateParams, nil)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result, "Coalesce should handle the original error scenario correctly")
			}
		})
	}
}

func TestFilterFunction(t *testing.T) {
	data := map[string]any{
		"items": []any{
			map[string]any{"name": "alice", "age": 30},
			map[string]any{"name": "bob", "age": 25},
			map[string]any{"name": "charlie", "age": 30},
		},
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "filter function",
			template: `{{filter "items" "age" "eq" 30}}`,
			expected: "[map[age:30 name:alice] map[age:30 name:charlie]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HydrateString(tt.template, &data)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNestedFunctionWithNilHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expectedError  bool
		expectedErrMsg string
		expected       any
	}{
		{
			name:        "Nested function with missing key should error",
			template:    "{{len (get \"similar_memories\")}}",
			stateParams: map[string]any{
				// similar_memories is intentionally missing
			},
			expectedError:  true,
			expectedErrMsg: "info needed for keys", // Changed to match InfoNeededError
		},
		{
			name:     "Nested function with existing empty array",
			template: "{{len (get \"similar_memories\")}}",
			stateParams: map[string]any{
				"similar_memories": []any{},
			},
			expected: 0,
		},
		{
			name:     "Nested function with existing array",
			template: "{{len (get \"similar_memories\")}}",
			stateParams: map[string]any{
				"similar_memories": []any{"mem1", "mem2", "mem3"},
			},
			expected: 3,
		},
		{
			name:        "Complex nested function with missing inner key",
			template:    "{{toJSON (mapToDict \"missing_list\" \"id\")}}",
			stateParams: map[string]any{
				// missing_list is intentionally missing
			},
			expected: "[]", // mapToDict returns empty array for missing keys
		},
		{
			name:     "Nested function with nil value in data",
			template: "{{len (get \"null_field\")}}",
			stateParams: map[string]any{
				"null_field": nil,
			},
			expected: 0, // get returns nil for nil values, len of nil is 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := Hydrate(tt.template, &tt.stateParams, nil)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGreaterThanConditionWithMissingField(t *testing.T) {
	t.Parallel()

	// This test simulates the original panic scenario
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expectedError  bool
		expectedErrMsg string
		expected       any
	}{
		{
			name:        "GreaterThan condition with missing field",
			template:    "{{gt (len (get \"similar_memories\")) 0}}",
			stateParams: map[string]any{
				// similar_memories is missing, which was causing the panic
			},
			expectedError:  true,
			expectedErrMsg: "info needed for keys", // InfoNeededError is thrown
		},
		{
			name:     "GreaterThan condition with empty array",
			template: "{{gt (len (get \"similar_memories\")) 0}}",
			stateParams: map[string]any{
				"similar_memories": []any{},
			},
			expected: false,
		},
		{
			name:     "GreaterThan condition with populated array",
			template: "{{gt (len (get \"similar_memories\")) 0}}",
			stateParams: map[string]any{
				"similar_memories": []any{"mem1", "mem2"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := Hydrate(tt.template, &tt.stateParams, nil)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
