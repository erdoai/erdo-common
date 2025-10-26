package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStructToMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name:     "primitive int",
			input:    42,
			expected: 42,
		},
		{
			name:     "primitive string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "nil pointer",
			input:    (*string)(nil),
			expected: nil,
		},
		{
			name: "simple struct",
			input: struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}{
				ID:   1,
				Name: "test",
			},
			expected: map[string]any{
				"ID":   1,
				"Name": "test",
			},
		},
		{
			name: "struct with pointer fields",
			input: struct {
				ID   int     `json:"id"`
				Name *string `json:"name"`
			}{
				ID:   1,
				Name: stringPtr("test"),
			},
			expected: map[string]any{
				"ID":   1,
				"Name": "test",
			},
		},
		{
			name: "struct with nil pointer field",
			input: struct {
				ID   int     `json:"id"`
				Name *string `json:"name"`
			}{
				ID:   1,
				Name: nil,
			},
			expected: map[string]any{
				"ID":   1,
				"Name": nil,
			},
		},
		{
			name: "nested struct",
			input: struct {
				ID      int    `json:"id"`
				Dataset *struct {
					Name string `json:"name"`
					Type string `json:"type"`
				} `json:"dataset"`
			}{
				ID: 1,
				Dataset: &struct {
					Name string `json:"name"`
					Type string `json:"type"`
				}{
					Name: "test",
					Type: "file",
				},
			},
			expected: map[string]any{
				"ID": 1,
				"Dataset": map[string]any{
					"Name": "test",
					"Type": "file",
				},
			},
		},
		{
			name: "slice of primitives",
			input: []int{1, 2, 3},
			expected: []any{1, 2, 3},
		},
		{
			name: "slice of structs",
			input: []struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}{
				{ID: 1, Name: "first"},
				{ID: 2, Name: "second"},
			},
			expected: []any{
				map[string]any{"ID": 1, "Name": "first"},
				map[string]any{"ID": 2, "Name": "second"},
			},
		},
		{
			name: "map with primitive values",
			input: map[string]int{
				"a": 1,
				"b": 2,
			},
			expected: map[string]any{
				"a": 1,
				"b": 2,
			},
		},
		{
			name: "map with struct values",
			input: map[string]struct {
				ID int `json:"id"`
			}{
				"first": {ID: 1},
				"second": {ID: 2},
			},
			expected: map[string]any{
				"first":  map[string]any{"ID": 1},
				"second": map[string]any{"ID": 2},
			},
		},
		{
			name: "struct with unexported fields (should be skipped)",
			input: struct {
				ID   int    `json:"id"`
				name string // unexported
			}{
				ID:   1,
				name: "test",
			},
			expected: map[string]any{
				"ID": 1,
			},
		},
		{
			name: "complex nested structure",
			input: struct {
				ID        int                `json:"id"`
				Resources []*struct {
					ID      int    `json:"id"`
					Dataset *struct {
						Name       *string           `json:"name"`
						Type       string            `json:"type"`
						Parameters map[string]string `json:"parameters"`
					} `json:"dataset"`
				} `json:"resources"`
			}{
				ID: 1,
				Resources: []*struct {
					ID      int    `json:"id"`
					Dataset *struct {
						Name       *string           `json:"name"`
						Type       string            `json:"type"`
						Parameters map[string]string `json:"parameters"`
					} `json:"dataset"`
				}{
					{
						ID: 10,
						Dataset: &struct {
							Name       *string           `json:"name"`
							Type       string            `json:"type"`
							Parameters map[string]string `json:"parameters"`
						}{
							Name: stringPtr("my-dataset"),
							Type: "integration",
							Parameters: map[string]string{
								"key1": "value1",
								"key2": "value2",
							},
						},
					},
				},
			},
			expected: map[string]any{
				"ID": 1,
				"Resources": []any{
					map[string]any{
						"ID": 10,
						"Dataset": map[string]any{
							"Name": "my-dataset",
							"Type": "integration",
							"Parameters": map[string]any{
								"key1": "value1",
								"key2": "value2",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := StructToMap(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
