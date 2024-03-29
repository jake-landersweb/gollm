package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnthropicTextCompletion(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestAnthropicTextCompletion")
	ctx := context.Background()

	// make the messages
	raw := make([]*LLMMessage, 0)
	raw = append(raw, NewSystemMessage("You are a model that is being used to validate that method calls to your api work in a go testing environment."))
	raw = append(raw, NewUserMessage("Please respond with a single sentence."))
	messages := LLMMessagesToAnthropic(raw)

	response, err := anthropicCompletion(ctx, logger, ANTHROPIC_CLAUDE2, 0.5, false, "", messages)
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
	raw := make([]*LLMMessage, 0)
	raw = append(raw, NewSystemMessage("You are a model that is being used to validate that method calls to your api work in a go testing environment."))
	raw = append(raw, NewUserMessage("Please respond with a reasonable response."))
	messages := LLMMessagesToAnthropic(raw)

	fmt.Println(*messages[0])

	response, err := anthropicCompletion(ctx, logger, ANTHROPIC_CLAUDE2, 0.5, true, schema, messages)
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
