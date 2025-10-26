package template

import (
	"testing"
)

// Benchmark common hydration scenarios to identify optimization opportunities

func BenchmarkHydrate_SimpleString(b *testing.B) {
	params := map[string]any{
		"name": "John",
	}
	template := "Hello {{name}}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hydrate(template, &params, nil)
	}
}

func BenchmarkHydrate_MultipleSimpleReplacements(b *testing.B) {
	params := map[string]any{
		"name":    "John",
		"city":    "London",
		"country": "UK",
		"age":     "30",
	}
	template := "{{name}} from {{city}}, {{country}} is {{age}} years old"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hydrate(template, &params, nil)
	}
}

func BenchmarkHydrate_NestedDict(b *testing.B) {
	params := map[string]any{
		"user": map[string]any{
			"name": "John",
			"address": map[string]any{
				"city":    "London",
				"country": "UK",
			},
		},
	}
	template := "{{user.name}} lives in {{user.address.city}}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hydrate(template, &params, nil)
	}
}

func BenchmarkHydrate_ArrayAccess(b *testing.B) {
	params := map[string]any{
		"items": []any{"apple", "banana", "cherry"},
	}
	template := "First: {{items[0]}}, Last: {{items[2]}}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hydrate(template, &params, nil)
	}
}

func BenchmarkHydrate_OptionalParams(b *testing.B) {
	params := map[string]any{
		"name": "John",
		// country is missing
	}
	template := "{{name}} from {{country?}}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hydrate(template, &params, nil)
	}
}

func BenchmarkHydrate_FunctionCall(b *testing.B) {
	params := map[string]any{
		"data": map[string]any{"key": "value"},
	}
	template := "JSON: {{toJson(data)}}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hydrate(template, &params, nil)
	}
}

func BenchmarkHydrate_LargeDict(b *testing.B) {
	// Simulate large parameter dict (100 keys)
	params := make(map[string]any)
	for i := 0; i < 100; i++ {
		params["key"+string(rune('a'+i%26))+string(rune('0'+i%10))] = "value"
	}
	params["target"] = "found"

	template := "Looking for: {{target}}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hydrate(template, &params, nil)
	}
}

func BenchmarkHydrate_DictOutput(b *testing.B) {
	params := map[string]any{
		"query":       "search term",
		"num_results": 10,
	}
	template := map[string]any{
		"query":       "{{query}}",
		"num_results": "{{num_results}}",
		"country":     "{{country?}}",
		"language":    "{{language?}}",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hydrate(template, &params, nil)
	}
}

func BenchmarkHydrate_EchoScenario(b *testing.B) {
	// Simulate the utils.echo scenario from logs
	params := map[string]any{
		"tool_usage_loops": 0,
	}
	template := map[string]any{
		"value": "{{tool_usage_loops}}",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hydrate(template, &params, nil)
	}
}

func BenchmarkHydrate_ToolParamsScenario(b *testing.B) {
	// Simulate the search_web tool params scenario from logs
	params := map[string]any{
		"query":       "Google Analytics GA4 API",
		"num_results": 3,
		// country and language missing (optional)
	}
	template := map[string]any{
		"query":       "{{query}}",
		"num_results": "{{num_results}}",
		"country":     "{{country?}}",
		"language":    "{{language?}}",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hydrate(template, &params, nil)
	}
}

func BenchmarkHydrate_NoTemplates(b *testing.B) {
	// Best case: no templates to hydrate
	params := map[string]any{
		"key": "value",
	}
	template := "Plain string with no templates"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hydrate(template, &params, nil)
	}
}

func BenchmarkHydrate_ComplexNested(b *testing.B) {
	params := map[string]any{
		"bot": map[string]any{
			"name": "Assistant",
			"config": map[string]any{
				"model":       "claude-3",
				"temperature": 0.7,
				"tools": []any{
					map[string]any{"name": "search", "enabled": true},
					map[string]any{"name": "code", "enabled": false},
				},
			},
		},
	}
	template := map[string]any{
		"bot_name": "{{bot.name}}",
		"model":    "{{bot.config.model}}",
		"temp":     "{{bot.config.temperature}}",
		"tools": []any{
			"{{bot.config.tools[0].name}}",
			"{{bot.config.tools[1].name}}",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hydrate(template, &params, nil)
	}
}

func BenchmarkFindTemplateKeyStrings_SimpleString(b *testing.B) {
	template := "Hello {{name}} from {{city}}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FindTemplateKeyStringsToHydrate(template, true, nil)
	}
}

func BenchmarkFindTemplateKeyStrings_LargeDict(b *testing.B) {
	template := make(map[string]any)
	for i := 0; i < 100; i++ {
		template["key"+string(rune('a'+i%26))] = "value {{param}}"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FindTemplateKeyStringsToHydrate(template, true, nil)
	}
}

func BenchmarkFindTemplateKeyStrings_NoTemplates(b *testing.B) {
	template := map[string]any{
		"key1": "plain value 1",
		"key2": "plain value 2",
		"key3": "plain value 3",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FindTemplateKeyStringsToHydrate(template, true, nil)
	}
}

// Realistic production scenarios with large state dicts

func BenchmarkHydrate_LargeState_SimpleTemplate(b *testing.B) {
	// Simulate large invocation state (1000 keys like real production)
	params := make(map[string]any)

	// Add parameters (50 keys)
	for i := 0; i < 50; i++ {
		params["param_"+string(rune('a'+i%26))+string(rune('0'+i%10))] = "value"
	}

	// Add current state (200 keys with nested objects)
	for i := 0; i < 200; i++ {
		params["state_"+string(rune('a'+i%26))+string(rune('0'+i%10))] = map[string]any{
			"nested_field_1": "value1",
			"nested_field_2": 12345,
			"nested_field_3": true,
		}
	}

	// Add resources (100 resources with metadata)
	resources := make([]any, 100)
	for i := 0; i < 100; i++ {
		resources[i] = map[string]any{
			"id":   "resource_" + string(rune('a'+i%26)),
			"name": "Resource " + string(rune('0'+i%10)),
			"type": "table",
			"metadata": map[string]any{
				"columns": []any{"col1", "col2", "col3"},
				"rows":    1000,
			},
		}
	}
	params["resources"] = resources

	// Add system params (50 keys)
	system := make(map[string]any)
	for i := 0; i < 50; i++ {
		system["sys_"+string(rune('a'+i%26))] = "system_value"
	}
	params["system"] = system

	// Simple template that only uses one key
	template := "Value: {{param_a0}}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hydrate(template, &params, nil)
	}
}

func BenchmarkHydrate_LargeState_DictOutput(b *testing.B) {
	// Same large state as above
	params := make(map[string]any)
	for i := 0; i < 50; i++ {
		params["param_"+string(rune('a'+i%26))+string(rune('0'+i%10))] = "value"
	}
	for i := 0; i < 200; i++ {
		params["state_"+string(rune('a'+i%26))+string(rune('0'+i%10))] = map[string]any{
			"nested": "data",
		}
	}
	resources := make([]any, 100)
	for i := 0; i < 100; i++ {
		resources[i] = map[string]any{"id": "res_" + string(rune('a'+i%26))}
	}
	params["resources"] = resources

	// Template is a dict with 4 fields (2 present, 2 optional missing)
	template := map[string]any{
		"query":       "{{param_a0}}",
		"num_results": "{{param_b1}}",
		"country":     "{{country?}}",
		"language":    "{{language?}}",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hydrate(template, &params, nil)
	}
}

func BenchmarkHydrate_MassiveState_NestedAccess(b *testing.B) {
	// Extremely large state (5000 keys total)
	params := make(map[string]any)

	// 500 top-level params
	for i := 0; i < 500; i++ {
		key := "p" + string(rune('a'+i%26)) + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10))
		params[key] = "value" + string(rune('0'+i%10))
	}

	// Large nested structures
	params["bot_state"] = map[string]any{
		"invocation_id": "inv_123",
		"step_count":    50,
		"history": make([]any, 100), // 100 items in history
	}

	// Deep nesting
	params["data"] = map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"target": "found",
				},
			},
		},
	}

	template := "Result: {{data.level1.level2.level3.target}}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hydrate(template, &params, nil)
	}
}
