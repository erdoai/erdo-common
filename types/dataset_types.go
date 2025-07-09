package types

import (
	"time"

	"github.com/google/uuid"
)

// DatasetType represents the type of dataset
type DatasetType string

const (
	DatasetTypeFile        DatasetType = "file"
	DatasetTypeDatabase    DatasetType = "database"
	DatasetTypeIntegration DatasetType = "integration"
)

// Dataset represents a data source that can be analyzed
type Dataset struct {
	ID   uuid.UUID   `json:"id"`
	Type DatasetType `json:"type"`
	Key  *string     `json:"key"`

	Name        *string `json:"name"`
	Description *string `json:"description"`

	AnalysisSummary *string    `json:"analysis_summary"`
	LastAnalyzed    *time.Time `json:"last_analyzed"`

	Instructions *string `json:"instructions"`

	// File dataset fields
	File     *string `json:"file"`
	FileType *string `json:"file_type"`
	Filename *string `json:"filename"`
	URL      *string `json:"url"`

	// Integration dataset fields
	IntegrationID                   *uuid.UUID                  `json:"integration_id"`
	IntegrationConfigID             *uuid.UUID                  `json:"integration_config_id"`
	IntegrationConfig               *IntegrationConfig          `json:"integration_config"`
	EncryptedIntegrationCredentials *map[string]string          `json:"encrypted_integration_credentials"`
	CredentialSchema                *map[string]CredentialSchema `json:"credential_schema"`
	CodegenDetails                  *CodegenDetails             `json:"codegen_details"`
	AnalysisDetails                 *AnalysisDetails            `json:"analysis_details"`
	AvailableScopes                 *[]string                   `json:"available_scopes"`
	
	// Parameters for code execution
	Parameters *map[string]interface{} `json:"parameters"`
}

// BotResource represents a data resource with its dataset for bot invocations
type BotResource struct {
	ID        int      `json:"id"` // scoped to the invocation
	Dataset   *Dataset `json:"dataset"`
	CreatedBy string   `json:"created_by"`
	// Any additional data that's useful but not present
	// on the resource type
	Extra     map[string]any     `json:"extra"`
}