package ltypes

// Message represents a single message in a conversation.
type AnthropicMessage struct {
	Role    string              `json:"role"`    // "user" or "assistant"
	Content []*AnthropicContent `json:"content"` // Can be a string or an array of content blocks
}

// Metadata describes metadata about the request.
type AnthropicMetadata struct {
	StopSequences []string `json:"stop_sequences,omitempty"` // Custom text sequences to stop generation
}

// RequestConfig represents the configuration for a request to the model.
type AnthropicRequest struct {
	Model       string               `json:"model"`    // The model version, e.g., "claude-2.1"
	Messages    []*AnthropicMessage  `json:"messages"` // Array of input messages
	Tools       []*AnthropicTool     `json:"tools"`
	ToolChoice  *AnthropicToolChoice `json:"tool_choice,omitempty"` // Force the model to select a tool
	System      string               `json:"system,omitempty"`      // System prompt, if any
	MaxTokens   int                  `json:"max_tokens"`            // Maximum number of tokens to generate
	Metadata    *AnthropicMetadata   `json:"metadata,omitempty"`    // Metadata about the request
	Stream      bool                 `json:"stream,omitempty"`      // Whether to stream the response
	Temperature float64              `json:"temperature,omitempty"` // Randomness in response
	TopP        float64              `json:"top_p,omitempty"`       // Nucleus sampling probability
	TopK        int                  `json:"top_k,omitempty"`       // Sample from the top K options
}

type AnthropicToolChoice struct {
	Type string `json:"type"`
	Name string `json:"name"`
}
