package types

// Template function definitions for use in bot exports and state references
// These should match the functions available in backend/utils/parameters/parameters.go

// BasicTemplateFunctions are functions that don't require .Data and .MissingKeys
var BasicTemplateFunctions = []string{
	"truthy",
	"toJSON",
	"len",
	"add",
	"sub",
	"gt",
	"lt",
	"mergeRaw",
	"nilToEmptyString",
	"truthyValue",
	"toString",
	"truncateString",
	"regexReplace",
	"noop",
	"list",
}

// DataTemplateFunctions are functions that require .Data and .MissingKeys parameters
var DataTemplateFunctions = []string{
	"get",
	"concat",
	"getOrOriginal",
	"sliceEnd",
	"sliceEndKeepFirstUserMessage",
	"slice",
	"extractSlice",
	"dedupeBy",
	"find",
	"findByValue",
	"getAtIndex",
	"merge",
	"coalescelist",
	"addkey",
	"removekey",
	"mapToDict",
	"addkeytoall",
	"incrementCounter",
	"incrementCounterBy",
	"coalesce",
	"filter",
}

// AllTemplateFunctions combines basic and data template functions
var AllTemplateFunctions = append(BasicTemplateFunctions, DataTemplateFunctions...)
