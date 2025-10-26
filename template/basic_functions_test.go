package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLenWithTypeAliases(t *testing.T) {
	t.Parallel()

	// Test that _len handles type aliases like []map[string]any and []map[string]interface{}
	// This was broken after removing the JSON round-trip optimization which converted all types
	tests := []struct {
		name     string
		input    any
		expected int
	}{
		{
			name:     "len with []any",
			input:    []any{"a", "b", "c"},
			expected: 3,
		},
		{
			name: "len with []map[string]any",
			input: []map[string]any{
				{"id": "1"},
				{"id": "2"},
			},
			expected: 2,
		},
		{
			name:     "len with empty []map[string]any",
			input:    []map[string]any{},
			expected: 0,
		},
		{
			name:     "len with map[string]any",
			input:    map[string]any{"a": 1, "b": 2},
			expected: 2,
		},
		{
			name:     "len with string",
			input:    "hello",
			expected: 5,
		},
		{
			name:     "len with nil",
			input:    nil,
			expected: 0,
		},
		{
			name:     "len with []int using reflection",
			input:    []int{1, 2, 3, 4},
			expected: 4,
		},
		{
			name:     "len with []string using reflection",
			input:    []string{"a", "b"},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := _len(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruthyValueWithTypeAliases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "truthy with non-empty []any",
			input:    []any{"a", "b"},
			expected: true,
		},
		{
			name:     "truthy with empty []any",
			input:    []any{},
			expected: false,
		},
		{
			name: "truthy with non-empty []map[string]any",
			input: []map[string]any{
				{"id": "1"},
			},
			expected: true,
		},
		{
			name:     "truthy with empty []map[string]any",
			input:    []map[string]any{},
			expected: false,
		},
		{
			name:     "truthy with non-empty map[string]any",
			input:    map[string]any{"a": 1},
			expected: true,
		},
		{
			name:     "truthy with empty map[string]any",
			input:    map[string]any{},
			expected: false,
		},
		{
			name:     "truthy with true",
			input:    true,
			expected: true,
		},
		{
			name:     "truthy with false",
			input:    false,
			expected: false,
		},
		{
			name:     "truthy with non-empty string",
			input:    "hello",
			expected: true,
		},
		{
			name:     "truthy with empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "truthy with nil",
			input:    nil,
			expected: false,
		},
		{
			name:     "truthy with non-empty []int using reflection",
			input:    []int{1, 2},
			expected: true,
		},
		{
			name:     "truthy with empty []int using reflection",
			input:    []int{},
			expected: false,
		},
		{
			name:     "truthy with non-slice/map type",
			input:    42,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := truthyValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLenAndTruthyInTemplates(t *testing.T) {
	t.Parallel()

	// Integration tests using actual template hydration
	tests := []struct {
		name        string
		template    string
		stateParams map[string]any
		expected    any
	}{
		{
			name:     "len with []map[string]any in template",
			template: "{{len $.Data.items}}",
			stateParams: map[string]any{
				"items": []map[string]any{
					{"id": "1"},
					{"id": "2"},
				},
			},
			expected: 2,
		},
		{
			name:     "truthy with []map[string]any",
			template: "{{if (truthy \"items\" $.Data)}}has items{{else}}no items{{end}}",
			stateParams: map[string]any{
				"items": []map[string]any{
					{"id": "1"},
				},
			},
			expected: "has items",
		},
		{
			name:     "truthy with empty []map[string]any",
			template: "{{if (truthy \"items\" $.Data)}}has items{{else}}no items{{end}}",
			stateParams: map[string]any{
				"items": []map[string]any{},
			},
			expected: "no items",
		},
		{
			name:     "complex template with mixed types",
			template: "{{if (truthy \"memories\" $.Data)}}Found {{len $.Data.memories}} memories{{else}}No memories{{end}}",
			stateParams: map[string]any{
				"memories": []map[string]any{
					{"content": "Memory 1"},
					{"content": "Memory 2"},
				},
			},
			expected: "Found 2 memories",
		},
		{
			name:     "get with []map[string]any index access",
			template: "{{get \"items[0].id\"}}",
			stateParams: map[string]any{
				"items": []map[string]any{
					{"id": "first"},
					{"id": "second"},
				},
			},
			expected: "first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := Hydrate(tt.template, &tt.stateParams, nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
