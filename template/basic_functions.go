package template

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"text/template"
	"time"

	"github.com/google/uuid"
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
	"eq":               eq, // Override built-in eq to handle pointers
	"ne":               ne, // Override built-in ne to handle pointers
	"mergeRaw":         mergeRaw,
	"nilToEmptyString": nilToEmptyString,
	"truthyValue":      truthyValue,
	"toString":         toString,
	"truncateString":   truncateString,
	"regexReplace":     regexReplace,
	"noop":             noop,
	"genUUID":          genUUID,
	"generateUUID":     genUUID,
	"list":             list,
	"now":              now,
	"endsWith":         endsWith,
	"startsWith":       startsWith,
}

func genUUID() string {
	return uuid.New().String()
}

func now() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

func toJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return string(b)
}

func truthy(key string, data map[string]any) bool {
	// pass empty list for missing keys as we only want to check if the key exists & is truthy
	val := get(key, data, &[]string{})
	if val == nil {
		return false
	}

	return truthyValue(val)
}

func truthyValue(val any) bool {
	// Unwrap null types and dereference pointers first
	unwrapped, valid := unwrapNullValue(val)
	if !valid {
		return false
	}
	val = unwrapped

	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case []any:
		return len(v) > 0
	case []map[string]any:
		return len(v) > 0
	case map[string]any:
		return len(v) > 0
	case nil:
		return false
	default:
		// Use reflection to handle any other slice, array, or map type
		reflectVal := reflect.ValueOf(val)
		kind := reflectVal.Kind()
		if kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map {
			return reflectVal.Len() > 0
		}
		return true
	}
}

func mergeRaw(array1 []any, array2 []any) []any {
	result := make([]any, 0, len(array1)+len(array2))
	result = append(result, array1...)
	result = append(result, array2...)
	return result
}

// nilToEmptyString is a template function that converts nil to empty string
// Also dereferences pointers and unwraps SQL null types before converting
func nilToEmptyString(v any) string {
	unwrapped, valid := unwrapNullValue(v)
	if !valid || unwrapped == nil {
		return ""
	}
	return fmt.Sprintf("%v", unwrapped)
}

func add(a, b int) int {
	return a + b
}

func sub(a, b any) int {
	aInt := convertToIntForArithmetic(a)
	bInt := convertToIntForArithmetic(b)
	return aInt - bInt
}

// convertToIntForArithmetic converts various types to int, returning 0 for invalid types
func convertToIntForArithmetic(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
		return 0
	case nil:
		return 0
	default:
		// Handle the case where $.MissingKeys is passed instead of a proper value
		return 0
	}
}

func _len(a any) int {
	// Unwrap null types and dereference pointers first
	unwrapped, valid := unwrapNullValue(a)
	if !valid || unwrapped == nil {
		return 0
	}
	a = unwrapped

	switch v := a.(type) {
	case []any:
		return len(v)
	case []map[string]any:
		return len(v)
	case string:
		return len(v)
	case map[string]any:
		return len(v)
	default:
		// Use reflection to handle any other slice, array, or map type
		// This handles other slice/map types that may not match the specific types above
		val := reflect.ValueOf(a)
		kind := val.Kind()
		if kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map {
			return val.Len()
		}
	}

	log.Printf("unsupported type for len: %T %v", a, a)

	return 0
}

func gt(a, b int) bool {
	return a > b
}

func lt(a, b int) bool {
	return a < b
}

func toString(value any) string {
	// Dereference pointers and unwrap SQL null types before converting to string
	unwrapped, valid := unwrapNullValue(value)
	if !valid || unwrapped == nil {
		return ""
	}
	return fmt.Sprintf("%v", unwrapped)
}

// regexReplace performs regex replacement on a string
func regexReplace(pattern, replacement string, value any) string {
	str := toString(value)
	re, err := regexp.Compile(pattern)
	if err != nil {
		// If regex compilation fails, return original string
		return str
	}
	return re.ReplaceAllString(str, replacement)
}

// truncateString returns the first n characters of a string, adding "..." if truncated
// It can handle various input types by converting them to strings first
// Automatically unwraps SQL null types and dereferences pointers
func truncateString(value any, n int) string {
	if n <= 0 {
		return ""
	}

	// Convert the value to a string using toString (which handles null types and pointers)
	str := toString(value)

	// Convert to runes to handle Unicode characters properly
	runes := []rune(str)
	if n >= len(runes) {
		return str
	}

	// If we need to truncate and there's room for "...", use n-3 characters + "..."
	if n > 3 {
		return string(runes[:n-3]) + "..."
	}

	// If n <= 3, just return the first n characters without "..."
	return string(runes[:n])
}

// noop is a template function that does nothing and returns an empty string
// Useful for whitespace removal in templates where you want {{- noop}} syntax
func noop() string {
	return ""
}

// list creates a slice from the given arguments
// Usage: {{list "item1" "item2" "item3"}} returns []any{"item1", "item2", "item3"}
func list(args ...any) []any {
	return args
}

// derefValue dereferences a pointer value if it's a pointer, otherwise returns the value as-is
func derefValue(v any) any {
	if v == nil {
		return nil
	}

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		return val.Elem().Interface()
	}

	return v
}

// unwrapNullValue unwraps SQL null types (like sql.NullString, sql.NullInt64, etc.)
// and their JSON-serialized map representations.
// Returns the unwrapped value and whether it was valid (non-null).
// If the value is not a null type, returns the original value and true.
func unwrapNullValue(v any) (any, bool) {
	if v == nil {
		return nil, false
	}

	// First dereference any pointer
	v = derefValue(v)
	if v == nil {
		return nil, false
	}

	// Check for map representation (from JSON serialization)
	// e.g., map[string]any{"String": "value", "Valid": true}
	if m, ok := v.(map[string]any); ok {
		return unwrapNullMap(m)
	}

	// Check for struct types with Valid field (sql.NullString, sql.NullInt64, etc.)
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Struct {
		return unwrapNullStruct(val)
	}

	// Not a null type, return as-is
	return v, true
}

// unwrapNullMap handles map representations of SQL null types from JSON serialization
// e.g., map[string]any{"String": "value", "Valid": true}
func unwrapNullMap(m map[string]any) (any, bool) {
	// Check if this looks like a serialized null type
	validVal, hasValid := m["Valid"]
	if !hasValid {
		// Not a null type map
		return m, true
	}

	// Check if Valid is false
	valid, ok := validVal.(bool)
	if !ok {
		// Valid field is not a bool, not a null type
		return m, true
	}

	if !valid {
		return nil, false
	}

	// Try common null type value field names
	// sql.NullString has "String", sql.NullInt64 has "Int64", etc.
	valueFieldNames := []string{"String", "Int64", "Int32", "Int16", "Float64", "Bool", "Time", "Byte"}
	for _, fieldName := range valueFieldNames {
		if val, hasField := m[fieldName]; hasField {
			return val, true
		}
	}

	// Has Valid=true but no recognized value field, return as-is
	return m, true
}

// unwrapNullStruct handles actual SQL null type structs
func unwrapNullStruct(val reflect.Value) (any, bool) {
	validField := val.FieldByName("Valid")
	if !validField.IsValid() || validField.Kind() != reflect.Bool {
		// Not a null type struct
		return val.Interface(), true
	}

	if !validField.Bool() {
		return nil, false
	}

	// Try common value field names
	valueFieldNames := []string{"String", "Int64", "Int32", "Int16", "Float64", "Bool", "Time", "Byte"}
	for _, fieldName := range valueFieldNames {
		valueField := val.FieldByName(fieldName)
		if valueField.IsValid() {
			return valueField.Interface(), true
		}
	}

	// Has Valid=true but no recognized value field, return original
	return val.Interface(), true
}

// eq performs pointer-aware equality comparison, automatically dereferencing pointers
// and unwrapping SQL null types.
// Overrides Go template's built-in eq to handle pointer fields in structs
// Special case: nil pointers and invalid null types are treated as empty strings for comparison purposes
// Handles type aliases (like DatasetType string) by comparing underlying values
// Usage: {{if eq $r.Dataset.Name "foo"}}...{{end}}
func eq(args ...any) bool {
	if len(args) == 0 {
		return false
	}

	// Unwrap first argument (handles pointers and SQL null types)
	first, firstValid := unwrapNullValue(args[0])
	if !firstValid {
		first = nil
	}

	// Compare with all other arguments - all must be equal
	for i := 1; i < len(args); i++ {
		other, otherValid := unwrapNullValue(args[i])
		if !otherValid {
			other = nil
		}

		// Special handling for nil comparison with empty string
		// This makes nil pointer/invalid null equivalent to empty string for template convenience
		if (first == nil && other == "") || (first == "" && other == nil) {
			continue
		}

		// Handle nil cases
		if first == nil && other == nil {
			continue
		}
		if first == nil || other == nil {
			return false
		}

		// Try reflect.DeepEqual first (handles most cases)
		if reflect.DeepEqual(first, other) {
			continue
		}

		// If DeepEqual fails, check if BOTH values are string-based types
		// This handles type aliases like DatasetType string vs string literals
		// Only use string comparison if both have string as their underlying kind
		firstVal := reflect.ValueOf(first)
		otherVal := reflect.ValueOf(other)
		firstKind := firstVal.Kind()
		otherKind := otherVal.Kind()

		if firstKind == reflect.String && otherKind == reflect.String {
			// Both are strings (or string aliases), so string comparison is appropriate
			firstStr := fmt.Sprintf("%v", first)
			otherStr := fmt.Sprintf("%v", other)
			if firstStr != otherStr {
				return false
			}
		} else {
			// Different types, so DeepEqual failure means they're not equal
			return false
		}
	}

	return true
}

// ne performs pointer-aware inequality comparison, automatically dereferencing pointers
// Overrides Go template's built-in ne to handle pointer fields in structs
// Usage: {{if ne $r.Dataset.Description ""}}...{{end}}
func ne(args ...any) bool {
	return !eq(args...)
}

// endsWith checks if a string ends with a given suffix
// Automatically handles pointer dereferencing, SQL null types, and nil values
// Usage: {{endsWith .Data.filename ".csv"}}
func endsWith(str any, suffix string) bool {
	s := toString(str) // toString already handles null types, pointers, and nil
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

// startsWith checks if a string starts with a given prefix
// Automatically handles pointer dereferencing, SQL null types, and nil values
// Usage: {{startsWith .Data.filename "prefix_"}}
func startsWith(str any, prefix string) bool {
	s := toString(str) // toString already handles null types, pointers, and nil
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
