package types

import (
	"time"
)

// ResourceAnalysis represents analysis information for a resource
type ResourceAnalysis struct {
	Summary      *string    `json:"summary"`
	LastAnalyzed *time.Time `json:"last_analyzed"`
	EntityKey    string     `json:"entity_key"`
}

// ResourceField represents a single field/column in a resource, carrying its type and semantic description.
//
// Min/Max and MinDate/MaxDate are optional range signals derived from ColumnStats on the
// most recent analysis. For numeric columns, Min/Max are populated. For date/timestamp
// columns, MinDate/MaxDate are populated as ISO strings. Agents surface these to the LLM
// so it can see the actual extent of the data (e.g. "max date is 2025-10-15") without
// having to probe with min/max queries first.
type ResourceField struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Min         *float64 `json:"min,omitempty"`
	Max         *float64 `json:"max,omitempty"`
	MinDate     string   `json:"min_date,omitempty"`
	MaxDate     string   `json:"max_date,omitempty"`
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
