package template

import (
	"encoding/json"
	"os"
	"testing"
)

func BenchmarkHydrate_RealProductionState(b *testing.B) {
	// Load the largest JSON object from production
	data, err := os.ReadFile("/tmp/largest_state.json")
	if err != nil {
		b.Fatalf("Failed to load state: %v", err)
	}

	var params map[string]any
	if err := json.Unmarshal(data, &params); err != nil {
		b.Fatalf("Failed to parse state: %v", err)
	}

	// Simple template hydration (like utils.echo)
	template := map[string]any{
		"value": "{{loops}}",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hydrate(template, &params, nil)
	}
}

func BenchmarkHydrate_RealProductionState_DictOutput(b *testing.B) {
	data, err := os.ReadFile("/tmp/largest_state.json")
	if err != nil {
		b.Fatalf("Failed to load state: %v", err)
	}

	var params map[string]any
	if err := json.Unmarshal(data, &params); err != nil {
		b.Fatalf("Failed to parse state: %v", err)
	}

	// Tool params scenario
	template := map[string]any{
		"query":       "{{query}}",
		"num_results": 3,
		"country":     "{{country?}}",
		"language":    "{{language?}}",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hydrate(template, &params, nil)
	}
}
