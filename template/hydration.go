package template

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/template"

	common "github.com/erdoai/erdo-common/types"
	utils "github.com/erdoai/erdo-common/utils"
)

// Hydrate parameters
type Key struct {
	Key        string
	IsOptional bool
}

type KeyDefinitions map[string]Key

type MissingKeyInfo struct {
	Key  string
	Path string
}

type InfoNeededError struct {
	MissingKeys     []string
	MissingKeyPaths []MissingKeyInfo
	AvailableKeys   []string
	Err             error
}

func (e *InfoNeededError) Error() string {
	if len(e.MissingKeyPaths) > 0 {
		pathDetails := make([]string, len(e.MissingKeyPaths))
		for i, info := range e.MissingKeyPaths {
			if info.Path != "" {
				// Build the full path: path.key
				if strings.HasPrefix(info.Path, "[") {
					// For array indices, don't add a dot
					pathDetails[i] = info.Path + "." + info.Key
				} else {
					pathDetails[i] = info.Path + "." + info.Key
				}
			} else {
				pathDetails[i] = info.Key
			}
		}
		return fmt.Sprintf("info needed for keys %v: %v. Available keys: %v", pathDetails, e.Err, e.AvailableKeys)
	}
	return fmt.Sprintf("info needed for keys %v: %v. Available keys: %v", e.MissingKeys, e.Err, e.AvailableKeys)
}

func (e *InfoNeededError) Unwrap() error {
	return e.Err
}

// Helper to add path to error - prepends the path segment
func addPathToError(err error, pathSegment string) {
	if err == nil || pathSegment == "" {
		return
	}

	var infoErr *InfoNeededError
	if errors.As(err, &infoErr) {
		// Create MissingKeyPaths if needed
		if len(infoErr.MissingKeyPaths) == 0 && len(infoErr.MissingKeys) > 0 {
			infoErr.MissingKeyPaths = make([]MissingKeyInfo, len(infoErr.MissingKeys))
			for i, key := range infoErr.MissingKeys {
				infoErr.MissingKeyPaths[i] = MissingKeyInfo{Key: key, Path: ""}
			}
		}

		// Prepend the path segment to existing paths
		for i := range infoErr.MissingKeyPaths {
			if infoErr.MissingKeyPaths[i].Path == "" {
				infoErr.MissingKeyPaths[i].Path = pathSegment
			} else {
				// For arrays, don't add a dot
				if strings.HasPrefix(infoErr.MissingKeyPaths[i].Path, "[") {
					infoErr.MissingKeyPaths[i].Path = pathSegment + infoErr.MissingKeyPaths[i].Path
				} else {
					infoErr.MissingKeyPaths[i].Path = pathSegment + "." + infoErr.MissingKeyPaths[i].Path
				}
			}
		}
	}
}

func getData(stateParameters *map[string]any) (*map[string]any, error) {
	if stateParameters == nil {
		return nil, nil
	}

	// OPTIMIZATION: Don't clone or normalize the data - just return it as-is
	// The hydration functions don't mutate the input data, so cloning is unnecessary
	// The caller MUST pass in JSON-normalized data (map[string]any with lowercase keys)
	// This avoids the expensive JSON marshal/unmarshal that was taking ~180ms for large state
	return stateParameters, nil
}

func Hydrate(value any, stateParameters *map[string]any, parameterHydrationBehaviour *map[string]any) (any, error) {
	if stateParameters == nil {
		return value, nil
	}

	data, err := getData(stateParameters)
	if err != nil {
		return nil, fmt.Errorf("error getting data: %w", err)
	}

	switch v := value.(type) {
	case string:
		if parameterHydrationBehaviour != nil {
			panic(fmt.Sprintf("hydrating string with behaviour %+v", parameterHydrationBehaviour))
		}
		return hydrateString(v, data)
	case map[string]any:
		return hydrateDict(v, data, parameterHydrationBehaviour)
	case []any:
		return hydrateSlice(v, data, parameterHydrationBehaviour)
	case []map[string]any:
		// Convert []map[string]any to []any to reuse existing hydrateSlice logic
		anySlice := make([]any, len(v))
		for i, d := range v {
			anySlice[i] = d
		}
		res, err := hydrateSlice(anySlice, data, parameterHydrationBehaviour)
		if err != nil {
			return nil, err
		}
		// Convert back to []map[string]any
		dictSlice := make([]map[string]any, len(res))
		for i, item := range res {
			if dictItem, ok := item.(map[string]any); ok {
				dictSlice[i] = dictItem
			} else {
				return nil, fmt.Errorf("expected map[string]any, got %T", item)
			}
		}
		return dictSlice, nil
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128:
		return v, nil
	default:
		// Handle any other slice type using reflection as fallback
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice {
			// Convert any slice type to []any
			anySlice := make([]any, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				anySlice[i] = rv.Index(i).Interface()
			}
			return hydrateSlice(anySlice, data, parameterHydrationBehaviour)
		}
		// For non-slice types, just return as-is
		log.Printf("!! unable to hydrate unknown type %T", value)
		return value, nil
	}
}

// shouldHydrateField checks if a specific field should be hydrated based on the parameterHydrationBehaviour
func shouldHydrateField(key string, parameterHydrationBehaviour *map[string]any) (bool, *map[string]any) {
	if parameterHydrationBehaviour == nil {
		return true, nil // Default behavior is to hydrate
	}

	var childBehaviour *map[string]any

	// Check if this specific key has hydration behavior defined
	if nestedBehaviour, ok := (*parameterHydrationBehaviour)[key]; ok {
		// Direct raw behavior assignment
		if raw, ok := nestedBehaviour.(common.ParameterHydrationBehaviour); ok {
			if raw == common.ParameterHydrationBehaviourRaw {
				return false, nil // Don't hydrate if marked as raw
			}
		} else if behaviour, ok := nestedBehaviour.(map[string]any); ok {
			// The nested object has its own behavior configuration
			childBehaviour = &behaviour
		}
	}

	return true, childBehaviour // Default to hydrating if not explicitly marked as raw
}

var varRegexStr = `(?:{{|%\()\s*([^\s.$][^\s]*?(?:\.[^\s]+?)*)\s*(?:}}|\)s)`
var funcRegexStr = `{{([^}]+)}}`
var funcRegex = regexp.MustCompile(funcRegexStr)
var wholeVarRegexStr = fmt.Sprintf("^%s$", varRegexStr)
var wholeFuncRegexStr = fmt.Sprintf("^%s$", funcRegexStr)
var directVarRegex = regexp.MustCompile(varRegexStr)
var wholeVarRegex = regexp.MustCompile(wholeVarRegexStr)
var wholeFuncRegex = regexp.MustCompile(wholeFuncRegexStr)
var optionalVarRegex = regexp.MustCompile(`{{([^{}]+?)\?}}`)

// hasDataPrefix checks if a string starts with .Data. or $.Data.
func hasDataPrefix(str string) bool {
	return strings.HasPrefix(str, ".Data.") || strings.HasPrefix(str, "$.Data.")
}

// removeDataPrefix removes .Data. or $.Data. prefix if present
func removeDataPrefix(str string) string {
	if strings.HasPrefix(str, ".Data.") {
		return strings.TrimPrefix(str, ".Data.")
	}
	if strings.HasPrefix(str, "$.Data.") {
		return strings.TrimPrefix(str, "$.Data.")
	}
	return str
}

// containsDataSuffix checks if a string already has .Data and .MissingKeys parameters at the end
func containsDataSuffix(str string) bool {
	// We need to check if the function already has both .Data and .MissingKeys at the END
	// not inside nested function calls
	trimmed := strings.TrimSpace(str)

	// Check if it ends with the pattern: ... .Data .MissingKeys or ... $.Data $.MissingKeys
	// This regex will match if the string ends with either pattern (with optional whitespace)
	pattern1 := `\.Data\s+\.MissingKeys\s*$`
	pattern2 := `\$\.Data\s+\$\.MissingKeys\s*$`
	matched1, _ := regexp.MatchString(pattern1, trimmed)
	matched2, _ := regexp.MatchString(pattern2, trimmed)
	return matched1 || matched2
}

// containsMissingKeysSuffix checks if a string contains .MissingKeys or $.MissingKeys suffix
func containsMissingKeysSuffix(str string) bool {
	return strings.Contains(str, ".MissingKeys") || strings.Contains(str, "$.MissingKeys")
}

// ParseTemplateKey creates a Key struct from a string, handling optional parameters
// (those with a ? suffix) and special prefixes like ".Data." or "$.Data.".
func ParseTemplateKey(key string) Key {
	// Check if parameter is optional
	isOptional := strings.HasSuffix(key, "?")
	cleanKey := strings.TrimSuffix(key, "?")

	// Remove .Data. prefix if present
	cleanKey = removeDataPrefix(cleanKey)

	return Key{
		Key:        cleanKey,
		IsOptional: isOptional,
	}
}

// findTemplateKeysToHydrate extracts template keys from a string using a regex
// and creates Key objects representing each extracted key with its optionality.
// The regex should extract the exact content (including the optional ? suffix) from template variables.
func findTemplateKeysToHydrate(s any, regex *regexp.Regexp, parameterHydrationBehaviour *map[string]any) []Key {
	// Convert input to string if possible
	var str string
	switch v := s.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		// For non-string types, return empty result
		return []Key{}
	}

	// Early exit: if string doesn't contain template markers, return empty
	// This is a huge optimization when scanning hydrated output with many non-template strings
	if !strings.Contains(str, "{{") && !strings.Contains(str, "%(") {
		return []Key{}
	}

	matches := regex.FindAllStringSubmatch(str, -1)
	keys := make([]Key, 0, len(matches))
	for _, match := range matches {
		if len(match) != 2 {
			continue // Skip if we didn't get the expected capture group
		}

		// match[0] is the entire match including braces: "{{key?}}"
		// match[1] is the extracted content inside the braces: "key?"
		keyStr := strings.TrimSpace(match[1])

		// Check for optional marker (? suffix)
		isOptional := strings.HasSuffix(keyStr, "?")
		cleanKey := strings.TrimSuffix(keyStr, "?")

		// Remove .Data. prefix if present
		cleanKey = removeDataPrefix(cleanKey)

		keys = append(keys, Key{
			Key:        cleanKey,
			IsOptional: isOptional,
		})
	}

	return keys
}

// cleanKey processes a key string to determine if it's optional and returns the cleaned key.
// Returns the key without the optional marker and a boolean indicating if it was optional.
func cleanKey(key string) (string, bool) {
	// Check if parameter is optional (has ? suffix)
	isOptional := strings.HasSuffix(key, "?")
	cleanKey := strings.TrimSuffix(key, "?")

	// Remove .Data. prefix if present
	cleanKey = removeDataPrefix(cleanKey)

	return cleanKey, isOptional
}

func FindTemplateKeysToHydrate(s any, includeOptional bool, parameterHydrationBehaviour *map[string]any) []Key {
	// For simple string inputs, use the existing function
	if str, ok := s.(string); ok {
		keys := findTemplateKeysToHydrate(str, directVarRegex, parameterHydrationBehaviour)

		res := make([]Key, 0, len(keys))
		for _, key := range keys {
			if !includeOptional && key.IsOptional {
				continue
			}
			res = append(res, key)
		}

		return res
	}

	// For complex data structures, traverse recursively
	var keys []Key
	switch v := s.(type) {
	case map[string]any:
		keys = findKeysInDict(v, includeOptional, parameterHydrationBehaviour)
	case []any:
		keys = findKeysInSlice(v, includeOptional, parameterHydrationBehaviour)
	default:
		log.Printf("!! unable to find keys in %T", v)
		return []Key{}
	}

	// Remove duplicates while preserving order
	seen := make(map[string]bool)
	uniqueKeys := make([]Key, 0, len(keys))
	for _, key := range keys {
		if !seen[key.Key] {
			seen[key.Key] = true
			uniqueKeys = append(uniqueKeys, key)
		}
	}

	return uniqueKeys
}

func findKeysInDict(dict map[string]any, includeOptional bool, parameterHydrationBehaviour *map[string]any) []Key {
	var keys []Key

	for k, v := range dict {
		// Check if this specific field should be hydrated
		shouldHydrate, childBehaviour := shouldHydrateField(k, parameterHydrationBehaviour)
		if !shouldHydrate {
			continue
		}

		// Process this value based on its type
		switch tv := v.(type) {
		case string:
			fieldKeys := findTemplateKeysToHydrate(tv, directVarRegex, childBehaviour)
			for _, key := range fieldKeys {
				if includeOptional || !key.IsOptional {
					keys = append(keys, key)
				}
			}
		case map[string]any:
			fieldKeys := findKeysInDict(tv, includeOptional, childBehaviour)
			keys = append(keys, fieldKeys...)
		case []any:
			fieldKeys := findKeysInSlice(tv, includeOptional, childBehaviour)
			keys = append(keys, fieldKeys...)
		}
	}

	return keys
}

func findKeysInSlice(slice []any, includeOptional bool, parameterHydrationBehaviour *map[string]any) []Key {
	var keys []Key

	for _, v := range slice {
		// Process this value based on its type
		switch tv := v.(type) {
		case string:
			// We pass down the same behaviour for each element in the slice
			fieldKeys := findTemplateKeysToHydrate(tv, directVarRegex, parameterHydrationBehaviour)
			for _, key := range fieldKeys {
				if includeOptional || !key.IsOptional {
					keys = append(keys, key)
				}
			}
		case map[string]any:
			fieldKeys := findKeysInDict(tv, includeOptional, parameterHydrationBehaviour)
			keys = append(keys, fieldKeys...)
		case []any:
			fieldKeys := findKeysInSlice(tv, includeOptional, parameterHydrationBehaviour)
			keys = append(keys, fieldKeys...)
		}
	}

	return keys
}

func FindTemplateKeyStringsToHydrate(s any, includeOptional bool, parameterHydrationBehaviour *map[string]any) []string {
	keys := FindTemplateKeysToHydrate(s, includeOptional, parameterHydrationBehaviour)
	return getKeyStrings(keys)
}

var reservedWords = []string{"if", "range", "with", "end", "else", "template", "block", "define"}

func parseTemplate(input string) (string, error) {
	// First, replace function calls
	res := funcRegex.ReplaceAllStringFunc(input, func(match string) string {
		// Skip if already contains .Data
		if containsDataSuffix(match) {
			return match
		}

		// Extract the function call from the match
		funcCall := strings.TrimSpace(funcRegex.FindStringSubmatch(match)[1])

		// Check if this is actually a function call (has a space) vs a variable
		if !strings.Contains(funcCall, " ") {
			// This is a variable, not a function call
			return match
		}

		firstWord := strings.Fields(funcCall)[0]

		// Skip reserved words unless they contain nested functions
		if slices.Contains(reservedWords, firstWord) {
			// Check if this reserved word statement contains nested function calls
			if strings.Contains(funcCall, "(") && strings.Contains(funcCall, ")") {
				// Process any nested function calls within reserved word statements
				processedCall := processReservedWordWithNestedFunctions(funcCall)
				return fmt.Sprintf("{{%s}}", processedCall)
			}
			return match
		}

		// Skip special characters
		firstLetter := string(funcCall[0])
		if slices.Contains([]string{"$", "-"}, firstLetter) {
			return match
		}

		// Process the function call recursively to handle nested functions
		// This will add .Data and .MissingKeys to any data functions
		processedCall := processNestedFunctionCalls(funcCall)

		return fmt.Sprintf("{{%s}}", processedCall)
	})

	// Then, replace direct variable references
	res = directVarRegex.ReplaceAllStringFunc(res, func(match string) string {
		content := strings.TrimSpace(match[2 : len(match)-2])
		firstWord := strings.Fields(content)[0]

		if slices.Contains(reservedWords, firstWord) {
			return match
		}

		firstLetter := string(content[0])
		if firstLetter == "$" {
			return match
		}

		// Check if the match is already wrapped in getOrOriginal
		if strings.HasPrefix(content, "getOrOriginal") {
			return match
		}

		// Wrap the result in nilToEmptyString to handle nil values
		return fmt.Sprintf("{{nilToEmptyString (getOrOriginal \"%v\" $.KeyDefinitions $.Data $.MissingKeys)}}", content)
	})

	// Create a template with our custom functions to test for errors
	t := template.New("test").Funcs(funcMap)

	_, err := t.Parse(res)
	if err != nil {
		log.Printf("!!invalid template syntax: %v in template: %v", err, res)
		return "", fmt.Errorf("invalid template syntax: %w in template: %v", err, res)
	}

	return res, nil
}

// processNestedFunctionCalls recursively processes function calls and ensures that
// all functions have the necessary .Data and .MissingKeys parameters
func processNestedFunctionCalls(funcCall string) string {
	// Parse the function call to get structured fields
	fields := parseQuotedFields(funcCall)
	if len(fields) == 0 {
		return funcCall
	}

	funcName := fields[0]

	// Process each field
	var processedFields []string
	processedFields = append(processedFields, funcName)

	for i := 1; i < len(fields); i++ {
		field := fields[i]

		// Check if this field is a nested function call
		if strings.HasPrefix(field, "(") && strings.HasSuffix(field, ")") {
			// Extract the nested function
			nestedFunc := field[1 : len(field)-1]
			// Process the nested function recursively
			processedNested := processNestedFunctionCalls(nestedFunc)
			// Re-wrap in parentheses
			processedFields = append(processedFields, "("+processedNested+")")
		} else {
			// Regular field, keep as-is
			processedFields = append(processedFields, field)
		}
	}

	// Reconstruct the function call
	finalResult := strings.Join(processedFields, " ")

	// Check if this function needs .Data and .MissingKeys added
	_, requiresData := dataFuncMap[funcName]
	if requiresData {
		// Check if the function already has the required parameters
		hasMissingKeys := false
		hasData := false

		// Check the last few fields for .Data and .MissingKeys
		for i := len(processedFields) - 1; i >= 0 && i >= len(processedFields)-2; i-- {
			field := processedFields[i]
			if field == ".MissingKeys" || field == "$.MissingKeys" {
				hasMissingKeys = true
			}
			if field == ".Data" || field == "$.Data" {
				hasData = true
			}
		}

		if !hasMissingKeys {
			// Special handling for get with nested function as data parameter
			if funcName == "get" && len(processedFields) >= 2 {
				// Check if the second parameter is a nested function
				hasNestedDataParam := false
				for i := 1; i < len(processedFields); i++ {
					if strings.HasPrefix(processedFields[i], "(") && strings.HasSuffix(processedFields[i], ")") {
						// Found a nested function parameter
						if i == 2 && funcName == "get" {
							// This is the second parameter to get, which provides the data
							hasNestedDataParam = true
							break
						}
					}
				}

				if hasNestedDataParam {
					// The second parameter is a nested function that provides the data
					// Only add .MissingKeys
					finalResult = finalResult + " $.MissingKeys"
				} else if !hasData {
					// Normal case - add both if not already present
					finalResult = finalResult + " $.Data $.MissingKeys"
				} else {
					// Has data but not missingKeys
					finalResult = finalResult + " $.MissingKeys"
				}
			} else if !hasData {
				// For all other data functions, add both parameters if not present
				finalResult = finalResult + " $.Data $.MissingKeys"
			} else {
				// Has data but not missingKeys
				finalResult = finalResult + " $.MissingKeys"
			}
		}
	}
	return finalResult
}

// processReservedWordWithNestedFunctions handles reserved word statements (like if, range)
// that contain nested function calls, processing only the nested function calls
func processReservedWordWithNestedFunctions(statement string) string {

	// Parse parentheses manually to handle nested function calls correctly
	result := strings.Builder{}
	i := 0

	for i < len(statement) {
		if statement[i] == '(' {
			// Find the matching closing parenthesis
			parenCount := 1
			start := i
			i++ // skip the opening parenthesis

			for i < len(statement) && parenCount > 0 {
				if statement[i] == '(' {
					parenCount++
				} else if statement[i] == ')' {
					parenCount--
				}
				i++
			}

			if parenCount == 0 {
				// Extract the content within parentheses
				inner := strings.TrimSpace(statement[start+1 : i-1])

				// Check if this looks like a function call
				if strings.Contains(inner, " ") {
					parts := strings.Fields(inner)
					if len(parts) > 0 {
						funcName := parts[0]
						// Skip if it's a reserved word itself
						if slices.Contains(reservedWords, funcName) {
							result.WriteString(statement[start:i])
							continue
						}

						// Process the function call
						processed := processNestedFunctionCalls(inner)
						result.WriteString("(" + processed + ")")
						continue
					}
				}

				// Not a function call, keep original
				result.WriteString(statement[start:i])
			} else {
				// Malformed parentheses, keep original
				result.WriteString(statement[start:])
				break
			}
		} else {
			result.WriteByte(statement[i])
			i++
		}
	}

	return result.String()
}

func getKeys(data map[string]any) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	return keys
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

// HydrateString hydrates a string with the given state parameters
func HydrateString(s string, stateParameters *map[string]any, parameterHydrationBehaviour ...*map[string]any) (string, error) {
	var behaviour *map[string]any
	if len(parameterHydrationBehaviour) > 0 {
		behaviour = parameterHydrationBehaviour[0]
	}

	value, err := Hydrate(s, stateParameters, behaviour)
	if err != nil {
		return "", err
	}

	strValue, ok := value.(string)
	if !ok {
		// always cast to string
		return fmt.Sprintf("%v", value), nil
		// return "", fmt.Errorf("expected string, got %T", value)
	}

	return strValue, nil
}

// Adds additional processing for custom functions in Go templates.
// This ensures that variable arguments like "resource_id" get the actual value
// instead of being passed as literal strings.
func addCustomTemplateHelpers(t *template.Template, data map[string]any) *template.Template {
	customFuncMap := template.FuncMap{
		// Create a wrapper around addkeytoall that handles variable references
		"addkeytoall": func(listKey string, key string, valueArg any) any {
			// Handle when valueArg is a string that should be a variable name
			if valueStr, ok := valueArg.(string); ok && !strings.HasPrefix(valueStr, "\"") && !strings.HasPrefix(valueStr, "'") {
				// Check if this is intended to be a variable reference
				if val, exists := data[valueStr]; exists {
					// It's a variable reference, use the actual value
					return addkeytoall(listKey, key, val, data, &[]string{})
				} else if strings.Contains(valueStr, ".") {
					// It might be a nested variable reference like "foo.bar"
					parts := strings.Split(valueStr, ".")
					var current any = data
					var found bool = true

					// Navigate through the nested parts
					for _, part := range parts {
						if dictData, ok := current.(map[string]any); ok {
							if val, exists := dictData[part]; exists {
								current = val
							} else {
								found = false
								break
							}
						} else if mapData, ok := current.(map[string]any); ok {
							if val, exists := mapData[part]; exists {
								current = val
							} else {
								found = false
								break
							}
						} else {
							found = false
							break
						}
					}

					if found {
						// We found the nested value, use it
						return addkeytoall(listKey, key, current, data, &[]string{})
					}
				}
			}

			// Default case: just pass the value as is
			return addkeytoall(listKey, key, valueArg, data, &[]string{})
		},
	}

	// Add our custom functions to the template
	return t.Funcs(customFuncMap)
}

func getKeyStrings(keys []Key) []string {
	res := make([]string, 0, len(keys))
	for _, key := range keys {
		res = append(res, key.Key)
	}
	return res
}

func hydrateString(userTemplate string, data *map[string]any) (any, error) {
	if data == nil {
		data = &map[string]any{}
	}

	// Early exit: if string doesn't contain template markers, return as-is
	// This is a huge optimization - most strings don't have templates!
	if !strings.Contains(userTemplate, "{{") && !strings.Contains(userTemplate, "%(") {
		return userTemplate, nil
	}

	var missingKeys []string

	// Check if the entire string is a single template variable
	if keys := findTemplateKeysToHydrate(userTemplate, wholeVarRegex, nil); len(keys) == 1 {
		key := keys[0]

		// Before treating as variable, check if this is actually a known function
		// This handles parameterless functions like genUUID, now, noop, etc.
		if _, isKnownFunction := funcMap[key.Key]; isKnownFunction {
			// This is a function, not a variable - try to process as function
			if value, err := processSingleFunction(key.Key, *data, &missingKeys); err == nil {
				return value, nil
			}
			// If function processing fails, fall through to template parsing
		} else {
			// This is actually a variable, process normally
			if value, err := processSingleVariable(key, *data, &missingKeys); err == nil {
				return value, nil
			} else if err != nil {
				// If this is an optional key and it's missing, return nil
				if key.IsOptional {
					return nil, nil
				}
				return nil, err
			}

			if key.IsOptional {
				return nil, nil
			}

			return nil, &InfoNeededError{
				MissingKeys:   getKeyStrings(keys),
				AvailableKeys: getKeys(*data),
				Err:           fmt.Errorf("missing keys in template"),
			}
		}
	}

	// Or if the entire string is a function
	if keys := findTemplateKeysToHydrate(userTemplate, wholeFuncRegex, nil); len(keys) == 1 {
		key := keys[0]

		// Try to process as a single function call (optimization path)
		if value, err := processSingleFunction(key.Key, *data, &missingKeys); err == nil {
			return value, nil
		} else {
			// Single function processing failed, falling back to full template parsing
			// This should rarely happen now that we support slice arguments and complex nested calls
		}
	}

	// Find all template keys before processing
	allKeys := FindTemplateKeysToHydrate(userTemplate, true, nil)

	// Pre-process the template to handle optional parameters
	// Replace optional parameters that are missing with empty strings in the template
	preprocessedTemplate := userTemplate
	for _, key := range allKeys {
		if key.IsOptional {
			// Check if the parameter exists in data
			value := get(key.Key, *data, &[]string{})
			if value == nil {
				// Replace the optional parameter with empty string using string replacement
				optPattern := fmt.Sprintf("{{%s?}}", key.Key)
				preprocessedTemplate = strings.ReplaceAll(preprocessedTemplate, optPattern, "")
			}
		}
	}

	// Parse the user template and convert it to use our custom get function
	parsedTemplate, err := parseTemplate(preprocessedTemplate)
	if err != nil {
		return nil, err
	}

	// Create a custom template with our functions
	var t = template.New("custom").Funcs(funcMap)

	// Add custom helpers for variable processing
	t = addCustomTemplateHelpers(t, *data)

	// Set option to error on missing keys
	t.Option("missingkey=error")

	// Parse the modified template
	t, err = t.Parse(parsedTemplate)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %w", err)
	}

	keyDefinitions := KeyDefinitions{}
	// Only include non-optional keys or optional keys that exist in data
	for _, key := range allKeys {
		if !key.IsOptional || get(key.Key, *data, &[]string{}) != nil {
			keyDefinitions[key.Key] = key
		}
	}

	// Execute the template
	var result bytes.Buffer
	templateData := struct {
		Data           map[string]any
		MissingKeys    *[]string
		KeyDefinitions KeyDefinitions
	}{Data: *data, MissingKeys: &missingKeys, KeyDefinitions: keyDefinitions}

	err = t.Execute(&result, templateData)

	// Check for missing key errors
	if err != nil {
		// Only log actual template errors, not argument type mismatches which are expected
		if !strings.Contains(err.Error(), "invalid value; expected int") {
			log.Printf("template execution error: %v", err)
		}

		// Parse the error message to extract missing keys
		errorMsg := err.Error()
		re := regexp.MustCompile(`map has no entry for key "(.*?)"`)
		matches := re.FindAllStringSubmatch(errorMsg, -1)

		// If it's not a missing key error, return the original error
		if len(matches) == 0 {
			log.Printf("template error: %v", errorMsg)
			return nil, fmt.Errorf("template error (not a missing key error): %w", err)
		}
		for _, match := range matches {
			if len(match) > 1 {
				// Check if the key is optional before adding it to missingKeys
				key := match[1]
				isOptional := false
				for _, k := range allKeys {
					if k.Key == key && k.IsOptional {
						isOptional = true
						break
					}
				}
				if !isOptional {
					missingKeys = append(missingKeys, key)
				}
			}
		}
	}

	strResult := result.String()

	// Manually handle any remaining optional parameters in the template result
	// Replace any remaining {{key?}} patterns with empty strings
	strResult = optionalVarRegex.ReplaceAllString(strResult, "")

	// Check for missing keys from both error and get function
	// Filter out optional keys from missingKeys
	var requiredMissingKeys []string
	for _, key := range missingKeys {
		isOptional := false
		for _, k := range allKeys {
			if k.Key == key && k.IsOptional {
				isOptional = true
				break
			}
		}
		if !isOptional {
			requiredMissingKeys = append(requiredMissingKeys, key)
		}
	}

	if len(requiredMissingKeys) > 0 {
		// return partial result as may have some vars in a previous
		// hydration that would be missing from the current params
		// (e.g. a string that uses both system params and step output)
		return strResult, &InfoNeededError{
			MissingKeys:   requiredMissingKeys,
			AvailableKeys: getKeys(*data),
			Err:           fmt.Errorf("missing keys in template"),
		}
	}

	// Special case for optional variables that were returned as template strings
	// If the result matches the pattern {{key?}}, and key is optional, return nil
	if optVarMatches := wholeVarRegex.FindStringSubmatch(strResult); len(optVarMatches) > 0 {
		matchedKey := strings.TrimSpace(optVarMatches[1])
		isOptional := strings.HasSuffix(matchedKey, "?")
		if isOptional {
			// It's an optional parameter that wasn't hydrated, so return nil
			return nil, nil
		}
	}

	// Check if the value is a number and preserve its type
	if value, err := strconv.Atoi(strResult); err == nil {
		return value, nil
	}

	// If there was an error but no missing keys were found, return the original error
	if err != nil {
		return strResult, fmt.Errorf("error executing template: %w", err)
	}

	return strResult, nil
}

func processSingleVariable(key Key, data map[string]any, missingKeys *[]string) (any, error) {
	value := get(key.Key, data, missingKeys)
	if value != nil {
		return value, nil
	}

	if key.IsOptional {
		return nil, nil
	}

	return nil, &InfoNeededError{
		MissingKeys:   []string{key.Key},
		AvailableKeys: getKeys(data),
		Err:           fmt.Errorf("missing key in template"),
	}
}

func processSingleFunction(funcCall string, data map[string]any, missingKeys *[]string) (any, error) {
	// Handle reserved words
	if err := validateFunctionName(funcCall); err != nil {
		return nil, err
	}

	// Handle nested function calls
	if result, handled, err := processNestedFunction(funcCall, data, missingKeys); handled {
		return result, err
	}

	// Parse regular function call
	funcName, processedArgs, err := parseFunctionCall(funcCall, data, missingKeys)
	if err != nil {
		return nil, err
	}

	// Execute the function
	return executeFunctionCall(funcName, processedArgs, data, missingKeys)
}

// validateFunctionName checks if the function call starts with a reserved word
func validateFunctionName(funcCall string) error {
	for _, reserved := range reservedWords {
		if strings.HasPrefix(funcCall, reserved+" ") {
			return fmt.Errorf("reserved word used as function: %s", reserved)
		}
	}
	return nil
}

// processNestedFunction handles nested function calls like "toJSON (mapToDict ...)"
func processNestedFunction(funcCall string, data map[string]any, missingKeys *[]string) (any, bool, error) {
	if !strings.Contains(funcCall, "(") || !strings.Contains(funcCall, ")") {
		return nil, false, nil // Not a nested function
	}

	// Use parseQuotedFields to properly parse the function call
	parts := parseQuotedFields(funcCall)
	if len(parts) < 2 {
		return nil, false, nil // Not enough parts for a nested function
	}

	outerFunc := parts[0]

	// Check if this looks like "func1 (func2 args...)" pattern
	// This should have exactly 2 parts: the outer function and one parenthesized expression
	if len(parts) != 2 {
		return nil, false, nil // Multiple arguments, let parseFunctionCall handle it
	}

	innerCall := strings.TrimSpace(parts[1])
	if !strings.HasPrefix(innerCall, "(") || !strings.HasSuffix(innerCall, ")") {
		return nil, false, nil // Not a single parenthesized expression
	}

	// Extract and process inner function
	innerFunc := innerCall[1 : len(innerCall)-1]
	innerResult, err := processSingleFunction(innerFunc, data, missingKeys)
	if err != nil {
		return nil, true, err
	}

	// Execute outer function with inner result
	fn, ok := funcMap[outerFunc]
	if !ok {
		return nil, true, fmt.Errorf("unknown function: %s", outerFunc)
	}

	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	// Check if this is a data function (needs data and missingKeys)
	isDataFunction := false
	for _, funcName := range funcsThatNeedData {
		if funcName == outerFunc {
			isDataFunction = true
			break
		}
	}

	if fnType.NumIn() == 1 {
		// Simple function with one argument
		if innerResult == nil {
			return nil, true, fmt.Errorf("cannot call function %s with nil argument from inner function %s", outerFunc, innerFunc)
		}
		callArgs := []reflect.Value{reflect.ValueOf(innerResult)}
		results := fnValue.Call(callArgs)
		return processResults(results), true, nil
	} else if isDataFunction && fnType.NumIn() == 3 {
		// Data function with (arg, data, missingKeys)
		if innerResult == nil {
			return nil, true, fmt.Errorf("cannot call function %s with nil argument from inner function %s", outerFunc, innerFunc)
		}
		callArgs := []reflect.Value{
			reflect.ValueOf(innerResult),
			reflect.ValueOf(data),
			reflect.ValueOf(missingKeys),
		}
		results := fnValue.Call(callArgs)
		return processResults(results), true, nil
	}

	return nil, true, fmt.Errorf("unsupported function signature for nested call: %s (numIn: %d, isDataFunction: %v)", outerFunc, fnType.NumIn(), isDataFunction)
}

// parseQuotedFields parses a string into fields, respecting quoted strings (both single and double quotes)
// and parentheses for nested function calls
func parseQuotedFields(s string) []string {
	var fields []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)
	parenDepth := 0

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if !inQuotes {
			if ch == '"' || ch == '\'' {
				inQuotes = true
				quoteChar = ch
				current.WriteByte(ch)
			} else if ch == '(' {
				parenDepth++
				current.WriteByte(ch)
			} else if ch == ')' {
				parenDepth--
				current.WriteByte(ch)
			} else if (ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r') && parenDepth == 0 {
				// Whitespace outside quotes and parentheses - end current field
				if current.Len() > 0 {
					fields = append(fields, current.String())
					current.Reset()
				}
			} else {
				current.WriteByte(ch)
			}
		} else {
			// Inside quotes
			current.WriteByte(ch)
			if ch == quoteChar {
				inQuotes = false
				quoteChar = 0
			}
		}
	}

	// Add final field if any
	if current.Len() > 0 {
		fields = append(fields, current.String())
	}

	return fields
}

// parseFunctionCall parses a function call string and returns the function name and processed arguments
func parseFunctionCall(funcCall string, data map[string]any, missingKeys *[]string) (string, []any, error) {
	parts := parseQuotedFields(funcCall)
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("empty function call")
	}

	funcName := parts[0]
	if slices.Contains(reservedWords, funcName) {
		return "", nil, fmt.Errorf("unknown function: %s", funcName)
	}

	args := parts[1:]
	processedArgs := make([]any, len(args))

	// Process each argument, evaluating nested functions recursively
	for i, arg := range args {
		// Check if this argument is a nested function call
		if strings.HasPrefix(arg, "(") && strings.HasSuffix(arg, ")") {
			// Extract the nested function and evaluate it
			nestedFunc := arg[1 : len(arg)-1]
			if result, err := processSingleFunction(nestedFunc, data, missingKeys); err == nil {
				processedArgs[i] = result
			} else {
				// If evaluation fails, pass the string as-is
				processedArgs[i] = arg
			}
		} else {
			// Process non-function arguments normally
			processedArgs[i] = processArgument(arg, data, missingKeys)
		}
	}

	return funcName, processedArgs, nil
}

// interpretEscapeSequences converts common escape sequences to their actual characters
func interpretEscapeSequences(s string) string {
	// Handle common escape sequences
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\r", "\r")
	s = strings.ReplaceAll(s, "\\\"", "\"")
	s = strings.ReplaceAll(s, "\\'", "'")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

// processArgument processes a single function argument, handling data references and quote stripping
func processArgument(arg string, data map[string]any, missingKeys *[]string) any {
	// Note: Nested function calls are now handled in parseFunctionCall
	// This function only handles non-function arguments

	// Check if this is a nested function call that wasn't processed yet
	// (This can happen when processArgument is called directly from tests)
	if strings.HasPrefix(arg, "(") && strings.HasSuffix(arg, ")") {
		// Extract the function call and process it
		funcCall := arg[1 : len(arg)-1]
		if value, err := processSingleFunction(funcCall, data, missingKeys); err == nil {
			return value
		}
		// If processing failed, return the original
		return arg
	}

	// Handle quoted strings
	if strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"") {
		// Remove quotes and interpret escape sequences
		unquoted := strings.Trim(arg, "\"")
		return interpretEscapeSequences(unquoted)
	}

	clean := strings.Trim(arg, "\"'")

	// Handle data references
	if strings.HasPrefix(clean, "$.Data.") || strings.HasPrefix(clean, ".Data.") {
		var path string
		if strings.HasPrefix(clean, "$.Data.") {
			path = strings.TrimPrefix(clean, "$.Data.")
		} else {
			path = strings.TrimPrefix(clean, ".Data.")
		}
		return get(path, data, missingKeys)
	}

	// Handle special parameters
	if clean == ".Data" || clean == "$.Data" {
		return data
	}
	if clean == ".MissingKeys" || clean == "$.MissingKeys" {
		return missingKeys
	}

	return clean
}

// executeFunctionCall executes a function with the given arguments
func executeFunctionCall(funcName string, processedArgs []any, data map[string]any, missingKeys *[]string) (any, error) {
	fn, ok := funcMap[funcName]
	if !ok {
		return nil, fmt.Errorf("unknown function: %s", funcName)
	}

	_, isBasicFunction := basicFuncMap[funcName]
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	var callArgs []reflect.Value
	var err error

	if isBasicFunction {
		callArgs, err = prepareBasicFunctionArgs(funcName, processedArgs, fnType)
	} else {
		callArgs, err = prepareDataFunctionArgs(funcName, processedArgs, fnType, data, missingKeys)
	}

	if err != nil {
		return nil, err
	}

	results := fnValue.Call(callArgs)
	return processResults(results), nil
}

// prepareBasicFunctionArgs prepares arguments for basic functions (no data/missingKeys needed)
func prepareBasicFunctionArgs(funcName string, processedArgs []any, fnType reflect.Type) ([]reflect.Value, error) {
	expectedArgCount := len(processedArgs)
	if fnType.NumIn() != expectedArgCount {
		return nil, fmt.Errorf("incorrect number of arguments for function %s: expected %d, got %d", funcName, fnType.NumIn(), len(processedArgs))
	}

	callArgs := make([]reflect.Value, fnType.NumIn())
	for i := 0; i < len(processedArgs); i++ {
		arg, err := convertArgument(processedArgs[i], fnType.In(i), i)
		if err != nil {
			return nil, err
		}
		callArgs[i] = arg
	}

	return callArgs, nil
}

// prepareDataFunctionArgs prepares arguments for data functions (need data/missingKeys)
func prepareDataFunctionArgs(funcName string, processedArgs []any, fnType reflect.Type, data map[string]any, missingKeys *[]string) ([]reflect.Value, error) {
	if fnType.NumIn() < 2 {
		return nil, fmt.Errorf("function %s must have data and missingKeys parameters", funcName)
	}

	// Check if the last two args are already data and missingKeys
	hasDataAndMissingKeys := false
	if len(processedArgs) >= 2 {
		lastArg := processedArgs[len(processedArgs)-1]
		secondLastArg := processedArgs[len(processedArgs)-2]

		// Check if these are the data and missingKeys parameters
		if _, isDataMap := secondLastArg.(map[string]any); isDataMap {
			if _, isMissingKeys := lastArg.(*[]string); isMissingKeys {
				hasDataAndMissingKeys = true
			}
		}
	}

	var expectedArgCount int
	if hasDataAndMissingKeys {
		// Data and missingKeys are already in processedArgs
		expectedArgCount = len(processedArgs)
	} else {
		// Need to add data and missingKeys
		expectedArgCount = len(processedArgs) + 2
	}

	if fnType.NumIn() != expectedArgCount {
		return nil, fmt.Errorf("incorrect number of arguments for function %s: expected %d, got %d", funcName, fnType.NumIn()-2, len(processedArgs))
	}

	callArgs := make([]reflect.Value, fnType.NumIn())

	if hasDataAndMissingKeys {
		// All args are already in processedArgs
		for i := 0; i < len(processedArgs); i++ {
			arg, err := convertArgument(processedArgs[i], fnType.In(i), i)
			if err != nil {
				return nil, err
			}
			callArgs[i] = arg
		}
	} else {
		// Need to add data and missingKeys
		for i := 0; i < len(processedArgs); i++ {
			arg, err := convertArgument(processedArgs[i], fnType.In(i), i)
			if err != nil {
				return nil, err
			}
			callArgs[i] = arg
		}
		callArgs[len(processedArgs)] = reflect.ValueOf(data)
		callArgs[len(processedArgs)+1] = reflect.ValueOf(missingKeys)
	}

	return callArgs, nil
}

// convertArgument converts a processed argument to the expected type using reflection
func convertArgument(paramValue any, paramType reflect.Type, argIndex int) (reflect.Value, error) {
	// Handle nil values
	if paramValue == nil {
		// For nullable types like interfaces, pointers, slices, maps, nil is valid
		switch paramType.Kind() {
		case reflect.Interface, reflect.Ptr, reflect.Slice, reflect.Map:
			return reflect.Zero(paramType), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert nil to non-nullable type %v for argument %d", paramType, argIndex)
		}
	}

	switch paramType.Kind() {
	case reflect.String:
		if strVal, ok := paramValue.(string); ok {
			return reflect.ValueOf(strVal), nil
		}
		return reflect.ValueOf(fmt.Sprintf("%v", paramValue)), nil

	case reflect.Int:
		intVal, err := convertToInt(paramValue)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("error converting argument %d to int: %w", argIndex, err)
		}
		return reflect.ValueOf(intVal), nil

	case reflect.Interface:
		return reflect.ValueOf(paramValue), nil

	case reflect.Slice:
		// Handle slice types - convert []any to the expected slice type
		if sliceVal, ok := paramValue.([]any); ok {
			return reflect.ValueOf(sliceVal), nil
		}
		// If it's not []any, try to convert it to []any
		valueReflect := reflect.ValueOf(paramValue)
		if valueReflect.Kind() == reflect.Slice {
			// Convert any slice type to []any
			resultSlice := make([]any, valueReflect.Len())
			for i := 0; i < valueReflect.Len(); i++ {
				resultSlice[i] = valueReflect.Index(i).Interface()
			}
			return reflect.ValueOf(resultSlice), nil
		}
		return reflect.Value{}, fmt.Errorf("cannot convert argument %d to slice: expected slice, got %T", argIndex, paramValue)

	default:
		return reflect.Value{}, fmt.Errorf("unsupported argument type for argument %d: %v", argIndex, paramType.Kind())
	}
}

// convertToInt converts various types to int
func convertToInt(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot convert type %T to int", value)
	}
}

// processResults processes the results from a function call
func processResults(results []reflect.Value) any {
	if len(results) == 0 {
		return nil
	}

	if len(results) == 2 && !results[1].IsNil() {
		// If there's an error in the second result, we should handle it in the caller
		// For now, just return the first result
		return results[0].Interface()
	}

	return results[0].Interface()
}

func hydrateDict(dict any, stateParameters *map[string]any, parameterHydrationBehaviour *map[string]any) (map[string]any, error) {
	var typedDict map[string]any

	switch d := dict.(type) {
	case map[string]any:
		typedDict = d
	default:
		return nil, fmt.Errorf("expected map[string]any or map[string]any, got %T", dict)
	}

	if stateParameters == nil {
		return typedDict, nil
	}

	// Pre-allocate result map with same capacity as input to avoid resizing
	result := make(map[string]any, len(typedDict))
	var missingKeys []string
	var missingKeyPaths []MissingKeyInfo

	// First process all values that need hydration
	for key, value := range typedDict {
		// Check if this specific field should be hydrated
		shouldHydrate, childBehaviour := shouldHydrateField(key, parameterHydrationBehaviour)

		if !shouldHydrate {
			// If not hydrating, just copy the original value
			result[key] = value
			continue
		}

		// For values that need hydration, process them
		hydratedValue, err := Hydrate(value, stateParameters, childBehaviour)

		// Handle string values that might contain unhydrated optional templates
		if strValue, ok := hydratedValue.(string); ok {
			if optVarMatches := wholeVarRegex.FindStringSubmatch(strValue); len(optVarMatches) > 0 {
				matchedKey := strings.TrimSpace(optVarMatches[1])
				isOptional := strings.HasSuffix(matchedKey, "?")
				if isOptional {
					// It's an optional parameter that wasn't hydrated, we should return nil
					result[key] = nil
					continue
				}
			}
		}

		// Check for optional string patterns in the original value
		if hydratedValue == nil && err == nil {
			if strOriginal, ok := value.(string); ok {
				if optVarMatches := wholeVarRegex.FindStringSubmatch(strOriginal); len(optVarMatches) > 0 {
					matchedKey := strings.TrimSpace(optVarMatches[1])
					isOptional := strings.HasSuffix(matchedKey, "?")
					if isOptional {
						// It's an optional parameter, set to nil explicitly
						result[key] = nil
						continue
					}
				}
			}
		}

		// Check for other nil values
		if hydratedValue == nil && err == nil {
			// It's an optional parameter that returned nil, keep it as nil
			result[key] = nil
			continue
		}

		// For all other cases, store the hydrated value if not nil
		if hydratedValue != nil {
			result[key] = hydratedValue
		} else {
			// Default - use original value if hydrated is nil but there was an error
			result[key] = value
		}

		if err != nil {
			var infoNeededErr *InfoNeededError
			if errors.As(err, &infoNeededErr) {
				// Add path to the error
				addPathToError(err, key)
				missingKeys = append(missingKeys, infoNeededErr.MissingKeys...)
				// Collect the paths with updated path info
				missingKeyPaths = append(missingKeyPaths, infoNeededErr.MissingKeyPaths...)
				continue
			}

			return nil, fmt.Errorf("error hydrating key '%s': %w", key, err)
		}
	}

	if len(missingKeys) > 0 || len(missingKeyPaths) > 0 {
		return result, &InfoNeededError{
			MissingKeys:     missingKeys,
			MissingKeyPaths: missingKeyPaths,
			AvailableKeys:   getKeys(*stateParameters),
			Err:             fmt.Errorf("missing keys in dict"),
		}
	}

	return result, nil
}

func hydrateSlice(slice []any, stateParameters *map[string]any, parameterHydrationBehaviour *map[string]any) ([]any, error) {
	if stateParameters == nil {
		return slice, nil
	}

	result := make([]any, len(slice))
	var missingKeys []string
	var missingKeyPaths []MissingKeyInfo

	// Process each slice element
	for i, v := range slice {
		// Hydrate the value
		// We pass down the same behaviour for each element in the slice
		hydratedValue, err := Hydrate(v, stateParameters, parameterHydrationBehaviour)

		// Handle string values that might contain unhydrated optional templates
		if strValue, ok := hydratedValue.(string); ok {
			if optVarMatches := wholeVarRegex.FindStringSubmatch(strValue); len(optVarMatches) > 0 {
				matchedKey := strings.TrimSpace(optVarMatches[1])
				isOptional := strings.HasSuffix(matchedKey, "?")
				if isOptional {
					// It's an optional parameter that wasn't hydrated, we should return nil
					result[i] = nil
					continue
				}
			}
		}

		// Check for optional string patterns in the original value
		if hydratedValue == nil && err == nil {
			if strOriginal, ok := v.(string); ok {
				if optVarMatches := wholeVarRegex.FindStringSubmatch(strOriginal); len(optVarMatches) > 0 {
					matchedKey := strings.TrimSpace(optVarMatches[1])
					isOptional := strings.HasSuffix(matchedKey, "?")
					if isOptional {
						// It's an optional parameter, set to nil explicitly
						result[i] = nil
						continue
					}
				}
			}
		}

		// Check for other nil values
		if hydratedValue == nil && err == nil {
			// It's an optional parameter that returned nil, keep it as nil
			result[i] = nil
			continue
		}

		// Set the hydrated value or original if error occurred
		if hydratedValue != nil {
			result[i] = hydratedValue
		} else {
			result[i] = v
		}

		// Handle errors
		if err != nil {
			var infoNeededErr *InfoNeededError
			if errors.As(err, &infoNeededErr) {
				// Add path to the error with array index
				addPathToError(err, fmt.Sprintf("[%d]", i))
				missingKeys = append(missingKeys, infoNeededErr.MissingKeys...)
				// Collect the paths with updated path info
				missingKeyPaths = append(missingKeyPaths, infoNeededErr.MissingKeyPaths...)
			} else {
				return nil, fmt.Errorf("error hydrating slice index %d: %w", i, err)
			}
		}
	}

	if len(missingKeys) > 0 || len(missingKeyPaths) > 0 {
		return result, &InfoNeededError{
			MissingKeys:     missingKeys,
			MissingKeyPaths: missingKeyPaths,
			AvailableKeys:   getKeys(*stateParameters),
			Err:             fmt.Errorf("missing keys in slice"),
		}
	}

	return result, nil
}

// Public API helper methods - these use Hydrate but then case back to the original type,
// as Hydrate contains parameter pre-processing that we want to do for all
// hydrations (deep copy the params etc so they're not modified by the template),
// so can't export hydrateString etc. above directly
func HydrateDict(dict any, stateParameters *map[string]any, parameterHydrationBehaviour ...*map[string]any) (map[string]any, error) {
	var behaviour *map[string]any
	if len(parameterHydrationBehaviour) > 0 {
		behaviour = parameterHydrationBehaviour[0]
	}

	value, err := Hydrate(dict, stateParameters, behaviour)
	if err != nil {
		return nil, err
	}

	typedDict, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected map[string]any, got %T", value)
	}

	return typedDict, nil
}

func HydrateSlice(slice []any, stateParameters *map[string]any, parameterHydrationBehaviour ...*map[string]any) ([]any, error) {
	var behaviour *map[string]any
	if len(parameterHydrationBehaviour) > 0 {
		behaviour = parameterHydrationBehaviour[0]
	}

	value, err := Hydrate(slice, stateParameters, behaviour)
	if err != nil {
		return nil, err
	}

	typedSlice, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("expected []any, got %T", value)
	}

	return typedSlice, nil
}

func Get(key string, data map[string]any, missingKeys *[]string) any {
	if missingKeys == nil {
		missingKeys = &[]string{}
	}
	value := get(key, data, missingKeys)
	if value == nil {
		return nil
	}
	return value
}

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
	copiedData, err := utils.JSONToDict(utils.JSON(data))
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

	// Convert map[string]any to map[string]any for consistent handling
	if mapValue, ok := current.(map[string]any); ok {
		current = map[string]any(mapValue)
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

func MergeSources(sources ...map[string]any) (map[string]any, error) {
	combinedParams := map[string]any{}

	for _, source := range sources {
		for k, v := range source {
			combinedParams[k] = v
		}
	}

	return combinedParams, nil
}

// isKeyOptional determines if a key is optional either by its suffix (?) or by its definition
// in the keyDefinitions map.
func isKeyOptional(key string, keyDefinitions KeyDefinitions) bool {
	// Check if key has a ? suffix directly
	if strings.HasSuffix(key, "?") {
		return true
	}

	// Clean key for definitions lookup
	cleanKeyStr := strings.TrimSuffix(key, "?")
	cleanKeyStr = removeDataPrefix(cleanKeyStr)

	// Then check definitions
	if keyDefinition, ok := keyDefinitions[cleanKeyStr]; ok {
		return keyDefinition.IsOptional
	}

	return false
}

// handleMissingKey adds a key to the missingKeys slice only if the key is not optional.
// This is used to track which required keys are missing from the data during hydration.
func handleMissingKey(key string, isOptional bool, missingKeys *[]string) {
	if !isOptional {
		// Check if key already exists before adding
		for _, k := range *missingKeys {
			if k == key {
				return
			}
		}
		*missingKeys = append(*missingKeys, key)
	}
}
