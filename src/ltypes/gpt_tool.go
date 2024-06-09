package ltypes

type GPTTool struct {
	// The type of the tool. Currently, only function is supported.
	Type string `json:"type"`

	// Function object for this tool
	Function *GPTToolFunction `json:"function"`
}

type GPTToolFunction struct {
	// The name of the function to be called. Must be a-z, A-Z, 0-9, or contain underscores and dashes, with a maximum length of 64.
	Name string `json:"name"`

	// A description of what the function does, used by the model to choose when and how to call the function.
	Description string `json:"description,omitempty"`

	// The parameters the functions accepts, described as a JSON Schema object. See the guide for examples, and the JSON Schema reference for documentation about the format.
	Parameters *ToolSchema `json:"parameters,omitempty"`
}
