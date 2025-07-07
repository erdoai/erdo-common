package template

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTemplateFunctions(t *testing.T) {
	t.Run("addkey function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"object": Dict{"existingKey": "existingValue"},
			"value":  "newValue",
		}

		result := addkey("object", "newKey", "value", data, &missingKeys)
		expected := Dict{"existingKey": "existingValue", "newKey": "newValue"}
		assert.Equal(t, expected, result)
	})

	t.Run("removekey function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"object": Dict{"key1": "value1", "key2": "value2"},
		}

		result := removekey("object", "key1", data, &missingKeys)
		expected := Dict{"key2": "value2"}
		assert.Equal(t, expected, result)
	})

	t.Run("mapToDict function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"stringList": []any{"value1", "value2", "value3"},
		}

		result := mapToDict("stringList", "key", data, &missingKeys)
		expected := []Dict{
			{"key": "value1"},
			{"key": "value2"},
			{"key": "value3"},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("coalesce function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"existingKey": "value",
		}

		// Test with existing key
		result := coalesce("existingKey", "default", data, &missingKeys)
		assert.Equal(t, "value", result)

		// Test with missing key
		result = coalesce("missingKey", "default", data, &missingKeys)
		assert.Equal(t, "default", result)

		// Test with optional key syntax
		result = coalesce("missingKey?", 0, data, &missingKeys)
		assert.Equal(t, 0, result)
	})

	t.Run("extractSlice function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"items": []any{
				Dict{"memory": Dict{"ID": "1", "Content": "Memory 1"}, "distance": 0.5},
				Dict{"memory": Dict{"ID": "2", "Content": "Memory 2"}, "distance": 0.3},
			},
		}

		result := extractSlice("items", "memory", data, &missingKeys)
		expected := []any{
			Dict{"ID": "1", "Content": "Memory 1"},
			Dict{"ID": "2", "Content": "Memory 2"},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("dedupeBy function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"simpleItems": []any{
				Dict{"id": "1", "name": "Item 1"},
				Dict{"id": "2", "name": "Item 2"},
				Dict{"id": "1", "name": "Item 1 Duplicate"},
				Dict{"id": "3", "name": "Item 3"},
			},
		}

		result := dedupeBy("simpleItems", "id", data, &missingKeys)
		assert.Equal(t, 3, len(result))
	})

	t.Run("find function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"items": []any{
				Dict{"id": "1", "name": "Item 1"},
				Dict{"id": "2", "name": "Item 2"},
				Dict{"id": "3", "name": "Item 3"},
			},
			"targetId": "2",
		}

		result := find("items", "id", "targetId", data, &missingKeys)
		expected := Dict{"id": "2", "name": "Item 2"}
		assert.Equal(t, expected, result)
	})

	t.Run("findByValue function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"items": []any{
				Dict{"id": "1", "name": "Item 1"},
				Dict{"id": "2", "name": "Item 2"},
				Dict{"id": "3", "name": "Item 3"},
			},
		}

		result := findByValue("items", "id", "2", data, &missingKeys)
		expected := Dict{"id": "2", "name": "Item 2"}
		assert.Equal(t, expected, result)
	})

	t.Run("getAtIndex function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"items": []any{"first", "second", "third"},
		}

		result := getAtIndex("items", 1, data, &missingKeys)
		assert.Equal(t, "second", result)
	})

	t.Run("slice function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"items": []any{"a", "b", "c", "d", "e"},
		}

		result := slice("items", 1, 4, data, &missingKeys)
		expected := []any{"b", "c", "d"}
		assert.Equal(t, expected, result)
	})

	t.Run("merge function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"target": Dict{"a": 1, "b": 2},
			"source": Dict{"c": 3, "d": 4},
		}

		result := merge("target", "source", data, &missingKeys)
		expected := Dict{"a": 1, "b": 2, "c": 3, "d": 4}
		assert.Equal(t, expected, result)
	})

	t.Run("incrementCounter function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"counter": 5,
		}

		result := incrementCounter("counter", data, &missingKeys)
		assert.Equal(t, 6, result)

		// Test with missing counter
		result = incrementCounter("missingCounter", data, &missingKeys)
		assert.Equal(t, 1, result)
	})

	t.Run("incrementCounterBy function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"counter": 10,
		}

		result := incrementCounterBy("counter", 5, data, &missingKeys)
		assert.Equal(t, 15, result)
	})

	t.Run("addkeytoall function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"memories": []any{
				Dict{"ID": "1", "content": "memory 1"},
				Dict{"ID": "2", "content": "memory 2"},
			},
		}

		result := addkeytoall("memories", "resource_id", "resource-123", data, &missingKeys)
		expected := []any{
			Dict{"ID": "1", "content": "memory 1", "resource_id": "resource-123"},
			Dict{"ID": "2", "content": "memory 2", "resource_id": "resource-123"},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("concat function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"first":  "Hello",
			"second": "World",
		}

		result := concat("first", "second", data, &missingKeys)
		assert.Equal(t, "HelloWorld", result)
	})

	t.Run("getOrOriginal function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"existingKey": "value",
		}

		// Test with existing key
		result := getOrOriginal("existingKey", data, &missingKeys)
		assert.Equal(t, "value", result)

		// Test with missing key
		result = getOrOriginal("missingKey", data, &missingKeys)
		assert.Equal(t, "missingKey", result)
	})

	t.Run("coalescelist function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"myList": []any{"a", "b", "c"},
		}

		result := coalescelist([]string{"missingList", "myList"}, data, &missingKeys)
		expected := []any{"a", "b", "c"}
		assert.Equal(t, expected, result)
	})

	t.Run("filter function", func(t *testing.T) {
		var missingKeys []string
		data := Dict{
			"items": []any{
				Dict{"type": "A", "value": 1},
				Dict{"type": "B", "value": 2},
				Dict{"type": "A", "value": 3},
			},
		}

		result := filter("items", "type", "A", data, &missingKeys)
		expected := []any{
			Dict{"type": "A", "value": 1},
			Dict{"type": "A", "value": 3},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("truthy function", func(t *testing.T) {
		data := Dict{
			"trueValue":  true,
			"falseValue": false,
			"emptyStr":   "",
			"nonEmpty":   "value",
			"zero":       0,
			"nonZero":    42,
			"emptyList":  []any{},
			"nonEmptyList": []any{1, 2, 3},
		}

		assert.True(t, truthy("trueValue", data))
		assert.False(t, truthy("falseValue", data))
		assert.False(t, truthy("emptyStr", data))
		assert.True(t, truthy("nonEmpty", data))
		assert.False(t, truthy("zero", data))
		assert.True(t, truthy("nonZero", data))
		assert.False(t, truthy("emptyList", data))
		assert.True(t, truthy("nonEmptyList", data))
		assert.False(t, truthy("missingKey", data))
	})

	t.Run("truthyValue function", func(t *testing.T) {
		assert.True(t, truthyValue(true))
		assert.False(t, truthyValue(false))
		assert.False(t, truthyValue(""))
		assert.True(t, truthyValue("non-empty"))
		assert.False(t, truthyValue([]any{}))
		assert.True(t, truthyValue([]any{1, 2, 3}))
		assert.False(t, truthyValue(nil))
		assert.True(t, truthyValue(42))
	})

	t.Run("toString function", func(t *testing.T) {
		assert.Equal(t, "42", toString(42))
		assert.Equal(t, "hello", toString("hello"))
		assert.Equal(t, "true", toString(true))
		assert.Equal(t, "", toString(nil))
	})

	t.Run("truncateString function", func(t *testing.T) {
		assert.Equal(t, "hello", truncateString("hello", 10))
		assert.Equal(t, "hello...", truncateString("hello world", 5))
		assert.Equal(t, "test", truncateString("test", 4))
	})

	t.Run("regexReplace function", func(t *testing.T) {
		result := regexReplace(`\d+`, "X", "abc123def456")
		assert.Equal(t, "abcXdefX", result)

		result = regexReplace(`[aeiou]`, "*", "hello world")
		assert.Equal(t, "h*ll* w*rld", result)
	})

	t.Run("list function", func(t *testing.T) {
		result := list("a", "b", "c")
		expected := []any{"a", "b", "c"}
		assert.Equal(t, expected, result)

		result = list()
		expected = []any{}
		assert.Equal(t, expected, result)
	})

	t.Run("add function", func(t *testing.T) {
		result, err := add(10, 5)
		assert.NoError(t, err)
		assert.Equal(t, float64(15), result)

		result, err = add("10", "5")
		assert.NoError(t, err)
		assert.Equal(t, float64(15), result)
	})

	t.Run("sub function", func(t *testing.T) {
		result, err := sub(10, 5)
		assert.NoError(t, err)
		assert.Equal(t, float64(5), result)
	})

	t.Run("gt function", func(t *testing.T) {
		result, err := gt(10, 5)
		assert.NoError(t, err)
		assert.True(t, result)

		result, err = gt(5, 10)
		assert.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("lt function", func(t *testing.T) {
		result, err := lt(5, 10)
		assert.NoError(t, err)
		assert.True(t, result)

		result, err = lt(10, 5)
		assert.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("len function", func(t *testing.T) {
		assert.Equal(t, 3, _len([]any{1, 2, 3}))
		assert.Equal(t, 5, _len("hello"))
		assert.Equal(t, 2, _len(map[string]any{"a": 1, "b": 2}))
		assert.Equal(t, 0, _len(nil))
	})

	t.Run("toJSON function", func(t *testing.T) {
		result := toJSON(Dict{"key": "value", "number": 42})
		assert.Contains(t, result, `"key":"value"`)
		assert.Contains(t, result, `"number":42`)
	})

	t.Run("mergeRaw function", func(t *testing.T) {
		result := mergeRaw([]any{1, 2}, []any{3, 4})
		expected := []any{1, 2, 3, 4}
		assert.Equal(t, expected, result)
	})

	t.Run("nilToEmptyString function", func(t *testing.T) {
		assert.Equal(t, "", nilToEmptyString(nil))
		assert.Equal(t, "42", nilToEmptyString(42))
		assert.Equal(t, "hello", nilToEmptyString("hello"))
	})

	t.Run("noop function", func(t *testing.T) {
		assert.Equal(t, "test", noop("test"))
		assert.Equal(t, 42, noop(42))
		assert.Equal(t, nil, noop(nil))
	})
}

func TestDedupeBy(t *testing.T) {
	t.Parallel()

	// Create test data
	testData := Dict{
		"simpleItems": []any{
			Dict{"id": "1", "name": "Item 1"},
			Dict{"id": "2", "name": "Item 2"},
			Dict{"id": "1", "name": "Item 1 Duplicate"},
			Dict{"id": "3", "name": "Item 3"},
		},
		"complexItems": []any{
			Dict{
				"ID":      "1",
				"Content": "First item",
				"CreatedAt": map[string]any{
					"Time":  time.Now().Format(time.RFC3339),
					"Valid": true,
				},
			},
			Dict{
				"ID":      "2",
				"Content": "Second item",
				"CreatedAt": map[string]any{
					"Time":  time.Now().Format(time.RFC3339),
					"Valid": true,
				},
			},
			Dict{
				"ID":      "1", // Duplicate ID
				"Content": "First item duplicate",
				"CreatedAt": map[string]any{
					"Time":  time.Now().Format(time.RFC3339),
					"Valid": true,
				},
			},
			Dict{
				"ID":      "3",
				"Content": "Third item",
				"CreatedAt": map[string]any{
					"Time":  time.Now().Format(time.RFC3339),
					"Valid": true,
				},
			},
		},
		"nestedItems": []any{
			Dict{
				"metadata": Dict{
					"id":   "A",
					"type": "first",
				},
				"content": "Content A",
			},
			Dict{
				"metadata": Dict{
					"id":   "B",
					"type": "second",
				},
				"content": "Content B",
			},
			Dict{
				"metadata": Dict{
					"id":   "A", // Duplicate nested ID
					"type": "third",
				},
				"content": "Content A duplicate",
			},
		},
	}

	// Test cases
	testCases := []struct {
		name          string
		arrayKey      string
		fieldKey      string
		expectedCount int
	}{
		{
			name:          "Simple deduplication by ID",
			arrayKey:      "simpleItems",
			fieldKey:      "id",
			expectedCount: 3, // 3 unique IDs (1, 2, 3)
		},
		{
			name:          "Complex object deduplication by ID",
			arrayKey:      "complexItems",
			fieldKey:      "ID",
			expectedCount: 3, // 3 unique IDs (1, 2, 3)
		},
		{
			name:          "Nested field deduplication",
			arrayKey:      "nestedItems",
			fieldKey:      "metadata.id",
			expectedCount: 2, // 2 unique IDs (A, B)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var missingKeys []string
			result := dedupeBy(tc.arrayKey, tc.fieldKey, testData, &missingKeys)

			assert.Equal(t, tc.expectedCount, len(result), "Expected %d items after deduplication", tc.expectedCount)

			// Verify no duplicates exist in result
			seen := make(map[string]bool)
			for _, item := range result {
				itemDict, ok := item.(Dict)
				if !ok {
					t.Fatalf("Expected Dict item, got %T", item)
				}

				// Extract the field value, handling nested fields
				var fieldValue any
				if tc.name == "Nested field deduplication" {
					metadata, ok := itemDict["metadata"].(Dict)
					if !ok {
						t.Fatalf("Expected metadata to be Dict, got %T", itemDict["metadata"])
					}
					fieldValue = metadata["id"]
				} else {
					fieldValue = itemDict[tc.fieldKey]
				}

				valueStr := toString(fieldValue)
				assert.False(t, seen[valueStr], "Found duplicate ID %s after deduplication", valueStr)
				seen[valueStr] = true
			}
		})
	}
}

func TestAddkey(t *testing.T) {
	tests := []struct {
		name        string
		toObj       string
		key         string
		valueKey    string
		data        Dict
		missingKeys *[]string
		expected    Dict
	}{
		{
			name:        "Add key to existing object",
			toObj:       "object",
			key:         "newKey",
			valueKey:    "value",
			data:        Dict{"object": Dict{"existingKey": "existingValue"}, "value": "newValue"},
			missingKeys: &[]string{},
			expected:    Dict{"existingKey": "existingValue", "newKey": "newValue"},
		},
		{
			name:        "Add key to empty object",
			toObj:       "emptyObject",
			key:         "firstKey",
			valueKey:    "value",
			data:        Dict{"emptyObject": Dict{}, "value": "someValue"},
			missingKeys: &[]string{},
			expected:    Dict{"firstKey": "someValue"},
		},
		{
			name:        "Overwrite existing key",
			toObj:       "object",
			key:         "existingKey",
			valueKey:    "newValue",
			data:        Dict{"object": Dict{"existingKey": "oldValue"}, "newValue": "updatedValue"},
			missingKeys: &[]string{},
			expected:    Dict{"existingKey": "updatedValue"},
		},
		{
			name:        "Add nested value",
			toObj:       "object",
			key:         "nested",
			valueKey:    "nestedValue",
			data:        Dict{"object": Dict{}, "nestedValue": Dict{"a": 1, "b": 2}},
			missingKeys: &[]string{},
			expected:    Dict{"nested": Dict{"a": 1, "b": 2}},
		},
		{
			name:        "Object not found",
			toObj:       "nonExistentObject",
			key:         "key",
			valueKey:    "value",
			data:        Dict{"value": "someValue"},
			missingKeys: &[]string{},
			expected:    nil, // Should return nil since object doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addkey(tt.toObj, tt.key, tt.valueKey, tt.data, tt.missingKeys)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.Equal(t, tt.expected, result)

			// For the case where we're missing a key, check that it was added to missingKeys
			if tt.name == "Object not found" {
				assert.Contains(t, *tt.missingKeys, tt.toObj)
			}
		})
	}
}

func TestExtractSlice(t *testing.T) {
	tests := []struct {
		name        string
		array       string
		field       string
		data        Dict
		missingKeys *[]string
		expected    []any
	}{
		{
			name:  "Extract objects from array",
			array: "items",
			field: "memory",
			data: Dict{
				"items": []any{
					Dict{
						"memory": Dict{
							"ID":      "1",
							"Content": "Memory 1",
						},
						"distance": 0.5,
					},
					Dict{
						"memory": Dict{
							"ID":      "2",
							"Content": "Memory 2",
						},
						"distance": 0.3,
					},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				Dict{
					"ID":      "1",
					"Content": "Memory 1",
				},
				Dict{
					"ID":      "2",
					"Content": "Memory 2",
				},
			},
		},
		{
			name:  "Extract string values",
			array: "people",
			field: "name",
			data: Dict{
				"people": []any{
					Dict{"name": "Alice", "age": 25},
					Dict{"name": "Bob", "age": 30},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				"Alice",
				"Bob",
			},
		},
		{
			name:  "Extract numeric values",
			array: "people",
			field: "age",
			data: Dict{
				"people": []any{
					Dict{"name": "Alice", "age": 25},
					Dict{"name": "Bob", "age": 30},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				25,
				30,
			},
		},
		{
			name:  "Empty array",
			array: "items",
			field: "memory",
			data: Dict{
				"items": []any{},
			},
			missingKeys: &[]string{},
			expected:    []any{},
		},
		{
			name:        "Array not found",
			array:       "nonexistent",
			field:       "memory",
			data:        Dict{},
			missingKeys: &[]string{},
			expected:    []any{},
		},
		{
			name:  "Missing field in some items",
			array: "items",
			field: "memory",
			data: Dict{
				"items": []any{
					Dict{
						"memory": Dict{
							"ID":      "1",
							"Content": "Memory 1",
						},
					},
					Dict{
						"other": "value",
					},
					Dict{
						"memory": Dict{
							"ID":      "2",
							"Content": "Memory 2",
						},
					},
				},
			},
			missingKeys: &[]string{},
			expected: []any{
				Dict{
					"ID":      "1",
					"Content": "Memory 1",
				},
				Dict{
					"ID":      "2",
					"Content": "Memory 2",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSlice(tt.array, tt.field, tt.data, tt.missingKeys)
			assert.Equal(t, tt.expected, result)
		})
	}
}