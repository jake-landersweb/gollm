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

	// store token records internally incase users want to store state inside the record
	usageRecords []*tokens.UsageRecord
}

type CompletionInput struct {
	Model        string
	Temperature  float64
	Json         bool
	JsonSchema   string
	Conversation []*Message
	Tools        []*Tool
	RequiredTool *Tool
	ProhibitTool bool // if set to true, will not use tools
}

// Valiate the completion input
func (input *CompletionInput) Validate() error {
	if input.Model == "" {
		return fmt.Errorf("`Model` cannot be empty")
	}
	if input.Json && input.JsonSchema == "" {
		return fmt.Errorf("if `Json` is true, then `JsonSchema` cannot be empty")
	}
	if len(input.Conversation) == 0 {
		return fmt.Errorf("the conversation cannot be empty")
	}
	if input.Conversation[len(input.Conversation)-1].Role == RoleAI || input.Conversation[len(input.Conversation)-1].Role == RoleSystem {
		return fmt.Errorf("the last message cannot be an ai message or a system message")
	}
	return nil
}

type CompletionResponse struct {
	Model       string
	StopReason  string
	Message     *Message
	UsageRecord *tokens.UsageRecord
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
	args *NewLanguageModelArgs,
) *LanguageModel {
	args = parseArguments(args)

	return &LanguageModel{
		userId:       userId,
		logger:       logger,
		args:         args,
		usageRecords: make([]*tokens.UsageRecord, 0),
	}
}

/*
Estimates the token usage for a given input request. The accuracy can vary based
on what model you are using:

- GPT3/4: Rough approximation, but should NOT be used for billing reasons

- Gemini: Uses the production tokenization endpoint, will be exact token counts.

- Anthropic: Uses approximate function, should NOT be used for billing reasons
*/
func TokenEstimate(model string, message string) (int, error) {
	if strings.HasPrefix(model, "gpt") {
		return gptTokenizerApproximate("avg", message)
	} else if strings.HasPrefix(model, "gemini") {
		return geminiTokenizerAccurate(message, model)
	} else if strings.HasPrefix(model, "claude") {
		return anthropicTokenizerAproximate(message), nil
	} else {
		return 0, fmt.Errorf("invalid model: %s", model)
	}
}

// Uses the `Model` passed in the `input` to dynamically parse which completion method to use.
func (l *LanguageModel) Completion(ctx context.Context, input *CompletionInput) (*CompletionResponse, error) {
	// parse the input
	if input == nil {
		return nil, fmt.Errorf("the input cannot be nil")
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// create a copy of the conversation
	conversation := make([]*Message, len(input.Conversation))
	copy(conversation, input.Conversation)

	// check the token usage and trim the conversation if needed
	// TODO --

	var response *CompletionResponse
	var err error

	if strings.HasPrefix(input.Model, "gpt") {
		response, err = l.gpt(ctx, input, conversation)
	} else if strings.HasPrefix(input.Model, "gemini") {
		response, err = l.gemini(ctx, input, conversation)
	} else if strings.HasPrefix(input.Model, "claude") {
		response, err = l.anthropic(ctx, input, conversation)
	} else {
		return nil, fmt.Errorf("invalid model type: %s", input.Model)
	}

	if err != nil {
		return nil, err
	}

	// store the token record internally as well
	l.usageRecords = append(l.usageRecords, response.UsageRecord)
	return response, nil
}

// Perform a completion specifically using OpenAI as the provider.
// To be used only when wanting a direct gpt completion. Otherwise, use `DynamicCompletion`.
func (l *LanguageModel) gpt(ctx context.Context, input *CompletionInput, conversation []*Message) (*CompletionResponse, error) {
	logger := l.logger.With("model", input.Model, "temperature", input.Temperature, "json", input.Json, "jsonSchema", input.JsonSchema)
	logger.InfoContext(ctx, "Beginning GPT completion ...")

	requiredTool := ""
	if input.RequiredTool != nil {
		requiredTool = input.RequiredTool.Title
	}

	// send the request
	response, err := l.gptCompletion(
		ctx,
		logger,
		l.userId,
		input.Model,
		input.Temperature,
		input.Json,
		input.JsonSchema,
		MessagesToOpenAI(conversation),
		ToolsToOpenAI(input.Tools),
		input.ProhibitTool,
		requiredTool,
	)
	if err != nil {
		return nil, fmt.Errorf("there was an issue sending the request: %v", err)
	}

	// add the response message to the conversation
	choice := &response.Choices[0]

	// Create a token record for this request
	tokenRecord := tokens.NewUsageRecordFromGPTUsage(input.Model, &response.Usage)

	logger.InfoContext(ctx, "Completed GPT completion")
	logger.DebugContext(ctx, "GPT completion stats", "response", response, "inTokens", response.Usage.PromptTokens, "outTokens", response.Usage.CompletionTokens, "totalTokens", response.Usage.TotalTokens)

	// return the text string of the completion to let the caller parse as needed
	return &CompletionResponse{
		Model:       input.Model,
		StopReason:  choice.FinishReason,
		Message:     NewMessageFromOpenAI(&choice.Message),
		UsageRecord: tokenRecord,
	}, nil
}

// Perform a completion specifically using Google as the provider.
// To be used only when wanting a direct gpt completion. Otherwise, use `DynamicCompletion`.
func (l *LanguageModel) gemini(ctx context.Context, input *CompletionInput, conversation []*Message) (*CompletionResponse, error) {
	logger := l.logger.With("model", input.Model, "temperature", input.Temperature, "json", input.Json, "jsonSchema", input.JsonSchema)
	logger.InfoContext(ctx, "Beginning Gemini completion ...")

	requiredTool := ""
	if input.RequiredTool != nil {
		requiredTool = input.RequiredTool.Title
	}

	// send the request
	response, err := l.geminiCompletion(
		ctx,
		logger,
		input.Model,
		input.Temperature,
		input.Json,
		input.JsonSchema,
		MessagesToGemini(conversation),
		ToolsToGemini(input.Tools),
		input.ProhibitTool,
		requiredTool,
	)
	if err != nil {
		return nil, fmt.Errorf("there was an issue sending the request: %v", err)
	}

	// add the conversation record
	candidate := &response.Candidates[0]

	// parse token usage from the response
	tokenRecord := tokens.NewUsageRecordFromGeminiUsage(input.Model, response.UsageMetadata)

	logger.InfoContext(ctx, "Completed Gemini completion")
	logger.DebugContext(ctx, "Gemini completion stats", "record", *tokenRecord)

	return &CompletionResponse{
		Model:       input.Model,
		StopReason:  candidate.FinishReason,
		Message:     NewMessageFromGemini(&candidate.Content),
		UsageRecord: tokenRecord,
	}, nil
}

// Perform a completion specifically using Anthropic as the provider.
// To be used only when wanting a direct gpt completion. Otherwise, use `DynamicCompletion`.
func (l *LanguageModel) anthropic(ctx context.Context, input *CompletionInput, conversation []*Message) (*CompletionResponse, error) {
	logger := l.logger.With("model", input.Model, "temperature", input.Temperature, "json", input.Json, "jsonSchema", input.JsonSchema)
	logger.InfoContext(ctx, "Beginning Anthropic completion ...")

	requiredTool := ""
	if input.RequiredTool != nil {
		requiredTool = input.RequiredTool.Title
	}

	// send the request
	response, err := l.anthropicCompletion(
		ctx,
		logger,
		input.Model,
		input.Temperature,
		input.Json,
		input.JsonSchema,
		MessagesToAnthropic(conversation),
		ToolsToAnthropic(input.Tools),
		input.ProhibitTool,
		requiredTool,
	)
	if err != nil {
		return nil, fmt.Errorf("there was an issue sending the request: %v", err)
	}

	// add the tokens to the internal counts
	tokenRecord := tokens.NewUsageRecordFromAnthropicUsage(input.Model, response.Usage)

	logger.InfoContext(ctx, "Completed Anthropic completion")
	logger.DebugContext(ctx, "Anthropic completion stats", "response", response, "inTokens", response.Usage.InputTokens, "outTokens", response.Usage.OutputTokens, "totalTokens", response.Usage.InputTokens+response.Usage.OutputTokens)

	// return the text string of the completion to let the caller parse as needed
	return &CompletionResponse{
		Model:       input.Model,
		StopReason:  response.StopReason,
		Message:     NewMessageFromAnthropic(response),
		UsageRecord: tokenRecord,
	}, nil
}

func PrintConversation(conversation []*Message) {
	fmt.Println("\n\n --- LLM Conversation --- ")
	for _, item := range conversation {
		fmt.Println("[[", item.Role.ToString(), "]]")
		switch item.Role {
		case RoleSystem:
			fallthrough
		case RoleAI:
			fallthrough
		case RoleUser:
			fmt.Println(">", item.Message)
		case RoleToolCall:
			fmt.Printf("Tool Call ID: %s\n", item.ToolUseID)
			// compose a function list
			calls := ""
			for k, v := range item.ToolArguments {
				calls += fmt.Sprintf("%s: \"%s\", ", k, v)
			}
			calls = calls[:len(calls)-2]
			fmt.Printf("%s(%s)\n", item.ToolName, calls)
		case RoleToolResult:
			fmt.Printf("Tool Call ID: %s\n", item.ToolUseID)
			fmt.Printf("Tool Name: %s\n", item.ToolName)
			fmt.Printf("Result: %s\n", item.Message)
		}
		fmt.Println("")
	}
}

func (l *LanguageModel) GetUsageRecords() []*tokens.UsageRecord {
	return l.usageRecords
}
