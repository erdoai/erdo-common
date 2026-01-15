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
		{
			name:     "truthy with loop variable - has key",
			template: "{{range $i, $item := $.Data.items}}{{if (truthy \"name\" $item)}}{{$item.name}},{{end}}{{end}}",
			stateParams: map[string]any{
				"items": []any{
					map[string]any{"name": "Alice"},
					map[string]any{"id": "123"}, // no name key
					map[string]any{"name": "Bob"},
				},
			},
			expected: "Alice,Bob,",
		},
		{
			name:     "truthy with loop variable - empty array",
			template: "{{range $i, $item := $.Data.items}}{{if (truthy \"name\" $item)}}{{$item.name}},{{end}}{{end}}",
			stateParams: map[string]any{
				"items": []any{},
			},
			expected: "",
		},
		{
			name:     "truthy with loop variable - all missing key",
			template: "{{range $i, $item := $.Data.items}}{{if (truthy \"name\" $item)}}{{$item.name}},{{else}}no-name,{{end}}{{end}}",
			stateParams: map[string]any{
				"items": []any{
					map[string]any{"id": "1"},
					map[string]any{"id": "2"},
				},
			},
			expected: "no-name,no-name,",
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

func TestEndsWith(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		suffix   string
		expected bool
	}{
		{
			name:     "ends with .csv",
			input:    "data.csv",
			suffix:   ".csv",
			expected: true,
		},
		{
			name:     "does not end with .csv",
			input:    "data.xlsx",
			suffix:   ".csv",
			expected: false,
		},
		{
			name:     "ends with .xlsx",
			input:    "report.xlsx",
			suffix:   ".xlsx",
			expected: true,
		},
		{
			name:     "ends with .xls",
			input:    "report.xls",
			suffix:   ".xls",
			expected: true,
		},
		{
			name:     "empty string does not end with suffix",
			input:    "",
			suffix:   ".csv",
			expected: false,
		},
		{
			name:     "nil input returns false",
			input:    nil,
			suffix:   ".csv",
			expected: false,
		},
		{
			name:     "pointer to string",
			input:    stringPtr("file.csv"),
			suffix:   ".csv",
			expected: true,
		},
		{
			name:     "nil pointer returns false",
			input:    (*string)(nil),
			suffix:   ".csv",
			expected: false,
		},
		{
			name:     "suffix longer than string",
			input:    "a",
			suffix:   ".csv",
			expected: false,
		},
		{
			name:     "exact match",
			input:    ".csv",
			suffix:   ".csv",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := endsWith(tt.input, tt.suffix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStartsWith(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		prefix   string
		expected bool
	}{
		{
			name:     "starts with prefix_",
			input:    "prefix_file.txt",
			prefix:   "prefix_",
			expected: true,
		},
		{
			name:     "does not start with prefix_",
			input:    "file.txt",
			prefix:   "prefix_",
			expected: false,
		},
		{
			name:     "starts with http://",
			input:    "http://example.com",
			prefix:   "http://",
			expected: true,
		},
		{
			name:     "starts with https://",
			input:    "https://example.com",
			prefix:   "https://",
			expected: true,
		},
		{
			name:     "empty string does not start with prefix",
			input:    "",
			prefix:   "prefix",
			expected: false,
		},
		{
			name:     "nil input returns false",
			input:    nil,
			prefix:   "prefix",
			expected: false,
		},
		{
			name:     "pointer to string",
			input:    stringPtr("prefix_data"),
			prefix:   "prefix_",
			expected: true,
		},
		{
			name:     "nil pointer returns false",
			input:    (*string)(nil),
			prefix:   "prefix",
			expected: false,
		},
		{
			name:     "prefix longer than string",
			input:    "a",
			prefix:   "prefix",
			expected: false,
		},
		{
			name:     "exact match",
			input:    "prefix",
			prefix:   "prefix",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := startsWith(tt.input, tt.prefix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEndsWithInTemplates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		template    string
		stateParams map[string]any
		expected    any
	}{
		{
			name:     "endsWith .csv",
			template: "{{endsWith $.Data.filename \".csv\"}}",
			stateParams: map[string]any{
				"filename": "data.csv",
			},
			expected: true,
		},
		{
			name:     "endsWith .xlsx or .xls (xlsx)",
			template: "{{or (endsWith $.Data.filename \".xlsx\") (endsWith $.Data.filename \".xls\")}}",
			stateParams: map[string]any{
				"filename": "report.xlsx",
			},
			expected: true, // Custom or function returns bool (not string) for proper type preservation
		},
		{
			name:     "endsWith .xlsx or .xls (xls)",
			template: "{{or (endsWith $.Data.filename \".xlsx\") (endsWith $.Data.filename \".xls\")}}",
			stateParams: map[string]any{
				"filename": "report.xls",
			},
			expected: true, // Custom or function returns bool (not string) for proper type preservation
		},
		{
			name:     "does not end with .csv",
			template: "{{endsWith $.Data.filename \".csv\"}}",
			stateParams: map[string]any{
				"filename": "data.txt",
			},
			expected: false,
		},
		{
			name:     "conditional based on endsWith",
			template: "{{if endsWith $.Data.filename \".csv\"}}CSV file{{else}}Not CSV{{end}}",
			stateParams: map[string]any{
				"filename": "data.csv",
			},
			expected: "CSV file",
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

// Helper function to create pointer to string
func stringPtr(s string) *string {
	return &s
}

func TestPrepend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		item     any
		slice    any
		expected []any
	}{
		{
			name:     "prepend to []any",
			item:     "first",
			slice:    []any{"second", "third"},
			expected: []any{"first", "second", "third"},
		},
		{
			name:     "prepend to []string",
			item:     "first",
			slice:    []string{"second", "third"},
			expected: []any{"first", "second", "third"},
		},
		{
			name:     "prepend to nil",
			item:     "only",
			slice:    nil,
			expected: []any{"only"},
		},
		{
			name:     "prepend to empty slice",
			item:     "first",
			slice:    []any{},
			expected: []any{"first"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := prepend(tt.item, tt.slice)
			assert.Equal(t, tt.expected, result)
		})
	}
}
