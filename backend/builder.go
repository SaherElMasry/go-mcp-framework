package backend

import "time"

// ToolBuilder provides fluent API for building tool definitions
type ToolBuilder struct {
	name        string
	description string
	parameters  []Parameter
	streaming   bool            // Existing
	cache       ToolCacheConfig // NEW
}

// NewTool creates a new tool builder
func NewTool(name string) *ToolBuilder {
	return &ToolBuilder{
		name:       name,
		parameters: make([]Parameter, 0),
		cache:      DefaultCacheConfig(), // NEW: Safe default
	}
}

// Description sets the tool description
func (b *ToolBuilder) Description(desc string) *ToolBuilder {
	b.description = desc
	return b
}

// StringParam adds a string parameter
func (b *ToolBuilder) StringParam(name, description string, required bool) *ToolBuilder {
	b.parameters = append(b.parameters, Parameter{
		Name:        name,
		Description: description,
		Type:        "string",
		Required:    required,
	})
	return b
}

// IntParam adds an integer parameter
func (b *ToolBuilder) IntParam(name, description string, required bool, min, max *int) *ToolBuilder {
	param := Parameter{
		Name:        name,
		Description: description,
		Type:        "integer",
		Required:    required,
		Minimum:     min,
		Maximum:     max,
	}
	b.parameters = append(b.parameters, param)
	return b
}

// BoolParam adds a boolean parameter
func (b *ToolBuilder) BoolParam(name, description string, required bool, defaultVal *bool) *ToolBuilder {
	param := Parameter{
		Name:        name,
		Description: description,
		Type:        "boolean",
		Required:    required,
	}
	if defaultVal != nil {
		param.Default = *defaultVal
	}
	b.parameters = append(b.parameters, param)
	return b
}

// EnumParam adds an enum parameter
func (b *ToolBuilder) EnumParam(name, description string, required bool, values []string, defaultVal *string) *ToolBuilder {
	param := Parameter{
		Name:        name,
		Description: description,
		Type:        "string",
		Required:    required,
		Enum:        values,
	}
	if defaultVal != nil {
		param.Default = *defaultVal
	}
	b.parameters = append(b.parameters, param)
	return b
}

// Streaming marks the tool as supporting streaming (Existing)
func (b *ToolBuilder) Streaming(enabled bool) *ToolBuilder {
	b.streaming = enabled
	return b
}

// ============================================================
// NEW: Cache Configuration Methods
// ============================================================

// WithCache configures caching for this tool
//
// Example:
//
//	NewTool("read_file").
//	    WithCache(true, 5*time.Minute).  // Cache for 5 minutes
//	    Build()
func (b *ToolBuilder) WithCache(cacheable bool, ttl time.Duration) *ToolBuilder {
	b.cache = ToolCacheConfig{
		Cacheable: cacheable,
	}
	if cacheable && ttl > 0 {
		b.cache.TTL = &ttl
	}
	return b
}

// Cacheable marks the tool as cacheable with default TTL
func (b *ToolBuilder) Cacheable() *ToolBuilder {
	b.cache = ToolCacheConfig{
		Cacheable: true,
	}
	return b
}

// NonCacheable explicitly marks the tool as non-cacheable
func (b *ToolBuilder) NonCacheable() *ToolBuilder {
	b.cache = ToolCacheConfig{
		Cacheable: false,
	}
	return b
}

// CacheTags adds cache tags for categorization
func (b *ToolBuilder) CacheTags(tags ...string) *ToolBuilder {
	b.cache.Tags = append(b.cache.Tags, tags...)
	return b
}

// Build creates the tool definition
func (b *ToolBuilder) Build() ToolDefinition {
	return ToolDefinition{
		Name:        b.name,
		Description: b.description,
		Parameters:  b.parameters,
		Streaming:   b.streaming,
		Cache:       b.cache, // NEW
	}
}
