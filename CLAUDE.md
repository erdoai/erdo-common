# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build and Test
```bash
go build ./...        # Build all packages
go test ./...         # Run all tests
go test -v ./...      # Run tests with verbose output
go test ./template    # Run tests for template package only
go mod tidy           # Clean up module dependencies
```

## Architecture Overview

This is the common library for the Erdo AI system, providing shared types and utilities for building AI-powered automation bots.

### Core Components

**Types Package (`/types/`)**: Defines all core data structures for the Erdo system:
- Bot definitions with steps, parameters, and execution modes
- Resource management (tables, endpoints, documents, entities)
- Integration configurations for external services
- Dataset types for various data sources

**Template Engine (`/template/`)**: Sophisticated template processing system:
- `hydration.go` (1,751 lines): Core hydration logic that processes templates with parameter substitution
- Two categories of template functions:
  - Basic functions: Pure operations (UUID, JSON, string manipulation, comparison)
  - Data functions: Stateful operations requiring access to `.Data` and `.MissingKeys`
- Error tracking with `InfoNeededError` that preserves path information for missing parameters
- Supports optional parameters with `?` suffix and nested data access
- **Pointer-aware comparisons**: `eq` and `ne` functions automatically dereference pointers and treat `nil` pointers as empty strings for convenient template conditionals

### Key Design Patterns

**Parameter Hydration**: The system uses a multi-pass approach to template processing:
1. First pass identifies missing required parameters
2. Optional parameters (with `?`) are silently replaced with empty strings if missing
3. Missing keys are tracked with full path information for debugging
4. Functions automatically receive `.Data` and `.MissingKeys` parameters

**Bot Execution Model**: Steps can have:
- Dependencies on other steps
- Three execution modes: sequential, parallel, background
- Result handlers with conditional logic
- Different hydration behaviors: hydrate (default), raw, none

**Template Function Injection**: When calling template functions, the system automatically:
- Prepends `.Data` as the first argument for data functions
- Adds `.MissingKeys` parameter for functions that track missing keys
- Handles reserved Go template words by escaping them

### Testing Approach

Tests use the `testify` framework and are comprehensive, especially in `/template/`:
- Test files follow `*_test.go` naming convention
- Many tests use table-driven test patterns
- Template tests verify both successful hydration and error cases

## Important Guidelines

### Pointer Fields in Templates
Templates can directly access pointer fields from Go structs without dereferencing. The template system handles this automatically:

**Pointer Comparisons**: The `eq` and `ne` functions automatically dereference pointers:
```go
// Go struct
type Dataset struct {
    Name        *string  `json:"name"`
    Description *string  `json:"description"`
}

// Template usage - works directly with pointers
{{if ne .Data.dataset.Description ""}}
  Description: {{.Data.dataset.Description}}
{{end}}
```

**Key behaviors**:
- Pointers are automatically dereferenced for comparison
- `nil` pointers are treated as equivalent to empty strings (`""`)
- This allows natural template conditionals: `{{if ne $r.Dataset.Name ""}}` works whether `Name` is `nil` or points to an empty string

### JSON Struct Tags
**NEVER add `omitempty` to JSON struct tags in shared types** - pointer fields should use `*type` instead of `omitempty` for optional values. The template system works directly with Go struct fields (not JSON serialization), so:
- ✅ Use pointer types for optional fields: `Name *string`
- ❌ Don't use `omitempty` tags to make fields optional
- Templates access struct fields directly and handle pointers automatically