package template

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHydrateStringAdvanced(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       string
		expectedError  bool
		expectedErrMsg string
	}{
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
			result, err := HydrateString(tt.template, &tt.stateParams, nil)

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

func TestHydrateDictAdvanced(t *testing.T) {
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
			expectedErrMsg: "info needed for keys [name]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HydrateDict(tt.template, &tt.stateParams, nil)

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

func TestHydrateSliceWithBehavior(t *testing.T) {
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
			expectedErrMsg: "info needed for keys [name]",
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
				"raw":    ParameterHydrationBehaviourRaw, // Applies to all elements
				"hidden": ParameterHydrationBehaviourRaw, // Applies to all elements
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
					// This map[string]any applies to the "details" field, which is a slice
					// The map[string]any will be passed down to all elements in that slice
					"raw": ParameterHydrationBehaviourRaw, // Applied to "raw" key in all elements of the details slice
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
							"raw":    "{{rawParam}}", // This will be hydrated because details is processed as a map[string]any, not as part of the slice behavior
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
					"raw": ParameterHydrationBehaviourRaw, // This correctly targets the 'raw' field in all elements of the details array
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
				"raw": ParameterHydrationBehaviourRaw,
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
					"rawValue": ParameterHydrationBehaviourRaw, // Only this specific path is configured not to be hydrated
				},
				"raw": ParameterHydrationBehaviourRaw, // This only applies to top-level keys named "raw", not those in nested objects
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
				result, err = HydrateSlice(tt.template, &tt.stateParams, nil)
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
			result, err := HydrateString(tt.template, &stateParams, nil)
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
				"tools": ParameterHydrationBehaviourRaw,
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
					"parameters": ParameterHydrationBehaviourRaw,
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
						"raw": ParameterHydrationBehaviourRaw,
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
					"parameters": ParameterHydrationBehaviourRaw,
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
					"should_leave": ParameterHydrationBehaviourRaw,
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
					"parameters": ParameterHydrationBehaviourRaw,
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
			assert.Equal(t, tt.expected, result)
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
					"parameters": ParameterHydrationBehaviourRaw,
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
				"raw": ParameterHydrationBehaviourRaw,
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
				"someParam": ParameterHydrationBehaviourRaw,
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
			result, err := HydrateSlice(tt.template, &tt.stateParams, nil)

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

func TestMapToDict(t *testing.T) {
	tests := []struct {
		name           string
		listKey        string
		dictKey        string
		stateParams    map[string]any
		expected       []map[string]any
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:    "Convert string list to dict list",
			listKey: "stringList",
			dictKey: "key",
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
			name:    "Empty list",
			listKey: "emptyList",
			dictKey: "key",
			stateParams: map[string]any{
				"emptyList": []any{},
			},
			expected: []map[string]any{},
		},
		{
			name:        "Non-existent list",
			listKey:     "nonExistentList",
			dictKey:     "key",
			stateParams: map[string]any{},
			expected:    []map[string]any{},
		},
		{
			name:    "List with mixed types",
			listKey: "mixedList",
			dictKey: "key",
			stateParams: map[string]any{
				"mixedList": []any{"string", 123, true, nil},
			},
			expected: []map[string]any{
				{"key": "string"},
				{"key": 123},
				{"key": true},
				{"key": nil},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call mapToDict directly
			var missingKeys []string
			result := mapToDict(tt.listKey, tt.dictKey, tt.stateParams, &missingKeys)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNestedTemplateFunctions(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       string
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:     "mapToDict nested function",
			template: "{{toJSON (mapToDict \"stringList\" \"key\")}}",
			stateParams: map[string]any{
				"stringList": []any{"value1", "value2"},
			},
			expected: `[{"key":"value1"},{"key":"value2"}]`,
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
			// Use HydrateString to get string results
			result, err := HydrateString(tt.template, &tt.stateParams, nil)

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

func TestDedupeByConcurrentSafety(t *testing.T) {
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
					t.Fatalf("Expected map[string]any item, got %T", item)
				}

				// Extract the field value, handling nested fields
				var fieldValue any
				if tc.name == "Nested field deduplication" {
					metadata, ok := itemDict["metadata"].(map[string]any)
					if !ok {
						t.Fatalf("Expected metadata to be map[string]any, got %T", itemDict["metadata"])
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