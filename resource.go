package erdotypes

// ResourceType represents the type of resource
type ResourceType string

const (
	ResourceTypeTable           ResourceType = "table"
	ResourceTypeEndpoint        ResourceType = "endpoint"
	ResourceTypeDocument        ResourceType = "document"
	ResourceTypePartialDocument ResourceType = "partial_document"
	ResourceTypeEntity          ResourceType = "entity"
	ResourceTypeSheet           ResourceType = "sheet"
)

// ResourceState represents the state of a resource
type ResourceState string

const (
	ResourceStateActive  ResourceState = "active"
	ResourceStateRemoved ResourceState = "removed"
	ResourceStateDeleted ResourceState = "deleted"
)

// ResourceRelationshipType represents the type of relationship between resources
type ResourceRelationshipType string

const (
	ResourceRelationshipTypeAccepts ResourceRelationshipType = "accepts"
	ResourceRelationshipTypeReturns ResourceRelationshipType = "returns"
	ResourceRelationshipTypeLinksTo ResourceRelationshipType = "links_to"
)

// ResourceAttachType represents how a resource should be attached to dataset operations
type ResourceAttachType string

const (
	ResourceAttachTypeAlways     ResourceAttachType = "always"
	ResourceAttachTypeSearchable ResourceAttachType = "searchable"
)
