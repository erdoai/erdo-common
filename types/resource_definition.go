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
}
