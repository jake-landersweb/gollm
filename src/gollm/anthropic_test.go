package gollm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/jake-landersweb/gollm/v2/src/ltypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnthropicTextCompletion(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestAnthropicTextCompletion")
	ctx := context.Background()

	// make the messages
	raw := make([]*Message, 0)
	raw = append(raw, NewSystemMessage("You are a model that is being used to validate that method calls to your api work in a go testing environment."))
	raw = append(raw, NewUserMessage("Please respond with a single sentence."))
	messages := MessagesToAnthropic(raw)

	llm := NewLanguageModel(test_user_id, logger, nil)
	response, err := llm.anthropicCompletion(ctx, logger, anthropic_claude3, 0.5, false, "", messages, nil, false, "")
	assert.Nil(t, err)
	if err != nil {
		return
	}
	assert.NotEmpty(t, response.Content)
	assert.NotEmpty(t, response.Content[0].Text)
}

func TestAnthropicJSONCompletion(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestAnthropicJSONCompletion")
	ctx := context.Background()

	schema := `{"message": string, "date": int}`

	// make the messages
	raw := make([]*Message, 0)
	raw = append(raw, NewSystemMessage("You are a model that is being used to validate that method calls to your api work in a go testing environment."))
	raw = append(raw, NewUserMessage("Please respond with a reasonable response."))
	messages := MessagesToAnthropic(raw)

	fmt.Println(*messages[0])

	llm := NewLanguageModel(test_user_id, logger, nil)
	response, err := llm.anthropicCompletion(ctx, logger, anthropic_claude3, 0.5, true, schema, messages, nil, false, "")
	assert.Nil(t, err)
	if err != nil {
		return
	}

	// parse the json
	tmp := struct {
		Message string `json:"message"`
		Date    int    `json:"date"`
	}{}
	err = json.Unmarshal([]byte(response.Content[0].Text), &tmp)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	assert.NotEmpty(t, tmp.Message)
	assert.NotEmpty(t, tmp.Date)
}

func TestAnthropicToolUse(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestAnthropicToolUse")
	tools := make([]*Tool, 0)
	tools = append(tools, &Tool{
		Title:       "get_weather",
		Description: "Gets the weather in celcius for the specified city.",
		Schema: &ltypes.ToolSchema{
			Type: "object",
			Properties: map[string]*ltypes.ToolSchema{
				"city_name": {
					Type:        "string",
					Description: "The name of a US city in the form of '<CITY>, <STATE_CODE>'. Such as 'Portland, OR'.",
				},
			},
		},
	})

	fmt.Println("INTERNAL TOOLS:")
	debugPrint(tools)
	fmt.Println("ANTHROPIC TOOLS:")
	debugPrint(ToolsToAnthropic(tools))

	messages := make([]*Message, 0)
	messages = append(messages, NewSystemMessage("You are a model in a testing environment to test the implementation of tool use for language models. Act as normal."))
	messages = append(messages, NewUserMessage("What is the weather in San Francisco today?"))

	fmt.Println("ANTHROPIC MESSAGES:")
	debugPrint(MessagesToAnthropic(messages))

	// send the tool use request
	llm := NewLanguageModel(test_user_id, logger, nil)
	response, err := llm.anthropicCompletion(context.TODO(), logger, anthropic_claude3, 0.5, false, "", MessagesToAnthropic(messages), ToolsToAnthropic(tools), true, tools[0].Title)
	require.NoError(t, err)

	debugPrint(response)

	// add the message
	latestMessage := NewMessageFromAnthropic(response)
	messages = append(messages, latestMessage)
	fmt.Println(latestMessage)
	require.Equal(t, RoleToolCall, latestMessage.Role)

	fmt.Println("INTERNAL MESSAGES")
	PrintConversation(messages)

	// add a tool response
	messages = append(messages, NewToolResultMessage(latestMessage.ToolUseID, latestMessage.ToolName, "35 degrees"))

	fmt.Println("INPUT ANTH MESSAGES:")
	anthMsg := MessagesToAnthropic(messages)
	debugPrint(anthMsg)
	response, err = llm.anthropicCompletion(context.TODO(), logger, anthropic_claude3, 0.5, false, "", anthMsg, ToolsToAnthropic(tools), false, "")
	require.NoError(t, err)

	// ensure the response was a valid assistant message
	latestMessage = NewMessageFromAnthropic(response)
	messages = append(messages, latestMessage)
	fmt.Println(latestMessage)
	require.Equal(t, RoleAI, latestMessage.Role)

	fmt.Println("AHTHROPIC CONVERSATION:")
	debugPrint(MessagesToAnthropic(messages))

	fmt.Println("CONVERSATION:")
	PrintConversation(messages)
}
