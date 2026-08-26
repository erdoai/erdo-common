package template

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

// captureLog collects everything the standard logger writes while fn runs, so a
// test can assert on what a lookup did or did not report.
func captureLog(t *testing.T, fn func()) string {
	t.Helper()

	var buf bytes.Buffer
	out, prefix, flags := log.Writer(), log.Prefix(), log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(out)
		log.SetPrefix(prefix)
		log.SetFlags(flags)
	})

	fn()
	return buf.String()
}

func TestLookupsReportOnlyMissesThatMatter(t *testing.T) {
	data := map[string]any{"present": "here"}

	t.Run("a required key that is missing is logged and recorded", func(t *testing.T) {
		var missing []string
		out := captureLog(t, func() {
			require.Nil(t, get("absent", data, &missing))
		})

		// This is the case worth keeping: the template expected a value and did
		// not get one, which is how "the agent ran without its context" is found.
		require.Contains(t, out, `key "absent" not found`)
		require.Equal(t, []string{"absent"}, missing)
	})

	t.Run("an optional key that is missing is silent", func(t *testing.T) {
		var missing []string
		out := captureLog(t, func() {
			require.Nil(t, get("absent?", data, &missing))
		})

		// The engine's contract is that an optional parameter is silently
		// substituted, so reporting the miss contradicts it.
		require.Empty(t, out)
		require.Empty(t, missing)
	})

	t.Run("a presence test on a missing key is silent", func(t *testing.T) {
		out := captureLog(t, func() {
			require.False(t, truthy("absent", data))
		})
		require.Empty(t, out)
	})

	t.Run("a presence test on a present key still answers", func(t *testing.T) {
		out := captureLog(t, func() {
			require.True(t, truthy("present", data))
		})
		require.Empty(t, out)
	})

	t.Run("a coalesce miss is silent and returns the fallback", func(t *testing.T) {
		var missing []string
		out := captureLog(t, func() {
			require.Equal(t, "fallback", coalesce("absent", "fallback", data, &missing))
		})

		// A miss is the whole reason coalesce was called.
		require.Empty(t, out)
		require.Empty(t, missing)
	})

	t.Run("nested misses follow the same rule", func(t *testing.T) {
		nested := map[string]any{"outer": map[string]any{"inner": "value"}}

		required := captureLog(t, func() {
			var missing []string
			require.Nil(t, get("outer.absent", nested, &missing))
		})
		require.Contains(t, required, `key "absent" not found`)

		optional := captureLog(t, func() {
			var missing []string
			require.Nil(t, get("outer.absent?", nested, &missing))
		})
		require.Empty(t, optional)
	})

	t.Run("a bad array index follows the same rule", func(t *testing.T) {
		listed := map[string]any{"items": []any{"first"}}

		required := captureLog(t, func() {
			var missing []string
			require.Nil(t, get("items.7", listed, &missing))
		})
		require.Contains(t, required, "invalid array index")

		optional := captureLog(t, func() {
			var missing []string
			require.Nil(t, get("items.7?", listed, &missing))
		})
		require.Empty(t, optional)
	})

	t.Run("nil data follows the same rule", func(t *testing.T) {
		required := captureLog(t, func() {
			var missing []string
			require.Nil(t, get("anything", nil, &missing))
		})
		require.Contains(t, required, "data is nil")

		optional := captureLog(t, func() {
			var missing []string
			require.Nil(t, get("anything?", nil, &missing))
		})
		require.Empty(t, optional)
	})

	t.Run("a present key is returned unchanged and reports nothing", func(t *testing.T) {
		var missing []string
		out := captureLog(t, func() {
			require.Equal(t, "here", get("present", data, &missing))
		})
		require.Empty(t, out)
		require.Empty(t, missing)
	})
}
