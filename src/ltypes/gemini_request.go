package ltypes

// RequestBody represents the body of a request with content, tools, safety settings, and generation configuration.
type GemRequestBody struct {
	Contents         []*GemContent       `json:"contents"`                   // Required. The content of the current conversation.
	Tools            []*GemTool          `json:"tools,omitempty"`            // Optional. A list of Tools the model may use.
	SafetySettings   []*GemSafetySetting `json:"safetySettings,omitempty"`   // Optional. A list of unique SafetySetting instances.
	GenerationConfig GemGenerationConfig `json:"generationConfig,omitempty"` // Optional. Configuration options for model generation.
}
