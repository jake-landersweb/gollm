package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/jake-landersweb/gollm/src/ltypes"
	"github.com/stretchr/testify/assert"
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

	response, err := gptCompletion(context.TODO(), logger, TEST_USER_ID, GPT3_MODEL, 1.0, false, "", messages)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	assert.NotEmpty(t, response.Choices)
	assert.NotEmpty(t, response.Choices[0].Message)
	assert.Equal(t, response.Usage.TotalTokens, response.Usage.CompletionTokens+response.Usage.PromptTokens)

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

	response, err := gptCompletion(context.TODO(), logger, TEST_USER_ID, GPT3_MODEL, 1.0, true, schema, messages)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	assert.NotEmpty(t, response.Choices)
	assert.NotEmpty(t, response.Choices[0].Message)
	assert.Equal(t, response.Usage.TotalTokens, response.Usage.CompletionTokens+response.Usage.PromptTokens)

	fmt.Println(response.Choices[0].Message)

	// parse the json
	tmp := struct {
		Message string `json:"message"`
		Date    int    `json:"date"`
	}{}
	err = json.Unmarshal([]byte(response.Choices[0].Message.Content), &tmp)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	assert.NotEmpty(t, tmp.Message)
	assert.NotEmpty(t, tmp.Date)
}
