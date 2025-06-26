package erdotypes

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Bot represents a bot for export/import across CLI and backend
type Bot struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Code            string    `json:"code"`
	FilePath        string    `json:"file_path"`
	Persona         *string   `json:"persona"`
	RunningMessage  *string   `json:"running_message"`
	FinishedMessage *string   `json:"finished_message"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	OrganizationID  string    `json:"organization_id"`
	Visibility      string    `json:"visibility"`
	Source          string    `json:"source"`
}

// Step represents a bot step for export/import across CLI and backend
type Step struct {
	ID                          string          `json:"id"`
	BotID                       string          `json:"bot_id"`
	ActionType                  string          `json:"action_type"`
	Parameters                  json.RawMessage `json:"parameters"`
	DependsOn                   *[]string       `json:"depends_on,omitempty"` // Able to be nil (so dependencies can be automatically resolved by the API), or empty (so that the deps are explicitly empty)
	Key                         *string         `json:"key,omitempty"`
	StepOrder                   int32           `json:"step_order"`
	OutputContentType           string          `json:"output_content_type"`
	UserOutputVisibility        string          `json:"user_output_visibility"`
	BotOutputVisibility         string          `json:"bot_output_visibility"`
	ExecutionMode               json.RawMessage `json:"execution_mode"`
	OutputBehaviour             json.RawMessage `json:"output_behaviour"`
	OutputChannels              json.RawMessage `json:"output_channels"`
	RunningMessage              *string         `json:"running_message,omitempty"`
	FinishedMessage             *string         `json:"finished_message,omitempty"`
	HistoryContentType          *string         `json:"history_content_type,omitempty"`
	UiContentType               *string         `json:"ui_content_type,omitempty"`
	ParameterHydrationBehaviour json.RawMessage `json:"parameter_hydration_behaviour"`
	ResultHandlerID             *string         `json:"result_handler_id,omitempty"`
	CreatedAt                   time.Time       `json:"created_at"`
	UpdatedAt                   time.Time       `json:"updated_at"`
}

// Parameter Types
// ===============

// ParameterType represents the type of a parameter
type ParameterType string

const (
	ParameterTypeString  ParameterType = "string"
	ParameterTypeInteger ParameterType = "integer"
	ParameterTypeFloat   ParameterType = "float"
	ParameterTypeBoolean ParameterType = "bool"
	ParameterTypeJson    ParameterType = "json"
)

// ParameterHydrationBehaviour represents how parameters should be hydrated
type ParameterHydrationBehaviour string

const (
	ParameterHydrationBehaviourHydrate ParameterHydrationBehaviour = "hydrate"
	ParameterHydrationBehaviourRaw     ParameterHydrationBehaviour = "raw"
	ParameterHydrationBehaviourNone    ParameterHydrationBehaviour = "none"
)

// OutputVisibility represents visibility levels for any type of output
type OutputVisibility string

const (
	OutputVisibilityVisible OutputVisibility = "visible"
	OutputVisibilityHidden  OutputVisibility = "hidden"
)

// Aliases for backward compatibility and semantic clarity
type OutputContentVisibility = OutputVisibility
type UserOutputVisibility = OutputVisibility
type BotOutputVisibility = OutputVisibility

// OutputContentType represents the type of output content
type OutputContentType string

const (
	OutputContentTypeText OutputContentType = "text"
	OutputContentTypeJSON OutputContentType = "json"
	OutputContentTypeHTML OutputContentType = "html"
)

// HandlerType represents the type of result handler
type HandlerType string

const (
	HandlerTypeIntermediate HandlerType = "intermediate"
	HandlerTypeFinal        HandlerType = "final"
)

// ExecutionModeType represents the execution mode for steps
type ExecutionModeType string

const (
	ExecutionModeTypeAll           ExecutionModeType = "all"
	ExecutionModeTypeIterateOver   ExecutionModeType = "iterate_over"
	ExecutionModeTypeAllBackground ExecutionModeType = "all_background"
)

// Model represents LLM model types
type Model string

const (
	// Anthropic models
	ModelClaude3Dot7Sonnet Model = "claude-3-7-sonnet-20250219"
	ModelClaude4Sonnet     Model = "claude-sonnet-4-20250514"
	ModelClaude3Dot5Haiku  Model = "claude-3-5-haiku-20241022"

	// OpenAI models
	ModelGPT4o        Model = "gpt-4o"
	ModelGPT4oMini    Model = "gpt-4o-mini"
	ModelGPT4Dot1     Model = "gpt-4.1"
	ModelGPT4Dot1Mini Model = "gpt-4.1-mini"
	ModelGPT4Dot1Nano Model = "gpt-4.1-nano"
)

// ParameterDefinition represents a parameter definition (shared across CLI and backend)
type ParameterDefinition struct {
	ID          uuid.UUID     `json:"id"`
	BotID       uuid.UUID     `json:"bot_id"`
	Name        string        `json:"name"`
	Key         string        `json:"key"`
	Description *string       `json:"description"`
	Type        ParameterType `json:"type"`
	IsRequired  bool          `json:"is_required"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	// Extended fields for agent discovery
	ValueSources []ParameterValueSource `json:"value_sources,omitempty"`
	Interpreters []ParameterInterpreter `json:"interpreters,omitempty"`
}

// ParameterValueSourceType represents the type of parameter value source
type ParameterValueSourceType string

// ParameterValueSource represents a source for parameter values
type ParameterValueSource struct {
	ID                    uuid.UUID                `json:"id"`
	ParameterDefinitionID uuid.UUID                `json:"parameter_definition_id"`
	Type                  ParameterValueSourceType `json:"type"`
	Parameters            json.RawMessage          `json:"parameters"`
	CreatedAt             time.Time                `json:"created_at"`
	UpdatedAt             time.Time                `json:"updated_at"`
	// Extended fields for agent discovery
	OnPopulate []ParameterValueSourceHandler `json:"on_populate,omitempty"`
}

// ParameterValueSourceHandler represents a handler for parameter value source events
type ParameterValueSourceHandler struct {
	ID                     uuid.UUID       `json:"id"`
	ParameterValueSourceID uuid.UUID       `json:"parameter_value_source_id"`
	ActionType             string          `json:"action_type"`
	Parameters             json.RawMessage `json:"parameters"`
	ExecutionMode          string          `json:"execution_mode"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
}

// ParameterInterpreter represents a parameter interpreter
type ParameterInterpreter struct {
	ID                    uuid.UUID       `json:"id"`
	ParameterDefinitionID uuid.UUID       `json:"parameter_definition_id"`
	ActionType            string          `json:"action_type"`
	Parameters            json.RawMessage `json:"parameters"`
	InterpreterOrder      int32           `json:"interpreter_order"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

// Agent Discovery Types (for CLI introspection)
// =============================================

// AgentDiscovery represents a discovered agent from Python source code
type AgentDiscovery struct {
	Bot                  Bot                   `json:"bot"`
	ParameterDefinitions []ParameterDefinition `json:"parameter_definitions"`
	Steps                []Step                `json:"steps"`
	FilePath             string                `json:"file_path"`
	SourceCode           string                `json:"source_code"`
}

type ExecutionMode struct {
	Mode        string               `json:"mode"`
	Data        json.RawMessage      `json:"data"`
	IfCondition *ConditionDefinition `json:"if_condition"`
}

type OutputBehavior json.RawMessage

// API Request Types
// ================

// UpsertBotRequest represents a request to create or update a bot
type UpsertBotRequest struct {
	Bot    Bot                `json:"bot"`
	Steps  []StepWithHandlers `json:"steps"`
	Source string             `json:"source"`
}

// Response Types
// ==============

// BotsResponse represents a response containing multiple bots
type BotsResponse struct {
	Bots []Bot `json:"bots"`
}

type StepsResponse struct {
	Steps []Step `json:"steps"`
}

// Service and Integration Types (for CLI and backend)
// ==================================================

// ServiceDefinition represents a service with its actions
type ServiceDefinition struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Actions     []ActionDefinition `json:"actions"`
}

// ActionDefinition represents an action within a service
type ActionDefinition struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Parameters  []ParameterDefinition `json:"parameters"`
}

// IntegrationSchema represents an integration configuration
type IntegrationSchema struct {
	Key              string                      `json:"key"`
	Name             string                      `json:"name"`
	Description      string                      `json:"description"`
	Type             string                      `json:"type"`
	AuthTypes        []string                    `json:"auth_types"`
	CredentialSchema map[string]CredentialSchema `json:"credential_schema"`
	AvailableScopes  []string                    `json:"available_scopes"`
	DocumentationUrl string                      `json:"documentation_url"`
	CodegenDetails   *CodegenDetails             `json:"codegen_details,omitempty"`
	AnalysisDetails  *AnalysisDetails            `json:"analysis_details,omitempty"`
}

// CodegenDetails represents code generation details for integrations
type CodegenDetails struct {
	Code    string   `json:"code"`
	Imports []string `json:"imports"`
	Hint    string   `json:"hint,omitempty"`
}

// AnalysisDetails represents analysis details for integrations
type AnalysisDetails struct {
	Imports []string `json:"imports"`
	Code    string   `json:"code"`
}

// CredentialSource defines where a credential value comes from
type CredentialSource string

const (
	// CredentialSourceIntegrationCredentials indicates the credential comes from the integration's stored credentials
	CredentialSourceIntegrationCredentials CredentialSource = "integration_credentials"
	// CredentialSourceConfigProviderCredentials indicates the credential comes from the integration config's provider credentials
	CredentialSourceConfigProviderCredentials CredentialSource = "config_provider_credentials"
	// CredentialSourceSegment indicates the credential comes from selected segments
	CredentialSourceSegment CredentialSource = "segment"
	// CredentialSourceDatasetParameters indicates the credential comes from dataset parameters
	CredentialSourceDatasetParameters CredentialSource = "dataset_parameters"
)

// SensitivityLevel defines how sensitive a credential is, determining who can view it
type SensitivityLevel string

const (
	// SensitivityLevelNeverViewable indicates credentials that should never be returned to clients
	SensitivityLevelNeverViewable SensitivityLevel = "never_viewable"
	// SensitivityLevelOwnerViewable indicates credentials that can be shown to the owner of the integration
	SensitivityLevelOwnerViewable SensitivityLevel = "owner_viewable"
	// SensitivityLevelAdminViewable indicates credentials that can be shown to admin users only
	SensitivityLevelAdminViewable SensitivityLevel = "admin_viewable"
	// SensitivityLevelEditViewable indicates credentials that can be shown to users with edit or higher access
	SensitivityLevelEditViewable SensitivityLevel = "edit_viewable"
	// SensitivityLevelAllViewable indicates credentials that can be shown to any user with view access
	SensitivityLevelAllViewable SensitivityLevel = "all_viewable"
)

// CredentialSchema defines the schema for a credential field (enhanced version of CredentialField)
type CredentialSchema struct {
	Type        string           `json:"type"`
	Description string           `json:"description"`
	Required    bool             `json:"required"`
	Source      CredentialSource `json:"source"`
	Sensitivity SensitivityLevel `json:"sensitivity,omitempty"`
	JQ          string           `json:"jq,omitempty"`
	Header      string           `json:"header,omitempty"`
}

// ExportActionsResponse represents the response from the actions export endpoint
type ExportActionsResponse struct {
	Services     map[string]ServiceDefinition `json:"services"`
	Integrations map[string]IntegrationSchema `json:"integrations"`
}

// ConditionDefinition represents a conditional expression
type ConditionDefinition struct {
	Type       string          `json:"type"`
	Conditions json.RawMessage `json:"conditions,omitempty"`
	Leaf       json.RawMessage `json:"leaf,omitempty"`
}

type StepWithHandlers struct {
	Step           Step            `json:"step"`
	ResultHandlers []ResultHandler `json:"result_handlers"`
}

type ResultHandler struct {
	Type               string             `json:"type"`
	IfConditions       json.RawMessage    `json:"if_conditions"`
	ResultHandlerOrder int32              `json:"result_handler_order"`
	OutputContentType  string             `json:"output_content_type"`
	HistoryContentType *string            `json:"history_content_type,omitempty"`
	UiContentType      *string            `json:"ui_content_type,omitempty"`
	Steps              []StepWithHandlers `json:"steps"`
}

// Tool Definition Types for LLM Function Calling
// ==============================================

// JSONSchemaType represents valid JSON schema types
type JSONSchemaType string

const (
	JSONSchemaTypeString  JSONSchemaType = "string"
	JSONSchemaTypeNumber  JSONSchemaType = "number"
	JSONSchemaTypeBoolean JSONSchemaType = "boolean"
	JSONSchemaTypeObject  JSONSchemaType = "object"
	JSONSchemaTypeArray   JSONSchemaType = "array"
)

// JSONSchemaProperty represents a property in a JSON schema
type JSONSchemaProperty struct {
	Type        JSONSchemaType      `json:"type"`
	Description string              `json:"description,omitempty"`
	Items       *JSONSchemaProperty `json:"items,omitempty"` // For array types
	Enum        []string            `json:"enum,omitempty"`  // For enum types
}

// JSONSchema represents a JSON schema for tool input validation
type JSONSchema struct {
	Type       JSONSchemaType                `json:"type"`
	Properties map[string]JSONSchemaProperty `json:"properties,omitempty"`
	Required   []string                      `json:"required,omitempty"`
	Items      *JSONSchemaProperty           `json:"items,omitempty"` // For array types
}

// Tool represents a tool definition for LLM function calling
type Tool struct {
	Name                string          `json:"name"`
	Description         string          `json:"description"`
	InputSchema         JSONSchema      `json:"input_schema"`
	ActionType          string          `json:"action_type"`
	Parameters          json.RawMessage `json:"parameters"`
	BotOutputVisibility string          `json:"bot_output_visibility,omitempty"`
	HistoryContentType  string          `json:"history_content_type,omitempty"`
	UiContentType       string          `json:"ui_content_type,omitempty"`
	AsRoot              bool            `json:"as_root,omitempty"`
}

// Result Types (for actions and step execution)
// =============================================

// Dict represents a dictionary/map type for JSON data
type Dict = map[string]any

// Message represents a message in the conversation
type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// SystemParameters represents system-provided parameters available to agents
type SystemParameters struct {
	CurrentDate     string    `json:"current_date"`     // Current date in ISO format
	Messages        []Message `json:"messages"`         // All messages in the thread
	SessionMessages []Message `json:"session_messages"` // Messages in current session
	CurrentMessage  string    `json:"current_message"`  // The current user message
	ThreadID        uuid.UUID `json:"thread_id"`        // Current thread ID
	OrganizationID  string    `json:"organization_id"`  // Current organization ID
	InvocationID    uuid.UUID `json:"invocation_id"`    // Current invocation ID
}

// InvocationEventType represents SSE event types for bot invocations
type InvocationEventType string

const (
	InvocationEventTypeBotStarted InvocationEventType = "bot started"

	// 'output' from the bot's perspective, not from the user's -
	// these may be translated into step outputs based on the output channels
	InvocationEventTypeMessageCreated  InvocationEventType = "message created"
	InvocationEventTypeMessageFinished InvocationEventType = "message finished" // so we know no more content will be added

	// Intermediate message content events
	InvocationEventTypeMessageContentDelta  InvocationEventType = "message content delta"
	InvocationEventTypeCreateMessageContent InvocationEventType = "create message content" // flushes content and creates new content
	// for the final (non-incremental) message content so we can create the db records
	InvocationEventTypeMessageContentResult InvocationEventType = "message content result"

	// Messages from actions get converted into step outputs if the output channel
	// does not include MessageOutputChannel
	InvocationEventTypeStepOutputCreated  InvocationEventType = "step output created"
	InvocationEventTypeStepOutputFinished InvocationEventType = "step output finished"

	InvocationEventTypeStepOutputContentDelta  InvocationEventType = "step output content delta"
	InvocationEventTypeCreateStepOutputContent InvocationEventType = "create step output content"
	// for the final (non-incremental) message content so we can create the db records
	InvocationEventTypeStepOutputContentResult InvocationEventType = "step output content result"

	// Intermediate invocation events from steps
	InvocationEventTypeStepStarted          InvocationEventType = "step started"
	InvocationEventTypeStepResult           InvocationEventType = "step result"
	InvocationEventTypeResultHandlerStarted InvocationEventType = "result handler started"

	// Step/invocation errors (not status as these affect control flow, where status
	// only provides information)
	InvocationEventTypeRequiresInfo InvocationEventType = "requires info"
	InvocationEventTypeError        InvocationEventType = "error"

	// Status updates
	InvocationEventTypeStatus InvocationEventType = "status"

	// Final action invocation event
	InvocationEventTypeResult InvocationEventType = "result"

	// Dataset events
	InvocationEventTypeDatasetCreated InvocationEventType = "dataset created"
)

// Status represents the status of an action or invocation operation
type Status string

const (
	// Actions can invoke bots and bots rely on the result of actions,
	// so share statuses
	StatusSkipped      Status = "skipped"
	StatusSuccess      Status = "success"
	StatusBreak        Status = "break" // exit step loop
	StatusError        Status = "error"
	StatusRequiresInfo Status = "requires info"
	StatusGoToStep     Status = "go to step"
)

// Error represents error types that can occur during execution
type Error string

const (
	ErrorActionNotFound Error = "action not found"
	ErrorInternalError  Error = "internal error"
	ErrorInfoNeeded     Error = "info needed"
	ErrorTerminated     Error = "terminated"
	ErrorBadRequest     Error = "bad request"
)

// Result of an action or invocation operation - gathering params, running a handler, a step etc
type Result struct {
	Status     Status  `json:"status"`
	Parameters *Dict   `json:"parameters"` // input parameters
	Output     *Dict   `json:"output"`
	Message    *string `json:"message"`
	Error      *Error  `json:"error"`
}
