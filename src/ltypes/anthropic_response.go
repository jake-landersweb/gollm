package ltypes

type AnthropicResponse struct {
	ID           string              `json:"id"`
	Type         string              `json:"type"`
	Role         string              `json:"role"`
	Content      []*AnthropicContent `json:"content"`
	Model        string              `json:"model"`
	StopReason   string              `json:"stop_reason"`
	StopSequence interface{}         `json:"stop_sequence"`
	Usage        *AnthropicUsage     `json:"usage"`
	Error        *AnthropicError     `json:"error"`
}

// ContentBlock represents a block of content within a message.
type AnthropicContent struct {
	// NORMAL MESSAGES

	Type string `json:"type"`
	Text string `json:"text,omitempty"`

	// FOR TOOL USE

	ID    string         `json:"id,omitempty"`
	Name  string         `json:"name,omitempty"`
	Input map[string]any `json:"input,omitempty"`

	// FOR TOOL RESULT
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`
}

// Usage provides information on token usage for the request.
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type AnthropicErrorType string

const (
	ANTHROPIC_API_ERROR             AnthropicErrorType = "apierror"
	ANTHROPIC_AUTHENTICATION_ERROR  AnthropicErrorType = "authentication_error"
	ANTHROPIC_INVALID_REQUEST_ERROR AnthropicErrorType = "invalid_request_error"
	ANTHROPIC_NOT_FOUND_ERROR       AnthropicErrorType = "not_found_error"
	ANTHROPIC_OVERLOADED_ERROR      AnthropicErrorType = "overloaded_error"
	ANTHROPIC_PERMISSION_ERROR      AnthropicErrorType = "permission_error"
	ANTHROPIC_RATE_LIMIT_ERROR      AnthropicErrorType = "rate_limit_error"
)

type AnthropicError struct {
	Type    AnthropicErrorType `json:"type"`
	Message string             `json:"message"`
}
