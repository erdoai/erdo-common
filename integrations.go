package erdotypes

import (
	"time"
)

// Integration Type Constants
// =========================

// IntegrationType represents the type of integration
type IntegrationType string

const (
	IntegrationTypeApi      IntegrationType = "api"
	IntegrationTypeDatabase IntegrationType = "database"
	IntegrationTypeFile     IntegrationType = "file"
)

// AuthType represents authentication methods
type AuthType string

const (
	AuthTypeOauth2    AuthType = "oauth2"
	AuthTypeApiKey    AuthType = "api_key"
	AuthTypeDatabase  AuthType = "database"
	AuthTypeBasicAuth AuthType = "basic_auth"
)

// IntegrationStatus represents the status of an integration configuration
type IntegrationStatus string

const (
	IntegrationStatusActive      IntegrationStatus = "active"
	IntegrationStatusInactive    IntegrationStatus = "inactive"
	IntegrationStatusBeta        IntegrationStatus = "beta"
	IntegrationStatusComingSoon  IntegrationStatus = "coming_soon"
	IntegrationStatusError       IntegrationStatus = "error"
	IntegrationStatusNeedsReauth IntegrationStatus = "needs_reauth"
)

// SegmentSelectionType represents how segments can be selected
type SegmentSelectionType string

const (
	SegmentSelectionTypeSingle   SegmentSelectionType = "single"
	SegmentSelectionTypeMultiple SegmentSelectionType = "multiple"
	SegmentSelectionTypeRequired SegmentSelectionType = "required"
	SegmentSelectionTypeOptional SegmentSelectionType = "optional"
)

// ExpiryType represents how token expiry is calculated
type ExpiryType string

const (
	ExpiryTypeOAuthDefault ExpiryType = "oauth_default"
	ExpiryTypeConstant     ExpiryType = "constant"
	ExpiryTypeOAuthField   ExpiryType = "oauth_field"
)

// Integration Configuration Types
// ==============================

// UIConfigIcon represents icon configuration for integrations
type UIConfigIcon struct {
	Set  string `json:"set"`
	Name string `json:"name"`
}

// UIConfig represents UI configuration for integration display
type UIConfig struct {
	BrandLogoIcon UIConfigIcon `json:"brand_logo_icon"`
	BrandColor    string       `json:"brand_color"`
	ButtonStyle   string       `json:"button_style"`
}

// ErrorHandlingConfig defines how to handle errors in API calls
type ErrorHandlingConfig struct {
	IgnoreErrors []string `json:"ignore_errors,omitempty"` // List of error codes/messages to ignore
	ErrorPath    string   `json:"error_path,omitempty"`    // JSONPath to error code/message in response
}

// SegmentLevel defines how to fetch and parse each level of segments
type SegmentLevel struct {
	Name                string   `json:"name"`                           // e.g., "account", "property"
	Type                string   `json:"type"`                           // e.g., "project", "campaign"
	Selectable          bool     `json:"selectable"`                     // Whether this level can be selected
	URLTemplate         string   `json:"url_template,omitempty"`         // Template for API call
	Method              string   `json:"method,omitempty"`               // HTTP method (GET by default)
	Body                string   `json:"body,omitempty"`                 // Request body for POST/PUT requests
	IDPath              string   `json:"id_path,omitempty"`              // JSONPath to ID in response
	IDRegex             string   `json:"id_regex,omitempty"`             // Regex to extract ID from IDPath
	NamePath            string   `json:"name_path,omitempty"`            // JSONPath to name in response
	ParentKey           string   `json:"parent_key,omitempty"`           // How to reference parent in API call
	RequiredCredentials []string `json:"required_credentials,omitempty"` // Required credentials for this level
}

// SegmentConfig combines selection rules with API configuration
type SegmentConfig struct {
	SelectionType SegmentSelectionType `json:"selection_type"`
	MinSelections *int                 `json:"min_selections,omitempty"` // Only for MultiSelect
	MaxSelections *int                 `json:"max_selections,omitempty"` // Only for MultiSelect
	Description   string               `json:"description"`              // Explains what segments represent
	Hierarchical  bool                 `json:"hierarchical"`             // Whether segments have parent/child relationships
	BaseURL       string               `json:"base_url,omitempty"`       // Base URL for API calls
	Levels        []SegmentLevel       `json:"levels"`                   // Ordered segment hierarchy
	ErrorHandling *ErrorHandlingConfig `json:"error_handling,omitempty"`
}

// ResourceTypeConfig defines how to fetch a specific type of resource
type ResourceTypeConfig struct {
	Type            string `json:"type"`                       // e.g., "sheet", "table", "range"
	URLTemplate     string `json:"url_template"`               // Template for API call to get resources
	Method          string `json:"method,omitempty"`           // HTTP method (GET by default)
	Body            string `json:"body,omitempty"`             // Request body template
	IDPath          string `json:"id_path"`                    // JSONPath to ID in response
	NamePath        string `json:"name_path"`                  // JSONPath to name in response
	DescriptionPath string `json:"description_path,omitempty"` // JSONPath to description
}

// DatasetResourceDiscoveryConfig defines how to discover resources for a dataset integration
type DatasetResourceDiscoveryConfig struct {
	MinResourcesRequired int `json:"min_resources_required"` // Minimum number of resources required for the dataset to be saved

	// For databases
	TableQuery        string `json:"table_query,omitempty"`        // SQL to fetch tables
	RelationshipQuery string `json:"relationship_query,omitempty"` // SQL to fetch relationships

	// Additional metadata queries for enhanced analysis
	IndexQuery      string `json:"index_query,omitempty"`      // SQL to fetch table indexes
	ConstraintQuery string `json:"constraint_query,omitempty"` // SQL to fetch table constraints
	StatisticsQuery string `json:"statistics_query,omitempty"` // SQL to fetch table statistics
	SizeQuery       string `json:"size_query,omitempty"`       // SQL to fetch table sizes

	// For APIs
	BaseURL       string               `json:"base_url,omitempty"`
	ResourceTypes []ResourceTypeConfig `json:"resource_types,omitempty"`
	ErrorHandling *ErrorHandlingConfig `json:"error_handling,omitempty"`
}

// ExpiryConfig defines token expiry configuration
type ExpiryConfig struct {
	Type ExpiryType `json:"type"`
	// For constant type
	Duration time.Duration `json:"duration,omitempty"`
	// For oauth_field type
	FieldName string `json:"field_name,omitempty"`
	// For refresh tokens
	RefreshTokenDuration *time.Duration `json:"refresh_token_duration,omitempty"`
}

// SegmentOption represents a segment option for a provider
type SegmentOption struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`      // e.g., "project", "campaign", "account"
	ParentID *string           `json:"parent_id"` // for hierarchy
	Metadata map[string]string `json:"metadata"`  // additional provider-specific info
	Children []SegmentOption   `json:"children"`  // nested segments
}

// IntegrationDefinition represents a complete integration definition for Python SDK
type IntegrationDefinition struct {
	// Basic configuration
	Name                    string            `json:"name"`
	Key                     string            `json:"key"`
	Type                    IntegrationType   `json:"type"`
	AuthTypes               []AuthType        `json:"auth_types"`
	Status                  IntegrationStatus `json:"status"`
	Description             *string           `json:"description,omitempty"`
	ProviderName            *string           `json:"provider_name,omitempty"`
	DocumentationUrl        *string           `json:"documentation_url,omitempty"`
	OpenapiDocumentationUrl *string           `json:"openapi_documentation_url,omitempty"`
	HealthcheckUrl          *string           `json:"healthcheck_url,omitempty"`

	// OAuth-specific fields
	AuthUrl         *string  `json:"auth_url,omitempty"`
	TokenUrl        *string  `json:"token_url,omitempty"`
	ClientID        *string  `json:"client_id,omitempty"`
	ClientSecret    *string  `json:"client_secret,omitempty"`
	AvailableScopes []string `json:"available_scopes,omitempty"`
	OptionalScopes  []string `json:"optional_scopes,omitempty"`
	ScopeSeparator  *string  `json:"scope_separator,omitempty"`

	// API version info
	ApiVersion            *string `json:"api_version,omitempty"`
	ApiVersionDescription *string `json:"api_version_description,omitempty"`

	// Configuration objects
	CredentialSchema               map[string]CredentialSchema     `json:"credential_schema,omitempty"`
	UIConfig                       *UIConfig                       `json:"ui_config,omitempty"`
	CodegenDetails                 *CodegenDetails                 `json:"codegen_details,omitempty"`
	AnalysisDetails                *AnalysisDetails                `json:"analysis_details,omitempty"`
	SegmentConfig                  *SegmentConfig                  `json:"segment_config,omitempty"`
	DatasetResourceDiscoveryConfig *DatasetResourceDiscoveryConfig `json:"dataset_resource_discovery_config,omitempty"`
	ExpiryConfig                   *ExpiryConfig                   `json:"expiry_config,omitempty"`

	// Provider credentials (for app-level secrets)
	ProviderCredentials map[string]interface{} `json:"provider_credentials,omitempty"`

	// Resource configuration
	DefaultResourceAttachType *string `json:"default_resource_attach_type,omitempty"`
}

// Python SDK Integration Types
// ============================

type IntegrationConfig struct {
	Definition IntegrationDefinition `json:"definition"`
	Source     string                `json:"source"` // "python" to indicate Python-defined
	FilePath   string                `json:"file_path,omitempty"`
}

type PythonIntegrationInstance struct {
	Config      IntegrationConfig      `json:"config"`
	Credentials map[string]interface{} `json:"credentials"`
	Methods     []string               `json:"methods,omitempty"` // Available methods on the integration
}

// Integration Discovery Types (for CLI introspection)
// ===================================================

// IntegrationDiscovery represents a discovered integration from Python source code
type IntegrationDiscovery struct {
	Config     IntegrationConfig          `json:"config"`
	Instance   *PythonIntegrationInstance `json:"instance,omitempty"`
	FilePath   string                     `json:"file_path"`
	SourceCode string                     `json:"source_code"`
}

// Request/Response Types for Integration Management
// ================================================

// UpsertIntegrationRequest represents a request to create or update an integration
type UpsertIntegrationRequest struct {
	Integration IntegrationDefinition `json:"integration"`
	Source      string                `json:"source"`
}

// UpsertIntegrationResponse represents the response from upserting an integration
type UpsertIntegrationResponse struct {
	IntegrationID string `json:"integration_id"`
	Status        string `json:"status"`
}

// ListIntegrationsResponse represents a response containing multiple integrations
type ListIntegrationsResponse struct {
	Integrations []IntegrationDefinition `json:"integrations"`
}

// ExportIntegrationsResponse represents the response from the integrations export endpoint
type ExportIntegrationsResponse struct {
	Integrations map[string]IntegrationDefinition `json:"integrations"`
}
