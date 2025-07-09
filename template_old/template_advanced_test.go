package template

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestHydrateStringWithDataReferences tests that function arguments with $.Data or .Data prefixes
// are properly hydrated when used in functions - this was from the original test suite
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
			name:     "Nested value using dot notation",
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

// TestTemplateWithAddkeyFunction tests using the addkey function in templates
// to modify objects by adding new key-value pairs - from original
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
			template: "{{addkey \"object\" \"newKey\" \"value\"}}",
			stateParams: map[string]any{
				"object": map[string]any{"existingKey": "existingValue"},
				"value":  "newValue",
			},
			expected: "map[existingKey:existingValue newKey:newValue]", // Template returns string representation
		},
		{
			name:     "Add key to empty object",
			template: "{{addkey \"emptyObject\" \"firstKey\" \"value\"}}",
			stateParams: map[string]any{
				"emptyObject": map[string]any{},
				"value":       "someValue",
			},
			expected: "map[firstKey:someValue]", // Template returns string representation
		},
		{
			name:     "Object not found",
			template: "{{addkey \"nonExistentObject\" \"key\" \"value\"}}",
			stateParams: map[string]any{
				"value": "someValue",
			},
			expected: "map[]", // Template returns string representation of empty map
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

// TestIncrementCounter tests the incrementCounter and incrementCounterBy functions
func TestIncrementCounter(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(data map[string]any)
		operation   func(data map[string]any, missingKeys *[]string) int
		expected    int
		checkData   map[string]int
	}{
		{
			name:  "Increment non-existent counter",
			setup: func(data map[string]any) {},
			operation: func(data map[string]any, missingKeys *[]string) int {
				return incrementCounter("counter", data, missingKeys)
			},
			expected: 1,
			checkData: map[string]int{
				"counter": 1,
			},
		},
		{
			name: "Increment existing counter",
			setup: func(data map[string]any) {
				data["counter"] = 5
			},
			operation: func(data map[string]any, missingKeys *[]string) int {
				return incrementCounter("counter", data, missingKeys)
			},
			expected: 6,
			checkData: map[string]int{
				"counter": 6,
			},
		},
		{
			name:  "Increment by amount on non-existent counter",
			setup: func(data map[string]any) {},
			operation: func(data map[string]any, missingKeys *[]string) int {
				return incrementCounterBy("counter", 3, data, missingKeys)
			},
			expected: 3,
			checkData: map[string]int{
				"counter": 3,
			},
		},
		{
			name: "Increment by amount on existing counter",
			setup: func(data map[string]any) {
				data["counter"] = 10
			},
			operation: func(data map[string]any, missingKeys *[]string) int {
				return incrementCounterBy("counter", 5, data, missingKeys)
			},
			expected: 15,
			checkData: map[string]int{
				"counter": 15,
			},
		},
		{
			name: "Increment string number",
			setup: func(data map[string]any) {
				data["counter"] = "10"
			},
			operation: func(data map[string]any, missingKeys *[]string) int {
				return incrementCounter("counter", data, missingKeys)
			},
			expected: 11,
			checkData: map[string]int{
				"counter": 11,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]any{}
			missingKeys := []string{}
			
			tt.setup(data)
			result := tt.operation(data, &missingKeys)
			
			assert.Equal(t, tt.expected, result)
			
			// Check that data was modified correctly
			for key, expectedValue := range tt.checkData {
				actualValue, exists := data[key]
				assert.True(t, exists, "Key %s should exist in data", key)
				assert.Equal(t, expectedValue, actualValue, "Value for key %s should match", key)
			}
		})
	}
}

// TestFindByValue tests the findByValue function that searches arrays for specific values
func TestFindByValue(t *testing.T) {
	tests := []struct {
		name        string
		arrayKey    string
		fieldKey    string
		targetValue any
		data        map[string]any
		missingKeys *[]string
		expected    any
	}{
		{
			name:        "Find by string value",
			arrayKey:    "items",
			fieldKey:    "id",
			targetValue: "item2",
			data: map[string]any{
				"items": []any{
					map[string]any{"id": "item1", "name": "First"},
					map[string]any{"id": "item2", "name": "Second"},
					map[string]any{"id": "item3", "name": "Third"},
				},
			},
			missingKeys: &[]string{},
			expected:    map[string]any{"id": "item2", "name": "Second"},
		},
		{
			name:        "Find by numeric value",
			arrayKey:    "items",
			fieldKey:    "count",
			targetValue: 42,
			data: map[string]any{
				"items": []any{
					map[string]any{"id": "a", "count": 10},
					map[string]any{"id": "b", "count": 42},
					map[string]any{"id": "c", "count": 100},
				},
			},
			missingKeys: &[]string{},
			expected:    map[string]any{"id": "b", "count": 42},
		},
		{
			name:        "Find by float stored as int",
			arrayKey:    "items",
			fieldKey:    "value",
			targetValue: 5,
			data: map[string]any{
				"items": []any{
					map[string]any{"id": "a", "value": float64(5)}, // JSON unmarshaling
					map[string]any{"id": "b", "value": float64(10)},
				},
			},
			missingKeys: &[]string{},
			expected:    map[string]any{"id": "a", "value": float64(5)},
		},
		{
			name:        "Not found",
			arrayKey:    "items",
			fieldKey:    "id",
			targetValue: "nonexistent",
			data: map[string]any{
				"items": []any{
					map[string]any{"id": "item1", "name": "First"},
					map[string]any{"id": "item2", "name": "Second"},
				},
			},
			missingKeys: &[]string{},
			expected:    nil,
		},
		{
			name:        "Array not found",
			arrayKey:    "nonexistent",
			fieldKey:    "id",
			targetValue: "value",
			data:        map[string]any{},
			missingKeys: &[]string{},
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findByValue(tt.arrayKey, tt.fieldKey, tt.targetValue, tt.data, tt.missingKeys)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestComplexNestedTemplates tests complex nested template scenarios from original
func TestComplexNestedTemplates(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       string
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "Nested if with function calls",
			template: `{{if gt (len .Data.items) 0}}{{if lt (len .Data.items) 5}}Small list{{else}}Large list{{end}}{{else}}Empty{{end}}`,
			stateParams: map[string]any{
				"items": []any{"a", "b", "c"},
			},
			expected: "Small list",
		},
		{
			name: "Nested if with function calls - large list",
			template: `{{if gt (len .Data.items) 0}}{{if lt (len .Data.items) 5}}Small list{{else}}Large list{{end}}{{else}}Empty{{end}}`,
			stateParams: map[string]any{
				"items": []any{"a", "b", "c", "d", "e", "f"},
			},
			expected: "Large list",
		},
		{
			name: "Nested if with function calls - empty",
			template: `{{if gt (len .Data.items) 0}}{{if lt (len .Data.items) 5}}Small list{{else}}Large list{{end}}{{else}}Empty{{end}}`,
			stateParams: map[string]any{
				"items": []any{},
			},
			expected: "Empty",
		},
		{
			name: "Complex range with conditionals",
			template: `{{range .Data.users}}{{if eq .status "active"}}{{.name}} is active. {{end}}{{end}}`,
			stateParams: map[string]any{
				"users": []any{
					map[string]any{"name": "Alice", "status": "active"},
					map[string]any{"name": "Bob", "status": "inactive"},
					map[string]any{"name": "Carol", "status": "active"},
				},
			},
			expected: "Alice is active. Carol is active. ",
		},
		{
			name: "Nested data access with conditionals",
			template: `{{if .Data.user}}{{if .Data.user.profile}}{{if .Data.user.profile.email}}Email: {{.Data.user.profile.email}}{{else}}No email{{end}}{{else}}No profile{{end}}{{else}}No user{{end}}`,
			stateParams: map[string]any{
				"user": map[string]any{
					"profile": map[string]any{
						"email": "test@example.com",
					},
				},
			},
			expected: "Email: test@example.com",
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

// TestWhitespaceHandling tests whitespace trimming with noop function
func TestWhitespaceHandling(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		stateParams map[string]any
		expected    string
	}{
		{
			name: "Trim whitespace with dash",
			template: `{{- "  trimmed  " -}}`,
			stateParams: map[string]any{},
			expected: "  trimmed  ", // The dash only trims whitespace around the directive, not inside strings
		},
		{
			name: "Noop for whitespace control",
			template: `Line 1{{- noop}}
Line 2`,
			stateParams: map[string]any{},
			expected: "Line 1\nLine 2",
		},
		{
			name: "Complex whitespace control",
			template: `{{- range .Data.items -}}
{{- .name -}},
{{- end -}}`,
			stateParams: map[string]any{
				"items": []any{
					map[string]any{"name": "A"},
					map[string]any{"name": "B"},
					map[string]any{"name": "C"},
				},
			},
			expected: "A,B,C,",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HydrateString(tt.template, &tt.stateParams, nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetAtIndex tests array indexing functionality
func TestGetAtIndex(t *testing.T) {
	tests := []struct {
		name        string
		array       string
		index       any
		data        map[string]any
		missingKeys *[]string
		expected    any
	}{
		{
			name:  "Get at valid index",
			array: "items",
			index: 1,
			data: map[string]any{
				"items": []any{"first", "second", "third"},
			},
			missingKeys: &[]string{},
			expected:    "second",
		},
		{
			name:  "Get at index 0",
			array: "items",
			index: 0,
			data: map[string]any{
				"items": []any{"first", "second", "third"},
			},
			missingKeys: &[]string{},
			expected:    "first",
		},
		{
			name:  "Get at index with string",
			array: "items",
			index: "2",
			data: map[string]any{
				"items": []any{"first", "second", "third"},
			},
			missingKeys: &[]string{},
			expected:    "third",
		},
		{
			name:  "Index out of bounds",
			array: "items",
			index: 10,
			data: map[string]any{
				"items": []any{"first", "second", "third"},
			},
			missingKeys: &[]string{},
			expected:    nil,
		},
		{
			name:  "Negative index",
			array: "items",
			index: -1,
			data: map[string]any{
				"items": []any{"first", "second", "third"},
			},
			missingKeys: &[]string{},
			expected:    nil,
		},
		{
			name:        "Array not found",
			array:       "nonexistent",
			index:       0,
			data:        map[string]any{},
			missingKeys: &[]string{},
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert index to int
			var idx int
			switch v := tt.index.(type) {
			case int:
				idx = v
			case string:
				// Parse string to int
				parsedIdx, err := strconv.Atoi(v)
				if err != nil {
					idx = -1 // Invalid index
				} else {
					idx = parsedIdx
				}
			default:
				idx = -1 // Invalid index
			}
			
			result := getAtIndex(tt.array, idx, tt.data, tt.missingKeys)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestListFunction tests the list function that creates arrays
func TestListFunction(t *testing.T) {
	tests := []struct {
		name     string
		items    []any
		expected []any
	}{
		{
			name:     "Create list with multiple items",
			items:    []any{"a", "b", "c"},
			expected: []any{"a", "b", "c"},
		},
		{
			name:     "Create list with mixed types",
			items:    []any{"string", 123, true, nil},
			expected: []any{"string", 123, true, nil},
		},
		{
			name:     "Create empty list",
			items:    []any{},
			expected: []any{},
		},
		{
			name:     "Create list with single item",
			items:    []any{"single"},
			expected: []any{"single"},
		},
		{
			name:     "Create list with nested structures",
			items:    []any{map[string]any{"key": "value"}, []any{1, 2, 3}},
			expected: []any{map[string]any{"key": "value"}, []any{1, 2, 3}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := list(tt.items...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTemplateWithListFunction tests using list function in templates
func TestTemplateWithListFunction(t *testing.T) {
	// Note: The list function behavior in templates is limited by Go's template engine
	// Variables in function arguments are not automatically resolved
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expected       any
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:        "Create list with literals",
			template:    `{{list "a" "b" "c"}}`,
			stateParams: map[string]any{},
			expected:    "[a b c]",  // When template returns array, it's converted to string
		},
		{
			name:     "Use list in range",
			template: `{{range list "x" "y" "z"}}{{.}},{{end}}`,
			stateParams: map[string]any{},
			expected: "x,y,z,",
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

// TestSliceFunction tests the slice function for array slicing
func TestSliceFunction(t *testing.T) {
	tests := []struct {
		name        string
		array       string
		start       any
		end         any
		data        map[string]any
		missingKeys *[]string
		expected    []any
	}{
		{
			name:  "Slice middle portion",
			array: "items",
			start: 1,
			end:   3,
			data: map[string]any{
				"items": []any{"a", "b", "c", "d", "e"},
			},
			missingKeys: &[]string{},
			expected:    []any{"b", "c"},
		},
		{
			name:  "Slice from beginning",
			array: "items",
			start: 0,
			end:   2,
			data: map[string]any{
				"items": []any{"a", "b", "c", "d", "e"},
			},
			missingKeys: &[]string{},
			expected:    []any{"a", "b"},
		},
		{
			name:  "Slice to end",
			array: "items",
			start: 3,
			end:   5,
			data: map[string]any{
				"items": []any{"a", "b", "c", "d", "e"},
			},
			missingKeys: &[]string{},
			expected:    []any{"d", "e"},
		},
		{
			name:  "Slice with string indices",
			array: "items",
			start: "1",
			end:   "4",
			data: map[string]any{
				"items": []any{"a", "b", "c", "d", "e"},
			},
			missingKeys: &[]string{},
			expected:    []any{"b", "c", "d"},
		},
		{
			name:  "Slice beyond bounds",
			array: "items",
			start: 2,
			end:   10,
			data: map[string]any{
				"items": []any{"a", "b", "c"},
			},
			missingKeys: &[]string{},
			expected:    []any{"c"},
		},
		{
			name:  "Slice with negative start",
			array: "items",
			start: -1,
			end:   2,
			data: map[string]any{
				"items": []any{"a", "b", "c"},
			},
			missingKeys: &[]string{},
			expected:    []any{"a", "b"},
		},
		{
			name:  "Empty slice when start >= end",
			array: "items",
			start: 3,
			end:   2,
			data: map[string]any{
				"items": []any{"a", "b", "c"},
			},
			missingKeys: &[]string{},
			expected:    []any{},
		},
		{
			name:        "Array not found",
			array:       "nonexistent",
			start:       0,
			end:         2,
			data:        map[string]any{},
			missingKeys: &[]string{},
			expected:    []any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert start and end to int
			var start, end int
			
			switch v := tt.start.(type) {
			case int:
				start = v
			case string:
				parsedStart, err := strconv.Atoi(v)
				if err != nil {
					start = 0
				} else {
					start = parsedStart
				}
			default:
				start = 0
			}
			
			switch v := tt.end.(type) {
			case int:
				end = v
			case string:
				parsedEnd, err := strconv.Atoi(v)
				if err != nil {
					end = 0
				} else {
					end = parsedEnd
				}
			default:
				end = 0
			}
			
			result := slice(tt.array, start, end, tt.data, tt.missingKeys)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTimeFormatting tests that the template system handles time values correctly
func TestTimeFormatting(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		template    string
		stateParams map[string]any
		expected    string
	}{
		{
			name:     "Direct time value",
			template: "Time: {{timestamp}}",
			stateParams: map[string]any{
				"timestamp": now.Format(time.RFC3339),
			},
			expected: fmt.Sprintf("Time: %s", now.Format(time.RFC3339)),
		},
		{
			name:     "Time in nested structure",
			template: "Created at: {{event.created_at}}",
			stateParams: map[string]any{
				"event": map[string]any{
					"created_at": now.Format(time.RFC3339),
				},
			},
			expected: fmt.Sprintf("Created at: %s", now.Format(time.RFC3339)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HydrateString(tt.template, &tt.stateParams, nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestUnicodeHandling tests Unicode character handling in templates
func TestUnicodeHandling(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		stateParams map[string]any
		expected    string
	}{
		{
			name:     "Unicode in template",
			template: "Hello {{name}} 👋",
			stateParams: map[string]any{
				"name": "世界",
			},
			expected: "Hello 世界 👋",
		},
		{
			name:     "Emoji handling",
			template: "Status: {{status}}",
			stateParams: map[string]any{
				"status": "✅ Complete",
			},
			expected: "Status: ✅ Complete",
		},
		{
			name:     "Mixed scripts",
			template: "{{greeting}} {{name}}!",
			stateParams: map[string]any{
				"greeting": "Привет",
				"name":     "العالم",
			},
			expected: "Привет العالم!",
		},
		{
			name:     "Truncate Unicode correctly",
			template: "{{truncateString .Data.text 10}}",
			stateParams: map[string]any{
				"text": "Hello 世界 from Tokyo",
			},
			expected: "Hello 世界 f...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HydrateString(tt.template, &tt.stateParams, nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestErrorContextInTemplates tests that errors provide good context
func TestErrorContextInTemplates(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateParams    map[string]any
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:           "Missing required variable",
			template:       "Hello {{name}} from {{location}}!",
			stateParams:    map[string]any{"name": "Alice"},
			expectedError:  true,
			expectedErrMsg: "location",
		},
		{
			name:     "Invalid function arguments",
			template: "Result: {{add \"string\" \"another\"}}",
			stateParams: map[string]any{},
			expectedError: true,
			expectedErrMsg: "expected integer",  // Go template error for type mismatch
		},
		{
			name:           "Nested missing variable",
			template:       "User: {{user.profile.name}}",
			stateParams:    map[string]any{"user": map[string]any{}},
			expectedError:  true,
			expectedErrMsg: "user.profile.name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := HydrateString(tt.template, &tt.stateParams, nil)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSpecialCharacterEscaping tests handling of special characters
func TestSpecialCharacterEscaping(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		stateParams map[string]any
		expected    string
	}{
		{
			name:     "HTML entities",
			template: "{{content}}",
			stateParams: map[string]any{
				"content": "<script>alert('xss')</script>",
			},
			expected: "<script>alert('xss')</script>",
		},
		{
			name:     "Quotes in content",
			template: `He said "{{quote}}"`,
			stateParams: map[string]any{
				"quote": `Hello "world"`,
			},
			expected: `He said "Hello "world""`,
		},
		{
			name:     "Newlines and tabs",
			template: "{{content}}",
			stateParams: map[string]any{
				"content": "Line 1\nLine 2\tTabbed",
			},
			expected: "Line 1\nLine 2\tTabbed",
		},
		{
			name:     "Backslashes",
			template: "Path: {{path}}",
			stateParams: map[string]any{
				"path": `C:\Users\Name\Documents`,
			},
			expected: `Path: C:\Users\Name\Documents`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HydrateString(tt.template, &tt.stateParams, nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestLargeDataSetPerformance tests performance with large data sets
func TestLargeDataSetPerformance(t *testing.T) {
	// Create a large array
	largeArray := make([]any, 1000)
	for i := 0; i < 1000; i++ {
		largeArray[i] = map[string]any{
			"id":    fmt.Sprintf("item-%d", i),
			"value": i,
			"active": i%2 == 0,
		}
	}

	tests := []struct {
		name        string
		template    string
		stateParams map[string]any
		validate    func(t *testing.T, result any)
	}{
		{
			name:     "Process large array with range",
			template: `{{range .Data.items}}{{if .active}}{{.id}},{{end}}{{end}}`,
			stateParams: map[string]any{
				"items": largeArray,
			},
			validate: func(t *testing.T, result any) {
				str := result.(string)
				// Should have 500 active items
				assert.Equal(t, 500, strings.Count(str, ","))
				assert.Contains(t, str, "item-0,")
				assert.Contains(t, str, "item-998,")
				assert.NotContains(t, str, "item-1,")
			},
		},
		{
			name:     "Filter large array",
			template: `{{len (filter "items" "active" true)}}`,
			stateParams: map[string]any{
				"items": largeArray,
			},
			validate: func(t *testing.T, result any) {
				// This will be a string representation of the number
				assert.Equal(t, "500", result)
			},
		},
		{
			name:     "Extract slice from large array",
			template: `{{len (extractSlice "items" "id")}}`,
			stateParams: map[string]any{
				"items": largeArray,
			},
			validate: func(t *testing.T, result any) {
				assert.Equal(t, "1000", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			result, err := Hydrate(tt.template, &tt.stateParams, nil)
			duration := time.Since(start)

			assert.NoError(t, err)
			tt.validate(t, result)
			
			// Performance assertion - should complete within reasonable time
			assert.Less(t, duration, 1*time.Second, "Operation took too long: %v", duration)
		})
	}
}