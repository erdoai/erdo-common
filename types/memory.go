package types

import (
	"time"
)

// Memory represents a stored memory item in the system
type Memory struct {
	ID                  string            `json:"id"`
	Content             string            `json:"content"`
	Description         string            `json:"description"`
	Tags                []string          `json:"tags,omitempty"`
	State               string            `json:"state"`
	EstimatedStaleAt    *time.Time        `json:"estimated_stale_at,omitempty"`
	StaleWhenText       *string           `json:"stale_when_text,omitempty"`
	CreatedByEntityType string            `json:"created_by_entity_type"`
	CreatedByID         string            `json:"created_by_id"`
	CreatedFrom         *string           `json:"created_from,omitempty"`
	OrganizationID      *string           `json:"organization_id,omitempty"`
	Extra               map[string]interface{} `json:"extra,omitempty"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
	Type                string            `json:"type"`
	DatasetID           *string           `json:"dataset_id,omitempty"`
	IntegrationConfigID *string           `json:"integration_config_id,omitempty"`
	UserID              *string           `json:"user_id,omitempty"`
	ThreadID            *string           `json:"thread_id,omitempty"`
	CurrentVersion      int32             `json:"current_version"`
	ApprovalStatus      string            `json:"approval_status"`
}