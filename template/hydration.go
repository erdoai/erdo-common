package template

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"text/template"
)

// ParameterHydrationBehaviour constants
const (
	ParameterHydrationBehaviourRaw      = "raw"
	ParameterHydrationBehaviourHydrated = "hydrated"
)

// hydrateString processes template strings with variable substitution
func hydrateString(userTemplate string, data *Dict) (string, error) {
	if data == nil {
		return userTemplate, nil
	}

	var missingKeys []string

	// Check if this is a simple variable replacement (whole string is a template)
	if matches := wholeVarRegex.FindStringSubmatch(userTemplate); len(matches) > 0 {
		key := ParseTemplateKey(matches[1])
		if key.IsOptional {
			// For optional keys, return empty string if missing
			if value := get(key.Key, *data, &missingKeys); value != nil {
				return fmt.Sprintf("%v", value), nil
			}
			return "", nil
		} else {
			// For required keys, return the value or error if missing
			if value := get(key.Key, *data, &missingKeys); value != nil {
				return fmt.Sprintf("%v", value), nil
			}
			if len(missingKeys) > 0 {
				availableKeys := getAvailableKeys(*data)
				return "", &InfoNeededError{
					MissingKeys:   missingKeys,
					AvailableKeys: availableKeys,
					Err:           fmt.Errorf("missing key: %s", key.Key),
				}
			}
		}
	}

	// Check if this is a single function call
	if matches := wholeFuncRegex.FindStringSubmatch(userTemplate); len(matches) > 0 {
		key := ParseTemplateKey(matches[1])
		
		// Check if this contains nested function calls that need special parameter handling
		needsTemplateParsing := false
		if strings.Contains(key.Key, "(") && strings.Contains(key.Key, ")") {
			// Check if any nested function needs .Data and .MissingKeys
			for funcName := range dataFuncMap {
				if strings.Contains(key.Key, funcName) {
					needsTemplateParsing = true
					break
				}
			}
		}
		
		if !needsTemplateParsing {
			// Process the function call directly for simple cases
			if value, err := processSingleFunction(key.Key, *data, &missingKeys); err == nil {
				return fmt.Sprintf("%v", value), nil
			} else {
				log.Printf("error hydrating function key: %s, error: %v", key.Key, err)
				// ignore errors and fallback to parsing template
			}
		}
	}

	// For complex templates or multiple substitutions, use Go template engine
	return executeTemplate(userTemplate, *data, &missingKeys)
}

// hydrateDict processes dictionary values with template substitution
func hydrateDict(dict Dict, data *Dict, parameterHydrationBehaviour *Dict) (Dict, error) {
	result := make(Dict)
	
	for key, value := range dict {
		shouldHydrate, fieldBehaviour := shouldHydrateField(key, parameterHydrationBehaviour)
		if shouldHydrate {
			hydratedValue, err := Hydrate(value, data, fieldBehaviour)
			if err != nil {
				return nil, fmt.Errorf("error hydrating key %s: %w", key, err)
			}
			result[key] = hydratedValue
		} else {
			result[key] = value
		}
	}
	
	return result, nil
}

// hydrateSlice processes slice values with template substitution
func hydrateSlice(slice []any, data *Dict, parameterHydrationBehaviour *Dict) ([]any, error) {
	result := make([]any, len(slice))
	
	for i, item := range slice {
		hydratedItem, err := Hydrate(item, data, parameterHydrationBehaviour)
		if err != nil {
			return nil, fmt.Errorf("error hydrating slice item %d: %w", i, err)
		}
		result[i] = hydratedItem
	}
	
	return result, nil
}

// shouldHydrateField determines if a field should be hydrated based on behaviour config
func shouldHydrateField(key string, parameterHydrationBehaviour *Dict) (bool, *Dict) {
	if parameterHydrationBehaviour == nil {
		return true, nil
	}

	// Check if there's a specific behaviour for this key
	if behaviour, exists := (*parameterHydrationBehaviour)[key]; exists {
		switch b := behaviour.(type) {
		case string:
			// If the behavior is a string, it applies to this field directly
			return b != ParameterHydrationBehaviourRaw, nil
		case Dict:
			// If the behavior is a Dict, it contains behaviors for nested fields
			// The field itself should be hydrated, but with the nested behavior
			return true, &b
		case map[string]any:
			// Convert to Dict
			dictB := make(Dict)
			for k, v := range b {
				dictB[k] = v
			}
			// Same as Dict case
			return true, &dictB
		}
	}

	return true, nil
}

// executeTemplate executes a Go template with the provided data
func executeTemplate(userTemplate string, data Dict, missingKeys *[]string) (string, error) {
	// Preprocess the template to handle simple variables and functions
	processedTemplate := preprocessSimpleVariables(userTemplate, data, missingKeys)
	processedTemplate = preprocessTemplate(processedTemplate)
	
	// Create template with custom functions
	tmpl, err := template.New("template").Funcs(funcMap).Parse(processedTemplate)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	// Prepare template data
	templateData := map[string]any{
		"Data":        data,
		"MissingKeys": missingKeys,
	}
	
	// Also add direct access to data values for simple variable substitution
	for k, v := range data {
		templateData[k] = v
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	result := buf.String()

	// Check if we have missing keys after execution
	if len(*missingKeys) > 0 {
		// Deduplicate missing keys
		uniqueKeys := deduplicateMissingKeys(*missingKeys)
		availableKeys := getAvailableKeys(data)
		return "", &InfoNeededError{
			MissingKeys:   uniqueKeys,
			AvailableKeys: availableKeys,
			Err:           fmt.Errorf("missing keys after template execution"),
		}
	}

	return result, nil
}

// preprocessSimpleVariables converts simple {{variable}} patterns to template function calls
func preprocessSimpleVariables(userTemplate string, data Dict, missingKeys *[]string) string {
	// Reserved template words that should not be processed
	reserved := map[string]bool{
		"if": true, "else": true, "end": true, "range": true, 
		"with": true, "define": true, "template": true, "block": true,
	}
	
	// Regular expression to match simple variables (not function calls)
	simpleVarRegex := regexp.MustCompile(`{{(\s*[a-zA-Z_][a-zA-Z0-9_\.]*\??\s*)}}`)
	
	return simpleVarRegex.ReplaceAllStringFunc(userTemplate, func(match string) string {
		// Extract the variable name (without braces and whitespace)
		varName := strings.TrimSpace(match[2:len(match)-2])
		
		// Check if it's an optional variable
		isOptional := strings.HasSuffix(varName, "?")
		cleanVarName := varName
		if isOptional {
			cleanVarName = varName[:len(varName)-1]
		}
		
		// Skip reserved words
		if reserved[cleanVarName] {
			return match
		}
		
		// Skip if it contains a space (likely a function call)
		if strings.Contains(cleanVarName, " ") {
			return match
		}
		
		// Check if the variable exists in data
		val := get(cleanVarName, data, &[]string{})
		if val == nil && !isOptional {
			// Add to missing keys
			*missingKeys = append(*missingKeys, cleanVarName)
			// Convert to a template expression that will be handled properly
			return fmt.Sprintf("{{get %q .Data .MissingKeys}}", cleanVarName)
		}
		
		// For optional variables, always use getOptional
		// This avoids the template parser error with '?' character
		if isOptional {
			return fmt.Sprintf("{{getOptional %q .Data .MissingKeys}}", cleanVarName)
		}
		
		// Return the actual value wrapped in Go template syntax
		return fmt.Sprintf("{{index . %q}}", cleanVarName)
	})
}

// preprocessTemplate adds .Data and .MissingKeys to functions that need them
func preprocessTemplate(userTemplate string) string {
	// Process functions that need .Data and .MissingKeys parameters
	result := userTemplate
	
	// Add data parameters to functions that need them
	for funcName := range dataFuncMap {
		// Simple function call pattern: {{funcName arg1 arg2}}
		_ = fmt.Sprintf(`\{\{\s*%s\s+([^}]+)\}\}`, funcName)
		_ = fmt.Sprintf(`{{%s $1 $.Data $.MissingKeys}}`, funcName)
		
		// Only replace if it doesn't already have .Data or .MissingKeys
		matches := funcRegex.FindAllString(result, -1)
		for _, match := range matches {
			if strings.Contains(match, funcName) && 
			   !strings.Contains(match, ".Data") && 
			   !strings.Contains(match, ".MissingKeys") {
				// This is a more complex replacement that preserves the original structure
				processed := processNestedFunctionCalls(strings.Trim(match, "{}"))
				if processed != strings.Trim(match, "{}") {
					result = strings.ReplaceAll(result, match, "{{"+processed+"}}")
				}
			}
		}
	}
	
	return result
}

// processNestedFunctionCalls handles nested function calls
func processNestedFunctionCalls(funcCall string) string {
	// Check if this contains a nested function call with parentheses
	openParenIndex := strings.Index(funcCall, "(")
	if openParenIndex <= 0 {
		// No nested function, check if this function needs .Data and .MissingKeys
		parts := strings.Fields(funcCall)
		if len(parts) == 0 {
			return funcCall
		}
		
		funcName := parts[0]
		_, requiresData := dataFuncMap[funcName]
		if requiresData && !containsDataSuffix(funcCall) && !containsMissingKeysSuffix(funcCall) {
			return appendDataParams(funcCall)
		}
		return funcCall
	}

	// Extract the outer function name
	outerFunc := strings.TrimSpace(funcCall[:openParenIndex])

	// Process all arguments after the function name
	allArgs := funcCall[openParenIndex+1:] // +1 to skip the opening paren
	// Remove the closing paren if present
	if strings.HasSuffix(allArgs, ")") {
		allArgs = allArgs[:len(allArgs)-1]
	}
	processedArgs := processAllArguments(allArgs)

	// Reconstruct the function call with processed arguments
	result := outerFunc + " (" + processedArgs + ")"

	// Check if the outer function requires .Data and .MissingKeys
	_, outerRequiresData := dataFuncMap[outerFunc]
	if outerRequiresData && !containsDataSuffix(result) && !containsMissingKeysSuffix(result) {
		result = appendDataParams(result)
	}

	return result
}

// processAllArguments processes all arguments in a function call
func processAllArguments(args string) string {
	// First check if this is a function call that should be processed as a whole
	trimmedArgs := strings.TrimSpace(args)
	parts := strings.Fields(trimmedArgs)
	
	// If the first part is a data function and we have more parts, process as a complete function call
	if len(parts) > 0 {
		funcName := parts[0]
		if _, isDataFunc := dataFuncMap[funcName]; isDataFunc && len(parts) > 1 {
			// This is a complete function call that needs data params
			if !containsDataSuffix(trimmedArgs) && !containsMissingKeysSuffix(trimmedArgs) {
				// Add data params after all the existing arguments
			result := trimmedArgs + " $.Data $.MissingKeys"
				return result
			}
			return trimmedArgs
		}
	}
	
	// Otherwise, process arguments individually
	var result strings.Builder
	var currentArg strings.Builder
	parenCount := 0
	inQuotes := false
	quote := byte(0)

	for i := 0; i < len(args); i++ {
		char := args[i]
		
		if !inQuotes && (char == '"' || char == '\'') {
			inQuotes = true
			quote = char
		} else if inQuotes && char == quote {
			inQuotes = false
			quote = 0
		} else if !inQuotes {
			if char == '(' {
				parenCount++
			} else if char == ')' {
				parenCount--
			}
		}
		
		// If we're at a space and not in quotes/parentheses, we've found an argument boundary
		if !inQuotes && parenCount == 0 && char == ' ' {
			arg := strings.TrimSpace(currentArg.String())
			if arg != "" {
				if result.Len() > 0 {
					result.WriteByte(' ')
				}
				result.WriteString(processArgumentExpression(arg))
			}
			currentArg.Reset()
		} else {
			currentArg.WriteByte(char)
		}
	}
	
	// Process the last argument
	arg := strings.TrimSpace(currentArg.String())
	if arg != "" {
		if result.Len() > 0 {
			result.WriteByte(' ')
		}
		result.WriteString(processArgumentExpression(arg))
	}
	
	return result.String()
}

// processArgumentExpression processes a single argument expression
func processArgumentExpression(expr string) string {
	expr = strings.TrimSpace(expr)
	
	// If it's quoted, return as-is
	if (strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"")) ||
	   (strings.HasPrefix(expr, "'") && strings.HasSuffix(expr, "'")) {
		return expr
	}
	
	// Check if this is a nested function call (starts with a function name and has parentheses)
	if strings.Contains(expr, "(") && strings.Contains(expr, ")") {
		// Process it as a nested function call
		return processNestedFunctionCalls(expr)
	}
	
	// For simple expressions without parentheses, check if it's a function that needs parameters
	parts := strings.Fields(expr)
	if len(parts) > 0 {
		funcName := parts[0]
		_, requiresData := dataFuncMap[funcName]
		if requiresData && !containsDataSuffix(expr) && !containsMissingKeysSuffix(expr) {
			return appendDataParams(expr)
		}
	}
	
	return expr
}

// processSingleFunction processes a single function call
func processSingleFunction(funcCall string, data Dict, missingKeys *[]string) (any, error) {
	// Parse the function call
	parts := strings.Fields(funcCall)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty function call")
	}
	
	funcName := parts[0]
	args := parts[1:]
	
	// Process arguments
	processedArgs := make([]any, len(args))
	for i, arg := range args {
		// Convert string arguments to appropriate types
		if strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"") {
			processedArgs[i] = strings.Trim(arg, "\"")
		} else if strings.HasPrefix(arg, "'") && strings.HasSuffix(arg, "'") {
			processedArgs[i] = strings.Trim(arg, "'")
		} else {
			// Try to get from data first
			if value := get(arg, data, missingKeys); value != nil {
				processedArgs[i] = value
			} else {
				processedArgs[i] = arg
			}
		}
	}
	
	// Get the function
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
	
	// Call the function
	results := fnValue.Call(callArgs)
	if len(results) == 0 {
		return nil, nil
	}
	
	// Handle error return values
	if len(results) == 2 && !results[1].IsNil() {
		if err, ok := results[1].Interface().(error); ok {
			return nil, err
		}
	}
	
	return results[0].Interface(), nil
}

// prepareBasicFunctionArgs prepares arguments for basic functions
func prepareBasicFunctionArgs(funcName string, args []any, fnType reflect.Type) ([]reflect.Value, error) {
	numIn := fnType.NumIn()
	callArgs := make([]reflect.Value, numIn)
	
	for i := 0; i < numIn && i < len(args); i++ {
		argType := fnType.In(i)
		arg := args[i]
		
		argValue := reflect.ValueOf(arg)
		if argValue.Type().ConvertibleTo(argType) {
			callArgs[i] = argValue.Convert(argType)
		} else {
			callArgs[i] = argValue
		}
	}
	
	// Fill remaining args with zero values
	for i := len(args); i < numIn; i++ {
		callArgs[i] = reflect.Zero(fnType.In(i))
	}
	
	return callArgs, nil
}

// prepareDataFunctionArgs prepares arguments for data functions
func prepareDataFunctionArgs(funcName string, args []any, fnType reflect.Type, data Dict, missingKeys *[]string) ([]reflect.Value, error) {
	numIn := fnType.NumIn()
	callArgs := make([]reflect.Value, numIn)
	
	// Fill function-specific args
	argIndex := 0
	for i := 0; i < numIn-2 && argIndex < len(args); i++ {
		argType := fnType.In(i)
		arg := args[argIndex]
		
		argValue := reflect.ValueOf(arg)
		if argValue.Type().ConvertibleTo(argType) {
			callArgs[i] = argValue.Convert(argType)
		} else {
			callArgs[i] = argValue
		}
		argIndex++
	}
	
	// Fill remaining function args with zero values
	for i := argIndex; i < numIn-2; i++ {
		callArgs[i] = reflect.Zero(fnType.In(i))
	}
	
	// Add data and missingKeys at the end
	if numIn >= 2 {
		callArgs[numIn-2] = reflect.ValueOf(data)
		callArgs[numIn-1] = reflect.ValueOf(missingKeys)
	}
	
	return callArgs, nil
}

// getAvailableKeys returns available keys from data for error reporting
func getAvailableKeys(data Dict) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	return keys
}

// deduplicateMissingKeys removes duplicate keys from the missing keys list
func deduplicateMissingKeys(keys []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		if !seen[key] {
			seen[key] = true
			result = append(result, key)
		}
	}
	return result
}