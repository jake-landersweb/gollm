package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jake-landersweb/gollm/src/tokens"
)

type LanguageModel struct {
	userId string
	logger *slog.Logger

	conversation []*LLMMessage
	tokenRecords []*tokens.TokenRecord
}

type CompletionInput struct {
	Model       string
	Temperature float64
	Json        bool
	JsonSchema  string
	Input       string
}

func NewLLM(userId string, logger *slog.Logger, sysMsg string) *LanguageModel {
	// initialize the conversation
	conversation := make([]*LLMMessage, 0)
	conversation = append(conversation, NewSystemMessage(sysMsg))

	return &LanguageModel{
		userId:       userId,
		logger:       logger,
		conversation: conversation,
		tokenRecords: make([]*tokens.TokenRecord, 0),
	}
}

/*
Estimates the token usage for a given input request. The accuracy can vary based
on what model you are using:

- GPT3/4: Rough approximation, but should NOT be used for billing reasons

- Gemini: Uses the production tokenization endpoint, will be exact token counts.
*/
func (l *LanguageModel) TokenEstimate(input *CompletionInput) (int, error) {
	switch input.Model {
	case GPT3_MODEL:
		fallthrough
	case GPT4_MODEL:
		return gptTokenizerApproximate("avg", input.Input)
	case GEMINI_MODEL:
		return geminiTokenizerAccurate(input.Input)
	case ANTHROPIC_CLAUDE_INSTANT:
		fallthrough
	case ANTHROPIC_CLAUDE2:
		return anthropicTokenizerAproximate(input.Input), nil

	default:
		return 0, fmt.Errorf("invalid model: %s", input.Model)
	}
}

func (l *LanguageModel) PrintConversation() {
	fmt.Println("\n\n --- LLM Conversation --- ")
	for _, item := range l.conversation {
		fmt.Println("[[", item.Role.ToString(), "]]")
		fmt.Println(">", item.Message)
	}
}

func (l *LanguageModel) GPTCompletion(ctx context.Context, input *CompletionInput) (string, error) {
	logger := l.logger.With("model", input.Model, "temperature", input.Temperature, "json", input.Json, "jsonSchema", input.JsonSchema, "input", input.Input)
	logger.InfoContext(ctx, "Beginning completion ...")

	// add a new conversation record for the specified input
	l.conversation = append(l.conversation, &LLMMessage{
		Role:    RoleUser,
		Message: input.Input,
	})

	// send the request
	response, err := gptCompletion(ctx, logger, l.userId, input.Model, input.Temperature, input.Json, input.JsonSchema, LLMMessagesToGPT(l.conversation))
	if err != nil {
		// remove the message from the conversation
		l.conversation = l.conversation[:len(l.conversation)-1]
		return "", fmt.Errorf("there was an issue sending the request: %v", err)
	}

	// add the response message to the conversation
	l.conversation = append(l.conversation, NewMessageFromGPT(&response.Choices[0].Message))

	// add the tokens to the internal counts
	l.tokenRecords = append(l.tokenRecords, tokens.NewTokenRecordFromGPTUsage(input.Model, &response.Usage))

	logger.InfoContext(ctx, "Completed GPT completion")
	logger.DebugContext(ctx, "GPT completion stats", "response", response, "inTokens", response.Usage.PromptTokens, "outTokens", response.Usage.CompletionTokens, "totalTokens", response.Usage.TotalTokens)

	// return the text string of the completion to let the caller parse as needed
	return response.Choices[0].Message.Content, nil
}

func (l *LanguageModel) GeminiCompletion(ctx context.Context, input *CompletionInput) (string, error) {
	logger := l.logger.With("model", input.Model, "temperature", input.Temperature, "json", input.Json, "jsonSchema", input.JsonSchema, "input", input.Input)
	logger.InfoContext(ctx, "Beginning completion ...")

	logger.DebugContext(ctx, "Calculating the input tokens ...")
	inTokens, err := geminiTokenizerAccurate(input.Input)
	if err != nil {
		return "", fmt.Errorf("there was an issue calculating the token usage: %v", err)
	}
	logger.DebugContext(ctx, "Got input tokens", "tokens", inTokens)

	l.conversation = append(l.conversation, &LLMMessage{
		Role:    RoleUser,
		Message: input.Input,
	})

	// send the request
	response, err := geminiCompletion(ctx, logger, input.Model, input.Temperature, input.Json, input.JsonSchema, LLMMessagesToGemini(l.conversation))
	if err != nil {
		// remove the message from the conversation
		l.conversation = l.conversation[:len(l.conversation)-1]
		return "", fmt.Errorf("there was an issue sending the request: %v", err)
	}

	logger.DebugContext(ctx, "Calculating the output tokens ...")
	outTokens, err := geminiTokenizerAccurate(response.Candidates[0].Content.Parts[0].Text)
	if err != nil {
		return "", fmt.Errorf("there was an issue calculating the token usage for this api call: %v", err)
	}
	logger.DebugContext(ctx, "Got output tokens", "tokens", outTokens)

	// add the conversation record
	l.conversation = append(l.conversation, NewMessageFromGemini(&response.Candidates[0].Content))

	// add the parsed tokens
	l.tokenRecords = append(l.tokenRecords, tokens.NewTokenRecord(input.Model, inTokens, outTokens, inTokens+outTokens))

	logger.InfoContext(ctx, "Completed Gemini completion")
	logger.DebugContext(ctx, "Gemini completion stats", "response", response, "inTokens", inTokens, "outTokens", outTokens, "totalTokens", inTokens+outTokens)

	// return the message
	return response.Candidates[0].Content.Parts[0].Text, nil
}

func (l *LanguageModel) AnthropicCompletion(ctx context.Context, input *CompletionInput) (string, error) {
	logger := l.logger.With("model", input.Model, "temperature", input.Temperature, "json", input.Json, "jsonSchema", input.JsonSchema, "input", input.Input)
	logger.InfoContext(ctx, "Beginning completion ...")

	// add a new conversation record for the specified input
	l.conversation = append(l.conversation, &LLMMessage{
		Role:    RoleUser,
		Message: input.Input,
	})

	// send the request
	response, err := anthropicCompletion(ctx, logger, input.Model, input.Temperature, input.Json, input.JsonSchema, LLMMessagesToAnthropic(l.conversation))
	if err != nil {
		// remove the message from the conversation
		l.conversation = l.conversation[:len(l.conversation)-1]
		return "", fmt.Errorf("there was an issue sending the request: %v", err)
	}

	// add the response message to the conversation
	l.conversation = append(l.conversation, NewMessageFromAnthropic(response.Content[0]))

	// add the tokens to the internal counts
	l.tokenRecords = append(l.tokenRecords, tokens.NewTokenRecordFromAnthropicUsage(input.Model, response.Usage))

	logger.InfoContext(ctx, "Completed GPT completion")
	logger.DebugContext(ctx, "GPT completion stats", "response", response, "inTokens", response.Usage.InputTokens, "outTokens", response.Usage.OutputTokens, "totalTokens", response.Usage.InputTokens+response.Usage.OutputTokens)

	// return the text string of the completion to let the caller parse as needed
	return response.Content[0].Text, nil
}
