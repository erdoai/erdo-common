package types

import (
	"time"
)

// ResourceAnalysis represents analysis information for a resource
type ResourceAnalysis struct {
	Summary      *string              `json:"summary"`
	LastAnalyzed *time.Time           `json:"last_analyzed"`
	EntityKey    string               `json:"entity_key"`
	ColumnRanges []ColumnDateRange    `json:"column_ranges,omitempty"`
}

// ColumnDateRange carries the observed min/max for a single date or numeric column,
// derived from ColumnStats during analysis. Agents render these so the LLM sees the
// actual extent of each column's data without probing with min/max queries.
type ColumnDateRange struct {
	Column  string   `json:"column"`
	MinDate string   `json:"min_date,omitempty"`
	MaxDate string   `json:"max_date,omitempty"`
	Min     *float64 `json:"min,omitempty"`
	Max     *float64 `json:"max,omitempty"`
}

// ResourceField represents a single field/column in a resource, carrying its type and semantic description.
type ResourceField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// Resource represents a resource definition in the system
type Resource struct {
	ID                  string             `json:"id"`
	Key                 string             `json:"key"`
	Name                string             `json:"name"`
	Type                ResourceType       `json:"type"`
	Description         *string            `json:"description"`
	UseCases            []string           `json:"use_cases"`
	Tags                []string           `json:"tags"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
	IntegrationConfigID string             `json:"integration_config_id"`
	State               ResourceState      `json:"state"`
	AttachType          ResourceAttachType `json:"attach_type"`
	DatasetID           string             `json:"dataset_id"`
	Instructions        *string            `json:"instructions"`
	Analyses            []ResourceAnalysis `json:"analyses"`
	RelatedResources    []Resource         `json:"related_resources"`
	Metadata            map[string]any     `json:"metadata"` // Structured metadata (e.g., sheet_name, table_index for Excel/Sheets)
	Fields              []ResourceField    `json:"fields"`   // Column/field definitions with types and descriptions
}
