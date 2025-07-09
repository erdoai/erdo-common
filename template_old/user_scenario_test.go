package template

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
)

// TestUserScenarioRawBehavior tests the exact scenario from the user's error logs
func TestUserScenarioRawBehavior(t *testing.T) {
	// This is the exact structure from the user's logs
	invocationStateParameters := map[string]any{
		"model": "claude-sonnet-4-20250514",
		"tools": []any{
			map[string]any{
				"name":        "analyze_data",
				"description": "Use this tool to run a specific analysis on the data.",
				"parameters": map[string]any{
					"query":               "{{query?}}",
					"url":                 "{{url?}}",
					"num_results":         "{{num_results?}}",
					"target_selector":     "{{target_selector?}}",
					"remove_selector":     "{{remove_selector?}}",
					"include_links":       "{{include_links?}}",
					"include_images":      "{{include_images?}}",
					"language":            "{{language?}}",
					"country":             "{{country?}}",
					"additional_context":  "{{additional_context?}}",
					"system_current_date": "{{system.current_date}}",
				},
			},
		},
		"system_prompt": `Current system date: {{system.current_date}}`,
		"message_history": []any{
			map[string]any{
				"role":    "user",
				"content": "test message",
			},
		},
	}
	
	// The behavior configuration
	behavior := &map[string]any{
		"tools": map[string]any{
			"parameters": "raw",
		},
	}
	
	// Available data including system.current_date
	availableData := map[string]any{
		"system": map[string]any{
			"current_date": "2025-07-08 15:35:05 BST",
		},
	}
	
	// Extract keys that need hydration
	keys := FindTemplateKeyStringsToHydrate(invocationStateParameters, false, behavior)
	
	// Should only find system.current_date from system_prompt
	// Should NOT find any keys from tools[].parameters since it's marked as raw
	assert.Equal(t, []string{"system.current_date"}, keys)
	
	// Hydrate the parameters
	result, err := Hydrate(invocationStateParameters, &availableData, behavior)
	assert.NoError(t, err)
	
	resultMap := result.(map[string]any)
	
	// Check system_prompt is hydrated
	assert.Equal(t, "Current system date: 2025-07-08 15:35:05 BST", resultMap["system_prompt"])
	
	// Check tools[].parameters is NOT hydrated
	tools := resultMap["tools"].([]any)
	tool := tools[0].(map[string]any)
	params := tool["parameters"].(map[string]any)
	
	// All template strings should remain unchanged
	assert.Equal(t, "{{query?}}", params["query"])
	assert.Equal(t, "{{url?}}", params["url"])
	assert.Equal(t, "{{num_results?}}", params["num_results"])
	assert.Equal(t, "{{target_selector?}}", params["target_selector"])
	assert.Equal(t, "{{remove_selector?}}", params["remove_selector"])
	assert.Equal(t, "{{include_links?}}", params["include_links"])
	assert.Equal(t, "{{include_images?}}", params["include_images"])
	assert.Equal(t, "{{language?}}", params["language"])
	assert.Equal(t, "{{country?}}", params["country"])
	assert.Equal(t, "{{additional_context?}}", params["additional_context"])
	assert.Equal(t, "{{system.current_date}}", params["system_current_date"])
	
	// Tool name and description should still be the same (not templates)
	assert.Equal(t, "analyze_data", tool["name"])
	assert.Equal(t, "Use this tool to run a specific analysis on the data.", tool["description"])
}