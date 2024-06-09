package gollm

import (
	"encoding/json"

	"github.com/jake-landersweb/gollm/v2/src/ltypes"
)

type Tool struct {
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Schema      *ltypes.ToolSchema `json:"schema"`
}

func (t *Tool) ToOpenAI() *ltypes.GPTTool {
	return &ltypes.GPTTool{
		Type: "function",
		Function: &ltypes.GPTToolFunction{
			Name:        t.Title,
			Description: t.Description,
			Parameters:  t.Schema,
		},
	}
}

func (t *Tool) ToGemini() *ltypes.GemFunctionDeclaration {
	return &ltypes.GemFunctionDeclaration{
		Name:        t.Title,
		Description: t.Description,
		Parameters:  t.Schema,
	}
}

func (t *Tool) ToAnthropic() *ltypes.AnthropicTool {
	return &ltypes.AnthropicTool{
		Name:        t.Title,
		Description: t.Description,
		InputSchema: t.Schema,
	}
}

// Converts to OpenAI tools
func ToolsToOpenAI(tools []*Tool) []*ltypes.GPTTool {
	resp := make([]*ltypes.GPTTool, len(tools))
	for i, item := range tools {
		resp[i] = item.ToOpenAI()
	}
	return resp
}

// Converts to Gemini Tools. Gemini has an odd shape. The multi-nested schema is abstracted away
func ToolsToGemini(tools []*Tool) []*ltypes.GemTool {
	funcs := make([]*ltypes.GemFunctionDeclaration, len(tools))
	for i, item := range tools {
		funcs[i] = item.ToGemini()
	}

	resp := make([]*ltypes.GemTool, 1)
	obj := &ltypes.GemTool{
		FunctionDeclarations: funcs,
	}
	resp = append(resp, obj)
	return resp
}

// Converts to Anthropic tools
func ToolsToAnthropic(tools []*Tool) []*ltypes.AnthropicTool {
	resp := make([]*ltypes.AnthropicTool, len(tools))
	for i, item := range tools {
		resp[i] = item.ToAnthropic()
	}
	return resp
}

type ToolCall struct {
	ID        string         `json:"id"`        // Identifier of the tool call. Not applicable for all providers
	Name      string         `json:"name"`      // Name of the calling function. Will match the name of a supplied `Tool` object `Schema`
	Arguments map[string]any `json:"arguments"` // JSON schema of the arguments in the form of a map[string]any
}

func (t *ToolCall) ToOpenAI() []*ltypes.GPTCompletionToolCall {
	// encode
	enc, _ := json.Marshal(t.Arguments)
	resp := make([]*ltypes.GPTCompletionToolCall, 0)
	resp = append(resp, &ltypes.GPTCompletionToolCall{
		ID:   t.ID,
		Type: "function",
		Function: &ltypes.GPTToolCallFunction{
			Name:      t.Name,
			Arguments: string(enc),
		},
	})
	return resp
}

func (t *ToolCall) ToAnthropic() *ltypes.AnthropicContent {
	return &ltypes.AnthropicContent{
		Type:  "tool_use",
		ID:    t.ID,
		Name:  t.Name,
		Input: t.Arguments,
	}
}

func ToolCallFromOpenAI(call []*ltypes.GPTCompletionToolCall) *ToolCall {
	// decode
	args := make(map[string]any)
	json.Unmarshal([]byte(call[0].Function.Arguments), &args)
	return &ToolCall{
		ID:        call[0].ID,
		Name:      call[0].Function.Name,
		Arguments: args,
	}
}

func ToolCallFromAnthropic(call *ltypes.AnthropicContent) *ToolCall {
	return &ToolCall{
		ID:        call.ID,
		Name:      call.Name,
		Arguments: call.Input,
	}
}
