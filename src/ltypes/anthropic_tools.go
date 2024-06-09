package ltypes

type AnthropicTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema *ToolSchema `json:"input_schema"`
}
