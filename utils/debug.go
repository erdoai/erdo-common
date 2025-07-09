package utils

import (
	"fmt"
	"reflect"
	"strings"
)

const maxDebugDepth = 10

const maxRecursionDepth = 100

type pathEntry struct {
	depth int
	path  []string
}

// FindCircularReferences traverses an object graph to find any circular references
func FindCircularReferences(v any) bool {
	return findCircularReferencesInPath(v, make(map[uintptr]pathEntry), []string{"root"}, 0)
}

func findCircularReferencesInPath(v any, currentPath map[uintptr]pathEntry, path []string, depth int) bool {
	if v == nil {
		return false
	}

	if depth > maxRecursionDepth {
		fmt.Printf("Max recursion depth exceeded at depth %d\nPath: %v\n",
			depth, strings.Join(path, " -> "))
		return true
	}

	val := reflect.ValueOf(v)

	// Get the underlying value if it's an interface
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	// Handle pointers - get the actual value
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return false
		}
		ptr := val.Pointer()

		// If we see this pointer in our current path, it's a circular reference
		if entry, seen := currentPath[ptr]; seen {
			fmt.Printf("Found circular reference!\n  Current: depth %d, path: %v\n  Previous: depth %d, path: %v\n  Type: %v\n",
				depth, strings.Join(path, " -> "),
				entry.depth, strings.Join(entry.path, " -> "),
				val.Type())
			return true
		}

		// Add this pointer to our current path with current depth and path
		currentPath[ptr] = pathEntry{depth: depth, path: path}
		defer delete(currentPath, ptr) // Remove it when we're done with this branch

		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Map:
		for _, k := range val.MapKeys() {
			keyStr := fmt.Sprintf("%v", k.Interface())
			newPath := append([]string{}, path...)
			newPath = append(newPath, keyStr)
			if findCircularReferencesInPath(val.MapIndex(k).Interface(), currentPath, newPath, depth+1) {
				return true
			}
		}
	case reflect.Struct:
		t := val.Type()
		for i := 0; i < val.NumField(); i++ {
			if val.Field(i).CanInterface() {
				newPath := append([]string{}, path...)
				newPath = append(newPath, t.Field(i).Name)
				if findCircularReferencesInPath(val.Field(i).Interface(), currentPath, newPath, depth+1) {
					return true
				}
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			newPath := append([]string{}, path...)
			newPath = append(newPath, fmt.Sprintf("[%d]", i))
			if findCircularReferencesInPath(val.Index(i).Interface(), currentPath, newPath, depth+1) {
				return true
			}
		}
	}
	return false
}
