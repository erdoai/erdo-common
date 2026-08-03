package types

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
	BrandLogoIcon      UIConfigIcon `json:"brand_logo_icon"`
	BrandColor         string       `json:"brand_color"`
	BrandColorDarkMode string       `json:"brand_color_dark_mode"`
	ButtonStyle        string       `json:"button_style"`
	Category           string       `json:"category"`
}

// ErrorHandlingConfig defines how to handle errors in API calls
type ErrorHandlingConfig struct {
	IgnoreErrors []string `json:"ignore_errors,omitempty"` // List of error codes/messages to ignore
	ErrorPath    string   `json:"error_path,omitempty"`    // JSONPath to error code/message in response
}

// SegmentLevel defines how to fetch and parse each level of segments
type SegmentLevel struct {
	Name       string `json:"name"`       // e.g., "account", "property"
	Type       string `json:"type"`       // e.g., "project", "campaign"
	Selectable bool   `json:"selectable"` // Whether this level can be selected
	Required   bool   `json:"required"`   // Must user explicitly select from this level? (false = optional, all by default)

	// ConnectionScope marks a level as choosing WHICH PROVIDER ACCOUNT the
	// connection operates, rather than filtering what one dataset reads from
	// it. An advertising OAuth grant reaches every account the authenticating
	// login can see, so "which of them is this connection for" is a property of
	// the connection: it decides the account for every dataset, sync, and agent
	// action that connection serves. Levels below it (campaigns, lists, flows)
	// stay per-dataset filters.
	//
	// Connection-scope selections are stored on the integration and are asked
	// for when connecting; the segment picker shown while configuring a dataset
	// offers only the levels below them.
	ConnectionScope bool `json:"connection_scope,omitempty"`

	URLTemplate         string   `json:"url_template,omitempty"`         // Template for API call
	Method              string   `json:"method,omitempty"`               // HTTP method (GET by default)
	Body                string   `json:"body,omitempty"`                 // Request body for POST/PUT requests
	IDPath              string   `json:"id_path,omitempty"`              // JSONPath to ID in response
	IDRegex             string   `json:"id_regex,omitempty"`             // Regex to extract ID from IDPath
	NamePath            string   `json:"name_path,omitempty"`            // JSONPath to name in response
	ParentKey           string   `json:"parent_key,omitempty"`           // How to reference parent in API call
	RequiredCredentials []string `json:"required_credentials,omitempty"` // Required credentials for this level

	// GroupBy* synthesize a parent level by grouping this level's segments on a
	// field of each response item (e.g. Meta ad accounts grouped by their owning
	// business). This avoids a separate parent-level fetch when the parent entity
	// only appears embedded in the child response — and keeps working when the
	// token cannot list the parents directly. Items missing the group field stay
	// ungrouped at this level's position in the tree.
	GroupByLevelName string `json:"group_by_level_name,omitempty"` // Type/name for synthetic parent segments, e.g. "business"
	GroupByItemsPath string `json:"group_by_items_path,omitempty"` // JSONPath to the response items array, e.g. "$.data[*]"
	GroupByIDPath    string `json:"group_by_id_path,omitempty"`    // JSONPath to the group ID relative to an item, e.g. "$.business.id"
	GroupByNamePath  string `json:"group_by_name_path,omitempty"`  // JSONPath to the group name relative to an item, e.g. "$.business.name"
	// GroupByUngroupedName, when set, buckets items missing the group field under
	// a single synthetic parent with this display name (type "group", id
	// "__ungrouped__") instead of leaving them ungrouped at this level — keeping
	// the tree a uniform depth for consumers that render strict two-level trees.
	GroupByUngroupedName string `json:"group_by_ungrouped_name,omitempty"`

	// Name enrichment - for APIs where the initial call only returns IDs and a secondary
	// call is needed to get display names (e.g., Google Ads listAccessibleCustomers).
	// The provider first tries batch enrichment (if configured), then falls back to individual calls.
	// URL templates have access to {{.id}} (the segment ID).
	//
	// Batch enrichment (tried first): Query each segment as a potential "manager" that can return
	// names for multiple child segments. For Google Ads, manager accounts can query customer_client
	// to get all linked account names in one call.
	NameEnrichmentBatchBody     string `json:"name_enrichment_batch_body,omitempty"`      // Batch query body (e.g., customer_client query)
	NameEnrichmentBatchIDPath   string `json:"name_enrichment_batch_id_path,omitempty"`   // JSONPath to IDs in batch response
	NameEnrichmentBatchNamePath string `json:"name_enrichment_batch_name_path,omitempty"` // JSONPath to names in batch response
	//
	// Individual enrichment (fallback): Called for segments not found in batch responses.
	NameEnrichmentURLTemplate string `json:"name_enrichment_url_template,omitempty"` // URL to fetch name
	NameEnrichmentMethod      string `json:"name_enrichment_method,omitempty"`       // HTTP method (GET by default)
	NameEnrichmentBody        string `json:"name_enrichment_body,omitempty"`         // Request body for individual enrichment
	NameEnrichmentPath        string `json:"name_enrichment_path,omitempty"`         // JSONPath to name in enrichment response
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
	LlmsTxtUrl              *string           `json:"llms_txt_url,omitempty"`
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
