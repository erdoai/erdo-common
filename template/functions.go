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
	"getOptional":                  getOptional,
	"concat":                       concat,
	"getOrOriginal":                getOrOriginal,
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

func truthy(key string, data Dict) bool {
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

func _len(v any) int {
	if v == nil {
		return 0
	}
	
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
		return val.Len()
	default:
		return 0
	}
}

func add(a, b any) (float64, error) {
	aVal, err := toFloat64(a)
	if err != nil {
		return 0, err
	}
	bVal, err := toFloat64(b)
	if err != nil {
		return 0, err
	}
	return aVal + bVal, nil
}

func sub(a, b any) (float64, error) {
	aVal, err := toFloat64(a)
	if err != nil {
		return 0, err
	}
	bVal, err := toFloat64(b)
	if err != nil {
		return 0, err
	}
	return aVal - bVal, nil
}

func gt(a, b any) (bool, error) {
	aVal, err := toFloat64(a)
	if err != nil {
		return false, err
	}
	bVal, err := toFloat64(b)
	if err != nil {
		return false, err
	}
	return aVal > bVal, nil
}

func lt(a, b any) (bool, error) {
	aVal, err := toFloat64(a)
	if err != nil {
		return false, err
	}
	bVal, err := toFloat64(b)
	if err != nil {
		return false, err
	}
	return aVal < bVal, nil
}

func mergeRaw(items ...any) any {
	// Check if all items are slices
	allSlices := true
	for _, item := range items {
		if _, ok := item.([]any); !ok {
			allSlices = false
			break
		}
	}
	
	// If all items are slices, concatenate them
	if allSlices && len(items) > 0 {
		var result []any
		for _, item := range items {
			if slice, ok := item.([]any); ok {
				result = append(result, slice...)
			}
		}
		return result
	}
	
	// Otherwise, merge as dictionaries
	result := make(Dict)
	for _, m := range items {
		if dict, ok := m.(Dict); ok {
			for k, v := range dict {
				result[k] = v
			}
		} else if mapVal, ok := m.(map[string]any); ok {
			for k, v := range mapVal {
				result[k] = v
			}
		}
	}
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
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

func regexReplace(pattern, replacement, text string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Printf("Error compiling regex pattern %s: %v", pattern, err)
		return text
	}
	return re.ReplaceAllString(text, replacement)
}

func noop(v ...any) any {
	if len(v) > 0 {
		return v[0]
	}
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
func Get(key string, data Dict, missingKeys *[]string) any {
	return get(key, data, missingKeys)
}

func get(key string, data Dict, missingKeys *[]string) any {
	keys := strings.Split(key, ".")
	current := any(data)
	
	for _, k := range keys {
		switch v := current.(type) {
		case Dict:
			if val, exists := v[k]; exists {
				current = val
			} else {
				if missingKeys != nil {
					*missingKeys = append(*missingKeys, key)
				}
				return nil
			}
		case map[string]any:
			if val, exists := v[k]; exists {
				current = val
			} else {
				if missingKeys != nil {
					*missingKeys = append(*missingKeys, key)
				}
				return nil
			}
		case []any:
			if idx, err := strconv.Atoi(k); err == nil && idx >= 0 && idx < len(v) {
				current = v[idx]
			} else {
				if missingKeys != nil {
					*missingKeys = append(*missingKeys, key)
				}
				return nil
			}
		default:
			if missingKeys != nil {
				*missingKeys = append(*missingKeys, key)
			}
			return nil
		}
	}
	
	return current
}

// getOptional retrieves an optional value from data using dot notation key
// Returns empty string if the key is missing (no error)
func getOptional(key string, data Dict, missingKeys *[]string) any {
	keys := strings.Split(key, ".")
	current := any(data)
	
	for _, k := range keys {
		switch v := current.(type) {
		case Dict:
			if val, exists := v[k]; exists {
				current = val
			} else {
				// For optional variables, don't add to missing keys
				return ""
			}
		case map[string]any:
			if val, exists := v[k]; exists {
				current = val
			} else {
				// For optional variables, don't add to missing keys
				return ""
			}
		case []any:
			if idx, err := strconv.Atoi(k); err == nil && idx >= 0 && idx < len(v) {
				current = v[idx]
			} else {
				// For optional variables, don't add to missing keys
				return ""
			}
		default:
			// For optional variables, don't add to missing keys
			return ""
		}
	}
	
	return current
}

func concat(key1, key2 string, data Dict, missingKeys *[]string) string {
	val1 := get(key1, data, missingKeys)
	val2 := get(key2, data, missingKeys)
	return fmt.Sprintf("%v%v", val1, val2)
}

func getOrOriginal(key string, data Dict, missingKeys *[]string) any {
	val := get(key, data, missingKeys)
	if val != nil {
		return val
	}
	return key
}

func slice(key string, start, end int, data Dict, missingKeys *[]string) []any {
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

func extractSlice(key string, field string, data Dict, missingKeys *[]string) []any {
	val := get(key, data, missingKeys)
	if arr, ok := val.([]any); ok {
		result := make([]any, 0, len(arr))
		for _, item := range arr {
			if dict, ok := item.(Dict); ok {
				if fieldVal, exists := dict[field]; exists {
					result = append(result, fieldVal)
				}
			} else if mapVal, ok := item.(map[string]any); ok {
				if fieldVal, exists := mapVal[field]; exists {
					result = append(result, fieldVal)
				}
			}
		}
		return result
	}
	return []any{}
}

func dedupeBy(key string, field string, data Dict, missingKeys *[]string) []any {
	val := get(key, data, missingKeys)
	if arr, ok := val.([]any); ok {
		seen := make(map[string]bool)
		result := make([]any, 0)
		
		for _, item := range arr {
			var fieldVal any
			if dict, ok := item.(Dict); ok {
				// Use get to handle nested fields like "metadata.id"
				fieldVal = get(field, dict, &[]string{})
			} else if mapVal, ok := item.(map[string]any); ok {
				// Convert to Dict and use get
				tempDict := make(Dict)
				for k, v := range mapVal {
					tempDict[k] = v
				}
				fieldVal = get(field, tempDict, &[]string{})
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

func find(key string, field string, value any, data Dict, missingKeys *[]string) any {
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
			if dict, ok := item.(Dict); ok {
				fieldVal = dict[field]
			} else if mapVal, ok := item.(map[string]any); ok {
				fieldVal = mapVal[field]
			}
			
			if reflect.DeepEqual(fieldVal, value) {
				return item
			}
		}
	}
	return nil
}

func findByValue(arrayKey string, fieldKey string, targetValue any, data Dict, missingKeys *[]string) any {
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
		itemDict, ok := item.(Dict)
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

func getAtIndex(key string, index int, data Dict, missingKeys *[]string) any {
	val := get(key, data, missingKeys)
	if arr, ok := val.([]any); ok {
		if index >= 0 && index < len(arr) {
			return arr[index]
		}
	}
	return nil
}

func merge(target string, source string, data Dict, missingKeys *[]string) Dict {
	targetVal := get(target, data, missingKeys)
	sourceVal := get(source, data, missingKeys)
	
	result := make(Dict)
	
	if targetDict, ok := targetVal.(Dict); ok {
		for k, v := range targetDict {
			result[k] = v
		}
	} else if targetMap, ok := targetVal.(map[string]any); ok {
		for k, v := range targetMap {
			result[k] = v
		}
	}
	
	if sourceDict, ok := sourceVal.(Dict); ok {
		for k, v := range sourceDict {
			result[k] = v
		}
	} else if sourceMap, ok := sourceVal.(map[string]any); ok {
		for k, v := range sourceMap {
			result[k] = v
		}
	}
	
	return result
}

func coalescelist(keys []string, data Dict, missingKeys *[]string) any {
	for _, key := range keys {
		if val := get(key, data, missingKeys); val != nil {
			return val
		}
	}
	return nil
}

func addkey(toObj string, key string, valueKey string, data Dict, missingKeys *[]string) Dict {
	_obj := get(toObj, data, missingKeys)
	obj, ok := _obj.(Dict)
	if !ok {
		log.Printf("Error casting to dict in addkey: %T %v", _obj, _obj)
		return nil
	}

	value := get(valueKey, data, missingKeys)

	result := make(Dict)
	for k, v := range obj {
		result[k] = v
	}
	result[key] = value

	return result
}

func removekey(toObj string, key string, data Dict, missingKeys *[]string) Dict {
	_obj := get(toObj, data, missingKeys)
	obj, ok := _obj.(Dict)
	if !ok {
		log.Printf("Error casting to dict in removekey: %T %v", _obj, _obj)
		return obj
	}

	result := make(Dict)
	for k, v := range obj {
		if k != key {
			result[k] = v
		}
	}

	return result
}

func mapToDict(listKey string, dictKey string, data Dict, missingKeys *[]string) []Dict {
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
		return []Dict{}
	}

	list, ok := _list.([]any)
	if !ok {
		log.Printf("Error casting to list in mapToDict: %T %v", _list, _list)
		return []Dict{}
	}

	result := make([]Dict, 0, len(list))
	for _, item := range list {
		result = append(result, Dict{dictKey: item})
	}

	return result
}

func addkeytoall(key string, newKey string, value any, data Dict, missingKeys *[]string) []any {
	val := get(key, data, missingKeys)
	if arr, ok := val.([]any); ok {
		result := make([]any, len(arr))
		for i, item := range arr {
			if dict, ok := item.(Dict); ok {
				newDict := make(Dict)
				for k, v := range dict {
					newDict[k] = v
				}
				newDict[newKey] = value
				result[i] = newDict
			} else if mapVal, ok := item.(map[string]any); ok {
				newDict := make(Dict)
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

func incrementCounter(key string, data Dict, missingKeys *[]string) int {
	return incrementCounterBy(key, 1, data, missingKeys)
}

func incrementCounterBy(key string, amount int, data Dict, missingKeys *[]string) int {
	val := get(key, data, missingKeys)
	if val == nil {
		return amount
	}
	
	if num, err := toFloat64(val); err == nil {
		return int(num) + amount
	}
	
	return amount
}

func coalesce(key any, fallbackValue any, data Dict, missingKeys *[]string) any {
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

func filter(key string, field string, value any, data Dict, missingKeys *[]string) []any {
	val := get(key, data, missingKeys)
	if arr, ok := val.([]any); ok {
		result := make([]any, 0)
		for _, item := range arr {
			var fieldVal any
			if dict, ok := item.(Dict); ok {
				fieldVal = dict[field]
			} else if mapVal, ok := item.(map[string]any); ok {
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
func Set(data Dict, key string, value any) (Dict, error) {
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

	return result.(Dict), nil
}

func setRecursive(current any, parts []string, value any) (any, error) {
	if len(parts) == 0 {
		return value, nil
	}

	// Convert map[string]any to Dict for consistent handling
	if mapValue, ok := current.(map[string]any); ok {
		current = Dict(mapValue)
	}

	switch m := current.(type) {
	case Dict:
		if len(parts) == 1 {
			m[parts[0]] = value
			return m, nil
		}

		var next any
		var exists bool

		// Check if the next part exists
		if next, exists = m[parts[0]]; !exists {
			// Create missing nested objects as we go
			next = Dict{}
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
			item = Dict{}
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

// sliceEndKeepFirstUserMessage gets the last n messages from a slice, 
// but ensures the first user message is included if it would be cut off
func sliceEndKeepFirstUserMessage(sliceKey string, n int, data Dict, missingKeys *[]string) []any {
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
	if msgDict, ok := msg.(Dict); ok {
		if role, exists := msgDict["role"]; exists {
			if roleStr, ok := role.(string); ok && roleStr == "user" {
				return true
			}
		}
	} else if msgMap, ok := msg.(map[string]any); ok {
		if role, exists := msgMap["role"]; exists {
			if roleStr, ok := role.(string); ok && roleStr == "user" {
				return true
			}
		}
	}
	return false
}