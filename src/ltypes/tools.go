package ltypes

// Schema represents the Schema object allowing the definition of input and output data types.
// This object is valid for OpenAI, Gemini, and Anthropic tool use
type ToolSchema struct {
	Type        string                 `json:"type"`                  // Required. Data type.
	Format      string                 `json:"format,omitempty"`      // Optional. The format of the data.
	Description string                 `json:"description,omitempty"` // Optional. A brief description of the parameter.
	Nullable    bool                   `json:"nullable,omitempty"`    // Optional. Indicates if the value may be null.
	Enum        []string               `json:"enum,omitempty"`        // Optional. Possible values of the element.
	Properties  map[string]*ToolSchema `json:"properties,omitempty"`  // Optional. Properties of Type.OBJECT.
	Required    []string               `json:"required,omitempty"`    // Optional. Required properties of Type.OBJECT.
	Items       *ToolSchema            `json:"items,omitempty"`       // Optional. Schema of the elements of Type.ARRAY.
}
