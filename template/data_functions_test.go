package template

import (
	"testing"

	. "github.com/erdoai/erdo-common/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestMissingKeysDeduplication(t *testing.T) {
	tests := []struct {
		name            string
		data            map[string]any
		testFunc        func(map[string]any, *[]string)
		expectedMissing []string // Should be deduped
	}{
		{
			name: "Multiple functions adding same missing key should be deduped",
			data: map[string]any{
				"existing": "value",
			},
			testFunc: func(data map[string]any, missingKeys *[]string) {
				// Try to get the same missing key multiple times
				get("missing_key", data, missingKeys)
				get("missing_key", data, missingKeys)
				get("missing_key", data, missingKeys)
			},
			expectedMissing: []string{"missing_key"}, // Should only appear once
		},
		{
			name: "merge function adding duplicate missing keys should be deduped",
			data: map[string]any{
				"arr1": []any{"a"},
			},
			testFunc: func(data map[string]any, missingKeys *[]string) {
				// First merge with missing second array
				merge("arr1", "missing_arr", data, missingKeys)
				// Second merge with same missing array
				merge("arr1", "missing_arr", data, missingKeys)
			},
			expectedMissing: []string{"missing_arr"}, // Should only appear once
		},
		{
			name: "slice function with nil value adding duplicate keys should be deduped",
			data: map[string]any{
				"nil_arr": nil,
			},
			testFunc: func(data map[string]any, missingKeys *[]string) {
				// Multiple slice calls on nil array
				slice("nil_arr", 0, 1, data, missingKeys)
				slice("nil_arr", 1, 2, data, missingKeys)
			},
			expectedMissing: []string{"nil_arr"}, // Should only appear once
		},
		{
			name: "Different missing keys should all be preserved",
			data: map[string]any{},
			testFunc: func(data map[string]any, missingKeys *[]string) {
				get("missing1", data, missingKeys)
				get("missing2", data, missingKeys)
				get("missing3", data, missingKeys)
				get("missing1", data, missingKeys) // Duplicate
			},
			expectedMissing: []string{"missing1", "missing2", "missing3"}, // Three unique keys
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missingKeys := []string{}
			tt.testFunc(tt.data, &missingKeys)

			// Check for duplicates
			seen := make(map[string]bool)
			var dedupedKeys []string
			for _, key := range missingKeys {
				if !seen[key] {
					seen[key] = true
					dedupedKeys = append(dedupedKeys, key)
				}
			}

			// Verify no duplicates in result
			if len(dedupedKeys) != len(missingKeys) {
				t.Errorf("Missing keys contain duplicates. Original: %v, Deduped: %v", missingKeys, dedupedKeys)
			}

			// Verify we have the expected keys
			if len(missingKeys) != len(tt.expectedMissing) {
				t.Errorf("Expected %d missing keys, got %d. Expected: %v, Got: %v", 
					len(tt.expectedMissing), len(missingKeys), tt.expectedMissing, missingKeys)
			}
		})
	}
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

// TestStructTypeHandling tests that template functions work with struct types (not just maps)
func TestStructTypeHandling(t *testing.T) {
	// Define a Message struct similar to types.Message
	type Message struct {
		Role    string
		Content string
		ID      string
	}

	t.Run("sliceEndKeepFirstUserMessage with []Message structs", func(t *testing.T) {
		data := map[string]any{
			"messages": []Message{
				{Role: "user", Content: "first user message", ID: "1"},
				{Role: "assistant", Content: "first assistant response", ID: "2"},
				{Role: "assistant", Content: "second assistant response", ID: "3"},
				{Role: "user", Content: "second user message", ID: "4"},
			},
		}

		missingKeys := []string{}
		result := sliceEndKeepFirstUserMessage("messages", 2, data, &missingKeys)

		require.Empty(t, missingKeys, "Should not have missing keys")
		require.Len(t, result, 3, "Should return last 2 messages plus first user message")

		// Verify first element is the first user message
		firstMsg := result[0].(Message)
		assert.Equal(t, "user", firstMsg.Role)
		assert.Equal(t, "first user message", firstMsg.Content)
	})

	t.Run("filter with struct array", func(t *testing.T) {
		type Item struct {
			Name   string
			Status string
			Count  int
		}

		data := map[string]any{
			"items": []Item{
				{Name: "item1", Status: "active", Count: 5},
				{Name: "item2", Status: "inactive", Count: 3},
				{Name: "item3", Status: "active", Count: 7},
			},
		}

		missingKeys := []string{}
		result := filter("items", "Status", "eq", "active", data, &missingKeys)

		require.Empty(t, missingKeys)
		require.Len(t, result, 2, "Should filter to 2 active items")

		// Verify filtered results
		item1 := result[0].(Item)
		assert.Equal(t, "active", item1.Status)
	})

	t.Run("concat with struct array", func(t *testing.T) {
		type User struct {
			Name  string
			Email string
		}

		data := map[string]any{
			"users": []User{
				{Name: "Alice", Email: "alice@example.com"},
				{Name: "Bob", Email: "bob@example.com"},
				{Name: "Charlie", Email: "charlie@example.com"},
			},
		}

		missingKeys := []string{}
		result := concat(", ", "users", "Name", data, &missingKeys)

		assert.Equal(t, "Alice, Bob, Charlie", result)
		require.Empty(t, missingKeys)
	})

	t.Run("ToAnySlice with different slice types", func(t *testing.T) {
		type Custom struct {
			Value string
		}

		tests := []struct {
			name     string
			input    any
			expected int
			isNil    bool
		}{
			{
				name:     "[]any",
				input:    []any{1, 2, 3},
				expected: 3,
			},
			{
				name:     "[]string",
				input:    []string{"a", "b", "c"},
				expected: 3,
			},
			{
				name:     "[]int",
				input:    []int{1, 2, 3},
				expected: 3,
			},
			{
				name:     "[]Custom",
				input:    []Custom{{Value: "a"}, {Value: "b"}},
				expected: 2,
			},
			{
				name:     "[]Message",
				input:    []Message{{Role: "user"}},
				expected: 1,
			},
			{
				name:  "not a slice",
				input: "string",
				isNil: true,
			},
			{
				name:  "nil",
				input: nil,
				isNil: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := ToAnySlice(tt.input)
				if tt.isNil {
					assert.Nil(t, result)
				} else {
					require.NotNil(t, result)
					assert.Len(t, result, tt.expected)
				}
			})
		}
	})

	t.Run("GetFieldValue with different types", func(t *testing.T) {
		type Nested struct {
			Value string
		}

		type TestStruct struct {
			Name   string
			Count  int
			Active *bool
			Nested Nested
		}

		active := true
		testStruct := TestStruct{
			Name:   "test",
			Count:  42,
			Active: &active,
			Nested: Nested{Value: "nested value"},
		}

		// Test struct field access
		assert.Equal(t, "test", GetFieldValue(testStruct, "name"))
		assert.Equal(t, "test", GetFieldValue(testStruct, "Name"))
		assert.Equal(t, 42, GetFieldValue(testStruct, "count"))
		assert.Equal(t, 42, GetFieldValue(testStruct, "Count"))
		assert.Equal(t, true, GetFieldValue(testStruct, "active"))

		// Test map field access
		testMap := map[string]any{
			"name":   "test",
			"count":  42,
			"active": true,
		}

		assert.Equal(t, "test", GetFieldValue(testMap, "name"))
		assert.Equal(t, "test", GetFieldValue(testMap, "Name")) // Should try PascalCase
		assert.Equal(t, 42, GetFieldValue(testMap, "count"))

		// Test with PascalCase map keys
		pascalMap := map[string]any{
			"Name":   "test",
			"Count":  42,
			"Active": true,
		}

		assert.Equal(t, "test", GetFieldValue(pascalMap, "name")) // Should find Name
		assert.Equal(t, 42, GetFieldValue(pascalMap, "count"))    // Should find Count

		// Test nil/missing
		assert.Nil(t, GetFieldValue(testStruct, "nonexistent"))
		assert.Nil(t, GetFieldValue(testMap, "nonexistent"))
		assert.Nil(t, GetFieldValue(nil, "name"))
	})

	t.Run("isUserMessage with struct vs map", func(t *testing.T) {
		// Test with struct
		structMsg := Message{Role: "user", Content: "test"}
		assert.True(t, isUserMessage(structMsg))

		structAssistant := Message{Role: "assistant", Content: "test"}
		assert.False(t, isUserMessage(structAssistant))

		// Test with map
		mapMsg := map[string]any{"role": "user", "content": "test"}
		assert.True(t, isUserMessage(mapMsg))

		mapAssistant := map[string]any{"role": "assistant", "content": "test"}
		assert.False(t, isUserMessage(mapAssistant))

		// Test with PascalCase map
		pascalMsg := map[string]any{"Role": "user", "Content": "test"}
		assert.True(t, isUserMessage(pascalMsg))
	})
}
