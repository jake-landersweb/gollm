package gollm

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLLMGPT(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestLLMGPT")
	ctx := context.Background()

	model := NewLanguageModel(test_user_id, logger, "You are being used in a go test environment to validate your API calls are working.", nil)

	response, err := model.GPTCompletion(ctx, &CompletionInput{
		Model:       gpt3_model,
		Temperature: 1.0,
		Json:        false,
		Input:       "Respond with a single sentence validating your method call.",
	})
	assert.Nil(t, err)
	if err != nil {
		return
	}
	logger.DebugContext(ctx, response)

	assert.Equal(t, 3, len(model.conversation))
	assert.Equal(t, 3, len(model.conversation))
}

func TestLLMGemini(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestLLMGemini")
	ctx := context.Background()

	model := NewLanguageModel(test_user_id, logger, "You are being used in a go test environment to validate your API calls are working.", nil)

	response, err := model.GeminiCompletion(ctx, &CompletionInput{
		Model:       gemini_model,
		Temperature: 0.5,
		Json:        false,
		Input:       "Respond with a single sentence validating your method call.",
	})
	assert.Nil(t, err)
	if err != nil {
		return
	}
	logger.DebugContext(ctx, response)

	assert.Equal(t, 3, len(model.conversation))
	assert.Equal(t, 3, len(model.conversation))
}

func TestLLMAnthropic(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestLLMAnthropic")
	ctx := context.Background()

	model := NewLanguageModel(test_user_id, logger, "You are being used in a go test environment to validate your API calls are working.", nil)

	response, err := model.AnthropicCompletion(ctx, &CompletionInput{
		Model:       anthropic_claude2,
		Temperature: 0.5,
		Json:        false,
		Input:       "Respond with a single sentence validating your method call.",
	})
	assert.Nil(t, err)
	if err != nil {
		return
	}
	logger.DebugContext(ctx, response)

	assert.Equal(t, 3, len(model.conversation))
	assert.Equal(t, 3, len(model.conversation))
}

func TestLLMMulti(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestLLMMulti")
	ctx := context.Background()

	model := NewLanguageModel(test_user_id, logger, "You are a pirate on a deserted island", nil)

	var err error
	input1 := &CompletionInput{
		Model:       gemini_model,
		Temperature: 0.7,
		Json:        false,
		Input:       "Where is the treasure matey?",
	}
	_, err = model.TokenEstimate(input1)
	assert.Nil(t, err)
	// run a gpt completion
	_, err = model.GeminiCompletion(ctx, input1)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	input2 := &CompletionInput{
		Model:       gpt3_model,
		Temperature: 1.3,
		Json:        false,
		Input:       "Are you sure? You must show me now or suffer!",
	}
	_, err = model.TokenEstimate(input2)
	assert.Nil(t, err)

	// run a gemini completion
	_, err = model.GPTCompletion(ctx, input2)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	input3 := &CompletionInput{
		Model:       anthropic_claude2,
		Temperature: 0.7,
		Json:        false,
		Input:       "Aha! Thats more like it! Treasure for everyone!",
	}
	_, err = model.TokenEstimate(input3)
	assert.Nil(t, err)

	// run an anthropic completion
	_, err = model.AnthropicCompletion(ctx, input3)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	model.PrintConversation()

	assert.Equal(t, 7, len(model.conversation))
}
