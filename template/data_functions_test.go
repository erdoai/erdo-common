package template

import (
	"testing"
)

func TestMergeVsMergeRawOriginalIssue(t *testing.T) {
	// This test reproduces the original issue where merge was used incorrectly
	// with function results instead of string keys

	data := map[string]any{
		"system": map[string]any{
			"messages": []any{
				map[string]any{"role": "user", "content": "show aapl ytd"},
			},
		},
		"steps": map[string]any{
			"prepare_dynamic_context": map[string]any{
				"dynamic_messages": []any{
					map[string]any{"role": "assistant", "content": "system info"},
				},
			},
		},
	}

	t.Run("merge with correct string keys - should work", func(t *testing.T) {
		missingKeys := []string{}
		result := merge("system.messages", "steps.prepare_dynamic_context.dynamic_messages", data, &missingKeys)

		if len(missingKeys) > 0 {
			t.Errorf("Expected merge with correct keys to work, but got missingKeys: %v", missingKeys)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 merged messages, got %d", len(result))
		}
	})

	t.Run("merge with function result (wrong usage) - should fail fast", func(t *testing.T) {
		missingKeys := []string{}

		// This simulates what would happen with the original broken template:
		// {{merge (sliceEndKeepFirstUserMessage system.messages 10) steps.prepare_dynamic_context.dynamic_messages}}
		//
		// The invalid key should be caught immediately
		result := merge("(sliceEndKeepFirstUserMessage system.messages 10)", "steps.prepare_dynamic_context.dynamic_messages", data, &missingKeys)

		// This should fail because "(sliceEndKeepFirstUserMessage system.messages 10)" is not a valid key
		if len(missingKeys) == 0 {
			t.Errorf("Expected merge with invalid key to fail, but it didn't. Result: %v", result)
		}
	})
}

func TestMergeFailFastBehavior(t *testing.T) {
	tests := []struct {
		name           string
		data           map[string]any
		array1Key      string
		array2Key      string
		shouldFail     bool
		expectedResult []any
	}{
		{
			name: "both arrays exist and non-empty",
			data: map[string]any{
				"arr1": []any{"a", "b"},
				"arr2": []any{"c", "d"},
			},
			array1Key:      "arr1",
			array2Key:      "arr2",
			shouldFail:     false,
			expectedResult: []any{"a", "b", "c", "d"},
		},
		{
			name: "both arrays exist but empty - should NOT fail",
			data: map[string]any{
				"arr1": []any{},
				"arr2": []any{},
			},
			array1Key:      "arr1",
			array2Key:      "arr2",
			shouldFail:     false,
			expectedResult: []any{},
		},
		{
			name: "one array empty, one non-empty - should NOT fail",
			data: map[string]any{
				"arr1": []any{},
				"arr2": []any{"c", "d"},
			},
			array1Key:      "arr1",
			array2Key:      "arr2",
			shouldFail:     false,
			expectedResult: []any{"c", "d"},
		},
		{
			name: "first array missing - should fail",
			data: map[string]any{
				"arr2": []any{"c", "d"},
			},
			array1Key:  "missing_arr1",
			array2Key:  "arr2",
			shouldFail: true,
		},
		{
			name: "second array missing - should fail",
			data: map[string]any{
				"arr1": []any{"a", "b"},
			},
			array1Key:  "arr1",
			array2Key:  "missing_arr2",
			shouldFail: true,
		},
		{
			name: "first array is nil - should fail",
			data: map[string]any{
				"arr1": nil,
				"arr2": []any{"c", "d"},
			},
			array1Key:  "arr1",
			array2Key:  "arr2",
			shouldFail: true,
		},
		{
			name: "first array is wrong type - should fail",
			data: map[string]any{
				"arr1": "not an array",
				"arr2": []any{"c", "d"},
			},
			array1Key:  "arr1",
			array2Key:  "arr2",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missingKeys := []string{}
			result := merge(tt.array1Key, tt.array2Key, tt.data, &missingKeys)

			if tt.shouldFail {
				if len(missingKeys) == 0 {
					t.Errorf("Expected merge to fail (add to missingKeys), but it didn't. Result: %v", result)
				}
			} else {
				if len(missingKeys) > 0 {
					t.Errorf("Expected merge to succeed, but it failed with missingKeys: %v", missingKeys)
				}
				if len(result) != len(tt.expectedResult) {
					t.Errorf("Expected result length %d, got %d. Expected: %v, Got: %v", len(tt.expectedResult), len(result), tt.expectedResult, result)
				}
				for i, expected := range tt.expectedResult {
					if i < len(result) && result[i] != expected {
						t.Errorf("Expected result[%d] = %v, got %v", i, expected, result[i])
					}
				}
			}
		})
	}
}

func TestSliceFailFastBehavior(t *testing.T) {
	tests := []struct {
		name           string
		data           map[string]any
		arrayKey       string
		start          int
		end            int
		shouldFail     bool
		expectedResult []any
	}{
		{
			name: "array exists and non-empty",
			data: map[string]any{
				"arr": []any{"a", "b", "c", "d"},
			},
			arrayKey:       "arr",
			start:          1,
			end:            3,
			shouldFail:     false,
			expectedResult: []any{"b", "c"},
		},
		{
			name: "array exists but empty - should NOT fail",
			data: map[string]any{
				"arr": []any{},
			},
			arrayKey:       "arr",
			start:          0,
			end:            1,
			shouldFail:     false,
			expectedResult: []any{},
		},
		{
			name:       "array missing - should fail",
			data:       map[string]any{},
			arrayKey:   "missing_arr",
			start:      0,
			end:        1,
			shouldFail: true,
		},
		{
			name: "array is nil - should fail",
			data: map[string]any{
				"arr": nil,
			},
			arrayKey:   "arr",
			start:      0,
			end:        1,
			shouldFail: true,
		},
		{
			name: "array is wrong type - should fail",
			data: map[string]any{
				"arr": "not an array",
			},
			arrayKey:   "arr",
			start:      0,
			end:        1,
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missingKeys := []string{}
			result := slice(tt.arrayKey, tt.start, tt.end, tt.data, &missingKeys)

			if tt.shouldFail {
				if len(missingKeys) == 0 {
					t.Errorf("Expected slice to fail (add to missingKeys), but it didn't. Result: %v", result)
				}
			} else {
				if len(missingKeys) > 0 {
					t.Errorf("Expected slice to succeed, but it failed with missingKeys: %v", missingKeys)
				}
				if len(result) != len(tt.expectedResult) {
					t.Errorf("Expected result length %d, got %d. Expected: %v, Got: %v", len(tt.expectedResult), len(result), tt.expectedResult, result)
				}
				for i, expected := range tt.expectedResult {
					if i < len(result) && result[i] != expected {
						t.Errorf("Expected result[%d] = %v, got %v", i, expected, result[i])
					}
				}
			}
		})
	}
}

func TestSliceEndKeepFirstUserMessageFailFastBehavior(t *testing.T) {
	tests := []struct {
		name           string
		data           map[string]any
		arrayKey       string
		n              int
		shouldFail     bool
		expectedLength int
	}{
		{
			name: "array exists with messages",
			data: map[string]any{
				"messages": []any{
					map[string]any{"role": "user", "content": "msg1"},
					map[string]any{"role": "assistant", "content": "msg2"},
					map[string]any{"role": "user", "content": "msg3"},
				},
			},
			arrayKey:       "messages",
			n:              2,
			shouldFail:     false,
			expectedLength: 3, // function keeps first user message + last 2, so all 3 in this case
		},
		{
			name: "array exists but empty - should NOT fail",
			data: map[string]any{
				"messages": []any{},
			},
			arrayKey:       "messages",
			n:              2,
			shouldFail:     false,
			expectedLength: 0,
		},
		{
			name:       "array missing - should fail",
			data:       map[string]any{},
			arrayKey:   "missing_messages",
			n:          2,
			shouldFail: true,
		},
		{
			name: "array is nil - should fail",
			data: map[string]any{
				"messages": nil,
			},
			arrayKey:   "messages",
			n:          2,
			shouldFail: true,
		},
		{
			name: "array is wrong type - should fail",
			data: map[string]any{
				"messages": "not an array",
			},
			arrayKey:   "messages",
			n:          2,
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missingKeys := []string{}
			result := sliceEndKeepFirstUserMessage(tt.arrayKey, tt.n, tt.data, &missingKeys)

			if tt.shouldFail {
				if len(missingKeys) == 0 {
					t.Errorf("Expected sliceEndKeepFirstUserMessage to fail (add to missingKeys), but it didn't. Result: %v", result)
				}
			} else {
				if len(missingKeys) > 0 {
					t.Errorf("Expected sliceEndKeepFirstUserMessage to succeed, but it failed with missingKeys: %v", missingKeys)
				}
				if len(result) != tt.expectedLength {
					t.Errorf("Expected result length %d, got %d. Result: %v", tt.expectedLength, len(result), result)
				}
			}
		})
	}
}
