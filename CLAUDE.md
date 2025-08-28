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
  - Basic functions: Pure operations (UUID, JSON, string manipulation)
  - Data functions: Stateful operations requiring access to `.Data` and `.MissingKeys`
- Error tracking with `InfoNeededError` that preserves path information for missing parameters
- Supports optional parameters with `?` suffix and nested data access

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

### JSON Struct Tags
**NEVER add `omitempty` to JSON struct tags in shared types** that may be used in templates. The `omitempty` tag breaks template accessors because it causes fields to be excluded from JSON marshaling when they have zero values, making them inaccessible in templates even when explicitly referenced.