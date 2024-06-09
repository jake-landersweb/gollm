package ltypes

// Tool represents the details of a tool that the model may use to generate a response.
type GemTool struct {
	FunctionDeclarations []*GemFunctionDeclaration `json:"functionDeclarations"` // Optional. A list of FunctionDeclarations available to the model.
}

// FunctionDeclaration represents a function declaration as defined by the OpenAPI 3.03 specification.
type GemFunctionDeclaration struct {
	Name        string      `json:"name"`        // Required. The name of the function.
	Description string      `json:"description"` // Required. A brief description of the function.
	Parameters  *ToolSchema `json:"parameters"`  // Optional. Describes the parameters to this function.
}
