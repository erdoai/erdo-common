package template

import (
	"sort"
	"testing"

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
			name:     "Function call",
			template: "{{toString (get \"count\" .Data .MissingKeys)}}",
			stateParams: map[string]any{
				"count": 42,
			},
			expected: "42",
		},
		{
			name:     "Missing variable",
			template: "Hello, {{missing}}!",
			stateParams: map[string]any{
				"name": "World",
			},
			expectedError:  true,
			expectedErrMsg: "missing",
		},
		{
			name:     "Optional variable present",
			template: "Hello, {{name?}}!",
			stateParams: map[string]any{
				"name": "World",
			},
			expected: "Hello, World!",
		},
		{
			name:     "Optional variable missing",
			template: "Hello, {{missing?}}!",
			stateParams: map[string]any{
				"name": "World",
			},
			expected: "Hello, !",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HydrateString(tt.template, &tt.stateParams, nil)
			
			if tt.expectedError {
				assert.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestHydrateDict(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]any
		stateParams map[string]any
		expected    map[string]any
	}{
		{
			name: "Simple value substitution",
			input: map[string]any{
				"greeting": "Hello, {{name}}!",
				"static":   "unchanged",
			},
			stateParams: map[string]any{
				"name": "World",
			},
			expected: map[string]any{
				"greeting": "Hello, World!",
				"static":   "unchanged",
			},
		},
		{
			name: "Nested dict",
			input: map[string]any{
				"outer": map[string]any{
					"inner": "Value: {{value}}",
				},
			},
			stateParams: map[string]any{
				"value": "test",
			},
			expected: map[string]any{
				"outer": map[string]any{
					"inner": "Value: test",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HydrateDict(tt.input, &tt.stateParams, nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHydrateSlice(t *testing.T) {
	tests := []struct {
		name        string
		input       []any
		stateParams map[string]any
		expected    []any
	}{
		{
			name: "String slice with templates",
			input: []any{
				"Hello {{name}}",
				"Count: {{count}}",
				"static",
			},
			stateParams: map[string]any{
				"name":  "World",
				"count": 42,
			},
			expected: []any{
				"Hello World",
				"Count: 42",
				"static",
			},
		},
		{
			name: "Mixed type slice",
			input: []any{
				"Template: {{value}}",
				42,
				true,
				map[string]any{"nested": "{{value}}"},
			},
			stateParams: map[string]any{
				"value": "test",
			},
			expected: []any{
				"Template: test",
				42,
				true,
				map[string]any{"nested": "test"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HydrateSlice(tt.input, &tt.stateParams, nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseTemplateKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Key
	}{
		{
			name:  "Simple key",
			input: "username",
			expected: Key{
				Key:        "username",
				IsOptional: false,
			},
		},
		{
			name:  "Optional key",
			input: "username?",
			expected: Key{
				Key:        "username",
				IsOptional: true,
			},
		},
		{
			name:  "Key with .Data prefix",
			input: ".Data.username",
			expected: Key{
				Key:        "username",
				IsOptional: false,
			},
		},
		{
			name:  "Optional key with .Data prefix",
			input: ".Data.username?",
			expected: Key{
				Key:        "username",
				IsOptional: true,
			},
		},
		{
			name:  "Key with $.Data prefix",
			input: "$.Data.username",
			expected: Key{
				Key:        "username",
				IsOptional: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseTemplateKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMergeSources(t *testing.T) {
	tests := []struct {
		name     string
		sources  []map[string]any
		expected map[string]any
	}{
		{
			name: "Merge two sources",
			sources: []map[string]any{
				{"a": 1, "b": 2},
				{"c": 3, "d": 4},
			},
			expected: map[string]any{"a": 1, "b": 2, "c": 3, "d": 4},
		},
		{
			name: "Overlapping keys - later wins",
			sources: []map[string]any{
				{"a": 1, "b": 2},
				{"b": 3, "c": 4},
			},
			expected: map[string]any{"a": 1, "b": 3, "c": 4},
		},
		{
			name: "Empty source",
			sources: []map[string]any{
				{"a": 1},
				{},
				{"b": 2},
			},
			expected: map[string]any{"a": 1, "b": 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MergeSources(tt.sources...)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGet(t *testing.T) {
	data := map[string]any{
		"simple": "value",
		"nested": map[string]any{
			"inner": "nested_value",
			"deep": map[string]any{
				"deeper": "deep_value",
			},
		},
		"number": 42,
	}

	tests := []struct {
		name     string
		key      string
		expected any
	}{
		{
			name:     "Simple key",
			key:      "simple",
			expected: "value",
		},
		{
			name:     "Nested key",
			key:      "nested.inner",
			expected: "nested_value",
		},
		{
			name:     "Deep nested key",
			key:      "nested.deep.deeper",
			expected: "deep_value",
		},
		{
			name:     "Number value",
			key:      "number",
			expected: 42,
		},
		{
			name:     "Missing key",
			key:      "missing",
			expected: nil,
		},
		{
			name:     "Missing nested key",
			key:      "nested.missing",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(tt.key, data, nil)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractTemplateParameters(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected []string
	}{
		{
			name:     "Single parameter",
			template: "Hello {{name}}",
			expected: []string{"name"},
		},
		{
			name:     "Multiple parameters",
			template: "{{greeting}} {{name}}, you have {{count}} messages",
			expected: []string{"greeting", "name", "count"},
		},
		{
			name:     "Optional parameter",
			template: "Hello {{name?}}",
			expected: []string{"name"},
		},
		{
			name:     "Python-style parameter",
			template: "Hello %(name)s",
			expected: []string{"name"},
		},
		{
			name:     "Mixed syntax",
			template: "{{greeting}} %(name)s",
			expected: []string{"greeting", "name"},
		},
		{
			name:     "Dot notation",
			template: "{{user.name}} has {{user.count}} items",
			expected: []string{"user.name", "user.count"},
		},
		{
			name:     "No parameters",
			template: "Static text with no templates",
			expected: []string{},
		},
		{
			name:     "Duplicate parameters",
			template: "{{name}} and {{name}} again",
			expected: []string{"name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTemplateParameters(tt.template)
			
			// Sort both slices for comparison since order doesn't matter
			sort.Strings(result)
			sort.Strings(tt.expected)
			
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateTemplateSyntax(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected []string
	}{
		{
			name:     "Valid template",
			template: "Hello {{name}}",
			expected: []string{},
		},
		{
			name:     "Invalid template - unclosed",
			template: "Hello {{name",
			expected: []string{"Invalid template syntax in 'Hello {{name': unclosed template"},
		},
		{
			name:     "Invalid template - extra close",
			template: "Hello name}}",
			expected: []string{"Invalid template syntax in 'Hello name}}': unexpected template close"},
		},
		{
			name:     "Valid optional parameter",
			template: "Hello {{name?}}",
			expected: []string{},
		},
		{
			name:     "Valid python style",
			template: "Hello %(name)s",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateTemplateSyntax(tt.template)
			assert.Equal(t, tt.expected, result)
		})
	}
}