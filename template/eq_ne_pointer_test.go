package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Custom type alias for testing (like DatasetType)
type TestStringAlias string

func TestEqWithPointers(t *testing.T) {
	t.Parallel()

	hello := "hello"
	world := "world"
	empty := ""
	num42 := 42

	tests := []struct {
		name     string
		args     []any
		expected bool
	}{
		{
			name:     "pointer and literal string (equal)",
			args:     []any{&hello, "hello"},
			expected: true,
		},
		{
			name:     "pointer and literal string (not equal)",
			args:     []any{&hello, "world"},
			expected: false,
		},
		{
			name:     "nil pointer and empty string (treated as equal)",
			args:     []any{(*string)(nil), ""},
			expected: true, // nil is treated as empty string
		},
		{
			name:     "nil pointer and nil",
			args:     []any{(*string)(nil), nil},
			expected: true,
		},
		{
			name:     "two nil pointers",
			args:     []any{(*string)(nil), (*int)(nil)},
			expected: true,
		},
		{
			name:     "pointer to empty string and literal empty string",
			args:     []any{&empty, ""},
			expected: true,
		},
		{
			name:     "two pointers to same value",
			args:     []any{&hello, &hello},
			expected: true,
		},
		{
			name:     "two pointers to different values",
			args:     []any{&hello, &world},
			expected: false,
		},
		{
			name:     "pointer to int and literal int (equal)",
			args:     []any{&num42, 42},
			expected: true,
		},
		{
			name:     "pointer to int and literal int (not equal)",
			args:     []any{&num42, 43},
			expected: false,
		},
		{
			name:     "non-pointer values (equal)",
			args:     []any{"test", "test"},
			expected: true,
		},
		{
			name:     "non-pointer values (not equal)",
			args:     []any{"test", "other"},
			expected: false,
		},
		{
			name:     "multiple args all equal",
			args:     []any{&hello, "hello", &hello},
			expected: true,
		},
		{
			name:     "multiple args one not equal",
			args:     []any{&hello, "hello", "world"},
			expected: false,
		},
		{
			name:     "empty args",
			args:     []any{},
			expected: false,
		},
		{
			name:     "single arg",
			args:     []any{"test"},
			expected: true,
		},
		{
			name:     "different types (int and string)",
			args:     []any{&num42, "42"},
			expected: false,
		},
		{
			name:     "type alias and underlying type (equal)",
			args:     []any{TestStringAlias("integration"), "integration"},
			expected: true,
		},
		{
			name:     "type alias and underlying type (not equal)",
			args:     []any{TestStringAlias("integration"), "file"},
			expected: false,
		},
		{
			name:     "pointer to type alias and string literal",
			args:     []any{func() *TestStringAlias { v := TestStringAlias("integration"); return &v }(), "integration"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := eq(tt.args...)
			assert.Equal(t, tt.expected, result, "eq(%v)", tt.args)
		})
	}
}

func TestNeWithPointers(t *testing.T) {
	t.Parallel()

	hello := "hello"
	world := "world"
	empty := ""

	tests := []struct {
		name     string
		args     []any
		expected bool
	}{
		{
			name:     "pointer and literal string (equal)",
			args:     []any{&hello, "hello"},
			expected: false, // ne returns false when equal
		},
		{
			name:     "pointer and literal string (not equal)",
			args:     []any{&hello, "world"},
			expected: true, // ne returns true when not equal
		},
		{
			name:     "nil pointer and empty string (treated as equal)",
			args:     []any{(*string)(nil), ""},
			expected: false, // ne returns false when equal
		},
		{
			name:     "nil pointer and nil",
			args:     []any{(*string)(nil), nil},
			expected: false,
		},
		{
			name:     "pointer to empty string and literal empty string",
			args:     []any{&empty, ""},
			expected: false,
		},
		{
			name:     "two pointers to different values",
			args:     []any{&hello, &world},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ne(tt.args...)
			assert.Equal(t, tt.expected, result, "ne(%v)", tt.args)
		})
	}
}

func TestEqNeInTemplates(t *testing.T) {
	t.Parallel()

	name := "Alice"
	description := "A person"
	empty := ""
	_ = empty // Mark as used since it's needed in test data

	tests := []struct {
		name           string
		template       string
		data           map[string]any
		expected       string
		shouldContain  string
		shouldNotError bool
	}{
		{
			name:     "ne with non-nil pointer and empty string",
			template: `{{- if ne .Data.name ""}}Name: {{.Data.name}}{{end}}`,
			data: map[string]any{
				"name": &name,
			},
			expected:       "Name: Alice",
			shouldNotError: true,
		},
		{
			name:     "ne with nil pointer and empty string",
			template: `{{- if ne .Data.name ""}}Name: {{.Data.name}}{{else}}No name{{end}}`,
			data: map[string]any{
				"name": (*string)(nil),
			},
			expected:       "No name",
			shouldNotError: true,
		},
		{
			name:     "eq with pointer and literal",
			template: `{{- if eq .Data.name "Alice"}}Match!{{end}}`,
			data: map[string]any{
				"name": &name,
			},
			expected:       "Match!",
			shouldNotError: true,
		},
		{
			name:     "eq with nil pointer",
			template: `{{- if eq .Data.name nil}}Nil!{{else}}Not nil{{end}}`,
			data: map[string]any{
				"name": (*string)(nil),
			},
			expected:       "Nil!",
			shouldNotError: true,
		},
		{
			name:     "ne with empty string pointer",
			template: `{{- if ne .Data.name ""}}Has content{{else}}Empty{{end}}`,
			data: map[string]any{
				"name": &empty,
			},
			expected:       "Empty",
			shouldNotError: true,
		},
		{
			name: "real-world use case: dataset description check",
			template: `{{- range $r := .Data.resources}}
{{- if ne $r.description ""}}
Description: {{$r.description}}
{{- end}}
{{- end}}`,
			data: map[string]any{
				"resources": []map[string]any{
					{"description": &description},
					{"description": (*string)(nil)},
					{"description": &empty},
				},
			},
			shouldContain:  "Description: A person",
			shouldNotError: true,
		},
		{
			name: "nested struct pointer field",
			template: `{{- if ne .Data.dataset.name ""}}
Dataset: {{.Data.dataset.name}}
{{- end}}`,
			data: map[string]any{
				"dataset": map[string]any{
					"name": &name,
				},
			},
			shouldContain:  "Dataset: Alice",
			shouldNotError: true,
		},
		{
			name:     "eq with multiple args",
			template: `{{- if eq .Data.a .Data.b .Data.c}}All equal{{else}}Not all equal{{end}}`,
			data: map[string]any{
				"a": &name,
				"b": "Alice",
				"c": &name,
			},
			expected:       "All equal",
			shouldNotError: true,
		},
		{
			name:     "ne works like 'not eq'",
			template: `{{- if ne .Data.a .Data.b}}Different{{else}}Same{{end}}`,
			data: map[string]any{
				"a": &name,
				"b": "Bob",
			},
			expected:       "Different",
			shouldNotError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := Hydrate(tt.template, &tt.data, nil)

			if tt.shouldNotError {
				require.NoError(t, err, "template should not error")
				resultStr, ok := result.(string)
				require.True(t, ok, "result should be a string")

				if tt.expected != "" {
					assert.Equal(t, tt.expected, resultStr)
				}
				if tt.shouldContain != "" {
					assert.Contains(t, resultStr, tt.shouldContain)
				}
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDerefValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name:     "nil pointer",
			input:    (*string)(nil),
			expected: nil,
		},
		{
			name: "non-nil string pointer",
			input: func() *string {
				s := "hello"
				return &s
			}(),
			expected: "hello",
		},
		{
			name: "non-nil int pointer",
			input: func() *int {
				i := 42
				return &i
			}(),
			expected: 42,
		},
		{
			name:     "non-pointer string",
			input:    "world",
			expected: "world",
		},
		{
			name:     "non-pointer int",
			input:    123,
			expected: 123,
		},
		{
			name:     "nil value",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := derefValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToStringWithPointers(t *testing.T) {
	t.Parallel()

	hello := "hello"
	num42 := 42

	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "nil pointer",
			input:    (*string)(nil),
			expected: "",
		},
		{
			name:     "string pointer",
			input:    &hello,
			expected: "hello",
		},
		{
			name:     "int pointer",
			input:    &num42,
			expected: "42",
		},
		{
			name:     "non-pointer string",
			input:    "world",
			expected: "world",
		},
		{
			name:     "non-pointer int",
			input:    123,
			expected: "123",
		},
		{
			name:     "nil value",
			input:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := toString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateStringWithPointers(t *testing.T) {
	t.Parallel()

	longStr := "This is a very long string that needs truncation"
	shortStr := "Short"

	tests := []struct {
		name     string
		input    any
		length   int
		expected string
	}{
		{
			name:     "nil pointer",
			input:    (*string)(nil),
			length:   10,
			expected: "",
		},
		{
			name:     "pointer to long string - truncated",
			input:    &longStr,
			length:   10,
			expected: "This is...",
		},
		{
			name:     "pointer to short string - not truncated",
			input:    &shortStr,
			length:   10,
			expected: "Short",
		},
		{
			name:     "pointer to string - exact length",
			input:    &shortStr,
			length:   5,
			expected: "Short",
		},
		{
			name:     "non-pointer string truncated",
			input:    "Hello World",
			length:   8,
			expected: "Hello...",
		},
		{
			name:     "zero length",
			input:    &longStr,
			length:   0,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := truncateString(tt.input, tt.length)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNilToEmptyStringWithPointers(t *testing.T) {
	t.Parallel()

	hello := "hello"
	num42 := 42

	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "nil pointer",
			input:    (*string)(nil),
			expected: "",
		},
		{
			name:     "string pointer",
			input:    &hello,
			expected: "hello",
		},
		{
			name:     "int pointer",
			input:    &num42,
			expected: "42",
		},
		{
			name:     "non-pointer string",
			input:    "world",
			expected: "world",
		},
		{
			name:     "nil value",
			input:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := nilToEmptyString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPointerDereferencingInTemplates(t *testing.T) {
	t.Parallel()

	name := "Alice"
	longName := "This is a very long name that should be truncated"

	tests := []struct {
		name           string
		template       string
		data           map[string]any
		expected       string
		shouldNotError bool
	}{
		{
			name:           "toString with pointer",
			template:       `Name: {{toString .Data.name}}`,
			data:           map[string]any{"name": &name},
			expected:       "Name: Alice",
			shouldNotError: true,
		},
		{
			name:           "truncateString with pointer",
			template:       `Name: {{truncateString .Data.name 10}}`,
			data:           map[string]any{"name": &longName},
			expected:       "Name: This is...",
			shouldNotError: true,
		},
		{
			name:           "nilToEmptyString with pointer",
			template:       `Name: [{{nilToEmptyString .Data.name}}]`,
			data:           map[string]any{"name": (*string)(nil)},
			expected:       "Name: []",
			shouldNotError: true,
		},
		{
			name:           "direct pointer output in template",
			template:       `Name: {{.Data.name}}`,
			data:           map[string]any{"name": &name},
			expected:       "Name: Alice",
			shouldNotError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := Hydrate(tt.template, &tt.data, nil)

			if tt.shouldNotError {
				require.NoError(t, err, "template should not error")
				resultStr, ok := result.(string)
				require.True(t, ok, "result should be a string")
				assert.Equal(t, tt.expected, resultStr)
			} else {
				require.Error(t, err)
			}
		})
	}
}
