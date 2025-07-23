package template

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestAutomaticDataParamsAddition(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		stateParams map[string]any
		expected    any
		wantErr     bool
		errContains string
	}{
		{
			name:     "get with nested find should work WITHOUT explicit params",
			template: `{{get "dataset.id" (find "resources" "id" "memory.resource_id")}}`,
			stateParams: map[string]any{
				"resources": []any{
					map[string]any{
						"id": "res-1",
						"dataset": map[string]any{
							"id": "dataset-123",
						},
					},
					map[string]any{
						"id": "res-2",
						"dataset": map[string]any{
							"id": "dataset-456",
						},
					},
				},
				"memory": map[string]any{
					"resource_id": "res-2",
				},
			},
			expected: "dataset-456",
		},
		{
			name:     "simple get without params should auto-add them",
			template: `{{get "user.name"}}`,
			stateParams: map[string]any{
				"user": map[string]any{
					"name": "John Doe",
				},
			},
			expected: "John Doe",
		},
		{
			name:     "find without params should auto-add them",
			template: `{{find "users" "id" "current_user_id"}}`,
			stateParams: map[string]any{
				"users": []any{
					map[string]any{"id": "u1", "name": "Alice"},
					map[string]any{"id": "u2", "name": "Bob"},
				},
				"current_user_id": "u2",
			},
			expected: map[string]any{"id": "u2", "name": "Bob"},
		},
		{
			name:     "addkey with nested get should work without params",
			template: `{{addkey "memory" "dataset_id" (get "dataset_id")}}`,
			stateParams: map[string]any{
				"memory": map[string]any{
					"id": "mem-1",
				},
				"dataset_id": "dataset-789",
			},
			expected: map[string]any{
				"id":         "mem-1",
				"dataset_id": "dataset-789",
			},
		},
		{
			name:     "getAtIndex with nested get for index",
			template: `{{get "name" (getAtIndex "users" (get "current_index"))}}`,
			stateParams: map[string]any{
				"users": []any{
					map[string]any{"name": "Alice"},
					map[string]any{"name": "Bob"},
					map[string]any{"name": "Charlie"},
				},
				"current_index": 1,
			},
			expected: "Bob",
		},
		{
			name:     "template with explicit params should still work",
			template: `{{get "dataset.id" (find "resources" "id" "memory.resource_id" .Data .MissingKeys) .MissingKeys}}`,
			stateParams: map[string]any{
				"resources": []any{
					map[string]any{
						"id": "res-1",
						"dataset": map[string]any{
							"id": "dataset-123",
						},
					},
					map[string]any{
						"id": "res-2",
						"dataset": map[string]any{
							"id": "dataset-456",
						},
					},
				},
				"memory": map[string]any{
					"resource_id": "res-2",
				},
			},
			expected: "dataset-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Hydrate(tt.template, &tt.stateParams, nil)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}