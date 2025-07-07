package template

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"

	json "github.com/goccy/go-json"
)

// Common types that were previously in backend/types
type Dict map[string]any

// JSON marshals any value to JSON
func ToJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// JSONToDict converts JSON string to Dict
func JSONToDict(jsonStr string) (Dict, error) {
	var result Dict
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}

// Template processing types and errors
type Key struct {
	Key        string
	IsOptional bool
}

type KeyDefinitions map[string]Key

type InfoNeededError struct {
	MissingKeys   []string
	AvailableKeys []string
	Err           error
}

func (e *InfoNeededError) Error() string {
	return fmt.Sprintf("info needed for keys %v: %v. Available keys: %v", e.MissingKeys, e.Err, e.AvailableKeys)
}

func (e *InfoNeededError) Unwrap() error {
	return e.Err
}

// Template regex patterns
var varRegexStr = `(?:{{|%\()\s*([^\s.$][^\s]*?(?:\.[^\s]+?)*)\s*(?:}}|\)s)`
var funcRegexStr = `{{([^}]+)}}`
var funcRegex = regexp.MustCompile(funcRegexStr)
var wholeVarRegexStr = fmt.Sprintf("^%s$", varRegexStr)
var wholeFuncRegexStr = fmt.Sprintf("^%s$", funcRegexStr)
var directVarRegex = regexp.MustCompile(varRegexStr)
var wholeVarRegex = regexp.MustCompile(wholeVarRegexStr)
var wholeFuncRegex = regexp.MustCompile(wholeFuncRegexStr)
var optionalVarRegex = regexp.MustCompile(`{{([^{}]+?)\?}}`)

// Helper functions for template processing
func hasDataPrefix(str string) bool {
	return strings.HasPrefix(str, ".Data") || strings.HasPrefix(str, "$.Data")
}

func removeDataPrefix(str string) string {
	if strings.HasPrefix(str, ".Data.") {
		return str[6:] // Remove ".Data."
	}
	if strings.HasPrefix(str, "$.Data.") {
		return str[7:] // Remove "$.Data."
	}
	if str == ".Data" || str == "$.Data" {
		return ""
	}
	return str
}

func containsDataSuffix(str string) bool {
	return strings.Contains(str, ".Data") || strings.Contains(str, "$.Data")
}

func containsMissingKeysSuffix(str string) bool {
	return strings.Contains(str, ".MissingKeys") || strings.Contains(str, "$.MissingKeys")
}

func appendDataParams(funcCall string) string {
	if !containsDataSuffix(funcCall) && !containsMissingKeysSuffix(funcCall) {
		// Parse the function call to insert parameters in the right place
		parts := strings.Fields(funcCall)
		if len(parts) > 0 {
			funcName := parts[0]
			args := parts[1:]
			// Reconstruct with data params after the function name and before other args
			result := funcName
			for _, arg := range args {
				result += " " + arg
			}
			result += " $.Data $.MissingKeys"
			return result
		}
		return funcCall + " $.Data $.MissingKeys"
	}
	return funcCall
}

// ParseTemplateKey parses a template key and returns Key struct
func ParseTemplateKey(key string) Key {
	isOptional := false
	cleanedKey := key

	// Check for optional syntax {{key?}}
	if strings.HasSuffix(key, "?") {
		isOptional = true
		cleanedKey = key[:len(key)-1]
	}

	// Remove .Data prefix if present
	cleanedKey = removeDataPrefix(cleanedKey)

	return Key{
		Key:        cleanedKey,
		IsOptional: isOptional,
	}
}

// Combined function map for templates
var funcMap = template.FuncMap{}
var funcsThatNeedData = []string{}

func init() {
	// Initialize the combined funcMap
	for name, fn := range basicFuncMap {
		funcMap[name] = fn
	}
	for name, fn := range dataFuncMap {
		funcMap[name] = fn
		funcsThatNeedData = append(funcsThatNeedData, name)
	}
}

// Hydrate processes a value with template substitution
func Hydrate(value any, stateParameters *Dict, parameterHydrationBehaviour *Dict) (any, error) {
	if stateParameters == nil {
		return value, nil
	}

	data, err := getData(stateParameters)
	if err != nil {
		return nil, fmt.Errorf("error getting data: %w", err)
	}

	switch v := value.(type) {
	case string:
		// Check if this is a simple variable reference (entire string is just {{var}} or {{var?}})
		if matches := wholeVarRegex.FindStringSubmatch(v); len(matches) > 0 {
			key := ParseTemplateKey(matches[1])
			val := Get(key.Key, *data, nil)
			if val == nil {
				if key.IsOptional {
					return nil, nil
				}
				// For required variables, continue with normal processing to get proper error
			} else {
				// For simple variable substitution, preserve the original type
				return val, nil
			}
		}
		return hydrateString(v, data)
	case Dict:
		return hydrateDict(v, data, parameterHydrationBehaviour)
	case []any:
		return hydrateSlice(v, data, parameterHydrationBehaviour)
	case []Dict:
		result := make([]Dict, len(v))
		for i, dict := range v {
			// Each dict in the slice should be hydrated with the same parameterHydrationBehaviour
			hydrated, err := hydrateDict(dict, data, parameterHydrationBehaviour)
			if err != nil {
				return nil, fmt.Errorf("error hydrating slice item %d: %w", i, err)
			}
			result[i] = hydrated
		}
		return result, nil
	default:
		return value, nil
	}
}

func getData(stateParameters *Dict) (*Dict, error) {
	if stateParameters == nil {
		return nil, nil
	}

	jsonStr, err := ToJSON(*stateParameters)
	if err != nil {
		return nil, fmt.Errorf("error marshaling data: %w", err)
	}

	data, err := JSONToDict(jsonStr)
	if err != nil {
		return nil, fmt.Errorf("error cloning data: %w", err)
	}

	// Preserve original types from the source data
	// JSON unmarshaling converts numbers to float64, so we need to restore original types
	for key, value := range *stateParameters {
		if jsonValue, exists := data[key]; exists {
			// Check if we need to preserve the original type
			if shouldPreserveType(value, jsonValue) {
				data[key] = value
			}
		} else {
			// Copy non-JSONable values
			data[key] = value
		}
	}

	return &data, nil
}

// shouldPreserveType checks if we should preserve the original type over the JSON-unmarshaled type
func shouldPreserveType(original, jsonValue any) bool {
	// Preserve integer types that were converted to float64
	switch original.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		if _, isFloat := jsonValue.(float64); isFloat {
			return true
		}
	}
	
	// For nested structures, we need to preserve types recursively
	switch orig := original.(type) {
	case Dict:
		if jsonDict, ok := jsonValue.(Dict); ok {
			preserveNestedTypes(orig, jsonDict)
		}
	case map[string]any:
		if jsonMap, ok := jsonValue.(map[string]any); ok {
			preserveNestedTypes(orig, jsonMap)
		}
	}
	
	return false
}

// preserveNestedTypes recursively preserves types in nested structures
func preserveNestedTypes(original, jsonValue map[string]any) {
	for key, origVal := range original {
		if jsonVal, exists := jsonValue[key]; exists {
			if shouldPreserveType(origVal, jsonVal) {
				jsonValue[key] = origVal
			}
		}
	}
}

// HydrateString hydrates a string template with data
func HydrateString(templateStr string, data *Dict, behaviour *Dict) (string, error) {
	result, err := Hydrate(templateStr, data, behaviour)
	if err != nil {
		return "", err
	}
	if str, ok := result.(string); ok {
		return str, nil
	}
	return fmt.Sprintf("%v", result), nil
}

// HydrateDict hydrates a Dict with template data
func HydrateDict(dict Dict, data *Dict, behaviour *Dict) (Dict, error) {
	result, err := Hydrate(dict, data, behaviour)
	if err != nil {
		return nil, err
	}
	if resultDict, ok := result.(Dict); ok {
		return resultDict, nil
	}
	return nil, fmt.Errorf("result is not a Dict")
}

// HydrateSlice hydrates a slice with template data
func HydrateSlice(slice []any, data *Dict, behaviour *Dict) ([]any, error) {
	result, err := Hydrate(slice, data, behaviour)
	if err != nil {
		return nil, err
	}
	if resultSlice, ok := result.([]any); ok {
		return resultSlice, nil
	}
	return nil, fmt.Errorf("result is not a slice")
}

// MergeSources merges multiple data sources into a single Dict
func MergeSources(sources ...Dict) (Dict, error) {
	result := make(Dict)
	for _, source := range sources {
		for k, v := range source {
			result[k] = v
		}
	}
	return result, nil
}

