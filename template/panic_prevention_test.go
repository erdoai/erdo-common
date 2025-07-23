package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPreventReflectPanicWithNilValue specifically tests the fix for the panic:
// "panic: reflect: Call using zero Value argument"
// This was happening when a nested function returned nil and we tried to call
// reflect.ValueOf(nil).Call() on it.
func TestPreventReflectPanicWithNilValue(t *testing.T) {
	t.Parallel()

	// Test the original scenario that was causing issues
	t.Run("Original GreaterThan scenario with missing field", func(t *testing.T) {
		// This simulates the template: {{len(data.similar_memories)}} > 0
		// where similar_memories doesn't exist
		template := "{{gt (len (get \"similar_memories\")) 0}}"
		data := map[string]any{
			// similar_memories is missing
		}

		// Should return InfoNeededError, not panic
		_, err := Hydrate(template, &data, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "info needed for keys")
	})

	// Test that nil values in data are handled correctly
	t.Run("Explicit nil value in data", func(t *testing.T) {
		template := "{{len (get \"null_field\")}}"
		data := map[string]any{
			"null_field": nil,
		}

		// len should handle nil gracefully and return 0
		result, err := Hydrate(template, &data, nil)
		assert.NoError(t, err)
		assert.Equal(t, 0, result)
	})

	// Test that the fix prevents panic in nested function calls
	t.Run("Nested function with nil from mapToDict", func(t *testing.T) {
		// mapToDict returns empty array for missing keys
		template := "{{len (mapToDict \"missing_key\" \"id\")}}"
		data := map[string]any{}

		result, err := Hydrate(template, &data, nil)
		assert.NoError(t, err)
		assert.Equal(t, 0, result) // len of empty array is 0
	})

	// Test complex nested function that used to panic
	t.Run("Complex nested function with toJSON", func(t *testing.T) {
		template := "{{toJSON (get \"missing_field\")}}"
		data := map[string]any{}

		// This should return an InfoNeededError for the missing field
		_, err := Hydrate(template, &data, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "info needed for keys")
	})
}