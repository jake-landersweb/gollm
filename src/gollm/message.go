package gollm

import (
	"fmt"
	"strings"

	"github.com/jake-landersweb/gollm/v2/src/ltypes"
)

type LanguageModelRole int

const (
	RoleSystem LanguageModelRole = iota
	RoleUser
	RoleAI
)

func (r LanguageModelRole) ToString() string {
	switch r {
	case RoleSystem:
		return "System"
	case RoleUser:
		return "User"
	case RoleAI:
		return "Assistant"
	default:
		return "Unknown"
	}
}

type LanguageModelMessage struct {
	Role    LanguageModelRole
	Message string
}

func NewSystemMessage(input string) *LanguageModelMessage {
	return &LanguageModelMessage{
		Role:    RoleSystem,
		Message: input,
	}
}

func NewUserMessage(input string) *LanguageModelMessage {
	return &LanguageModelMessage{
		Role:    RoleUser,
		Message: input,
	}
}

// Creates a new `LLMMessage` from an input `GPTCompletionMessage`
func NewMessageFromGPT(input *ltypes.GPTCompletionMessage) *LanguageModelMessage {
	msg := &LanguageModelMessage{
		Message: input.Content,
	}
	switch input.Role {
	case "system":
		msg.Role = RoleSystem
	case "assistant":
		msg.Role = RoleAI
	default:
		msg.Role = RoleUser
	}
	return msg
}

/*
For parsing the response of the gemini api into an `LLMMessage`. This is NOT for use
to convert a list of `GemContent` messages, as you should use `LLMMessagesFromGemini` to
ensure the system message is parsed correctly
*/
func NewMessageFromGemini(input *ltypes.GemContent) *LanguageModelMessage {
	msg := &LanguageModelMessage{
		Message: input.Parts[0].Text,
	}
	switch input.Role {
	case "model":
		msg.Role = RoleAI
	default:
		msg.Role = RoleUser
	}
	return msg
}

/*
Parses an `AnthropicContent` into an `LLMMessage`.
This function should only be used to parse the response from the Anthropic API, as it
does not take into account different roles and message types.
*/
func NewMessageFromAnthropic(input *ltypes.AnthropicContent) *LanguageModelMessage {
	return &LanguageModelMessage{
		Role:    RoleAI,
		Message: input.Text,
	}
}

/*
Parses a list of `GPTCompletionMessage` into a list of `LLMMessage`. These methods should
be used over manual converstion to ensure correct serialization and message parsing from
the implementation specific messaging system and the `LLMMessage` abstraction.
*/
func LLMMessagesFromGPT(input []*ltypes.GPTCompletionMessage) []*LanguageModelMessage {
	resp := make([]*LanguageModelMessage, 0)

	for _, item := range input {
		resp = append(resp, NewMessageFromGPT(item))
	}

	return resp
}

func LLMMessagesToGPT(messages []*LanguageModelMessage) []*ltypes.GPTCompletionMessage {
	resp := make([]*ltypes.GPTCompletionMessage, 0)

	for _, item := range messages {
		var role string

		switch item.Role {
		case RoleSystem:
			role = "system"
		case RoleAI:
			role = "assistant"
		default:
			role = "user"
		}

		resp = append(resp, &ltypes.GPTCompletionMessage{
			Role:    role,
			Content: item.Message,
		})
	}

	return resp
}

/*
Parses a list of `GemContent` into a list of `LLMMessage`. These methods should
be used over manual converstion to ensure correct serialization and message parsing from
the implementation specific messaging system and the `LLMMessage` abstraction.
*/
func LLMMessagesFromGemini(messages []*ltypes.GemContent) []*LanguageModelMessage {
	resp := make([]*LanguageModelMessage, 0)

	// loop over messages and perform parsing
	for _, item := range messages {

		// check for system message
		if strings.HasPrefix(item.Parts[0].Text, gemini_system_message) {
			// split into two messages
			parsed := strings.Split(item.Parts[0].Text, "\n\n")
			// remove the system message from the first parsed message
			sys := strings.ReplaceAll(parsed[0], fmt.Sprintf("%s: ", gemini_system_message), "")
			// add the system message
			resp = append(resp, NewSystemMessage(sys))
			// add the user message
			resp = append(resp, NewUserMessage(parsed[1]))
		} else {
			// basic message
			msg := &LanguageModelMessage{
				Message: item.Parts[0].Text,
			}
			switch item.Role {
			case "model":
				msg.Role = RoleAI
			default:
				msg.Role = RoleUser
			}
			resp = append(resp, msg)
		}
	}

	return resp
}

func LLMMessagesToGemini(messages []*LanguageModelMessage) []*ltypes.GemContent {
	resp := make([]*ltypes.GemContent, 0)

	for _, item := range messages {
		var role string
		message := item.Message

		switch item.Role {
		case RoleSystem:
			// create a custom message as there is not a 'system' message in gemini
			role = "user"
			message = fmt.Sprintf("%s: %s", gemini_system_message, item.Message)
		case RoleAI:
			role = "model"
		default:
			if len(resp) == 1 && messages[0].Role == RoleSystem {
				// append to the first user message as there is no system message in gemini
				resp[0].Parts[0].Text = fmt.Sprintf("%s\n\n%s", resp[0].Parts[0].Text, item.Message)
				continue
			} else {
				role = "user"
			}
		}

		// if got here, normal append
		resp = append(resp, &ltypes.GemContent{
			Role: role,
			Parts: []ltypes.GemPart{
				{
					Text: message,
				},
			},
		})
	}

	return resp
}

func LLMMessagesFromAnthropic(messages []*ltypes.AnthropicMessage) []*LanguageModelMessage {
	resp := make([]*LanguageModelMessage, 0)

	for _, item := range messages {
		switch item.Role {
		case "system":
			resp = append(resp, NewSystemMessage(item.Content))
		case "user":
			resp = append(resp, NewUserMessage(item.Content))
		case "assistant":
			resp = append(resp, &LanguageModelMessage{
				Role:    RoleAI,
				Message: item.Content,
			})
		}
	}

	return resp
}

func LLMMessagesToAnthropic(messages []*LanguageModelMessage) []*ltypes.AnthropicMessage {
	resp := make([]*ltypes.AnthropicMessage, 0)

	for _, item := range messages {
		msg := &ltypes.AnthropicMessage{
			Content: item.Message,
		}
		switch item.Role {
		case RoleSystem:
			msg.Role = "system"
		case RoleAI:
			msg.Role = "assistant"
		default:
			msg.Role = "user"
		}
		resp = append(resp, msg)
	}

	return resp
}
