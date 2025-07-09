package template

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

// cleanKey processes a key string to determine if it's optional and returns the cleaned key.
// Returns the key without the optional marker and a boolean indicating if it was optional.
func cleanKey(key string) (string, bool) {
	// Check if parameter is optional (has ? suffix)
	isOptional := strings.HasSuffix(key, "?")
	cleanKey := strings.TrimSuffix(key, "?")

	// Remove .Data. prefix if present
	if strings.HasPrefix(cleanKey, ".Data.") {
		cleanKey = strings.TrimPrefix(cleanKey, ".Data.")
	}
	if strings.HasPrefix(cleanKey, "$.Data.") {
		cleanKey = strings.TrimPrefix(cleanKey, "$.Data.")
	}

	return cleanKey, isOptional
}

// handleMissingKey adds a key to the missingKeys slice only if the key is not optional.
// This is used to track which required keys are missing from the data during hydration.
func handleMissingKey(key string, isOptional bool, missingKeys *[]string) {
	if !isOptional && missingKeys != nil {
		*missingKeys = append(*missingKeys, key)
	}
}

// Basic functions that don't require .Data and .MissingKeys
var basicFuncMap = template.FuncMap{
	"truthy":           truthy,
	"toJSON":           toJSON,
	"len":              _len,
	"add":              add,
	"sub":              sub,
	"gt":               gt,
	"lt":               lt,
	"mergeRaw":         mergeRaw,
	"nilToEmptyString": nilToEmptyString,
	"truthyValue":      truthyValue,
	"toString":         toString,
	"truncateString":   truncateString,
	"regexReplace":     regexReplace,
	"noop":             noop,
	"list":             list,
}

// Functions that require .Data and .MissingKeys parameters
var dataFuncMap = template.FuncMap{
	"get":                          get,
	"concat":                       concat,
	"getOrOriginal":                getOrOriginal,
	"sliceEnd":                     sliceEnd,
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
	"addkeytoall":                  addkeytoall,
	"incrementCounter":             incrementCounter,
	"incrementCounterBy":           incrementCounterBy,
	"coalesce":                     coalesce,
	"filter":                       filter,
	"sliceEndKeepFirstUserMessage": sliceEndKeepFirstUserMessage,
}

// Basic template functions (no data dependency)

func truthy(key string, data map[string]any) bool {
	value, exists := data[key]
	if !exists {
		return false
	}
	
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() != 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() != 0
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0
	case []any:
		return len(v) > 0
	case map[string]any:
		return len(v) > 0
	case nil:
		return false
	default:
		return true
	}
}

func toJSON(v any) string {
	jsonStr, err := ToJSON(v)
	if err != nil {
		log.Printf("Error marshaling to JSON: %v", err)
		return ""
	}
	return jsonStr
}

func _len(a any) int {
	if a == nil {
		return 0
	}

	switch v := a.(type) {
	case []any:
		return len(v)
	case string:
		return len(v)
	case map[string]any:
		return len(v)
	case []map[string]any:
		return len(v)
	}

	// Try reflection for other slice types
	rv := reflect.ValueOf(a)
	if rv.Kind() == reflect.Slice {
		return rv.Len()
	}

	log.Printf("unsupported type for len: %T %v", a, a)

	return 0
}

func add(a, b int) int {
	return a + b
}

func sub(a, b int) int {
	return a - b
}

func gt(a, b int) bool {
	return a > b
}

func lt(a, b int) bool {
	return a < b
}

func mergeRaw(array1 []any, array2 []any) []any {
	result := make([]any, 0, len(array1)+len(array2))
	result = append(result, array1...)
	result = append(result, array2...)
	return result
}

func nilToEmptyString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func truthyValue(val any) bool {
	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case []any:
		return len(v) > 0
	case nil:
		return false
	default:
		return true
	}
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func truncateString(s string, length int) string {
	runes := []rune(s)
	if len(runes) <= length {
		return s
	}
	return string(runes[:length]) + "..."
}

func regexReplace(pattern, replacement, text string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Printf("Error compiling regex pattern %s: %v", pattern, err)
		return text
	}
	return re.ReplaceAllString(text, replacement)
}

func noop() string {
	return ""
}

func list(items ...any) []any {
	if len(items) == 0 {
		return []any{}
	}
	return items
}

// Data-dependent functions

// Get retrieves a value from data by key (with dot notation support)
func Get(key string, data map[string]any, missingKeys *[]string) any {
	return get(key, data, missingKeys)
}

func get(key string, data map[string]any, missingKeys *[]string) any {
	lookupKey, isOptional := cleanKey(key)

	parts := strings.FieldsFunc(lookupKey, func(r rune) bool {
		return r == '.' || r == '[' || r == ']'
	})
	current := any(data)

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
				log.Printf("get: key %q not found in map at path %q", part, lookupKey)
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
		default:
			log.Printf("get: cannot access %q in type %T at path %q", part, current, lookupKey)
			handleMissingKey(lookupKey, isOptional, missingKeys)
			return nil
		}
	}

	// If we get a float64 that's actually an int, convert it back
	if num, ok := current.(float64); ok && num == float64(int(num)) {
		return int(num)
	}

	return current
}

func concat(key1, key2 string, data map[string]any, missingKeys *[]string) string {
	val1 := get(key1, data, missingKeys)
	val2 := get(key2, data, missingKeys)
	return fmt.Sprintf("%v%v", val1, val2)
}

func getOrOriginal(key string, data map[string]any, missingKeys *[]string) any {
	// Check if parameter is optional
	isOptional := strings.HasSuffix(key, "?")
	cleanKey := strings.TrimSuffix(key, "?")

	// Remove .Data. prefix if present
	cleanKey = removeDataPrefix(cleanKey)

	value := get(cleanKey, data, missingKeys)

	if value == nil {
		if isOptional {
			return nil
		}
		return fmt.Sprintf("{{%s}}", cleanKey)
	}
	return value
}

func slice(key string, start, end int, data map[string]any, missingKeys *[]string) []any {
	val := get(key, data, missingKeys)
	if arr, ok := val.([]any); ok {
		if start < 0 {
			start = 0
		}
		if end > len(arr) {
			end = len(arr)
		}
		if start >= end {
			return []any{}
		}
		return arr[start:end]
	}
	return []any{}
}

func extractSlice(key string, field string, data map[string]any, missingKeys *[]string) []any {
	val := get(key, data, missingKeys)
	if arr, ok := val.([]any); ok {
		result := make([]any, 0, len(arr))
		for _, item := range arr {
			if mapVal, ok := item.(map[string]any); ok {
				if fieldVal, exists := mapVal[field]; exists {
					result = append(result, fieldVal)
				}
			}
		}
		return result
	}
	return []any{}
}

func dedupeBy(key string, field string, data map[string]any, missingKeys *[]string) []any {
	val := get(key, data, missingKeys)
	if arr, ok := val.([]any); ok {
		seen := make(map[string]bool)
		result := make([]any, 0)
		
		for _, item := range arr {
			var fieldVal any
			if mapVal, ok := item.(map[string]any); ok {
				// Use get to handle nested fields like "metadata.id"
				fieldVal = get(field, mapVal, &[]string{})
			}
			
			key := fmt.Sprintf("%v", fieldVal)
			if !seen[key] {
				seen[key] = true
				result = append(result, item)
			}
		}
		return result
	}
	return []any{}
}

func find(key string, field string, value any, data map[string]any, missingKeys *[]string) any {
	// If value is a string, check if it's a key in data
	if strValue, ok := value.(string); ok {
		if dataValue, exists := data[strValue]; exists {
			value = dataValue
		}
	}
	
	val := get(key, data, missingKeys)
	if arr, ok := val.([]any); ok {
		for _, item := range arr {
			var fieldVal any
			if mapVal, ok := item.(map[string]any); ok {
				fieldVal = mapVal[field]
			}
			
			if reflect.DeepEqual(fieldVal, value) {
				return item
			}
		}
	}
	return nil
}

func findByValue(arrayKey string, fieldKey string, targetValue any, data map[string]any, missingKeys *[]string) any {
	_items := get(arrayKey, data, missingKeys)
	if _items == nil {
		return nil
	}

	targetStr := fmt.Sprintf("%v", targetValue)

	items, ok := _items.([]any)
	if !ok {
		return nil
	}

	for _, item := range items {
		itemDict, ok := item.(map[string]any)
		if !ok {
			continue
		}

		fieldValue := get(fieldKey, itemDict, &[]string{})
		if fieldValue != nil {
			fieldStr := fmt.Sprintf("%v", fieldValue)
			if fieldStr == targetStr {
				return item
			}
		}
	}

	return nil
}

func getAtIndex(key string, index int, data map[string]any, missingKeys *[]string) any {
	val := get(key, data, missingKeys)
	if arr, ok := val.([]any); ok {
		if index >= 0 && index < len(arr) {
			return arr[index]
		}
	}
	return nil
}

func merge(array1 string, array2 string, data map[string]any, missingKeys *[]string) any {
	items1 := get(array1, data, missingKeys)
	items2 := get(array2, data, missingKeys)

	// Check if we're merging dictionaries
	dict1, isDict1 := items1.(map[string]any)
	dict2, isDict2 := items2.(map[string]any)
	
	if isDict1 && isDict2 {
		// Merge dictionaries
		result := make(map[string]any)
		for k, v := range dict1 {
			result[k] = v
		}
		for k, v := range dict2 {
			result[k] = v
		}
		return result
	}
	
	// Otherwise, merge as arrays
	var slice1, slice2 []any

	// Convert first array
	switch v := items1.(type) {
	case []any:
		slice1 = v
	case nil:
		slice1 = []any{}
	default:
		slice1 = []any{v}
	}

	// Convert second array
	switch v := items2.(type) {
	case []any:
		slice2 = v
	case nil:
		slice2 = []any{}
	default:
		slice2 = []any{v}
	}

	// Combine slices
	result := make([]any, 0, len(slice1)+len(slice2))
	result = append(result, slice1...)
	result = append(result, slice2...)
	return result
}

func coalescelist(list string, data map[string]any, missingKeys *[]string) []any {
	_list := get(list, data, missingKeys)
	if _list == nil {
		return []any{}
	}
	if slice, ok := _list.([]any); ok {
		return slice
	}
	return []any{}
}

func addkey(toObj string, key string, valueKey string, data map[string]any, missingKeys *[]string) map[string]any {
	_obj := get(toObj, data, missingKeys)
	if _obj == nil {
		*missingKeys = append(*missingKeys, toObj)
		return nil
	}
	
	// Try to convert to map[string]any
	obj, ok := _obj.(map[string]any)
	if !ok {
		log.Printf("Error casting to map[string]any in addkey: %T %v", _obj, _obj)
		return nil
	}

	value := get(valueKey, data, missingKeys)

	result, err := Set(obj, key, value)
	if err != nil {
		log.Printf("Error setting key %v to valueKey %v in addkey: %v", key, valueKey, err)
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

	list, ok := _list.([]any)
	if !ok {
		log.Printf("Error casting to list in mapToDict: %T %v", _list, _list)
		return []map[string]any{}
	}

	result := make([]map[string]any, 0, len(list))
	for _, item := range list {
		result = append(result, map[string]any{dictKey: item})
	}

	return result
}

func addkeytoall(key string, newKey string, value any, data map[string]any, missingKeys *[]string) []any {
	val := get(key, data, missingKeys)
	if arr, ok := val.([]any); ok {
		result := make([]any, len(arr))
		for i, item := range arr {
			if mapVal, ok := item.(map[string]any); ok {
				newDict := make(map[string]any)
				for k, v := range mapVal {
					newDict[k] = v
				}
				newDict[newKey] = value
				result[i] = newDict
			} else {
				result[i] = item
			}
		}
		return result
	}
	return []any{}
}

func incrementCounter(key string, data map[string]any, missingKeys *[]string) int {
	return incrementCounterBy(key, 1, data, missingKeys)
}

func incrementCounterBy(key string, amount int, data map[string]any, missingKeys *[]string) int {
	val := get(key, data, missingKeys)
	var current int
	
	if val == nil {
		current = 0
	} else if num, err := toFloat64(val); err == nil {
		current = int(num)
	} else {
		current = 0
	}
	
	newValue := current + amount
	data[key] = newValue  // Store the new value back in data
	return newValue
}

func coalesce(key any, fallbackValue any, data map[string]any, missingKeys *[]string) any {
	// Handle the first argument (key) - try to look it up in data if it's a string
	if strKey, ok := key.(string); ok {
		// Remove optional suffix if present
		if strings.HasSuffix(strKey, "?") {
			strKey = strKey[:len(strKey)-1]
		}
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
		// Try to parse as number first
		if num, err := strconv.ParseInt(strFallback, 10, 64); err == nil {
			return num
		}
		if num, err := strconv.ParseFloat(strFallback, 64); err == nil {
			return num
		}
	}
	
	return fallbackValue
}

func filter(key string, field string, value any, data map[string]any, missingKeys *[]string) []any {
	val := get(key, data, missingKeys)
	if arr, ok := val.([]any); ok {
		result := make([]any, 0)
		for _, item := range arr {
			var fieldVal any
			if mapVal, ok := item.(map[string]any); ok {
				fieldVal = mapVal[field]
			}
			
			if reflect.DeepEqual(fieldVal, value) {
				result = append(result, item)
			}
		}
		return result
	}
	return []any{}
}

// Helper function to convert values to float64
func toFloat64(v any) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, nil
		}
		return 0, fmt.Errorf("cannot convert string %q to float64", val)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// Set creates a new dict with the specified key set to value
func Set(data map[string]any, key string, value any) (map[string]any, error) {
	if data == nil {
		return nil, fmt.Errorf("data is nil")
	}

	parts := strings.FieldsFunc(key, func(r rune) bool {
		return r == '.' || r == '[' || r == ']'
	})

	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid key: %s", key)
	}

	// Deep copy data to avoid modifying the original
	jsonData, err := ToJSON(data)
	if err != nil {
		return nil, fmt.Errorf("error copying data: %w", err)
	}
	copiedData, err := JSONToDict(jsonData)
	if err != nil {
		return nil, fmt.Errorf("error copying data: %w", err)
	}

	result, err := setRecursive(copiedData, parts, value)
	if err != nil {
		return nil, err
	}

	return result.(map[string]any), nil
}

func setRecursive(current any, parts []string, value any) (any, error) {
	if len(parts) == 0 {
		return value, nil
	}


	switch m := current.(type) {
	case map[string]any:
		if len(parts) == 1 {
			m[parts[0]] = value
			return m, nil
		}

		var next any
		var exists bool

		// Check if the next part exists
		if next, exists = m[parts[0]]; !exists {
			// Create missing nested objects as we go
			next = map[string]any{}
			m[parts[0]] = next
		}

		updated, err := setRecursive(next, parts[1:], value)
		if err != nil {
			return nil, err
		}
		m[parts[0]] = updated
		return m, nil

	case []any:
		index, err := strconv.Atoi(parts[0])
		if err != nil || index < 0 || index >= len(m) {
			return nil, fmt.Errorf("invalid array index: %s", parts[0])
		}
		if len(parts) == 1 {
			m[index] = value
			return m, nil
		}

		// Get the item at the index
		item := m[index]

		// If the item doesn't exist or isn't a Dict/map, create a new one
		if item == nil {
			item = map[string]any{}
			m[index] = item
		}

		updated, err := setRecursive(item, parts[1:], value)
		if err != nil {
			return nil, err
		}
		m[index] = updated
		return m, nil

	default:
		return nil, fmt.Errorf("cannot set key %s on type %T", parts[0], current)
	}
}

// sliceEnd gets the last n items from a slice
func sliceEnd(sliceKey string, n int, data map[string]any, missingKeys *[]string) []any {
	_slice := get(sliceKey, data, missingKeys)
	if _slice == nil {
		return nil
	}

	slice, ok := _slice.([]any)
	if !ok {
		*missingKeys = append(*missingKeys, sliceKey)
		return nil
	}

	if len(slice) == 0 {
		return slice
	}

	// If we want all items or more than available, return everything
	if n >= len(slice) {
		return slice
	}

	// Get the last n items
	return slice[len(slice)-n:]
}

// sliceEndKeepFirstUserMessage gets the last n messages from a slice, 
// but ensures the first user message is included if it would be cut off
func sliceEndKeepFirstUserMessage(sliceKey string, n int, data map[string]any, missingKeys *[]string) []any {
	_slice := get(sliceKey, data, missingKeys)
	if _slice == nil {
		return nil
	}

	slice, ok := _slice.([]any)
	if !ok {
		*missingKeys = append(*missingKeys, sliceKey)
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
func isUserMessage(msg any) bool {
	if msgMap, ok := msg.(map[string]any); ok {
		if role, exists := msgMap["role"]; exists {
			if roleStr, ok := role.(string); ok && roleStr == "user" {
				return true
			}
		}
	}
	return false
}