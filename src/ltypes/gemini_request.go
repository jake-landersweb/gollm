package ltypes

// RequestBody represents the body of a request with content, tools, safety settings, and generation configuration.
type GemRequestBody struct {
	Contents         []*GemContent        `json:"contents"`                   // Required. The content of the current conversation.
	Tools            []*GemTool           `json:"tools,omitempty"`            // Optional. A list of Tools the model may use.
	SafetySettings   []*GemSafetySetting  `json:"safetySettings,omitempty"`   // Optional. A list of unique SafetySetting instances.
	ToolConfig       *GemToolConfig       `json:"toolConfig,omitempty"`       // Optional. defines the configuration for tools
	GenerationConfig *GemGenerationConfig `json:"generationConfig,omitempty"` // Optional. Configuration options for model generation.
}

// Tool calling configuration
type GemToolConfig struct {
	FunctionCallingConfig *FunctionCallingConfig `json:"functionCallingConfig"`
}

// When `Mode` is "ANY", then ONLY the values in `AllowedFunctionNames` will be called
type FunctionCallingConfig struct {
	Mode                 string   `json:"mode"`
	AllowedFunctionNames []string `json:"allowedFunctionNames"`
}
