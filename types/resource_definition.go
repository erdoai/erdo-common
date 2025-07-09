package types

import (
	"time"
)

// Resource represents a resource definition in the system
type Resource struct {
	ID                  string                 `json:"id"`
	Key                 string                 `json:"key"`
	Name                string                 `json:"name"`
	Type                ResourceType           `json:"type"`
	Description         *string                `json:"description,omitempty"`
	UseCases            []string               `json:"use_cases,omitempty"`
	Tags                []string               `json:"tags,omitempty"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
	IntegrationConfigID string                 `json:"integration_config_id"`
	State               ResourceState          `json:"state"`
	AttachType          ResourceAttachType     `json:"attach_type"`
	DatasetID           string                 `json:"dataset_id"`
	Instructions        *string                `json:"instructions,omitempty"`
}