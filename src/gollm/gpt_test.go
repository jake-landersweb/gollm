package gollm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/jake-landersweb/gollm/v2/src/ltypes"
	"github.com/stretchr/testify/require"
)

func TestGPTTextCompletion(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestGPTTextCompletion")
	messages := make([]*ltypes.GPTCompletionMessage, 0)
	messages = append(messages, &ltypes.GPTCompletionMessage{
		Content: "You are a model that is being used to validate that method calls to your api work in a go testing environment.",
		Role:    "system",
	})
	messages = append(messages, &ltypes.GPTCompletionMessage{
		Content: "Please respond with a single sentence.",
		Role:    "user",
	})

	llm := NewLanguageModel(test_user_id, logger, nil)
	response, err := llm.gptCompletion(context.TODO(), logger, test_user_id, gpt3_model, 1.0, false, "", messages, nil)
	require.Nil(t, err)

	require.NotEmpty(t, response.Choices)
	require.NotEmpty(t, response.Choices[0].Message)
	require.Equal(t, response.Usage.TotalTokens, response.Usage.CompletionTokens+response.Usage.PromptTokens)

	fmt.Println(response.Choices[0].Message)
}

func TestGPTJSONCompletion(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestGPTJSONCompletion")
	schema := `{"message": string, "date": int}`

	messages := make([]*ltypes.GPTCompletionMessage, 0)
	messages = append(messages, &ltypes.GPTCompletionMessage{
		Content: "You are a model that is being used to validate that method calls to your api work in a go testing environment.",
		Role:    "system",
	})
	messages = append(messages, &ltypes.GPTCompletionMessage{
		Content: "Please give a reasonable response.",
		Role:    "user",
	})

	llm := NewLanguageModel(test_user_id, logger, nil)
	response, err := llm.gptCompletion(context.TODO(), logger, test_user_id, gpt3_model, 1.0, true, schema, messages, nil)
	require.Nil(t, err)

	require.NotEmpty(t, response.Choices)
	require.NotEmpty(t, response.Choices[0].Message)
	require.Equal(t, response.Usage.TotalTokens, response.Usage.CompletionTokens+response.Usage.PromptTokens)

	fmt.Println(response.Choices[0].Message)

	// parse the json
	tmp := struct {
		Message string `json:"message"`
		Date    int    `json:"date"`
	}{}
	err = json.Unmarshal([]byte(response.Choices[0].Message.Content), &tmp)
	require.Nil(t, err)
	if err != nil {
		return
	}

	require.NotEmpty(t, tmp.Message)
	require.NotEmpty(t, tmp.Date)
}

func TestGPTToolUsage(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestGPTToolUsage")
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

	enc, _ := json.MarshalIndent(tools, "", "    ")
	fmt.Println(string(enc))
	enc, _ = json.MarshalIndent(ToolsToOpenAI(tools), "", "    ")
	fmt.Println(string(enc))

	messages := make([]*Message, 0)
	messages = append(messages, NewSystemMessage("You are a model in a testing environment to test the implementation of tool use for language models. Act as normal."))
	messages = append(messages, NewUserMessage("What is the weather in San Francisco today?"))

	llm := NewLanguageModel(test_user_id, logger, nil)
	response, err := llm.gptCompletion(context.TODO(), logger, test_user_id, gpt3_model, 1.0, false, "", MessagesToOpenAI(messages), ToolsToOpenAI(tools))
	require.Nil(t, err)

	enc, _ = json.MarshalIndent(response, "", "    ")
	fmt.Println(string(enc))

	latestMessage := NewMessageFromOpenAI(&response.Choices[0].Message)
	messages = append(messages, latestMessage)
	require.Equal(t, latestMessage.Role, RoleToolCall)

	// add a tool use result onto the message
	messages = append(messages, NewToolResultMessage(latestMessage.ToolUseID, latestMessage.ToolName, "35 degrees"))

	response, err = llm.gptCompletion(context.TODO(), logger, test_user_id, gpt3_model, 1.0, false, "", MessagesToOpenAI(messages), ToolsToOpenAI(tools))
	require.NoError(t, err)

	enc, _ = json.MarshalIndent(response, "", "    ")
	fmt.Println(string(enc))

	latestMessage = NewMessageFromOpenAI(&response.Choices[0].Message)
	messages = append(messages, latestMessage)
	require.Equal(t, latestMessage.Role, RoleAI)

	fmt.Println("INTERNAL MESSAGES:")
	enc, _ = json.MarshalIndent(messages, "", "    ")
	fmt.Println(string(enc))
	fmt.Println("OPEN AI MESSAGES:")
	enc, _ = json.MarshalIndent(MessagesToOpenAI(messages), "", "    ")
	fmt.Println(string(enc))

	PrintConversation(messages)
}
