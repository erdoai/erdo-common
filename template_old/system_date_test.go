package template

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSystemCurrentDateHydration tests that system.current_date is properly resolved
func TestSystemCurrentDateHydration(t *testing.T) {
	template := `Current date: {{system.current_date}}`
	
	// System data with current date
	systemData := map[string]any{
		"system": map[string]any{
			"current_date": "2025-07-08 18:02:00 BST",
		},
	}
	
	result, err := Hydrate(template, &systemData, nil)
	assert.NoError(t, err)
	assert.Equal(t, "Current date: 2025-07-08 18:02:00 BST", result)
}

// TestSliceEndKeepFirstUserMessageFunction tests the specific function mentioned in logs
func TestSliceEndKeepFirstUserMessageFunction(t *testing.T) {
	template := `{{sliceEndKeepFirstUserMessage "system.messages" 10}}`
	
	// System data with messages
	systemData := map[string]any{
		"system": map[string]any{
			"messages": []any{
				map[string]any{"role": "user", "content": "First user message"},
				map[string]any{"role": "assistant", "content": "Assistant response"},
				map[string]any{"role": "user", "content": "Second user message"},
				map[string]any{"role": "assistant", "content": "Another response"},
			},
		},
	}
	
	result, err := Hydrate(template, &systemData, nil)
	assert.NoError(t, err)
	// Should return string representation of the sliced messages
	assert.True(t, strings.Contains(result.(string), "First user message"))
}

// TestComplexTemplateFromLogs tests a template similar to what's in the actual logs
func TestComplexTemplateFromLogs(t *testing.T) {
	// Template similar to the message_history field from the logs
	template := `{{sliceEndKeepFirstUserMessage "system.messages" 10}}`
	
	// System data structure like what's available
	systemData := map[string]any{
		"system": map[string]any{
			"current_date": "2025-07-08 18:02:00 BST",
			"messages": []any{
				map[string]any{
					"id": "23efbeb4-5731-43ca-b2ee-c4c021e41c5b",
					"role": "user",
					"content": []any{
						map[string]any{
							"id": "aba8ff1f-1c42-421f-aeaa-20eebc4a0433",
							"timestamp": "2025-07-08T18:02:00.001027+01:00",
							"content_type": "text",
							"content": "show aapl ytd",
							"visibility": "visible",
						},
					},
				},
			},
		},
		"invocation_id": "2dd5e08e-b4a9-42d8-8e2b-405d8dfead8c",
		"loops": 0,
	}
	
	result, err := Hydrate(template, &systemData, nil)
	assert.NoError(t, err)
	// Should return the sliced messages without error
	assert.NotNil(t, result)
}