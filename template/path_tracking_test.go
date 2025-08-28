package template

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorPathTracking(t *testing.T) {
	tests := []struct {
		name           string
		template       any
		data           map[string]any
		expectedError  bool
		expectedPaths  []string
	}{
		{
			name: "nested dict missing key",
			template: map[string]any{
				"outer": map[string]any{
					"inner": map[string]any{
						"value": "{{missing_key}}",
					},
				},
			},
			data:          map[string]any{},
			expectedError: true,
			expectedPaths: []string{"outer.inner.value.missing_key"},
		},
		{
			name: "multiple nested missing keys",
			template: map[string]any{
				"config": map[string]any{
					"database": map[string]any{
						"host": "{{db_host}}",
						"port": "{{db_port}}",
					},
					"cache": map[string]any{
						"url": "{{cache_url}}",
					},
				},
			},
			data:          map[string]any{},
			expectedError: true,
			expectedPaths: []string{"config.database.host.db_host", "config.database.port.db_port", "config.cache.url.cache_url"},
		},
		{
			name: "array with missing keys",
			template: map[string]any{
				"servers": []any{
					map[string]any{
						"name": "{{server1_name}}",
						"ip":   "{{server1_ip}}",
					},
					map[string]any{
						"name": "{{server2_name}}",
						"ip":   "{{server2_ip}}",
					},
				},
			},
			data:          map[string]any{},
			expectedError: true,
			expectedPaths: []string{"servers[0].name.server1_name", "servers[0].ip.server1_ip", "servers[1].name.server2_name", "servers[1].ip.server2_ip"},
		},
		{
			name: "deeply nested structure",
			template: map[string]any{
				"level1": map[string]any{
					"level2": map[string]any{
						"level3": map[string]any{
							"level4": map[string]any{
								"value": "{{deep_value}}",
							},
						},
					},
				},
			},
			data:          map[string]any{},
			expectedError: true,
			expectedPaths: []string{"level1.level2.level3.level4.value.deep_value"},
		},
		{
			name: "complex structure with many missing keys",
			template: map[string]any{
				"app": map[string]any{
					"database": map[string]any{
						"primary": map[string]any{
							"host": "{{primary_db_host}}",
							"port": "{{primary_db_port}}",
							"user": "{{primary_db_user}}",
						},
						"replica": map[string]any{
							"host": "{{replica_db_host}}",
							"port": "{{replica_db_port}}",
						},
					},
					"cache": map[string]any{
						"servers": []any{
							map[string]any{
								"host": "{{cache1_host}}",
								"port": "{{cache1_port}}",
							},
							map[string]any{
								"host": "{{cache2_host}}",
								"port": "{{cache2_port}}",
							},
						},
					},
					"api": map[string]any{
						"key":     "{{api_key}}",
						"secret":  "{{api_secret}}",
						"timeout": "{{api_timeout}}",
					},
				},
			},
			data:          map[string]any{},
			expectedError: true,
			expectedPaths: []string{
				"app.database.primary.host.primary_db_host",
				"app.database.primary.port.primary_db_port",
				"app.database.primary.user.primary_db_user",
				"app.database.replica.host.replica_db_host",
				"app.database.replica.port.replica_db_port",
				"app.cache.servers[0].host.cache1_host",
				"app.cache.servers[0].port.cache1_port",
				"app.cache.servers[1].host.cache2_host",
				"app.cache.servers[1].port.cache2_port",
				"app.api.key.api_key",
				"app.api.secret.api_secret",
				"app.api.timeout.api_timeout",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Hydrate(tt.template, &tt.data, nil)
			
			if tt.expectedError {
				assert.Error(t, err)
				
				// Check if it's an InfoNeededError
				var infoErr *InfoNeededError
				if assert.ErrorAs(t, err, &infoErr) {
					// Check that we get missing keys
					assert.NotEmpty(t, infoErr.MissingKeys)
					
					// Check that we have path information
					assert.NotEmpty(t, infoErr.MissingKeyPaths)
					
					// Extract actual paths for comparison
					actualPaths := make([]string, len(infoErr.MissingKeyPaths))
					for i, info := range infoErr.MissingKeyPaths {
						// Build the full path as it appears in the error message
						if info.Path != "" {
							if strings.HasPrefix(info.Path, "[") {
								actualPaths[i] = info.Path + "." + info.Key
							} else {
								actualPaths[i] = info.Path + "." + info.Key
							}
						} else {
							actualPaths[i] = info.Key
						}
						t.Logf("Missing key '%s' at path '%s' -> full path: '%s'", info.Key, info.Path, actualPaths[i])
					}
					
					// Verify expected paths are present
					for _, expectedPath := range tt.expectedPaths {
						found := false
						for _, actualPath := range actualPaths {
							if actualPath == expectedPath {
								found = true
								break
							}
						}
						assert.True(t, found, "Expected path %s not found in error. Actual paths: %v", expectedPath, actualPaths)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}