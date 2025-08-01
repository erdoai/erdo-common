package template

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"text/template"
	"time"

	json "github.com/goccy/go-json"
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

func mergeRaw(array1 []any, array2 []any) []any {
	result := make([]any, 0, len(array1)+len(array2))
	result = append(result, array1...)
	result = append(result, array2...)
	return result
}

// nilToEmptyString is a template function that converts nil to empty string
func nilToEmptyString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
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
	return fmt.Sprintf("%v", value)
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
func truncateString(value any, n int) string {
	if n <= 0 {
		return ""
	}

	// Handle nil values
	if value == nil {
		return ""
	}

	// Handle sql.NullString types specifically
	if reflect.TypeOf(value).String() == "sql.NullString" {
		v := reflect.ValueOf(value)
		valid := v.FieldByName("Valid").Bool()
		if !valid {
			return ""
		}
		str := v.FieldByName("String").String()

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

	// Convert the value to a string using the existing toString function
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
