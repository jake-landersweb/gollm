package gollm

import (
	"github.com/google/uuid"
	"github.com/jake-landersweb/gollm/v2/src/ltypes"
)

type Role int

const (
	RoleSystem Role = iota
	RoleUser
	RoleAI
	RoleToolCall
	RoleToolResult
)

func (r Role) ToString() string {
	switch r {
	case RoleSystem:
		return "System"
	case RoleUser:
		return "User"
	case RoleAI:
		return "Assistant"
	case RoleToolCall:
		return "Tool Call"
	case RoleToolResult:
		return "Tool Result"
	default:
		return "Unknown"
	}
}

type Message struct {
	Role    Role   `json:"role"`    // Role of the message
	Message string `json:"message"` // Plain text of the message

	ToolUseID     string         `json:"id"`        // If applicable - ID of the tool call
	ToolName      string         `json:"name"`      // If applicable - Name of the tool call
	ToolArguments map[string]any `json:"arguments"` // If applicable - Argments of the tool call. This will only be set on the role: `RoleToolCall`
}

func (m *Message) GetToolCall() *ToolCall {
	if m.ToolUseID == "" {
		return nil
	}
	return &ToolCall{
		ID:        m.ToolUseID,
		Name:      m.ToolName,
		Arguments: m.ToolArguments,
	}
}

func (m *Message) SetToolCall(tc *ToolCall) {
	m.ToolUseID = tc.ID
	m.ToolName = tc.Name
	m.ToolArguments = tc.Arguments
}

// Creates a new system message
func NewSystemMessage(input string) *Message {
	return &Message{
		Role:    RoleSystem,
		Message: input,
	}
}

// Creates a new user message
func NewUserMessage(input string) *Message {
	return &Message{
		Role:    RoleUser,
		Message: input,
	}
}

// Creates a new assistant message
func NewAssistantMessage(input string) *Message {
	return &Message{
		Role:    RoleAI,
		Message: input,
	}
}

func NewToolCallMessage(
	id string,
	name string,
	arguments map[string]any,
	message string,
) *Message {
	return &Message{
		Role:          RoleToolCall,
		Message:       message,
		ToolUseID:     id,
		ToolName:      name,
		ToolArguments: arguments,
	}
}

func NewToolResultMessage(
	id string,
	name string,
	message string,
) *Message {
	return &Message{
		Role:      RoleToolResult,
		Message:   message,
		ToolUseID: id,
		ToolName:  name,
	}
}

// Creates a new conversation from a system message, and returns the
// list with the system message embedded as the first element
func NewConversation(sysMessage string) []*Message {
	conversation := make([]*Message, 0)
	conversation = append(conversation, NewSystemMessage(sysMessage))
	return conversation
}

// Creates a new `LLMMessage` from an input `GPTCompletionMessage`
func NewMessageFromOpenAI(input *ltypes.GPTCompletionMessage) *Message {
	msg := &Message{
		Message: input.Content,
	}
	switch input.Role {
	case "system":
		msg.Role = RoleSystem
	case "assistant":
		// parse if there was a tool call
		if input.ToolCalls != nil && len(input.ToolCalls) != 0 {
			msg.SetToolCall(ToolCallFromOpenAI(input.ToolCalls))
			msg.Role = RoleToolCall
		} else {
			msg.Role = RoleAI
		}
	case "tool":
		msg.Role = RoleToolResult
		msg.ToolUseID = input.ToolCallId
		msg.ToolName = input.Name
	default:
		msg.Role = RoleUser
	}
	return msg
}

/*
For parsing the response of the gemini api into an `LanguageModelMessage`. This is NOT for use
to convert a list of `GemContent` messages, as you should use `LLMMessagesFromGemini` to
ensure the system message is parsed correctly
*/
func NewMessageFromGemini(input *ltypes.GemContent) *Message {
	// return the message parsed from the response, as it will never be a user message,
	// which would trigger an index fault
	return MessagesFromGemini([]*ltypes.GemContent{
		input,
	})[0]
}

/*
Parses an `AnthropicContent` into an `LanguageModelMessage`.
This function should only be used to parse the response from the Anthropic API, as it
does not take into account different roles and message types.
*/
func NewMessageFromAnthropic(input *ltypes.AnthropicResponse) *Message {
	// need to parse the anthropic messages using the messages parser
	msgs := make([]*ltypes.AnthropicMessage, 0)
	msgs = append(msgs, &ltypes.AnthropicMessage{
		Role:    input.Role,
		Content: input.Content,
	})
	// only one message will be parsed here
	return MessagesFromAnthropic(msgs)[0]
}

/*
Parses a list of `GPTCompletionMessage` into a list of `LanguageModelMessage`. These methods should
be used over manual converstion to ensure correct serialization and message parsing from
the implementation specific messaging system and the `LanguageModelMessage` abstraction.
*/
func MessagesFromOpenAI(input []*ltypes.GPTCompletionMessage) []*Message {
	resp := make([]*Message, 0)

	for _, item := range input {
		resp = append(resp, NewMessageFromOpenAI(item))
	}

	return resp
}

func MessagesToOpenAI(messages []*Message) []*ltypes.GPTCompletionMessage {
	resp := make([]*ltypes.GPTCompletionMessage, 0)

	for _, item := range messages {
		message := &ltypes.GPTCompletionMessage{
			Content: item.Message,
		}
		switch item.Role {
		case RoleSystem:
			message.Role = "system"
		case RoleAI:
			message.Role = "assistant"
		case RoleToolCall:
			message.Role = "assistant"
			message.ToolCalls = item.GetToolCall().ToOpenAI()
		case RoleToolResult:
			message.Role = "tool"
			message.ToolCallId = item.ToolUseID
			message.Name = item.ToolName
		default:
			message.Role = "user"
		}

		resp = append(resp, message)
	}

	return resp
}

/*
Parses a list of `GemContent` into a list of `LanguageModelMessage`. These methods should
be used over manual converstion to ensure correct serialization and message parsing from
the implementation specific messaging system and the `LanguageModelMessage` abstraction.
*/
func MessagesFromGemini(messages []*ltypes.GemContent) []*Message {
	resp := make([]*Message, 0)

	// loop over messages and perform parsing
	for index, item := range messages {
		switch item.Role {
		case "system":
			resp = append(resp, NewSystemMessage(item.Parts[0].Text))
		case "model":
			// parse a function call if there was one
			if item.Parts[0].FunctionCall != nil {
				resp = append(resp, NewToolCallMessage(uuid.New().String(), item.Parts[0].FunctionCall.Name, item.Parts[0].FunctionCall.Args, ""))
			} else {
				resp = append(resp, NewAssistantMessage(item.Parts[0].Text))
			}
		default:
			if item.Parts[0].FunctionResponse != nil {
				// parse the tool use id from the previous messages
				toolUseId := resp[index-1].ToolUseID
				resp = append(resp, NewToolResultMessage(toolUseId, item.Parts[0].FunctionResponse.Name, item.Parts[0].FunctionResponse.Response["function_response"].(string)))
			} else {
				resp = append(resp, NewUserMessage(item.Parts[0].Text))
			}
		}
	}

	return resp
}

func MessagesToGemini(messages []*Message) []*ltypes.GemContent {
	resp := make([]*ltypes.GemContent, 0)

	for _, item := range messages {
		switch item.Role {
		case RoleAI:
			resp = append(resp, &ltypes.GemContent{
				Role:  "model",
				Parts: []ltypes.GemPart{{Text: item.Message}},
			})
		case RoleSystem:
			resp = append(resp, &ltypes.GemContent{
				Role:  "system",
				Parts: []ltypes.GemPart{{Text: item.Message}},
			})
		case RoleToolCall:
			resp = append(resp, &ltypes.GemContent{
				Role: "model",
				Parts: []ltypes.GemPart{{FunctionCall: &ltypes.GemFunctionCall{
					Name: item.ToolName,
					Args: item.ToolArguments,
				}}},
			})
		case RoleToolResult:
			resp = append(resp, &ltypes.GemContent{
				Role: "user",
				Parts: []ltypes.GemPart{{FunctionResponse: &ltypes.GemFunctionResponse{
					Name: item.ToolName,
					Response: map[string]any{
						"function_response": item.Message,
					},
				}}},
			})
		default:
			resp = append(resp, &ltypes.GemContent{
				Role:  "user",
				Parts: []ltypes.GemPart{{Text: item.Message}},
			})
		}
	}

	return resp
}

func MessagesFromAnthropic(messages []*ltypes.AnthropicMessage) []*Message {
	resp := make([]*Message, 0)

	for index, msg := range messages {
		switch msg.Role {
		case "system":
			// system messages never use tools
			resp = append(resp, NewSystemMessage(msg.Content[0].Text))
		case "user":
			// check if tool result
			if msg.Content[0].Type == "tool_result" {
				// get the toolId from the previous message in the list
				toolName := ""
				for _, item := range messages[index-1].Content {
					if item.Type == "tool_use" {
						toolName = item.Name
					}
				}

				resp = append(resp, NewToolResultMessage(msg.Content[0].ToolUseID, toolName, msg.Content[0].Content))
			} else {
				// normal message
				resp = append(resp, NewUserMessage(msg.Content[0].Text))
			}
		case "assistant":
			// check for all possible options
			var tmp *Message
			for _, item := range msg.Content {
				switch item.Type {
				case "tool_use":
					tmp = NewToolCallMessage(
						item.ID,
						item.Name,
						item.Input,
						"",
					)
				case "text":
					tmp = NewAssistantMessage(item.Text)
				}
			}
			resp = append(resp, tmp)
		}
	}

	return resp
}

func MessagesToAnthropic(messages []*Message) []*ltypes.AnthropicMessage {
	resp := make([]*ltypes.AnthropicMessage, 0)

	for _, msg := range messages {
		content := make([]*ltypes.AnthropicContent, 0)

		switch msg.Role {
		case RoleUser:
			content = append(content, &ltypes.AnthropicContent{
				Type: "text",
				Text: msg.Message,
			})
			resp = append(resp, &ltypes.AnthropicMessage{
				Role:    "user",
				Content: content,
			})
		case RoleSystem:
			// the system message is parsed from the message array during the request
			// because Anthropic does not handle system messages the same way
			content = append(content, &ltypes.AnthropicContent{
				Type: "text",
				Text: msg.Message,
			})
			resp = append(resp, &ltypes.AnthropicMessage{
				Role:    "system",
				Content: content,
			})
		case RoleAI:
			// initalize the array and seed as a text message
			content = append(content, &ltypes.AnthropicContent{
				Type: "text",
				Text: msg.Message,
			})
			resp = append(resp, &ltypes.AnthropicMessage{
				Role:    "assistant",
				Content: content,
			})
		case RoleToolCall:
			// create the ai message structure with a text message and a tool use message
			content = append(content, &ltypes.AnthropicContent{
				Type: "text",
				Text: "thinking ...",
			})
			content = append(content, &ltypes.AnthropicContent{
				Type:  "tool_use",
				ID:    msg.ToolUseID,
				Name:  msg.ToolName,
				Input: msg.ToolArguments,
			})
			resp = append(resp, &ltypes.AnthropicMessage{
				Role:    "assistant",
				Content: content,
			})
		case RoleToolResult:
			// add tool call results as user messages
			content = append(content, &ltypes.AnthropicContent{
				Type:      "tool_result",
				ToolUseID: msg.ToolUseID,
				Content:   msg.Message,
			})
			resp = append(resp, &ltypes.AnthropicMessage{
				Role:    "user",
				Content: content,
			})
		}
	}

	return resp
}
