package template

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
)

// TestRawParameterBehavior tests that fields marked as "raw" are not hydrated
func TestRawParameterBehavior(t *testing.T) {
	// Test data structure similar to the actual use case
	data := map[string]any{
		"tools": []any{
			map[string]any{
				"action_type": "bot.invoke",
				"parameters": map[string]any{
					"bot_name": "data analyst",
					"parameters": map[string]any{
						"query": "{{query}}",
						"context": "{{additional_context?}}",
					},
				},
			},
		},
		"system_prompt": "Date: {{system.current_date}}",
	}
	
	// Hydration behavior that marks tools[].parameters as raw
	behavior := &map[string]any{
		"tools": map[string]any{
			"parameters": "raw",
		},
	}
	
	// Available data for hydration
	availableData := map[string]any{
		"system": map[string]any{
			"current_date": "2025-07-09",
		},
		"query": "What is the weather?",
		"additional_context": "User is in London",
	}
	
	// Hydrate with behavior
	result, err := Hydrate(data, &availableData, behavior)
	assert.NoError(t, err)
	
	resultMap := result.(map[string]any)
	tools := resultMap["tools"].([]any)
	tool := tools[0].(map[string]any)
	params := tool["parameters"].(map[string]any)
	
	// The parameters field should NOT be hydrated (marked as raw)
	assert.Equal(t, "data analyst", params["bot_name"])
	
	// The nested parameters should still contain template strings
	nestedParams := params["parameters"].(map[string]any)
	assert.Equal(t, "{{query}}", nestedParams["query"])
	assert.Equal(t, "{{additional_context?}}", nestedParams["context"])
	
	// But system_prompt should be hydrated
	assert.Equal(t, "Date: 2025-07-09", resultMap["system_prompt"])
}

// TestNestedRawBehavior tests that only the specific field marked as raw is not hydrated
func TestNestedRawBehavior(t *testing.T) {
	// Complex nested structure
	data := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"raw_field": map[string]any{
					"template": "{{should_not_be_hydrated}}",
				},
				"normal_field": "{{should_be_hydrated}}",
			},
		},
	}
	
	// Mark only level1.level2.raw_field as raw
	behavior := &map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"raw_field": "raw",
			},
		},
	}
	
	// Available data
	availableData := map[string]any{
		"should_not_be_hydrated": "FAIL",
		"should_be_hydrated": "SUCCESS",
	}
	
	// Hydrate
	result, err := Hydrate(data, &availableData, behavior)
	assert.NoError(t, err)
	
	resultMap := result.(map[string]any)
	level1 := resultMap["level1"].(map[string]any)
	level2 := level1["level2"].(map[string]any)
	
	// raw_field should NOT be hydrated
	rawField := level2["raw_field"].(map[string]any)
	assert.Equal(t, "{{should_not_be_hydrated}}", rawField["template"])
	
	// normal_field SHOULD be hydrated
	assert.Equal(t, "SUCCESS", level2["normal_field"])
}

// TestKeyExtractionWithRawBehavior verifies that raw fields don't contribute keys
func TestKeyExtractionWithRawBehavior(t *testing.T) {
	// Structure with some raw fields
	data := map[string]any{
		"tools": []any{
			map[string]any{
				"parameters": map[string]any{
					"query": "{{query}}",
					"url": "{{url}}",
				},
			},
		},
		"prompt": "{{system.current_date}}",
	}
	
	// Mark tools[].parameters as raw
	behavior := &map[string]any{
		"tools": map[string]any{
			"parameters": "raw",
		},
	}
	
	// Extract keys
	keys := FindTemplateKeyStringsToHydrate(data, false, behavior)
	
	// Should only find system.current_date, not query or url
	assert.Equal(t, []string{"system.current_date"}, keys)
}