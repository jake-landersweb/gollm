package ltypes

// Tool represents the details of a tool that the model may use to generate a response.
type GemTool struct {
	FunctionDeclarations []GemFunctionDeclaration `json:"functionDeclarations"` // Optional. A list of FunctionDeclarations available to the model.
}

// FunctionDeclaration represents a function declaration as defined by the OpenAPI 3.03 specification.
type GemFunctionDeclaration struct {
	Name        string               `json:"name"`        // Required. The name of the function.
	Description string               `json:"description"` // Required. A brief description of the function.
	Parameters  map[string]GemSchema `json:"parameters"`  // Optional. Describes the parameters to this function.
}

// Schema represents the Schema object allowing the definition of input and output data types.
type GemSchema struct {
	Type        string               `json:"type"`        // Required. Data type.
	Format      string               `json:"format"`      // Optional. The format of the data.
	Description string               `json:"description"` // Optional. A brief description of the parameter.
	Nullable    bool                 `json:"nullable"`    // Optional. Indicates if the value may be null.
	Enum        []string             `json:"enum"`        // Optional. Possible values of the element.
	Properties  map[string]GemSchema `json:"properties"`  // Optional. Properties of Type.OBJECT.
	Required    []string             `json:"required"`    // Optional. Required properties of Type.OBJECT.
	Items       *GemSchema           `json:"items"`       // Optional. Schema of the elements of Type.ARRAY.
}
