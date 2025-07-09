package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestComplexTemplateKeyExtraction tests that the key extraction correctly handles
// complex templates with range variables, tool parameters, and nested keys
func TestComplexTemplateKeyExtraction(t *testing.T) {
	tests := []struct {
		name                            string
		template                        string
		parameterHydrationBehaviour     *map[string]any
		expectedKeys                    []string
		shouldIncludeOptional          bool
	}{
		{
			name: "Range variables should not be extracted",
			template: `{{range $r := .Data.resources}}{{$r.id}} {{$r.dataset.filename}}{{end}}`,
			expectedKeys: []string{"resources"},
		},
		{
			name: "Basic template key extraction",
			template: `{"query": "{{query?}}", "url": "{{url}}"}`,
			expectedKeys: []string{"url"}, // query is optional and excluded by default
			shouldIncludeOptional: false,
		},
		{
			name: "System nested keys should be extracted correctly",
			template: `Current date: {{system.current_date}}`,
			expectedKeys: []string{"system.current_date"},
		},
		{
			name: "Mixed template with valid and invalid keys",
			template: `{{range $r := .Data.resources}}Resource: {{$r.id}} - Name: {{resource_name}}{{end}} Date: {{system.current_date}}`,
			expectedKeys: []string{"resources", "resource_name", "system.current_date"},
		},
		{
			name: "Reserved words should not be extracted",
			template: `{{if something}}{{something}}{{end}}`,
			expectedKeys: []string{"something"}, // if/end are reserved, but something is not
		},
		{
			name: "Optional parameters",
			template: `{{optional_param?}} {{required_param}}`,
			expectedKeys: []string{"required_param"}, // optional excluded by default
			shouldIncludeOptional: false,
		},
		{
			name: "Optional parameters when included",
			template: `{{optional_param?}} {{required_param}}`,
			expectedKeys: []string{"optional_param", "required_param"},
			shouldIncludeOptional: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := FindTemplateKeysToHydrate(tt.template, tt.shouldIncludeOptional, tt.parameterHydrationBehaviour)
			
			// Extract just the key names for comparison
			var keyNames []string
			for _, key := range keys {
				keyNames = append(keyNames, key.Key)
			}
			
			assert.ElementsMatch(t, tt.expectedKeys, keyNames, 
				"Expected keys %v but got %v", tt.expectedKeys, keyNames)
		})
	}
}

// TestParameterHydrationBehavior tests that raw parameters are not processed for key extraction
func TestParameterHydrationBehavior(t *testing.T) {
	// Using a map structure to properly test hydration behavior
	templateData := map[string]any{
		"tools": []any{
			map[string]any{
				"parameters": map[string]any{
					"query": "{{query?}}",
					"url": "{{url}}",
					"resource_keys": "{{resource_keys?}}",
				},
			},
		},
		"system_prompt": "Current date: {{system.current_date}}",
		"resources": "{{resources}}",
	}
	
	// This mimics the behavior configuration from the error logs
	behavior := &map[string]any{
		"tools": map[string]any{
			"parameters": "raw",
		},
	}
	
	keys := FindTemplateKeysToHydrate(templateData, false, behavior)
	
	// Extract just the key names for comparison
	var keyNames []string
	for _, key := range keys {
		keyNames = append(keyNames, key.Key)
	}
	
	// Should only include system.current_date and resources, not the tool parameters
	expectedKeys := []string{"system.current_date", "resources"}
	assert.ElementsMatch(t, expectedKeys, keyNames, 
		"Expected only non-raw keys to be extracted, but got %v", keyNames)
}