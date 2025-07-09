package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCoalesceOriginalErrorScenario tests the exact scenario from the original error logs
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
			expected: "0", // When used in templates, numbers get converted to strings
		},
		{
			name:     "Original error scenario - code_retry_loops exists",
			template: `{{coalesce "code_retry_loops?" 0}}`,
			stateParams: map[string]any{
				"code_retry_loops": 1, // After one retry
			},
			expected: "1", // When used in templates, numbers get converted to strings
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

// TestCoalesceFunctionDirect tests the coalesce function directly with correct types
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
			assert.Equal(t, tt.expectedType, func() string {
				switch result.(type) {
				case int:
					return "int"
				case string:
					return "string"
				case float64:
					return "float64"
				default:
					return "unknown"
				}
			}())
		})
	}
}

// TestBasicFunctionInTemplate tests that basic functions work correctly in templates
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
			expected:    "5",  // Template returns string
		},
		{
			name:        "add two numbers",
			template:    `{{add 5 3}}`,
			stateParams: map[string]any{},
			expected:    "8",  // Template returns string
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
			expected:    "true",  // Template returns string
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
			expected:    "hello...",
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

// TestTemplateWithSliceEndKeepFirstUserMessage tests the specialized message function
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
			expected: "[map[content:u3 role:user] map[content:a role:assistant] map[content:a role:assistant] map[content:a role:assistant]]",
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
			expected: "[map[content:User message 1 role:user] map[content:Assistant message 2 role:assistant]]",
		},
		{
			name:        "Array not found",
			template:    "{{sliceEndKeepFirstUserMessage \"nonexistent\" 3}}",
			stateParams: map[string]any{},
			expected:    "[]",
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

// TestMergeFunction tests the merge function that combines arrays
func TestMergeFunction(t *testing.T) {
	tests := []struct {
		name        string
		array1      string
		array2      string
		data        map[string]any
		missingKeys *[]string
		expected    []any
	}{
		{
			name:   "Merge two arrays",
			array1: "arr1",
			array2: "arr2",
			data: map[string]any{
				"arr1": []any{"a", "b"},
				"arr2": []any{"c", "d"},
			},
			missingKeys: &[]string{},
			expected:    []any{"a", "b", "c", "d"},
		},
		{
			name:   "Merge with nil array",
			array1: "arr1",
			array2: "missing",
			data: map[string]any{
				"arr1": []any{"a", "b"},
			},
			missingKeys: &[]string{},
			expected:    []any{"a", "b"},
		},
		{
			name:   "Merge with single values",
			array1: "val1",
			array2: "val2",
			data: map[string]any{
				"val1": "single1",
				"val2": "single2",
			},
			missingKeys: &[]string{},
			expected:    []any{"single1", "single2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := merge(tt.array1, tt.array2, tt.data, tt.missingKeys)
			assert.Equal(t, tt.expected, result)
		})
	}
}