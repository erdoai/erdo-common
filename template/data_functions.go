package template

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	. "github.com/erdoai/erdo-common/utils"
)

// addMissingKey adds a key to the missingKeys slice if it's not already present
func addMissingKey(missingKeys *[]string, key string) {
	for _, k := range *missingKeys {
		if k == key {
			return // Key already exists, don't add duplicate
		}
	}
	*missingKeys = append(*missingKeys, key)
}

// Functions that require .Data and .MissingKeys parameters
var dataFuncMap = template.FuncMap{
	"get":                          get,
	"concat":                       concat,
	"getOrOriginal":                getOrOriginal,
	"sliceEnd":                     sliceEnd,
	"sliceEndKeepFirstUserMessage": sliceEndKeepFirstUserMessage,
	"slice":                        slice,
	"extractSlice":                 extractSlice,
	"dedupeBy":                     dedupeBy,
	"find":                         find,
	"findByValue":                  findByValue,
	"getAtIndex":                   getAtIndex,
	"merge":                        merge,
	"coalescelist":                 coalescelist,
	"addkey":                       addkey,
	"removekey":                    removekey,
	"mapToDict":                    mapToDict,
	"mapToArray":                   mapToArray,
	"addkeytoall":                  addkeytoall,
	"incrementCounter":             incrementCounter,
	"incrementCounterBy":           incrementCounterBy,
	"coalesce":                     coalesce,
	"filter":                       filter,
}

func addkey(toObj string, key string, value any, data map[string]any, missingKeys *[]string) map[string]any {
	_obj := get(toObj, data, missingKeys)
	obj, ok := _obj.(map[string]any)
	if !ok {
		log.Printf("Error casting to dict in addkey: %T %v", _obj, _obj)
		return nil
	}

	result, err := Set(obj, key, value)
	if err != nil {
		log.Printf("Error setting key %v to value %v in addkey: %v", key, value, err)
		return obj
	}

	return result
}

func removekey(toObj string, key string, data map[string]any, missingKeys *[]string) map[string]any {
	_obj := get(toObj, data, missingKeys)
	obj, ok := _obj.(map[string]any)
	if !ok {
		log.Printf("Error casting to dict in removekey: %T %v", _obj, _obj)
		return obj
	}

	delete(obj, key)

	return obj
}

// mapToDict converts a list of values to a list of dictionaries with a specified key
// Example: {{map "myList" "myKey"}} will convert ["value1", "value2"] to [{"myKey": "value1"}, {"myKey": "value2"}]
func mapToDict(listKey string, dictKey string, data map[string]any, missingKeys *[]string) []map[string]any {
	_list := get(listKey, data, missingKeys)
	if _list == nil {
		// Remove the key from missingKeys if it was added
		// This is because we want to return an empty list for non-existent lists
		// rather than treating it as a missing key
		for i, key := range *missingKeys {
			if key == listKey {
				*missingKeys = append((*missingKeys)[:i], (*missingKeys)[i+1:]...)
				break
			}
		}
		return []map[string]any{}
	}

	list := ToAnySlice(_list)
	if list == nil {
		log.Printf("Error casting to list in mapToDict: %T %v", _list, _list)
		return []map[string]any{}
	}

	result := make([]map[string]any, 0, len(list))
	for _, item := range list {
		result = append(result, map[string]any{dictKey: item})
	}

	return result
}

func coalescelist(list string, data map[string]any, missingKeys *[]string) []any {
	_list := get(list, data, missingKeys)
	slice := ToAnySlice(_list)
	if slice == nil {
		return []any{}
	}
	return slice
}

func getSliceInt(v any, data map[string]any, missingKeys *[]string) (*int, bool) {
	var ret int
	ok := true
	switch v := v.(type) {
	case int:
		ret = v
	case string:
		// Check if numeric
		if num, err := strconv.Atoi(v); err == nil {
			// If it's a numeric string, use the number directly
			ret = num
		} else {
			// If it's not a numeric string, try to get it from data
			_ret := get(v, data, missingKeys)
			ret, ok = _ret.(int)
			if !ok {
				log.Printf("slice start not ok: %T %v %T %v", _ret, _ret, v, v)
				return nil, false
			}
		}
	default:
		log.Printf("slice start unexpected type: %T %v", v, v)
		return nil, false
	}

	return &ret, ok
}

func slice(array string, start any, end any, data map[string]any, missingKeys *[]string) []any {
	_items := get(array, data, missingKeys)
	if _items == nil {
		log.Printf("slice items is nil")
		addMissingKey(missingKeys, array)
		return []any{}
	}

	items := ToAnySlice(_items)
	if items == nil {
		log.Printf("slice items not ok: %T %v", _items, _items)
		addMissingKey(missingKeys, array)
		return []any{}
	}

	_startInt, ok := getSliceInt(start, data, missingKeys)
	if !ok {
		return []any{}
	}
	startInt := *_startInt

	_endInt, ok := getSliceInt(end, data, missingKeys)
	if !ok {
		return []any{}
	}
	endInt := *_endInt

	if startInt < 0 {
		startInt = 0
	}

	if endInt > len(items) {
		endInt = len(items)
	}

	return items[startInt:endInt]
}

// extractSlice extracts a field from each item in a list and returns those values as a new list
// Example: {{extractSlice "items" "name"}} will extract the name field from each item in the items list
// It supports extracting any type of value - strings, numbers, objects, arrays, etc.
func extractSlice(array string, propertyPath string, data map[string]any, missingKeys *[]string) []any {
	_items := get(array, data, missingKeys)
	if _items == nil {
		return []any{}
	}

	items := ToAnySlice(_items)
	if items == nil {
		return []any{}
	}

	result := make([]any, 0, len(items))
	for _, item := range items {
		if val := get(propertyPath, item, missingKeys); val != nil {
			result = append(result, val)
		}
	}
	return result
}

// dedupeBy removes duplicates from a slice based on a specific field
func dedupeBy(array string, field string, data map[string]any, missingKeys *[]string) []any {
	_items := get(array, data, missingKeys)
	if _items == nil {
		return []any{}
	}

	items := ToAnySlice(_items)
	if items == nil {
		log.Printf("dedupeBy: items is not a slice, got type %T for array %q", _items, array)
		return []any{}
	}

	if len(items) == 0 {
		return []any{}
	}

	seen := make(map[string]bool)
	result := make([]any, 0, len(items))

	for _, item := range items {
		// Use get to extract the value, handling nested fields and different data types properly
		fieldValue := get(field, item, missingKeys)
		if fieldValue == nil {
			// If we can't find the field, just add the item and continue
			result = append(result, item)
			continue
		}

		// Convert the field value to a string for the map key
		valueStr := fmt.Sprintf("%v", fieldValue)

		if !seen[valueStr] {
			seen[valueStr] = true
			result = append(result, item)
		}
	}

	return result
}

// find searches for an item in a slice where the specified field matches the target value
func find(arrayKey string, fieldKey string, targetKey any, data map[string]any, missingKeys *[]string) any {
	_items := get(arrayKey, data, missingKeys)
	if _items == nil {
		return nil
	}

	// Handle targetKey which can be either a string key to look up or an actual value
	var targetValue any
	if targetKeyStr, ok := targetKey.(string); ok {
		// It's a string, try to look it up in data
		targetValue = get(targetKeyStr, data, missingKeys)
		if targetValue == nil {
			// If not found in data, use the string itself as the target value
			targetValue = targetKeyStr
		}
	} else {
		// It's already a resolved value
		targetValue = targetKey
	}

	targetStr := fmt.Sprintf("%v", targetValue)

	items := ToAnySlice(_items)
	if items == nil {
		return nil
	}

	for _, item := range items {
		if value := get(fieldKey, item, missingKeys); value != nil {
			// Convert both values to strings for comparison
			var valueStr string
			switch v := value.(type) {
			case float64:
				// Handle integers stored as float64 in JSON
				if v == float64(int64(v)) {
					valueStr = fmt.Sprintf("%d", int64(v))
				} else {
					valueStr = fmt.Sprintf("%g", v)
				}
			case int:
				valueStr = fmt.Sprintf("%d", v)
			case string:
				valueStr = v
			default:
				valueStr = fmt.Sprintf("%v", v)
			}

			if valueStr == targetStr {
				return item
			}
		}
	}

	return nil
}

// find searches for an item in a slice where the specified field matches the target value
func findByValue(arrayKey string, fieldKey string, targetValue any, data map[string]any, missingKeys *[]string) any {
	_items := get(arrayKey, data, missingKeys)
	if _items == nil {
		return nil
	}

	targetStr := fmt.Sprintf("%v", targetValue)

	items := ToAnySlice(_items)
	if items == nil {
		return nil
	}

	for _, item := range items {
		if value := get(fieldKey, item, missingKeys); value != nil {
			// Convert both values to strings for comparison
			var valueStr string
			switch v := value.(type) {
			case float64:
				// Handle integers stored as float64 in JSON
				if v == float64(int64(v)) {
					valueStr = fmt.Sprintf("%d", int64(v))
				} else {
					valueStr = fmt.Sprintf("%g", v)
				}
			case int:
				valueStr = fmt.Sprintf("%d", v)
			case string:
				valueStr = v
			default:
				valueStr = fmt.Sprintf("%v", v)
			}

			if valueStr == targetStr {
				return item
			}
		}
	}

	log.Printf("findByValue: no item found for arrayKey %s, fieldKey %s, targetValue %v", arrayKey, fieldKey, targetValue)

	return nil
}

// getAtIndex returns the item at the specified index in the slice
func getAtIndex(array string, index any, data map[string]any, missingKeys *[]string) any {
	_items := get(array, data, missingKeys)
	if _items == nil {
		return nil
	}

	items := ToAnySlice(_items)
	if items == nil {
		return nil
	}

	// Resolve the index value
	var indexValue int
	switch v := index.(type) {
	case int:
		indexValue = v
	case string:
		// First try to parse as integer
		if indexInt, err := strconv.Atoi(v); err == nil {
			indexValue = indexInt
		} else {
			// If not an integer, treat as a path and resolve it
			resolvedIndex := get(v, data, missingKeys)
			if resolvedIndex != nil {
				switch idx := resolvedIndex.(type) {
				case int:
					indexValue = idx
				case float64:
					indexValue = int(idx)
				case string:
					if parsed, err := strconv.Atoi(idx); err == nil {
						indexValue = parsed
					} else {
						return nil
					}
				default:
					return nil
				}
			} else {
				return nil
			}
		}
	default:
		return nil
	}

	if indexValue >= 0 && indexValue < len(items) {
		return items[indexValue]
	}

	return nil
}

// merge combines multiple slices into one
func merge(array1 string, array2 string, data map[string]any, missingKeys *[]string) []any {
	initialMissingCount := len(*missingKeys)

	items1 := get(array1, data, missingKeys)
	items2 := get(array2, data, missingKeys)

	// Fail fast if either key was missing
	if len(*missingKeys) > initialMissingCount {
		log.Printf("merge: key lookup failed for array1=%q or array2=%q", array1, array2)
		return []any{}
	}

	var slice1, slice2 []any

	// Convert first array
	slice1 = ToAnySlice(items1)
	if slice1 == nil {
		log.Printf("merge: array1 key %q is not a slice, got %T", array1, items1)
		addMissingKey(missingKeys, array1)
		return []any{}
	}

	// Convert second array
	slice2 = ToAnySlice(items2)
	if slice2 == nil {
		log.Printf("merge: array2 key %q is not a slice, got %T", array2, items2)
		addMissingKey(missingKeys, array2)
		return []any{}
	}

	// Combine slices
	result := make([]any, 0, len(slice1)+len(slice2))
	result = append(result, slice1...)
	result = append(result, slice2...)
	return result
}

// getOrOriginal retrieves a value from data using the specified key.
// If the key is optional (has ? suffix or is defined as optional in keyDefinitions)
// and the value is nil, it returns nil.
// Otherwise, it returns the original template string if the value is nil.
func getOrOriginal(key string, keyDefinitions KeyDefinitions, data map[string]any, missingKeys *[]string) any {
	// Check if parameter is optional
	isOptional := strings.HasSuffix(key, "?")
	cleanKey := strings.TrimSuffix(key, "?")

	// Remove .Data. prefix if present
	cleanKey = removeDataPrefix(cleanKey)

	value := get(cleanKey, data, missingKeys)

	if value == nil {
		if isOptional || isKeyOptional(key, keyDefinitions) {
			return nil
		}
		return fmt.Sprintf("{{%s}}", cleanKey)
	}
	return value
}

func sliceEnd(sliceKey string, n int, data map[string]any, missingKeys *[]string) []any {
	_slice := get(sliceKey, data, missingKeys)
	if _slice == nil {
		log.Printf("sliceEnd: key %q resolved to nil", sliceKey)
		addMissingKey(missingKeys, sliceKey)
		return nil
	}

	slice := ToAnySlice(_slice)
	if slice == nil {
		log.Printf("sliceEnd: key %q is not a slice, got %T", sliceKey, _slice)
		addMissingKey(missingKeys, sliceKey)
		return nil
	}

	if n >= len(slice) {
		return slice
	}

	return slice[len(slice)-n:]
}

// sliceEndKeepFirstUserMessage returns the last n messages but always keeps the first user message
// if it exists. This is useful for maintaining context while limiting message history.
func sliceEndKeepFirstUserMessage(sliceKey string, n int, data map[string]any, missingKeys *[]string) []any {
	_slice := get(sliceKey, data, missingKeys)
	if _slice == nil {
		addMissingKey(missingKeys, sliceKey)
		return nil
	}

	// Convert to []any regardless of the underlying type ([]any, []Message, etc.)
	slice := ToAnySlice(_slice)
	if slice == nil {
		addMissingKey(missingKeys, sliceKey)
		return nil
	}

	if len(slice) == 0 {
		return slice
	}

	// If we want all messages or more than available, return everything
	if n >= len(slice) {
		return slice
	}

	// Get the last n messages first
	lastNMessages := slice[len(slice)-n:]

	// Check if the first message in the slice is already a user message
	if len(lastNMessages) > 0 {
		if isUserMessage(lastNMessages[0]) {
			return lastNMessages
		}
	}

	// Find the first user message in the part that would be cut off (before the slice start)
	// We search backwards from the slice start point to find the most recent user message
	sliceStartIndex := len(slice) - n
	var firstUserMessage any
	for i := sliceStartIndex - 1; i >= 0; i-- {
		if isUserMessage(slice[i]) {
			firstUserMessage = slice[i]
			break
		}
	}

	// If there's no user message in the cut-off part, just return the last n messages
	if firstUserMessage == nil {
		return lastNMessages
	}

	// Prepend the first user message to the last n messages
	result := make([]any, 0, n+1)
	result = append(result, firstUserMessage)
	result = append(result, lastNMessages...)

	return result
}

// isUserMessage checks if a message has role "user"
// Handles both map[string]any and struct types with Role field
func isUserMessage(msg any) bool {
	// Use common utility function that works for both maps and structs
	roleVal := GetFieldValue(msg, "role")
	if roleVal == nil {
		return false
	}

	roleStr, ok := roleVal.(string)
	if !ok {
		return false
	}

	return roleStr == "user"
}

// get retrieves a value from a nested data structure using a dot-separated key path.
// It resolves paths through dictionaries and slices, handles optional keys, and tracks
// missing required keys in the missingKeys slice.
// For example:
// - "user.name" will navigate to the "name" field in the "user" dictionary
// - "items.0.name" will navigate to the "name" field in the first element of the "items" slice
// Returns nil if the key is not found or can't be accessed.
func get(key string, data any, missingKeys *[]string) any {
	lookupKey, isOptional := cleanKey(key)

	parts := strings.FieldsFunc(lookupKey, func(r rune) bool {
		return r == '.' || r == '[' || r == ']'
	})
	current := data

	if current == nil {
		log.Printf("get: data is nil for key %q", lookupKey)
		handleMissingKey(lookupKey, isOptional, missingKeys)
		return nil
	}

	for _, part := range parts {
		if strings.HasPrefix(part, "\"") || strings.HasSuffix(part, "\"") || strings.HasPrefix(part, "'") {
			log.Printf("get: key part incorrectly has wrapping quotes: %s", part)
		}
		switch m := current.(type) {
		case map[string]any:
			if val, exists := m[part]; exists {
				current = val
			} else {
				log.Printf("get: key %q not found in dict at path %q", part, lookupKey)
				handleMissingKey(lookupKey, isOptional, missingKeys)
				return nil
			}
		case []any:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(m) {
				log.Printf("get: invalid array index %q at path %q", part, lookupKey)
				handleMissingKey(lookupKey, isOptional, missingKeys)
				return nil
			}
			current = m[index]
		case []map[string]any:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(m) {
				log.Printf("get: invalid array index %q at path %q", part, lookupKey)
				handleMissingKey(lookupKey, isOptional, missingKeys)
				return nil
			}
			current = m[index]
		default:
			// Use reflection to handle other slice/map types
			reflectVal := reflect.ValueOf(current)
			kind := reflectVal.Kind()
			if kind == reflect.Map {
				// Try to access as a map
				key := reflect.ValueOf(part)
				val := reflectVal.MapIndex(key)
				if val.IsValid() {
					current = val.Interface()
				} else {
					log.Printf("get: key %q not found in map at path %q", part, lookupKey)
					handleMissingKey(lookupKey, isOptional, missingKeys)
					return nil
				}
			} else if kind == reflect.Slice || kind == reflect.Array {
				// Try to access as an array
				index, err := strconv.Atoi(part)
				if err != nil || index < 0 || index >= reflectVal.Len() {
					log.Printf("get: invalid array index %q at path %q", part, lookupKey)
					handleMissingKey(lookupKey, isOptional, missingKeys)
					return nil
				}
				current = reflectVal.Index(index).Interface()
			} else {
				log.Printf("get: cannot access %q in type %T at path %q", part, current, lookupKey)
				handleMissingKey(lookupKey, isOptional, missingKeys)
				return nil
			}
		}
	}

	// If we get a float64 that's actually an int, convert it back
	if num, ok := current.(float64); ok && num == float64(int(num)) {
		return int(num)
	}

	return current
}

// concat extracts values from a slice of dictionaries and joins them with a separator.
// It retrieves a slice using the 'key' parameter, then for each item in the slice,
// it extracts the value at 'property' and concatenates these values with 'sep'.
// For example, concat(", ", "users", "name", ...) would extract the "name" field from
// each object in the "users" slice and join them with commas.
// Returns the original template if the key is not found or the slice is empty.
func concat(sep string, key string, property string, data any, missingKeys *[]string) string {
	// Escape special characters in sep and property
	escapedSep := template.JSEscapeString(sep)
	escapedProperty := template.JSEscapeString(property)
	escapedKey := template.JSEscapeString(key)
	funcCall := fmt.Sprintf("concat \"%s\" \"%s\" \"%s\"", escapedSep, escapedKey, escapedProperty)
	original := fmt.Sprintf("{{%s $.Data $.MissingKeys}}", funcCall)

	items := get(key, data, missingKeys)
	if items == nil {
		return original
	}

	itemsSlice := ToAnySlice(items)
	if itemsSlice == nil {
		addMissingKey(missingKeys, escapedKey)
		return original
	}

	itemsStrs := []string{}
	for _, item := range itemsSlice {
		// Use GetFieldValue to work with both maps and structs
		propVal := GetFieldValue(item, property)
		if propVal == nil {
			// if we can't get the property, skip it
			continue
		}
		itemStr, ok := propVal.(string)
		if !ok {
			// if property is not a string, skip it
			continue
		}
		itemsStrs = append(itemsStrs, itemStr)
	}

	if len(itemsStrs) == 0 {
		addMissingKey(missingKeys, escapedKey)
		return original
	}

	return strings.Join(itemsStrs, sep)
}

// addkeytoall adds a key with the same value to all items in a list
// Example: {{addkeytoall "myList" "newKey" value}} adds the key "newKey" with value to each item in "myList"
// Supports nested keys using dot notation (e.g., "metadata.resource_id")
func addkeytoall(listKey string, key string, value any, data map[string]any, missingKeys *[]string) []any {
	_list := get(listKey, data, missingKeys)
	if _list == nil {
		return []any{}
	}

	list := ToAnySlice(_list)
	if list == nil {
		log.Printf("Error casting to list in addkeytoall: %T %v", _list, _list)
		return []any{}
	}

	result := make([]any, 0, len(list))
	for _, item := range list {
		// Try to convert to Dict regardless of original type
		var dictItem map[string]any

		if mapItem, ok := item.(map[string]any); ok {
			dictItem = map[string]any(mapItem)
		} else if di, ok := item.(map[string]any); ok {
			dictItem = di
		} else {
			// For non-dictionary types, we can't add a key, so just preserve the item
			result = append(result, item)
			continue
		}

		// Now that we have a Dict, use Set
		updatedDict, err := Set(dictItem, key, value)
		if err != nil {
			log.Printf("Error setting key %s in addkeytoall: %v", key, err)
			result = append(result, item) // Add original if error
		} else {
			result = append(result, updatedDict)
		}
	}

	return result
}

// incrementCounter increments a named counter in the data and returns the new value
// If the counter doesn't exist, it starts at 1. If increment is not provided, defaults to 1.
func incrementCounter(counterName string, data map[string]any, missingKeys *[]string) int {
	return incrementCounterBy(counterName, 1, data, missingKeys)
}

// incrementCounterBy increments a named counter by a specific amount and returns the new value
func incrementCounterBy(counterName string, increment int, data map[string]any, missingKeys *[]string) int {
	currentValue := 0

	// Get current value if it exists
	if existingValue := get(counterName, data, &[]string{}); existingValue != nil {
		switch v := existingValue.(type) {
		case int:
			currentValue = v
		case float64:
			currentValue = int(v)
		case string:
			if parsed, err := strconv.Atoi(v); err == nil {
				currentValue = parsed
			}
		}
	}

	newValue := currentValue + increment

	// Update the data map directly
	data[counterName] = newValue

	return newValue
}

// coalesce returns the first non-nil value from the arguments
// The first argument (key) is treated as a template variable to look up in data
// The second argument (fallbackValue) is treated as a literal value to return if the key is missing
// Example: {{coalesce "optional_key?" 0}} returns the value of optional_key or 0 if nil/missing
// Example: {{coalesce "optional_key?" "default"}} returns the value of optional_key or "default" if nil/missing
func coalesce(key any, fallbackValue any, data map[string]any, missingKeys *[]string) any {
	// Handle the first argument (key) - try to look it up in data if it's a string
	if strKey, ok := key.(string); ok {
		dataValue := get(strKey, data, &[]string{}) // Don't add to missingKeys for coalesce
		if dataValue != nil {
			return dataValue
		}
	} else if key != nil {
		// If the key is not a string but is not nil, return it
		return key
	}

	// If the key was nil/missing, return the fallback value
	// Handle the case where Go's template engine converted numeric literals to strings
	if strFallback, ok := fallbackValue.(string); ok {
		// Try to parse as number first (same logic as processArgument)
		if value, err := strconv.ParseFloat(strFallback, 64); err == nil {
			// Convert float64 that's actually an int back to int
			if value == float64(int(value)) {
				return int(value)
			}
			return value
		}

		// Try to parse as boolean
		if value, err := strconv.ParseBool(strFallback); err == nil {
			return value
		}

		// If it's a quoted string, remove quotes and return as string literal
		if (strings.HasPrefix(strFallback, "\"") && strings.HasSuffix(strFallback, "\"")) ||
			(strings.HasPrefix(strFallback, "'") && strings.HasSuffix(strFallback, "'")) {
			return strFallback[1 : len(strFallback)-1]
		}

		// Return as string if not parseable as number/boolean
		return strFallback
	}

	// Return the fallback value as-is for non-string types
	return fallbackValue
}

// filter filters a slice to only include items where the specified field matches the given value
// Supports operators: "eq" (equals), "in" (value in list)
func filter(key string, field string, operator string, value any, data map[string]any, missingKeys *[]string) []any {
	val := get(key, data, missingKeys)
	arr := ToAnySlice(val)
	if arr == nil {
		return []any{}
	}

	// Pre-allocate with input capacity (worst case all items match)
	result := make([]any, 0, len(arr))
	for _, item := range arr {
		// Use GetFieldValue to work with both maps and structs
		fieldVal := GetFieldValue(item, field)

		var matches bool
		switch operator {
		case "eq":
			// Convert both values to the same type for comparison
			// Handle the case where value comes from template as string but field is numeric
			if fieldVal != nil && value != nil {
				if _, ok := fieldVal.(int); ok {
					if valueStr, ok := value.(string); ok {
						if valueInt, err := strconv.Atoi(valueStr); err == nil {
							value = valueInt
						}
					}
				} else if _, ok := fieldVal.(float64); ok {
					if valueStr, ok := value.(string); ok {
						if valueFloat, err := strconv.ParseFloat(valueStr, 64); err == nil {
							value = valueFloat
						}
					}
				}
			}
			matches = reflect.DeepEqual(fieldVal, value)
		case "in":
			// Check if fieldVal is in the value list
			if valueList, ok := value.([]any); ok {
				for _, listItem := range valueList {
					if reflect.DeepEqual(fieldVal, listItem) {
						matches = true
						break
					}
				}
			}
		default:
			// Default to equals for unknown operators
			matches = reflect.DeepEqual(fieldVal, value)
		}

		if matches {
			result = append(result, item)
		}
	}
	return result
}

// mapToArray converts a map to an array of objects with "key" and "value" fields
// Example: {"old_id_1": "new_id_1", "old_id_2": "new_id_2"} becomes
// [{"key": "old_id_1", "value": "new_id_1"}, {"key": "old_id_2", "value": "new_id_2"}]
func mapToArray(mapKey string, data map[string]any, missingKeys *[]string) []map[string]any {
	_map := get(mapKey, data, missingKeys)
	if _map == nil {
		// Remove the key from missingKeys if it was added
		// This is because we want to return an empty array for non-existent maps
		// rather than treating it as a missing key
		for i, key := range *missingKeys {
			if key == mapKey {
				*missingKeys = append((*missingKeys)[:i], (*missingKeys)[i+1:]...)
				break
			}
		}
		return []map[string]any{}
	}

	mapVal, ok := _map.(map[string]any)
	if !ok {
		log.Printf("Error casting to map in mapToArray: %T %v", _map, _map)
		return []map[string]any{}
	}

	result := make([]map[string]any, 0, len(mapVal))
	for key, value := range mapVal {
		result = append(result, map[string]any{
			"key":   key,
			"value": value,
		})
	}

	return result
}
