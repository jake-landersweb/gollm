package gollm

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLLMGPT(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestLLMGPT")
	ctx := context.Background()

	model := NewLanguageModel(test_user_id, logger, nil)
	conversation := NewConversation("You are being used in a go test environment to validate your API calls are working.")
	conversation = append(conversation, NewUserMessage("Testing 1,2,3 ..."))

	response, err := model.Completion(ctx, &CompletionInput{
		Model:        gpt3_model,
		Temperature:  1.0,
		Conversation: conversation,
	})
	require.Nil(t, err)
	logger.DebugContext(ctx, response.Message.Message)
}

func TestLLMGemini(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestLLMGemini")
	ctx := context.Background()

	model := NewLanguageModel(test_user_id, logger, nil)
	conversation := NewConversation("You are being used in a go test environment to validate your API calls are working.")
	conversation = append(conversation, NewUserMessage("Testing 1,2,3 ..."))

	response, err := model.Completion(ctx, &CompletionInput{
		Model:        gemini_model,
		Temperature:  0.5,
		Json:         false,
		Conversation: conversation,
	})
	require.Nil(t, err)
	logger.DebugContext(ctx, response.Message.Message)
}

func TestLLMAnthropic(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestLLMAnthropic")
	ctx := context.Background()

	model := NewLanguageModel(test_user_id, logger, nil)
	conversation := NewConversation("You are being used in a go test environment to validate your API calls are working.")
	conversation = append(conversation, NewUserMessage("Testing 1,2,3 ..."))

	response, err := model.Completion(ctx, &CompletionInput{
		Model:        anthropic_claude3,
		Temperature:  0.5,
		Conversation: conversation,
	})
	require.Nil(t, err)
	logger.DebugContext(ctx, response.Message.Message)
}

func TestLLMMulti(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestLLMMulti")
	ctx := context.Background()

	model := NewLanguageModel(test_user_id, logger, nil)
	conversation := NewConversation("You are a pirate on a deserted island")
	conversation = append(conversation, NewUserMessage("Where is the treasure matey?"))

	var err error
	input1 := &CompletionInput{
		Model:        gemini_model,
		Temperature:  0.7,
		Json:         false,
		Conversation: conversation,
	}
	_, err = TokenEstimate(input1.Model, input1.Conversation[len(input1.Conversation)-1].Message)
	require.Nil(t, err)
	// run a gpt completion
	response1, err := model.Completion(ctx, input1)
	require.Nil(t, err)
	if err != nil {
		return
	}
	conversation = append(conversation, response1.Message)
	conversation = append(conversation, NewUserMessage("Are you sure? You must show me now or suffer!"))

	input2 := &CompletionInput{
		Model:        gpt3_model,
		Temperature:  1.3,
		Json:         false,
		Conversation: conversation,
	}
	_, err = TokenEstimate(input2.Model, input2.Conversation[len(input2.Conversation)-1].Message)
	require.Nil(t, err)

	// run a gemini completion
	response2, err := model.Completion(ctx, input2)
	require.Nil(t, err)
	if err != nil {
		return
	}

	conversation = append(conversation, response2.Message)
	conversation = append(conversation, NewUserMessage("Aha! Thats more like it! Treasure for everyone!"))

	input3 := &CompletionInput{
		Model:        anthropic_claude3,
		Temperature:  0.7,
		Json:         false,
		Conversation: conversation,
	}
	_, err = TokenEstimate(input3.Model, input3.Conversation[len(input3.Conversation)-1].Message)
	require.Nil(t, err)

	// run an anthropic completion
	response3, err := model.Completion(ctx, input3)
	require.Nil(t, err)
	if err != nil {
		return
	}
	conversation = append(conversation, response3.Message)

	PrintConversation(conversation)
	require.Equal(t, 7, len(conversation))
}
