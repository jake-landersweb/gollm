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

func TestGeminiTokens(t *testing.T) {
	tokens, err := geminiTokenizerAccurate("This is an input string where I would like to know how many tokens make it up. Some grammer, can also be us'ed potentially (hopefully): yes.", gemini_model)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	fmt.Println("Tokens:", tokens)
	assert.Equal(t, 35, tokens)
}

func TestGeminiTextCompletion(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestGeminiTextCompletion")
	ctx := context.Background()

	// make the messages
	raw := make([]*Message, 0)
	raw = append(raw, NewSystemMessage("You are a model that is being used to validate that method calls to your api work in a go testing environment."))
	raw = append(raw, NewUserMessage("Please respond with a single sentence."))
	messages := MessagesToGemini(raw)

	llm := NewLanguageModel(test_user_id, logger, nil)
	response, err := llm.geminiCompletion(ctx, logger, gemini_model, 0.5, false, "", messages, nil)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	fmt.Println(*response)
}

func TestGeminiJSONCompletion(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestGeminiTextCompletion")
	ctx := context.Background()

	schema := `{"message": string, "date": int}`

	// make the messages
	raw := make([]*Message, 0)
	raw = append(raw, NewSystemMessage("You are a model that is being used to validate that method calls to your api work in a go testing environment."))
	raw = append(raw, NewUserMessage("Please respond with a reasonable response."))
	messages := MessagesToGemini(raw)

	fmt.Println(*messages[0])

	llm := NewLanguageModel(test_user_id, logger, nil)
	response, err := llm.geminiCompletion(ctx, logger, gemini_model, 0.5, true, schema, messages, nil)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	// parse the json
	tmp := struct {
		Message string `json:"message"`
		Date    int    `json:"date"`
	}{}
	err = json.Unmarshal([]byte(response.Candidates[0].Content.Parts[0].Text), &tmp)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	assert.NotEmpty(t, tmp.Message)
	assert.NotEmpty(t, tmp.Date)
}

func TestGeminiToolUse(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestGPTToolUsage")
	tools := make([]*Tool, 0)
	tools = append(tools, &Tool{
		Title:       "get_weather",
		Description: "Gets the weather in celcius for the specified city. Use this function when a user requests the weather.",
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

	debugPrint(ToolsToGemini(tools))

	messages := make([]*Message, 0)
	messages = append(messages, NewSystemMessage("You are a model in a testing environment to test the implementation of tool use for language models. Act as normal."))
	messages = append(messages, NewUserMessage("What is the weather in San Francisco today?"))

	fmt.Println("Gem Messages:")
	debugPrint(MessagesToGemini(messages))

	llm := NewLanguageModel(test_user_id, logger, nil)
	response, err := llm.geminiCompletion(context.TODO(), logger, gemini_model, 0.5, false, "", MessagesToGemini(messages), ToolsToGemini(tools))
	require.Nil(t, err)
	fmt.Println("Gem Response:")
	debugPrint(response)

	latestMessage := NewMessageFromGemini(&response.Candidates[0].Content)
	messages = append(messages, latestMessage)
	PrintConversation(messages)
	require.Equal(t, latestMessage.Role, RoleToolCall)

	// create a fake tool response
	messages = append(messages, NewToolResultMessage(latestMessage.ToolUseID, latestMessage.ToolName, "35 degrees"))

	fmt.Println("Gem Messages:")
	debugPrint(MessagesToGemini(messages))

	// create a new completion
	response, err = llm.geminiCompletion(context.TODO(), logger, gemini_model, 0.5, false, "", MessagesToGemini(messages), ToolsToGemini(tools))
	require.NoError(t, err)
	fmt.Println("Gem Response:")
	debugPrint(response)

	// add message to the array
	latestMessage = NewMessageFromGemini(&response.Candidates[0].Content)
	messages = append(messages, latestMessage)
	require.Equal(t, RoleAI, latestMessage.Role)
	PrintConversation(messages)
}
