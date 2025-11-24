package template

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests for SQL null type handling in templates.
// SQL null types like sql.NullString get serialized to maps like:
// map[string]any{"String": "value", "Valid": true}
// These tests ensure template functions handle both the struct and map forms.

func TestUnwrapNullValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         any
		expectedValue any
		expectedValid bool
	}{
		// Map form (JSON-serialized null types)
		{
			name:          "valid null string map",
			input:         map[string]any{"String": "hello world", "Valid": true},
			expectedValue: "hello world",
			expectedValid: true,
		},
		{
			name:          "invalid null string map",
			input:         map[string]any{"String": "", "Valid": false},
			expectedValue: nil,
			expectedValid: false,
		},
		{
			name:          "valid null int64 map",
			input:         map[string]any{"Int64": int64(42), "Valid": true},
			expectedValue: int64(42),
			expectedValid: true,
		},
		{
			name:          "invalid null int64 map",
			input:         map[string]any{"Int64": int64(0), "Valid": false},
			expectedValue: nil,
			expectedValid: false,
		},
		{
			name:          "valid null bool map",
			input:         map[string]any{"Bool": true, "Valid": true},
			expectedValue: true,
			expectedValid: true,
		},
		{
			name:          "valid null float64 map",
			input:         map[string]any{"Float64": 3.14, "Valid": true},
			expectedValue: 3.14,
			expectedValid: true,
		},
		// Regular maps (not null types) should pass through
		{
			name:          "regular map without Valid field",
			input:         map[string]any{"foo": "bar", "baz": 123},
			expectedValue: map[string]any{"foo": "bar", "baz": 123},
			expectedValid: true,
		},
		{
			name:          "regular map with non-bool Valid",
			input:         map[string]any{"String": "hello", "Valid": "true"},
			expectedValue: map[string]any{"String": "hello", "Valid": "true"},
			expectedValid: true,
		},
		// Struct form (actual sql.Null* types)
		{
			name:          "valid sql.NullString struct",
			input:         sql.NullString{String: "test value", Valid: true},
			expectedValue: "test value",
			expectedValid: true,
		},
		{
			name:          "invalid sql.NullString struct",
			input:         sql.NullString{String: "", Valid: false},
			expectedValue: nil,
			expectedValid: false,
		},
		{
			name:          "valid sql.NullInt64 struct",
			input:         sql.NullInt64{Int64: 12345, Valid: true},
			expectedValue: int64(12345),
			expectedValid: true,
		},
		{
			name:          "invalid sql.NullInt64 struct",
			input:         sql.NullInt64{Int64: 0, Valid: false},
			expectedValue: nil,
			expectedValid: false,
		},
		{
			name:          "valid sql.NullBool struct",
			input:         sql.NullBool{Bool: true, Valid: true},
			expectedValue: true,
			expectedValid: true,
		},
		{
			name:          "valid sql.NullFloat64 struct",
			input:         sql.NullFloat64{Float64: 2.718, Valid: true},
			expectedValue: 2.718,
			expectedValid: true,
		},
		// Regular values should pass through unchanged
		{
			name:          "plain string",
			input:         "hello",
			expectedValue: "hello",
			expectedValid: true,
		},
		{
			name:          "plain int",
			input:         42,
			expectedValue: 42,
			expectedValid: true,
		},
		{
			name:          "nil",
			input:         nil,
			expectedValue: nil,
			expectedValid: false,
		},
		{
			name:          "pointer to string",
			input:         stringPtr("pointer value"),
			expectedValue: "pointer value",
			expectedValid: true,
		},
		{
			name:          "nil pointer",
			input:         (*string)(nil),
			expectedValue: nil,
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			value, valid := unwrapNullValue(tt.input)
			assert.Equal(t, tt.expectedValid, valid, "valid mismatch")
			assert.Equal(t, tt.expectedValue, value, "value mismatch")
		})
	}
}

func TestToStringWithNullTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected string
	}{
		// Map form (JSON-serialized)
		{
			name:     "valid null string map",
			input:    map[string]any{"String": "hello world", "Valid": true},
			expected: "hello world",
		},
		{
			name:     "invalid null string map",
			input:    map[string]any{"String": "ignored", "Valid": false},
			expected: "",
		},
		{
			name:     "valid null int64 map",
			input:    map[string]any{"Int64": int64(42), "Valid": true},
			expected: "42",
		},
		// Struct form
		{
			name:     "valid sql.NullString",
			input:    sql.NullString{String: "test string", Valid: true},
			expected: "test string",
		},
		{
			name:     "invalid sql.NullString",
			input:    sql.NullString{String: "ignored", Valid: false},
			expected: "",
		},
		// Regular values
		{
			name:     "plain string",
			input:    "regular string",
			expected: "regular string",
		},
		{
			name:     "nil",
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

func TestNilToEmptyStringWithNullTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected string
	}{
		// Map form
		{
			name:     "valid null string map",
			input:    map[string]any{"String": "content", "Valid": true},
			expected: "content",
		},
		{
			name:     "invalid null string map",
			input:    map[string]any{"String": "ignored", "Valid": false},
			expected: "",
		},
		// Struct form
		{
			name:     "valid sql.NullString",
			input:    sql.NullString{String: "value", Valid: true},
			expected: "value",
		},
		{
			name:     "invalid sql.NullString",
			input:    sql.NullString{Valid: false},
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

func TestTruncateStringWithNullTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		n        int
		expected string
	}{
		// Map form
		{
			name:     "valid null string map truncated",
			input:    map[string]any{"String": "this is a very long string", "Valid": true},
			n:        10,
			expected: "this is...",
		},
		{
			name:     "invalid null string map",
			input:    map[string]any{"String": "ignored", "Valid": false},
			n:        10,
			expected: "",
		},
		{
			name:     "valid null string map not truncated",
			input:    map[string]any{"String": "short", "Valid": true},
			n:        10,
			expected: "short",
		},
		// Struct form
		{
			name:     "valid sql.NullString truncated",
			input:    sql.NullString{String: "a very long string here", Valid: true},
			n:        10,
			expected: "a very ...",
		},
		{
			name:     "invalid sql.NullString",
			input:    sql.NullString{String: "ignored", Valid: false},
			n:        10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := truncateString(tt.input, tt.n)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruthyValueWithNullTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		// Map form
		{
			name:     "valid null string with content",
			input:    map[string]any{"String": "content", "Valid": true},
			expected: true,
		},
		{
			name:     "valid null string empty",
			input:    map[string]any{"String": "", "Valid": true},
			expected: false,
		},
		{
			name:     "invalid null string",
			input:    map[string]any{"String": "ignored", "Valid": false},
			expected: false,
		},
		// Struct form
		{
			name:     "valid sql.NullString with content",
			input:    sql.NullString{String: "hello", Valid: true},
			expected: true,
		},
		{
			name:     "valid sql.NullString empty",
			input:    sql.NullString{String: "", Valid: true},
			expected: false,
		},
		{
			name:     "invalid sql.NullString",
			input:    sql.NullString{Valid: false},
			expected: false,
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

func TestLenWithNullTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected int
	}{
		// Map form (string length)
		{
			name:     "valid null string length",
			input:    map[string]any{"String": "hello", "Valid": true},
			expected: 5,
		},
		{
			name:     "invalid null string",
			input:    map[string]any{"String": "ignored", "Valid": false},
			expected: 0,
		},
		// Struct form
		{
			name:     "valid sql.NullString length",
			input:    sql.NullString{String: "hello world", Valid: true},
			expected: 11,
		},
		{
			name:     "invalid sql.NullString",
			input:    sql.NullString{Valid: false},
			expected: 0,
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

func TestEqWithNullTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []any
		expected bool
	}{
		// Map form
		{
			name:     "valid null string equals string",
			args:     []any{map[string]any{"String": "hello", "Valid": true}, "hello"},
			expected: true,
		},
		{
			name:     "valid null string not equals string",
			args:     []any{map[string]any{"String": "hello", "Valid": true}, "world"},
			expected: false,
		},
		{
			name:     "invalid null string equals empty string",
			args:     []any{map[string]any{"String": "ignored", "Valid": false}, ""},
			expected: true,
		},
		// Struct form
		{
			name:     "valid sql.NullString equals string",
			args:     []any{sql.NullString{String: "test", Valid: true}, "test"},
			expected: true,
		},
		{
			name:     "invalid sql.NullString equals empty string",
			args:     []any{sql.NullString{Valid: false}, ""},
			expected: true,
		},
		// Two null types
		{
			name: "two valid null strings equal",
			args: []any{
				map[string]any{"String": "same", "Valid": true},
				sql.NullString{String: "same", Valid: true},
			},
			expected: true,
		},
		{
			name: "two invalid null types equal",
			args: []any{
				map[string]any{"String": "ignored", "Valid": false},
				sql.NullString{Valid: false},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := eq(tt.args...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNeWithNullTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []any
		expected bool
	}{
		// Map form
		{
			name:     "valid null string not equals different string",
			args:     []any{map[string]any{"String": "hello", "Valid": true}, "world"},
			expected: true,
		},
		{
			name:     "valid null string not equals same string",
			args:     []any{map[string]any{"String": "hello", "Valid": true}, "hello"},
			expected: false,
		},
		{
			name:     "invalid null string not equals empty - should be false",
			args:     []any{map[string]any{"String": "ignored", "Valid": false}, ""},
			expected: false,
		},
		{
			name:     "invalid null string not equals non-empty",
			args:     []any{map[string]any{"String": "ignored", "Valid": false}, "something"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ne(tt.args...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEndsWithNullTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		suffix   string
		expected bool
	}{
		// Map form
		{
			name:     "valid null string ends with suffix",
			input:    map[string]any{"String": "file.csv", "Valid": true},
			suffix:   ".csv",
			expected: true,
		},
		{
			name:     "invalid null string",
			input:    map[string]any{"String": "file.csv", "Valid": false},
			suffix:   ".csv",
			expected: false,
		},
		// Struct form
		{
			name:     "valid sql.NullString ends with suffix",
			input:    sql.NullString{String: "document.pdf", Valid: true},
			suffix:   ".pdf",
			expected: true,
		},
		{
			name:     "invalid sql.NullString",
			input:    sql.NullString{Valid: false},
			suffix:   ".pdf",
			expected: false,
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

func TestStartsWithNullTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		prefix   string
		expected bool
	}{
		// Map form
		{
			name:     "valid null string starts with prefix",
			input:    map[string]any{"String": "http://example.com", "Valid": true},
			prefix:   "http://",
			expected: true,
		},
		{
			name:     "invalid null string",
			input:    map[string]any{"String": "http://example.com", "Valid": false},
			prefix:   "http://",
			expected: false,
		},
		// Struct form
		{
			name:     "valid sql.NullString starts with prefix",
			input:    sql.NullString{String: "https://api.example.com", Valid: true},
			prefix:   "https://",
			expected: true,
		},
		{
			name:     "invalid sql.NullString",
			input:    sql.NullString{Valid: false},
			prefix:   "https://",
			expected: false,
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

func TestNullTypesInTemplates(t *testing.T) {
	t.Parallel()

	// Note: Direct $.Data.foo access gives raw values. Use functions like toString,
	// truncateString, nilToEmptyString, etc. to properly handle null types.
	tests := []struct {
		name        string
		template    string
		stateParams map[string]any
		expected    any
	}{
		{
			name:     "valid null string using toString",
			template: "{{toString $.Data.description}}",
			stateParams: map[string]any{
				"description": map[string]any{"String": "This is a description", "Valid": true},
			},
			expected: "This is a description",
		},
		{
			name:     "invalid null string using toString",
			template: "{{toString $.Data.description}}",
			stateParams: map[string]any{
				"description": map[string]any{"String": "ignored", "Valid": false},
			},
			expected: "",
		},
		{
			name:     "valid null string using nilToEmptyString",
			template: "{{nilToEmptyString $.Data.description}}",
			stateParams: map[string]any{
				"description": map[string]any{"String": "This is a description", "Valid": true},
			},
			expected: "This is a description",
		},
		{
			name:     "null type in conditional with ne (unwraps automatically)",
			template: "{{if ne $.Data.api_version \"\"}}v{{toString $.Data.api_version}}{{else}}no version{{end}}",
			stateParams: map[string]any{
				"api_version": map[string]any{"String": "2.0", "Valid": true},
			},
			expected: "v2.0",
		},
		{
			name:     "invalid null type in conditional (ne treats invalid as empty string)",
			template: "{{if ne $.Data.api_version \"\"}}v{{toString $.Data.api_version}}{{else}}no version{{end}}",
			stateParams: map[string]any{
				"api_version": map[string]any{"String": "ignored", "Valid": false},
			},
			expected: "no version",
		},
		{
			name:     "truncateString with null type",
			template: "{{truncateString $.Data.description 20}}",
			stateParams: map[string]any{
				"description": map[string]any{"String": "This is a very long description that needs truncation", "Valid": true},
			},
			expected: "This is a very lo...",
		},
		{
			name:     "truncateString with invalid null type",
			template: "{{truncateString $.Data.description 20}}",
			stateParams: map[string]any{
				"description": map[string]any{"String": "ignored", "Valid": false},
			},
			expected: "",
		},
		{
			name:     "len with null type string",
			template: "{{len $.Data.name}}",
			stateParams: map[string]any{
				"name": map[string]any{"String": "hello", "Valid": true},
			},
			expected: 5,
		},
		{
			name:     "len with invalid null type",
			template: "{{len $.Data.name}}",
			stateParams: map[string]any{
				"name": map[string]any{"String": "ignored", "Valid": false},
			},
			expected: 0,
		},
		{
			name:     "truthy with valid null type",
			template: "{{if truthyValue $.Data.name}}has name{{else}}no name{{end}}",
			stateParams: map[string]any{
				"name": map[string]any{"String": "test", "Valid": true},
			},
			expected: "has name",
		},
		{
			name:     "truthy with invalid null type",
			template: "{{if truthyValue $.Data.name}}has name{{else}}no name{{end}}",
			stateParams: map[string]any{
				"name": map[string]any{"String": "ignored", "Valid": false},
			},
			expected: "no name",
		},
		{
			name:     "endsWith with null type",
			template: "{{endsWith $.Data.filename \".csv\"}}",
			stateParams: map[string]any{
				"filename": map[string]any{"String": "data.csv", "Valid": true},
			},
			expected: true,
		},
		{
			name:     "startsWith with null type",
			template: "{{startsWith $.Data.url \"https://\"}}",
			stateParams: map[string]any{
				"url": map[string]any{"String": "https://example.com", "Valid": true},
			},
			expected: true,
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
