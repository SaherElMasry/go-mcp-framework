package backend

// ToolBuilder provides a fluent API for building tool definitions
type ToolBuilder struct {
	tool ToolDefinition
}

// NewTool creates a new tool builder
func NewTool(name string) *ToolBuilder {
	return &ToolBuilder{
		tool: ToolDefinition{
			Name: name,
			InputSchema: Schema{
				Type:       "object",
				Properties: make(map[string]Property),
				Required:   []string{},
			},
		},
	}
}

// Description sets the tool description
func (b *ToolBuilder) Description(desc string) *ToolBuilder {
	b.tool.Description = desc
	return b
}

// StringParam adds a string parameter
func (b *ToolBuilder) StringParam(name, description string, required bool) *ToolBuilder {
	b.tool.InputSchema.Properties[name] = Property{
		Type:        "string",
		Description: description,
	}
	if required {
		b.tool.InputSchema.Required = append(b.tool.InputSchema.Required, name)
	}
	return b
}

// IntParam adds an integer parameter
func (b *ToolBuilder) IntParam(name, description string, required bool, min, max *int) *ToolBuilder {
	var minF, maxF *float64
	if min != nil {
		f := float64(*min)
		minF = &f
	}
	if max != nil {
		f := float64(*max)
		maxF = &f
	}

	b.tool.InputSchema.Properties[name] = Property{
		Type:        "integer",
		Description: description,
		Minimum:     minF,
		Maximum:     maxF,
	}
	if required {
		b.tool.InputSchema.Required = append(b.tool.InputSchema.Required, name)
	}
	return b
}

// BoolParam adds a boolean parameter
func (b *ToolBuilder) BoolParam(name, description string, required bool, defaultVal *bool) *ToolBuilder {
	prop := Property{
		Type:        "boolean",
		Description: description,
	}
	if defaultVal != nil {
		prop.Default = *defaultVal
	}
	b.tool.InputSchema.Properties[name] = prop
	if required {
		b.tool.InputSchema.Required = append(b.tool.InputSchema.Required, name)
	}
	return b
}

// EnumParam adds an enum string parameter
func (b *ToolBuilder) EnumParam(name, description string, required bool, values []string, defaultVal *string) *ToolBuilder {
	prop := Property{
		Type:        "string",
		Description: description,
		Enum:        values,
	}
	if defaultVal != nil {
		prop.Default = *defaultVal
	}
	b.tool.InputSchema.Properties[name] = prop
	if required {
		b.tool.InputSchema.Required = append(b.tool.InputSchema.Required, name)
	}
	return b
}

// Build returns the completed tool definition
func (b *ToolBuilder) Build() ToolDefinition {
	return b.tool
}
