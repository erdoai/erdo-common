package template

import (
	"fmt"
	"regexp"
	"strings"
)

var reservedWords = []string{
	"if", "range", "with", "end", "else", "template", "block", "define",
}

// ValidateTemplateSyntax validates basic template syntax
func ValidateTemplateSyntax(templateStr string) []string {
	if templateStr == "" {
		return []string{}
	}
	
	// Check for basic template brace matching
	openCount := strings.Count(templateStr, "{{")
	closeCount := strings.Count(templateStr, "}}")
	
	if openCount > closeCount {
		return []string{fmt.Sprintf("Invalid template syntax in '%s': unclosed template", templateStr)}
	} else if closeCount > openCount {
		return []string{fmt.Sprintf("Invalid template syntax in '%s': unexpected template close", templateStr)}
	}
	
	return []string{}
}

// ExtractTemplateParameters extracts parameter references from template strings
func ExtractTemplateParameters(template string) []string {
	params := []string{}
	seen := make(map[string]bool)
	
	// Extract .Data.field references
	dataPattern := regexp.MustCompile(`\.Data\.([a-zA-Z_][a-zA-Z0-9_.]*)`)
	dataMatches := dataPattern.FindAllStringSubmatch(template, -1)
	for _, match := range dataMatches {
		if len(match) > 1 && !seen[match[1]] {
			params = append(params, match[1])
			seen[match[1]] = true
		}
	}
	
	// Extract $.Data.field references
	dollarDataPattern := regexp.MustCompile(`\$\.Data\.([a-zA-Z_][a-zA-Z0-9_.]*)`)
	dollarMatches := dollarDataPattern.FindAllStringSubmatch(template, -1)
	for _, match := range dollarMatches {
		if len(match) > 1 && !seen[match[1]] {
			params = append(params, match[1])
			seen[match[1]] = true
		}
	}
	
	// Extract simple variable references {{variable}} including optional ones
	varPattern := regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_.]*\??)\s*\}\}`)
	varMatches := varPattern.FindAllStringSubmatch(template, -1)
	for _, match := range varMatches {
		if len(match) > 1 {
			variable := strings.TrimSpace(match[1])
			// Remove optional suffix if present
			variable = strings.TrimSuffix(variable, "?")
			// Skip reserved words and function calls
			if !isReservedWord(variable) && !strings.Contains(variable, "(") && !seen[variable] {
				params = append(params, variable)
				seen[variable] = true
			}
		}
	}
	
	// Extract Python-style parameters
	pythonPattern := regexp.MustCompile(`%\(([a-zA-Z_][a-zA-Z0-9_.]*)\)s`)
	pythonMatches := pythonPattern.FindAllStringSubmatch(template, -1)
	for _, match := range pythonMatches {
		if len(match) > 1 && !seen[match[1]] {
			params = append(params, match[1])
			seen[match[1]] = true
		}
	}
	
	return params
}

func isReservedWord(word string) bool {
	for _, reserved := range reservedWords {
		if word == reserved {
			return true
		}
	}
	return false
}

// FindTemplateKeysToHydrate finds template keys that need hydration
func FindTemplateKeysToHydrate(s any, includeOptional bool, parameterHydrationBehaviour *map[string]any) []Key {
	return findAllTemplateKeysToHydrate(s, includeOptional, parameterHydrationBehaviour)
}

func findTemplateKeysToHydrate(s any, regex *regexp.Regexp, parameterHydrationBehaviour *map[string]any) []Key {
	var keys []Key
	seen := make(map[string]bool)
	includeOptional := regex == directVarRegex
	
	switch v := s.(type) {
	case string:
		matches := regex.FindAllStringSubmatch(v, -1)
		for _, match := range matches {
			if len(match) > 1 {
				key, isValid := cleanKeyForValidation(match[1])
				if isValid && !seen[key] {
					parsedKey := ParseTemplateKey(key)
					// Skip optional keys if includeOptional is false
					if !parsedKey.IsOptional || includeOptional {
						keys = append(keys, parsedKey)
						seen[key] = true
					}
				}
			}
		}
	case []any:
		for _, item := range v {
			subKeys := findTemplateKeysToHydrate(item, regex, parameterHydrationBehaviour)
			for _, key := range subKeys {
				if !seen[key.Key] {
					// Skip optional keys if includeOptional is false
					if !key.IsOptional || includeOptional {
						keys = append(keys, key)
						seen[key.Key] = true
					}
				}
			}
		}
	case map[string]any:
		for fieldKey, value := range v {
			// Check if this field should be hydrated based on behavior
			shouldHydrate, fieldBehaviour := shouldHydrateField(fieldKey, parameterHydrationBehaviour)
			if shouldHydrate {
				subKeys := findTemplateKeysToHydrate(value, regex, fieldBehaviour)
				for _, key := range subKeys {
					if !seen[key.Key] {
						// Skip optional keys if includeOptional is false
						if !key.IsOptional || includeOptional {
							keys = append(keys, key)
							seen[key.Key] = true
						}
					}
				}
			}
			// If field is raw, skip its template keys
		}
	}
	
	return keys
}

func cleanKeyForValidation(key string) (string, bool) {
	key = strings.TrimSpace(key)
	
	// Skip template variables (starting with $)
	if strings.HasPrefix(key, "$") {
		return "", false
	}
	
	// Skip reserved words
	if isReservedWord(key) {
		return "", false
	}
	
	// Skip function calls (contain parentheses)
	if strings.Contains(key, "(") {
		return "", false
	}
	
	return key, true
}

// findAllTemplateKeysToHydrate finds all template keys including those in function calls
func findAllTemplateKeysToHydrate(s any, includeOptional bool, parameterHydrationBehaviour *map[string]any) []Key {
	var keys []Key
	seen := make(map[string]bool)
	
	switch v := s.(type) {
	case string:
		// Find all template expressions in the string
		templateKeys := extractAllTemplateKeys(v, includeOptional)
		for _, key := range templateKeys {
			if !seen[key.Key] {
				keys = append(keys, key)
				seen[key.Key] = true
			}
		}
	case []any:
		for _, item := range v {
			subKeys := findAllTemplateKeysToHydrate(item, includeOptional, parameterHydrationBehaviour)
			for _, key := range subKeys {
				if !seen[key.Key] {
					keys = append(keys, key)
					seen[key.Key] = true
				}
			}
		}
	case map[string]any:
		// Convert to Dict to use shouldHydrateField
		for fieldKey, value := range v {
			// Check if this field should be hydrated based on behavior
			shouldHydrate, fieldBehaviour := shouldHydrateField(fieldKey, parameterHydrationBehaviour)
			if shouldHydrate {
				subKeys := findAllTemplateKeysToHydrate(value, includeOptional, fieldBehaviour)
				for _, key := range subKeys {
					if !seen[key.Key] {
						keys = append(keys, key)
						seen[key.Key] = true
					}
				}
			}
			// If field is raw, skip its template keys
		}
	}
	
	return keys
}

// extractAllTemplateKeys extracts all template keys from a string, including those in function calls
func extractAllTemplateKeys(s string, includeOptional bool) []Key {
	var keys []Key
	seen := make(map[string]bool)
	
	// Find all template expressions {{ ... }}
	templateRegex := regexp.MustCompile(`\{\{\s*([^}]+)\s*\}\}`)
	matches := templateRegex.FindAllStringSubmatch(s, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			expr := strings.TrimSpace(match[1])
			
			// Extract keys from this expression
			exprKeys := extractKeysFromExpression(expr, includeOptional)
			for _, key := range exprKeys {
				if !seen[key.Key] {
					keys = append(keys, key)
					seen[key.Key] = true
				}
			}
		}
	}
	
	return keys
}

// extractKeysFromExpression extracts template keys from a single template expression
func extractKeysFromExpression(expr string, includeOptional bool) []Key {
	var keys []Key
	
	// Check if this is a simple variable (no spaces, no function calls)
	if !strings.Contains(expr, " ") && !strings.Contains(expr, "(") {
		if key, isValid := cleanKeyForValidation(expr); isValid {
			parsedKey := ParseTemplateKey(key)
			// Include the key if we want optional keys or if it's not optional
			if includeOptional || !parsedKey.IsOptional {
				keys = append(keys, parsedKey)
			}
		}
		return keys
	}
	
	// For function calls, extract quoted string parameters
	// This handles cases like: get "system.messages" or slice "items" 0 5
	quotedStringRegex := regexp.MustCompile(`"([^"]+)"`)
	quotedMatches := quotedStringRegex.FindAllStringSubmatch(expr, -1)
	
	for _, match := range quotedMatches {
		if len(match) > 1 {
			keyStr := match[1]
			// Only add if it looks like a data key (contains dots or is a simple identifier)
			if isDataKey(keyStr) {
				parsedKey := ParseTemplateKey(keyStr)
				// Include the key if we want optional keys or if it's not optional
				if includeOptional || !parsedKey.IsOptional {
					keys = append(keys, parsedKey)
				}
			}
		}
	}
	
	// Also check for direct references to Data fields
	if strings.Contains(expr, ".Data.") || strings.Contains(expr, "$.Data.") {
		dataRegex := regexp.MustCompile(`(?:\$?\.Data\.)([a-zA-Z_][a-zA-Z0-9_.]*)`)
		dataMatches := dataRegex.FindAllStringSubmatch(expr, -1)
		for _, match := range dataMatches {
			if len(match) > 1 {
				keyStr := match[1]
				parsedKey := ParseTemplateKey(keyStr)
				// Include the key if we want optional keys or if it's not optional
				if includeOptional || !parsedKey.IsOptional {
					keys = append(keys, parsedKey)
				}
			}
		}
	}
	
	return keys
}

// isDataKey checks if a string looks like a data key
func isDataKey(s string) bool {
	// Skip if empty or starts with a number
	if s == "" || (s[0] >= '0' && s[0] <= '9') {
		return false
	}
	
	// Skip common non-data strings
	nonDataStrings := map[string]bool{
		"true": true, "false": true, "null": true, "nil": true,
		"eq": true, "ne": true, "lt": true, "le": true, "gt": true, "ge": true,
		"and": true, "or": true, "not": true,
	}
	
	if nonDataStrings[s] {
		return false
	}
	
	// If it contains dots or is a valid identifier, consider it a data key
	return strings.Contains(s, ".") || regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(s)
}

// FindTemplateKeyStringsToHydrate returns string keys that need hydration
func FindTemplateKeyStringsToHydrate(s any, includeOptional bool, parameterHydrationBehaviour *map[string]any) []string {
	keys := FindTemplateKeysToHydrate(s, includeOptional, parameterHydrationBehaviour)
	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = key.Key
	}
	return result
}