package template

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
)

// TestRawBehaviorCompleteness verifies that raw fields are completely ignored during hydration
func TestRawBehaviorCompleteness(t *testing.T) {
	// Test case from the user's error logs
	data := map[string]any{
		"tools": []any{
			map[string]any{
				"action_type": "bot.invoke",
				"parameters": map[string]any{
					"bot_name": "data analyst",
					"parameters": map[string]any{
						"query": "{{query?}}",
						"url": "{{url}}",
						"system_current_date": "{{system.current_date}}",
					},
				},
			},
		},
		"system_prompt": "Current date: {{system.current_date}}",
	}
	
	// Mark tools[].parameters as raw
	behavior := &map[string]any{
		"tools": map[string]any{
			"parameters": "raw",
		},
	}
	
	// Extract keys that need hydration
	keys := FindTemplateKeyStringsToHydrate(data, false, behavior)
	
	// Should only find system.current_date from system_prompt
	// Should NOT find query, url, or system.current_date from tools[].parameters
	assert.Equal(t, []string{"system.current_date"}, keys)
	
	// Available data for hydration
	availableData := map[string]any{
		"system": map[string]any{
			"current_date": "2025-07-09",
		},
		"query": "test query",
		"url": "https://example.com",
	}
	
	// Hydrate the data
	result, err := Hydrate(data, &availableData, behavior)
	assert.NoError(t, err)
	
	// Check the result
	resultMap := result.(map[string]any)
	
	// system_prompt should be hydrated
	assert.Equal(t, "Current date: 2025-07-09", resultMap["system_prompt"])
	
	// tools[].parameters should NOT be hydrated
	tools := resultMap["tools"].([]any)
	tool := tools[0].(map[string]any)
	params := tool["parameters"].(map[string]any)
	
	// bot_name should remain as is
	assert.Equal(t, "data analyst", params["bot_name"])
	
	// The nested parameters should still contain template strings
	nestedParams := params["parameters"].(map[string]any)
	assert.Equal(t, "{{query?}}", nestedParams["query"])
	assert.Equal(t, "{{url}}", nestedParams["url"])
	assert.Equal(t, "{{system.current_date}}", nestedParams["system_current_date"])
}

// TestRawBehaviorDeepNesting tests that raw behavior works no matter how deep the nesting
func TestRawBehaviorDeepNesting(t *testing.T) {
	data := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"raw_field": map[string]any{
						"level4": map[string]any{
							"level5": map[string]any{
								"template": "{{should_not_hydrate}}",
							},
						},
					},
					"normal_field": "{{should_hydrate}}",
				},
			},
		},
	}
	
	// Mark only the raw_field as raw
	behavior := &map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"raw_field": "raw",
				},
			},
		},
	}
	
	// Extract keys
	keys := FindTemplateKeyStringsToHydrate(data, false, behavior)
	
	// Should only find should_hydrate, not should_not_hydrate
	assert.Equal(t, []string{"should_hydrate"}, keys)
	
	// Hydrate
	availableData := map[string]any{
		"should_not_hydrate": "FAIL",
		"should_hydrate": "SUCCESS",
	}
	
	result, err := Hydrate(data, &availableData, behavior)
	assert.NoError(t, err)
	
	// Check the deeply nested raw field is untouched
	resultMap := result.(map[string]any)
	level1 := resultMap["level1"].(map[string]any)
	level2 := level1["level2"].(map[string]any)
	level3 := level2["level3"].(map[string]any)
	rawField := level3["raw_field"].(map[string]any)
	level4 := rawField["level4"].(map[string]any)
	level5 := level4["level5"].(map[string]any)
	
	// Should still be the template string
	assert.Equal(t, "{{should_not_hydrate}}", level5["template"])
	
	// But normal_field should be hydrated
	assert.Equal(t, "SUCCESS", level3["normal_field"])
}