package template

import (
	"fmt"
	"regexp"
	"strings"
)

var reservedWords = []string{"if", "range", "with", "end", "else", "template", "block", "define"}

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
func FindTemplateKeysToHydrate(s any, includeOptional bool, parameterHydrationBehaviour *Dict) []Key {
	var regex *regexp.Regexp
	if includeOptional {
		regex = directVarRegex
	} else {
		regex = funcRegex
	}
	
	return findTemplateKeysToHydrate(s, regex, parameterHydrationBehaviour)
}

func findTemplateKeysToHydrate(s any, regex *regexp.Regexp, parameterHydrationBehaviour *Dict) []Key {
	var keys []Key
	seen := make(map[string]bool)
	includeOptional := regex == directVarRegex
	
	switch v := s.(type) {
	case string:
		matches := regex.FindAllStringSubmatch(v, -1)
		for _, match := range matches {
			if len(match) > 1 {
				key, isValid := cleanKey(match[1])
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
	case Dict:
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
		// Convert to Dict to use shouldHydrateField
		dict := make(Dict)
		for k, val := range v {
			dict[k] = val
		}
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

func cleanKey(key string) (string, bool) {
	key = strings.TrimSpace(key)
	
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

// FindTemplateKeyStringsToHydrate returns string keys that need hydration
func FindTemplateKeyStringsToHydrate(s any, includeOptional bool, parameterHydrationBehaviour *Dict) []string {
	keys := FindTemplateKeysToHydrate(s, includeOptional, parameterHydrationBehaviour)
	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = key.Key
	}
	return result
}