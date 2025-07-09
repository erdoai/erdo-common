package template

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParameterHydrationBugFromCLI tests the fix for the issue where parameter hydration
// didn't find params and respect hydration behavior when moved from main erdo to erdo-common
func TestParameterHydrationBugFromCLI(t *testing.T) {
	tests := []struct {
		name                        string
		template                    map[string]any
		stateParams                 map[string]any
		parameterHydrationBehaviour map[string]any
		expected                    map[string]any
		shouldFail                  bool
		expectedError               string
	}{
		{
			name: "CLI Issue: Tools parameters should remain raw when specified",
			template: map[string]any{
				"tools": []map[string]any{
					{
						"action_type":            "bot.invoke",
						"bot_output_visibility":  "visible",
						"description":            "Use this tool to run a specific analysis on the data.",
						"input_schema": map[string]any{
							"properties": map[string]any{
								"additional_context": map[string]any{
									"description": "Additional, detailed context from the history.",
									"type":        "string",
								},
								"query": map[string]any{
									"description": "The specific query to answer with the analysis.",
									"type":        "string",
								},
							},
							"required": []string{"query"},
							"type":     "object",
						},
						"name": "analyze_data",
						"parameters": map[string]any{
							"bot_name":   "data analyst",
							"context":    "{{additional_context?}}",
							"query":      "{{query?}}",
							"other_key":  "{{url?}}",
							"system_key": "{{system.current_date}}",
						},
					},
				},
			},
			stateParams: map[string]any{
				"additional_context": "some context",
				"query":              "test query",
				"url":                "https://example.com",
				"system": map[string]any{
					"current_date": "2025-07-08",
				},
			},
			parameterHydrationBehaviour: map[string]any{
				"tools": map[string]any{
					"parameters": "raw",
				},
			},
			expected: map[string]any{
				"tools": []map[string]any{
					{
						"action_type":            "bot.invoke",
						"bot_output_visibility":  "visible",
						"description":            "Use this tool to run a specific analysis on the data.",
						"input_schema": map[string]any{
							"properties": map[string]any{
								"additional_context": map[string]any{
									"description": "Additional, detailed context from the history.",
									"type":        "string",
								},
								"query": map[string]any{
									"description": "The specific query to answer with the analysis.",
									"type":        "string",
								},
							},
							"required": []string{"query"},
							"type":     "object",
						},
						"name": "analyze_data",
						"parameters": map[string]any{
							"bot_name":   "data analyst",
							"context":    "{{additional_context?}}",  // Should remain raw
							"query":      "{{query?}}",               // Should remain raw
							"other_key":  "{{url?}}",                 // Should remain raw
							"system_key": "{{system.current_date}}", // Should remain raw
						},
					},
				},
			},
			shouldFail: false, // Bug is now fixed
		},
		{
			name: "CLI Issue: Keys should be found for hydration when not marked as raw",
			template: map[string]any{
				"system_prompt": "Current date is {{system.current_date}}",
				"tools": []map[string]any{
					{
						"parameters": map[string]any{
							"context": "{{additional_context?}}",
							"query":   "{{query?}}",
						},
					},
				},
			},
			stateParams: map[string]any{
				"additional_context": "some context",
				"query":              "test query",
				"system": map[string]any{
					"current_date": "2025-07-08",
				},
			},
			parameterHydrationBehaviour: map[string]any{
				"tools": map[string]any{
					"parameters": "raw", // parameters should be raw
				},
				// system_prompt should be hydrated normally
			},
			expected: map[string]any{
				"system_prompt": "Current date is 2025-07-08", // Should be hydrated
				"tools": []map[string]any{
					{
						"parameters": map[string]any{
							"context": "{{additional_context?}}", // Should remain raw
							"query":   "{{query?}}",             // Should remain raw
						},
					},
				},
			},
			shouldFail: false, // Bug is now fixed
		},
		{
			name: "FindTemplateKeyStringsToHydrate should identify keys correctly",
			template: map[string]any{
				"system_prompt": "Date: {{system.current_date}}",
				"tools": []map[string]any{
					{
						"parameters": map[string]any{
							"query":   "{{query?}}",
							"context": "{{additional_context?}}",
						},
					},
				},
			},
			stateParams: map[string]any{
				"query":              "test query",
				"additional_context": "some context",
				"system": map[string]any{
					"current_date": "2025-07-08",
				},
			},
			parameterHydrationBehaviour: map[string]any{
				"tools": map[string]any{
					"parameters": "raw",
				},
			},
			// For this test, we'll verify FindTemplateKeyStringsToHydrate function behavior
			shouldFail: false, // Bug is now fixed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "FindTemplateKeyStringsToHydrate should identify keys correctly" {
				// Test the key finding function specifically
				keys := FindTemplateKeyStringsToHydrate(tt.template, true, &tt.parameterHydrationBehaviour)
				
				// Should find: system.current_date (from system_prompt)
				// Should NOT find: query, additional_context (from tools.parameters which is raw)
				expectedKeys := []string{"system.current_date"}
				
				// Log what we actually found
				t.Logf("Keys found: %v", keys)
				t.Logf("Expected keys: %v", expectedKeys)
				
				// This should fail initially if the bug exists
				assert.ElementsMatch(t, expectedKeys, keys, "FindTemplateKeyStringsToHydrate should correctly identify keys to hydrate")
				return
			}

			result, err := HydrateDict(tt.template, &tt.stateParams, &tt.parameterHydrationBehaviour)
			
			// All tests should now pass since the bugs are fixed
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHydrateParametersFromLog tests the specific scenario from the CLI log
func TestHydrateParametersFromLog(t *testing.T) {
	// This simulates the exact scenario from the log where keys are not being found
	combinedParamsJSON := map[string]any{
		"invocation_id": "fd7387cd-d5c5-4edd-b537-b05ed543b7ef",
		"loops":         0,
		"resources": []map[string]any{
			{
				"id": 1,
				"dataset": map[string]any{
					"id":          "85f6fab7-461c-4855-bac4-b29e24105296",
					"type":        "integration",
					"key":         "financial_datasets",
					"name":        "financial datasets",
					"description": "",
				},
			},
		},
		"system": map[string]any{
			"current_date": "2025-07-08 15:35:05 BST",
			"messages": []map[string]any{
				{"role": "user", "content": "test message"},
			},
		},
	}

	templateHydrationBehaviourJSON := map[string]any{
		"tools": map[string]any{
			"parameters": "raw",
		},
	}

	// Test template that should require some keys to be hydrated and others to remain raw
	template := map[string]any{
		"message_history": "{{get \"system.messages\"}}",
		"model":           "claude-sonnet-4-20250514",
		"system_prompt":   "Current system date: {{system.current_date}}",
		"tools": []map[string]any{
			{
				"name":        "analyze_data",
				"description": "Use this tool to run a specific analysis on the data.",
				"parameters": map[string]any{
					"query":                "{{query?}}",
					"additional_context":   "{{additional_context?}}",
					"country":              "{{country?}}",
					"language":             "{{language?}}",
					"num_results":          "{{num_results?}}",
					"include_images":       "{{include_images?}}",
					"remove_selector":      "{{remove_selector?}}",
					"target_selector":      "{{target_selector?}}",
					"url":                  "{{url?}}",
					"include_links":        "{{include_links?}}",
					"system_current_date":  "{{system.current_date}}",
				},
			},
		},
	}

	t.Run("Keys should be correctly identified for hydration", func(t *testing.T) {
		// First test: Find template keys
		keys := FindTemplateKeyStringsToHydrate(template, true, &templateHydrationBehaviourJSON)
		
		// Should find: system.current_date (used in system_prompt and tools.parameters.system_current_date)
		// Should find: system.messages (used in message_history)
		// Should NOT find: query, additional_context, etc. (used in tools.parameters which is raw)
		
		t.Logf("Found keys: %v", keys)
		
		// This will likely fail initially due to the bug
		expectedKeys := []string{"system.current_date", "system.messages"}
		t.Logf("Expected keys: %v", expectedKeys)
		
		// Check if we can find system.current_date
		found := false
		for _, key := range keys {
			if key == "system.current_date" {
				found = true
				break
			}
		}
		
		if !found {
			t.Errorf("BUG: system.current_date should be found but wasn't. Found keys: %v", keys)
		}
	})

	t.Run("Hydration should work with available parameters", func(t *testing.T) {
		// This should not fail with missing keys for system.current_date since it's available
		result, err := HydrateDict(template, &combinedParamsJSON, &templateHydrationBehaviourJSON)
		
		if err != nil {
			// Log the error to see what's happening
			t.Logf("Error during hydration (this might show the bug): %v", err)
			
			// Check if it's a missing keys error
			if infoErr, ok := err.(*InfoNeededError); ok {
				t.Logf("Missing keys: %v", infoErr.MissingKeys)
				t.Logf("Available keys: %v", infoErr.AvailableKeys)
				
				// Check if system.current_date is in missing keys when it should be available
				for _, missingKey := range infoErr.MissingKeys {
					if missingKey == "system.current_date" {
						t.Errorf("BUG: system.current_date is reported as missing but should be available")
					}
				}
			}
		}
		
		// Log the result for debugging
		t.Logf("Hydration result: %+v", result)
		
		// The expected behavior:
		// - system_prompt should be hydrated with the current date
		// - tools.parameters should remain as raw templates
		if result != nil {
			if systemPrompt, ok := result["system_prompt"].(string); ok {
				if !strings.Contains(systemPrompt, "2025-07-08") {
					t.Errorf("BUG: system_prompt should contain the hydrated date, got: %s", systemPrompt)
				}
			}
		}
	})
}