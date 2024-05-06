package gollm

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jake-landersweb/gollm/v2/src/tokens"
)

type LanguageModel struct {
	userId string
	logger *slog.Logger
	args   *NewLanguageModelArgs

	conversation []*LanguageModelMessage
	tokenRecords []*tokens.TokenRecord
}

type CompletionInput struct {
	Model       string
	Temperature float64
	Json        bool
	JsonSchema  string
	Input       string
}

type NewLanguageModelArgs struct {
	// OpenAI Configs
	GptBaseUrl   string
	GptMaxTokens int
	OpenAIApiKey string // If not defined, the env variable `OPENAI_API_KEY` will be used

	// Gemini Configs
	GeminiBaseUrl string
	GeminiApiKey  string // If not defined, the env variable `GEMINI_API_KEY` will be used

	// Anthropic Configs
	AnthropicBaseUrl   string
	AnthropicVersion   string
	AnthropicMaxTokens int
	AnthropicApiKey    string // If not defined, the env variable `ANTHROPIC_API_KEY` will be used
}

func parseArguments(args *NewLanguageModelArgs) *NewLanguageModelArgs {
	if args == nil {
		args = &NewLanguageModelArgs{}
	}
	if args.GptBaseUrl == "" {
		args.GptBaseUrl = gpt_base_url
	}
	if args.GptMaxTokens == 0 {
		args.GptMaxTokens = gpt_max_tokens
	}
	if args.GeminiBaseUrl == "" {
		args.GeminiBaseUrl = gemini_base_url
	}
	if args.AnthropicBaseUrl == "" {
		args.AnthropicBaseUrl = anthropic_base_url
	}
	if args.AnthropicVersion == "" {
		args.AnthropicVersion = anthropic_version
	}
	if args.AnthropicMaxTokens == 0 {
		args.AnthropicMaxTokens = anthropic_max_tokens
	}
	return args
}

// Create a new language model from a system message.
func NewLanguageModel(
	userId string,
	logger *slog.Logger,
	sysMessage string,
	args *NewLanguageModelArgs,
) *LanguageModel {
	args = parseArguments(args)

	if sysMessage == "" {
		sysMessage = default_system_message
	}

	// initialize the conversation
	conversation := make([]*LanguageModelMessage, 0)
	conversation = append(conversation, NewSystemMessage(sysMessage))

	return &LanguageModel{
		userId:       userId,
		logger:       logger,
		args:         args,
		conversation: conversation,
		tokenRecords: make([]*tokens.TokenRecord, 0),
	}
}

// Create a new language model that inherets from a previous conversation
func NewLanguageModelFromConversation(
	userId string,
	logger *slog.Logger,
	conversation []*LanguageModelMessage,
	args *NewLanguageModelArgs,
) *LanguageModel {
	args = parseArguments(args)

	if conversation == nil {
		// initialize the conversation
		conversation = make([]*LanguageModelMessage, 0)
		conversation = append(conversation, NewSystemMessage(default_system_message))
	}

	return &LanguageModel{
		userId:       userId,
		logger:       logger,
		args:         args,
		conversation: conversation,
		tokenRecords: make([]*tokens.TokenRecord, 0),
	}
}

func (l *LanguageModel) GetConversation(input *CompletionInput) []*LanguageModelMessage {
	return l.conversation
}

/*
Estimates the token usage for a given input request. The accuracy can vary based
on what model you are using:

- GPT3/4: Rough approximation, but should NOT be used for billing reasons

- Gemini: Uses the production tokenization endpoint, will be exact token counts.
*/
func (l *LanguageModel) TokenEstimate(input *CompletionInput) (int, error) {
	if strings.HasPrefix(input.Model, "gpt") {
		return gptTokenizerApproximate("avg", input.Input)
	} else if strings.HasPrefix(input.Model, "gemini") {
		return l.geminiTokenizerAccurate(input.Input, input.Model)
	} else if strings.HasPrefix(input.Model, "claude") {
		return anthropicTokenizerAproximate(input.Input), nil
	} else {
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

// Uses the `Model` passed in the `input` to dynamically parse which completion method to use.
// This should be the default method used in most cases, but the provider-specific functions
// (i.e GPTCompletion) are available if needed.
func (l *LanguageModel) DynamicCompletion(ctx context.Context, input *CompletionInput) (string, error) {
	if strings.HasPrefix(input.Model, "gpt") {
		return l.GPTCompletion(ctx, input)
	} else if strings.HasPrefix(input.Model, "gemini") {
		return l.GeminiCompletion(ctx, input)
	} else if strings.HasPrefix(input.Model, "claude") {
		return l.AnthropicCompletion(ctx, input)
	} else {
		return "", fmt.Errorf("invalid model type: %s", input.Model)
	}
}

// Perform a completion specifically using OpenAI as the provider.
// To be used only when wanting a direct gpt completion. Otherwise, use `DynamicCompletion`.
func (l *LanguageModel) GPTCompletion(ctx context.Context, input *CompletionInput) (string, error) {
	logger := l.logger.With("model", input.Model, "temperature", input.Temperature, "json", input.Json, "jsonSchema", input.JsonSchema, "input", input.Input)
	logger.InfoContext(ctx, "Beginning completion ...")

	// add a new conversation record for the specified input
	l.conversation = append(l.conversation, &LanguageModelMessage{
		Role:    RoleUser,
		Message: input.Input,
	})

	// send the request
	response, err := l.gptCompletion(ctx, logger, l.userId, input.Model, input.Temperature, input.Json, input.JsonSchema, LLMMessagesToGPT(l.conversation))
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

// Perform a completion specifically using Google as the provider.
// To be used only when wanting a direct gpt completion. Otherwise, use `DynamicCompletion`.
func (l *LanguageModel) GeminiCompletion(ctx context.Context, input *CompletionInput) (string, error) {
	logger := l.logger.With("model", input.Model, "temperature", input.Temperature, "json", input.Json, "jsonSchema", input.JsonSchema, "input", input.Input)
	logger.InfoContext(ctx, "Beginning completion ...")

	logger.DebugContext(ctx, "Calculating the input tokens ...")
	inTokens, err := l.geminiTokenizerAccurate(input.Input, input.Model)
	if err != nil {
		return "", fmt.Errorf("there was an issue calculating the token usage: %v", err)
	}
	logger.DebugContext(ctx, "Got input tokens", "tokens", inTokens)

	l.conversation = append(l.conversation, &LanguageModelMessage{
		Role:    RoleUser,
		Message: input.Input,
	})

	// send the request
	response, err := l.geminiCompletion(ctx, logger, input.Model, input.Temperature, input.Json, input.JsonSchema, LLMMessagesToGemini(l.conversation))
	if err != nil {
		// remove the message from the conversation
		l.conversation = l.conversation[:len(l.conversation)-1]
		return "", fmt.Errorf("there was an issue sending the request: %v", err)
	}

	logger.DebugContext(ctx, "Calculating the output tokens ...")
	outTokens, err := l.geminiTokenizerAccurate(response.Candidates[0].Content.Parts[0].Text, input.Model)
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

// Perform a completion specifically using Anthropic as the provider.
// To be used only when wanting a direct gpt completion. Otherwise, use `DynamicCompletion`.
func (l *LanguageModel) AnthropicCompletion(ctx context.Context, input *CompletionInput) (string, error) {
	logger := l.logger.With("model", input.Model, "temperature", input.Temperature, "json", input.Json, "jsonSchema", input.JsonSchema, "input", input.Input)
	logger.InfoContext(ctx, "Beginning completion ...")

	// add a new conversation record for the specified input
	l.conversation = append(l.conversation, &LanguageModelMessage{
		Role:    RoleUser,
		Message: input.Input,
	})

	// send the request
	response, err := l.anthropicCompletion(ctx, logger, input.Model, input.Temperature, input.Json, input.JsonSchema, LLMMessagesToAnthropic(l.conversation))
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

func (l *LanguageModel) GetTokenRecords() []*tokens.TokenRecord {
	return l.tokenRecords
}
